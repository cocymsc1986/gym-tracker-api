package memory

import (
	"fmt"
	"strings"
	"sync"

	"gym-tracker-api/internal/models"
)

type ExerciseRepository struct {
	mu    sync.RWMutex
	items map[string]*models.Exercise
}

func NewExerciseRepository() *ExerciseRepository {
	return &ExerciseRepository{items: make(map[string]*models.Exercise)}
}

func exerciseKey(userID, exerciseID string) string {
	return userID + "\x00" + exerciseID
}

func (r *ExerciseRepository) GetByID(userID, exerciseID string) (*models.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.items[exerciseKey(userID, exerciseID)]
	if !ok {
		return nil, fmt.Errorf("exercise not found")
	}
	c := *e
	return &c, nil
}

func (r *ExerciseRepository) ListByUserID(userID string) ([]*models.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	prefix := userID + "\x00"
	var out []*models.Exercise
	for k, e := range r.items {
		if strings.HasPrefix(k, prefix) {
			c := *e
			out = append(out, &c)
		}
	}
	return out, nil
}

func (r *ExerciseRepository) ListByType(userID, exerciseType string) ([]*models.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	prefix := userID + "\x00"
	var out []*models.Exercise
	for k, e := range r.items {
		if strings.HasPrefix(k, prefix) && e.ExerciseType == exerciseType {
			c := *e
			out = append(out, &c)
		}
	}
	return out, nil
}

func (r *ExerciseRepository) ListByName(userID, exerciseName string) ([]*models.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	prefix := userID + "\x00"
	var out []*models.Exercise
	for k, e := range r.items {
		if strings.HasPrefix(k, prefix) && e.Name == exerciseName {
			c := *e
			out = append(out, &c)
		}
	}
	return out, nil
}

func (r *ExerciseRepository) Create(userID string, exercise *models.Exercise) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := exerciseKey(userID, exercise.ExerciseID)
	if _, exists := r.items[k]; exists {
		return fmt.Errorf("exercise already exists")
	}
	c := *exercise
	r.items[k] = &c
	return nil
}

func (r *ExerciseRepository) Update(userID string, exercise *models.Exercise) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := exerciseKey(userID, exercise.ExerciseID)
	if _, exists := r.items[k]; !exists {
		return fmt.Errorf("exercise not found")
	}
	c := *exercise
	r.items[k] = &c
	return nil
}

func (r *ExerciseRepository) Delete(userID, exerciseID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := exerciseKey(userID, exerciseID)
	if _, exists := r.items[k]; !exists {
		return fmt.Errorf("exercise not found")
	}
	delete(r.items, k)
	return nil
}
