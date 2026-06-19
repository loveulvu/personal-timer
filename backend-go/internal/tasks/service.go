package tasks

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
)

const estimateHistoryLimit = 20

var (
	ErrInvalidProjectID        = errors.New("project_id is required")
	ErrInvalidEstimatedMinutes = errors.New("estimated_minutes must be greater than 0")
	ErrProjectNotFound         = errors.New("project not found")
)

type estimatePreviewRepository interface {
	ProjectExists(ctx context.Context, projectID int64) (bool, error)
	ListEstimateHistorySamples(ctx context.Context, projectID int64, limit int) ([]EstimateHistorySample, error)
}

type Service struct {
	repo estimatePreviewRepository
}

func NewService(repo estimatePreviewRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) EstimatePreview(ctx context.Context, req EstimatePreviewRequest) (*EstimatePreviewResponse, error) {
	if req.ProjectID <= 0 {
		return nil, ErrInvalidProjectID
	}
	if req.EstimatedMinutes <= 0 {
		return nil, ErrInvalidEstimatedMinutes
	}
	req.Title = strings.TrimSpace(req.Title)

	exists, err := s.repo.ProjectExists(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProjectNotFound
	}

	samples, err := s.repo.ListEstimateHistorySamples(ctx, req.ProjectID, estimateHistoryLimit)
	if err != nil {
		return nil, err
	}

	resp := &EstimatePreviewResponse{
		ProjectID:             req.ProjectID,
		InputEstimatedMinutes: req.EstimatedMinutes,
		SampleCount:           len(samples),
		SuggestedMinutes:      req.EstimatedMinutes,
	}
	if len(samples) < 3 {
		resp.RiskLevel = "insufficient_data"
		resp.Reason = "同项目可用历史完成任务少于 3 条，暂时无法可靠判断估时偏差。"
		return resp, nil
	}

	var estimatedTotal, actualTotal int
	for _, sample := range samples {
		estimatedTotal += sample.EstimatedMinutes
		actualTotal += sample.ActualSeconds / 60 // ponytail: floor minutes, matching existing stats integer conversion.
	}
	resp.AvgEstimatedMinutes = int(math.Round(float64(estimatedTotal) / float64(len(samples))))
	resp.AvgActualMinutes = int(math.Round(float64(actualTotal) / float64(len(samples))))
	resp.OverrunRate = round2((float64(resp.AvgActualMinutes) - float64(resp.AvgEstimatedMinutes)) / float64(resp.AvgEstimatedMinutes))
	resp.SuggestedMinutes = roundUpTo5(resp.AvgActualMinutes)
	if resp.SuggestedMinutes < req.EstimatedMinutes {
		resp.SuggestedMinutes = req.EstimatedMinutes
	}
	resp.SplitRecommended = resp.AvgActualMinutes >= 90
	resp.RiskLevel = estimateRiskLevel(req.EstimatedMinutes, resp.AvgActualMinutes)
	resp.Reason = buildEstimateReason(req.EstimatedMinutes, resp)

	return resp, nil
}

func estimateRiskLevel(inputEstimatedMinutes, avgActualMinutes int) string {
	input := float64(inputEstimatedMinutes)
	actual := float64(avgActualMinutes)
	if input < actual*0.7 {
		return "high"
	}
	if input < actual*0.9 {
		return "medium"
	}
	return "low"
}

func buildEstimateReason(inputEstimatedMinutes int, resp *EstimatePreviewResponse) string {
	bias := "高于"
	if resp.OverrunRate < 0 {
		bias = "低于"
	}
	reason := fmt.Sprintf("该项目最近完成任务的平均实际耗时约 %d 分钟，%s平均估时约 %.0f%%。",
		resp.AvgActualMinutes,
		bias,
		math.Abs(resp.OverrunRate*100),
	)
	if resp.RiskLevel == "low" {
		reason += fmt.Sprintf("当前估时 %d 分钟基本接近历史实际耗时。", inputEstimatedMinutes)
	} else {
		reason += fmt.Sprintf("当前估时 %d 分钟偏低，建议提高到约 %d 分钟。", inputEstimatedMinutes, resp.SuggestedMinutes)
	}
	if resp.SplitRecommended {
		reason += "平均实际耗时已达到 90 分钟以上，建议拆分任务。"
	}
	return reason
}

func roundUpTo5(minutes int) int {
	if minutes <= 0 {
		return 0
	}
	return ((minutes + 4) / 5) * 5
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
