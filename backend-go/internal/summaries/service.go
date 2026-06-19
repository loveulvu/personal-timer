package summaries

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"personal/internal/llm"
	"personal/internal/memories"
	"personal/internal/stats"
)

const DailyRecentActiveDaysLimit = 5

var (
	ErrStatsQueryFailed            = errors.New("stats query failed")
	ErrLLMGenerationFailed         = errors.New("LLM generation failed")
	ErrSummaryPersistFailed        = errors.New("summary persistence failed")
	ErrSummaryAlreadyExists        = errors.New("summary already exists")
	ErrActionItemNotAcceptable     = errors.New("action item is not acceptable")
	ErrActionItemIndexInvalid      = errors.New("action item index is invalid")
	ErrActionItemProjectInvalid    = errors.New("action item project is not available")
	ErrActionItemTargetDateInvalid = errors.New("target_date is invalid")
)

type summaryRepository interface {
	CreateSummary(ctx context.Context, input CreateSummaryInput) (int64, error)
	SummaryExists(ctx context.Context, summaryType, startDate, endDate string) (bool, error)
	ListSummaries(ctx context.Context, summaryType string) ([]GeneratedSummary, error)
	GetSummaryByID(ctx context.Context, id int64) (*GeneratedSummary, error)
	DeleteSummary(ctx context.Context, id int64) error
	ListRecentDailyActiveDates(ctx context.Context, beforeDate string, limit int) ([]string, error)
	ListDailySummaryTasks(ctx context.Context, dates []string) ([]dailySummaryTaskRow, error)
	ListDailySummarySessions(ctx context.Context, dates []string) ([]dailySummarySessionRow, error)
	FindSummaryProjectByName(ctx context.Context, name string) (*summaryProjectRow, error)
	FindAcceptedDailyTask(ctx context.Context, targetDate string, projectID int64, title string) (*AcceptedDailyTask, error)
	CreateAcceptedDailyTask(ctx context.Context, targetDate string, projectID int64, title string, estimatedMinutes int) (*AcceptedDailyTask, error)
}

type weeklyStatsProvider interface {
	GetWeeklyStats(startDate, endDate string) (*stats.WeeklyStats, error)
}

type memoryExtractor interface {
	ExtractFromSummary(ctx context.Context, summaryID int64) (memories.ExtractionResult, error)
}

type memoryRecallProvider interface {
	RecallRelevantMemories(ctx context.Context, input memories.RecallInput) ([]memories.StudyMemory, error)
}

type Service struct {
	repo            summaryRepository
	statsService    weeklyStatsProvider
	llmClient       llm.Client
	memoryExtractor memoryExtractor
	memoryRecall    memoryRecallProvider
}

func NewService(repo *Repository, statsService *stats.Service, llmClient llm.Client) *Service {
	return &Service{
		repo:         repo,
		statsService: statsService,
		llmClient:    llmClient,
	}
}

func (s *Service) SetMemoryExtractor(extractor memoryExtractor) {
	s.memoryExtractor = extractor
}

func (s *Service) SetMemoryRecall(recall memoryRecallProvider) {
	s.memoryRecall = recall
}

func (s *Service) GenerateDailySummary(ctx context.Context, date string) (*GenerateSummaryResult, error) {
	exists, err := s.repo.SummaryExists(ctx, "daily", date, date)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSummaryAlreadyExists
	}

	recentDates, err := s.repo.ListRecentDailyActiveDates(ctx, date, DailyRecentActiveDaysLimit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}

	windowDates := append([]string{date}, recentDates...)
	tasks, err := s.repo.ListDailySummaryTasks(ctx, windowDates)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}
	sessions, err := s.repo.ListDailySummarySessions(ctx, windowDates)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}

	dailyContext := buildDailySummarySourceData(date, recentDates, tasks, sessions)
	s.recallDailyMemories(ctx, &dailyContext)
	sourceData, err := json.Marshal(dailyContext)
	if err != nil {
		return nil, err
	}

	actionItemList := BuildDailyActionItems(dailyContext)
	var content string
	if isEmptyIncludedDailyData(dailyContext) {
		content = buildEmptyDailySummaryFallback(dailyContext, actionItemList)
		log.Printf(
			"LLM summary generation skipped: summary_type=daily skipped_llm=true reason=empty_included_daily_data source_data_bytes=%d action_items_count=%d",
			len(sourceData),
			len(actionItemList),
		)
	} else {
		dailyPrompt := buildDailyPrompt(string(sourceData))
		content, err = s.generateSummaryWithLog(ctx, "daily", dailyPrompt, len(sourceData))
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrLLMGenerationFailed, err)
		}
	}
	actionItems := marshalSummaryActionItems("daily", actionItemList)

	id, err := s.repo.CreateSummary(ctx, CreateSummaryInput{
		SummaryType: "daily",
		StartDate:   date,
		EndDate:     date,
		Content:     content,
		SourceData:  sourceData,
		ActionItems: actionItems,
	})
	if err != nil {
		if errors.Is(err, ErrSummaryAlreadyExists) {
			return nil, ErrSummaryAlreadyExists
		}
		return nil, fmt.Errorf("%w: %v", ErrSummaryPersistFailed, err)
	}
	s.extractMemories(ctx, id, "daily")

	return &GenerateSummaryResult{SummaryID: id, Content: content, ActionItems: actionItems}, nil
}

func isEmptyIncludedDailyData(source DailySummarySourceData) bool {
	return source.Today.TotalFocusMinutes == 0 &&
		source.Today.TaskCount == 0 &&
		len(source.Today.ProjectBreakdown) == 0
}

