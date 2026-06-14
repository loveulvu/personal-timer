package stats

import (
	"os"
	"strings"
	"testing"
)

func TestStatsQueriesPreferActualSecondsOverride(t *testing.T) {
	source, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	required := []string{
		"COALESCE(dt.actual_seconds_override, COALESCE(SUM(ts.duration_seconds), 0))",
		"COALESCE(dt.actual_seconds_override, ts.total_seconds, 0)",
	}
	for _, expression := range required {
		if !strings.Contains(text, expression) {
			t.Fatalf("stats query is missing override precedence expression %q", expression)
		}
	}
}
