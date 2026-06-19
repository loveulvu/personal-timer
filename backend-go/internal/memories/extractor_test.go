package memories

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestExtractRepeatedBlockersFromNotes(t *testing.T) {
	store := newFakeExtractionStore()
	store.summaries[1] = fakeSummary(1, "daily", map[string]any{
		"recent_context": map[string]any{
			"repeated_notes": []any{
				"Go channel context timeout",
				map[string]any{"note": "SQL JOIN MySQL", "count": float64(3)},
			},
		},
	})

	result, err := newTestExtractor(store).ExtractFromSummary(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.CreatedCount != 3 || result.EvidenceCount != 3 {
		t.Fatalf("result = %+v, want 3 created/evidence", result)
	}
	assertMemoryTitle(t, store, "重复卡点：Go 并发与运行时")
	assertMemoryTitle(t, store, "重复卡点：SQL 与数据库查询")
	assertMemoryTitle(t, store, "重复卡点：网络请求与接口稳定性")
}

func TestExtractEstimateBiasRespectsThresholdAndProjectScope(t *testing.T) {
	store := newFakeExtractionStore()
	store.projectsByName["Backend Study"] = projectForExtraction{ID: 10, Name: "Backend Study", IncludeInSummary: true}
	store.projectsByName["Meal"] = projectForExtraction{ID: 11, Name: "Meal", IncludeInSummary: false}
	store.summaries[1] = fakeSummary(1, "daily", map[string]any{
		"today": map[string]any{"project_breakdown": []any{
			map[string]any{"project_name": "Backend Study", "estimated_minutes": float64(60), "actual_minutes": float64(110), "overrun_minutes": float64(50)},
			map[string]any{"project_name": "Backend Study", "estimated_minutes": float64(60), "actual_minutes": float64(70), "overrun_minutes": float64(10)},
			map[string]any{"project_name": "Meal", "estimated_minutes": float64(10), "actual_minutes": float64(80), "overrun_minutes": float64(70)},
		}},
	})

	result, err := newTestExtractor(store).ExtractFromSummary(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.CreatedCount != 1 || len(store.memories) != 1 {
		t.Fatalf("created = %d memories=%+v, want one included estimate_bias", result.CreatedCount, store.memories)
	}
	memory := assertMemoryTitle(t, store, "估时偏差：Backend Study 经常超时")
	if memory.ProjectID == nil || *memory.ProjectID != 10 {
		t.Fatalf("project id = %v, want 10", memory.ProjectID)
	}
}

func TestExtractWeeklyTimePatternOnly(t *testing.T) {
	source := map[string]any{"week": map[string]any{
		"total_focus_minutes": float64(200),
		"time_distribution": map[string]any{
			"morning_minutes": float64(20), "afternoon_minutes": float64(140), "evening_minutes": float64(40), "night_minutes": float64(0),
		},
	}}

	weeklyStore := newFakeExtractionStore()
	weeklyStore.summaries[1] = fakeSummary(1, "weekly", source)
	weeklyResult, err := newTestExtractor(weeklyStore).ExtractFromSummary(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if weeklyResult.CreatedCount != 1 {
		t.Fatalf("weekly result = %+v, want one time_pattern", weeklyResult)
	}
	assertMemoryTitle(t, weeklyStore, "时间规律：学习主要集中在下午")

	dailyStore := newFakeExtractionStore()
	dailyStore.summaries[2] = fakeSummary(2, "daily", source)
	dailyResult, err := newTestExtractor(dailyStore).ExtractFromSummary(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if dailyResult.CreatedCount != 0 {
		t.Fatalf("daily result = %+v, want no time_pattern", dailyResult)
	}
}

func TestExtractIsIdempotentForSameSummary(t *testing.T) {
	store := newFakeExtractionStore()
	store.summaries[1] = fakeSummary(1, "daily", map[string]any{
		"recent_context": map[string]any{"repeated_notes": []any{"Go channel context"}},
	})
	extractor := newTestExtractor(store)

	first, err := extractor.ExtractFromSummary(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := extractor.ExtractFromSummary(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	memory := assertMemoryTitle(t, store, "重复卡点：Go 并发与运行时")
	if first.CreatedCount != 1 || second.SkippedCount != 1 || len(store.evidence) != 1 || memory.SupportCount != 1 {
		t.Fatalf("first=%+v second=%+v evidence=%d support=%d, want idempotent", first, second, len(store.evidence), memory.SupportCount)
	}
}

func TestExtractHandlesBadInput(t *testing.T) {
	store := newFakeExtractionStore()
	store.summaries[1] = summaryForExtraction{ID: 1, SummaryType: "daily", SourceData: json.RawMessage(`{bad`)}
	store.summaries[2] = summaryForExtraction{ID: 2, SummaryType: "daily"}
	store.summaries[3] = summaryForExtraction{ID: 3, SummaryType: "daily", SourceData: json.RawMessage(`{}`), ActionItems: json.RawMessage(`null`)}
	extractor := newTestExtractor(store)

	if _, err := extractor.ExtractFromSummary(context.Background(), 999); !errors.Is(err, ErrSummaryNotFound) {
		t.Fatalf("missing summary error = %v, want ErrSummaryNotFound", err)
	}
	if _, err := extractor.ExtractFromSummary(context.Background(), 1); !errors.Is(err, ErrInvalidSourceData) {
		t.Fatalf("bad source error = %v, want ErrInvalidSourceData", err)
	}
	if result, err := extractor.ExtractFromSummary(context.Background(), 2); err != nil || result.CreatedCount != 0 {
		t.Fatalf("empty source result=%+v err=%v, want empty success", result, err)
	}
	if result, err := extractor.ExtractFromSummary(context.Background(), 3); err != nil || result.CreatedCount != 0 {
		t.Fatalf("empty json result=%+v err=%v, want empty success", result, err)
	}
}

func TestExtractConfidenceCapsAtNinetyFive(t *testing.T) {
	store := newFakeExtractionStore()
	store.summaries[1] = fakeSummary(1, "daily", map[string]any{
		"recent_context": map[string]any{"repeated_notes": []any{"Go channel context"}},
	})
	store.summaries[2] = fakeSummary(2, "daily", map[string]any{
		"recent_context": map[string]any{"repeated_notes": []any{"Go channel context again"}},
	})
	extractor := newTestExtractor(store)
	if _, err := extractor.ExtractFromSummary(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	memory := assertMemoryTitle(t, store, "重复卡点：Go 并发与运行时")
	memory.Confidence = 0.94
	store.memories[memory.ID] = memory

	result, err := extractor.ExtractFromSummary(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	memory = assertMemoryTitle(t, store, "重复卡点：Go 并发与运行时")
	if result.UpdatedCount != 1 || memory.Confidence != 0.95 {
		t.Fatalf("result=%+v confidence=%v, want update capped at 0.95", result, memory.Confidence)
	}
}

type fakeExtractionStore struct {
	summaries      map[int64]summaryForExtraction
	projectsByName map[string]projectForExtraction
	projectsByID   map[int64]projectForExtraction
	memories       map[int64]StudyMemory
	evidence       []StudyMemoryEvidence
	nextMemoryID   int64
	nextEvidenceID int64
}

func newFakeExtractionStore() *fakeExtractionStore {
	return &fakeExtractionStore{
		summaries:      map[int64]summaryForExtraction{},
		projectsByName: map[string]projectForExtraction{},
		projectsByID:   map[int64]projectForExtraction{},
		memories:       map[int64]StudyMemory{},
		nextMemoryID:   1,
		nextEvidenceID: 1,
	}
}

func newTestExtractor(store *fakeExtractionStore) *Extractor {
	extractor := NewExtractor(store)
	extractor.now = func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	return extractor
}

func fakeSummary(id int64, summaryType string, source map[string]any) summaryForExtraction {
	data, _ := json.Marshal(source)
	return summaryForExtraction{
		ID:          id,
		SummaryType: summaryType,
		StartDate:   "2026-06-18",
		EndDate:     "2026-06-21",
		SourceData:  data,
	}
}

func (f *fakeExtractionStore) GetSummaryForExtraction(ctx context.Context, id int64) (summaryForExtraction, error) {
	summary, ok := f.summaries[id]
	if !ok {
		return summaryForExtraction{}, ErrSummaryNotFound
	}
	return summary, nil
}

func (f *fakeExtractionStore) FindProjectForExtraction(ctx context.Context, projectID *int64, name string) (*projectForExtraction, error) {
	if projectID != nil {
		project, ok := f.projectsByID[*projectID]
		if !ok {
			for _, item := range f.projectsByName {
				if item.ID == *projectID {
					project = item
					ok = true
					break
				}
			}
		}
		if !ok {
			return nil, nil
		}
		return &project, nil
	}
	project, ok := f.projectsByName[name]
	if !ok {
		return nil, nil
	}
	return &project, nil
}

func (f *fakeExtractionStore) FindActiveMemoryByIdentity(ctx context.Context, memoryType, scopeType string, projectID *int64, title string) (StudyMemory, error) {
	for _, memory := range f.memories {
		if memory.MemoryType == memoryType &&
			memory.ScopeType == scopeType &&
			memory.Title == title &&
			memory.Status == "active" &&
			int64PointerEqual(memory.ProjectID, projectID) {
			return memory, nil
		}
	}
	return StudyMemory{}, ErrMemoryNotFound
}

func (f *fakeExtractionStore) EvidenceExists(ctx context.Context, memoryID int64, sourceType string, sourceID int64) (bool, error) {
	for _, evidence := range f.evidence {
		if evidence.MemoryID == memoryID && evidence.SourceType == sourceType && evidence.SourceID != nil && *evidence.SourceID == sourceID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeExtractionStore) CreateMemory(ctx context.Context, input CreateMemoryInput) (StudyMemory, error) {
	memory := StudyMemory{
		ID:                 f.nextMemoryID,
		MemoryType:         input.MemoryType,
		ScopeType:          input.ScopeType,
		ProjectID:          input.ProjectID,
		Title:              input.Title,
		Content:            input.Content,
		StructuredData:     input.StructuredData,
		Confidence:         input.Confidence,
		SupportCount:       input.SupportCount,
		ContradictionCount: input.ContradictionCount,
		FirstSeenAt:        input.FirstSeenAt,
		LastSeenAt:         input.LastSeenAt,
		Status:             input.Status,
	}
	f.memories[memory.ID] = memory
	f.nextMemoryID++
	return memory, nil
}

func (f *fakeExtractionStore) UpdateMemory(ctx context.Context, id int64, input UpdateMemoryInput) (StudyMemory, error) {
	memory := f.memories[id]
	if input.Content != nil {
		memory.Content = *input.Content
	}
	if input.StructuredData != nil {
		memory.StructuredData = *input.StructuredData
	}
	if input.Confidence != nil {
		memory.Confidence = *input.Confidence
	}
	if input.SupportCount != nil {
		memory.SupportCount = *input.SupportCount
	}
	if input.LastSeenAt != nil {
		memory.LastSeenAt = *input.LastSeenAt
	}
	f.memories[id] = memory
	return memory, nil
}

func (f *fakeExtractionStore) AddEvidence(ctx context.Context, input AddEvidenceInput) (StudyMemoryEvidence, error) {
	evidence := StudyMemoryEvidence{
		ID:           f.nextEvidenceID,
		MemoryID:     input.MemoryID,
		SourceType:   input.SourceType,
		SourceID:     input.SourceID,
		EvidenceDate: input.EvidenceDate,
		Excerpt:      input.Excerpt,
		Weight:       input.Weight,
	}
	f.evidence = append(f.evidence, evidence)
	f.nextEvidenceID++
	return evidence, nil
}

func assertMemoryTitle(t *testing.T, store *fakeExtractionStore, title string) StudyMemory {
	t.Helper()
	for _, memory := range store.memories {
		if memory.Title == title {
			return memory
		}
	}
	t.Fatalf("memory title %q not found in %+v", title, store.memories)
	return StudyMemory{}
}

func int64PointerEqual(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
