// Package memory provides in-memory implementations of the repository
// interfaces for local development and testing. They are not used by the
// production Lambda entrypoint.
package memory

import (
	"fmt"
	"sync"
	"time"

	"gym-tracker-api/internal/models"
)

type WorkoutRepository struct {
	mu    sync.RWMutex
	items map[string]*models.Workout
}

func NewWorkoutRepository() *WorkoutRepository {
	return &WorkoutRepository{items: make(map[string]*models.Workout)}
}

func workoutKey(userID, workoutID string) string {
	return userID + "\x00" + workoutID
}

func (r *WorkoutRepository) GetByID(userID, workoutID string) (*models.Workout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.items[workoutKey(userID, workoutID)]
	if !ok {
		return nil, fmt.Errorf("workout not found")
	}
	copy := *w
	return &copy, nil
}

func (r *WorkoutRepository) ListByUserID(userID string) ([]*models.Workout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*models.Workout
	for _, w := range r.items {
		if w.UserID == userID {
			c := *w
			out = append(out, &c)
		}
	}
	return out, nil
}

func (r *WorkoutRepository) Create(workout *models.Workout) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := workoutKey(workout.UserID, workout.WorkoutID)
	if _, exists := r.items[k]; exists {
		return fmt.Errorf("workout already exists")
	}
	if workout.CreatedAt.IsZero() {
		workout.CreatedAt = time.Now()
	}
	c := *workout
	r.items[k] = &c
	return nil
}

func (r *WorkoutRepository) Update(workout *models.Workout) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := workoutKey(workout.UserID, workout.WorkoutID)
	if _, exists := r.items[k]; !exists {
		return fmt.Errorf("workout not found")
	}
	c := *workout
	r.items[k] = &c
	return nil
}

func (r *WorkoutRepository) Delete(workoutID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := workoutKey(userID, workoutID)
	if _, exists := r.items[k]; !exists {
		return fmt.Errorf("workout not found")
	}
	delete(r.items, k)
	return nil
}
