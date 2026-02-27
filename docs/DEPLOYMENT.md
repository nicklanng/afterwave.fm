# Deployment strategy

How we build, test, and ship the application and infrastructure. Architecture (what we run) is in [Architecture](./ARCHITECTURE.md). Vision caps team size; deployment should be operable by that team without heavy orchestration.

## Implementation checklist

- ~~Make test (or equivalent) for API tests~~
- ~~Local dev: API + optional frontend; DynamoDB Local / env config~~
- CI/CD: GitHub Actions — build, test, lint on push/PR
- Deploy: Docker image → ECR; ECS Fargate; ALB; Terraform for infra
- Secrets: AWS Secrets Manager; no secrets in repo
- Staging and production environments; branch/tag strategy
- Health check endpoint for ALB/orchestrator
- Monitoring: Grafana Cloud (or equivalent); logs, metrics, alerts
- Rollback: redeploy previous ECS task definition

---

## Environments

- **Local** — Developers run the API (and optionally frontend) on their machine. Config via env (e.g. `.env` not in repo, or defaults). Local DynamoDB or DynamoDB Local for dev; optional local Stripe test mode.
- **Staging** — Mirrors production (same services, smaller scale). Used for QA and integration before release. Branch: deploy from `main` or a release branch; or tag-based. Staging URL TBD (e.g. staging.afterwave.fm or a separate domain).
- **Production** — [www.afterwave.fm](http://www.afterwave.fm), *.afterwave.fm (artist pages), api.afterwave.fm. Deploy from `main` on merge, or from a tag/release branch. Manual approval for prod deploys is recommended until the pipeline is trusted.

**Branch strategy:** `main` is the default branch. Feature branches merge into `main`. Optionally: release branches (e.g. `release/1.0`) for a formal release cycle; or continuous deploy from `main` with tags for versioning. Decide when we add CI.

---

## CI/CD

We use **GitHub Actions** for CI/CD: build, test, and deploy on push (or PR) to `main`.

- **Build** — On push (or on PR) to `main`: run tests, lint (e.g. `go vet`, staticcheck, or golangci-lint), build the API binary. Fail the pipeline if tests or lint fail.
- **Deploy** — On successful build (and optionally manual approval for prod): build Docker image, push to ECR, update ECS service (new task definition revision). GitHub Actions: build → test → (optional manual approval for prod) → deploy to ECS Fargate. Exact workflow (branch → env mapping, approval gates) TBD.
- **Secrets** — Not in repo. **AWS Secrets Manager** for API secrets (JWT signing keys, Stripe keys, DB config, etc.). CI or runtime fetches from Secrets Manager; injected as env vars at startup. GitHub Actions uses OIDC or stored credentials to assume an AWS role and deploy; no long-lived access keys in GitHub where possible.

**Goal:** One command (or one merge) from code to running service; clear rollback path. Prefer simple pipeline (build → test → deploy) over multi-stage orchestration.

---

## Infrastructure (where we host)

- **Provider** — AWS as primary. Design supports **multiple global sites** (API in several regions) with **one central DynamoDB**; see [Architecture](./ARCHITECTURE.md).
- **API** — **api.afterwave.fm**. Versioned paths (e.g. `/v1/users/me`). Compute: **containers on ECS (Fargate)**; ALB in front; TLS terminated at ALB. API does **not** use Host for routing; all API traffic goes to api.afterwave.fm.
- **Frontend (web)** — **React + Vite + Bun** build; **served on a CDN**, preferably **AWS CloudFront** in front of **S3** (static site hosting). Same origin for [www.afterwave.fm](http://www.afterwave.fm) and *.afterwave.fm (or routing by Host at CloudFront). Frontend decides what to render based on domain (www vs artist handle); API is always api.afterwave.fm/v1/...
- **Database** — DynamoDB (managed). One primary region (or DynamoDB Global Tables for multi-region reads). **KMS** for per-customer or sensitive-field encryption in DynamoDB where required. Backup and point-in-time recovery via AWS.
- **Object storage** — **S3**. Buckets for music, images; IAM and bucket policy for least privilege. No public bucket; access via **presigned CloudFront (or S3) URLs** issued by the API for signed-in users only. CloudFront in front of S3 for media delivery.
- **DNS** — Route 53. Records: [www.afterwave.fm](http://www.afterwave.fm), *.afterwave.fm, api.afterwave.fm. SSL: ACM certificates (e.g. afterwave.fm, *.afterwave.fm, api.afterwave.fm).
- **Email** — **AWS SES** for transactional email (signup, password reset, invites, notifications). See [Architecture](./ARCHITECTURE.md).
- **Search** — **OpenSearch (AWS)**. Managed OpenSearch Service for discovery: free-text search (artist name, handle, bio), filters (genre, location), and relevance ranking. API indexes artist data into OpenSearch (sync from DynamoDB on write or batch); discovery and browse queries hit OpenSearch. Terraform defines the OpenSearch domain (and any required VPC/security group). See [Discovery](./DISCOVERY.md).

**Provisioning:** We use **Terraform** for infrastructure as code: all AWS resources (VPC, ECS, DynamoDB, S3, OpenSearch, Route 53, ACM, etc.) are defined in Terraform (e.g. `infra/` or `terraform/` in the repo). Apply via CI or manually; state in S3 + DynamoDB lock. No manual one-off resources for core infra; codify from the start so staging and prod stay in sync.

### Containers on ECS Fargate

We deploy the API as **containers on ECS (Fargate)**. No EC2 to manage; we run tasks behind an ALB.

- **Build** — CI builds a Docker image (e.g. multi-stage: compile Go in builder, copy binary into minimal runtime image). Push image to **ECR**.
- **Run** — Terraform defines ECS cluster (or default), task definition, service, and ALB. ALB routes api.afterwave.fm to the ECS service. Fargate runs the tasks; we scale by task count.
- **Deploy** — GitHub Actions builds the image, pushes to ECR, and updates the ECS service (new task definition revision). Rollback = deploy previous task definition.

Terraform defines ECS, ALB, and target group. We don’t use EC2 (we’d manage instances and process managers), EKS (more than we need for one API), or Lambda (our API is a long-lived HTTP server).

Terraform defines ECS, ALB, and target group. We don't use EC2 (we'd manage instances and process managers), EKS (more than we need for one API), or Lambda (our API is a long-lived HTTP server).

### Rate limiting (ALB + WAF)

**ALB does not support rate limiting.** To throttle or block abusive traffic (e.g. brute-force on `/auth/login`, signup abuse), put **AWS WAF** in front of the ALB and add **rate-based rules**:

- **Associate WAF with the ALB** — In the AWS Console or Terraform: create a WAF Web ACL and associate it with the ALB that fronts the Fargate service. All traffic to the API then passes through WAF before reaching the ALB.
- **Rate-based rule** — In WAF, add a rule of type "Rate-based rule". You define:
  - **Limit** — e.g. 2000 requests per 5 minutes per IP (minimum is 100 per 5 minutes).
  - **Evaluation window** — 60, 120, 300, or 600 seconds.
  - **Scope** — Typically "IP" so each client is limited by source IP. You can also scope by URI path (e.g. stricter limit for `/v1/auth/login` and `/v1/auth/signup`) using a separate rule or WAF scope-down.
- **Action** — Block or count when the limit is exceeded. Start with "Count" to tune the threshold, then switch to "Block".

**Terraform:** Use `aws_wafv2_web_acl` and `aws_wafv2_web_acl_association` to attach the Web ACL to the ALB. Add an `aws_wafv2_rule_group` or inline `rate_based_statement` in a rule.

**Note:** If you put **CloudFront** in front of the ALB, you can associate WAF with the CloudFront distribution instead; rate limiting then applies at the edge. Preserve the client IP (e.g. CloudFront forwards `X-Forwarded-For` or the WAF "forwarded IP" config) so rate-based rules count per real client.

---

## Subdomains in deployment (frontend only)

- **DNS** — [www.afterwave.fm](http://www.afterwave.fm) and *.afterwave.fm → **CloudFront** (frontend). api.afterwave.fm → **ALB** (API). One certificate for *.afterwave.fm; separate cert or SAN for api.afterwave.fm.
- **Frontend** — Same CloudFront distribution (or same S3 origin) serves www and artist subdomains. The **frontend** reads the Host and decides what to render (landing/discovery vs artist page); the API never sees Host for routing. Handles are resolved by the frontend calling api.afterwave.fm/v1/artists/:handle.
- **API** — No subdomain routing; all requests to api.afterwave.fm/v1/...

---

## Secrets and config

- **AWS Secrets Manager** — JWT signing keys, Stripe secret key, DB config, SES credentials, etc. **KMS** for per-customer or sensitive data encryption in DynamoDB (see Architecture). No secrets in repo.
- **How we inject** — Local: `.env` or shell export; file in `.gitignore`. CI/staging/prod: API fetches from **Secrets Manager** at startup (or CI injects env from Secrets Manager). Mapped into process env. No default secrets in code; required vars fail fast at startup.

---

## Monitoring and rollback

- **Observability** — **Grafana Cloud** for metrics, logs, and traces to start. We export OpenTelemetry (and optionally logs) from the API to Grafana Cloud. Can bring observability in-house (e.g. self-hosted Grafana, Loki, Tempo, or other stack) if cost becomes an issue.
- **Health** — HTTP health check (e.g. `GET /health` or `GET /live`) returning 200. ALB or orchestrator uses it. Endpoint does minimal work (no DB if possible, or a cheap DynamoDB read).
- **Logging** — Structured logs (slog JSON) to stdout; collected and sent to Grafana Cloud (or the platform’s log collector that forwards there). Include trace_id in logs when OpenTelemetry is used.
- **Alerts** — Basic: deploy failure, health check failure, 5xx rate above threshold. Configure in Grafana Cloud (or CI). Channel: email or Slack. Avoid alert fatigue; only alert when someone should act.
- **Rollback** — If deploy is bad: deploy previous ECS task definition (last known good). Prefer “redeploy last known good” over complex rollback logic. Database migrations: backward-compatible changes so old code still works during rollback.

---

## Team and operability

- **Who deploys** — Small team; any engineer with access can deploy. Document steps in README or runbook. Optional: require approval for prod (e.g. GitHub Environments, or manual “promote” step).
- **Scope** — One API service, one frontend (when we have it), DynamoDB, S3, Stripe. No microservices until we have a clear need. Keep deployment understandable so a single person can own it.

---

## Open decisions

- Staging URL and whether staging shares AWS account with prod or uses a separate account.
- GitHub Actions: exact workflow (branch → env), approval gates for prod, OIDC vs stored AWS credentials.
- Terraform: repo layout (`infra/` vs `terraform/`), state backend (S3 + DynamoDB lock), and whether to run Terraform apply in GitHub Actions or manually.
- Docker image: multi-stage build and minimal runtime image (e.g. scratch or alpine) — exact Dockerfile layout TBD.

