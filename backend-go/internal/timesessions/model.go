package timesessions

import "time"

type UpdateTimeSessionRequest struct {
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
}

type UpdateTimeSessionInput struct {
	StartedAt       time.Time
	EndedAt         time.Time
	DurationSeconds int
}
