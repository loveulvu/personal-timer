package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"personal/internal/dailytasks"
	"personal/internal/memories"
	"personal/internal/plans"
	"personal/internal/summaries"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type contextTaskLister struct{}

func (f contextTaskLister) ListDailyTasksByDate(date string) ([]dailytasks.DailyTask, error) {
	projectID := int64(1)
	return []dailytasks.DailyTask{
		{ID: 1, ProjectID: &projectID, TaskDate: date, Title: "Go review", EstimatedMinutes: 60, ActualSeconds: 1800, Status: "planned"},
	}, nil
}

type contextPlanRiskGetter struct{}

func (f contextPlanRiskGetter) GetPlanRisk(ctx context.Context, date string) (*plans.PlanRiskResponse, error) {
	return &plans.PlanRiskResponse{Date: date, PlannedTotalMinutes: 60, RiskLevel: plans.PlanRiskLow}, nil
}

type contextSummaryLister struct{}

func (f contextSummaryLister) ListSummaries(ctx context.Context, summaryType string) ([]summaries.GeneratedSummary, error) {
	return []summaries.GeneratedSummary{
		{
			ID:          10,
			SummaryType: "daily",
			StartDate:   "2026-06-22",
			EndDate:     "2026-06-22",
			Content:     strings.Repeat("summary ", 120),
			ActionItems: json.RawMessage(`[{"title":"Review database indexes"}]`),
			CreatedAt:   time.Date(2026, 6, 22, 20, 0, 0, 0, time.UTC),
		},
	}, nil
}

func (f contextSummaryLister) ListActionItemAcceptances(ctx context.Context, summaryID int64) ([]summaries.ActionItemAcceptance, error) {
	taskID := int64(99)
	return []summaries.ActionItemAcceptance{{ID: 1, SummaryID: summaryID, ItemIndex: 0, TargetDate: "2026-06-23", TargetTaskID: &taskID}}, nil
}

type contextMemoryStore struct{}

func (f contextMemoryStore) ListMemories(ctx context.Context, filter memories.ListMemoriesFilter) ([]memories.StudyMemory, error) {
	return []memories.StudyMemory{
		{ID: 1, MemoryType: "time_pattern", ScopeType: "global", Title: "Good slot", Content: "Evening focus works", Confidence: 0.9, SupportCount: 3, Status: "active", LastSeenAt: time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)},
		{ID: 2, MemoryType: "estimate_bias", ScopeType: "global", Title: "Too weak", Content: "Low confidence", Confidence: 0.3, Status: "active"},
		{ID: 3, MemoryType: "repeated_blocker", ScopeType: "global", Title: "Archived", Content: "Old", Confidence: 0.9, Status: "archived"},
	}, nil
}

func (f contextMemoryStore) ListMemoryEvidence(ctx context.Context, memoryID int64) ([]memories.StudyMemoryEvidence, error) {
	excerpt := strings.Repeat("evidence ", 80)
	return []memories.StudyMemoryEvidence{{ID: 1, MemoryID: memoryID, SourceType: "daily_summary", Excerpt: &excerpt}}, nil
}

func testContextBuilder() *ContextPackBuilder {
	return NewContextPackBuilder(contextTaskLister{}, contextPlanRiskGetter{}, contextSummaryLister{}, contextMemoryStore{})
}

func TestContextPackBuilderRejectsInvalidInput(t *testing.T) {
	_, err := testContextBuilder().Build(context.Background(), ContextPreviewRequest{TargetDate: "2026-06-23"})
	if err != ErrInvalidContextPreviewInput {
		t.Fatalf("empty goal err = %v, want ErrInvalidContextPreviewInput", err)
	}
	_, err = testContextBuilder().Build(context.Background(), ContextPreviewRequest{Goal: "plan", TargetDate: "bad"})
	if err != ErrInvalidContextPreviewInput {
		t.Fatalf("bad date err = %v, want ErrInvalidContextPreviewInput", err)
	}
}

func TestContextPackBuilderDefaultsRecentDaysToFive(t *testing.T) {
	pack, err := testContextBuilder().Build(context.Background(), ContextPreviewRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Build err = %v", err)
	}
	if len(pack.RecentSummaries) != 1 {
		t.Fatalf("recent summaries = %d, want 1", len(pack.RecentSummaries))
	}
}

func TestContextPackBuilderCapsRecentDays(t *testing.T) {
	pack, err := testContextBuilder().Build(context.Background(), ContextPreviewRequest{Goal: "plan", TargetDate: "2026-06-23", RecentDays: 99})
	if err != nil {
		t.Fatalf("Build err = %v", err)
	}
	if !contains(pack.OmittedSections, "recent_days_capped_to_14") {
		t.Fatalf("omitted = %v, want recent_days_capped_to_14", pack.OmittedSections)
	}
}

func TestContextPackBuilderFiltersMemories(t *testing.T) {
	pack, err := testContextBuilder().Build(context.Background(), ContextPreviewRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Build err = %v", err)
	}
	if len(pack.Memories) != 1 || pack.Memories[0].ID != 1 {
		t.Fatalf("memories = %+v, want only high-confidence active memory", pack.Memories)
	}
	if !contains(pack.OmittedSections, "low_confidence_memories_omitted") ||
		!contains(pack.OmittedSections, "archived_memories_omitted") {
		t.Fatalf("omitted = %v, want memory omission reasons", pack.OmittedSections)
	}
}

func TestContextPackBuilderConstraintsIncludeWriteConfirmation(t *testing.T) {
	pack, err := testContextBuilder().Build(context.Background(), ContextPreviewRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Build err = %v", err)
	}
	if !contains(pack.Constraints, "write tools require user confirmation") {
		t.Fatalf("constraints = %v, want write confirmation rule", pack.Constraints)
	}
}

func TestContextPreviewHandlerReturnsContextPack(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(NewToolRegistry(), testContextBuilder())
	router.POST("/api/agent/context-preview", handler.ContextPreview)

	body := bytes.NewBufferString(`{"goal":"plan today","target_date":"2026-06-23"}`)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, "/api/agent/context-preview", body))

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", resp.Code, resp.Body.String())
	}
	var decoded ContextPreviewResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.ContextPack.UserGoal != "plan today" || len(decoded.ContextPack.TodayTasks) != 1 {
		t.Fatalf("context_pack = %+v", decoded.ContextPack)
	}
}

func TestContextPreviewHandlerRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(NewToolRegistry(), testContextBuilder())
	router.POST("/api/agent/context-preview", handler.ContextPreview)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, "/api/agent/context-preview", bytes.NewBufferString(`{"goal":`)))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.Code)
	}
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
