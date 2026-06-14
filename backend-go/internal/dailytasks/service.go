package dailytasks

import (
	"errors"
	"strings"
)

var (
	ErrInvalidDailyTaskStatus           = errors.New("invalid daily task status")
	ErrInvalidDailyTaskStatusTransition = errors.New("daily task status must be changed through timer endpoints")
	ErrDailyTaskMustBeCompleted         = errors.New("daily task status must be completed")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateDailyTask(req CreateDailyTaskRequest) (int64, error) {
	input := CreateDailyTaskInput{
		ProjectID:        req.ProjectID,
		TaskDate:         req.TaskDate,
		Title:            strings.TrimSpace(req.Title),
		EstimatedMinutes: req.EstimatedMinutes,
	}

	return s.repo.CreateDailyTask(input)
}

func (s *Service) ListDailyTasksByDate(date string) ([]DailyTask, error) {
	return s.repo.ListDailyTasksByDate(date)
}

func (s *Service) GetDailyTaskByID(id int64) (*DailyTask, error) {
	return s.repo.GetDailyTaskByID(id)
}

func (s *Service) UpdateDailyTask(id int64, req UpdateDailyTaskRequest) error {
	currentTask, err := s.repo.GetDailyTaskByID(id)
	if err != nil {
		return err
	}

	status := strings.TrimSpace(req.Status)
	if !isValidDailyTaskStatus(status) {
		return ErrInvalidDailyTaskStatus
	}
	if status != currentTask.Status && !isManualStatusTransitionAllowed(currentTask.Status, status) {
		return ErrInvalidDailyTaskStatusTransition
	}

	input := UpdateDailyTaskInput{
		ProjectID:        req.ProjectID,
		TaskDate:         req.TaskDate,
		Title:            strings.TrimSpace(req.Title),
		EstimatedMinutes: req.EstimatedMinutes,
		Status:           status,
	}

	return s.repo.UpdateDailyTask(id, input)
}

func (s *Service) DeleteDailyTask(id int64) error {
	task, err := s.repo.GetDailyTaskByID(id)
	if err != nil {
		return err
	}
	if task.Status != "completed" {
		return ErrDailyTaskMustBeCompleted
	}
	return s.repo.DeleteDailyTask(id)
}

func isValidDailyTaskStatus(status string) bool {
	switch status {
	case "planned", "running", "paused", "completed", "cancelled":
		return true
	default:
		return false
	}
}

func isManualStatusTransitionAllowed(currentStatus, newStatus string) bool {
	return currentStatus == "planned" && newStatus == "cancelled" ||
		currentStatus == "cancelled" && newStatus == "planned"
}