func buildEmptyDailySummaryFallback(source DailySummarySourceData, actionItems []SummaryActionItem) string {
	var b strings.Builder
	b.WriteString("# 每日学习总结\n\n")
	b.WriteString("## 1. 今日概览\n\n")
	b.WriteString("今日没有记录纳入学习统计的专注任务。\n")
	if source.Excluded.ExcludedTotalMinutes > 0 {
		b.WriteString("系统检测到存在被排除的非学习记录，例如生活类或休息类任务，这些记录未计入学习总结。\n")
	}

	b.WriteString("\n## 2. 时间分布\n\n")
	b.WriteString("今日没有学习专注时长，无法分析时间段分布。\n")
	b.WriteString("\n## 3. 项目推进\n\n")
	b.WriteString("今日没有纳入学习统计的项目推进记录。\n")
	b.WriteString("\n## 4. 与近期记录的对比\n\n")
	if source.DataQuality.ComparisonDaysWithData > 0 {
		b.WriteString("近期有历史学习记录，但目标日期为空白。今天未形成纳入学习统计的学习记录。\n")
	} else {
		b.WriteString("当前历史样本不足，无法对比。\n")
	}

	b.WriteString("\n## 5. 发现的问题\n\n")
	b.WriteString("- 今天没有纳入学习统计的数据。\n")
	b.WriteString("- 可能是未学习、未使用计时器，或只记录了生活类任务。\n")
	if source.Excluded.UnassignedTaskCount > 0 {
		b.WriteString("- 存在未绑定任务被排除，建议补充项目归属。\n")
	}

	b.WriteString("\n## 6. 明日建议\n\n")
	if len(actionItems) == 0 {
		b.WriteString("- 明天先记录一段明确的学习专注任务。\n")
		return b.String()
	}
	for _, item := range actionItems {
		b.WriteString("- ")
		b.WriteString(item.Title)
		b.WriteString("\n")
	}
	return b.String()
}

func (s *Service) GenerateWeeklySummary(ctx context.Context, startDate, endDate string) (*GenerateSummaryResult, error) {
	exists, err := s.repo.SummaryExists(ctx, "weekly", startDate, endDate)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSummaryAlreadyExists
	}

	weekDates, err := dateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}
	weekTasks, err := s.repo.ListDailySummaryTasks(ctx, weekDates)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}
	weekSessions, err := s.repo.ListDailySummarySessions(ctx, weekDates)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}

	previousWeekDates, err := previousWeekDateRange(startDate)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}
	previousWeekTasks, err := s.repo.ListDailySummaryTasks(ctx, previousWeekDates)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}
	previousWeekSessions, err := s.repo.ListDailySummarySessions(ctx, previousWeekDates)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}

	weeklyContext := buildWeeklySummarySourceData(startDate, endDate, weekDates, weekTasks, weekSessions, previousWeekDates, previousWeekTasks, previousWeekSessions)
	s.recallWeeklyMemories(ctx, &weeklyContext)
	sourceData, err := json.Marshal(weeklyContext)
	if err != nil {
		return nil, err
	}

	weeklyPrompt := buildWeeklyPrompt(string(sourceData))
	content, err := s.generateSummaryWithLog(ctx, "weekly", weeklyPrompt, len(sourceData))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMGenerationFailed, err)
	}
	actionItems := marshalSummaryActionItems("weekly", BuildWeeklyActionItems(weeklyContext))

	id, err := s.repo.CreateSummary(ctx, CreateSummaryInput{
		SummaryType: "weekly",
		StartDate:   startDate,
		EndDate:     endDate,
		Content:     content,
		SourceData:  sourceData,
		ActionItems: actionItems,
	})
	if err != nil {
		if errors.Is(err, ErrSummaryAlreadyExists) {
			return nil, ErrSummaryAlreadyExists
		}
		return nil, fmt.Errorf("%w: %v", ErrSummaryPersistFailed, err)
	}
	s.extractMemories(ctx, id, "weekly")

	return &GenerateSummaryResult{SummaryID: id, Content: content, ActionItems: actionItems}, nil
}

func (s *Service) recallDailyMemories(ctx context.Context, source *DailySummarySourceData) {
	source.RelevantMemories = []SummaryRelevantMemory{}
	if s.memoryRecall == nil {
		return
	}
	s.recallMemories(ctx, "daily", dailyProjectNames(source), &source.RelevantMemories, &source.Warnings)
}

func (s *Service) recallWeeklyMemories(ctx context.Context, source *WeeklySummarySourceData) {
	source.RelevantMemories = []SummaryRelevantMemory{}
	if s.memoryRecall == nil {
		return
	}
	s.recallMemories(ctx, "weekly", weeklyProjectNames(source), &source.RelevantMemories, &source.Warnings)
}

func (s *Service) recallMemories(ctx context.Context, summaryType string, projectNames []string, target *[]SummaryRelevantMemory, warnings *[]string) {
	start := time.Now()
	items, err := s.memoryRecall.RecallRelevantMemories(ctx, memories.RecallInput{
		SummaryType:  summaryType,
		ProjectNames: projectNames,
		Limit:        8,
	})
	elapsed := time.Since(start)
	if err != nil {
		*warnings = append(*warnings, "memory recall failed: "+err.Error())
		log.Printf("memory_recall failed summary_type=%s error=%v elapsed=%s", summaryType, err, elapsed)
		return
	}
	*target = compactRelevantMemories(items)
	log.Printf("memory_recall completed summary_type=%s memory_count=%d elapsed=%s", summaryType, len(items), elapsed)
}

func dailyProjectNames(source *DailySummarySourceData) []string {
	names := make([]string, 0, len(source.Today.ProjectBreakdown))
	for _, project := range source.Today.ProjectBreakdown {
		names = append(names, project.ProjectName)
	}
	return names
}

func weeklyProjectNames(source *WeeklySummarySourceData) []string {
	names := make([]string, 0, len(source.Week.ProjectBreakdown))
	for _, project := range source.Week.ProjectBreakdown {
		names = append(names, project.ProjectName)
	}
	return names
}

