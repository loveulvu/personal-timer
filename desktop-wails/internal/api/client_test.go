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
