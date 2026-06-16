package summaries

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"personal/internal/stats"
)

func TestDailySummarySourceDataSingleDayDoesNotAnalyzeTrend(t *testing.T) {
	source := buildDailySummarySourceData(
		"2026-06-18",
		nil,
		[]dailySummaryTaskRow{
			taskRow("2026-06-18", "Personal Timer", "completed", 60, 3600, nil, "", ""),
		},
		nil,
	)

	if source.DataQuality.DaysWithData != 1 {
		t.Fatalf("days_with_data = %d, want 1", source.DataQuality.DaysWithData)
	}
	if source.DataQuality.CanAnalyzeTrend {
		t.Fatal("can_analyze_trend = true, want false")
	}
	if source.DataQuality.ComparisonDaysWithData != 0 {
		t.Fatalf("comparison_days_with_data = %d, want 0", source.DataQuality.ComparisonDaysWithData)
	}

	prompt := buildDailyPrompt(`{"data_quality":{"days_with_data":1,"can_analyze_trend":false}}`)
	required := []string{
		"你必须使用中文输出",
		"不能编造输入数据中不存在的项目、任务、时间、备注或趋势",
		"不要写空泛鼓励",
		"不要只复述每个项目花了多少时间",
		"days_with_data < 2：不要声称存在趋势",
		"days_with_data >= 2：可以描述初步变化，但要说明样本有限",
		"days_with_data >= 3：可以分析短期模式",
		"当前数据不足以判断",
		"# 每日学习总结",
		"## 1. 今日概览",
		"## 2. 时间分布",
		"## 3. 项目推进",
		"## 4. 与近期记录的对比",
		"## 5. 发现的问题",
		"## 6. 明日建议",
	}
	for _, text := range required {
		if !strings.Contains(prompt, text) {
			t.Fatalf("daily prompt is missing %q", text)
		}
	}
}

func TestDailySummaryUsesRecentActiveDatesInsteadOfNaturalWindow(t *testing.T) {
	recentDates := []string{"2026-06-15", "2026-06-10", "2026-06-01"}
	source := buildDailySummarySourceData(
		"2026-06-18",
		recentDates,
		[]dailySummaryTaskRow{
			taskRow("2026-06-18", "A", "completed", 30, 1800, nil, "", ""),
			taskRow("2026-06-15", "A", "completed", 30, 1800, nil, "", ""),
			taskRow("2026-06-10", "B", "completed", 20, 1200, nil, "", ""),
			taskRow("2026-06-01", "C", "completed", 10, 600, nil, "", ""),
		},
		nil,
	)

	got := make([]string, 0, len(source.RecentContext.RecentActiveDays))
	for _, day := range source.RecentContext.RecentActiveDays {
		got = append(got, day.Date)
	}
	if !reflect.DeepEqual(got, recentDates) {
		t.Fatalf("recent_active_days dates = %v, want %v", got, recentDates)
	}
	for _, gotDate := range got {
		if gotDate == source.TargetDate {
			t.Fatalf("recent_active_days must not include target date %s", source.TargetDate)
		}
	}
	if source.DataQuality.ComparisonDaysWithData != len(recentDates) {
		t.Fatalf("comparison_days_with_data = %d, want %d", source.DataQuality.ComparisonDaysWithData, len(recentDates))
	}
}

