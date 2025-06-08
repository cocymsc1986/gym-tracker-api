package models

import "errors"

type WeightItem struct {
	Weight float64 `json:"weight"`
	Unit   string  `json:"unit"`
	Reps   int     `json:"reps"`
}

type Exercise struct {
	ExerciseID 		string  `json:"exerciseId" dynamodbav:"ExerciseID" validate:"required"`
	Name       		string  `json:"name"`
	ExerciseType  string  `json:"exerciseType" dynamodbav:"ExerciseType" validate:"required"`
	Time       		string  `json:"time,omitempty"`
	Distance	 		float64 `json:"distance,omitempty"`
	DistanceUnit 	string  `json:"distanceUnit,omitempty"`
	Level      		string  `json:"level,omitempty"`
	Sets       		[]WeightItem  `json:"sets,omitempty"`
}

func (e *Exercise) Validate() error {
	if e.ExerciseID == "" {
			return errors.New("ExerciseID is required")
	}
	if e.Name == "" {
		return errors.New("name is required")
	}
	if e.ExerciseType == "" {
			return errors.New("ExerciseType is required")
	}
	return nil
}
