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
