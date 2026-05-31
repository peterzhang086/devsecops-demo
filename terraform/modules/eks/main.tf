variable "name_prefix"        { type = string }
variable "cluster_version"    { type = string }
variable "vpc_id"             { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "public_subnet_ids"  { type = list(string) }
variable "node_instance_type" { type = string }
variable "node_desired_size"  { type = number }

module "eks" {
  #checkov:skip=CKV_TF_1: Pinning to commit hash is a Week 2 supply-chain hardening item; version constraint used for now
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = var.name_prefix
  cluster_version = var.cluster_version

  vpc_id     = var.vpc_id
  subnet_ids = var.private_subnet_ids

  # Public endpoint is fine for demo; lock down in prod via cluster_endpoint_public_access_cidrs
  cluster_endpoint_public_access  = true
  cluster_endpoint_private_access = true

  # CRITICAL: control plane logs to CloudWatch — auditable
  cluster_enabled_log_types = [
    "api", "audit", "authenticator", "controllerManager", "scheduler"
  ]

  # Envelope encryption for K8s secrets at rest with a customer-managed KMS key
  cluster_encryption_config = {
    resources = ["secrets"]
  }

  # IRSA enabled by default in v20
  enable_irsa = true

  # API + ConfigMap auth — modern auth mode
  authentication_mode                      = "API_AND_CONFIG_MAP"
  enable_cluster_creator_admin_permissions = true

  eks_managed_node_groups = {
    default = {
      instance_types = [var.node_instance_type]
      min_size       = 1
      max_size       = 3
      desired_size   = var.node_desired_size

      # Bottlerocket is the security-conscious choice
      # (immutable, minimal, SELinux), but stick to AL2023 for simpler demo
      ami_type = "AL2023_x86_64_STANDARD"

      # Use IMDSv2 only — common CIS finding to call out
      metadata_options = {
        http_endpoint               = "enabled"
        http_tokens                 = "required"
        http_put_response_hop_limit = 2
      }

      labels = {
        role = "default"
      }
    }
  }

  # Enable common addons; pin versions in week 2 when locking down supply chain
  cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
    aws-ebs-csi-driver = {
      most_recent              = true
      service_account_role_arn = aws_iam_role.ebs_csi_driver.arn
    }
  }
}

resource "aws_iam_role" "ebs_csi_driver" {
  name = "${var.name_prefix}-ebs-csi-driver"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Federated = module.eks.oidc_provider_arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "${replace(module.eks.cluster_oidc_issuer_url, "https://", "")}:sub" = "system:serviceaccount:kube-system:ebs-csi-controller-sa"
          "${replace(module.eks.cluster_oidc_issuer_url, "https://", "")}:aud" = "sts.amazonaws.com"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ebs_csi_driver" {
  role       = aws_iam_role.ebs_csi_driver.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
}

output "cluster_name"     { value = module.eks.cluster_name }
output "cluster_arn"      { value = module.eks.cluster_arn }
output "cluster_endpoint" { value = module.eks.cluster_endpoint }
output "cluster_ca_data"  { value = module.eks.cluster_certificate_authority_data }
output "oidc_provider_arn" { value = module.eks.oidc_provider_arn }
