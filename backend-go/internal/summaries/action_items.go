package summaries

import (
	"sort"
	"strconv"
	"strings"
)

const maxSummaryActionItems = 5

var fixedDailyProjectMinutes = map[string]int{
	"算法练习": 45,
	"背单词":  15,
}

var goReviewKeywords = map[string]bool{
	"append":      true,
	"channel":     true,
	"close":       true,
	"context":     true,
	"map":         true,
	"slice":       true,
	"withcancel":  true,
	"withtimeout": true,
}

func BuildDailyActionItems(source DailySummarySourceData) []SummaryActionItem {
	items := make([]SummaryActionItem, 0)
	projects := dailyBreakdownByProject(source.Today.ProjectBreakdown)

	for _, projectName := range []string{"算法练习", "背单词"} {
		if _, ok := projects[projectName]; ok {
			continue
		}
		items = append(items, SummaryActionItem{
			Type:             "schedule",
			Priority:         "high",
			Title:            "明天补上" + projectName,
			Reason:           projectName + "是固定每日项目，但目标日期未出现相关记录。",
			SuggestedProject: projectName,
			SuggestedMinutes: fixedDailyProjectMinutes[projectName],
			Source:           "fixed_daily_project",
		})
	}

	if project, ok := projects["项目推进"]; ok && project.OverrunMinutes > 30 {
		items = append(items, SummaryActionItem{
			Type:             "split_task",
			Priority:         "medium",
			Title:            "拆分项目推进任务",
			Reason:           "项目推进目标日实际时长超过预估 " + strconv.Itoa(project.OverrunMinutes) + " 分钟，建议拆小任务或重新估时。",
			SuggestedProject: "项目推进",
			SuggestedMinutes: 60,
			Source:           "estimate_bias",
		})
	}

	if containsGoReviewKeyword(source.RecentContext.RepeatedNotes) {
		items = append(items, SummaryActionItem{
			Type:             "focus_topic",
			Priority:         "medium",
			Title:            "整理 Go 高频卡点",
			Reason:           "近期备注中反复出现 Go 基础或并发关键词。",
			SuggestedProject: "后端学习",
			SuggestedMinutes: 30,
			Source:           "repeated_notes",
		})
	}

	if source.Excluded.UnassignedTaskCount > 0 {
		items = append(items, SummaryActionItem{
			Type:     "cleanup",
			Priority: "low",
			Title:    "补充未绑定任务的项目归属",
			Reason:   "存在未绑定项目的任务，已从学习总结中排除。",
			Source:   "unassigned_tasks",
		})
	}

	return normalizeActionItems(items)
}

func BuildWeeklyActionItems(source WeeklySummarySourceData) []SummaryActionItem {
	items := make([]SummaryActionItem, 0)
	projects := weeklyBreakdownByProject(source.Week.ProjectBreakdown)
	daysWithData := source.DataQuality.DaysWithData

	for _, projectName := range []string{"算法练习", "背单词"} {
		activeDays := 0
		if project, ok := projects[projectName]; ok {
			activeDays = project.ActiveDays
		}
		if activeDays < daysWithData {
			items = append(items, SummaryActionItem{
				Type:             "consistency",
				Priority:         "high",
				Title:            "下周保持" + projectName + "每天出现",
				Reason:           projectName + "是固定每日项目，本周活跃天数少于本周有学习数据天数。",
				SuggestedProject: projectName,
				SuggestedMinutes: fixedDailyProjectMinutes[projectName],
				Source:           "fixed_daily_project",
			})
		}
	}

	if project, ok := projects["项目推进"]; ok && (project.OverrunMinutes > 60 || project.OverrunRate >= 0.4) {
		items = append(items, SummaryActionItem{
			Type:             "split_task",
			Priority:         "high",
			Title:            "拆分项目推进任务",
			Reason:           "项目推进本周实际时长明显高于预估，说明任务粒度或预估存在偏差。",
			SuggestedProject: "项目推进",
			SuggestedMinutes: 60,
			Source:           "estimate_bias",
		})
	}

	backendActiveDays := 0
	if project, ok := projects["后端学习"]; ok {
		backendActiveDays = project.ActiveDays
	}
	if daysWithData > 0 && backendActiveDays < 3 {
		items = append(items, SummaryActionItem{
			Type:             "schedule",
			Priority:         "medium",
			Title:            "下周增加后端学习频率",
			Reason:           "后端学习本周活跃天数较少，不要求每天出现，但需要保持连续输入。",
			SuggestedProject: "后端学习",
			SuggestedMinutes: 45,
			Source:           "low_active_days",
		})
	}

	if containsGoReviewKeyword(source.Week.RepeatedNotes) {
		items = append(items, SummaryActionItem{
			Type:             "focus_topic",
			Priority:         "medium",
			Title:            "整理 Go 并发与基础数据结构笔记",
			Reason:           "本周备注中反复出现 context/channel/slice/map 等关键词，说明这些概念仍在反复出现。",
			SuggestedProject: "后端学习",
			SuggestedMinutes: 30,
			Source:           "repeated_notes",
		})
	}

	if source.Excluded.UnassignedTaskCount > 0 {
		items = append(items, SummaryActionItem{
			Type:     "cleanup",
			Priority: "low",
			Title:    "清理未绑定项目任务",
			Reason:   "本周存在未绑定项目任务，已从学习总结中排除，建议补充项目归属。",
			Source:   "unassigned_tasks",
		})
	}

	return normalizeActionItems(items)
}

func dailyBreakdownByProject(projects []DailyProjectBreakdown) map[string]DailyProjectBreakdown {
	result := make(map[string]DailyProjectBreakdown, len(projects))
	for _, project := range projects {
		result[project.ProjectName] = project
	}
	return result
}

func weeklyBreakdownByProject(projects []WeeklyProjectBreakdown) map[string]WeeklyProjectBreakdown {
	result := make(map[string]WeeklyProjectBreakdown, len(projects))
	for _, project := range projects {
		result[project.ProjectName] = project
	}
	return result
}

func containsGoReviewKeyword(notes []string) bool {
	for _, note := range notes {
		fields := strings.FieldsFunc(note, func(r rune) bool {
			return !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_'
		})
		for _, field := range fields {
			if goReviewKeywords[strings.ToLower(field)] {
				return true
			}
		}
		if goReviewKeywords[strings.ToLower(strings.TrimSpace(note))] {
			return true
		}
	}
	return false
}

func normalizeActionItems(items []SummaryActionItem) []SummaryActionItem {
	deduped := make([]SummaryActionItem, 0, len(items))
	seen := make(map[string]bool)
	for _, item := range items {
		key := item.Type + "\x00" + item.SuggestedProject
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, item)
	}

	sort.SliceStable(deduped, func(i, j int) bool {
		return priorityRank(deduped[i].Priority) < priorityRank(deduped[j].Priority)
	})
	if len(deduped) > maxSummaryActionItems {
		deduped = deduped[:maxSummaryActionItems]
	}
	return deduped
}

func priorityRank(priority string) int {
	switch priority {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 3
	}
}
