package services

import (
	"gym-tracker-api/internal/models"
	"gym-tracker-api/internal/repository"
	"strings"
)

type ExerciseService interface {
	GetExercise(userID, exerciseID string) (*models.Exercise, error)
	GetExercises(userID string) ([]*models.Exercise, error)
	CreateExercise(userID string, exercise *models.Exercise, storeRpm bool) error
	UpdateExercise(userID, exerciseID string, exercise *models.Exercise, storeRpm bool) error
	DeleteExercise(userID, exerciseID string) error
	ListExercisesByType(userID, exerciseType string) ([]*models.Exercise, error)
	ListExercisesByName(userID, exerciseName string) ([]*models.Exercise, error)
}

type exerciseService struct {
	repo repository.ExerciseRepository
}

func NewExerciseService(repo repository.ExerciseRepository) ExerciseService {
	return &exerciseService{
		repo: repo,
	}
}

func (s *exerciseService) GetExercise(userID, exerciseID string) (*models.Exercise, error) {
	exercise, err := s.repo.GetByID(userID, exerciseID)
	if err != nil {
		return nil, err
	}
	return exercise, nil
}

func (s *exerciseService) GetExercises(userID string) ([]*models.Exercise, error) {
	exercises, err := s.repo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	return exercises, nil
}

func (s *exerciseService) CreateExercise(userID string, exercise *models.Exercise, storeRpm bool) error {
	if err := exercise.Validate(); err != nil {
		return err
	}
	if storeRpm {
		exercise.RPM = calculateRPM(exercise)
	}
	return s.repo.Create(userID, exercise)
}

func (s *exerciseService) UpdateExercise(userID string, exerciseID string, exercise *models.Exercise, storeRpm bool) error {
	if err := exercise.Validate(); err != nil {
		return err
	}
	exercise.ExerciseID = exerciseID
	if storeRpm {
		exercise.RPM = calculateRPM(exercise)
	}
	return s.repo.Update(userID, exercise)
}

// calculateRPM computes revolutions per minute for a cardio exercise.
// Requires cardio type, a positive time (seconds), and distance in miles or km.
// Uses the rule: 6.2 metres = 1 revolution.
func calculateRPM(exercise *models.Exercise) float64 {
	if exercise.ExerciseType != models.ExerciseTypeCardio {
		return 0
	}
	if exercise.Time <= 0 || exercise.Distance <= 0 {
		return 0
	}

	var distanceMeters float64
	switch strings.ToLower(exercise.DistanceUnit) {
	case "miles", "mile":
		distanceMeters = exercise.Distance * 1609.344
	case "km", "kilometers", "kilometre", "kilometres":
		distanceMeters = exercise.Distance * 1000
	default:
		return 0
	}

	revolutions := distanceMeters / 6.2
	minutes := float64(exercise.Time) / 60.0
	return revolutions / minutes
}

func (s *exerciseService) DeleteExercise(userID, exerciseID string) error {
	return s.repo.Delete(userID, exerciseID)
}

func (s *exerciseService) ListExercisesByType(userID, exerciseType string) ([]*models.Exercise, error) {
	exercises, err := s.repo.ListByType(userID, exerciseType)
	if err != nil {
		return nil, err
	}
	return exercises, nil
}

func (s *exerciseService) ListExercisesByName(userID, exerciseName string) ([]*models.Exercise, error) {
	exercises, err := s.repo.ListByName(userID, exerciseName)
	if err != nil {
		return nil, err
	}
	return exercises, nil
}
