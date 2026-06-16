package summaries

import (
	"os"
	"strings"
	"testing"
)

func TestRecentActiveDateQueryUsesDatesBeforeTargetAndLimit(t *testing.T) {
	source, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	required := []string{
		"WHERE dt.task_date < ?",
		"ORDER BY dt.task_date DESC",
		"LIMIT ?",
	}
	for _, expression := range required {
		if !strings.Contains(text, expression) {
			t.Fatalf("recent active date query is missing %q", expression)
		}
	}
	if strings.Contains(text, "INTERVAL") || strings.Contains(text, "DATE_SUB") {
		t.Fatal("recent active date query should not use a natural-day date subtraction window")
	}
}
