resource "aws_lambda_function" "api_handler" {
  filename         = "./main.zip"
  function_name    = "GymTrackerAPIHandler"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "main"
  runtime          = "go1.x"
  source_code_hash = filebase64sha256("./main.zip")

  environment {
    variables = {
      AWS_REGION              = "us-east-1"
      DYNAMODB_WORKOUTS_TABLE = aws_dynamodb_table.workouts.name
      DYNAMODB_EXERCISES_TABLE = aws_dynamodb_table.exercises.name
    }
  }
}
