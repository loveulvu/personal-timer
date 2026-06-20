package plans

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

const recentActiveDaysLimit = 5

var ErrInvalidPlanRiskDate = errors.New("date must be YYYY-MM-DD")

type planRiskRepository interface {
	GetPlannedTotalMinutes(ctx context.Context, date string) (int, error)
	ListRecentActiveDayActualMinutes(ctx context.Context, beforeDate string, limit int) ([]ActiveDayActualMinutes, error)
}

type Service struct {
	repo planRiskRepository
}

func NewService(repo planRiskRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetPlanRisk(ctx context.Context, date string) (*PlanRiskResponse, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, ErrInvalidPlanRiskDate
	}

	plannedTotal, err := s.repo.GetPlannedTotalMinutes(ctx, date)
	if err != nil {
		return nil, err
	}
	days, err := s.repo.ListRecentActiveDayActualMinutes(ctx, date, recentActiveDaysLimit)
	if err != nil {
		return nil, err
	}

	return buildPlanRisk(date, plannedTotal, days), nil
}

func buildPlanRisk(date string, plannedTotal int, days []ActiveDayActualMinutes) *PlanRiskResponse {
	resp := &PlanRiskResponse{
		Date:                date,
		PlannedTotalMinutes: plannedTotal,
		RecentActiveDays:    len(days),
	}
	if len(days) < 3 {
		resp.RiskLevel = PlanRiskInsufficientData
		resp.Reason = "最近可用学习记录少于 3 天，暂时无法可靠判断今日计划风险。"
		resp.Suggestions = []string{"先按当前计划执行", "完成后继续积累实际学习数据"}
		return resp
	}

	total := 0
	for _, day := range days {
		total += day.ActualMinutes
	}
	resp.RecentAvgActualMinutes = int(math.Round(float64(total) / float64(len(days))))
	if plannedTotal <= 0 {
		resp.RiskLevel = PlanRiskLow
		resp.Reason = "今日暂无纳入学习统计的计划任务。"
		resp.Suggestions = []string{"可以先添加 1-2 个核心学习任务。"}
		return resp
	}

	resp.PlanRatio = round2(float64(plannedTotal) / float64(resp.RecentAvgActualMinutes))
	switch {
	case resp.PlanRatio > 1.4:
		resp.RiskLevel = PlanRiskHigh
		resp.Reason = fmt.Sprintf("今日计划时长约为近期平均实际学习时长的 %.2f 倍，存在较高完不成风险。", resp.PlanRatio)
		resp.Suggestions = []string{"优先保留 1-2 个核心任务", "将低优先级任务移到明日", "把超过 90 分钟的任务拆分"}
	case resp.PlanRatio > 1.2:
		resp.RiskLevel = PlanRiskMedium
		resp.Reason = fmt.Sprintf("今日计划时长约为近期平均实际学习时长的 %.2f 倍，略高于近期平均水平。", resp.PlanRatio)
		resp.Suggestions = []string{"检查是否有任务估时偏低", "优先完成最重要的任务", "预留 30-60 分钟缓冲时间"}
	default:
		resp.RiskLevel = PlanRiskLow
		resp.Reason = "今日计划负载基本合理。"
		resp.Suggestions = []string{"当前计划负载基本合理", "保持任务完成记录，便于后续优化估时"}
	}
	return resp
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
