resource "aws_lambda_function" "api_handler" {
  filename         = var.lambda_zip_file
  function_name    = "GymTrackerAPIHandler"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "main"
  runtime          = "provided.al2"
  source_code_hash = filebase64sha256(var.lambda_zip_file)

  environment {
    variables = {
      DYNAMODB_WORKOUTS_TABLE = aws_dynamodb_table.workouts.name
      DYNAMODB_EXERCISES_TABLE = aws_dynamodb_table.exercises.name
    }
  }
}
