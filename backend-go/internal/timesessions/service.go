package timesessions

import (
	"context"
	"errors"
)

var ErrInvalidTimeRange = errors.New("ended_at must be later than started_at")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdateFinishedSession(ctx context.Context, id int64, input UpdateTimeSessionInput) error {
	if !input.EndedAt.After(input.StartedAt) {
		return ErrInvalidTimeRange
	}

	input.DurationSeconds = int(input.EndedAt.Sub(input.StartedAt).Seconds())
	return s.repo.UpdateFinishedSession(ctx, id, input)
}
