locals {
  name_prefix = "${var.project}-${var.environment}"
}

module "vpc" {
  source = "../../modules/vpc"

  name_prefix = local.name_prefix
  vpc_cidr    = var.vpc_cidr
  region      = var.region
}

module "ecr" {
  source = "../../modules/ecr"

  name_prefix    = local.name_prefix
  repository_name = "crypto-asset-api"
}

module "eks" {
  source = "../../modules/eks"

  name_prefix         = local.name_prefix
  cluster_version     = var.cluster_version
  vpc_id              = module.vpc.vpc_id
  private_subnet_ids  = module.vpc.private_subnet_ids
  public_subnet_ids   = module.vpc.public_subnet_ids
  node_instance_type  = var.node_instance_type
  node_desired_size   = var.node_desired_size
}

module "github_oidc" {
  source = "../../modules/iam-oidc"

  name_prefix  = local.name_prefix
  github_repo  = var.github_repo
  ecr_repo_arn = module.ecr.repository_arn
  eks_cluster_arn = module.eks.cluster_arn
}

output "cluster_name" {
  value = module.eks.cluster_name
}

output "cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "ecr_repository_url" {
  value = module.ecr.repository_url
}

output "github_actions_role_arn" {
  value       = module.github_oidc.role_arn
  description = "Set this as AWS_ROLE_ARN secret/variable in GitHub Actions"
}

output "kubeconfig_command" {
  value = "aws eks update-kubeconfig --name ${module.eks.cluster_name} --region ${var.region}"
}

resource "aws_eks_access_entry" "github_actions" {
  cluster_name  = module.eks.cluster_name
  principal_arn = module.github_oidc.role_arn
  type          = "STANDARD"
}

resource "aws_eks_access_policy_association" "github_actions_deploy" {
  cluster_name  = module.eks.cluster_name
  principal_arn = module.github_oidc.role_arn
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSAdminPolicy"

  access_scope {
    type = "cluster"
  }

  depends_on = [aws_eks_access_entry.github_actions]
}
