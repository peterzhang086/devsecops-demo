# Crypto Asset API — DevSecOps & Cloud Security Demo

End-to-end security engineering demo for a simulated crypto-asset management API,
showcasing DevSecOps tooling, AWS cloud security, Kubernetes runtime security,
and continuous compliance.

> Built as a portfolio piece for cloud security / DevSecOps engineering roles.

## Architecture

```
┌─────────────────┐
│  GitHub Actions │  ──OIDC──▶  AWS IAM Role  (no long-lived keys)
└────────┬────────┘
         │
         ├─ test          (go vet + go test -race)
         ├─ scan-code     (gosec SAST + govulncheck SCA)
         ├─ scan-iac      (checkov Terraform)
         │
         ▼
    build-and-push  (ECR — immutable tags, KMS encrypted)
         │
         ▼
     scan-image     (Trivy — CRITICAL/HIGH gate)
         │
         ▼
       deploy       (Kyverno admission control + Falco runtime detection)
         │
         ▼
        dast        (nuclei HTTP security scan)

Weekly: security-audit (Prowler CSPM — CIS AWS Benchmark)
```

```
   ┌──────────┐         ┌─────────────────────────────────────┐
   │   ECR    │────────▶│           EKS Cluster               │
   └──────────┘         │  ┌─────────────────────────────┐    │
                        │  │  crypto-asset-api (x2)      │    │
                        │  │  distroless · nonroot · RO  │    │
                        │  └──────────────┬──────────────┘    │
                        │                 │                    │
                        │  Kyverno  ·  Falco  ·  PSS          │
                        └─────────────────┼────────────────────┘
                                          │
                                 ┌────────▼────────┐
                                 │       ALB       │
                                 └─────────────────┘
```

## Security Controls

| Layer | Control | Tool |
|-------|---------|------|
| Source code | SAST | gosec |
| Dependencies | SCA | govulncheck |
| Infrastructure | IaC scanning | checkov |
| Container image | CVE scanning | Trivy |
| Kubernetes | Admission control | Kyverno |
| Runtime | Threat detection | Falco |
| Cloud posture | CSPM | Prowler |
| API | DAST | nuclei |
| Credentials | OIDC federation | GitHub Actions + AWS |

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

- [x] Week 1: Infra + app + CI/CD (EKS, ECR, OIDC, Go service)
- [x] Week 2: Security scanning pipeline (gosec, govulncheck, Trivy, checkov)
- [x] Week 3: Runtime security (Falco, Kyverno) + CSPM (Prowler)
- [x] Week 4: Pentest report + DAST (nuclei)