func TestDailySummaryActualSecondsOverrideTakesPrecedence(t *testing.T) {
	overrideSeconds := 7200
	source := buildDailySummarySourceData(
		"2026-06-18",
		nil,
		[]dailySummaryTaskRow{
			taskRow("2026-06-18", "Override", "completed", 60, 1800, &overrideSeconds, "", ""),
			taskRow("2026-06-18", "Session", "completed", 20, 1800, nil, "", ""),
		},
		nil,
	)

	if source.Today.TotalFocusMinutes != 150 {
		t.Fatalf("total_focus_minutes = %d, want 150", source.Today.TotalFocusMinutes)
	}

	byProject := map[string]DailyProjectBreakdown{}
	for _, project := range source.Today.ProjectBreakdown {
		byProject[project.ProjectName] = project
	}
	if byProject["Override"].ActualMinutes != 120 {
		t.Fatalf("override project actual_minutes = %d, want 120", byProject["Override"].ActualMinutes)
	}
	if byProject["Session"].ActualMinutes != 30 {
		t.Fatalf("session project actual_minutes = %d, want 30", byProject["Session"].ActualMinutes)
	}

	byPattern := map[string]DailyProjectPattern{}
	for _, pattern := range source.RecentContext.ProjectPatterns {
		byPattern[pattern.ProjectName] = pattern
	}
	if byPattern["Override"].AvgActualMinutes != 120 {
		t.Fatalf("override project pattern avg_actual_minutes = %d, want 120", byPattern["Override"].AvgActualMinutes)
	}
	if byPattern["Session"].AvgActualMinutes != 30 {
		t.Fatalf("session project pattern avg_actual_minutes = %d, want 30", byPattern["Session"].AvgActualMinutes)
	}
}

func TestDailySummaryTimeDistributionUsesSessionStartPeriod(t *testing.T) {
	source := buildDailySummarySourceData(
		"2026-06-18",
		nil,
		nil,
		[]dailySummarySessionRow{
			sessionRow("2026-06-18", "Personal Timer", "2026-06-18T05:59:00", 600),
			sessionRow("2026-06-18", "Personal Timer", "2026-06-18T06:00:00", 1200),
			sessionRow("2026-06-18", "Personal Timer", "2026-06-18T15:30:00", 2700),
			sessionRow("2026-06-18", "Personal Timer", "2026-06-18T18:00:00", 1800),
		},
	)

	if source.Today.TimeDistribution.AfternoonMinutes != 45 {
		t.Fatalf("afternoon_minutes = %d, want 45", source.Today.TimeDistribution.AfternoonMinutes)
	}
	if source.Today.TimeDistribution.MorningMinutes != 20 {
		t.Fatalf("morning_minutes = %d, want 20", source.Today.TimeDistribution.MorningMinutes)
	}
	if source.Today.TimeDistribution.EveningMinutes != 30 {
		t.Fatalf("evening_minutes = %d, want 30", source.Today.TimeDistribution.EveningMinutes)
	}
	if source.Today.TimeDistribution.NightMinutes != 10 {
		t.Fatalf("night_minutes = %d, want 10", source.Today.TimeDistribution.NightMinutes)
	}
	if source.Today.FirstStartTime != "05:59" {
		t.Fatalf("today.first_start_time = %q, want 05:59", source.Today.FirstStartTime)
	}
}

func TestTimeDistributionAggregatesSecondsBeforeConvertingToMinutes(t *testing.T) {
	distribution := buildTimeDistribution([]dailySummarySessionRow{
		sessionRow("2026-06-18", "Personal Timer", "2026-06-18T15:00:00", 90),
		sessionRow("2026-06-18", "Personal Timer", "2026-06-18T16:00:00", 90),
	})

	if distribution.AfternoonMinutes != 3 {
		t.Fatalf("afternoon_minutes = %d, want 3 after summing 180 seconds", distribution.AfternoonMinutes)
	}
}

