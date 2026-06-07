package projects

import (
	"database/sql"
	"errors"
)

var ErrProjectNotFound = errors.New("project not found")

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
func (r *Repository) ListProjects() ([]Project, error) {
	query := `
		SELECT id, name, description, is_fixed, created_at, updated_at
FROM projects
ORDER BY id DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]Project, 0)
	for rows.Next() {
		var p Project

		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.IsFixed,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}
func (r *Repository) GetProjectByID(id int64) (*Project, error) {
	query := `
		SELECT id, name, description, is_fixed, created_at, updated_at
		FROM projects
		WHERE id = ?
	`

	var p Project

	err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.IsFixed,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return &p, nil
}

func (r *Repository) UpdateProject(id int64, input UpdateProjectInput) error {
	query := `UPDATE projects SET name=?, description=?, is_fixed=? WHERE id=?`

	result, err := r.db.Exec(query, input.Name, input.Description, input.IsFixed, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		exists, err := r.projectExists(id)
		if err != nil {
			return err
		}
		if !exists {
			return ErrProjectNotFound
		}
	}

	return nil
}

func (r *Repository) DeleteProject(id int64) error {
	query := `DELETE FROM projects WHERE id=?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *Repository) projectExists(id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = ?)`

	var exists bool
	if err := r.db.QueryRow(query, id).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
