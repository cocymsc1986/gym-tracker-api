package services

import (
	"errors"
	"math"
	"testing"

	"gym-tracker-api/internal/models"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

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
		ExerciseType: "weights",
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

	if err := svc.CreateExercise("user-1", sampleExercise(), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateExercise_MissingName(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleExercise()
	e.Name = ""

	if err := svc.CreateExercise("user-1", e, false); err == nil {
		t.Error("expected validation error for missing name, got nil")
	}
}

func TestCreateExercise_MissingExerciseType(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleExercise()
	e.ExerciseType = ""

	if err := svc.CreateExercise("user-1", e, false); err == nil {
		t.Error("expected validation error for missing exerciseType, got nil")
	}
}

func TestCreateExercise_MissingID(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleExercise()
	e.ExerciseID = ""

	if err := svc.CreateExercise("user-1", e, false); err == nil {
		t.Error("expected validation error for missing exerciseId, got nil")
	}
}

func TestCreateExercise_RepoError(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{err: errors.New("write failed")})

	if err := svc.CreateExercise("user-1", sampleExercise(), false); err == nil {
		t.Error("expected repo error, got nil")
	}
}

// UpdateExercise

func TestUpdateExercise_Success(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	if err := svc.UpdateExercise("user-1", "ex-1", sampleExercise(), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateExercise_ValidationError(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleExercise()
	e.Name = ""

	if err := svc.UpdateExercise("user-1", "ex-1", e, false); err == nil {
		t.Error("expected validation error, got nil")
	}
}

// calculateRPM

func sampleCardioExercise() *models.Exercise {
	return &models.Exercise{
		ExerciseID:   "ex-2",
		Name:         "Cycling",
		ExerciseType: "cardio",
		Distance:     10,
		DistanceUnit: "km",
		Time:         3600, // 60 minutes
	}
}

func TestCalculateRPM_KM(t *testing.T) {
	// 10 km = 10000 m; revolutions = 10000 / 6.2 ≈ 1612.9; minutes = 3600/60 = 60; rpm ≈ 26.88
	e := sampleCardioExercise()
	rpm := calculateRPM(e)
	expected := (10000.0 / 6.2) / 60.0
	if !approxEqual(rpm, expected) {
		t.Errorf("expected rpm %f, got %f", expected, rpm)
	}
}

func TestCalculateRPM_Miles(t *testing.T) {
	// 5 miles = 5 * 1609.344 = 8046.72 m; revolutions = 8046.72 / 6.2 ≈ 1297.86; minutes = 1800/60 = 30; rpm ≈ 43.26
	e := &models.Exercise{
		ExerciseID:   "ex-3",
		Name:         "Running",
		ExerciseType: "cardio",
		Distance:     5,
		DistanceUnit: "miles",
		Time:         1800,
	}
	rpm := calculateRPM(e)
	expected := (5 * 1609.344 / 6.2) / 30.0
	if !approxEqual(rpm, expected) {
		t.Errorf("expected rpm %f, got %f", expected, rpm)
	}
}

func TestCalculateRPM_NonCardio(t *testing.T) {
	e := sampleExercise() // weights exercise
	if rpm := calculateRPM(e); rpm != 0 {
		t.Errorf("expected 0 for non-cardio, got %f", rpm)
	}
}

func TestCalculateRPM_MissingTime(t *testing.T) {
	e := sampleCardioExercise()
	e.Time = 0
	if rpm := calculateRPM(e); rpm != 0 {
		t.Errorf("expected 0 when time is missing, got %f", rpm)
	}
}

func TestCalculateRPM_MissingDistance(t *testing.T) {
	e := sampleCardioExercise()
	e.Distance = 0
	if rpm := calculateRPM(e); rpm != 0 {
		t.Errorf("expected 0 when distance is missing, got %f", rpm)
	}
}

func TestCalculateRPM_UnsupportedUnit(t *testing.T) {
	e := sampleCardioExercise()
	e.DistanceUnit = "meters"
	if rpm := calculateRPM(e); rpm != 0 {
		t.Errorf("expected 0 for unsupported distance unit, got %f", rpm)
	}
}

func TestCreateExercise_StoreRPM(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleCardioExercise()
	if err := svc.CreateExercise("user-1", e, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.RPM == 0 {
		t.Error("expected RPM to be populated when storeRpm is true")
	}
}

func TestCreateExercise_StoreRPM_False(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleCardioExercise()
	if err := svc.CreateExercise("user-1", e, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.RPM != 0 {
		t.Error("expected RPM to be 0 when storeRpm is false")
	}
}

func TestUpdateExercise_StoreRPM(t *testing.T) {
	svc := NewExerciseService(&mockExerciseRepo{})

	e := sampleCardioExercise()
	if err := svc.UpdateExercise("user-1", "ex-2", e, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.RPM == 0 {
		t.Error("expected RPM to be populated when storeRpm is true")
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
