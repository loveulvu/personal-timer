package projects

import (
	"errors"
	"strings"
)

const DefaultProjectCategory = "study"

var ErrInvalidProjectCategory = errors.New("invalid project category")

type Service struct {
	repo projectRepository
}

type projectRepository interface {
	ListProjects() ([]Project, error)
	CreateProject(input CreateProjectInput) (int64, error)
	GetProjectByID(id int64) (*Project, error)
	UpdateProject(id int64, input UpdateProjectInput) error
	DeleteProject(id int64) error
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}
func (s *Service) ListProjects() ([]Project, error) {
	return s.repo.ListProjects()
}

func (s *Service) CreateProject(req CreateProjectRequest) (int64, error) {
	category, err := normalizeProjectCategory(req.Category)
	if err != nil {
		return 0, err
	}
	includeInSummary := true
	if req.IncludeInSummary != nil {
		includeInSummary = *req.IncludeInSummary
	}

	input := CreateProjectInput{
		Name:             strings.TrimSpace(req.Name),
		Description:      req.Description,
		IsFixed:          req.IsFixed,
		Category:         category,
		IncludeInSummary: includeInSummary,
	}

	return s.repo.CreateProject(input)
}
func (s *Service) GetProjectByID(id int64) (*Project, error) {
	return s.repo.GetProjectByID(id)
}

func (s *Service) UpdateProject(id int64, req UpdateProjectRequest) error {
	category, err := normalizeProjectCategory(req.Category)
	if err != nil {
		return err
	}
	includeInSummary := true
	if req.IncludeInSummary != nil {
		includeInSummary = *req.IncludeInSummary
	}

	input := UpdateProjectInput{
		Name:             strings.TrimSpace(req.Name),
		Description:      req.Description,
		IsFixed:          req.IsFixed,
		Category:         category,
		IncludeInSummary: includeInSummary,
	}

	return s.repo.UpdateProject(id, input)
}

func (s *Service) DeleteProject(id int64) error {
	return s.repo.DeleteProject(id)
}

func normalizeProjectCategory(category string) (string, error) {
	value := strings.TrimSpace(category)
	if value == "" {
		return DefaultProjectCategory, nil
	}
	switch value {
	case "study", "project", "life", "break", "other":
		return value, nil
	default:
		return "", ErrInvalidProjectCategory
	}
}
