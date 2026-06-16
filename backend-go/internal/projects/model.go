package projects

import "time"

type Project struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	IsFixed          bool      `json:"is_fixed"`
	Category         string    `json:"category"`
	IncludeInSummary bool      `json:"include_in_summary"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateProjectRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	IsFixed          bool   `json:"is_fixed"`
	Category         string `json:"category"`
	IncludeInSummary *bool  `json:"include_in_summary"`
}
type CreateProjectInput struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	IsFixed          bool   `json:"is_fixed"`
	Category         string `json:"category"`
	IncludeInSummary bool   `json:"include_in_summary"`
}

type UpdateProjectRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	IsFixed          bool   `json:"is_fixed"`
	Category         string `json:"category"`
	IncludeInSummary *bool  `json:"include_in_summary"`
}

type UpdateProjectInput struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	IsFixed          bool   `json:"is_fixed"`
	Category         string `json:"category"`
	IncludeInSummary bool   `json:"include_in_summary"`
}
