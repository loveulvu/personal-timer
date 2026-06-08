package timesessions

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrTimeSessionNotFound = errors.New("time session not found")
	ErrTimeSessionRunning  = errors.New("running time session cannot be corrected")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpdateFinishedSession(ctx context.Context, id int64, input UpdateTimeSessionInput) error {
	var originalEndedAt sql.NullTime
	err := r.db.QueryRowContext(
		ctx,
		`SELECT ended_at FROM time_sessions WHERE id = ?`,
		id,
	).Scan(&originalEndedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTimeSessionNotFound
	}
	if err != nil {
		return err
	}
	if !originalEndedAt.Valid {
		return ErrTimeSessionRunning
	}

	_, err = r.db.ExecContext(
		ctx,
		`UPDATE time_sessions
		 SET started_at = ?, ended_at = ?, duration_seconds = ?
		 WHERE id = ?`,
		input.StartedAt,
		input.EndedAt,
		input.DurationSeconds,
		id,
	)
	return err
}
