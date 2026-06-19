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
