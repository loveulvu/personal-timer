package feedback

import (
	"os"
	"strings"
	"testing"
)

func TestApplyMemoryFeedbackQueryAdjustsConfidenceAndArchivesLowConfidence(t *testing.T) {
	source, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	required := []string{
		"support_count = support_count + ?",
		"contradiction_count = contradiction_count + ?",
		"confidence = LEAST(1.0, GREATEST(0.0, confidence + ?))",
		"THEN 'archived'",
		"WHERE id = ?",
	}
	for _, expression := range required {
		if !strings.Contains(text, expression) {
			t.Fatalf("memory feedback query is missing %q", expression)
		}
	}
}
