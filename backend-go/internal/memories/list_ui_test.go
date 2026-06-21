package memories

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestNormalizeMemoryItemsFilter(t *testing.T) {
	filter, all, err := normalizeMemoryItemsFilter(ListMemoryItemsFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if filter.Status != "active" || filter.Limit != defaultListLimit || all {
		t.Fatalf("default filter = %+v all=%v, want active/default limit", filter, all)
	}

	filter, all, err = normalizeMemoryItemsFilter(ListMemoryItemsFilter{Status: "all", Limit: 999})
	if err != nil {
		t.Fatal(err)
	}
	if !all || filter.Limit != maxUIListLimit {
		t.Fatalf("all/max filter = %+v all=%v, want all and max %d", filter, all, maxUIListLimit)
	}

	filter, all, err = normalizeMemoryItemsFilter(ListMemoryItemsFilter{Status: "archived", MemoryType: "estimate_bias", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if all || filter.Status != "archived" || filter.MemoryType != "estimate_bias" || filter.Limit != 10 {
		t.Fatalf("archived filter = %+v all=%v", filter, all)
	}

	if _, _, err := normalizeMemoryItemsFilter(ListMemoryItemsFilter{Status: "bad"}); !errors.Is(err, ErrInvalidMemoryInput) {
		t.Fatalf("bad status error = %v, want ErrInvalidMemoryInput", err)
	}
}

func TestMemoryUIListQueryDoesNotChangeRecallFilter(t *testing.T) {
	source, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, expression := range []string{
		"LEFT JOIN projects p ON p.id = m.project_id",
		"p.name AS project_name",
		"m.status = ?",
		"status = 'active'",
		"AND confidence >= ?",
	} {
		if !strings.Contains(text, expression) {
			t.Fatalf("repository query is missing %q", expression)
		}
	}
}
