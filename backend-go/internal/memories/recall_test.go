package memories

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestRecallRelevantMemoriesFiltersMatchesSortsAndLimits(t *testing.T) {
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	projectID := int64(7)
	otherProjectID := int64(8)
	store := &fakeRecallStore{
		projects: map[string]projectForExtraction{
			"Backend Study": {ID: projectID, Name: "Backend Study", IncludeInSummary: true},
		},
		memories: []StudyMemory{
			memoryForRecall(1, "time_pattern", "global", nil, "global", 0.8, "active", now.Add(-time.Hour)),
			memoryForRecall(2, "repeated_blocker", "topic", nil, "topic", 0.9, "active", now.Add(-2*time.Hour)),
			memoryForRecall(3, "estimate_bias", "project", &projectID, "project", 0.9, "active", now.Add(-time.Hour)),
			memoryForRecall(4, "estimate_bias", "project", &otherProjectID, "other project", 0.99, "active", now),
			memoryForRecall(5, "repeated_blocker", "topic", nil, "low confidence", 0.49, "active", now),
			memoryForRecall(6, "repeated_blocker", "topic", nil, "archived", 0.99, "archived", now),
			memoryForRecall(2, "repeated_blocker", "topic", nil, "duplicate", 0.9, "active", now.Add(-2*time.Hour)),
		},
	}

	got, err := NewRecallService(store).RecallRelevantMemories(context.Background(), RecallInput{
		ProjectNames: []string{"Backend Study"},
		Limit:        3,
	})
	if err != nil {
		t.Fatal(err)
	}
	ids := memoryIDs(got)
	want := []int64{3, 2, 1}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("memory ids = %v, want %v", ids, want)
	}
}

func TestRecallRelevantMemoriesSortsTypeAfterConfidenceAndLastSeen(t *testing.T) {
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	store := &fakeRecallStore{memories: []StudyMemory{
		memoryForRecall(1, "time_pattern", "global", nil, "time", 0.8, "active", now),
		memoryForRecall(2, "estimate_bias", "global", nil, "estimate", 0.8, "active", now),
		memoryForRecall(3, "repeated_blocker", "global", nil, "blocker", 0.8, "active", now),
	}}

	got, err := NewRecallService(store).RecallRelevantMemories(context.Background(), RecallInput{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{3, 2, 1}
	if ids := memoryIDs(got); !reflect.DeepEqual(ids, want) {
		t.Fatalf("memory ids = %v, want %v", ids, want)
	}
}

type fakeRecallStore struct {
	projects map[string]projectForExtraction
	memories []StudyMemory
}

func (f *fakeRecallStore) FindProjectForExtraction(ctx context.Context, projectID *int64, name string) (*projectForExtraction, error) {
	project, ok := f.projects[name]
	if !ok {
		return nil, nil
	}
	return &project, nil
}

func (f *fakeRecallStore) ListActiveMemoriesForRecall(ctx context.Context, projectIDs []int64, limit int) ([]StudyMemory, error) {
	return f.memories, nil
}

func memoryForRecall(id int64, memoryType, scopeType string, projectID *int64, title string, confidence float64, status string, lastSeen time.Time) StudyMemory {
	return StudyMemory{
		ID:           id,
		MemoryType:   memoryType,
		ScopeType:    scopeType,
		ProjectID:    projectID,
		Title:        title,
		Content:      title + " content",
		Confidence:   confidence,
		SupportCount: 2,
		LastSeenAt:   lastSeen,
		Status:       status,
	}
}

func memoryIDs(items []StudyMemory) []int64 {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}
