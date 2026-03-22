package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	repoDb "gym-tracker-api/internal/repository/db"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	userID := flag.String("user-id", "", "Cognito UserID (sub) whose data should be deleted (required)")
	env := flag.String("env", "prod", "Environment suffix for DynamoDB table names (prod or test)")
	dryRun := flag.Bool("dry-run", false, "List what would be deleted without deleting anything")
	flag.Parse()

	if *userID == "" {
		log.Fatal("--user-id is required")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewEnvCredentials(),
	}))
	dynamo := dynamodb.New(sess)

	workoutsTable := fmt.Sprintf("Workouts-%s", *env)
	exercisesTable := fmt.Sprintf("Exercises-%s", *env)

	workoutRepo := repoDb.NewDynamoWorkoutRepository(dynamo, workoutsTable)
	exerciseRepo := repoDb.NewDynamoExerciseRepository(dynamo, exercisesTable)

	if *dryRun {
		fmt.Println("DRY RUN — no data will be deleted")
	}

	// --- Delete exercises ---
	exercises, err := exerciseRepo.ListByUserID(*userID)
	if err != nil {
		log.Fatalf("failed to list exercises: %v", err)
	}

	fmt.Printf("Found %d exercises\n", len(exercises))
	deletedExercises := 0
	for _, ex := range exercises {
		if *dryRun {
			fmt.Printf("  [exercise] %s (%s) %s\n", ex.Name, ex.ExerciseType, ex.ExerciseID)
			continue
		}
		if err := exerciseRepo.Delete(*userID, ex.ExerciseID); err != nil {
			log.Printf("WARNING: failed to delete exercise %s (%s): %v", ex.ExerciseID, ex.Name, err)
			continue
		}
		deletedExercises++
	}

	// --- Delete workouts ---
	workouts, err := workoutRepo.ListByUserID(*userID)
	if err != nil {
		log.Fatalf("failed to list workouts: %v", err)
	}

	fmt.Printf("Found %d workouts\n", len(workouts))
	deletedWorkouts := 0
	for _, w := range workouts {
		if *dryRun {
			fmt.Printf("  [workout] %s — %s %s\n", w.Date, w.Name, w.WorkoutID)
			continue
		}
		if err := workoutRepo.Delete(w.WorkoutID, *userID); err != nil {
			log.Printf("WARNING: failed to delete workout %s (%s %s): %v", w.WorkoutID, w.Date, w.Name, err)
			continue
		}
		deletedWorkouts++
	}

	if !*dryRun {
		fmt.Printf("\nDone. Deleted %d exercises and %d workouts.\n", deletedExercises, deletedWorkouts)
	}
}
