# Gym Tracker API Refactoring Plan

## Current Issues Analysis

### Critical Data Model Issues
- `GetWorkout` only retrieves by UserID but ignores WorkoutID (main.go:56-71)
- `CreateWorkout` doesn't store the exercises array (main.go:73-96)
- `UpdateWorkout` sets exercises to empty array (main.go:134)
- Exercise-workout relationship is broken

### API Design Problems
- Inconsistent HTTP methods: `PUT` for create, `POST` for update
- Missing CRUD operations (no delete endpoints)
- No pagination for exercises list
- No filtering or search capabilities

### Error Handling & Validation
- No input validation
- Generic error messages
- No structured error responses
- Missing status code consistency

### Security & Best Practices
- No authentication/authorization
- No request size limits
- No CORS handling
- Hardcoded port (8080)

### Code Quality
- No tests
- No logging framework
- No graceful shutdown
- No health check endpoint

### Performance Issues
- Uses `Scan` for exercises (expensive operation)
- No connection pooling
- No caching

### Infrastructure
- Lambda environment variables don't match Go code expectations
- No monitoring/alerting setup
- Missing API Gateway integration in main.go

## Suggested Modular Architecture

### Directory Structure
```
gym-tracker-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point, route setup
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── models/
│   │   ├── workout.go           # Workout struct & validation
│   │   └── exercise.go          # Exercise struct & validation
│   ├── handlers/
│   │   ├── workout_handler.go   # Workout HTTP handlers
│   │   └── exercise_handler.go  # Exercise HTTP handlers
│   ├── services/
│   │   ├── workout_service.go   # Business logic for workouts
│   │   └── exercise_service.go  # Business logic for exercises
│   ├── repository/
│   │   ├── interfaces.go        # Repository interfaces
│   │   ├── dynamodb/
│   │   │   ├── workout_repo.go  # DynamoDB workout operations
│   │   │   └── exercise_repo.go # DynamoDB exercise operations
│   │   └── connection.go        # DB connection setup
│   ├── middleware/
│   │   ├── auth.go             # Authentication middleware
│   │   ├── cors.go             # CORS middleware
│   │   └── logging.go          # Request logging
│   └── utils/
│       ├── response.go         # Standard API responses
│       └── validation.go       # Input validation helpers
├── pkg/
│   └── errors/
│       └── api_errors.go       # Custom error types
└── deployments/
    └── lambda/
        └── handler.go          # Lambda-specific entry point
```

### Key Benefits

#### 1. Separation of Concerns
- **Handlers**: HTTP request/response logic only
- **Services**: Business logic and validation
- **Repository**: Data access layer
- **Models**: Data structures and validation rules

#### 2. Testability
- Each layer can be unit tested independently
- Easy to mock dependencies with interfaces

#### 3. Maintainability
- Clear boundaries between components
- Easy to find and modify specific functionality

#### 4. Scalability
- New features fit into existing patterns
- Easy to swap implementations (e.g., different databases)

## Example Refactored Code

### handlers/workout_handler.go
```go
package handlers

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/cocymsc1986/gym-tracker-api/internal/services"
    "github.com/cocymsc1986/gym-tracker-api/internal/utils"
)

type WorkoutHandler struct {
    service services.WorkoutService
}

func NewWorkoutHandler(service services.WorkoutService) *WorkoutHandler {
    return &WorkoutHandler{service: service}
}

func (h *WorkoutHandler) GetWorkout(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    userID := vars["userId"]
    workoutID := vars["workoutId"]
    
    workout, err := h.service.GetWorkout(userID, workoutID)
    if err != nil {
        utils.WriteErrorResponse(w, err)
        return
    }
    utils.WriteJSONResponse(w, workout, http.StatusOK)
}

func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
    var workout models.Workout
    if err := utils.DecodeJSON(r.Body, &workout); err != nil {
        utils.WriteErrorResponse(w, err)
        return
    }
    
    if err := h.service.CreateWorkout(&workout); err != nil {
        utils.WriteErrorResponse(w, err)
        return
    }
    
    utils.WriteJSONResponse(w, workout, http.StatusCreated)
}
```

