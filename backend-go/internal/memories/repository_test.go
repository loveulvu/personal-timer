package memories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestRepositoryValidation(t *testing.T) {
	repo := NewRepository(nil)

	badMemory := validMemoryInput()
	badMemory.MemoryType = "bad"
	if _, err := repo.CreateMemory(context.Background(), badMemory); !errors.Is(err, ErrInvalidMemoryInput) {
		t.Fatalf("invalid memory_type error = %v, want ErrInvalidMemoryInput", err)
	}

	badConfidence := validMemoryInput()
	badConfidence.Confidence = 1.5
	if _, err := repo.CreateMemory(context.Background(), badConfidence); !errors.Is(err, ErrInvalidMemoryInput) {
		t.Fatalf("invalid confidence error = %v, want ErrInvalidMemoryInput", err)
	}

	if _, err := repo.AddEvidence(context.Background(), AddEvidenceInput{MemoryID: 1, SourceType: "manual", EvidenceDate: "2026-06-19", Weight: 2}); !errors.Is(err, ErrInvalidEvidenceInput) {
		t.Fatalf("invalid evidence weight error = %v, want ErrInvalidEvidenceInput", err)
	}

	if got := normalizeLimit(0); got != defaultListLimit {
		t.Fatalf("normalizeLimit(0) = %d, want %d", got, defaultListLimit)
	}
	if got := normalizeLimit(999); got != maxListLimit {
		t.Fatalf("normalizeLimit(999) = %d, want %d", got, maxListLimit)
	}
}

