package db

import (
	"fmt"

	"gym-tracker-api/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoExerciseRepository struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func NewDynamoExerciseRepository(db *dynamodb.DynamoDB, tableName string) *DynamoExerciseRepository {
	return &DynamoExerciseRepository{
		db:        db,
		tableName: tableName,
	}
}

func (r *DynamoExerciseRepository) GetByID(userID, exerciseID string) (*models.Exercise, error) {
	result, err := r.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {
				S: aws.String(userID),
			},
			"ExerciseID": {
				S: aws.String(exerciseID),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("exercise not found")
	}

	var exercise models.Exercise
	err = dynamodbattribute.UnmarshalMap(result.Item, &exercise)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal exercise: %w", err)
	}

	return &exercise, nil
}

func (r *DynamoExerciseRepository) ListByUserID(userID string) ([]*models.Exercise, error) {
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
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}

	var exercises []*models.Exercise
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &exercises)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal exercises: %w", err)
	}

	return exercises, nil
}

func (r *DynamoExerciseRepository) Create(userID string, exercise *models.Exercise) error {
	av, err := dynamodbattribute.MarshalMap(exercise)
	if err != nil {
		return fmt.Errorf("failed to marshal exercise: %w", err)
	}

	av["UserID"] = &dynamodb.AttributeValue{
		S: aws.String(userID),
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to create exercise: %w", err)
	}

	return nil
}

func (r *DynamoExerciseRepository) Update(userID string, exercise *models.Exercise) error {
	_, err := r.GetByID(userID, exercise.ExerciseID)
	if err != nil {
		return fmt.Errorf("exercise not found: %w", err)
	}

	av, err := dynamodbattribute.MarshalMap(exercise)
	if err != nil {
		return fmt.Errorf("failed to marshal exercise: %w", err)
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to update exercise: %w", err)
	}

	return nil
}

func (r *DynamoExerciseRepository) Delete(userID, exerciseID string) error {
	_, err := r.GetByID(userID, exerciseID)
	if err != nil {
		return fmt.Errorf("exercise not found: %w", err)
	}

	_, err = r.db.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {
				S: aws.String(userID),
			},
			"ExerciseID": {
				S: aws.String(exerciseID),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete exercise: %w", err)
	}

	return nil
}

func (r *DynamoExerciseRepository) ListByType(userID, exerciseType string) ([]*models.Exercise, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("ExerciseTypeIndex"),
		KeyConditionExpression: aws.String("ExerciseType = :exerciseType"),
		FilterExpression:       aws.String("UserID = :userID"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":exerciseType": {
				S: aws.String(exerciseType),
			},
			":userID": {
				S: aws.String(userID),
			},
		},
	}

	result, err := r.db.Query(queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list exercises by exerciseType: %w", err)
	}

	var exercises []*models.Exercise
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &exercises)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal exercises: %w", err)
	}

	return exercises, nil
}