func TestDailySummaryRepeatedNotesUsesFinishTextAndFiltersShortAndNumericTokens(t *testing.T) {
	source := buildDailySummarySourceData(
		"2026-06-18",
		[]string{"2026-06-15"},
		[]dailySummaryTaskRow{
			taskRow("2026-06-18", "A", "completed", 30, 1800, nil, "接口联调 a 10", "前端状态刷新 map slice channel context"),
			taskRow("2026-06-15", "A", "completed", 30, 1800, nil, "接口联调 10", "前端状态刷新 map slice channel context"),
		},
		nil,
	)

	got := strings.Join(source.RecentContext.RepeatedNotes, ",")
	if !strings.Contains(got, "接口联调") || !strings.Contains(got, "前端状态刷新") {
		t.Fatalf("repeated_notes = %v, want repeated tokens from finish_note and finish_description", source.RecentContext.RepeatedNotes)
	}
	for _, token := range []string{"map", "slice", "channel", "context"} {
		if !strings.Contains(got, token) {
			t.Fatalf("repeated_notes = %v, should keep technical token %q", source.RecentContext.RepeatedNotes, token)
		}
	}
	for _, token := range source.RecentContext.RepeatedNotes {
		if token == "a" || token == "10" {
			t.Fatalf("repeated_notes = %v, should filter short and numeric tokens", source.RecentContext.RepeatedNotes)
		}
	}
}

func TestWeeklySummarySourceDataBasicStructureAndAggregation(t *testing.T) {
	overrideSeconds := 3600
	weekDates := mustDateRange("2026-06-15", "2026-06-21")
	previousWeekDates := mustDateRange("2026-06-08", "2026-06-14")
	source := buildWeeklySummarySourceData(
		"2026-06-15",
		"2026-06-21",
		weekDates,
		[]dailySummaryTaskRow{
			taskRow("2026-06-15", "Personal Timer", "completed", 60, 4200, nil, "接口联调", "前端状态刷新"),
			taskRow("2026-06-16", "Go", "completed", 30, 600, &overrideSeconds, "map slice", "channel context"),
			taskRow("2026-06-18", "Personal Timer", "planned", 45, 1800, nil, "接口联调", "前端状态刷新"),
		},
		[]dailySummarySessionRow{
			sessionRow("2026-06-15", "Personal Timer", "2026-06-15T15:20:00", 4200),
			sessionRow("2026-06-16", "Go", "2026-06-16T14:10:00", 600),
			sessionRow("2026-06-18", "Personal Timer", "2026-06-18T19:00:00", 1800),
		},
		previousWeekDates,
		[]dailySummaryTaskRow{
			taskRow("2026-06-10", "Personal Timer", "completed", 60, 3000, nil, "", ""),
		},
		[]dailySummarySessionRow{
			sessionRow("2026-06-10", "Personal Timer", "2026-06-10T16:00:00", 3000),
		},
	)

	if source.SummaryType != "weekly" || source.WeekStart != "2026-06-15" || source.WeekEnd != "2026-06-21" {
		t.Fatalf("unexpected weekly identity fields: %+v", source)
	}
	if source.DataQuality.DaysWithData != 3 || !source.DataQuality.CanAnalyzeTrend || !source.DataQuality.HasPreviousWeek {
		t.Fatalf("unexpected data quality: %+v", source.DataQuality)
	}
	if len(source.Week.DailyTotals) != 3 {
		t.Fatalf("daily_totals length = %d, want only 3 active dates", len(source.Week.DailyTotals))
	}
	if source.Week.DailyTotals[0].Date != "2026-06-15" ||
		source.Week.DailyTotals[0].FirstStartTime != "15:20" ||
		source.Week.DailyTotals[0].MainProject != "Personal Timer" {
		t.Fatalf("unexpected first daily total: %+v", source.Week.DailyTotals[0])
	}
	if source.Week.TimeDistribution.AfternoonMinutes != 80 {
		t.Fatalf("afternoon_minutes = %d, want 80", source.Week.TimeDistribution.AfternoonMinutes)
	}
	if source.Week.TimeDistribution.EveningMinutes != 30 {
		t.Fatalf("evening_minutes = %d, want 30", source.Week.TimeDistribution.EveningMinutes)
	}

	projectByName := map[string]WeeklyProjectBreakdown{}
	for _, project := range source.Week.ProjectBreakdown {
		projectByName[project.ProjectName] = project
	}
	if projectByName["Go"].ActualMinutes != 60 || projectByName["Go"].TotalMinutes != 60 {
		t.Fatalf("Go actual minutes = %+v, want override value 60", projectByName["Go"])
	}
	if projectByName["Personal Timer"].ActiveDays != 2 {
		t.Fatalf("Personal Timer active_days = %d, want 2", projectByName["Personal Timer"].ActiveDays)
	}

	patternByName := map[string]WeeklyStartTimePattern{}
	for _, pattern := range source.Week.StartTimePatterns {
		patternByName[pattern.ProjectName] = pattern
	}
	if patternByName["Personal Timer"].AvgStartTime != "17:10" {
		t.Fatalf("Personal Timer avg_start_time = %q, want 17:10", patternByName["Personal Timer"].AvgStartTime)
	}
	if got := strings.Join(source.Week.RepeatedNotes, ","); !strings.Contains(got, "接口联调") || !strings.Contains(got, "前端状态刷新") {
		t.Fatalf("weekly repeated_notes = %v, want repeated finish text", source.Week.RepeatedNotes)
	}
	if !source.PreviousWeekComparison.Available ||
		source.PreviousWeekComparison.DaysWithData != 1 ||
		source.PreviousWeekComparison.TotalFocusMinutes != 50 {
		t.Fatalf("unexpected previous week comparison: %+v", source.PreviousWeekComparison)
	}
}

