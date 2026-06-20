package plans

import (
	"os"
	"strings"
	"testing"
)

func TestPlanRiskQueriesUseSummaryScopeStatusesAndOverridePrecedence(t *testing.T) {
	source, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	required := []string{
		"INNER JOIN projects p ON p.id = dt.project_id",
		"AND COALESCE(p.include_in_summary, TRUE)",
		"dt.status IN ('planned', 'running', 'paused', 'completed')",
		"dt.task_date = ?",
		"dt.task_date < ?",
		"WHEN dt.actual_seconds_override IS NOT NULL AND dt.actual_seconds_override > 0 THEN dt.actual_seconds_override",
		"ELSE COALESCE(ts.total_seconds, 0)",
		"HAVING SUM(actual_seconds) > 0",
		"ORDER BY task_date DESC",
	}
	for _, expression := range required {
		if !strings.Contains(text, expression) {
			t.Fatalf("plan risk query is missing %q", expression)
		}
	}
}