func compactRelevantMemories(items []memories.StudyMemory) []SummaryRelevantMemory {
	result := make([]SummaryRelevantMemory, 0, len(items))
	for _, item := range items {
		result = append(result, SummaryRelevantMemory{
			ID:           item.ID,
			MemoryType:   item.MemoryType,
			ScopeType:    item.ScopeType,
			ProjectID:    item.ProjectID,
			Title:        item.Title,
			Content:      trimMemoryContent(item.Content),
			Confidence:   item.Confidence,
			SupportCount: item.SupportCount,
			LastSeenAt:   item.LastSeenAt.Format(time.RFC3339),
		})
	}
	return result
}

func trimMemoryContent(content string) string {
	runes := []rune(content)
	if len(runes) <= 500 {
		return content
	}
	return string(runes[:500])
}

func (s *Service) extractMemories(ctx context.Context, summaryID int64, summaryType string) {
	if s.memoryExtractor == nil {
		return
	}
	start := time.Now()
	result, err := s.memoryExtractor.ExtractFromSummary(ctx, summaryID)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("memory_extraction failed summary_id=%d summary_type=%s error=%v elapsed=%s", summaryID, summaryType, err, elapsed)
		return
	}
	log.Printf(
		"memory_extraction completed summary_id=%d summary_type=%s created_count=%d updated_count=%d skipped_count=%d evidence_count=%d elapsed=%s",
		summaryID,
		summaryType,
		result.CreatedCount,
		result.UpdatedCount,
		result.SkippedCount,
		result.EvidenceCount,
		elapsed,
	)
}

func (s *Service) ListSummaries(ctx context.Context, summaryType string) ([]GeneratedSummary, error) {
	return s.repo.ListSummaries(ctx, summaryType)
}

func (s *Service) GetSummaryByID(ctx context.Context, id int64) (*GeneratedSummary, error) {
	return s.repo.GetSummaryByID(ctx, id)
}

func (s *Service) DeleteSummary(ctx context.Context, id int64) error {
	return s.repo.DeleteSummary(ctx, id)
}

func (s *Service) AcceptActionItem(ctx context.Context, summaryID int64, itemIndex int, targetDate string) (*AcceptActionItemResult, error) {
	if _, err := time.Parse("2006-01-02", targetDate); err != nil {
		return nil, ErrActionItemTargetDateInvalid
	}
	if itemIndex < 0 {
		return nil, ErrActionItemIndexInvalid
	}
	summary, err := s.repo.GetSummaryByID(ctx, summaryID)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return nil, ErrSummaryNotFound
	}

	var items []SummaryActionItem
	if len(summary.ActionItems) > 0 {
		if err := json.Unmarshal(summary.ActionItems, &items); err != nil {
			return nil, ErrActionItemIndexInvalid
		}
	}
	if itemIndex >= len(items) {
		return nil, ErrActionItemIndexInvalid
	}

	item := items[itemIndex]
	if !isAcceptableActionItem(item) {
		return nil, ErrActionItemNotAcceptable
	}
	project, err := s.repo.FindSummaryProjectByName(ctx, item.SuggestedProject)
	if err != nil {
		return nil, err
	}
	if project == nil || !project.IncludeInSummary {
		return nil, ErrActionItemProjectInvalid
	}

	task, err := s.repo.FindAcceptedDailyTask(ctx, targetDate, project.ID, item.Title)
	if err != nil {
		return nil, err
	}
	if task != nil {
		return &AcceptActionItemResult{Created: false, AlreadyExists: true, Task: task, Message: "明日计划中已存在"}, nil
	}

	task, err = s.repo.CreateAcceptedDailyTask(ctx, targetDate, project.ID, item.Title, item.SuggestedMinutes)
	if err != nil {
		return nil, err
	}
	return &AcceptActionItemResult{Created: true, AlreadyExists: false, Task: task}, nil
}

func isAcceptableActionItem(item SummaryActionItem) bool {
	return strings.TrimSpace(item.Title) != "" &&
		strings.TrimSpace(item.SuggestedProject) != "" &&
		item.SuggestedMinutes > 0 &&
		item.Type != "cleanup"
}

func marshalSummaryActionItems(summaryType string, items []SummaryActionItem) json.RawMessage {
	data, err := json.Marshal(items)
	if err != nil {
		log.Printf("summary action_items marshal failed: summary_type=%s error=%v", summaryType, err)
		return json.RawMessage("[]")
	}
	return data
}

func (s *Service) generateSummaryWithLog(ctx context.Context, summaryType, prompt string, sourceDataBytes int) (string, error) {
	startedAt := time.Now()
	content, err := s.llmClient.GenerateSummary(ctx, prompt)
	elapsed := time.Since(startedAt)
	if err != nil {
		log.Printf(
			"LLM summary generation failed: summary_type=%s elapsed=%s prompt_chars=%d source_data_bytes=%d error=%v",
			summaryType,
			elapsed.Round(time.Millisecond),
			utf8.RuneCountInString(prompt),
			sourceDataBytes,
			err,
		)
		return "", err
	}
	log.Printf(
		"LLM summary generation succeeded: summary_type=%s elapsed=%s prompt_chars=%d source_data_bytes=%d",
		summaryType,
		elapsed.Round(time.Millisecond),
		utf8.RuneCountInString(prompt),
		sourceDataBytes,
	)
	return content, nil
}

func buildDailyPrompt(sourceData string) string {
	return buildDailyPromptBase(sourceData) + "\n" + memoryPromptGuidance()
}

