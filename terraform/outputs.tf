# Deployment Output
output "api_endpoint" {
  value = aws_apigatewayv2_stage.api_stage.invoke_url
}
