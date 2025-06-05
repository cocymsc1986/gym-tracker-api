package db

import (
	"fmt"
	"time"

	"gym-tracker-api/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoWorkoutRepository struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func NewDynamoWorkoutRepository(db *dynamodb.DynamoDB, tableName string) *DynamoWorkoutRepository {
	return &DynamoWorkoutRepository{
		db:        db,
		tableName: tableName,
	}
}

func (r *DynamoWorkoutRepository) GetByID(userID, workoutID string) (*models.Workout, error) {
	result, err := r.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {
				S: aws.String(userID),
			},
			"WorkoutID": {
				S: aws.String(workoutID),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("workout not found")
	}

	var workout models.Workout
	err = dynamodbattribute.UnmarshalMap(result.Item, &workout)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workout: %w", err)
	}

	return &workout, nil
}

func (r *DynamoWorkoutRepository) ListByUserID(userID string) ([]*models.Workout, error) {
	result, err := r.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("UserID = :userID"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userID": {
				S: aws.String(userID),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query workouts: %w", err)
	}

	var workouts []*models.Workout
	for _, item := range result.Items {
		var workout models.Workout
		err = dynamodbattribute.UnmarshalMap(item, &workout)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal workout: %w", err)
		}
		workouts = append(workouts, &workout)
	}

	return workouts, nil
}

func (r *DynamoWorkoutRepository) Create(workout *models.Workout) error {
	if workout.CreatedAt.IsZero() {
		workout.CreatedAt = time.Now()
	}

	item, err := dynamodbattribute.MarshalMap(workout)
	if err != nil {
		return fmt.Errorf("failed to marshal workout: %w", err)
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		ConditionExpression: aws.String("attribute_not_exists(UserID) AND attribute_not_exists(WorkoutID)"),
	})
	if err != nil {
		return fmt.Errorf("failed to create workout: %w", err)
	}

	return nil
}

func (r *DynamoWorkoutRepository) Update(workout *models.Workout) error {
	item, err := dynamodbattribute.MarshalMap(workout)
	if err != nil {
		return fmt.Errorf("failed to marshal workout: %w", err)
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		ConditionExpression: aws.String("attribute_exists(UserID) AND attribute_exists(WorkoutID)"),
	})
	if err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}

	return nil
}

func (r *DynamoWorkoutRepository) Delete(workoutID string, userID string) error {
	_, err := r.db.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {
				S: aws.String(userID),
			},
			"WorkoutID": {
				S: aws.String(workoutID),
			},
		},
		ConditionExpression: aws.String("attribute_exists(UserID) AND attribute_exists(WorkoutID)"),
	})
	if err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	return nil
}