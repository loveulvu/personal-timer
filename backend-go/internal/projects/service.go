package projects

import "strings"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}
func (s *Service) ListProjects() ([]Project, error) {
	return s.repo.ListProjects()
}

func (s *Service) CreateProject(req CreateProjectRequest) (int64, error) {
	input := CreateProjectInput{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		IsFixed:     req.IsFixed,
	}

	return s.repo.CreateProject(input)
}
func (s *Service) GetProjectByID(id int64) (*Project, error) {
	return s.repo.GetProjectByID(id)
}

func (s *Service) UpdateProject(id int64, req UpdateProjectRequest) error {
	input := UpdateProjectInput{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		IsFixed:     req.IsFixed,
	}

	return s.repo.UpdateProject(id, input)
}

func (s *Service) DeleteProject(id int64) error {
	return s.repo.DeleteProject(id)
}
