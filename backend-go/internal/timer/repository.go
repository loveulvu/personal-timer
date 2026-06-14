package timer

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrTaskNotFound            = errors.New("task not found")
	ErrTaskMustBePlanned       = errors.New("task status must be planned")
	ErrTaskMustBeRunning       = errors.New("task status must be running")
	ErrTaskMustBePaused        = errors.New("task status must be paused")
	ErrTaskMustBeRunningPaused = errors.New("task status must be running or paused")
	ErrTaskMustBeCompleted     = errors.New("task status must be completed")
	ErrRunningSessionNotFound  = errors.New("running session not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) StartTask(taskID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string

	query := `
		SELECT status
		FROM daily_tasks
		WHERE id = ?
		FOR UPDATE
	`

	err = tx.QueryRow(query, taskID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrTaskNotFound
		}
		return err
	}

	if status != "planned" {
		return ErrTaskMustBePlanned
	}

	insertSessionQuery := `
		INSERT INTO time_sessions (daily_task_id)
		VALUES (?)
	`

	if _, err := tx.Exec(insertSessionQuery, taskID); err != nil {
		return err
	}

	updateTaskQuery := `
		UPDATE daily_tasks
		SET status = 'running'
		WHERE id = ?
	`

	if _, err := tx.Exec(updateTaskQuery, taskID); err != nil {
		return err
	}

	return tx.Commit()
}
func (r *Repository) PauseTask(taskID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `
		SELECT status FROM daily_tasks WHERE id = ? FOR UPDATE
	`
	var status string
	err = tx.QueryRow(query, taskID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrTaskNotFound
		}
		return err
	}
	if status != "running" {
		return ErrTaskMustBeRunning
	}
	var sessionID int64
	var startedAt time.Time
	sessionQuery := `
		SELECT id, started_at FROM time_sessions
   WHERE daily_task_id = ? AND ended_at IS NULL
   ORDER BY id DESC
   LIMIT 1
   FOR UPDATE
	`
	err = tx.QueryRow(sessionQuery, taskID).Scan(&sessionID, &startedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRunningSessionNotFound
		}
		return err
	}
	now := time.Now()
	durationSeconds := int(now.Sub(startedAt).Seconds())

	updateSessionQuery := `
		UPDATE time_sessions
		SET ended_at = ?, duration_seconds = ?
		WHERE id = ?
	`

	if _, err := tx.Exec(updateSessionQuery, now, durationSeconds, sessionID); err != nil {
		return err
	}

	updateTaskQuery := `
		UPDATE daily_tasks
		SET status = 'paused'
		WHERE id = ?
	`

	if _, err := tx.Exec(updateTaskQuery, taskID); err != nil {
		return err
	}

	return tx.Commit()
}
func (r *Repository) ResumeTask(taskID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string

	query := `
		SELECT status
		FROM daily_tasks
		WHERE id = ?
		FOR UPDATE
	`

	err = tx.QueryRow(query, taskID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrTaskNotFound
		}
		return err
	}

	if status != "paused" {
		return ErrTaskMustBePaused
	}

	insertSessionQuery := `
		INSERT INTO time_sessions (daily_task_id)
		VALUES (?)
	`

	if _, err := tx.Exec(insertSessionQuery, taskID); err != nil {
		return err
	}

	updateTaskQuery := `
		UPDATE daily_tasks
		SET status = 'running'
		WHERE id = ?
	`

	if _, err := tx.Exec(updateTaskQuery, taskID); err != nil {
		return err
	}

	return tx.Commit()
}
func (r *Repository) FinishTask(taskID int64, finishNote, finishDescription string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string

	statusQuery := `
		SELECT status
		FROM daily_tasks
		WHERE id = ?
		FOR UPDATE
	`

	err = tx.QueryRow(statusQuery, taskID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrTaskNotFound
		}
		return err
	}

	if status != "running" && status != "paused" {
		return ErrTaskMustBeRunningPaused
	}

	if status == "running" {
		var sessionID int64
		var startedAt time.Time

		sessionQuery := `
			SELECT id, started_at
			FROM time_sessions
			WHERE daily_task_id = ? AND ended_at IS NULL
			ORDER BY id DESC
			LIMIT 1
			FOR UPDATE
		`

		err = tx.QueryRow(sessionQuery, taskID).Scan(&sessionID, &startedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrRunningSessionNotFound
			}
			return err
		}

		now := time.Now()
		durationSeconds := int(now.Sub(startedAt).Seconds())

		updateSessionQuery := `
			UPDATE time_sessions
			SET ended_at = ?, duration_seconds = ?
			WHERE id = ?
		`

		if _, err := tx.Exec(updateSessionQuery, now, durationSeconds, sessionID); err != nil {
			return err
		}
	}

	updateTaskQuery := `
		UPDATE daily_tasks
		SET status = 'completed',
			finish_note = ?,
			finish_description = ?,
			completed_at = ?
		WHERE id = ?
	`

	if _, err := tx.Exec(updateTaskQuery, finishNote, finishDescription, time.Now(), taskID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) UpdateCompletedTask(taskID int64, finishNote, finishDescription string, actualSecondsOverride *int, updateActualOverride bool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	if err := tx.QueryRow(`SELECT status FROM daily_tasks WHERE id = ? FOR UPDATE`, taskID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTaskNotFound
		}
		return err
	}
	if status != "completed" {
		return ErrTaskMustBeCompleted
	}

	query := `
		UPDATE daily_tasks
		SET finish_note = ?, finish_description = ?
		WHERE id = ?
	`
	args := []any{finishNote, finishDescription, taskID}
	if updateActualOverride {
		query = `
			UPDATE daily_tasks
			SET finish_note = ?, finish_description = ?, actual_seconds_override = ?
			WHERE id = ?
		`
		args = []any{finishNote, finishDescription, actualSecondsOverride, taskID}
	}
	if _, err := tx.Exec(query, args...); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) DeleteCompletedTask(taskID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	if err := tx.QueryRow(`SELECT status FROM daily_tasks WHERE id = ? FOR UPDATE`, taskID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTaskNotFound
		}
		return err
	}
	if status != "completed" {
		return ErrTaskMustBeCompleted
	}

	if _, err := tx.Exec(`DELETE FROM time_sessions WHERE daily_task_id = ?`, taskID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM daily_tasks WHERE id = ?`, taskID); err != nil {
		return err
	}

	return tx.Commit()
}