func TestWeeklySummaryPreviousWeekComparisonUnavailable(t *testing.T) {
	source := buildWeeklySummarySourceData(
		"2026-06-15",
		"2026-06-21",
		mustDateRange("2026-06-15", "2026-06-21"),
		[]dailySummaryTaskRow{
			taskRow("2026-06-15", "Personal Timer", "completed", 60, 4200, nil, "", ""),
		},
		nil,
		mustDateRange("2026-06-08", "2026-06-14"),
		nil,
		nil,
	)

	if source.PreviousWeekComparison.Available {
		t.Fatalf("previous_week_comparison.available = true, want false")
	}
	if source.DataQuality.HasPreviousWeek {
		t.Fatalf("data_quality.has_previous_week = true, want false")
	}
}

func TestWeeklyPromptContainsRequiredChineseStructureAndRules(t *testing.T) {
	prompt := buildWeeklyPrompt(`{"summary_type":"weekly"}`)
	required := []string{
		"你必须使用中文输出",
		"不能编造输入数据中不存在的项目、任务、时间、备注或趋势",
		"不要写空泛鼓励",
		"不要只复述每个项目花了多少时间",
		"days_with_data < 2：不要声称存在周趋势",
		"days_with_data >= 2：可以描述初步变化，但要说明样本有限",
		"days_with_data >= 3：可以分析本周模式",
		"previous_week_comparison.available = false 时，不要做上周对比",
		"当前数据不足以判断",
		"# 每周学习总结",
		"## 1. 本周总览",
		"## 2. 项目推进",
		"## 3. 时间段分布",
		"## 4. 开始时间模式",
		"## 5. 预计与实际偏差",
		"## 6. 重复问题",
		"## 7. 与上一周对比",
		"## 8. 下周调整建议",
	}
	for _, text := range required {
		if !strings.Contains(prompt, text) {
			t.Fatalf("weekly prompt is missing %q", text)
		}
	}
}

