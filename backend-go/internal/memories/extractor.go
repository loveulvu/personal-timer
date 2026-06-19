package memories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

var ErrInvalidSourceData = errors.New("invalid summary source_data")

type memoryStore interface {
	GetSummaryForExtraction(ctx context.Context, id int64) (summaryForExtraction, error)
	FindProjectForExtraction(ctx context.Context, projectID *int64, name string) (*projectForExtraction, error)
	FindActiveMemoryByIdentity(ctx context.Context, memoryType, scopeType string, projectID *int64, title string) (StudyMemory, error)
	EvidenceExists(ctx context.Context, memoryID int64, sourceType string, sourceID int64) (bool, error)
	CreateMemory(ctx context.Context, input CreateMemoryInput) (StudyMemory, error)
	UpdateMemory(ctx context.Context, id int64, input UpdateMemoryInput) (StudyMemory, error)
	AddEvidence(ctx context.Context, input AddEvidenceInput) (StudyMemoryEvidence, error)
}

type Extractor struct {
	store memoryStore
	now   func() time.Time
}

func NewExtractor(store memoryStore) *Extractor {
	return &Extractor{store: store, now: time.Now}
}

func (e *Extractor) ExtractFromSummary(ctx context.Context, summaryID int64) (ExtractionResult, error) {
	result := ExtractionResult{SummaryID: summaryID}
	summary, err := e.store.GetSummaryForExtraction(ctx, summaryID)
	if err != nil {
		return result, err
	}
	if len(summary.SourceData) == 0 {
		return result, nil
	}

	var source map[string]any
	if err := json.Unmarshal(summary.SourceData, &source); err != nil {
		return result, ErrInvalidSourceData
	}
	if len(summary.ActionItems) > 0 && !json.Valid(summary.ActionItems) {
		result.Warnings = append(result.Warnings, "action_items JSON invalid, ignored")
	}

	candidates := e.dedupeCandidates(append(
		append(buildRepeatedBlockerCandidates(summary, source), buildEstimateBiasCandidates(summary, source)...),
		buildTimePatternCandidates(summary, source)...,
	))

	for _, candidate := range candidates {
		memory, created, updated, evidenceAdded, err := e.upsertCandidate(ctx, candidate)
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
			result.SkippedCount++
			continue
		}
		if created {
			result.CreatedCount++
		} else if updated {
			result.UpdatedCount++
		} else {
			result.SkippedCount++
		}
		if evidenceAdded {
			result.EvidenceCount++
		}
		result.Memories = append(result.Memories, memory)
	}

	return result, nil
}

func (e *Extractor) upsertCandidate(ctx context.Context, candidate memoryCandidate) (StudyMemory, bool, bool, bool, error) {
	if candidate.MemoryType == "estimate_bias" && candidate.ProjectID == nil {
		project, err := e.store.FindProjectForExtraction(ctx, nil, candidate.ProjectName)
		if err != nil || project == nil || !project.IncludeInSummary {
			return StudyMemory{}, false, false, false, fmt.Errorf("skip estimate_bias project %q", candidate.ProjectName)
		}
		candidate.ProjectID = &project.ID
		if candidate.ProjectName == "" {
			candidate.ProjectName = project.Name
			candidate.Title = "估时偏差：" + project.Name + " 经常超时"
		}
	}
	if candidate.MemoryType == "estimate_bias" && candidate.ProjectID != nil {
		project, err := e.store.FindProjectForExtraction(ctx, candidate.ProjectID, "")
		if err != nil || project == nil || !project.IncludeInSummary {
			return StudyMemory{}, false, false, false, fmt.Errorf("skip estimate_bias project id %d", *candidate.ProjectID)
		}
		if candidate.ProjectName == "" {
			candidate.ProjectName = project.Name
			candidate.Title = "估时偏差：" + project.Name + " 经常超时"
		}
	}

	now := e.now()
	existing, err := e.store.FindActiveMemoryByIdentity(ctx, candidate.MemoryType, candidate.ScopeType, candidate.ProjectID, candidate.Title)
	if errors.Is(err, ErrMemoryNotFound) {
		memory, err := e.store.CreateMemory(ctx, CreateMemoryInput{
			MemoryType:         candidate.MemoryType,
			ScopeType:          candidate.ScopeType,
			ProjectID:          candidate.ProjectID,
			Title:              candidate.Title,
			Content:            candidate.Content,
			StructuredData:     candidate.StructuredData,
			Confidence:         candidate.InitialConfidence,
			SupportCount:       1,
			ContradictionCount: 0,
			FirstSeenAt:        now,
			LastSeenAt:         now,
			Status:             "active",
		})
		if err != nil {
			return StudyMemory{}, false, false, false, err
		}
		if err := e.addEvidence(ctx, memory.ID, candidate); err != nil {
			return memory, true, false, false, err
		}
		return memory, true, false, true, nil
	}
	if err != nil {
		return StudyMemory{}, false, false, false, err
	}

	exists, err := e.store.EvidenceExists(ctx, existing.ID, candidate.EvidenceSourceType, candidate.EvidenceSourceID)
	if err != nil {
		return StudyMemory{}, false, false, false, err
	}
	if exists {
		return existing, false, false, false, nil
	}

	confidence := math.Min(0.95, existing.Confidence+0.05)
	supportCount := existing.SupportCount + 1
	memory, err := e.store.UpdateMemory(ctx, existing.ID, UpdateMemoryInput{
		Content:        &candidate.Content,
		StructuredData: &candidate.StructuredData,
		Confidence:     &confidence,
		SupportCount:   &supportCount,
		LastSeenAt:     &now,
	})
	if err != nil {
		return StudyMemory{}, false, false, false, err
	}
	if err := e.addEvidence(ctx, memory.ID, candidate); err != nil {
		return memory, false, true, false, err
	}
	return memory, false, true, true, nil
}

