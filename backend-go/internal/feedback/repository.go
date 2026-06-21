package feedback

import (
	"context"
	"database/sql"
	"errors"
)

var ErrFeedbackTargetNotFound = errors.New("feedback target not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateFeedback(ctx context.Context, input CreateFeedbackInput) (Feedback, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO study_feedback (target_type, target_id, target_index, feedback_value, feedback_note)
		VALUES (?, ?, ?, ?, ?)
	`, input.TargetType, input.TargetID, input.TargetIndex, input.FeedbackValue, input.FeedbackNote)
	if err != nil {
		return Feedback{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Feedback{}, err
	}
	return r.GetFeedbackByID(ctx, id)
}

func (r *Repository) GetFeedbackByID(ctx context.Context, id int64) (Feedback, error) {
	var item Feedback
	err := r.db.QueryRowContext(ctx, `
		SELECT id, target_type, target_id, target_index, feedback_value, feedback_note, created_at, updated_at
		FROM study_feedback
		WHERE id = ?
	`, id).Scan(&item.ID, &item.TargetType, &item.TargetID, &item.TargetIndex, &item.FeedbackValue, &item.FeedbackNote, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) ApplyMemoryFeedback(ctx context.Context, memoryID int64, impact MemoryFeedbackImpact) error {
	archiveBelow := -1.0
	if impact.ArchiveBelow != nil {
		archiveBelow = *impact.ArchiveBelow
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE study_memories
		SET support_count = support_count + ?,
			contradiction_count = contradiction_count + ?,
			confidence = LEAST(1.0, GREATEST(0.0, confidence + ?)),
			status = CASE
				WHEN LEAST(1.0, GREATEST(0.0, confidence + ?)) < ? THEN 'archived'
				ELSE status
			END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, impact.SupportDelta, impact.ContradictionDelta, impact.ConfidenceDelta, impact.ConfidenceDelta, archiveBelow, memoryID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFeedbackTargetNotFound
	}
	return nil
}