func buildDailyPromptBase(sourceData string) string {
	return `你是一个理性的学习记录分析助手。
你必须使用中文输出。
你不能编造输入数据中不存在的项目、任务、时间、备注或趋势。
不要写空泛鼓励，不要写鸡汤。
不要只复述每个项目花了多少时间。
你需要基于输入数据分析：
1. 时间分布
2. 项目推进情况
3. 预计时间与实际时间偏差
4. 重复出现的问题
5. 行为模式
6. 下一步具体调整建议

样本不足规则：
- days_with_data < 2：不要声称存在趋势，只能描述当天情况。
- days_with_data >= 2：可以描述初步变化，但要说明样本有限。
- days_with_data >= 3：可以分析短期模式。
- 如果某项数据缺失，明确说“当前数据不足以判断”。
- source_data 中 excluded 的内容只用于说明数据范围，不要把 excluded 项目计入学习总时长。
- 如果存在 excluded_projects，可以简要说明生活类/未纳入总结的任务已排除。
- 不要把 life/break 类项目当作学习效率问题分析。
- 如果存在 unassigned_task_count，建议用户补充项目分类或绑定项目。

Daily Summary 输出结构必须固定为：

# 每日学习总结

## 1. 今日概览
说明今日总时长、完成任务数、首次开始时间、主要项目。

## 2. 时间分布
分析今天主要学习发生在哪些时间段。

## 3. 项目推进
分析各项目投入和推进情况，不要只列时间。

## 4. 与近期记录的对比
根据 recent_context 分析变化。
如果样本不足，明确说明。

## 5. 发现的问题
指出预估偏差、开始时间偏晚、任务拆分过大、阻塞重复等。

## 6. 明日建议
给出 2-4 条具体建议。

输入 JSON：
` + sourceData
}

func buildWeeklyPrompt(sourceData string) string {
	return buildWeeklyPromptBase(sourceData) + "\n" + memoryPromptGuidance()
}

func buildWeeklyPromptBase(sourceData string) string {
	return `你是一个理性的学习记录分析助手。
你必须使用中文输出。
你不能编造输入数据中不存在的项目、任务、时间、备注或趋势。
不要写空泛鼓励，不要写鸡汤。
不要只复述每个项目花了多少时间。
你需要基于输入数据分析：
1. 本周总投入
2. 活跃天数
3. 项目推进情况
4. 时间段分布
5. 开始时间模式
6. 预计时间与实际时间偏差
7. 重复出现的问题
8. 与上一周的变化
9. 下周具体调整建议

样本不足规则：
- days_with_data < 2：不要声称存在周趋势，只能描述已有记录。
- days_with_data >= 2：可以描述初步变化，但要说明样本有限。
- days_with_data >= 3：可以分析本周模式。
- previous_week_comparison.available = false 时，不要做上周对比。
- 如果某项数据缺失，明确说“当前数据不足以判断”。
- source_data 中 excluded 的内容只用于说明数据范围，不要把 excluded 项目计入学习总时长。
- 如果存在 excluded_projects，可以简要说明生活类/未纳入总结的任务已排除。
- 不要把 life/break 类项目当作学习效率问题分析。
- 如果存在 unassigned_task_count，建议用户补充项目分类或绑定项目。

Weekly Summary 输出结构必须固定为：

# 每周学习总结

## 1. 本周总览
说明总时长、活跃天数、完成任务数、主要项目。

## 2. 项目推进
分析项目投入、稳定性、完成情况，不要只列时间。

## 3. 时间段分布
分析上午 / 下午 / 晚上 / 深夜的投入情况。

## 4. 开始时间模式
分析本周学习启动时间，以及不同项目通常什么时候开始。

## 5. 预计与实际偏差
分析哪些项目经常超时或预估过高。

## 6. 重复问题
根据 finish_note / finish_description 分析反复出现的阻塞或技术点。

## 7. 与上一周对比
如果 previous_week_comparison.available = true，则分析变化。
如果没有上一周数据，明确说明当前数据不足以做上周对比。

## 8. 下周调整建议
给出 3-5 条具体、可执行的建议。

输入 JSON：
` + sourceData
}

func memoryPromptGuidance() string {
	return `长期记忆参考：
- source_data.relevant_memories 是历史数据沉淀出的长期规律；为空数组时表示暂无可用长期记忆。
- 可以用长期记忆辅助分析今天/本周，但不要机械重复 memory 内容。
- 只有当 memory 与当前数据相关时才引用。
- 如果当前数据与 memory 矛盾，要指出可能发生变化，不要强行套用旧规律。
- 不要把 memory 当成绝对事实。`
}

func buildDailySummarySourceData(targetDate string, recentDates []string, tasks []dailySummaryTaskRow, sessions []dailySummarySessionRow) DailySummarySourceData {
	includedTasks, includedSessions, excluded := splitSummaryScope(tasks, sessions)
	contextDates := append([]string{targetDate}, recentDates...)
	tasksByDate := groupTasksByDate(includedTasks)
	sessionsByDate := groupSessionsByDate(includedSessions)
	recentDates = datesWithData(recentDates, tasksByDate, sessionsByDate)
	contextDates = append([]string{targetDate}, recentDates...)
	todayTasks := tasksByDate[targetDate]
	todaySessions := sessionsByDate[targetDate]
	todayHasData := len(todayTasks) > 0 || len(todaySessions) > 0
	daysWithData := len(recentDates)
	if todayHasData {
		daysWithData++
	}

	source := DailySummarySourceData{
		SummaryType: "daily",
		TargetDate:  targetDate,
		DataQuality: DailyDataQuality{
			DaysWithData:           daysWithData,
			CanAnalyzeTrend:        daysWithData >= 3,
			ComparisonWindowDays:   DailyRecentActiveDaysLimit,
			ComparisonDaysWithData: len(recentDates),
		},
		Today: DailySummaryToday{
			TotalFocusMinutes: totalTaskMinutes(todayTasks),
			CompletedTasks:    countCompletedTasks(todayTasks),
			TaskCount:         len(todayTasks),
			FirstStartTime:    firstStartTime(todaySessions),
			ProjectBreakdown:  buildProjectBreakdown(todayTasks),
			TimeDistribution:  buildTimeDistribution(todaySessions),
		},
		RecentContext: DailySummaryContext{
			RecentActiveDays: buildRecentActiveDays(recentDates, tasksByDate, sessionsByDate),
			ProjectPatterns:  buildProjectPatterns(contextDates, tasksByDate, sessionsByDate),
			RepeatedNotes:    extractRepeatedNotes(tasksForDates(contextDates, tasksByDate)),
		},
		RelevantMemories: []SummaryRelevantMemory{},
		Excluded:         excluded,
		Warnings:         buildSummaryWarnings(excluded),
	}

	return source
}

