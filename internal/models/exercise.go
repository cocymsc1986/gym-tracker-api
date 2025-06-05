package models

import "errors"

type Exercise struct {
	ExerciseID string  `json:"exerciseId"`
	Name       string  `json:"name"`
	Tag        string  `json:"tag"`
	Time       string  `json:"time,omitempty"`
	Level      string  `json:"level,omitempty"`
	Sets       int     `json:"sets,omitempty"`
	Reps       int     `json:"reps,omitempty"`
	Weight     float64 `json:"weight,omitempty"`
}

func (e *Exercise) Validate() error {
	if e.ExerciseID == "" {
			return errors.New("ExerciseID is required")
	}
	if e.Name == "" {
		return errors.New("name is required")
	}
	if e.Tag == "" {
			return errors.New("tag is required")
	}
	return nil
}
