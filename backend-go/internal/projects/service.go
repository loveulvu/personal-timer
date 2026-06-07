package projects

import "strings"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProject(req CreateProjectRequest) (int64, error) {
	input := CreateProjectInput{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		IsFixed:     req.IsFixed,
	}

	return s.repo.CreateProject(input)
}
