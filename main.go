package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var db *dynamodb.DynamoDB

// Models
type Workout struct {
	UserID    string     `json:"userId"`
	WorkoutID string     `json:"workoutId"`
	Name      string     `json:"name"`
	Exercises []Exercise `json:"exercises"`
}

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

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	fmt.Println("Loaded AWS Region:", os.Getenv("AWS_REGION"))

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewEnvCredentials(),
	}))
	db = dynamodb.New(sess)
}

// Handlers
func GetWorkout(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userId"]
	result, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("TABLE_WORKOUTS")),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {
				S: aws.String(userID),
			},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error fetching workout")
		return
	}
	json.NewEncoder(w).Encode(result.Item)
}

func CreateWorkout(w http.ResponseWriter, r *http.Request) {
	var workout Workout
	json.NewDecoder(r.Body).Decode(&workout)
	_, err := db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("TABLE_WORKOUTS")),
		Item: map[string]*dynamodb.AttributeValue{
			"UserID": {
				S: aws.String(workout.UserID),
			},
			"WorkoutID": {
				S: aws.String(workout.WorkoutID),
			},
			"Name": {
				S: aws.String(workout.Name),
			},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error creating workout")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func AddExercise(w http.ResponseWriter, r *http.Request) {
	var exercise Exercise
	json.NewDecoder(r.Body).Decode(&exercise)
	_, err := db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("TABLE_EXERCISES")),
		Item: map[string]*dynamodb.AttributeValue{
			"ExerciseID": {S: aws.String(exercise.ExerciseID)},
			"Name":       {S: aws.String(exercise.Name)},
			"Tag":        {S: aws.String(exercise.Tag)},
			"Time":       {S: aws.String(exercise.Time)},
			"Level":      {S: aws.String(exercise.Level)},
			"Sets":       {N: aws.String(fmt.Sprint(exercise.Sets))},
			"Reps":       {N: aws.String(fmt.Sprint(exercise.Reps))},
			"Weight":     {N: aws.String(fmt.Sprint(exercise.Weight))},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error adding exercise")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	var workout Workout
	json.NewDecoder(r.Body).Decode(&workout)
	_, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(os.Getenv("TABLE_WORKOUTS")),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID":    {S: aws.String(workout.UserID)},
			"WorkoutID": {S: aws.String(workout.WorkoutID)},
		},
		UpdateExpression: aws.String("SET Name = :name, Exercises = :exercises"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name":      {S: aws.String(workout.Name)},
			":exercises": {L: []*dynamodb.AttributeValue{}},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error updating workout")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetExercises(w http.ResponseWriter, r *http.Request) {
	result, err := db.Scan(&dynamodb.ScanInput{
		TableName: aws.String(os.Getenv("TABLE_EXERCISES")),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error fetching exercises", err)
		return
	}
	json.NewEncoder(w).Encode(result.Items)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/workouts/{userId}", GetWorkout).Methods("GET")
	r.HandleFunc("/workouts", CreateWorkout).Methods("PUT")
	r.HandleFunc("/workouts/{workoutId}/exercises", AddExercise).Methods("POST")
	r.HandleFunc("/workouts", UpdateWorkout).Methods("POST")
	r.HandleFunc("/exercises", GetExercises).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", r))
}