func TestGenerateDailySummaryPersistsStructuredSourceData(t *testing.T) {
	repo := &fakeSummaryRepo{
		recentDates: []string{"2026-06-15", "2026-06-10"},
		tasks: []dailySummaryTaskRow{
			taskRow("2026-06-18", "Personal Timer", "completed", 60, 4200, nil, "接口联调", "前端状态刷新"),
			taskRow("2026-06-15", "Personal Timer", "completed", 45, 3900, nil, "接口联调", ""),
			taskRow("2026-06-10", "Other", "completed", 30, 1800, nil, "前端状态刷新", ""),
		},
		sessions: []dailySummarySessionRow{
			sessionRow("2026-06-18", "Personal Timer", "2026-06-18T15:20:00", 4200),
			sessionRow("2026-06-15", "Personal Timer", "2026-06-15T15:40:00", 3900),
			sessionRow("2026-06-10", "Other", "2026-06-10T20:00:00", 1800),
		},
	}
	var llmPrompt string
	service := &Service{
		repo:         repo,
		statsService: fakeWeeklyStatsProvider{},
		llmClient:    fakeLLM{content: "ok", prompt: &llmPrompt},
	}

	result, err := service.GenerateDailySummary(context.Background(), "2026-06-18")
	if err != nil {
		t.Fatalf("GenerateDailySummary returned error: %v", err)
	}
	if result.SummaryID != 42 {
		t.Fatalf("summary id = %d, want 42", result.SummaryID)
	}
	if len(repo.created.SourceData) == 0 {
		t.Fatal("created source_data is empty")
	}
	if !strings.Contains(llmPrompt, string(repo.created.SourceData)) {
		t.Fatal("persisted source_data does not match the structured JSON sent to the LLM")
	}
	if !reflect.DeepEqual(repo.requestedTaskDates, []string{"2026-06-18", "2026-06-15", "2026-06-10"}) {
		t.Fatalf("task query dates = %v, want target plus recent active dates", repo.requestedTaskDates)
	}
	if repo.requestedRecentBeforeDate != "2026-06-18" {
		t.Fatalf("recent active date query beforeDate = %q, want target date", repo.requestedRecentBeforeDate)
	}

	var source DailySummarySourceData
	if err := json.Unmarshal(repo.created.SourceData, &source); err != nil {
		t.Fatalf("source_data is not DailySummarySourceData JSON: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(repo.created.SourceData, &raw); err != nil {
		t.Fatalf("source_data is not an object: %v", err)
	}
	for _, key := range []string{"summary_type", "target_date", "data_quality", "today", "recent_context"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("source_data is missing top-level key %q: %s", key, string(repo.created.SourceData))
		}
	}
	recentContext, ok := raw["recent_context"].(map[string]any)
	if !ok {
		t.Fatalf("source_data.recent_context is not an object: %s", string(repo.created.SourceData))
	}
	for _, key := range []string{"recent_active_days", "project_patterns", "repeated_notes"} {
		if _, ok := recentContext[key]; !ok {
			t.Fatalf("source_data.recent_context is missing key %q: %s", key, string(repo.created.SourceData))
		}
	}
	today, ok := raw["today"].(map[string]any)
	if !ok {
		t.Fatalf("source_data.today is not an object: %s", string(repo.created.SourceData))
	}
	if _, ok := today["first_start_time"]; !ok {
		t.Fatalf("source_data.today is missing first_start_time: %s", string(repo.created.SourceData))
	}

	if source.SummaryType != "daily" ||
		source.TargetDate != "2026-06-18" ||
		source.DataQuality.DaysWithData != 3 ||
		source.Today.FirstStartTime != "15:20" ||
		source.Today.TotalFocusMinutes == 0 ||
		len(source.RecentContext.RecentActiveDays) != 2 {
		t.Fatalf("persisted source_data missing required daily context: %+v", source)
	}
}