### services/workout_service.go
```go
package services

import (
    "github.com/cocymsc1986/gym-tracker-api/internal/models"
    "github.com/cocymsc1986/gym-tracker-api/internal/repository"
)

type WorkoutService interface {
    GetWorkout(userID, workoutID string) (*models.Workout, error)
    CreateWorkout(workout *models.Workout) error
    UpdateWorkout(workout *models.Workout) error
    DeleteWorkout(userID, workoutID string) error
    ListWorkouts(userID string) ([]*models.Workout, error)
}

type workoutService struct {
    repo repository.WorkoutRepository
}

func NewWorkoutService(repo repository.WorkoutRepository) WorkoutService {
    return &workoutService{repo: repo}
}

func (s *workoutService) GetWorkout(userID, workoutID string) (*models.Workout, error) {
    if userID == "" || workoutID == "" {
        return nil, errors.NewValidationError("userID and workoutID are required")
    }
    
    return s.repo.GetByID(userID, workoutID)
}

func (s *workoutService) CreateWorkout(workout *models.Workout) error {
    if err := workout.Validate(); err != nil {
        return err
    }
    
    return s.repo.Create(workout)
}
```

### repository/interfaces.go
```go
package repository

import "github.com/cocymsc1986/gym-tracker-api/internal/models"

type WorkoutRepository interface {
    GetByID(userID, workoutID string) (*models.Workout, error)
    Create(workout *models.Workout) error
    Update(workout *models.Workout) error
    Delete(userID, workoutID string) error
    ListByUserID(userID string) ([]*models.Workout, error)
}

type ExerciseRepository interface {
    GetByID(exerciseID string) (*models.Exercise, error)
    Create(exercise *models.Exercise) error
    Update(exercise *models.Exercise) error
    Delete(exerciseID string) error
    List() ([]*models.Exercise, error)
    ListByTag(tag string) ([]*models.Exercise, error)
}
```

### models/workout.go
```go
package models

import (
    "errors"
    "time"
)

type Workout struct {
    UserID    string     `json:"userId" validate:"required"`
    WorkoutID string     `json:"workoutId" validate:"required"`
    Name      string     `json:"name" validate:"required,min=1,max=100"`
    Exercises []Exercise `json:"exercises"`
    CreatedAt time.Time  `json:"createdAt"`
    UpdatedAt time.Time  `json:"updatedAt"`
}

func (w *Workout) Validate() error {
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
```

### utils/response.go
```go
package utils

import (
    "encoding/json"
    "net/http"
    "github.com/cocymsc1986/gym-tracker-api/pkg/errors"
)

type APIResponse struct {
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Message string      `json:"message,omitempty"`
}

func WriteJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    response := APIResponse{Data: data}
    json.NewEncoder(w).Encode(response)
}

func WriteErrorResponse(w http.ResponseWriter, err error) {
    w.Header().Set("Content-Type", "application/json")
    
    var statusCode int
    var message string
    
    switch e := err.(type) {
    case *errors.ValidationError:
        statusCode = http.StatusBadRequest
        message = e.Error()
    case *errors.NotFoundError:
        statusCode = http.StatusNotFound
        message = e.Error()
    default:
        statusCode = http.StatusInternalServerError
        message = "Internal server error"
    }
    
    w.WriteHeader(statusCode)
    response := APIResponse{Error: message}
    json.NewEncoder(w).Encode(response)
}
```

## Implementation Priority

1. **Phase 1**: Create basic directory structure and move models
2. **Phase 2**: Extract handlers and create service interfaces
3. **Phase 3**: Implement repository pattern with DynamoDB
4. **Phase 4**: Add middleware (auth, CORS, logging)
5. **Phase 5**: Add comprehensive error handling and validation
6. **Phase 6**: Add tests for each layer
7. **Phase 7**: Performance optimizations and caching

## Additional Improvements to Consider

- Add OpenAPI/Swagger documentation
- Implement health check endpoints
- Add metrics and monitoring
- Configure graceful shutdown
- Add rate limiting
- Implement proper logging with structured logs
- Add database migrations for schema changes
- Configure CI/CD pipeline
- Add integration tests
- Implement caching strategy