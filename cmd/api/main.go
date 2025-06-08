package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gym-tracker-api/internal/handlers"
	"gym-tracker-api/internal/middleware"
	"gym-tracker-api/internal/repository/db"
	"gym-tracker-api/internal/services"

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
	// Load .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
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
	
	r := mux.NewRouter()
	
	// Auth routes (no authentication required)
	r.HandleFunc("/auth/signup", authHandler.SignUp).Methods("POST")
	r.HandleFunc("/auth/confirm", authHandler.ConfirmSignUp).Methods("POST")
	r.HandleFunc("/auth/signin", authHandler.SignIn).Methods("POST")
	r.HandleFunc("/auth/refresh", authHandler.RefreshToken).Methods("POST")
	
	// Protected routes (authentication required)
	r.HandleFunc("/workouts/{userId}", authMiddleware.Authenticate(workoutHandler.ListWorkouts)).Methods("GET")
	r.HandleFunc("/workouts/{userId}/{workoutId}", authMiddleware.Authenticate(workoutHandler.GetWorkout)).Methods("GET")
	r.HandleFunc("/workouts", authMiddleware.Authenticate(workoutHandler.CreateWorkout)).Methods("POST")
	r.HandleFunc("/workouts/{userId}/{workoutId}", authMiddleware.Authenticate(workoutHandler.UpdateWorkout)).Methods("PUT")
	r.HandleFunc("/workouts/{userId}/{workoutId}", authMiddleware.Authenticate(workoutHandler.DeleteWorkout)).Methods("DELETE")
	r.HandleFunc("/workouts/{userId}/{workoutId}/exercises", authMiddleware.Authenticate(workoutHandler.ListExercisesInWorkout)).Methods("GET")
	r.HandleFunc("/workouts/{userId}/{workoutId}/exercises/{exerciseId}", authMiddleware.Authenticate(workoutHandler.AddExerciseToWorkout)).Methods("POST")
	r.HandleFunc("/workouts/{userId}/{workoutId}/exercises/{exerciseId}", authMiddleware.Authenticate(workoutHandler.RemoveExerciseFromWorkout)).Methods("DELETE")
	r.HandleFunc("/exercises", authMiddleware.Authenticate(exerciseHandler.GetExercises)).Methods("GET")
	r.HandleFunc("/exercises/{exerciseId}", authMiddleware.Authenticate(exerciseHandler.GetExercise)).Methods("GET")
	r.HandleFunc("/exercises", authMiddleware.Authenticate(exerciseHandler.CreateExercise)).Methods("POST")
	r.HandleFunc("/exercises/{exerciseId}", authMiddleware.Authenticate(exerciseHandler.UpdateExercise)).Methods("PUT")
	r.HandleFunc("/exercises/{exerciseId}", authMiddleware.Authenticate(exerciseHandler.DeleteExercise)).Methods("DELETE")
	
	log.Printf("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
