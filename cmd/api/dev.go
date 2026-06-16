package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gym-tracker-api/internal/handlers"
	"gym-tracker-api/internal/middleware"
	"gym-tracker-api/internal/models"
	"gym-tracker-api/internal/repository/memory"
	"gym-tracker-api/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// DevUserID is the well-known user ID that seed data is created under.
// QA agents should use this in path parameters and as the user_id returned
// from /auth/signin.
const DevUserID = "dev-user"

func runDev() {
	log.Println("DEV_MODE enabled — in-memory repositories, stubbed auth, no AWS dependencies")

	workoutRepo := memory.NewWorkoutRepository()
	exerciseRepo := memory.NewExerciseRepository()
	seedDevData(workoutRepo, exerciseRepo)

	workoutService := services.NewWorkoutService(workoutRepo)
	exerciseService := services.NewExerciseService(exerciseRepo)

	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)

	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	if originsEnv == "" {
		originsEnv = "*"
	}
	corsMiddleware := middleware.NewCORSMiddleware(strings.Split(originsEnv, ","))

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Printf("Request: %s %s", req.Method, req.URL.Path)
			next.ServeHTTP(w, req)
		})
	})

	// Stubbed auth — accept anything, return shaped responses.
	r.HandleFunc("/auth/signup", devSignUp).Methods("POST")
	r.HandleFunc("/auth/confirm", devMessage("Email confirmed successfully")).Methods("POST")
	r.HandleFunc("/auth/signin", devSignIn).Methods("POST")
	r.HandleFunc("/auth/refresh", devRefresh).Methods("POST")
	r.HandleFunc("/auth/reset", devMessage("Password reset code sent to your email.")).Methods("POST")
	r.HandleFunc("/auth/reset/confirm", devMessage("Password reset successfully.")).Methods("POST")

	// Pass-through auth — the Authorization header is not checked.
	r.HandleFunc("/workouts/{userId}", workoutHandler.ListWorkouts).Methods("GET")
	r.HandleFunc("/workouts/{userId}/{workoutId}", workoutHandler.GetWorkout).Methods("GET")
	r.HandleFunc("/workouts/{userId}", workoutHandler.CreateWorkout).Methods("POST")
	r.HandleFunc("/workouts/{userId}/{workoutId}", workoutHandler.UpdateWorkout).Methods("PUT")
	r.HandleFunc("/workouts/{userId}/{workoutId}", workoutHandler.DeleteWorkout).Methods("DELETE")
	r.HandleFunc("/workouts/{userId}/{workoutId}/exercises", workoutHandler.ListExercisesInWorkout).Methods("GET")
	r.HandleFunc("/workouts/{userId}/{workoutId}/exercises/{exerciseId}", workoutHandler.AddExerciseToWorkout).Methods("POST")
	r.HandleFunc("/workouts/{userId}/{workoutId}/exercises/{exerciseId}", workoutHandler.RemoveExerciseFromWorkout).Methods("DELETE")
	r.HandleFunc("/exercises/{userId}/{exerciseId}", exerciseHandler.GetExercise).Methods("GET")
	r.HandleFunc("/exercises/{userId}/name/{exerciseName}", exerciseHandler.ListExercisesByName).Methods("GET")
	r.HandleFunc("/exercises/{userId}", exerciseHandler.GetExercises).Methods("GET")
	r.HandleFunc("/exercises/{userId}", exerciseHandler.CreateExercise).Methods("POST")
	r.HandleFunc("/exercises/{userId}/{exerciseId}", exerciseHandler.UpdateExercise).Methods("PUT")
	r.HandleFunc("/exercises/{userId}/{exerciseId}", exerciseHandler.DeleteExercise).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Dev server listening on :%s (seeded user: %s)", port, DevUserID)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware.Handler(r)))
}

func devMessage(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
	}
}

func devSignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully. Please check your email for verification.",
	})
}

func devSignIn(w http.ResponseWriter, r *http.Request) {
	// Embed AuthResponse so the JSON shape exactly matches prod, with an
	// extra user_id field so callers don't need to decode the token.
	type devAuthResponse struct {
		handlers.AuthResponse
		UserID string `json:"user_id"`
	}
	resp := devAuthResponse{
		AuthResponse: handlers.AuthResponse{
			AccessToken:  "dev-access-token",
			RefreshToken: "dev-refresh-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
		},
		UserID: DevUserID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func devRefresh(w http.ResponseWriter, r *http.Request) {
	resp := handlers.AuthResponse{
		AccessToken: "dev-access-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func seedDevData(wr *memory.WorkoutRepository, er *memory.ExerciseRepository) {
	bench := &models.Exercise{
		ExerciseID:   uuid.New().String(),
		Name:         "Bench Press",
		ExerciseType: models.ExerciseTypeWeights,
		Sets: []models.WeightItem{
			{Weight: 60, Unit: "kg", Reps: 8},
			{Weight: 60, Unit: "kg", Reps: 8},
			{Weight: 60, Unit: "kg", Reps: 6},
		},
	}
	treadmill := &models.Exercise{
		ExerciseID:   uuid.New().String(),
		Name:         "Treadmill",
		ExerciseType: models.ExerciseTypeCardio,
		Time:         1200,
		Distance:     3,
		DistanceUnit: "km",
	}
	plank := &models.Exercise{
		ExerciseID:   uuid.New().String(),
		Name:         "Plank",
		ExerciseType: models.ExerciseTypeBodyWeight,
		Sets: []models.WeightItem{
			{Duration: 60},
			{Duration: 60},
		},
	}
	for _, ex := range []*models.Exercise{bench, treadmill, plank} {
		_ = er.Create(DevUserID, ex)
	}

	today := time.Now().Format("2006-01-02")
	pushDay := &models.Workout{
		UserID:    DevUserID,
		WorkoutID: uuid.New().String(),
		Name:      "Push Day",
		Date:      today,
		Exercises: []string{bench.ExerciseID},
		CreatedAt: time.Now(),
	}
	cardioDay := &models.Workout{
		UserID:    DevUserID,
		WorkoutID: uuid.New().String(),
		Name:      "Cardio Day",
		Date:      today,
		Exercises: []string{treadmill.ExerciseID, plank.ExerciseID},
		CreatedAt: time.Now(),
	}
	for _, w := range []*models.Workout{pushDay, cardioDay} {
		_ = wr.Create(w)
	}
}
