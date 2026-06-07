package dailytasks

import "strings"

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
	input := UpdateDailyTaskInput{
		ProjectID:        req.ProjectID,
		TaskDate:         req.TaskDate,
		Title:            strings.TrimSpace(req.Title),
		EstimatedMinutes: req.EstimatedMinutes,
	}

	return s.repo.UpdateDailyTask(id, input)
}

func (s *Service) DeleteDailyTask(id int64) error {
	return s.repo.DeleteDailyTask(id)
}