func buildWeeklySummarySourceData(weekStart, weekEnd string, weekDates []string, weekTasks []dailySummaryTaskRow, weekSessions []dailySummarySessionRow, previousWeekDates []string, previousWeekTasks []dailySummaryTaskRow, previousWeekSessions []dailySummarySessionRow) WeeklySummarySourceData {
	includedWeekTasks, includedWeekSessions, excluded := splitSummaryScope(weekTasks, weekSessions)
	includedPreviousWeekTasks, includedPreviousWeekSessions, _ := splitSummaryScope(previousWeekTasks, previousWeekSessions)
	tasksByDate := groupTasksByDate(includedWeekTasks)
	sessionsByDate := groupSessionsByDate(includedWeekSessions)
	previousTasksByDate := groupTasksByDate(includedPreviousWeekTasks)
	previousSessionsByDate := groupSessionsByDate(includedPreviousWeekSessions)
	daysWithData := countDatesWithData(weekDates, tasksByDate, sessionsByDate)
	previousDaysWithData := countDatesWithData(previousWeekDates, previousTasksByDate, previousSessionsByDate)

	return WeeklySummarySourceData{
		SummaryType: "weekly",
		WeekStart:   weekStart,
		WeekEnd:     weekEnd,
		DataQuality: WeeklyDataQuality{
			DaysWithData:    daysWithData,
			CanAnalyzeTrend: daysWithData >= 3,
			HasPreviousWeek: previousDaysWithData > 0,
		},
		Week: WeeklySummaryWeek{
			TotalFocusMinutes: totalTaskMinutes(includedWeekTasks),
			CompletedTasks:    countCompletedTasks(includedWeekTasks),
			TaskCount:         len(includedWeekTasks),
			DailyTotals:       buildWeeklyDailyTotals(weekDates, tasksByDate, sessionsByDate),
			ProjectBreakdown:  buildWeeklyProjectBreakdown(weekDates, tasksByDate),
			TimeDistribution:  buildTimeDistribution(includedWeekSessions),
			StartTimePatterns: buildWeeklyStartTimePatterns(weekDates, tasksByDate, sessionsByDate),
			RepeatedNotes:     extractRepeatedNotes(includedWeekTasks),
		},
		PreviousWeekComparison: buildPreviousWeekComparison(previousDaysWithData, includedPreviousWeekTasks),
		RelevantMemories:       []SummaryRelevantMemory{},
		Excluded:               excluded,
		Warnings:               buildSummaryWarnings(excluded),
	}
}

func splitSummaryScope(tasks []dailySummaryTaskRow, sessions []dailySummarySessionRow) ([]dailySummaryTaskRow, []dailySummarySessionRow, SummaryExcluded) {
	includedTasks := make([]dailySummaryTaskRow, 0, len(tasks))
	includedSessions := make([]dailySummarySessionRow, 0, len(sessions))
	excluded := SummaryExcluded{}
	excludedTotalSeconds := 0
	unassignedTotalSeconds := 0
	excludedSecondsByProject := make(map[string]int)
	excludedCategoryByProject := make(map[string]string)

	for _, task := range tasks {
		if isIncludedInSummary(task.ProjectID, task.IncludeInSummary) {
			includedTasks = append(includedTasks, task)
			continue
		}

		seconds := actualSeconds(task)
		excluded.ExcludedTaskCount++
		excludedTotalSeconds += seconds
		if !task.ProjectID.Valid {
			excluded.UnassignedTaskCount++
			unassignedTotalSeconds += seconds
			continue
		}
		excludedSecondsByProject[task.ProjectName] += seconds
		excludedCategoryByProject[task.ProjectName] = task.ProjectCategory
	}

	for _, session := range sessions {
		if isIncludedInSummary(session.ProjectID, session.IncludeInSummary) {
			includedSessions = append(includedSessions, session)
		}
	}

	excluded.ExcludedTotalMinutes = secondsToMinutes(excludedTotalSeconds)
	excluded.UnassignedTotalMinutes = secondsToMinutes(unassignedTotalSeconds)
	excluded.ExcludedProjects = buildExcludedProjects(excludedSecondsByProject, excludedCategoryByProject)
	return includedTasks, includedSessions, excluded
}

func isIncludedInSummary(projectID sql.NullInt64, includeInSummary bool) bool {
	return projectID.Valid && includeInSummary
}

func buildExcludedProjects(secondsByProject map[string]int, categoryByProject map[string]string) []SummaryExcludedProject {
	projects := make([]SummaryExcludedProject, 0, len(secondsByProject))
	for projectName, seconds := range secondsByProject {
		projects = append(projects, SummaryExcludedProject{
			ProjectName:  projectName,
			Category:     categoryByProject[projectName],
			TotalMinutes: secondsToMinutes(seconds),
		})
	}
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].TotalMinutes == projects[j].TotalMinutes {
			return projects[i].ProjectName < projects[j].ProjectName
		}
		return projects[i].TotalMinutes > projects[j].TotalMinutes
	})
	return projects
}

func buildSummaryWarnings(excluded SummaryExcluded) []string {
	warnings := make([]string, 0, 1)
	if excluded.UnassignedTaskCount > 0 {
		warnings = append(warnings, "存在未绑定项目的任务，已从学习总结中排除。")
	}
	return warnings
}

func groupTasksByDate(tasks []dailySummaryTaskRow) map[string][]dailySummaryTaskRow {
	result := make(map[string][]dailySummaryTaskRow)
	for _, task := range tasks {
		result[task.Date] = append(result[task.Date], task)
	}
	return result
}

func countDatesWithData(dates []string, tasksByDate map[string][]dailySummaryTaskRow, sessionsByDate map[string][]dailySummarySessionRow) int {
	count := 0
	for _, date := range dates {
		if len(tasksByDate[date]) > 0 || len(sessionsByDate[date]) > 0 {
			count++
		}
	}
	return count
}

