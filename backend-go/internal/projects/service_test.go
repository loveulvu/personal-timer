package projects

import (
	"errors"
	"testing"
)

func TestCreateProjectDefaultsCategoryAndSummaryScope(t *testing.T) {
	repo := &fakeProjectRepo{}
	service := &Service{repo: repo}

	id, err := service.CreateProject(CreateProjectRequest{Name: " Go Study "})
	if err != nil {
		t.Fatalf("CreateProject returned error: %v", err)
	}
	if id != 7 {
		t.Fatalf("id = %d, want 7", id)
	}
	if repo.created.Name != "Go Study" {
		t.Fatalf("name = %q, want trimmed name", repo.created.Name)
	}
	if repo.created.Category != "study" {
		t.Fatalf("category = %q, want study", repo.created.Category)
	}
	if !repo.created.IncludeInSummary {
		t.Fatal("include_in_summary = false, want default true")
	}
}

func TestCreateProjectRejectsInvalidCategory(t *testing.T) {
	service := &Service{repo: &fakeProjectRepo{}}

	_, err := service.CreateProject(CreateProjectRequest{Name: "Bad", Category: "food"})
	if !errors.Is(err, ErrInvalidProjectCategory) {
		t.Fatalf("error = %v, want ErrInvalidProjectCategory", err)
	}
}

func TestUpdateProjectSupportsCategoryAndSummaryScope(t *testing.T) {
	include := false
	repo := &fakeProjectRepo{}
	service := &Service{repo: repo}

	if err := service.UpdateProject(3, UpdateProjectRequest{
		Name:             "Dinner",
		Category:         "life",
		IncludeInSummary: &include,
	}); err != nil {
		t.Fatalf("UpdateProject returned error: %v", err)
	}
	if repo.updatedID != 3 {
		t.Fatalf("updated id = %d, want 3", repo.updatedID)
	}
	if repo.updated.Category != "life" {
		t.Fatalf("category = %q, want life", repo.updated.Category)
	}
	if repo.updated.IncludeInSummary {
		t.Fatal("include_in_summary = true, want false")
	}
}

type fakeProjectRepo struct {
	created   CreateProjectInput
	updated   UpdateProjectInput
	updatedID int64
}

func (r *fakeProjectRepo) ListProjects() ([]Project, error) {
	return nil, nil
}

func (r *fakeProjectRepo) CreateProject(input CreateProjectInput) (int64, error) {
	r.created = input
	return 7, nil
}

func (r *fakeProjectRepo) GetProjectByID(id int64) (*Project, error) {
	return nil, nil
}

func (r *fakeProjectRepo) UpdateProject(id int64, input UpdateProjectInput) error {
	r.updatedID = id
	r.updated = input
	return nil
}

func (r *fakeProjectRepo) DeleteProject(id int64) error {
	return nil
}
