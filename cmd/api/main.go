package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"gym-tracker-api/internal/handlers"
	"gym-tracker-api/internal/middleware"
	"gym-tracker-api/internal/repository/db"
	"gym-tracker-api/internal/services"

	"github.com/akrylysov/algnhsa"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var (
	dynamoClient *dynamodb.DynamoDB
	cognitoClient *cognitoidentityprovider.CognitoIdentityProvider
)

func init() {
	// Load .env file (non-fatal — Lambda uses env vars from Terraform)
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	fmt.Println("Loaded AWS Region:", os.Getenv("AWS_REGION"))

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewEnvCredentials(),
	}))
	dynamoClient = dynamodb.New(sess)
	cognitoClient = cognitoidentityprovider.New(sess)
}

func setupHandlers() (*handlers.WorkoutHandler, *handlers.ExerciseHandler, *handlers.AuthHandler) {
	// Repository layer
	workoutRepo := db.NewDynamoWorkoutRepository(dynamoClient, os.Getenv("DYNAMO_TABLE_WORKOUTS"))
	exerciseRepo := db.NewDynamoExerciseRepository(dynamoClient, os.Getenv("DYNAMO_TABLE_EXERCISES"))
	
	// Service layer
	workoutService := services.NewWorkoutService(workoutRepo)
	exerciseService := services.NewExerciseService(exerciseRepo)
	
	// Handler layer
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	authHandler := handlers.NewAuthHandler(cognitoClient)
	
	return workoutHandler, exerciseHandler, authHandler
} 

func main() {
	// Initialize handlers with proper dependency injection
	workoutHandler, exerciseHandler, authHandler := setupHandlers()
	
	// Setup middleware
	authMiddleware := middleware.NewAuthMiddleware(cognitoClient)
	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	if originsEnv == "" {
		originsEnv = "http://localhost:5173,capacitor://localhost"
	}
	allowedOrigins := strings.Split(originsEnv, ",")
	corsMiddleware := middleware.NewCORSMiddleware(allowedOrigins)
	
	r := mux.NewRouter()
	
	// Add basic logging middleware first to verify requests are coming in
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})
	
	// Auth routes (no authentication required)
	r.HandleFunc("/auth/signup", authHandler.SignUp).Methods("POST")
	r.HandleFunc("/auth/confirm", authHandler.ConfirmSignUp).Methods("POST")
	r.HandleFunc("/auth/signin", authHandler.SignIn).Methods("POST")
	r.HandleFunc("/auth/refresh", authHandler.RefreshToken).Methods("POST")
	
	// Protected routes (authentication required)
	r.HandleFunc("/workouts/{userId}", authMiddleware.Authenticate(workoutHandler.ListWorkouts)).Methods("GET")
	r.HandleFunc("/workouts/{userId}/{workoutId}", authMiddleware.Authenticate(workoutHandler.GetWorkout)).Methods("GET")
	r.HandleFunc("/workouts/{userId}", authMiddleware.Authenticate(workoutHandler.CreateWorkout)).Methods("POST")
	r.HandleFunc("/workouts/{userId}/{workoutId}", authMiddleware.Authenticate(workoutHandler.UpdateWorkout)).Methods("PUT")
	r.HandleFunc("/workouts/{userId}/{workoutId}", authMiddleware.Authenticate(workoutHandler.DeleteWorkout)).Methods("DELETE")
	r.HandleFunc("/workouts/{userId}/{workoutId}/exercises", authMiddleware.Authenticate(workoutHandler.ListExercisesInWorkout)).Methods("GET")
	r.HandleFunc("/workouts/{userId}/{workoutId}/exercises/{exerciseId}", authMiddleware.Authenticate(workoutHandler.AddExerciseToWorkout)).Methods("POST")
	r.HandleFunc("/workouts/{userId}/{workoutId}/exercises/{exerciseId}", authMiddleware.Authenticate(workoutHandler.RemoveExerciseFromWorkout)).Methods("DELETE")
	r.HandleFunc("/exercises/{userId}/{exerciseId}", authMiddleware.Authenticate(exerciseHandler.GetExercise)).Methods("GET")
	r.HandleFunc("/exercises/{userId}/name/{exerciseName}", authMiddleware.Authenticate(exerciseHandler.ListExercisesByName)).Methods("GET")
	r.HandleFunc("/exercises/{userId}", authMiddleware.Authenticate(exerciseHandler.GetExercises)).Methods("GET")
	r.HandleFunc("/exercises/{userId}", authMiddleware.Authenticate(exerciseHandler.CreateExercise)).Methods("POST")
	r.HandleFunc("/exercises/{exerciseId}", authMiddleware.Authenticate(exerciseHandler.UpdateExercise)).Methods("PUT")
	r.HandleFunc("/exercises/{exerciseId}", authMiddleware.Authenticate(exerciseHandler.DeleteExercise)).Methods("DELETE")
	
	// Wrap router with CORS middleware so it runs before routing —
	// gorilla/mux r.Use() only runs when a route matches, which
	// excludes OPTIONS preflight requests that have no registered route.
	algnhsa.ListenAndServe(corsMiddleware.Handler(r), nil)
}