func datesWithData(dates []string, tasksByDate map[string][]dailySummaryTaskRow, sessionsByDate map[string][]dailySummarySessionRow) []string {
	result := make([]string, 0, len(dates))
	for _, date := range dates {
		if len(tasksByDate[date]) > 0 || len(sessionsByDate[date]) > 0 {
			result = append(result, date)
		}
	}
	return result
}

func groupSessionsByDate(sessions []dailySummarySessionRow) map[string][]dailySummarySessionRow {
	result := make(map[string][]dailySummarySessionRow)
	for _, session := range sessions {
		result[session.Date] = append(result[session.Date], session)
	}
	return result
}

func tasksForDates(dates []string, tasksByDate map[string][]dailySummaryTaskRow) []dailySummaryTaskRow {
	result := make([]dailySummaryTaskRow, 0)
	for _, date := range dates {
		result = append(result, tasksByDate[date]...)
	}
	return result
}

func totalTaskMinutes(tasks []dailySummaryTaskRow) int {
	totalSeconds := 0
	for _, task := range tasks {
		totalSeconds += actualSeconds(task)
	}
	return secondsToMinutes(totalSeconds)
}

func countCompletedTasks(tasks []dailySummaryTaskRow) int {
	completed := 0
	for _, task := range tasks {
		if task.Status == "completed" {
			completed++
		}
	}
	return completed
}

func buildProjectBreakdown(tasks []dailySummaryTaskRow) []DailyProjectBreakdown {
	type projectAccumulator struct {
		item          DailyProjectBreakdown
		actualSeconds int
	}
	byProject := make(map[string]*projectAccumulator)
	for _, task := range tasks {
		acc := byProject[task.ProjectName]
		if acc == nil {
			acc = &projectAccumulator{item: DailyProjectBreakdown{ProjectName: task.ProjectName}}
			byProject[task.ProjectName] = acc
		}
		item := &acc.item
		item.TaskCount++
		if task.Status == "completed" {
			item.CompletedCount++
		}
		item.EstimatedMinutes += task.EstimatedMinutes
		acc.actualSeconds += actualSeconds(task)
	}

	result := make([]DailyProjectBreakdown, 0, len(byProject))
	for _, acc := range byProject {
		acc.item.ActualMinutes = secondsToMinutes(acc.actualSeconds)
		acc.item.TotalMinutes = acc.item.ActualMinutes
		acc.item.OverrunMinutes = acc.item.ActualMinutes - acc.item.EstimatedMinutes
		result = append(result, acc.item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalMinutes == result[j].TotalMinutes {
			return result[i].ProjectName < result[j].ProjectName
		}
		return result[i].TotalMinutes > result[j].TotalMinutes
	})
	return result
}

func buildTimeDistribution(sessions []dailySummarySessionRow) DailyTimeDistribution {
	var morningSeconds, afternoonSeconds, eveningSeconds, nightSeconds int
	for _, session := range sessions {
		hour := session.StartedAt.Hour()
		// First version simplification: assign the whole session to the period where started_at falls,
		// without splitting sessions that cross period boundaries.
		switch {
		case hour >= 6 && hour < 12:
			morningSeconds += session.DurationSeconds
		case hour >= 12 && hour < 18:
			afternoonSeconds += session.DurationSeconds
		case hour >= 18:
			eveningSeconds += session.DurationSeconds
		default:
			nightSeconds += session.DurationSeconds
		}
	}
	return DailyTimeDistribution{
		MorningMinutes:   secondsToMinutes(morningSeconds),
		AfternoonMinutes: secondsToMinutes(afternoonSeconds),
		EveningMinutes:   secondsToMinutes(eveningSeconds),
		NightMinutes:     secondsToMinutes(nightSeconds),
	}
}

func buildRecentActiveDays(recentDates []string, tasksByDate map[string][]dailySummaryTaskRow, sessionsByDate map[string][]dailySummarySessionRow) []DailyRecentActiveDay {
	result := make([]DailyRecentActiveDay, 0, len(recentDates))
	for _, date := range recentDates {
		tasks := tasksByDate[date]
		sessions := sessionsByDate[date]
		result = append(result, DailyRecentActiveDay{
			Date:              date,
			TotalFocusMinutes: totalTaskMinutes(tasks),
			FirstStartTime:    firstStartTime(sessions),
			MainProject:       mainProject(tasks),
		})
	}
	return result
}

func mainProject(tasks []dailySummaryTaskRow) string {
	secondsByProject := make(map[string]int)
	for _, task := range tasks {
		secondsByProject[task.ProjectName] += actualSeconds(task)
	}

	bestProject := ""
	bestSeconds := -1
	for project, seconds := range secondsByProject {
		if seconds > bestSeconds || (seconds == bestSeconds && (bestProject == "" || project < bestProject)) {
			bestProject = project
			bestSeconds = seconds
		}
	}
	return bestProject
}

func firstStartTime(sessions []dailySummarySessionRow) string {
	if len(sessions) == 0 {
		return ""
	}
	first := sessions[0].StartedAt
	for _, session := range sessions[1:] {
		if session.StartedAt.Before(first) {
			first = session.StartedAt
		}
	}
	return first.Format("15:04")
}

