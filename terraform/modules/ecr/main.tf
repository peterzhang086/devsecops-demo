variable "name_prefix"     { type = string }
variable "repository_name" { type = string }

resource "aws_ecr_repository" "this" {
  name                 = "${var.name_prefix}/${var.repository_name}"
  image_tag_mutability = "IMMUTABLE"  # signed tags only — important talking point

  image_scanning_configuration {
    scan_on_push = true  # ECR-native scanning (will add Trivy as second opinion)
  }

  encryption_configuration {
    encryption_type = "KMS"  # default AWS-managed KMS key
  }
}

# Keep only most recent images to avoid storage bloat
resource "aws_ecr_lifecycle_policy" "this" {
  repository = aws_ecr_repository.this.name
  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep last 20 images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 20
      }
      action = { type = "expire" }
    }]
  })
}

output "repository_url" { value = aws_ecr_repository.this.repository_url }
output "repository_arn" { value = aws_ecr_repository.this.arn }
