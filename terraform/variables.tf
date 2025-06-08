variable "environment" {
  description = "Environment name (test/prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "lambda_zip_file" {
  type = string
}

variable "cognito_user_pool_name" {
  description = "Name of the Cognito User Pool"
  type        = string
  default     = "gym-tracker-users"
}

variable "cognito_client_name" {
  description = "Name of the Cognito User Pool Client"
  type        = string
  default     = "gym-tracker-client"
}
