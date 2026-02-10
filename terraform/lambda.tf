resource "aws_lambda_function" "api_handler" {
  filename         = var.lambda_zip_file
  function_name    = "GymTrackerAPIHandler-${var.environment}"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "bootstrap"
  runtime          = "provided.al2"
  source_code_hash = filebase64sha256(var.lambda_zip_file)

  environment {
    variables = {
      ENVIRONMENT          = var.environment
      DYNAMO_TABLE_WORKOUTS  = aws_dynamodb_table.workouts.name
      DYNAMO_TABLE_EXERCISES = aws_dynamodb_table.exercises.name
      COGNITO_USER_POOL_ID = aws_cognito_user_pool.gym_tracker_pool.id
      COGNITO_CLIENT_ID    = aws_cognito_user_pool_client.gym_tracker_client.id
      CORS_ALLOWED_ORIGINS = var.cors_allowed_origins
    }
  }

  tags = {
    Environment = var.environment
    Project     = "gym-tracker"
  }
}

resource "aws_lambda_permission" "api_gateway_invoke" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api_handler.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.gym_tracker_api.execution_arn}/*/*"
}
