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
		"p.id IS NOT NULL",
		"COALESCE(p.include_in_summary, TRUE)",
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

func TestActionItemAcceptanceQueriesUseUniqueKeyAndListBySummary(t *testing.T) {
	migration, err := os.ReadFile("../../migrations/012_create_summary_action_item_acceptances.sql")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(migration) + "\n" + string(source)
	for _, expression := range []string{
		"UNIQUE KEY uniq_summary_item_target_date (summary_id, item_index, target_date)",
		"ON DUPLICATE KEY UPDATE",
		"INSERT INTO summary_action_item_acceptances",
		"WHERE summary_id = ?",
		"ORDER BY item_index ASC, id ASC",
	} {
		if !strings.Contains(text, expression) {
			t.Fatalf("acceptance query is missing %q", expression)
		}
	}
}
