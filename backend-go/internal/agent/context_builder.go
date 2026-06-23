package agent

import (
	"context"
	"encoding/json"
	"errors"
	"personal/internal/dailytasks"
	"personal/internal/memories"
	"personal/internal/summaries"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultContextRecentDays = 5
	maxContextRecentDays     = 14
	contextSummaryLimit      = 500
	contextActionItemsLimit  = 800
	contextMemoryLimit       = 8
	contextMemoryTextLimit   = 400
	contextEvidenceLimit     = 240
	contextActionItemLimit   = 20
)

var ErrInvalidContextPreviewInput = errors.New("invalid context preview input")

type ContextPreviewRequest struct {
	Goal       string `json:"goal"`
	TargetDate string `json:"target_date"`
	RecentDays int    `json:"recent_days"`
}

type ContextPreviewResponse struct {
	ContextPack ContextPack `json:"context_pack"`
}

type summaryLister interface {
	ListSummaries(ctx context.Context, summaryType string) ([]summaries.GeneratedSummary, error)
	ListActionItemAcceptances(ctx context.Context, summaryID int64) ([]summaries.ActionItemAcceptance, error)
}

type memoryContextStore interface {
	ListMemories(ctx context.Context, filter memories.ListMemoriesFilter) ([]memories.StudyMemory, error)
	ListMemoryEvidence(ctx context.Context, memoryID int64) ([]memories.StudyMemoryEvidence, error)
}

type ContextPackBuilder struct {
	tasks     dailyTaskLister
	planRisk  planRiskGetter
	summaries summaryLister
	memories  memoryContextStore
}

func NewContextPackBuilder(tasks dailyTaskLister, planRisk planRiskGetter, summaries summaryLister, memories memoryContextStore) *ContextPackBuilder {
	return &ContextPackBuilder{tasks: tasks, planRisk: planRisk, summaries: summaries, memories: memories}
}

func (b *ContextPackBuilder) Build(ctx context.Context, req ContextPreviewRequest) (ContextPack, error) {
	goal := strings.TrimSpace(req.Goal)
	if goal == "" {
		return ContextPack{}, ErrInvalidContextPreviewInput
	}
	target, err := time.Parse("2006-01-02", strings.TrimSpace(req.TargetDate))
	if err != nil {
		return ContextPack{}, ErrInvalidContextPreviewInput
	}

	recentDays := req.RecentDays
	omitted := make([]string, 0)
	if recentDays == 0 {
		recentDays = defaultContextRecentDays
	}
	if recentDays < 1 {
		return ContextPack{}, ErrInvalidContextPreviewInput
	}
	if recentDays > maxContextRecentDays {
		recentDays = maxContextRecentDays
		omitted = appendOmitted(omitted, "recent_days_capped_to_14")
	}

	pack := ContextPack{
		UserGoal:          goal,
		TargetDate:        target.Format("2006-01-02"),
		TodayTasks:        []ContextTask{},
		RecentSummaries:   []ContextSummary{},
		Memories:          []ContextMemory{},
		RecentActionItems: []ContextActionItem{},
		Constraints:       contextConstraints(),
		OmittedSections:   omitted,
	}

	if b.tasks != nil {
		tasks, err := b.tasks.ListDailyTasksByDate(pack.TargetDate)
		if err != nil {
			return ContextPack{}, err
		}
		pack.TodayTasks = compactTasks(tasks)
	}

	if b.planRisk != nil {
		risk, err := b.planRisk.GetPlanRisk(ctx, pack.TargetDate)
		if err != nil {
			pack.OmittedSections = appendOmitted(pack.OmittedSections, "plan_risk_unavailable")
		} else {
			pack.PlanRisk, _ = json.Marshal(risk)
		}
	} else {
		pack.OmittedSections = appendOmitted(pack.OmittedSections, "plan_risk_unavailable")
	}

	if b.summaries != nil {
		recent, actionItems, omitted, err := b.buildSummaries(ctx, target, recentDays)
		if err != nil {
			return ContextPack{}, err
		}
		pack.RecentSummaries = recent
		pack.RecentActionItems = actionItems
		for _, section := range omitted {
			pack.OmittedSections = appendOmitted(pack.OmittedSections, section)
		}
	}

	if b.memories != nil {
		items, omitted, err := b.buildMemories(ctx)
		if err != nil {
			return ContextPack{}, err
		}
		pack.Memories = items
		for _, section := range omitted {
			pack.OmittedSections = appendOmitted(pack.OmittedSections, section)
		}
	}

	return pack, nil
}

