package summaries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"personal/internal/llm"
	"personal/internal/stats"
)

const DailyRecentActiveDaysLimit = 5

var (
	ErrStatsQueryFailed     = errors.New("stats query failed")
	ErrLLMGenerationFailed  = errors.New("LLM generation failed")
	ErrSummaryPersistFailed = errors.New("summary persistence failed")
	ErrSummaryAlreadyExists = errors.New("summary already exists")
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
}

type weeklyStatsProvider interface {
	GetWeeklyStats(startDate, endDate string) (*stats.WeeklyStats, error)
}

type Service struct {
	repo         summaryRepository
	statsService weeklyStatsProvider
	llmClient    llm.Client
}

func NewService(repo *Repository, statsService *stats.Service, llmClient llm.Client) *Service {
	return &Service{
		repo:         repo,
		statsService: statsService,
		llmClient:    llmClient,
	}
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
	sourceData, err := json.Marshal(dailyContext)
	if err != nil {
		return nil, err
	}

	content, err := s.llmClient.GenerateSummary(ctx, buildDailyPrompt(string(sourceData)))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMGenerationFailed, err)
	}

	id, err := s.repo.CreateSummary(ctx, CreateSummaryInput{
		SummaryType: "daily",
		StartDate:   date,
		EndDate:     date,
		Content:     content,
		SourceData:  sourceData,
	})
	if err != nil {
		if errors.Is(err, ErrSummaryAlreadyExists) {
			return nil, ErrSummaryAlreadyExists
		}
		return nil, fmt.Errorf("%w: %v", ErrSummaryPersistFailed, err)
	}

	return &GenerateSummaryResult{SummaryID: id, Content: content}, nil
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
	sourceData, err := json.Marshal(weeklyContext)
	if err != nil {
		return nil, err
	}

	content, err := s.llmClient.GenerateSummary(ctx, buildWeeklyPrompt(string(sourceData)))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMGenerationFailed, err)
	}

	id, err := s.repo.CreateSummary(ctx, CreateSummaryInput{
		SummaryType: "weekly",
		StartDate:   startDate,
		EndDate:     endDate,
		Content:     content,
		SourceData:  sourceData,
	})
	if err != nil {
		if errors.Is(err, ErrSummaryAlreadyExists) {
			return nil, ErrSummaryAlreadyExists
		}
		return nil, fmt.Errorf("%w: %v", ErrSummaryPersistFailed, err)
	}

	return &GenerateSummaryResult{SummaryID: id, Content: content}, nil
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

func buildDailyPrompt(sourceData string) string {
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

func buildDailySummarySourceData(targetDate string, recentDates []string, tasks []dailySummaryTaskRow, sessions []dailySummarySessionRow) DailySummarySourceData {
	contextDates := append([]string{targetDate}, recentDates...)
	tasksByDate := groupTasksByDate(tasks)
	sessionsByDate := groupSessionsByDate(sessions)
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
	}

	return source
}

func buildWeeklySummarySourceData(weekStart, weekEnd string, weekDates []string, weekTasks []dailySummaryTaskRow, weekSessions []dailySummarySessionRow, previousWeekDates []string, previousWeekTasks []dailySummaryTaskRow, previousWeekSessions []dailySummarySessionRow) WeeklySummarySourceData {
	tasksByDate := groupTasksByDate(weekTasks)
	sessionsByDate := groupSessionsByDate(weekSessions)
	previousTasksByDate := groupTasksByDate(previousWeekTasks)
	previousSessionsByDate := groupSessionsByDate(previousWeekSessions)
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
			TotalFocusMinutes: totalTaskMinutes(weekTasks),
			CompletedTasks:    countCompletedTasks(weekTasks),
			TaskCount:         len(weekTasks),
			DailyTotals:       buildWeeklyDailyTotals(weekDates, tasksByDate, sessionsByDate),
			ProjectBreakdown:  buildWeeklyProjectBreakdown(weekDates, tasksByDate),
			TimeDistribution:  buildTimeDistribution(weekSessions),
			StartTimePatterns: buildWeeklyStartTimePatterns(weekDates, tasksByDate, sessionsByDate),
			RepeatedNotes:     extractRepeatedNotes(weekTasks),
		},
		PreviousWeekComparison: buildPreviousWeekComparison(previousDaysWithData, previousWeekTasks),
	}
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
