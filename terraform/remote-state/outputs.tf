output "s3_bucket_name" {
  description = "Name of the S3 bucket for terraform state"
  value       = aws_s3_bucket.terraform_state.bucket
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table for terraform locks"
  value       = aws_dynamodb_table.terraform_locks.name
}

output "aws_region" {
  description = "AWS region"
  value       = var.aws_region
}