func TestRepositoryCRUDWithMySQL(t *testing.T) {
	dsn := os.Getenv("MEMORIES_TEST_DSN")
	if dsn == "" {
		t.Skip("MEMORIES_TEST_DSN not set")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	setupMemoryTestDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()
	projectID := createProject(t, db, "Backend Study", true)

	input := validMemoryInput()
	input.ProjectID = &projectID
	input.StructuredData = json.RawMessage(`{"avg_actual_minutes":65}`)
	memory, err := repo.CreateMemory(ctx, input)
	if err != nil {
		t.Fatalf("CreateMemory error: %v", err)
	}
	if memory.ID == 0 || memory.ProjectID == nil || *memory.ProjectID != projectID {
		t.Fatalf("memory = %+v, want project memory", memory)
	}

	got, err := repo.GetMemoryByID(ctx, memory.ID)
	if err != nil {
		t.Fatalf("GetMemoryByID error: %v", err)
	}
	if got.Title != input.Title {
		t.Fatalf("GetMemoryByID title = %q, want %q", got.Title, input.Title)
	}

	if _, err := repo.CreateMemory(ctx, CreateMemoryInput{
		MemoryType: "time_pattern", ScopeType: "global", Title: "Archived", Content: "old",
		Confidence: 0.5, FirstSeenAt: time.Now(), LastSeenAt: time.Now(), Status: "archived",
	}); err != nil {
		t.Fatalf("Create archived memory error: %v", err)
	}

	active, err := repo.ListMemories(ctx, ListMemoriesFilter{})
	if err != nil {
		t.Fatalf("ListMemories error: %v", err)
	}
	if len(active) != 1 || active[0].ID != memory.ID {
		t.Fatalf("active memories = %+v, want only active memory", active)
	}

	byType, err := repo.ListMemories(ctx, ListMemoriesFilter{MemoryType: "estimate_bias"})
	if err != nil || len(byType) != 1 || byType[0].ID != memory.ID {
		t.Fatalf("ListMemories by type = %+v err=%v", byType, err)
	}
	byScope, err := repo.ListMemories(ctx, ListMemoriesFilter{ScopeType: "project", ProjectID: &projectID})
	if err != nil || len(byScope) != 1 || byScope[0].ID != memory.ID {
		t.Fatalf("ListMemories by scope/project = %+v err=%v", byScope, err)
	}

	confidence := 0.8
	supportCount := 3
	lastSeen := time.Now().Add(time.Hour).Truncate(time.Second)
	updated, err := repo.UpdateMemory(ctx, memory.ID, UpdateMemoryInput{
		Confidence: &confidence, SupportCount: &supportCount, LastSeenAt: &lastSeen,
	})
	if err != nil {
		t.Fatalf("UpdateMemory error: %v", err)
	}
	if updated.Confidence != confidence || updated.SupportCount != supportCount || updated.LastSeenAt.Format(time.DateTime) != lastSeen.Format(time.DateTime) {
		t.Fatalf("updated = %+v, want confidence/support/last_seen updated", updated)
	}

	sourceID := int64(42)
	excerpt := "context/channel repeated"
	evidence, err := repo.AddEvidence(ctx, AddEvidenceInput{
		MemoryID: memory.ID, SourceType: "daily_summary", SourceID: &sourceID,
		EvidenceDate: "2026-06-19", Excerpt: &excerpt, Weight: 0.9,
	})
	if err != nil {
		t.Fatalf("AddEvidence error: %v", err)
	}
	if evidence.ID == 0 || evidence.MemoryID != memory.ID {
		t.Fatalf("evidence = %+v, want evidence for memory", evidence)
	}
	evidenceItems, err := repo.ListEvidence(ctx, memory.ID)
	if err != nil || len(evidenceItems) != 1 {
		t.Fatalf("ListEvidence = %+v err=%v, want one item", evidenceItems, err)
	}

	if err := repo.ArchiveMemory(ctx, memory.ID); err != nil {
		t.Fatalf("ArchiveMemory error: %v", err)
	}
	active, err = repo.ListMemories(ctx, ListMemoriesFilter{})
	if err != nil {
		t.Fatalf("ListMemories after archive error: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("active after archive = %+v, want none", active)
	}

	if _, err := db.Exec(`DELETE FROM study_memories WHERE id = ?`, memory.ID); err != nil {
		t.Fatalf("delete memory error: %v", err)
	}
	evidenceItems, err = repo.ListEvidence(ctx, memory.ID)
	if err != nil {
		t.Fatalf("ListEvidence after delete error: %v", err)
	}
	if len(evidenceItems) != 0 {
		t.Fatalf("evidence after memory delete = %+v, want cascade delete", evidenceItems)
	}
}

func validMemoryInput() CreateMemoryInput {
	now := time.Now().Truncate(time.Second)
	return CreateMemoryInput{
		MemoryType:  "estimate_bias",
		ScopeType:   "project",
		Title:       "Backend study estimate bias",
		Content:     "Backend study tasks often take longer than estimated.",
		Confidence:  0.5,
		FirstSeenAt: now,
		LastSeenAt:  now,
		Status:      "active",
	}
}

func setupMemoryTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, stmt := range []string{
		`SET FOREIGN_KEY_CHECKS = 0`,
		`DROP TABLE IF EXISTS study_memory_evidence`,
		`DROP TABLE IF EXISTS study_memories`,
		`DROP TABLE IF EXISTS projects`,
		`SET FOREIGN_KEY_CHECKS = 1`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("setup statement %q failed: %v", stmt, err)
		}
	}
	execSQLFile(t, db, "../../migrations/001_create_projects.sql")
	execSQLFile(t, db, "../../migrations/007_add_project_category_and_summary_scope.sql")
	execSQLFile(t, db, "../../migrations/009_create_study_memories.sql")
	execSQLFile(t, db, "../../migrations/010_create_study_memory_evidence.sql")
}

func execSQLFile(t *testing.T, db *sql.DB, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range strings.Split(string(data), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec %s failed: %v", path, err)
		}
	}
}

func createProject(t *testing.T, db *sql.DB, name string, includeInSummary bool) int64 {
	t.Helper()
	result, err := db.Exec(`INSERT INTO projects (name, description, is_fixed, category, include_in_summary) VALUES (?, '', false, 'study', ?)`, name, includeInSummary)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
