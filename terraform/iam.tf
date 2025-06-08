resource "aws_iam_role" "lambda_exec" {
  name = "GymTrackerLambdaExecutionRole-${var.environment}"


  tags = {
    Environment = var.environment
    Project     = "gym-tracker"
  }

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = { Service = "lambda.amazonaws.com" }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy_attachment" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "lambda_dynamo_policy" {
  name = "LambdaDynamoDBPolicy-${var.environment}"
  role = aws_iam_role.lambda_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["dynamodb:PutItem", "dynamodb:GetItem", "dynamodb:UpdateItem", "dynamodb:Scan"]
        Effect   = "Allow"
        Resource = [
          aws_dynamodb_table.workouts.arn,
          aws_dynamodb_table.exercises.arn
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy" "lambda_cognito_policy" {
  name = "LambdaCognitoPolicy-${var.environment}"
  role = aws_iam_role.lambda_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "cognito-idp:AdminInitiateAuth",
          "cognito-idp:AdminCreateUser",
          "cognito-idp:AdminSetUserPassword",
          "cognito-idp:GetUser",
          "cognito-idp:InitiateAuth",
          "cognito-idp:SignUp",
          "cognito-idp:ConfirmSignUp",
          "cognito-idp:ResendConfirmationCode"
        ]
        Effect   = "Allow"
        Resource = aws_cognito_user_pool.gym_tracker_pool.arn
      }
    ]
  })
}
