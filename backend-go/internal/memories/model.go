package memories

import (
	"encoding/json"
	"time"
)

type StudyMemory struct {
	ID                 int64           `json:"id"`
	MemoryType         string          `json:"memory_type"`
	ScopeType          string          `json:"scope_type"`
	ProjectID          *int64          `json:"project_id,omitempty"`
	Title              string          `json:"title"`
	Content            string          `json:"content"`
	StructuredData     json.RawMessage `json:"structured_data,omitempty"`
	Confidence         float64         `json:"confidence"`
	SupportCount       int             `json:"support_count"`
	ContradictionCount int             `json:"contradiction_count"`
	FirstSeenAt        time.Time       `json:"first_seen_at"`
	LastSeenAt         time.Time       `json:"last_seen_at"`
	Status             string          `json:"status"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type StudyMemoryEvidence struct {
	ID           int64     `json:"id"`
	MemoryID     int64     `json:"memory_id"`
	SourceType   string    `json:"source_type"`
	SourceID     *int64    `json:"source_id,omitempty"`
	EvidenceDate string    `json:"evidence_date"`
	Excerpt      *string   `json:"excerpt,omitempty"`
	Weight       float64   `json:"weight"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateMemoryInput struct {
	MemoryType         string
	ScopeType          string
	ProjectID          *int64
	Title              string
	Content            string
	StructuredData     json.RawMessage
	Confidence         float64
	SupportCount       int
	ContradictionCount int
	FirstSeenAt        time.Time
	LastSeenAt         time.Time
	Status             string
}

type UpdateMemoryInput struct {
	Title              *string
	Content            *string
	StructuredData     *json.RawMessage
	Confidence         *float64
	SupportCount       *int
	ContradictionCount *int
	LastSeenAt         *time.Time
	Status             *string
}

type AddEvidenceInput struct {
	MemoryID     int64
	SourceType   string
	SourceID     *int64
	EvidenceDate string
	Excerpt      *string
	Weight       float64
}

type ListMemoriesFilter struct {
	MemoryType string
	ScopeType  string
	ProjectID  *int64
	Status     string
	Limit      int
}

type ExtractionResult struct {
	SummaryID     int64         `json:"summary_id"`
	CreatedCount  int           `json:"created_count"`
	UpdatedCount  int           `json:"updated_count"`
	SkippedCount  int           `json:"skipped_count"`
	EvidenceCount int           `json:"evidence_count"`
	Memories      []StudyMemory `json:"memories"`
	Warnings      []string      `json:"warnings,omitempty"`
}

type RecallInput struct {
	SummaryType  string
	ProjectIDs   []int64
	ProjectNames []string
	Limit        int
}
