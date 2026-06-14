package timer

import (
	"errors"
	"testing"
)

type fakeRepository struct {
	finishInput  FinishTaskInput
	updateInput  UpdateCompletedTaskInput
	updateSecond *int
	updateActual bool
	finishErr    error
	updateErr    error
	deleteErr    error
	deleted      bool
}

func (f *fakeRepository) StartTask(int64) error  { return nil }
func (f *fakeRepository) PauseTask(int64) error  { return nil }
func (f *fakeRepository) ResumeTask(int64) error { return nil }
func (f *fakeRepository) FinishTask(_ int64, note, description string) error {
	f.finishInput = FinishTaskInput{FinishNote: note, FinishDescription: description}
	return f.finishErr
}
func (f *fakeRepository) UpdateCompletedTask(_ int64, note, description string, seconds *int, updateActual bool) error {
	f.updateInput = UpdateCompletedTaskInput{FinishNote: note, FinishDescription: description}
	f.updateSecond = seconds
	f.updateActual = updateActual
	return f.updateErr
}
func (f *fakeRepository) DeleteCompletedTask(int64) error {
	f.deleted = true
	return f.deleteErr
}

func TestFinishTaskRequiresNote(t *testing.T) {
	service := NewService(&fakeRepository{})
	err := service.FinishTask(1, FinishTaskInput{FinishDescription: "done"})
	if !errors.Is(err, ErrFinishNoteRequired) {
		t.Fatalf("expected ErrFinishNoteRequired, got %v", err)
	}
}

func TestFinishTaskRequiresDescription(t *testing.T) {
	service := NewService(&fakeRepository{})
	err := service.FinishTask(1, FinishTaskInput{FinishNote: "done"})
	if !errors.Is(err, ErrFinishDescriptionRequired) {
		t.Fatalf("expected ErrFinishDescriptionRequired, got %v", err)
	}
}

func TestUpdateCompletedTaskConvertsMinutesToSeconds(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	minutes := 12
	err := service.UpdateCompletedTask(1, UpdateCompletedTaskInput{
		FinishNote:            " note ",
		FinishDescription:     " description ",
		ActualMinutesOverride: &minutes,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.updateSecond == nil || *repo.updateSecond != 720 {
		t.Fatalf("expected 720 seconds, got %v", repo.updateSecond)
	}
	if !repo.updateActual {
		t.Fatal("expected actual duration update")
	}
	if repo.updateInput.FinishNote != "note" || repo.updateInput.FinishDescription != "description" {
		t.Fatalf("expected trimmed completion text, got %#v", repo.updateInput)
	}
}

func TestUpdateCompletedTaskCanClearOverride(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	err := service.UpdateCompletedTask(1, UpdateCompletedTaskInput{
		FinishNote:          "note",
		FinishDescription:   "description",
		ClearActualOverride: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.updateSecond != nil {
		t.Fatalf("expected nil override, got %v", *repo.updateSecond)
	}
	if !repo.updateActual {
		t.Fatal("expected clear override update")
	}
}

func TestUpdateCompletedTaskWithoutOverrideKeepsExistingValue(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	err := service.UpdateCompletedTask(1, UpdateCompletedTaskInput{
		FinishNote:        "note",
		FinishDescription: "description",
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.updateActual {
		t.Fatal("expected existing override to remain unchanged")
	}
}

func TestUpdateCompletedTaskRejectsConflictingOverrideInputs(t *testing.T) {
	service := NewService(&fakeRepository{})
	minutes := 5
	err := service.UpdateCompletedTask(1, UpdateCompletedTaskInput{
		FinishNote:            "note",
		FinishDescription:     "description",
		ActualMinutesOverride: &minutes,
		ClearActualOverride:   true,
	})
	if !errors.Is(err, ErrActualMinutesConflict) {
		t.Fatalf("expected ErrActualMinutesConflict, got %v", err)
	}
}

func TestUpdateCompletedTaskRejectsNegativeMinutes(t *testing.T) {
	service := NewService(&fakeRepository{})
	minutes := -1
	err := service.UpdateCompletedTask(1, UpdateCompletedTaskInput{
		FinishNote:            "note",
		FinishDescription:     "description",
		ActualMinutesOverride: &minutes,
	})
	if !errors.Is(err, ErrActualMinutesInvalid) {
		t.Fatalf("expected ErrActualMinutesInvalid, got %v", err)
	}
}

func TestUpdateCompletedTaskReturnsRepositoryStatusError(t *testing.T) {
	repo := &fakeRepository{updateErr: ErrTaskMustBeCompleted}
	service := NewService(repo)
	err := service.UpdateCompletedTask(1, UpdateCompletedTaskInput{
		FinishNote:        "note",
		FinishDescription: "description",
	})
	if !errors.Is(err, ErrTaskMustBeCompleted) {
		t.Fatalf("expected ErrTaskMustBeCompleted, got %v", err)
	}
}

func TestDeleteCompletedTaskReturnsRepositoryStatusError(t *testing.T) {
	repo := &fakeRepository{deleteErr: ErrTaskMustBeCompleted}
	service := NewService(repo)
	if !errors.Is(service.DeleteCompletedTask(1), ErrTaskMustBeCompleted) {
		t.Fatal("expected completed status error")
	}
}

func TestDeleteCompletedTaskSucceeds(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	if err := service.DeleteCompletedTask(1); err != nil {
		t.Fatal(err)
	}
	if !repo.deleted {
		t.Fatal("expected completed task deletion")
	}
}
