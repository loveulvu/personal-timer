package projects

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateProject(input CreateProjectInput) (int64, error) {
	query := `
		INSERT INTO projects (name, description, is_fixed)
		VALUES (?, ?, ?)
	`

	result, err := r.db.Exec(query, input.Name, input.Description, input.IsFixed)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
