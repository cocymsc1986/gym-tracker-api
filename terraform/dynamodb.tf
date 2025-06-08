resource "aws_dynamodb_table" "workouts" {
  name         = "Workouts-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"

  hash_key = "UserID"
  range_key = "WorkoutID"

  attribute {
    name = "UserID"
    type = "S"
  }

  attribute {
    name = "WorkoutID"
    type = "S"
  }


  tags = {
    Environment = var.environment
    Project     = "gym-tracker"
  }
}

resource "aws_dynamodb_table" "exercises" {
  name         = "Exercises-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"

  hash_key = "ExerciseID"

  attribute {
    name = "ExerciseID"
    type = "S"
  }

  attribute {
    name = "ExerciseType"
    type = "S"
  }

  global_secondary_index {
    name            = "ExerciseTypeIndex"
    hash_key        = "ExerciseType"
    projection_type = "ALL"
  }


  tags = {
    Environment = var.environment
    Project     = "gym-tracker"
  }
}
