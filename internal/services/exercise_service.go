package services

import (
	"gym-tracker-api/internal/models"
	"gym-tracker-api/internal/repository"
)

type ExerciseService interface {
	GetExercise(userID, exerciseID string) (*models.Exercise, error)
	GetExercises(userID string) ([]*models.Exercise, error)
	CreateExercise(exercise *models.Exercise) error
	UpdateExercise(userID, exerciseID string, exercise *models.Exercise) error
	DeleteExercise(userID, exerciseID string) error
	ListExercisesByType(userID, exerciseType string) ([]*models.Exercise, error)
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

func (s *exerciseService) CreateExercise(exercise *models.Exercise) error {
	if err := exercise.Validate(); err != nil {
		return err
	}
	return s.repo.Create(exercise)
}

func (s *exerciseService) UpdateExercise(userID string, exerciseID string, exercise *models.Exercise) error {
	if err := exercise.Validate(); err != nil {
		return err
	}
	exercise.ExerciseID = exerciseID
	return s.repo.Update(userID, exercise)
}

func (s *exerciseService) DeleteExercise(userID, exerciseID string) error {
	return s.repo.Delete(exerciseID, userID)
}

func (s *exerciseService) ListExercisesByType(userID, exerciseType string) ([]*models.Exercise, error) {
	exercises, err := s.repo.ListByType(userID, exerciseType)
	if err != nil {
		return nil, err
	}
	return exercises, nil
}