func buildProjectPatterns(dates []string, tasksByDate map[string][]dailySummaryTaskRow, sessionsByDate map[string][]dailySummarySessionRow) []DailyProjectPattern {
	type projectDay struct {
		actualSeconds    int
		estimatedMinutes int
		firstStartMinute *int
	}
	type projectAccumulator struct {
		days           map[string]*projectDay
		overrunTasks   int
		estimatedTasks int
	}

	projects := make(map[string]*projectAccumulator)
	for _, date := range dates {
		for _, task := range tasksByDate[date] {
			acc := projects[task.ProjectName]
			if acc == nil {
				acc = &projectAccumulator{days: make(map[string]*projectDay)}
				projects[task.ProjectName] = acc
			}
			day := acc.days[date]
			if day == nil {
				day = &projectDay{}
				acc.days[date] = day
			}
			actualMinutes := secondsToMinutes(actualSeconds(task))
			day.actualSeconds += actualSeconds(task)
			day.estimatedMinutes += task.EstimatedMinutes
			if task.EstimatedMinutes > 0 {
				acc.estimatedTasks++
				if actualMinutes > task.EstimatedMinutes {
					acc.overrunTasks++
				}
			}
		}
	}

	for _, date := range dates {
		for _, session := range sessionsByDate[date] {
			acc := projects[session.ProjectName]
			if acc == nil {
				continue
			}
			day := acc.days[date]
			if day == nil {
				continue
			}
			minute := session.StartedAt.Hour()*60 + session.StartedAt.Minute()
			if day.firstStartMinute == nil || minute < *day.firstStartMinute {
				copied := minute
				day.firstStartMinute = &copied
			}
		}
	}

	result := make([]DailyProjectPattern, 0, len(projects))
	for project, acc := range projects {
		activeDays := 0
		totalActualSeconds := 0
		totalEstimatedMinutes := 0
		totalStartMinutes := 0
		startDays := 0
		for _, day := range acc.days {
			if day.actualSeconds > 0 || day.estimatedMinutes > 0 || day.firstStartMinute != nil {
				activeDays++
			}
			totalActualSeconds += day.actualSeconds
			totalEstimatedMinutes += day.estimatedMinutes
			if day.firstStartMinute != nil {
				totalStartMinutes += *day.firstStartMinute
				startDays++
			}
		}
		if activeDays == 0 {
			continue
		}

		overrunRate := 0.0
		if acc.estimatedTasks > 0 {
			overrunRate = math.Round((float64(acc.overrunTasks)/float64(acc.estimatedTasks))*100) / 100
		}
		pattern := DailyProjectPattern{
			ProjectName:         project,
			ActiveDays:          activeDays,
			AvgActualMinutes:    secondsToMinutes(totalActualSeconds) / activeDays,
			AvgEstimatedMinutes: totalEstimatedMinutes / activeDays,
			OverrunRate:         overrunRate,
		}
		if startDays > 0 {
			pattern.AvgStartTime = minutesToClock(totalStartMinutes / startDays)
		}
		result = append(result, pattern)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].ActiveDays == result[j].ActiveDays {
			return result[i].ProjectName < result[j].ProjectName
		}
		return result[i].ActiveDays > result[j].ActiveDays
	})
	return result
}

func extractRepeatedNotes(tasks []dailySummaryTaskRow) []string {
	counts := make(map[string]int)
	splitter := regexp.MustCompile(`[，,。.;；:：、\s\(\)（）\[\]【】{}"'“”‘’!?！？/\\|]+`)
	pureNumber := regexp.MustCompile(`^[0-9]+$`)
	for _, task := range tasks {
		text := strings.TrimSpace(task.FinishNote + " " + task.FinishDescription)
		for _, token := range splitter.Split(text, -1) {
			token = strings.TrimSpace(token)
			if utf8.RuneCountInString(token) < 2 {
				continue
			}
			if pureNumber.MatchString(token) || isMeaninglessNoteToken(token) {
				continue
			}
			counts[token]++
		}
	}

	type noteCount struct {
		token string
		count int
	}
	repeated := make([]noteCount, 0)
	for token, count := range counts {
		if count > 1 {
			repeated = append(repeated, noteCount{token: token, count: count})
		}
	}
	sort.Slice(repeated, func(i, j int) bool {
		if repeated[i].count == repeated[j].count {
			return repeated[i].token < repeated[j].token
		}
		return repeated[i].count > repeated[j].count
	})

	limit := 8
	if len(repeated) < limit {
		limit = len(repeated)
	}
	result := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, repeated[i].token)
	}
	return result
}

func isMeaninglessNoteToken(token string) bool {
	meaningless := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "by": true, "did": true, "do": true, "does": true, "done": true,
		"false": true, "for": true, "from": true, "had": true, "has": true, "have": true,
		"in": true, "is": true, "it": true, "no": true, "nil": true, "null": true,
		"of": true, "on": true, "or": true, "that": true, "the": true, "this": true,
		"to": true, "todo": true, "true": true, "was": true, "were": true, "with": true,
		"yes": true,
		"一个":  true, "以及": true, "但是": true, "并且": true, "然后": true,
		"任务": true, "今天": true, "这个": true, "那个": true, "项目": true,
	}
	return meaningless[strings.ToLower(token)]
}

func buildWeeklyDailyTotals(dates []string, tasksByDate map[string][]dailySummaryTaskRow, sessionsByDate map[string][]dailySummarySessionRow) []WeeklyDailyTotal {
	result := make([]WeeklyDailyTotal, 0)
	for _, date := range dates {
		tasks := tasksByDate[date]
		sessions := sessionsByDate[date]
		if len(tasks) == 0 && len(sessions) == 0 {
			continue
		}
		result = append(result, WeeklyDailyTotal{
			Date:              date,
			TotalFocusMinutes: totalTaskMinutes(tasks),
			CompletedTasks:    countCompletedTasks(tasks),
			TaskCount:         len(tasks),
			FirstStartTime:    firstStartTime(sessions),
			MainProject:       mainProject(tasks),
		})
	}
	return result
}

