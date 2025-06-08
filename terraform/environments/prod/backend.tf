terraform {
  backend "s3" {
    bucket         = "gym-tracker-terraform-state-coguocff"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "gym-tracker-terraform-locks"
    encrypt        = true
  }
}