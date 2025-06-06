package models

import (
	"errors"
	"time"
)

type Workout struct {
	UserID    string     `json:"userId" dynamodbav:"UserID" validate:"required"`
	WorkoutID string     `json:"workoutId" dynamodbav:"WorkoutID" validate:"required"`
	Name      string     `json:"name" validate:"required, min=1,max=100"`
	Exercises []Exercise `json:"exercises"`
	CreatedAt time.Time  `json:"createdAt"`
}

func(w *Workout) Validate() error {
	if w.UserID == "" {
		return errors.New("userID is required")
	}
	if w.WorkoutID == "" {
		return errors.New("workoutID is required")
	}
	if w.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
