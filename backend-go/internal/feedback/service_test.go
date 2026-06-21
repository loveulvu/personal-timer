package feedback

import (
	"context"
	"errors"
	"testing"
)

func TestSubmitFeedbackRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		req  SubmitFeedbackRequest
		want error
	}{
		{name: "bad type", req: SubmitFeedbackRequest{TargetType: "task", TargetID: 1, FeedbackValue: "useful"}, want: ErrInvalidFeedbackTargetType},
		{name: "bad value", req: SubmitFeedbackRequest{TargetType: "summary", TargetID: 1, FeedbackValue: "useful"}, want: ErrInvalidFeedbackValue},
		{name: "missing action item index", req: SubmitFeedbackRequest{TargetType: "action_item", TargetID: 1, FeedbackValue: "useful"}, want: ErrInvalidFeedbackTargetIndex},
		{name: "bad id", req: SubmitFeedbackRequest{TargetType: "summary", FeedbackValue: "accurate"}, want: ErrInvalidFeedbackTargetID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(&fakeFeedbackRepo{}).SubmitFeedback(context.Background(), tt.req)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestSubmitFeedbackSavesSummaryAndActionItem(t *testing.T) {
	index := 2
	repo := &fakeFeedbackRepo{}
	service := NewService(repo)
	if _, err := service.SubmitFeedback(context.Background(), SubmitFeedbackRequest{TargetType: "summary", TargetID: 7, FeedbackValue: "accurate"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SubmitFeedback(context.Background(), SubmitFeedbackRequest{TargetType: "action_item", TargetID: 7, TargetIndex: &index, FeedbackValue: "useful"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.created) != 2 || repo.created[1].TargetIndex == nil || *repo.created[1].TargetIndex != 2 {
		t.Fatalf("created = %+v, want summary and action item feedback", repo.created)
	}
	if len(repo.memoryImpacts) != 0 {
		t.Fatalf("memory impacts = %+v, want none", repo.memoryImpacts)
	}
}

func TestMemoryFeedbackImpacts(t *testing.T) {
	tests := []struct {
		value           string
		supportDelta    int
		contraDelta     int
		confidenceDelta float64
		archive         bool
	}{
		{value: "correct", supportDelta: 1, confidenceDelta: 0.05},
		{value: "wrong", contraDelta: 1, confidenceDelta: -0.15, archive: true},
		{value: "outdated", contraDelta: 1, confidenceDelta: -0.10, archive: true},
		{value: "too_broad", confidenceDelta: -0.05},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			impact, ok := memoryFeedbackImpact(tt.value)
			if !ok {
				t.Fatal("impact not found")
			}
			if impact.SupportDelta != tt.supportDelta || impact.ContradictionDelta != tt.contraDelta || impact.ConfidenceDelta != tt.confidenceDelta || (impact.ArchiveBelow != nil) != tt.archive {
				t.Fatalf("impact = %+v, want support=%d contradiction=%d confidence=%v archive=%v", impact, tt.supportDelta, tt.contraDelta, tt.confidenceDelta, tt.archive)
			}
		})
	}
}

func TestSubmitFeedbackAppliesMemoryFeedback(t *testing.T) {
	repo := &fakeFeedbackRepo{}
	_, err := NewService(repo).SubmitFeedback(context.Background(), SubmitFeedbackRequest{TargetType: "memory", TargetID: 9, FeedbackValue: "wrong"})
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.created) != 1 || len(repo.memoryImpacts) != 1 || repo.memoryIDs[0] != 9 {
		t.Fatalf("created=%+v memoryIDs=%+v impacts=%+v, want saved feedback and memory impact", repo.created, repo.memoryIDs, repo.memoryImpacts)
	}
	if repo.memoryImpacts[0].ContradictionDelta != 1 || repo.memoryImpacts[0].ConfidenceDelta != -0.15 || repo.memoryImpacts[0].ArchiveBelow == nil {
		t.Fatalf("impact = %+v, want wrong memory impact", repo.memoryImpacts[0])
	}
}

type fakeFeedbackRepo struct {
	created       []CreateFeedbackInput
	memoryIDs     []int64
	memoryImpacts []MemoryFeedbackImpact
}

func (r *fakeFeedbackRepo) CreateFeedback(ctx context.Context, input CreateFeedbackInput) (Feedback, error) {
	r.created = append(r.created, input)
	return Feedback{ID: int64(len(r.created)), TargetType: input.TargetType, TargetID: input.TargetID, TargetIndex: input.TargetIndex, FeedbackValue: input.FeedbackValue, FeedbackNote: input.FeedbackNote}, nil
}

func (r *fakeFeedbackRepo) ApplyMemoryFeedback(ctx context.Context, memoryID int64, impact MemoryFeedbackImpact) error {
	r.memoryIDs = append(r.memoryIDs, memoryID)
	r.memoryImpacts = append(r.memoryImpacts, impact)
	return nil
}
