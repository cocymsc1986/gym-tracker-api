package services

import (
	"errors"
	"testing"
	"time"

	"gym-tracker-api/internal/models"
)

// mockWorkoutRepo implements repository.WorkoutRepository for testing.
type mockWorkoutRepo struct {
	workout  *models.Workout
	workouts []*models.Workout
	err      error
	updated  *models.Workout
}

func (m *mockWorkoutRepo) GetByID(userID, workoutID string) (*models.Workout, error) {
	return m.workout, m.err
}

func (m *mockWorkoutRepo) ListByUserID(userID string) ([]*models.Workout, error) {
	return m.workouts, m.err
}

func (m *mockWorkoutRepo) Create(workout *models.Workout) error {
	return m.err
}

func (m *mockWorkoutRepo) Update(workout *models.Workout) error {
	m.updated = workout
	return m.err
}

func (m *mockWorkoutRepo) Delete(workoutID, userID string) error {
	return m.err
}

func sampleWorkout() *models.Workout {
	return &models.Workout{
		UserID:    "user-1",
		WorkoutID: "workout-1",
		Name:      "Push Day",
		Date:      "2024-01-15",
		Exercises: []string{"ex-1"},
		CreatedAt: time.Now(),
	}
}

// GetWorkout

func TestGetWorkout_Success(t *testing.T) {
	want := sampleWorkout()
	svc := NewWorkoutService(&mockWorkoutRepo{workout: want})

	got, err := svc.GetWorkout("user-1", "workout-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.WorkoutID != want.WorkoutID {
		t.Errorf("expected workoutId %s, got %s", want.WorkoutID, got.WorkoutID)
	}
}

func TestGetWorkout_RepoError(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{err: errors.New("db error")})

	_, err := svc.GetWorkout("user-1", "workout-1")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// GetWorkouts

func TestGetWorkouts_Success(t *testing.T) {
	workouts := []*models.Workout{sampleWorkout()}
	svc := NewWorkoutService(&mockWorkoutRepo{workouts: workouts})

	got, err := svc.GetWorkouts("user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 workout, got %d", len(got))
	}
}

func TestGetWorkouts_RepoError(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{err: errors.New("db error")})

	_, err := svc.GetWorkouts("user-1")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// CreateWorkout

func TestCreateWorkout_Success(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{})

	w := sampleWorkout()
	if err := svc.CreateWorkout(w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateWorkout_MissingName(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{})

	w := sampleWorkout()
	w.Name = ""

	if err := svc.CreateWorkout(w); err == nil {
		t.Error("expected validation error for missing name, got nil")
	}
}

func TestCreateWorkout_MissingDate(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{})

	w := sampleWorkout()
	w.Date = ""

	if err := svc.CreateWorkout(w); err == nil {
		t.Error("expected validation error for missing date, got nil")
	}
}

func TestCreateWorkout_RepoError(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{err: errors.New("write failed")})

	if err := svc.CreateWorkout(sampleWorkout()); err == nil {
		t.Error("expected repo error, got nil")
	}
}

// UpdateWorkout

func TestUpdateWorkout_Success(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{})

	w := sampleWorkout()
	if err := svc.UpdateWorkout("user-1", "workout-1", w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateWorkout_ValidationError(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{})

	w := sampleWorkout()
	w.Name = ""

	if err := svc.UpdateWorkout("user-1", "workout-1", w); err == nil {
		t.Error("expected validation error, got nil")
	}
}

// DeleteWorkout

func TestDeleteWorkout_Success(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{})

	if err := svc.DeleteWorkout("user-1", "workout-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteWorkout_RepoError(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{err: models.ErrWorkoutNotFound})

	if err := svc.DeleteWorkout("user-1", "missing"); err == nil {
		t.Error("expected error, got nil")
	}
}

// AddExerciseToWorkout

func TestAddExerciseToWorkout_Success(t *testing.T) {
	repo := &mockWorkoutRepo{workout: sampleWorkout()}
	svc := NewWorkoutService(repo)

	if err := svc.AddExerciseToWorkout("user-1", "workout-1", "ex-new"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, id := range repo.updated.Exercises {
		if id == "ex-new" {
			found = true
		}
	}
	if !found {
		t.Error("exercise was not appended to workout")
	}
}

func TestAddExerciseToWorkout_WorkoutNotFound(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{err: models.ErrWorkoutNotFound})

	if err := svc.AddExerciseToWorkout("user-1", "missing", "ex-1"); err == nil {
		t.Error("expected error, got nil")
	}
}

// RemoveExerciseFromWorkout

func TestRemoveExerciseFromWorkout_Success(t *testing.T) {
	repo := &mockWorkoutRepo{workout: sampleWorkout()}
	svc := NewWorkoutService(repo)

	if err := svc.RemoveExerciseFromWorkout("user-1", "workout-1", "ex-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, id := range repo.updated.Exercises {
		if id == "ex-1" {
			t.Error("exercise was not removed from workout")
		}
	}
}

func TestRemoveExerciseFromWorkout_ExerciseNotInWorkout(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{workout: sampleWorkout()})

	err := svc.RemoveExerciseFromWorkout("user-1", "workout-1", "not-there")
	if err == nil {
		t.Error("expected error for non-existent exercise, got nil")
	}
	if !errors.Is(err, models.ErrExerciseNotFound) {
		t.Errorf("expected ErrExerciseNotFound, got %v", err)
	}
}

func TestRemoveExerciseFromWorkout_WorkoutNotFound(t *testing.T) {
	svc := NewWorkoutService(&mockWorkoutRepo{err: models.ErrWorkoutNotFound})

	if err := svc.RemoveExerciseFromWorkout("user-1", "missing", "ex-1"); err == nil {
		t.Error("expected error, got nil")
	}
}
