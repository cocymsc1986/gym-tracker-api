resource "aws_api_gateway_rest_api" "gym_tracker_api" {
  name        = "GymTrackerAPI-${var.environment}"
  description = "API for tracking gym workouts - ${var.environment}"

  tags = {
    Environment = var.environment
    Project     = "gym-tracker"
  }
}

resource "aws_api_gateway_resource" "workouts_resource" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  parent_id   = aws_api_gateway_rest_api.gym_tracker_api.root_resource_id
  path_part   = "workouts"
}

resource "aws_api_gateway_method" "workouts_method" {
  rest_api_id   = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id   = aws_api_gateway_resource.workouts_resource.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id             = aws_api_gateway_resource.workouts_resource.id
  http_method             = aws_api_gateway_method.workouts_method.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.api_handler.invoke_arn

  depends_on = [
    aws_api_gateway_method.workouts_method
  ]
}

resource "aws_api_gateway_deployment" "api_deployment" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  stage_name  = var.environment

  depends_on = [
    aws_api_gateway_method.workouts_method,
    aws_api_gateway_integration.lambda_integration
  ]
}

resource "aws_api_gateway_api_key" "gym_tracker_api_key" {
  name        = "GymTrackerAPIKey-${var.environment}"
  description = "API key for securing Gym Tracker API - ${var.environment}"
  enabled     = true

  tags = {
    Environment = var.environment
    Project     = "gym-tracker"
  }
}

resource "aws_api_gateway_usage_plan" "gym_tracker_usage_plan" {
  name = "GymTrackerUsagePlan-${var.environment}"

  api_stages {
    api_id = aws_api_gateway_rest_api.gym_tracker_api.id
    stage  = aws_api_gateway_deployment.api_deployment.stage_name
  }
}

resource "aws_api_gateway_usage_plan_key" "api_key_association" {
  key_id        = aws_api_gateway_api_key.gym_tracker_api_key.id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.gym_tracker_usage_plan.id
}

output "api_endpoint" {
  value = aws_api_gateway_deployment.api_deployment.invoke_url
}
