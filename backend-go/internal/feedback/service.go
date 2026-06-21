package feedback

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

const maxFeedbackNoteLength = 1000

var (
	ErrInvalidFeedbackTargetType  = errors.New("invalid feedback target_type")
	ErrInvalidFeedbackTargetID    = errors.New("target_id must be greater than 0")
	ErrInvalidFeedbackTargetIndex = errors.New("action_item feedback requires target_index >= 0")
	ErrInvalidFeedbackValue       = errors.New("invalid feedback_value for target_type")
	ErrFeedbackNoteTooLong        = errors.New("feedback_note is too long")
)

type feedbackRepository interface {
	CreateFeedback(ctx context.Context, input CreateFeedbackInput) (Feedback, error)
	ApplyMemoryFeedback(ctx context.Context, memoryID int64, impact MemoryFeedbackImpact) error
}

type Service struct {
	repo feedbackRepository
}

func NewService(repo feedbackRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SubmitFeedback(ctx context.Context, req SubmitFeedbackRequest) (Feedback, error) {
	input, err := validateFeedback(req)
	if err != nil {
		return Feedback{}, err
	}
	item, err := s.repo.CreateFeedback(ctx, input)
	if err != nil {
		return Feedback{}, err
	}
	if input.TargetType == "memory" {
		impact, ok := memoryFeedbackImpact(input.FeedbackValue)
		if ok {
			if err := s.repo.ApplyMemoryFeedback(ctx, input.TargetID, impact); err != nil {
				return Feedback{}, err
			}
		}
	}
	return item, nil
}

func validateFeedback(req SubmitFeedbackRequest) (CreateFeedbackInput, error) {
	targetType := strings.TrimSpace(req.TargetType)
	value := strings.TrimSpace(req.FeedbackValue)
	note := strings.TrimSpace(req.FeedbackNote)
	if req.TargetID <= 0 {
		return CreateFeedbackInput{}, ErrInvalidFeedbackTargetID
	}
	if !validFeedbackValue(targetType, value) {
		if _, ok := feedbackValues[targetType]; !ok {
			return CreateFeedbackInput{}, ErrInvalidFeedbackTargetType
		}
		return CreateFeedbackInput{}, ErrInvalidFeedbackValue
	}
	if targetType == "action_item" {
		if req.TargetIndex == nil || *req.TargetIndex < 0 {
			return CreateFeedbackInput{}, ErrInvalidFeedbackTargetIndex
		}
	} else {
		req.TargetIndex = nil
	}
	if utf8.RuneCountInString(note) > maxFeedbackNoteLength {
		return CreateFeedbackInput{}, ErrFeedbackNoteTooLong
	}
	return CreateFeedbackInput{
		TargetType:    targetType,
		TargetID:      req.TargetID,
		TargetIndex:   req.TargetIndex,
		FeedbackValue: value,
		FeedbackNote:  note,
	}, nil
}

var feedbackValues = map[string]map[string]bool{
	"summary": {
		"accurate": true, "partially_accurate": true, "inaccurate": true,
	},
	"action_item": {
		"useful": true, "not_useful": true, "already_known": true, "too_vague": true,
	},
	"memory": {
		"correct": true, "outdated": true, "wrong": true, "too_broad": true,
	},
}

func validFeedbackValue(targetType, value string) bool {
	values, ok := feedbackValues[targetType]
	return ok && values[value]
}

func memoryFeedbackImpact(value string) (MemoryFeedbackImpact, bool) {
	archiveBelow := 0.3
	switch value {
	case "correct":
		return MemoryFeedbackImpact{SupportDelta: 1, ConfidenceDelta: 0.05}, true
	case "wrong":
		return MemoryFeedbackImpact{ContradictionDelta: 1, ConfidenceDelta: -0.15, ArchiveBelow: &archiveBelow}, true
	case "outdated":
		return MemoryFeedbackImpact{ContradictionDelta: 1, ConfidenceDelta: -0.10, ArchiveBelow: &archiveBelow}, true
	case "too_broad":
		return MemoryFeedbackImpact{ConfidenceDelta: -0.05}, true
	default:
		return MemoryFeedbackImpact{}, false
	}
}