func buildWeeklyProjectBreakdown(dates []string, tasksByDate map[string][]dailySummaryTaskRow) []WeeklyProjectBreakdown {
	type projectAccumulator struct {
		item           WeeklyProjectBreakdown
		activeDates    map[string]bool
		actualSeconds  int
		overrunTasks   int
		estimatedTasks int
	}

	projects := make(map[string]*projectAccumulator)
	for _, date := range dates {
		for _, task := range tasksByDate[date] {
			acc := projects[task.ProjectName]
			if acc == nil {
				acc = &projectAccumulator{
					item:        WeeklyProjectBreakdown{ProjectName: task.ProjectName},
					activeDates: make(map[string]bool),
				}
				projects[task.ProjectName] = acc
			}
			acc.activeDates[date] = true
			acc.item.TaskCount++
			if task.Status == "completed" {
				acc.item.CompletedCount++
			}
			acc.item.EstimatedMinutes += task.EstimatedMinutes
			actualTaskSeconds := actualSeconds(task)
			acc.actualSeconds += actualTaskSeconds
			if task.EstimatedMinutes > 0 {
				acc.estimatedTasks++
				if secondsToMinutes(actualTaskSeconds) > task.EstimatedMinutes {
					acc.overrunTasks++
				}
			}
		}
	}

	result := make([]WeeklyProjectBreakdown, 0, len(projects))
	for _, acc := range projects {
		acc.item.ActiveDays = len(acc.activeDates)
		acc.item.ActualMinutes = secondsToMinutes(acc.actualSeconds)
		acc.item.TotalMinutes = acc.item.ActualMinutes
		acc.item.OverrunMinutes = acc.item.ActualMinutes - acc.item.EstimatedMinutes
		if acc.estimatedTasks > 0 {
			acc.item.OverrunRate = math.Round((float64(acc.overrunTasks)/float64(acc.estimatedTasks))*100) / 100
		}
		result = append(result, acc.item)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalMinutes == result[j].TotalMinutes {
			return result[i].ProjectName < result[j].ProjectName
		}
		return result[i].TotalMinutes > result[j].TotalMinutes
	})
	return result
}

func buildWeeklyStartTimePatterns(dates []string, tasksByDate map[string][]dailySummaryTaskRow, sessionsByDate map[string][]dailySummarySessionRow) []WeeklyStartTimePattern {
	type patternAccumulator struct {
		firstStartMinutes []int
		activeDates       map[string]bool
	}

	projects := make(map[string]*patternAccumulator)
	for _, date := range dates {
		projectHasTask := make(map[string]bool)
		for _, task := range tasksByDate[date] {
			projectHasTask[task.ProjectName] = true
		}

		earliestByProject := make(map[string]int)
		for _, session := range sessionsByDate[date] {
			if !projectHasTask[session.ProjectName] {
				continue
			}
			minute := session.StartedAt.Hour()*60 + session.StartedAt.Minute()
			current, ok := earliestByProject[session.ProjectName]
			if !ok || minute < current {
				earliestByProject[session.ProjectName] = minute
			}
		}

		for project := range projectHasTask {
			acc := projects[project]
			if acc == nil {
				acc = &patternAccumulator{activeDates: make(map[string]bool)}
				projects[project] = acc
			}
			acc.activeDates[date] = true
			if minute, ok := earliestByProject[project]; ok {
				acc.firstStartMinutes = append(acc.firstStartMinutes, minute)
			}
		}
	}

	result := make([]WeeklyStartTimePattern, 0, len(projects))
	for project, acc := range projects {
		firstStartTimes := make([]string, 0, len(acc.firstStartMinutes))
		totalMinutes := 0
		for _, minute := range acc.firstStartMinutes {
			firstStartTimes = append(firstStartTimes, minutesToClock(minute))
			totalMinutes += minute
		}
		pattern := WeeklyStartTimePattern{
			ProjectName:     project,
			ActiveDays:      len(acc.activeDates),
			FirstStartTimes: firstStartTimes,
		}
		if len(acc.firstStartMinutes) > 0 {
			pattern.AvgStartTime = minutesToClock(totalMinutes / len(acc.firstStartMinutes))
		}
		result = append(result, pattern)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].ActiveDays == result[j].ActiveDays {
			return result[i].ProjectName < result[j].ProjectName
		}
		return result[i].ActiveDays > result[j].ActiveDays
	})
	return result
}

func buildPreviousWeekComparison(daysWithData int, tasks []dailySummaryTaskRow) PreviousWeekComparison {
	comparison := PreviousWeekComparison{
		Available:         daysWithData > 0,
		TotalFocusMinutes: totalTaskMinutes(tasks),
		DaysWithData:      daysWithData,
		MainProjects:      buildPreviousWeekMainProjects(tasks),
	}
	return comparison
}

func buildPreviousWeekMainProjects(tasks []dailySummaryTaskRow) []PreviousWeekMainProject {
	secondsByProject := make(map[string]int)
	for _, task := range tasks {
		secondsByProject[task.ProjectName] += actualSeconds(task)
	}

	result := make([]PreviousWeekMainProject, 0, len(secondsByProject))
	for project, seconds := range secondsByProject {
		result = append(result, PreviousWeekMainProject{
			ProjectName:  project,
			TotalMinutes: secondsToMinutes(seconds),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalMinutes == result[j].TotalMinutes {
			return result[i].ProjectName < result[j].ProjectName
		}
		return result[i].TotalMinutes > result[j].TotalMinutes
	})
	if len(result) > 5 {
		result = result[:5]
	}
	return result
}

func actualSeconds(task dailySummaryTaskRow) int {
	if task.ActualSecondsOverride.Valid {
		return int(task.ActualSecondsOverride.Int64)
	}
	return task.SessionSeconds
}

func secondsToMinutes(seconds int) int {
	return seconds / 60
}

func minutesToClock(minutes int) string {
	hour := minutes / 60
	minute := minutes % 60
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

func dateRange(startDate, endDate string) ([]string, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end date is before start date")
	}
	return dateRangeFromTimes(start, end), nil
}

func previousWeekDateRange(weekStart string) ([]string, error) {
	start, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, err
	}
	return dateRangeFromTimes(start.AddDate(0, 0, -7), start.AddDate(0, 0, -1)), nil
}

func dateRangeFromTimes(start, end time.Time) []string {
	dates := make([]string, 0)
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		dates = append(dates, current.Format("2006-01-02"))
	}
	return dates
}
