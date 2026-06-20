package plans

import (
	"context"
	"errors"
	"testing"
)

func TestPlanRiskInsufficientData(t *testing.T) {
	resp := buildPlanRisk("2026-06-20", 180, []ActiveDayActualMinutes{
		{Date: "2026-06-18", ActualMinutes: 120},
		{Date: "2026-06-19", ActualMinutes: 150},
	})
	if resp.RiskLevel != PlanRiskInsufficientData || resp.PlanRatio != 0 {
		t.Fatalf("response = %+v, want insufficient_data", resp)
	}
}

func TestPlanRiskLevels(t *testing.T) {
	days := []ActiveDayActualMinutes{
		{Date: "2026-06-17", ActualMinutes: 100},
		{Date: "2026-06-18", ActualMinutes: 100},
		{Date: "2026-06-19", ActualMinutes: 100},
	}
	tests := []struct {
		name    string
		planned int
		want    PlanRiskLevel
	}{
		{name: "high", planned: 141, want: PlanRiskHigh},
		{name: "medium", planned: 121, want: PlanRiskMedium},
		{name: "low", planned: 120, want: PlanRiskLow},
		{name: "empty plan", planned: 0, want: PlanRiskLow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := buildPlanRisk("2026-06-20", tt.planned, days)
			if resp.RiskLevel != tt.want {
				t.Fatalf("risk = %s, want %s: %+v", resp.RiskLevel, tt.want, resp)
			}
		})
	}
}

func TestGetPlanRiskQueriesBeforeTargetDate(t *testing.T) {
	repo := &fakePlanRiskRepo{planned: 240, days: []ActiveDayActualMinutes{
		{Date: "2026-06-19", ActualMinutes: 120},
		{Date: "2026-06-18", ActualMinutes: 120},
		{Date: "2026-06-17", ActualMinutes: 120},
	}}
	resp, err := NewService(repo).GetPlanRisk(context.Background(), "2026-06-20")
	if err != nil {
		t.Fatal(err)
	}
	if repo.requestedBeforeDate != "2026-06-20" || repo.requestedLimit != recentActiveDaysLimit {
		t.Fatalf("beforeDate=%s limit=%d, want target date and limit", repo.requestedBeforeDate, repo.requestedLimit)
	}
	if resp.RiskLevel != PlanRiskHigh {
		t.Fatalf("response = %+v, want high risk", resp)
	}
}

func TestGetPlanRiskRejectsBadDate(t *testing.T) {
	_, err := NewService(&fakePlanRiskRepo{}).GetPlanRisk(context.Background(), "2026/06/20")
	if !errors.Is(err, ErrInvalidPlanRiskDate) {
		t.Fatalf("error = %v, want ErrInvalidPlanRiskDate", err)
	}
}

type fakePlanRiskRepo struct {
	planned             int
	days                []ActiveDayActualMinutes
	requestedDate       string
	requestedBeforeDate string
	requestedLimit      int
}

func (r *fakePlanRiskRepo) GetPlannedTotalMinutes(ctx context.Context, date string) (int, error) {
	r.requestedDate = date
	return r.planned, nil
}

func (r *fakePlanRiskRepo) ListRecentActiveDayActualMinutes(ctx context.Context, beforeDate string, limit int) ([]ActiveDayActualMinutes, error) {
	r.requestedBeforeDate = beforeDate
	r.requestedLimit = limit
	return r.days, nil
}
