package memories

import (
	"context"
	"sort"
)

const defaultRecallLimit = 8

type recallStore interface {
	FindProjectForExtraction(ctx context.Context, projectID *int64, name string) (*projectForExtraction, error)
	ListActiveMemoriesForRecall(ctx context.Context, projectIDs []int64, limit int) ([]StudyMemory, error)
}

type RecallService struct {
	store recallStore
}

func NewRecallService(store recallStore) *RecallService {
	return &RecallService{store: store}
}

func (s *RecallService) RecallRelevantMemories(ctx context.Context, input RecallInput) ([]StudyMemory, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = defaultRecallLimit
	}

	projectIDs := map[int64]bool{}
	for _, id := range input.ProjectIDs {
		if id > 0 {
			projectIDs[id] = true
		}
	}
	for _, name := range input.ProjectNames {
		project, err := s.store.FindProjectForExtraction(ctx, nil, name)
		if err != nil {
			return nil, err
		}
		if project != nil && project.IncludeInSummary {
			projectIDs[project.ID] = true
		}
	}

	ids := make([]int64, 0, len(projectIDs))
	for id := range projectIDs {
		ids = append(ids, id)
	}
	memories, err := s.store.ListActiveMemoriesForRecall(ctx, ids, maxListLimit)
	if err != nil {
		return nil, err
	}

	seen := map[int64]bool{}
	filtered := make([]StudyMemory, 0, len(memories))
	for _, memory := range memories {
		if seen[memory.ID] || memory.Status != "active" || memory.Confidence < 0.5 || !recallMemoryType(memory.MemoryType) {
			continue
		}
		if memory.ScopeType == "project" && (memory.ProjectID == nil || !projectIDs[*memory.ProjectID]) {
			continue
		}
		seen[memory.ID] = true
		filtered = append(filtered, memory)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].Confidence != filtered[j].Confidence {
			return filtered[i].Confidence > filtered[j].Confidence
		}
		if !filtered[i].LastSeenAt.Equal(filtered[j].LastSeenAt) {
			return filtered[i].LastSeenAt.After(filtered[j].LastSeenAt)
		}
		return recallTypeRank(filtered[i].MemoryType) < recallTypeRank(filtered[j].MemoryType)
	})
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func recallMemoryType(value string) bool {
	switch value {
	case "repeated_blocker", "estimate_bias", "time_pattern":
		return true
	default:
		return false
	}
}

func recallTypeRank(value string) int {
	switch value {
	case "repeated_blocker":
		return 0
	case "estimate_bias":
		return 1
	case "time_pattern":
		return 2
	default:
		return 9
	}
}
