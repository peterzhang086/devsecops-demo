# Week 1 Execution Runbook

Day-by-day checklist to take this skeleton from zero to "git push deploys to EKS".

## Prerequisites

- AWS account with admin access (personal account is fine; set billing alert at $50)
- AWS CLI v2 configured locally
- Terraform >= 1.6
- kubectl, kustomize, helm
- Docker (for local app testing)
- GitHub repo created (private is fine)

## Day 1 — Bootstrap

```bash
# Create the repo on GitHub first, then:
git clone git@github.com:peterzhang086/devsecops-demo.git
cd devsecops-demo
# Drop this skeleton in, commit.

# Bootstrap state backend
cd terraform/bootstrap
terraform init
terraform apply
# Note the outputs: state_bucket, lock_table
```

Update `terraform/envs/dev/providers.tf` backend block with the actual bucket name
and uncomment the lines.

## Day 2 — VPC + ECR

```bash
cd terraform/envs/dev
cp example.tfvars terraform.tfvars
# Edit terraform.tfvars: set github_repo = "peterzhang086/devsecops-demo"

terraform init
# Target apply to bring up just network + registry first — faster feedback
terraform apply -target=module.vpc -target=module.ecr
```

Verify in AWS Console: VPC created with 4 subnets, NAT GW, IGW, Flow Logs
flowing to CloudWatch.

## Day 3 — EKS

```bash
terraform apply -target=module.eks
# Takes ~15 minutes. Coffee.

aws eks update-kubeconfig --name demo-dev --region us-east-2
kubectl get nodes  # Should see 2 nodes Ready
kubectl get pods -A  # coredns, kube-proxy, aws-node, ebs-csi running
```

## Day 4 — GitHub OIDC + first manual deploy

```bash
terraform apply  # finalize everything including OIDC role

# Grab the role ARN from output
terraform output github_actions_role_arn
```

In GitHub repo Settings → Secrets and variables → Actions → Variables tab:
- Add **repository variable** `AWS_ROLE_ARN` with the value above

Build and push the app manually once to verify the path:
```bash
ECR=$(terraform output -raw ecr_repository_url)
aws ecr get-login-password --region us-east-2 | docker login --username AWS --password-stdin $ECR
cd ../../../app
docker build -t $ECR:manual-test .
docker push $ECR:manual-test
```

Apply manifests with the manual tag:
```bash
cd ../k8s/overlays/dev
kustomize edit set image REPLACE_IN_OVERLAY=$ECR:manual-test
kustomize build . | kubectl apply -f -
kubectl -n crypto-api get pods  # should see 2/2 Running
```

## Day 5 — Local DX + tests

```bash
cd app
go mod tidy
go test ./...
go run ./cmd  # curl localhost:8080/healthz to verify
```

Fix any issues. Commit cleanly.

## Day 6 — First CI run

Push to main. Watch GitHub Actions:
- `test` job runs `go vet` + `go test`
- `build-and-push` builds image, pushes to ECR with timestamp-sha tag
- `deploy` runs `kustomize edit set image` + `kubectl apply` + waits for rollout

If OIDC fails: verify the `sub` claim format matches the trust policy
(`repo:owner/repo:ref:refs/heads/main`).

## Day 7 — Expose via ALB + smoke test

Install AWS Load Balancer Controller (Helm):
```bash
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=demo-dev \
  --set serviceAccount.create=true
```

Add an Ingress object (week 2 will move this into k8s/base):
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: crypto-asset-api
  namespace: crypto-api
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
spec:
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: crypto-asset-api
                port:
                  number: 80
```

Get the ALB DNS, hit `/healthz`. Done.

## Daily Hygiene

```bash
# At end of every working session
cd terraform/envs/dev
terraform destroy  # nuke everything; bootstrap state stays
```

Or at minimum, scale node group to 0:
```bash
aws eks update-nodegroup-config --cluster-name demo-dev \
  --nodegroup-name default --scaling-config minSize=0,maxSize=3,desiredSize=0
```

## Definition of Done — Week 1

- [ ] `git push` on main → app live on ALB within 10 min
- [ ] No long-lived AWS credentials anywhere (OIDC only)
- [ ] CloudTrail, VPC Flow Logs, EKS control plane logs all flowing
- [ ] Pod runs as non-root, read-only fs, no caps, restricted PSS
- [ ] NetworkPolicy default-deny in place
- [ ] README architecture diagram updated to match reality
