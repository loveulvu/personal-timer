package summaries

import "testing"

func TestBuildDailyActionItemsAddsMissingFixedProjects(t *testing.T) {
	source := DailySummarySourceData{
		Today: DailySummaryToday{
			ProjectBreakdown: []DailyProjectBreakdown{
				{ProjectName: "后端学习", TotalMinutes: 60},
			},
		},
	}

	items := BuildDailyActionItems(source)

	assertActionItem(t, items, "schedule", "high", "算法练习")
	assertActionItem(t, items, "schedule", "high", "背单词")
}

func TestBuildDailyActionItemsSkipsFixedProjectsWhenPresent(t *testing.T) {
	source := DailySummarySourceData{
		Today: DailySummaryToday{
			ProjectBreakdown: []DailyProjectBreakdown{
				{ProjectName: "算法练习", TotalMinutes: 45},
				{ProjectName: "背单词", TotalMinutes: 15},
			},
		},
	}

	items := BuildDailyActionItems(source)

	if hasActionItem(items, "schedule", "算法练习") || hasActionItem(items, "schedule", "背单词") {
		t.Fatalf("items = %+v, should not include fixed project schedule items", items)
	}
}

func TestBuildWeeklyActionItemsAddsFixedProjectConsistency(t *testing.T) {
	source := WeeklySummarySourceData{
		DataQuality: WeeklyDataQuality{DaysWithData: 6},
		Week: WeeklySummaryWeek{
			ProjectBreakdown: []WeeklyProjectBreakdown{
				{ProjectName: "算法练习", ActiveDays: 4},
				{ProjectName: "背单词", ActiveDays: 5},
			},
		},
	}

	items := BuildWeeklyActionItems(source)

	assertActionItem(t, items, "consistency", "high", "算法练习")
	assertActionItem(t, items, "consistency", "high", "背单词")
}

func TestBuildWeeklyActionItemsAddsFlexibleProjectSuggestions(t *testing.T) {
	source := WeeklySummarySourceData{
		DataQuality: WeeklyDataQuality{DaysWithData: 4},
		Week: WeeklySummaryWeek{
			ProjectBreakdown: []WeeklyProjectBreakdown{
				{ProjectName: "算法练习", ActiveDays: 4},
				{ProjectName: "背单词", ActiveDays: 4},
				{ProjectName: "后端学习", ActiveDays: 2},
				{ProjectName: "项目推进", ActiveDays: 1, OverrunMinutes: 75, OverrunRate: 0.2},
			},
		},
	}

	items := BuildWeeklyActionItems(source)

	assertActionItem(t, items, "schedule", "medium", "后端学习")
	assertActionItem(t, items, "split_task", "high", "项目推进")
	if hasActionItem(items, "consistency", "项目推进") {
		t.Fatalf("items = %+v, should not require 项目推进 to appear every day", items)
	}
}

func TestBuildActionItemsUsesRepeatedGoKeywordsOnly(t *testing.T) {
	dailyItems := BuildDailyActionItems(DailySummarySourceData{
		Today: DailySummaryToday{
			ProjectBreakdown: []DailyProjectBreakdown{
				{ProjectName: "算法练习"},
				{ProjectName: "背单词"},
			},
		},
		RecentContext: DailySummaryContext{RepeatedNotes: []string{"context", "channel", "slice", "map", "WithCancel", "WithTimeout"}},
	})
	assertActionItem(t, dailyItems, "focus_topic", "medium", "后端学习")

	weeklyItems := BuildWeeklyActionItems(WeeklySummarySourceData{
		DataQuality: WeeklyDataQuality{DaysWithData: 3},
		Week: WeeklySummaryWeek{
			ProjectBreakdown: []WeeklyProjectBreakdown{
				{ProjectName: "算法练习", ActiveDays: 3},
				{ProjectName: "背单词", ActiveDays: 3},
				{ProjectName: "后端学习", ActiveDays: 3},
			},
			RepeatedNotes: []string{"done", "todo", "normal"},
		},
	})
	if hasActionItem(weeklyItems, "focus_topic", "后端学习") {
		t.Fatalf("items = %+v, should not create focus_topic from meaningless notes", weeklyItems)
	}
}

func TestBuildActionItemsAddsCleanupForUnassignedTasks(t *testing.T) {
	dailyItems := BuildDailyActionItems(DailySummarySourceData{
		Today: DailySummaryToday{
			ProjectBreakdown: []DailyProjectBreakdown{
				{ProjectName: "算法练习"},
				{ProjectName: "背单词"},
			},
		},
		Excluded: SummaryExcluded{UnassignedTaskCount: 1},
	})
	assertActionItem(t, dailyItems, "cleanup", "low", "")

	weeklyItems := BuildWeeklyActionItems(WeeklySummarySourceData{
		DataQuality: WeeklyDataQuality{DaysWithData: 1},
		Week: WeeklySummaryWeek{
			ProjectBreakdown: []WeeklyProjectBreakdown{
				{ProjectName: "算法练习", ActiveDays: 1},
				{ProjectName: "背单词", ActiveDays: 1},
			},
		},
		Excluded: SummaryExcluded{UnassignedTaskCount: 2},
	})
	assertActionItem(t, weeklyItems, "cleanup", "low", "")
}

func TestBuildActionItemsLimitsToFiveAndSortsByPriority(t *testing.T) {
	source := WeeklySummarySourceData{
		DataQuality: WeeklyDataQuality{DaysWithData: 6},
		Week: WeeklySummaryWeek{
			ProjectBreakdown: []WeeklyProjectBreakdown{
				{ProjectName: "算法练习", ActiveDays: 4},
				{ProjectName: "背单词", ActiveDays: 4},
				{ProjectName: "后端学习", ActiveDays: 1},
				{ProjectName: "项目推进", ActiveDays: 1, OverrunRate: 0.5},
			},
			RepeatedNotes: []string{"context", "channel"},
		},
		Excluded: SummaryExcluded{UnassignedTaskCount: 1},
	}

	items := BuildWeeklyActionItems(source)

	if len(items) != 5 {
		t.Fatalf("len(items) = %d, want 5: %+v", len(items), items)
	}
	for i := 1; i < len(items); i++ {
		if priorityRank(items[i-1].Priority) > priorityRank(items[i].Priority) {
			t.Fatalf("items are not sorted by priority: %+v", items)
		}
	}
	if hasActionItem(items, "cleanup", "") {
		t.Fatalf("items = %+v, should drop lowest priority item when over limit", items)
	}
}

func assertActionItem(t *testing.T, items []SummaryActionItem, itemType, priority, project string) {
	t.Helper()
	for _, item := range items {
		if item.Type == itemType && item.Priority == priority && item.SuggestedProject == project {
			return
		}
	}
	t.Fatalf("items = %+v, want type=%s priority=%s project=%s", items, itemType, priority, project)
}

func hasActionItem(items []SummaryActionItem, itemType, project string) bool {
	for _, item := range items {
		if item.Type == itemType && item.SuggestedProject == project {
			return true
		}
	}
	return false
}
