resource "aws_dynamodb_table" "workouts" {
  name         = "Workouts"
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
}

resource "aws_dynamodb_table" "exercises" {
  name         = "Exercises"
  billing_mode = "PAY_PER_REQUEST"

  hash_key = "ExerciseID"

  attribute {
    name = "ExerciseID"
    type = "S"
  }
}
