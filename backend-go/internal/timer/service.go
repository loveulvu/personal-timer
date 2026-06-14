package timer

import (
	"errors"
	"strings"
)

var (
	ErrFinishNoteRequired        = errors.New("finish_note is required")
	ErrFinishDescriptionRequired = errors.New("finish_description is required")
	ErrActualMinutesInvalid      = errors.New("actual_minutes_override must be greater than or equal to 0")
	ErrActualMinutesConflict     = errors.New("actual_minutes_override and clear_actual_minutes_override cannot both be set")
)

type FinishTaskInput struct {
	FinishNote        string `json:"finish_note"`
	FinishDescription string `json:"finish_description"`
}

type UpdateCompletedTaskInput struct {
	FinishNote            string `json:"finish_note"`
	FinishDescription     string `json:"finish_description"`
	ActualMinutesOverride *int   `json:"actual_minutes_override"`
	ClearActualOverride   bool   `json:"clear_actual_minutes_override"`
}

type repository interface {
	StartTask(taskID int64) error
	PauseTask(taskID int64) error
	ResumeTask(taskID int64) error
	FinishTask(taskID int64, finishNote, finishDescription string) error
	UpdateCompletedTask(taskID int64, finishNote, finishDescription string, actualSecondsOverride *int, updateActualOverride bool) error
	DeleteCompletedTask(taskID int64) error
}

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) StartTask(taskID int64) error {
	return s.repo.StartTask(taskID)
}
func (s *Service) PauseTask(taskID int64) error {
	return s.repo.PauseTask(taskID)
}
func (s *Service) ResumeTask(taskID int64) error {
	return s.repo.ResumeTask(taskID)
}
func (s *Service) FinishTask(taskID int64, input FinishTaskInput) error {
	note := strings.TrimSpace(input.FinishNote)
	if note == "" {
		return ErrFinishNoteRequired
	}
	description := strings.TrimSpace(input.FinishDescription)
	if description == "" {
		return ErrFinishDescriptionRequired
	}
	return s.repo.FinishTask(taskID, note, description)
}

func (s *Service) UpdateCompletedTask(taskID int64, input UpdateCompletedTaskInput) error {
	note := strings.TrimSpace(input.FinishNote)
	if note == "" {
		return ErrFinishNoteRequired
	}
	description := strings.TrimSpace(input.FinishDescription)
	if description == "" {
		return ErrFinishDescriptionRequired
	}
	if input.ActualMinutesOverride != nil && *input.ActualMinutesOverride < 0 {
		return ErrActualMinutesInvalid
	}
	if input.ClearActualOverride && input.ActualMinutesOverride != nil {
		return ErrActualMinutesConflict
	}

	var secondsOverride *int
	updateActualOverride := input.ClearActualOverride || input.ActualMinutesOverride != nil
	if input.ActualMinutesOverride != nil {
		seconds := *input.ActualMinutesOverride * 60
		secondsOverride = &seconds
	}
	return s.repo.UpdateCompletedTask(taskID, note, description, secondsOverride, updateActualOverride)
}

func (s *Service) DeleteCompletedTask(taskID int64) error {
	return s.repo.DeleteCompletedTask(taskID)
}
