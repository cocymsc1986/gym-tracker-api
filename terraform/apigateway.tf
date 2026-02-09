resource "aws_api_gateway_rest_api" "gym_tracker_api" {
  name        = "GymTrackerAPI-${var.environment}"
  description = "API for tracking gym workouts - ${var.environment}"

  tags = {
    Environment = var.environment
    Project     = "gym-tracker"
  }
}

# ──────────────────────────────────────────────
# {proxy+} catch-all resource
# ──────────────────────────────────────────────

resource "aws_api_gateway_resource" "proxy_resource" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  parent_id   = aws_api_gateway_rest_api.gym_tracker_api.root_resource_id
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "proxy_any" {
  rest_api_id   = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id   = aws_api_gateway_resource.proxy_resource.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "proxy_lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id             = aws_api_gateway_resource.proxy_resource.id
  http_method             = aws_api_gateway_method.proxy_any.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.api_handler.invoke_arn

  depends_on = [
    aws_api_gateway_method.proxy_any
  ]
}

# OPTIONS mock integration on {proxy+} for CORS preflight
resource "aws_api_gateway_method" "proxy_options" {
  rest_api_id   = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id   = aws_api_gateway_resource.proxy_resource.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "proxy_options_integration" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id = aws_api_gateway_resource.proxy_resource.id
  http_method = aws_api_gateway_method.proxy_options.http_method
  type        = "MOCK"

  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }

  depends_on = [
    aws_api_gateway_method.proxy_options
  ]
}

resource "aws_api_gateway_method_response" "proxy_options_200" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id = aws_api_gateway_resource.proxy_resource.id
  http_method = aws_api_gateway_method.proxy_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = "Empty"
  }

  depends_on = [
    aws_api_gateway_method.proxy_options
  ]
}

resource "aws_api_gateway_integration_response" "proxy_options_integration_response" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id = aws_api_gateway_resource.proxy_resource.id
  http_method = aws_api_gateway_method.proxy_options.http_method
  status_code = aws_api_gateway_method_response.proxy_options_200.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${var.cors_allowed_origins}'"
    "method.response.header.Access-Control-Allow-Methods" = "'GET,POST,PUT,DELETE,OPTIONS'"
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,Authorization'"
  }

  depends_on = [
    aws_api_gateway_integration.proxy_options_integration
  ]
}

# ──────────────────────────────────────────────
# OPTIONS mock integration on root resource
# ──────────────────────────────────────────────

resource "aws_api_gateway_method" "root_options" {
  rest_api_id   = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id   = aws_api_gateway_rest_api.gym_tracker_api.root_resource_id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "root_options_integration" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id = aws_api_gateway_rest_api.gym_tracker_api.root_resource_id
  http_method = aws_api_gateway_method.root_options.http_method
  type        = "MOCK"

  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }

  depends_on = [
    aws_api_gateway_method.root_options
  ]
}

resource "aws_api_gateway_method_response" "root_options_200" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id = aws_api_gateway_rest_api.gym_tracker_api.root_resource_id
  http_method = aws_api_gateway_method.root_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = "Empty"
  }

  depends_on = [
    aws_api_gateway_method.root_options
  ]
}

resource "aws_api_gateway_integration_response" "root_options_integration_response" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  resource_id = aws_api_gateway_rest_api.gym_tracker_api.root_resource_id
  http_method = aws_api_gateway_method.root_options.http_method
  status_code = aws_api_gateway_method_response.root_options_200.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${var.cors_allowed_origins}'"
    "method.response.header.Access-Control-Allow-Methods" = "'GET,POST,PUT,DELETE,OPTIONS'"
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,Authorization'"
  }

  depends_on = [
    aws_api_gateway_integration.root_options_integration
  ]
}

# ──────────────────────────────────────────────
# Deployment
# ──────────────────────────────────────────────

resource "aws_api_gateway_deployment" "api_deployment" {
  rest_api_id = aws_api_gateway_rest_api.gym_tracker_api.id
  stage_name  = var.environment

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.proxy_resource.id,
      aws_api_gateway_method.proxy_any.id,
      aws_api_gateway_integration.proxy_lambda_integration.id,
      aws_api_gateway_method.proxy_options.id,
      aws_api_gateway_integration.proxy_options_integration.id,
      aws_api_gateway_integration_response.proxy_options_integration_response.id,
      aws_api_gateway_method.root_options.id,
      aws_api_gateway_integration.root_options_integration.id,
      aws_api_gateway_integration_response.root_options_integration_response.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    aws_api_gateway_method.proxy_any,
    aws_api_gateway_integration.proxy_lambda_integration,
    aws_api_gateway_method.proxy_options,
    aws_api_gateway_integration.proxy_options_integration,
    aws_api_gateway_method.root_options,
    aws_api_gateway_integration.root_options_integration
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