func (e *Extractor) addEvidence(ctx context.Context, memoryID int64, candidate memoryCandidate) error {
	excerpt := candidate.EvidenceExcerpt
	_, err := e.store.AddEvidence(ctx, AddEvidenceInput{
		MemoryID:     memoryID,
		SourceType:   candidate.EvidenceSourceType,
		SourceID:     &candidate.EvidenceSourceID,
		EvidenceDate: candidate.EvidenceDate,
		Excerpt:      &excerpt,
		Weight:       candidate.EvidenceWeight,
	})
	return err
}

func (e *Extractor) dedupeCandidates(candidates []memoryCandidate) []memoryCandidate {
	seen := map[string]bool{}
	result := make([]memoryCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		key := fmt.Sprintf("%s|%s|%d|%s", candidate.MemoryType, candidate.ScopeType, int64Value(candidate.ProjectID), candidate.Title)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, candidate)
	}
	return result
}

type memoryCandidate struct {
	MemoryType         string
	ScopeType          string
	ProjectID          *int64
	ProjectName        string
	Title              string
	Content            string
	StructuredData     json.RawMessage
	EvidenceSourceType string
	EvidenceSourceID   int64
	EvidenceDate       string
	EvidenceExcerpt    string
	EvidenceWeight     float64
	InitialConfidence  float64
}

type repeatedNote struct {
	Text  string
	Count int
}

func buildRepeatedBlockerCandidates(summary summaryForExtraction, source map[string]any) []memoryCandidate {
	notes := repeatedNotesFromSource(summary.SummaryType, source)
	topics := map[string]map[string]bool{}
	excerpts := map[string][]string{}
	weight := map[string]float64{}
	for _, note := range notes {
		for topic, keywords := range topicsForText(note.Text) {
			if topics[topic] == nil {
				topics[topic] = map[string]bool{}
			}
			for _, keyword := range keywords {
				topics[topic][keyword] = true
			}
			excerpts[topic] = append(excerpts[topic], note.Text)
			weight[topic] = 0.8
			if note.Count >= 3 {
				weight[topic] = 1.0
			}
		}
	}

	result := make([]memoryCandidate, 0, len(topics))
	for topic, keywordSet := range topics {
		keywords := sortedKeys(keywordSet)
		data := mustJSON(map[string]any{
			"topic":        topic,
			"keywords":     keywords,
			"source":       "repeated_notes",
			"summary_id":   summary.ID,
			"summary_type": summary.SummaryType,
			"note_count":   len(excerpts[topic]),
		})
		result = append(result, memoryCandidate{
			MemoryType:         "repeated_blocker",
			ScopeType:          "topic",
			Title:              "重复卡点：" + topic,
			Content:            fmt.Sprintf("最近 summary 的 repeated_notes 中多次出现 %s 相关问题，说明该主题仍是重复学习阻塞点。", strings.Join(keywords, "、")),
			StructuredData:     data,
			EvidenceSourceType: summarySourceType(summary.SummaryType),
			EvidenceSourceID:   summary.ID,
			EvidenceDate:       evidenceDate(summary),
			EvidenceExcerpt:    strings.Join(excerpts[topic], "；"),
			EvidenceWeight:     weight[topic],
			InitialConfidence:  0.60,
		})
	}
	return result
}

