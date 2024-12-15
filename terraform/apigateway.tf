resource "aws_apigatewayv2_api" "gym_tracker_api" {
  name          = "GymTrackerAPI"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_api_key" "gym_tracker_api_key" {
  name        = "GymTrackerAPIKey"
  description = "API key for securing Gym Tracker API"
  enabled     = true
}

resource "aws_apigatewayv2_usage_plan" "gym_tracker_usage_plan" {
  api_stages {
    api_id = aws_apigatewayv2_api.gym_tracker_api.id
    stage  = aws_apigatewayv2_stage.api_stage.name
  }

  name = "GymTrackerUsagePlan"
}

resource "aws_apigatewayv2_usage_plan_key" "api_key_association" {
  key_id        = aws_apigatewayv2_api_key.gym_tracker_api_key.id
  key_type      = "API_KEY"
  usage_plan_id = aws_apigatewayv2_usage_plan.gym_tracker_usage_plan.id
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id             = aws_apigatewayv2_api.gym_tracker_api.id
  integration_type   = "AWS_PROXY"
  integration_uri    = aws_lambda_function.api_handler.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "workouts_route" {
  api_id    = aws_apigatewayv2_api.gym_tracker_api.id
  route_key = "ANY /workouts/{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_stage" "api_stage" {
  api_id      = aws_apigatewayv2_api.gym_tracker_api.id
  name        = "production"
  auto_deploy = true
}
