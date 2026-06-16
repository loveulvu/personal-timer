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

	weeklyStats, err := s.statsService.GetWeeklyStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatsQueryFailed, err)
	}

	sourceData, err := json.Marshal(weeklyStats)
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
说明今日总时长、完成任务数、主要项目。

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
	return `Generate a concise weekly study review from the JSON statistics below.
Focus on total time, daily fluctuations, project time distribution, completion,
main problems, and practical suggestions for next week. Do not praise or use motivational language.

Statistics:
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

func groupTasksByDate(tasks []dailySummaryTaskRow) map[string][]dailySummaryTaskRow {
	result := make(map[string][]dailySummaryTaskRow)
	for _, task := range tasks {
		result[task.Date] = append(result[task.Date], task)
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
	byProject := make(map[string]*DailyProjectBreakdown)
	for _, task := range tasks {
		item := byProject[task.ProjectName]
		if item == nil {
			item = &DailyProjectBreakdown{ProjectName: task.ProjectName}
			byProject[task.ProjectName] = item
		}
		item.TaskCount++
		if task.Status == "completed" {
			item.CompletedCount++
		}
		item.EstimatedMinutes += task.EstimatedMinutes
		item.ActualMinutes += secondsToMinutes(actualSeconds(task))
	}

	result := make([]DailyProjectBreakdown, 0, len(byProject))
	for _, item := range byProject {
		item.TotalMinutes = item.ActualMinutes
		item.OverrunMinutes = item.ActualMinutes - item.EstimatedMinutes
		result = append(result, *item)
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
	var result DailyTimeDistribution
	for _, session := range sessions {
		minutes := secondsToMinutes(session.DurationSeconds)
		hour := session.StartedAt.Hour()
		// First version simplification: assign the whole session to the period where started_at falls,
		// without splitting sessions that cross period boundaries.
		switch {
		case hour >= 6 && hour < 12:
			result.MorningMinutes += minutes
		case hour >= 12 && hour < 18:
			result.AfternoonMinutes += minutes
		case hour >= 18:
			result.EveningMinutes += minutes
		default:
			result.NightMinutes += minutes
		}
	}
	return result
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
		totalActualMinutes := 0
		totalEstimatedMinutes := 0
		totalStartMinutes := 0
		startDays := 0
		for _, day := range acc.days {
			if day.actualSeconds > 0 || day.estimatedMinutes > 0 || day.firstStartMinute != nil {
				activeDays++
			}
			totalActualMinutes += secondsToMinutes(day.actualSeconds)
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
			AvgActualMinutes:    totalActualMinutes / activeDays,
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
	for _, task := range tasks {
		text := strings.TrimSpace(task.FinishNote + " " + task.FinishDescription)
		for _, token := range splitter.Split(text, -1) {
			token = strings.TrimSpace(token)
			if utf8.RuneCountInString(token) < 2 {
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

	limit := 5
	if len(repeated) < limit {
		limit = len(repeated)
	}
	result := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, repeated[i].token)
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
