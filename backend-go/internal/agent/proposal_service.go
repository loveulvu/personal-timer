package agent

import (
	"context"
	"encoding/json"
	"errors"
	"personal/internal/dailytasks"
	"personal/internal/timer"
	"strings"
)

type actionProposalStore interface {
	GetActionProposal(ctx context.Context, id int64) (*ActionProposal, error)
	ListActionProposals(ctx context.Context, filter ActionProposalFilter) ([]ActionProposal, error)
	UpdateActionProposal(ctx context.Context, id int64, input UpdateActionProposalInput) (*ActionProposal, error)
}

type dailyTaskCreator interface {
	CreateDailyTask(req dailytasks.CreateDailyTaskRequest) (int64, error)
	GetDailyTaskByID(id int64) (*dailytasks.DailyTask, error)
}

type taskFinisher interface {
	FinishTask(taskID int64, input timer.FinishTaskInput) error
}

type ProposalService struct {
	store actionProposalStore
	tasks dailyTaskCreator
	timer taskFinisher
}

func NewProposalService(store actionProposalStore, tasks dailyTaskCreator, timer taskFinisher) *ProposalService {
	return &ProposalService{store: store, tasks: tasks, timer: timer}
}

func (s *ProposalService) List(ctx context.Context, status string) ([]ActionProposal, error) {
	if status == "" {
		status = string(ActionProposalStatusPending)
	}
	return s.store.ListActionProposals(ctx, ActionProposalFilter{Status: status})
}

func (s *ProposalService) Get(ctx context.Context, id int64) (*ActionProposal, error) {
	return s.store.GetActionProposal(ctx, id)
}

func (s *ProposalService) Reject(ctx context.Context, id int64) (*ActionProposal, error) {
	proposal, err := s.store.GetActionProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	switch proposal.Status {
	case ActionProposalStatusRejected:
		return proposal, nil
	case ActionProposalStatusExecuted:
		return nil, ErrProposalConflict
	case ActionProposalStatusPending:
		return s.store.UpdateActionProposal(ctx, id, UpdateActionProposalInput{
			Status: ActionProposalStatusRejected,
			Decide: true,
		})
	default:
		return nil, ErrProposalConflict
	}
}

func (s *ProposalService) Accept(ctx context.Context, id int64) (*ActionProposal, error) {
	proposal, err := s.store.GetActionProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	if proposal.Status == ActionProposalStatusExecuted {
		return proposal, nil
	}
	if proposal.Status != ActionProposalStatusPending {
		return nil, ErrProposalConflict
	}

	result, targetType, targetID, err := s.execute(ctx, proposal)
	if err != nil {
		updated, updateErr := s.store.UpdateActionProposal(ctx, id, UpdateActionProposalInput{
			Status:       ActionProposalStatusFailed,
			ErrorMessage: err.Error(),
			Decide:       true,
		})
		if updateErr != nil {
			return nil, updateErr
		}
		return updated, err
	}
	return s.store.UpdateActionProposal(ctx, id, UpdateActionProposalInput{
		Status:           ActionProposalStatusExecuted,
		Result:           result,
		Decide:           true,
		Execute:          true,
		TargetEntityType: targetType,
		TargetEntityID:   targetID,
	})
}

func (s *ProposalService) execute(ctx context.Context, proposal *ActionProposal) (json.RawMessage, string, *int64, error) {
	switch proposal.ActionType {
	case "create_daily_task":
		var payload struct {
			Date             string `json:"date"`
			ProjectID        *int64 `json:"project_id"`
			Title            string `json:"title"`
			EstimatedMinutes int    `json:"estimated_minutes"`
		}
		if err := json.Unmarshal(proposal.Payload, &payload); err != nil {
			return nil, "", nil, ErrInvalidToolInput
		}
		if validateDate(payload.Date) != nil || payload.ProjectID == nil || *payload.ProjectID <= 0 ||
			strings.TrimSpace(payload.Title) == "" || payload.EstimatedMinutes <= 0 {
			return nil, "", nil, ErrInvalidToolInput
		}
		taskID, err := s.tasks.CreateDailyTask(dailytasks.CreateDailyTaskRequest{
			ProjectID:        payload.ProjectID,
			TaskDate:         payload.Date,
			Title:            payload.Title,
			EstimatedMinutes: payload.EstimatedMinutes,
		})
		if err != nil {
			return nil, "", nil, err
		}
		task, err := s.tasks.GetDailyTaskByID(taskID)
		if err != nil {
			return nil, "", nil, err
		}
		return mustJSON(map[string]any{"task_id": taskID, "task": task}), "daily_task", &taskID, nil
	case "finish_task":
		var payload struct {
			TaskID            int64  `json:"task_id"`
			FinishNote        string `json:"finish_note"`
			FinishDescription string `json:"finish_description"`
		}
		if err := json.Unmarshal(proposal.Payload, &payload); err != nil {
			return nil, "", nil, ErrInvalidToolInput
		}
		if payload.TaskID <= 0 {
			return nil, "", nil, ErrInvalidToolInput
		}
		err := s.timer.FinishTask(payload.TaskID, timer.FinishTaskInput{
			FinishNote:        payload.FinishNote,
			FinishDescription: payload.FinishDescription,
		})
		if err != nil {
			return nil, "", nil, err
		}
		return mustJSON(map[string]any{"task_id": payload.TaskID, "status": "finished"}), "daily_task", &payload.TaskID, nil
	default:
		return nil, "", nil, errors.New("unsupported action proposal")
	}
}
