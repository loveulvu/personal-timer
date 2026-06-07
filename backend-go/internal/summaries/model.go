package summaries

import (
	"encoding/json"
	"time"
)

type GeneratedSummary struct {
	ID          int64           `json:"id"`
	SummaryType string          `json:"summary_type"`
	StartDate   string          `json:"start_date"`
	EndDate     string          `json:"end_date"`
	Content     string          `json:"content"`
	SourceData  json.RawMessage `json:"source_data,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CreateSummaryInput struct {
	SummaryType string
	StartDate   string
	EndDate     string
	Content     string
	SourceData  json.RawMessage
}

type GenerateDailySummaryRequest struct {
	Date string `json:"date"`
}

type GenerateWeeklySummaryRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type GenerateSummaryResult struct {
	SummaryID int64  `json:"summary_id"`
	Content   string `json:"content"`
}
