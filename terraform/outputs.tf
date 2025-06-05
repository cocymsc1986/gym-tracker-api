output "cognito_user_pool_id" {
  value = aws_cognito_user_pool.gym_tracker_pool.id
}

output "cognito_user_pool_client_id" {
  value = aws_cognito_user_pool_client.gym_tracker_client.id
}

output "cognito_domain" {
  value = aws_cognito_user_pool_domain.gym_tracker_domain.domain
}