func buildEstimateBiasCandidates(summary summaryForExtraction, source map[string]any) []memoryCandidate {
	rows := projectBreakdownFromSource(summary.SummaryType, source)
	threshold := 30.0
	if summary.SummaryType == "weekly" {
		threshold = 60.0
	}
	result := make([]memoryCandidate, 0)
	for _, row := range rows {
		estimated := number(row, "estimated_minutes")
		actual := number(row, "actual_minutes")
		overrun := number(row, "overrun_minutes")
		rate := number(row, "overrun_rate")
		if overrun == 0 && estimated > 0 && actual > estimated {
			overrun = actual - estimated
		}
		if rate == 0 && estimated > 0 && overrun > 0 {
			rate = overrun / estimated
		}
		if estimated <= 0 || actual <= estimated || (overrun < threshold && rate < 0.4) {
			continue
		}

		projectName := stringValue(row, "project_name", "name")
		projectID := optionalInt64(row, "project_id")
		if projectName == "" && projectID == nil {
			continue
		}
		data := mustJSON(map[string]any{
			"project_name":      projectName,
			"estimated_minutes": int(math.Round(estimated)),
			"actual_minutes":    int(math.Round(actual)),
			"overrun_minutes":   int(math.Round(overrun)),
			"overrun_rate":      rate,
			"summary_id":        summary.ID,
			"summary_type":      summary.SummaryType,
		})
		result = append(result, memoryCandidate{
			MemoryType:         "estimate_bias",
			ScopeType:          "project",
			ProjectID:          projectID,
			ProjectName:        projectName,
			Title:              fmt.Sprintf("估时偏差：%s 经常超时", projectName),
			Content:            "最近 summary 数据显示，该项目实际耗时明显超过预估，说明当前任务拆分或估时偏乐观。",
			StructuredData:     data,
			EvidenceSourceType: summarySourceType(summary.SummaryType),
			EvidenceSourceID:   summary.ID,
			EvidenceDate:       evidenceDate(summary),
			EvidenceExcerpt:    fmt.Sprintf("%s：预计 %d 分钟，实际 %d 分钟，超时 %d 分钟，超时率 %.0f%%。", projectName, int(estimated), int(actual), int(overrun), rate*100),
			EvidenceWeight:     weightByRate(rate),
			InitialConfidence:  0.65,
		})
	}
	return result
}

func buildTimePatternCandidates(summary summaryForExtraction, source map[string]any) []memoryCandidate {
	if summary.SummaryType != "weekly" {
		return nil
	}
	week := mapValue(source, "week")
	distribution := mapValue(week, "time_distribution")
	total := number(week, "total_focus_minutes")
	if total <= 0 {
		total = number(distribution, "morning_minutes") + number(distribution, "afternoon_minutes") + number(distribution, "evening_minutes") + number(distribution, "night_minutes")
	}
	if total < 120 {
		return nil
	}
	periods := map[string]float64{
		"morning":   number(distribution, "morning_minutes"),
		"afternoon": number(distribution, "afternoon_minutes"),
		"evening":   number(distribution, "evening_minutes"),
		"night":     number(distribution, "night_minutes"),
	}
	var dominant string
	var minutes float64
	for period, value := range periods {
		if value > minutes {
			dominant, minutes = period, value
		}
	}
	share := minutes / total
	if share < 0.6 {
		return nil
	}
	label := periodLabel(dominant)
	return []memoryCandidate{{
		MemoryType:         "time_pattern",
		ScopeType:          "global",
		Title:              "时间规律：学习主要集中在" + label,
		Content:            fmt.Sprintf("本周 time_distribution 显示学习时间主要集中在%s，占比超过 60%%，说明当前学习节奏偏向%s时段。", label, label),
		StructuredData:     mustJSON(map[string]any{"dominant_period": dominant, "dominant_period_label": label, "share": share, "total_focus_minutes": int(total), "summary_id": summary.ID, "summary_type": summary.SummaryType}),
		EvidenceSourceType: "weekly_summary",
		EvidenceSourceID:   summary.ID,
		EvidenceDate:       evidenceDate(summary),
		EvidenceExcerpt:    fmt.Sprintf("本周%s学习 %d 分钟，占总学习时长 %.0f%%。", label, int(minutes), share*100),
		EvidenceWeight:     0.8,
		InitialConfidence:  0.55,
	}}
}

