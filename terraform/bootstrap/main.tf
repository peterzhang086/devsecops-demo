# Bootstrap: creates the S3 bucket + DynamoDB table that hold Terraform state
# for all other environments. Run once. Uses local state itself.

terraform {
  required_version = ">= 1.6"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

variable "region" {
  type    = string
  default = "us-east-2"
}

variable "project" {
  type    = string
  default = "demo"
}

resource "random_id" "suffix" {
  byte_length = 4
}

# State bucket — encrypted, versioned, public access blocked, object lock-ready
resource "aws_s3_bucket" "tfstate" {
  #checkov:skip=CKV_AWS_144: Cross-region replication not required for demo state backend
  #checkov:skip=CKV2_AWS_61: Lifecycle policy not needed for Terraform state bucket
  #checkov:skip=CKV_AWS_145: AES256 encryption is sufficient for demo; KMS adds cost
  #checkov:skip=CKV2_AWS_62: Event notifications not required for demo state bucket
  #checkov:skip=CKV_AWS_18: Access logging not required for demo state bucket
  bucket = "${var.project}-tfstate-${random_id.suffix.hex}"

  # Demo environment — prod would set lifecycle prevent_destroy = true
}

resource "aws_s3_bucket_versioning" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "tfstate" {
  bucket                  = aws_s3_bucket.tfstate.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# State lock table
resource "aws_dynamodb_table" "tflock" {
  #checkov:skip=CKV_AWS_119: CMK not required for demo lock table; AWS-managed key is sufficient
  #checkov:skip=CKV_AWS_28: PITR not needed for Terraform state lock table
  name         = "${var.project}-tflock"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }
}

output "state_bucket" {
  value = aws_s3_bucket.tfstate.id
}

output "lock_table" {
  value = aws_dynamodb_table.tflock.name
}
