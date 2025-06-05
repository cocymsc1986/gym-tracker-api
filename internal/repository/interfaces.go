package repository

import "gym-tracker-api/internal/models"

type WorkoutRepository interface {
	GetByID(userID, workoutID string) (*models.Workout, error)
	ListByUserID(userID string) ([]*models.Workout, error)
	Create(workout *models.Workout) error
	Update(workout *models.Workout) error
	Delete(workoutID string, userID string) error
}

type ExerciseRepository interface {
	GetByID(userID, exerciseID string) (*models.Exercise, error)
	ListByUserID(userID string) ([]*models.Exercise, error)
	ListByTag(userID, tag string) ([]*models.Exercise, error)
	Create(exercise *models.Exercise) error
	Update(userID string, exercise *models.Exercise) error
	Delete(exerciseID string, userID string) error
}
