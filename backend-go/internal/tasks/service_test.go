package tasks

import (
	"context"
	"errors"
	"math"
	"testing"
)

func TestEstimatePreviewInsufficientData(t *testing.T) {
	service := NewService(&fakeEstimateRepo{
		projectExists: true,
		samples: []EstimateHistorySample{
			{TaskID: 1, EstimatedMinutes: 30, ActualSeconds: 2400},
			{TaskID: 2, EstimatedMinutes: 40, ActualSeconds: 3000},
		},
	})

	resp, err := service.EstimatePreview(context.Background(), EstimatePreviewRequest{ProjectID: 7, EstimatedMinutes: 45})
	if err != nil {
		t.Fatalf("EstimatePreview returned error: %v", err)
	}
	if resp.RiskLevel != "insufficient_data" || resp.SuggestedMinutes != 45 || resp.SampleCount != 2 {
		t.Fatalf("response = %+v, want insufficient data with input suggested minutes", resp)
	}
}

func TestEstimatePreviewHighRiskSuggestsRoundedActualAverage(t *testing.T) {
	service := NewService(&fakeEstimateRepo{
		projectExists: true,
		samples: []EstimateHistorySample{
			{TaskID: 1, EstimatedMinutes: 50, ActualSeconds: 4680},
			{TaskID: 2, EstimatedMinutes: 50, ActualSeconds: 4680},
			{TaskID: 3, EstimatedMinutes: 50, ActualSeconds: 4680},
		},
	})

	resp, err := service.EstimatePreview(context.Background(), EstimatePreviewRequest{ProjectID: 7, EstimatedMinutes: 45})
	if err != nil {
		t.Fatalf("EstimatePreview returned error: %v", err)
	}
	if resp.RiskLevel != "high" || resp.AvgEstimatedMinutes != 50 || resp.AvgActualMinutes != 78 || resp.SuggestedMinutes != 80 {
		t.Fatalf("response = %+v, want high risk with 80 minute suggestion", resp)
	}
	if math.Abs(resp.OverrunRate-0.56) > 0.001 {
		t.Fatalf("overrun_rate = %v, want 0.56", resp.OverrunRate)
	}
}

func TestEstimatePreviewSplitRecommendedWhenAverageActualAtLeast90(t *testing.T) {
	service := NewService(&fakeEstimateRepo{
		projectExists: true,
		samples: []EstimateHistorySample{
			{TaskID: 1, EstimatedMinutes: 60, ActualSeconds: 5400},
			{TaskID: 2, EstimatedMinutes: 60, ActualSeconds: 5400},
			{TaskID: 3, EstimatedMinutes: 60, ActualSeconds: 5400},
		},
	})

	resp, err := service.EstimatePreview(context.Background(), EstimatePreviewRequest{ProjectID: 7, EstimatedMinutes: 90})
	if err != nil {
		t.Fatalf("EstimatePreview returned error: %v", err)
	}
	if !resp.SplitRecommended || resp.AvgActualMinutes != 90 {
		t.Fatalf("response = %+v, want split recommendation at 90 minutes", resp)
	}
}

func TestEstimatePreviewRejectsInvalidEstimatedMinutes(t *testing.T) {
	service := NewService(&fakeEstimateRepo{projectExists: true})

	_, err := service.EstimatePreview(context.Background(), EstimatePreviewRequest{ProjectID: 7, EstimatedMinutes: 0})
	if !errors.Is(err, ErrInvalidEstimatedMinutes) {
		t.Fatalf("error = %v, want ErrInvalidEstimatedMinutes", err)
	}
}

func TestEstimatePreviewReturnsProjectNotFound(t *testing.T) {
	service := NewService(&fakeEstimateRepo{projectExists: false})

	_, err := service.EstimatePreview(context.Background(), EstimatePreviewRequest{ProjectID: 7, EstimatedMinutes: 30})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("error = %v, want ErrProjectNotFound", err)
	}
}

type fakeEstimateRepo struct {
	projectExists bool
	samples       []EstimateHistorySample
}

func (r *fakeEstimateRepo) ProjectExists(ctx context.Context, projectID int64) (bool, error) {
	return r.projectExists, nil
}

func (r *fakeEstimateRepo) ListEstimateHistorySamples(ctx context.Context, projectID int64, limit int) ([]EstimateHistorySample, error) {
	if limit != estimateHistoryLimit {
		return nil, errors.New("unexpected limit")
	}
	return r.samples, nil
}
