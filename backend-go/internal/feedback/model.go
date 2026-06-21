package feedback

import "time"

type Feedback struct {
	ID            int64      `json:"id"`
	TargetType    string     `json:"target_type"`
	TargetID      int64      `json:"target_id"`
	TargetIndex   *int       `json:"target_index"`
	FeedbackValue string     `json:"feedback_value"`
	FeedbackNote  string     `json:"feedback_note"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type SubmitFeedbackRequest struct {
	TargetType    string `json:"target_type"`
	TargetID      int64  `json:"target_id"`
	TargetIndex   *int   `json:"target_index"`
	FeedbackValue string `json:"feedback_value"`
	FeedbackNote  string `json:"feedback_note"`
}

type CreateFeedbackInput struct {
	TargetType    string
	TargetID      int64
	TargetIndex   *int
	FeedbackValue string
	FeedbackNote  string
}

type MemoryFeedbackImpact struct {
	SupportDelta       int
	ContradictionDelta int
	ConfidenceDelta    float64
	ArchiveBelow       *float64
}
