package dailytasks

import (
	"os"
	"strings"
	"testing"
)

func TestDailyTaskQueriesIncludeRunningSessionStart(t *testing.T) {
	source, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	required := []string{
		"MAX(CASE WHEN ended_at IS NULL THEN started_at END) AS current_session_started_at",
		"WHEN dt.status = 'completed' THEN COALESCE(dt.actual_seconds_override",
	}
	for _, expression := range required {
		if !strings.Contains(text, expression) {
			t.Fatalf("daily task query is missing %q", expression)
		}
	}
}
