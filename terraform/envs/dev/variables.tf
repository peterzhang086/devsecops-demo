variable "region" {
  type    = string
  default = "us-east-2"
}

variable "project" {
  type    = string
  default = "demo"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "vpc_cidr" {
  type    = string
  default = "10.20.0.0/16"
}

variable "cluster_version" {
  type    = string
  default = "1.30"
}

variable "node_instance_type" {
  type    = string
  default = "t3.small"
}

variable "node_desired_size" {
  type    = number
  default = 2
}

# GitHub repo allowed to assume the OIDC role: "owner/repo"
variable "github_repo" {
  type        = string
  description = "GitHub repo in <owner>/<repo> form, e.g. peterzhang086/devsecops-demo"
}
