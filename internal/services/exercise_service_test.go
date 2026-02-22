package services

import (
	"errors"
	"testing"

	"gym-tracker-api/internal/models"
)

// mockExerciseRepo implements repository.ExerciseRepository for testing.
type mockExerciseRepo struct {
	exercise  *models.Exercise
	exercises []*models.Exercise
	err       error
}

func (m *mockExerciseRepo) GetByID(userID, exerciseID string) (*models.Exercise, error) {
	return m.exercise, m.err
}

func (m *mockExerciseRepo) ListByUserID(userID string) ([]*models.Exercise, error) {
	return m.exercises, m.err
}

func (m *mockExerciseRepo) ListByType(userID, exerciseType string) ([]*models.Exercise, error) {
	return m.exercises, m.err
}

func (m *mockExerciseRepo) ListByName(userID, exerciseName string) ([]*models.Exercise, error) {
	return m.exercises, m.err
}

func (m *mockExerciseRepo) Create(userID string, exercise *models.Exercise) error {
	return m.err
}

func (m *mockExerciseRepo) Update(userID string, exercise *models.Exercise) error {
	return m.err
}

func (m *mockExerciseRepo) Delete(userID, exerciseID string) error {
	return m.err
}

func sampleExercise() *models.Exercise {
	return &models.Exercise{
		ExerciseID:   "ex-1",
		Name:         "Bench Press",
		ExerciseType: "strength",
	}
}

// GetExercise

func TestGetExercise_Success(t *testing.T) {
	want := sampleExercise()
	svc := NewExerciseService(&mockExerciseRepo{exercise: want})

	got, err := svc.GetExercise("user-1", "ex-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ExerciseID != want.ExerciseID {
		t.Errorf("expected exerciseId %s, got %s", want.ExerciseID, got.ExerciseID)
	}
}

func TestGetExercise_RepoError(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{err: models.ErrExerciseNotFound})

	_, err := svc.GetExercise("user-1", "missing")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// GetExercises

func TestGetExercises_Success(t *testing.T) {
	exercises := []*models.Exercise{sampleExercise()}
	svc := NewExerciseService(&mockExerciseRepo{exercises: exercises})

	got, err := svc.GetExercises("user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 exercise, got %d", len(got))
	}
}

func TestGetExercises_RepoError(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{err: errors.New("db error")})

	_, err := svc.GetExercises("user-1")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// CreateExercise

func TestCreateExercise_Success(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	if err := svc.CreateExercise("user-1", sampleExercise()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateExercise_MissingName(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleExercise()
	e.Name = ""

	if err := svc.CreateExercise("user-1", e); err == nil {
		t.Error("expected validation error for missing name, got nil")
	}
}

func TestCreateExercise_MissingExerciseType(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleExercise()
	e.ExerciseType = ""

	if err := svc.CreateExercise("user-1", e); err == nil {
		t.Error("expected validation error for missing exerciseType, got nil")
	}
}

func TestCreateExercise_MissingID(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleExercise()
	e.ExerciseID = ""

	if err := svc.CreateExercise("user-1", e); err == nil {
		t.Error("expected validation error for missing exerciseId, got nil")
	}
}

func TestCreateExercise_RepoError(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{err: errors.New("write failed")})

	if err := svc.CreateExercise("user-1", sampleExercise()); err == nil {
		t.Error("expected repo error, got nil")
	}
}

// UpdateExercise

func TestUpdateExercise_Success(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	if err := svc.UpdateExercise("user-1", "ex-1", sampleExercise()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateExercise_ValidationError(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleExercise()
	e.Name = ""

	if err := svc.UpdateExercise("user-1", "ex-1", e); err == nil {
		t.Error("expected validation error, got nil")
	}
}

// DeleteExercise

func TestDeleteExercise_Success(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	if err := svc.DeleteExercise("user-1", "ex-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteExercise_RepoError(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{err: models.ErrExerciseNotFound})

	if err := svc.DeleteExercise("user-1", "missing"); err == nil {
		t.Error("expected error, got nil")
	}
}

// ListExercisesByName

func TestListExercisesByName_Success(t *testing.T) {
	exercises := []*models.Exercise{sampleExercise()}
	svc := NewExerciseService(&mockExerciseRepo{exercises: exercises})

	got, err := svc.ListExercisesByName("user-1", "Bench Press")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 result, got %d", len(got))
	}
}

func TestListExercisesByName_RepoError(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{err: errors.New("db error")})

	_, err := svc.ListExercisesByName("user-1", "Squat")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ListExercisesByType

func TestListExercisesByType_Success(t *testing.T) {
	exercises := []*models.Exercise{sampleExercise()}
	svc := NewExerciseService(&mockExerciseRepo{exercises: exercises})

	got, err := svc.ListExercisesByType("user-1", "strength")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 result, got %d", len(got))
	}
}

func TestListExercisesByType_RepoError(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{err: errors.New("db error")})

	_, err := svc.ListExercisesByType("user-1", "cardio")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
