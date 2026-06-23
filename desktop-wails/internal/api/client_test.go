package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGenerateWeeklySummaryUsesWeeklyRouteAndDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/summaries/weekly/generate" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var request GenerateWeeklySummaryRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.StartDate != "2026-06-08" || request.EndDate != "2026-06-14" {
			t.Fatalf("unexpected dates: %#v", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","data":{"summary_id":7,"content":"weekly"}}`))
	}))
	defer server.Close()

	result, err := NewClient(server.URL).GenerateWeeklySummary(context.Background(), "2026-06-08", "2026-06-14")
	if err != nil {
		t.Fatal(err)
	}
	if result.SummaryID != 7 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestEstimateTaskPreviewUsesRouteAndReturnsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tasks/estimate-preview" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var request EstimatePreviewRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.ProjectID != 3 || request.EstimatedMinutes != 45 {
			t.Fatalf("unexpected request: %#v", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","data":{"project_id":3,"input_estimated_minutes":45,"sample_count":8,"avg_estimated_minutes":50,"avg_actual_minutes":78,"overrun_rate":0.56,"risk_level":"high","suggested_minutes":80,"split_recommended":false,"reason":"ok"}}`))
	}))
	defer server.Close()

	result, err := NewClient(server.URL).EstimateTaskPreview(context.Background(), EstimatePreviewRequest{
		ProjectID:        3,
		Title:            "task",
		EstimatedMinutes: 45,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.RiskLevel != "high" || result.SuggestedMinutes != 80 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestGetPlanRiskUsesDateQueryAndReturnsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/plans/risk" || r.URL.Query().Get("date") != "2026-06-20" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","data":{"date":"2026-06-20","planned_total_minutes":360,"recent_avg_actual_minutes":220,"recent_active_days":5,"plan_ratio":1.64,"risk_level":"high","reason":"ok","suggestions":["a"]}}`))
	}))
	defer server.Close()

	result, err := NewClient(server.URL).GetPlanRisk(context.Background(), "2026-06-20")
	if err != nil {
		t.Fatal(err)
	}
	if result.RiskLevel != "high" || result.PlanRatio != 1.64 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestSubmitFeedbackUsesRouteAndJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/feedback" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var request FeedbackRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.TargetType != "action_item" || request.TargetID != 7 || request.TargetIndex == nil || *request.TargetIndex != 2 || request.FeedbackValue != "useful" {
			t.Fatalf("unexpected request: %#v", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","data":{"id":3,"target_type":"action_item","target_id":7,"target_index":2,"feedback_value":"useful","feedback_note":"","created_at":"2026-06-20T12:00:00Z"}}`))
	}))
	defer server.Close()

	index := 2
	result, err := NewClient(server.URL).SubmitFeedback(context.Background(), FeedbackRequest{
		TargetType:    "action_item",
		TargetID:      7,
		TargetIndex:   &index,
		FeedbackValue: "useful",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 3 || result.TargetIndex == nil || *result.TargetIndex != 2 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestListMemoriesUsesQueryAndReturnsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/memories" ||
			r.URL.Query().Get("status") != "archived" ||
			r.URL.Query().Get("memory_type") != "estimate_bias" ||
			r.URL.Query().Get("limit") != "25" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","data":[{"id":1,"memory_type":"estimate_bias","scope_type":"project","project_id":2,"project_name":"Backend","title":"t","content":"c","confidence":0.7,"support_count":2,"contradiction_count":1,"status":"archived","first_seen_at":"2026-06-19T00:00:00Z","last_seen_at":"2026-06-20T00:00:00Z","created_at":"2026-06-19T00:00:00Z","updated_at":"2026-06-20T00:00:00Z"}]}`))
	}))
	defer server.Close()

	result, err := NewClient(server.URL).ListMemories(context.Background(), "archived", "estimate_bias", 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].ProjectName == nil || *result[0].ProjectName != "Backend" {
		t.Fatalf("unexpected memories: %#v", result)
	}
}

func TestListMemoryEvidenceUsesPathAndReturnsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/memories/9/evidence" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","data":[{"id":1,"memory_id":9,"source_type":"daily_summary","source_id":12,"evidence_date":"2026-06-20","excerpt":"Go context repeated","weight":1,"created_at":"2026-06-20T12:00:00Z"}]}`))
	}))
	defer server.Close()

	result, err := NewClient(server.URL).ListMemoryEvidence(context.Background(), 9)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].SourceID == nil || *result[0].SourceID != 12 {
		t.Fatalf("unexpected evidence: %#v", result)
	}
}

func TestAcceptSummaryActionItemParsesAcceptanceFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/summaries/7/action-items/2/accept" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","data":{"summary_id":7,"item_index":2,"target_date":"2026-06-22","target_task_id":123,"created":true,"already_exists":false,"acceptance_status":"accepted"}}`))
	}))
	defer server.Close()

	result, err := NewClient(server.URL).AcceptSummaryActionItem(context.Background(), 7, 2, "2026-06-22")
	if err != nil {
		t.Fatal(err)
	}
	if result.SummaryID != 7 || result.ItemIndex != 2 || result.TargetTaskID == nil || *result.TargetTaskID != 123 || result.AcceptanceStatus != "accepted" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestListActionItemAcceptancesUsesPathAndReturnsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/summaries/7/action-item-acceptances" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","data":[{"id":1,"summary_id":7,"item_index":0,"target_date":"2026-06-22","target_task_id":123,"status":"accepted","created_at":"2026-06-21T12:00:00Z"}]}`))
	}))
	defer server.Close()

	result, err := NewClient(server.URL).ListActionItemAcceptances(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].TargetTaskID == nil || *result[0].TargetTaskID != 123 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestListDailyTasksPreservesCurrentSessionStartedAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","data":[{"id":1,"project_id":1,"task_date":"2026-06-15","title":"test","estimated_minutes":30,"status":"running","actual_seconds":120,"current_session_started_at":"2026-06-15T10:00:00+08:00"}]}`))
	}))
	defer server.Close()

	tasks, err := NewClient(server.URL).ListDailyTasks(context.Background(), "2026-06-15")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].CurrentSessionStartedAt == nil {
		t.Fatalf("expected current session started_at, got %#v", tasks)
	}
	if got := tasks[0].CurrentSessionStartedAt.Format(time.RFC3339); got != "2026-06-15T10:00:00+08:00" {
		t.Fatalf("unexpected current session started_at: %s", got)
	}
}

func TestDoJSONReportsTimeoutInsteadOfBackendOffline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.httpClient.Timeout = time.Millisecond
	_, err := client.GenerateWeeklySummary(context.Background(), "2026-06-08", "2026-06-14")
	if err == nil || err.Error() != "backend request timed out" {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestDoJSONPreservesBackendError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"start_date must be YYYY-MM-DD"}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).GenerateWeeklySummary(context.Background(), "", "")
	if err == nil || !strings.Contains(err.Error(), "start_date must be YYYY-MM-DD") {
		t.Fatalf("expected backend validation error, got %v", err)
	}
}
