# Crypto Asset API — DevSecOps & Cloud Security Demo

End-to-end security engineering demo for a simulated crypto-asset management API,
showcasing DevSecOps tooling, AWS cloud security, Kubernetes runtime security,
and continuous compliance.

> Built as a portfolio piece for cloud security / DevSecOps engineering roles
> in Web3 / DEX context.

## Architecture (Week 1 Baseline)

```
┌─────────────────┐
│  GitHub Actions │  ──OIDC──▶  AWS IAM Role  (no long-lived keys)
└────────┬────────┘
         │ push image
         ▼
   ┌──────────┐         ┌──────────────────────────────┐
   │   ECR    │────────▶│        EKS Cluster           │
   └──────────┘         │  ┌────────────────────────┐  │
                        │  │  crypto-asset-api      │  │
                        │  │  (Go service)          │  │
                        │  └───────────┬────────────┘  │
                        └──────────────┼───────────────┘
                                       │
                              ┌────────▼────────┐
                              │       ALB       │
                              └────────┬────────┘
                                       │
                                  Internet
```

## Quick Start

```bash
# 1. Configure AWS credentials
aws configure

# 2. Bootstrap state backend (one-time)
cd terraform/bootstrap && terraform init && terraform apply

# 3. Provision infrastructure
cd ../envs/dev && terraform init && terraform apply

# 4. Configure kubectl
aws eks update-kubeconfig --name demo-dev --region us-east-2

# 5. Deploy app (manual first time; CI takes over after)
kubectl apply -k k8s/overlays/dev
```

## Cost Guardrails

- Single NAT Gateway (~$32/month) — biggest cost
- 2× t3.small EKS nodes (~$30/month)
- ALB (~$16/month)
- **Estimated burn: ~$80/month if left running 24/7**
- **Strongly recommended:** `terraform destroy` when not actively demoing

## Status

- [x] Week 1: Infra + app + minimal CI/CD
- [ ] Week 2: Full security scanning pipeline (SAST/SCA/DAST/IaC)
- [ ] Week 3: Runtime security (Falco/Kyverno) + CSPM (Prowler)
- [ ] Week 4: Pentest report + custom tools