func TestGenerateWeeklySummaryPersistsStructuredSourceData(t *testing.T) {
	overrideSeconds := 3600
	repo := &fakeSummaryRepo{
		tasks: []dailySummaryTaskRow{
			taskRow("2026-06-15", "Personal Timer", "completed", 60, 4200, nil, "接口联调", "前端状态刷新"),
			taskRow("2026-06-16", "Go", "completed", 30, 600, &overrideSeconds, "map slice", "channel context"),
			taskRow("2026-06-10", "Personal Timer", "completed", 60, 3000, nil, "", ""),
		},
		sessions: []dailySummarySessionRow{
			sessionRow("2026-06-15", "Personal Timer", "2026-06-15T15:20:00", 4200),
			sessionRow("2026-06-16", "Go", "2026-06-16T14:10:00", 600),
			sessionRow("2026-06-10", "Personal Timer", "2026-06-10T16:00:00", 3000),
		},
	}
	var llmPrompt string
	service := &Service{
		repo:         repo,
		statsService: fakeWeeklyStatsProvider{},
		llmClient:    fakeLLM{content: "weekly ok", prompt: &llmPrompt},
	}

	result, err := service.GenerateWeeklySummary(context.Background(), "2026-06-15", "2026-06-21")
	if err != nil {
		t.Fatalf("GenerateWeeklySummary returned error: %v", err)
	}
	if result.SummaryID != 42 {
		t.Fatalf("summary id = %d, want 42", result.SummaryID)
	}
	if repo.created.SummaryType != "weekly" || repo.created.StartDate != "2026-06-15" || repo.created.EndDate != "2026-06-21" {
		t.Fatalf("unexpected created summary identity: %+v", repo.created)
	}
	if len(repo.created.SourceData) == 0 {
		t.Fatal("created weekly source_data is empty")
	}
	if !strings.Contains(llmPrompt, string(repo.created.SourceData)) {
		t.Fatal("persisted weekly source_data does not match the structured JSON sent to the LLM")
	}
	if len(repo.requestedTaskDateCalls) != 2 {
		t.Fatalf("task date query call count = %d, want week and previous week", len(repo.requestedTaskDateCalls))
	}
	if !reflect.DeepEqual(repo.requestedTaskDateCalls[0], mustDateRange("2026-06-15", "2026-06-21")) {
		t.Fatalf("week task dates = %v, want week range", repo.requestedTaskDateCalls[0])
	}
	if !reflect.DeepEqual(repo.requestedTaskDateCalls[1], mustDateRange("2026-06-08", "2026-06-14")) {
		t.Fatalf("previous week task dates = %v, want previous week range", repo.requestedTaskDateCalls[1])
	}

	var source WeeklySummarySourceData
	if err := json.Unmarshal(repo.created.SourceData, &source); err != nil {
		t.Fatalf("source_data is not WeeklySummarySourceData JSON: %v", err)
	}
	if source.SummaryType != "weekly" ||
		source.WeekStart != "2026-06-15" ||
		source.WeekEnd != "2026-06-21" ||
		len(source.Week.DailyTotals) == 0 ||
		len(source.Week.ProjectBreakdown) == 0 ||
		len(source.Week.StartTimePatterns) == 0 {
		t.Fatalf("persisted weekly source_data missing required context: %+v", source)
	}

	var raw map[string]any
	if err := json.Unmarshal(repo.created.SourceData, &raw); err != nil {
		t.Fatalf("weekly source_data is not an object: %v", err)
	}
	for _, key := range []string{"summary_type", "week_start", "week_end", "data_quality", "week", "previous_week_comparison"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("weekly source_data is missing top-level key %q: %s", key, string(repo.created.SourceData))
		}
	}
	week, ok := raw["week"].(map[string]any)
	if !ok {
		t.Fatalf("weekly source_data.week is not an object: %s", string(repo.created.SourceData))
	}
	for _, key := range []string{"daily_totals", "project_breakdown", "time_distribution", "start_time_patterns", "repeated_notes"} {
		if _, ok := week[key]; !ok {
			t.Fatalf("weekly source_data.week is missing key %q: %s", key, string(repo.created.SourceData))
		}
	}
}

func taskRow(date, project, status string, estimatedMinutes, sessionSeconds int, overrideSeconds *int, finishNote, finishDescription string) dailySummaryTaskRow {
	row := dailySummaryTaskRow{
		Date:              date,
		ProjectName:       project,
		Status:            status,
		EstimatedMinutes:  estimatedMinutes,
		SessionSeconds:    sessionSeconds,
		FinishNote:        finishNote,
		FinishDescription: finishDescription,
	}
	if overrideSeconds != nil {
		row.ActualSecondsOverride = sql.NullInt64{Int64: int64(*overrideSeconds), Valid: true}
	}
	return row
}

