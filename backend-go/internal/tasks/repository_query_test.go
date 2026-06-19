package tasks

import (
	"os"
	"strings"
	"testing"
)

func TestEstimatePreviewQueryUsesCompletedTasksAndOverridePrecedence(t *testing.T) {
	source, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	required := []string{
		"dt.status = 'completed'",
		"dt.estimated_minutes > 0",
		"WHEN dt.actual_seconds_override IS NOT NULL AND dt.actual_seconds_override > 0 THEN dt.actual_seconds_override",
		"ELSE COALESCE(ts.total_seconds, 0)",
		"WHERE actual_seconds > 0",
		"ORDER BY completed_at DESC, id DESC",
	}
	for _, expression := range required {
		if !strings.Contains(text, expression) {
			t.Fatalf("estimate query is missing %q", expression)
		}
	}
}