func (b *ContextPackBuilder) buildSummaries(ctx context.Context, target time.Time, recentDays int) ([]ContextSummary, []ContextActionItem, []string, error) {
	all, err := b.summaries.ListSummaries(ctx, "daily")
	if err != nil {
		return nil, nil, nil, err
	}
	start := target.AddDate(0, 0, -recentDays)
	filtered := make([]summaries.GeneratedSummary, 0)
	for _, summary := range all {
		end, err := time.Parse("2006-01-02", summary.EndDate)
		if err != nil || end.Before(start) || !end.Before(target) {
			continue
		}
		filtered = append(filtered, summary)
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].EndDate > filtered[j].EndDate
	})

	omitted := make([]string, 0)
	result := make([]ContextSummary, 0, len(filtered))
	actionItems := make([]ContextActionItem, 0)
	for _, summary := range filtered {
		content, truncated := excerpt(summary.Content, contextSummaryLimit)
		if truncated {
			omitted = appendOmitted(omitted, "summary_content_excerpted")
		}
		items, itemExcerpt, itemTruncated := compactSummaryActionItems(summary.ActionItems, contextActionItemsLimit)
		if itemTruncated {
			omitted = appendOmitted(omitted, "summary_action_items_excerpted")
		}
		result = append(result, ContextSummary{
			ID:                 summary.ID,
			SummaryType:        summary.SummaryType,
			StartDate:          summary.StartDate,
			EndDate:            summary.EndDate,
			ContentExcerpt:     content,
			ActionItemsExcerpt: itemExcerpt,
			CreatedAt:          summary.CreatedAt,
		})
		acceptances := map[int]summaries.ActionItemAcceptance{}
		for _, acceptance := range mustListAcceptances(ctx, b.summaries, summary.ID) {
			acceptances[acceptance.ItemIndex] = acceptance
		}
		for i, item := range items {
			accepted := acceptances[i]
			actionItems = append(actionItems, ContextActionItem{
				SummaryID:    summary.ID,
				ItemIndex:    i,
				Content:      item.Title,
				Accepted:     accepted.ID > 0,
				TargetDate:   accepted.TargetDate,
				TargetTaskID: accepted.TargetTaskID,
			})
			if len(actionItems) >= contextActionItemLimit {
				omitted = appendOmitted(omitted, "recent_action_items_truncated")
				break
			}
		}
	}
	return result, actionItems, omitted, nil
}

func (b *ContextPackBuilder) buildMemories(ctx context.Context) ([]ContextMemory, []string, error) {
	memoriesList, err := b.memories.ListMemories(ctx, memories.ListMemoriesFilter{Status: "active", Limit: contextMemoryLimit * 3})
	if err != nil {
		return nil, nil, err
	}
	omitted := make([]string, 0)
	result := make([]ContextMemory, 0, contextMemoryLimit)
	for _, memory := range memoriesList {
		if memory.Status == "archived" {
			omitted = appendOmitted(omitted, "archived_memories_omitted")
			continue
		}
		if memory.Status != "active" || memory.Confidence < 0.5 {
			omitted = appendOmitted(omitted, "low_confidence_memories_omitted")
			continue
		}
		content, truncated := excerpt(memory.Content, contextMemoryTextLimit)
		if truncated {
			omitted = appendOmitted(omitted, "memory_content_excerpted")
		}
		evidence, err := b.memories.ListMemoryEvidence(ctx, memory.ID)
		if err != nil {
			return nil, nil, err
		}
		evidenceExcerpt, evidenceTruncated := compactEvidenceExcerpt(evidence)
		if evidenceTruncated {
			omitted = appendOmitted(omitted, "memory_evidence_excerpted")
		}
		result = append(result, ContextMemory{
			ID:                 memory.ID,
			MemoryType:         memory.MemoryType,
			ScopeType:          memory.ScopeType,
			ProjectID:          memory.ProjectID,
			Title:              memory.Title,
			Content:            content,
			Confidence:         memory.Confidence,
			SupportCount:       memory.SupportCount,
			ContradictionCount: memory.ContradictionCount,
			EvidenceCount:      len(evidence),
			EvidenceExcerpt:    evidenceExcerpt,
			Status:             memory.Status,
			LastSeenAt:         memory.LastSeenAt,
		})
		if len(result) >= contextMemoryLimit {
			break
		}
	}
	return result, omitted, nil
}

func compactTasks(tasks []dailytasks.DailyTask) []ContextTask {
	result := make([]ContextTask, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, ContextTask{
			ID:               task.ID,
			ProjectID:        task.ProjectID,
			TaskDate:         task.TaskDate,
			Title:            task.Title,
			EstimatedMinutes: task.EstimatedMinutes,
			ActualMinutes:    task.ActualSeconds / 60,
			Status:           task.Status,
		})
	}
	return result
}

func contextConstraints() []string {
	return []string{
		"read tools may be executed automatically",
		"write tools require user confirmation",
		"destructive tools are disabled",
		"do not infer facts not present in context",
		"do not create or modify tasks without action proposal",
	}
}

func compactSummaryActionItems(data json.RawMessage, limit int) ([]summaries.SummaryActionItem, string, bool) {
	if len(data) == 0 || !json.Valid(data) {
		return []summaries.SummaryActionItem{}, "", false
	}
	var items []summaries.SummaryActionItem
	if err := json.Unmarshal(data, &items); err != nil {
		return []summaries.SummaryActionItem{}, "", false
	}
	excerpted, truncated := excerpt(string(data), limit)
	return items, excerpted, truncated
}

func compactEvidenceExcerpt(items []memories.StudyMemoryEvidence) (string, bool) {
	if len(items) == 0 {
		return "", false
	}
	value := items[0].SourceType
	if items[0].Excerpt != nil {
		value = *items[0].Excerpt
	}
	return excerpt(value, contextEvidenceLimit)
}

func mustListAcceptances(ctx context.Context, source summaryLister, summaryID int64) []summaries.ActionItemAcceptance {
	items, err := source.ListActionItemAcceptances(ctx, summaryID)
	if err != nil {
		return []summaries.ActionItemAcceptance{}
	}
	return items
}

func excerpt(value string, limit int) (string, bool) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) <= limit {
		return value, false
	}
	runes := []rune(value)
	return string(runes[:limit]), true
}

func appendOmitted(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}