func sessionRow(date, project, startedAt string, durationSeconds int) dailySummarySessionRow {
	parsed, err := time.Parse("2006-01-02T15:04:05", startedAt)
	if err != nil {
		panic(err)
	}
	return dailySummarySessionRow{
		Date:            date,
		ProjectName:     project,
		StartedAt:       parsed,
		DurationSeconds: durationSeconds,
	}
}

func mustDateRange(startDate, endDate string) []string {
	dates, err := dateRange(startDate, endDate)
	if err != nil {
		panic(err)
	}
	return dates
}

type fakeSummaryRepo struct {
	recentDates               []string
	tasks                     []dailySummaryTaskRow
	sessions                  []dailySummarySessionRow
	created                   CreateSummaryInput
	requestedRecentBeforeDate string
	requestedTaskDates        []string
	requestedSessionDates     []string
	requestedTaskDateCalls    [][]string
	requestedSessionDateCalls [][]string
}

func (r *fakeSummaryRepo) CreateSummary(ctx context.Context, input CreateSummaryInput) (int64, error) {
	r.created = input
	return 42, nil
}

func (r *fakeSummaryRepo) SummaryExists(ctx context.Context, summaryType, startDate, endDate string) (bool, error) {
	return false, nil
}

func (r *fakeSummaryRepo) ListSummaries(ctx context.Context, summaryType string) ([]GeneratedSummary, error) {
	return nil, nil
}

func (r *fakeSummaryRepo) GetSummaryByID(ctx context.Context, id int64) (*GeneratedSummary, error) {
	return nil, nil
}

func (r *fakeSummaryRepo) DeleteSummary(ctx context.Context, id int64) error {
	return nil
}

func (r *fakeSummaryRepo) ListRecentDailyActiveDates(ctx context.Context, beforeDate string, limit int) ([]string, error) {
	r.requestedRecentBeforeDate = beforeDate
	if limit != DailyRecentActiveDaysLimit {
		return nil, nil
	}
	return r.recentDates, nil
}

func (r *fakeSummaryRepo) ListDailySummaryTasks(ctx context.Context, dates []string) ([]dailySummaryTaskRow, error) {
	r.requestedTaskDates = append([]string(nil), dates...)
	r.requestedTaskDateCalls = append(r.requestedTaskDateCalls, append([]string(nil), dates...))
	allowedDates := dateSet(dates)
	tasks := make([]dailySummaryTaskRow, 0)
	for _, task := range r.tasks {
		if allowedDates[task.Date] {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *fakeSummaryRepo) ListDailySummarySessions(ctx context.Context, dates []string) ([]dailySummarySessionRow, error) {
	r.requestedSessionDates = append([]string(nil), dates...)
	r.requestedSessionDateCalls = append(r.requestedSessionDateCalls, append([]string(nil), dates...))
	allowedDates := dateSet(dates)
	sessions := make([]dailySummarySessionRow, 0)
	for _, session := range r.sessions {
		if allowedDates[session.Date] {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func dateSet(dates []string) map[string]bool {
	result := make(map[string]bool, len(dates))
	for _, date := range dates {
		result[date] = true
	}
	return result
}

type fakeWeeklyStatsProvider struct{}

func (fakeWeeklyStatsProvider) GetWeeklyStats(startDate, endDate string) (*stats.WeeklyStats, error) {
	return &stats.WeeklyStats{StartDate: startDate, EndDate: endDate}, nil
}

type fakeLLM struct {
	content string
	prompt  *string
}

func (f fakeLLM) GenerateSummary(ctx context.Context, prompt string) (string, error) {
	if f.prompt != nil {
		*f.prompt = prompt
	}
	return f.content, nil
}
