package models

import "errors"

// Error definitions for the application
var (
	ErrWorkoutNotFound       = errors.New("workout not found")
	ErrExerciseNotFound      = errors.New("exercise not found")
	ErrWorkoutAlreadyExists  = errors.New("workout already exists")
	ErrExerciseAlreadyExists = errors.New("exercise already exists")
	ErrInvalidWorkout        = errors.New("invalid workout data")
	ErrInvalidExercise       = errors.New("invalid exercise data")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrInternalServerError   = errors.New("internal server error")
)
// ErrUserNotFound is returned when a user is not found in the system
var ErrUserNotFound = errors.New("user not found")
// ErrInvalidCredentials is returned when the provided credentials are invalid
var ErrInvalidCredentials = errors.New("invalid credentials")
// ErrUserAlreadyExists is returned when trying to create a user that already exists
var ErrUserAlreadyExists = errors.New("user already exists")
// ErrPasswordTooShort is returned when the provided password is too short
var ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
// ErrEmailAlreadyExists is returned when trying to register with an email that already exists
var ErrEmailAlreadyExists = errors.New("email already exists")
// ErrInvalidEmailFormat is returned when the provided email format is invalid
var ErrInvalidEmailFormat = errors.New("invalid email format")
// ErrTokenExpired is returned when the provided token has expired
var ErrTokenExpired = errors.New("token has expired")
// ErrTokenInvalid is returned when the provided token is invalid
var ErrTokenInvalid = errors.New("invalid token")
