package tasks

type EstimatePreviewRequest struct {
	ProjectID        int64  `json:"project_id"`
	Title            string `json:"title"`
	EstimatedMinutes int    `json:"estimated_minutes"`
}

type EstimatePreviewResponse struct {
	ProjectID             int64   `json:"project_id"`
	InputEstimatedMinutes int     `json:"input_estimated_minutes"`
	SampleCount           int     `json:"sample_count"`
	AvgEstimatedMinutes   int     `json:"avg_estimated_minutes"`
	AvgActualMinutes      int     `json:"avg_actual_minutes"`
	OverrunRate           float64 `json:"overrun_rate"`
	RiskLevel             string  `json:"risk_level"`
	SuggestedMinutes      int     `json:"suggested_minutes"`
	SplitRecommended      bool    `json:"split_recommended"`
	Reason                string  `json:"reason"`
}

type EstimateHistorySample struct {
	TaskID           int64
	EstimatedMinutes int
	ActualSeconds    int
}
