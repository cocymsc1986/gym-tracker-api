package services

import (
	"gym-tracker-api/internal/models"
	"gym-tracker-api/internal/repository"
)

type WorkoutService interface {
	GetWorkout(userID, workoutID string) (*models.Workout, error)
	GetWorkouts(userID string) ([]*models.Workout, error)
	CreateWorkout(workout *models.Workout) error
	UpdateWorkout(userID, workoutID string, workout *models.Workout) error
	DeleteWorkout(userID, workoutID string) error
	AddExerciseToWorkout(userID, workoutID string, exercise *models.Exercise) error
	RemoveExerciseFromWorkout(userID, workoutID, exerciseID string) error
}

type workoutService struct {
	repo repository.WorkoutRepository
}

func NewWorkoutService(repo repository.WorkoutRepository) WorkoutService {
	return &workoutService{
		repo: repo,
	}
}

func (s *workoutService) GetWorkout(userID, workoutID string) (*models.Workout, error) {
	workout, err := s.repo.GetByID(userID, workoutID)
	if err != nil {
		return nil, err
	}
	return workout, nil
}

func (s *workoutService) GetWorkouts(userID string) ([]*models.Workout, error) {
	workouts, err := s.repo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	return workouts, nil
}

func (s *workoutService) CreateWorkout(workout *models.Workout) error {
	if err := workout.Validate(); err != nil {
		return err
	}
	return s.repo.Create(workout)
}

func (s *workoutService) UpdateWorkout(userID, workoutID string, workout *models.Workout) error {
	if err := workout.Validate(); err != nil {
		return err
	}
	return s.repo.Update(workout)
}

func (s *workoutService) DeleteWorkout(userID, workoutID string) error {
	return s.repo.Delete(workoutID, userID)
}

func (s *workoutService) AddExerciseToWorkout(userID, workoutID string, exercise *models.Exercise) error {
	if err := exercise.Validate(); err != nil {
		return err
	}
	workout, err := s.repo.GetByID(userID, workoutID)
	if err != nil {
		return err
	}

	if workout == nil {
		return models.ErrWorkoutNotFound
	}

	workout.Exercises = append(workout.Exercises, *exercise)
	return s.repo.Update(workout)
}

func (s *workoutService) RemoveExerciseFromWorkout(userID, workoutID, exerciseID string) error {
	workout, err := s.repo.GetByID(userID, workoutID)
	if err != nil {
		return err
	}

	if workout == nil {
		return models.ErrWorkoutNotFound
	}

	for i, ex := range workout.Exercises {
		if ex.ExerciseID == exerciseID {
			workout.Exercises = append(workout.Exercises[:i], workout.Exercises[i+1:]...)
			return s.repo.Update(workout)
		}
	}

	return models.ErrExerciseNotFound
}