func repeatedNotesFromSource(summaryType string, source map[string]any) []repeatedNote {
	if summaryType == "weekly" {
		return parseRepeatedNotes(mapValue(mapValue(source, "week"), "repeated_notes"))
	}
	return parseRepeatedNotes(mapValue(mapValue(source, "recent_context"), "repeated_notes"))
}

func parseRepeatedNotes(value any) []repeatedNote {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]repeatedNote, 0, len(items))
	for _, item := range items {
		switch typed := item.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				result = append(result, repeatedNote{Text: typed, Count: 1})
			}
		case map[string]any:
			text := stringValue(typed, "note", "text", "title", "value")
			if text != "" {
				result = append(result, repeatedNote{Text: text, Count: int(number(typed, "count"))})
			}
		}
	}
	return result
}

func topicsForText(text string) map[string][]string {
	checks := map[string][]string{
		"Go 并发与运行时":  {"Go", "goroutine", "channel", "context", "defer", "panic", "recover", "map", "slice"},
		"SQL 与数据库查询": {"SQL", "JOIN", "MySQL", "migration"},
		"前端与桌面端开发":   {"React", "TypeScript", "Wails"},
		"网络请求与接口稳定性": {"timeout", "network", "API"},
		"后端基础设施":     {"Redis", "RabbitMQ", "JWT", "Docker", "LLM"},
	}
	result := map[string][]string{}
	for topic, keywords := range checks {
		for _, keyword := range keywords {
			if containsKeyword(text, keyword) {
				result[topic] = append(result[topic], keyword)
			}
		}
	}
	return result
}

func projectBreakdownFromSource(summaryType string, source map[string]any) []map[string]any {
	var value any
	if summaryType == "weekly" {
		value = mapValue(mapValue(source, "week"), "project_breakdown")
	} else {
		value = mapValue(mapValue(source, "today"), "project_breakdown")
	}
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]any); ok {
			result = append(result, row)
		}
	}
	return result
}

func mapValue(m any, key string) any {
	values, ok := m.(map[string]any)
	if !ok {
		return nil
	}
	return values[key]
}

func number(m any, keys ...string) float64 {
	values, ok := m.(map[string]any)
	if !ok {
		return 0
	}
	for _, key := range keys {
		switch value := values[key].(type) {
		case float64:
			return value
		case int:
			return float64(value)
		case json.Number:
			v, _ := value.Float64()
			return v
		}
	}
	return 0
}

func stringValue(m any, keys ...string) string {
	values, ok := m.(map[string]any)
	if !ok {
		return ""
	}
	for _, key := range keys {
		if value, ok := values[key].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func optionalInt64(m any, keys ...string) *int64 {
	for _, key := range keys {
		value := number(m, key)
		if value > 0 {
			id := int64(value)
			return &id
		}
	}
	return nil
}

func containsKeyword(text, keyword string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(keyword))
}

func sortedKeys(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func mustJSON(value any) json.RawMessage {
	data, _ := json.Marshal(value)
	return data
}

func summarySourceType(summaryType string) string {
	if summaryType == "weekly" {
		return "weekly_summary"
	}
	return "daily_summary"
}

func evidenceDate(summary summaryForExtraction) string {
	if summary.SummaryType == "weekly" && summary.EndDate != "" {
		return summary.EndDate
	}
	return summary.StartDate
}

func weightByRate(rate float64) float64 {
	if rate >= 0.8 {
		return 1.0
	}
	return 0.8
}

func periodLabel(period string) string {
	switch period {
	case "morning":
		return "上午"
	case "afternoon":
		return "下午"
	case "evening":
		return "晚上"
	case "night":
		return "深夜"
	default:
		return period
	}
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
