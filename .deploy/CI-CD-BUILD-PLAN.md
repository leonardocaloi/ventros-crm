# CI/CD Build & Deployment Plan - Ventros CRM

**Version**: 1.0
**Date**: 2025-10-12
**Status**: 🟢 Production-Ready

---

## 📋 Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Build Strategy](#build-strategy)
4. [Workflow Design](#workflow-design)
5. [Variable Structure](#variable-structure)
6. [GitHub Actions Workflows](#github-actions-workflows)
7. [AWX Integration](#awx-integration)
8. [Security & Secrets](#security--secrets)
9. [Deployment Environments](#deployment-environments)
10. [Rollback Strategy](#rollback-strategy)
11. [Monitoring & Observability](#monitoring--observability)

---

## 🎯 Executive Summary

### Current State
- ✅ Ansible role ready in `.deploy/ansible/ventros_crm/`
- ✅ Multi-stage Containerfile configured
- ✅ Helm charts structure defined
- ✅ AWX playbook configured
- ⚠️ Build process needs definition

### Target Workflow
```
git push → GitHub Actions → Build/Test → Publish Artifacts → AWX Trigger → Deploy to K8s
```

### Key Decision: **GitHub Actions for Build, AWX for Deploy**

**Why GitHub Actions for Build:**
- ✅ Free for public repos, generous limits for private
- ✅ Native Git integration (automatic triggers)
- ✅ Docker layer caching (faster builds)
- ✅ Matrix testing (multiple Go versions)
- ✅ Build artifacts separate from deployment
- ✅ Clear separation of concerns

**Why AWX for Deploy:**
- ✅ Already configured with K8s clusters
- ✅ Centralized deployment management
- ✅ Ansible Vault for secrets
- ✅ RBAC and audit logs
- ✅ Survey variables for dynamic configs
- ✅ Rollback capabilities

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         GIT PUSH (main)                         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      GITHUB ACTIONS                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Build &    │  │   Publish    │  │   Publish    │         │
│  │   Test       │→ │   Docker     │→ │   Helm       │         │
│  │   (Go 1.25)  │  │   Image      │  │   Chart      │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│         │                  │                  │                 │
└─────────┼──────────────────┼──────────────────┼─────────────────┘
          │                  │                  │
          ▼                  ▼                  ▼
    ┌─────────┐      ┌─────────────┐   ┌─────────────┐
    │  Tests  │      │   Docker    │   │    Helm     │
    │  Pass   │      │   Registry  │   │  Repository │
    └─────────┘      │  (DockerHub)│   │  (GH Pages) │
          │          └──────┬──────┘   └──────┬──────┘
          │                 │                  │
          └─────────────────┴──────────────────┘
                            │
                            ▼
          ┌─────────────────────────────────┐
          │    Trigger AWX (API/Webhook)    │
          └─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                            AWX                                  │
│  ┌──────────────────────────────────────────────────────┐     │
│  │  Job Template: "Deploy Ventros CRM"                  │     │
│  │  - Read vars from git (.deploy/ansible/global_vars) │     │
│  │  - Execute playbook (ventros_crm role)              │     │
│  │  - Use published Docker image (from variable)        │     │
│  │  - Use published Helm chart (from variable)          │     │
│  └──────────────────────────────────────────────────────┘     │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    KUBERNETES CLUSTER                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  PostgreSQL  │  │   RabbitMQ   │  │    Redis     │         │
│  │  (Zalando)   │  │  (Operator)  │  │   (Helm)     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│  ┌──────────────┐  ┌──────────────┐                            │
│  │   Temporal   │  │  Ventros CRM │                            │
│  │   (Helm)     │  │  Deployment  │◄── Main Application        │
│  └──────────────┘  └──────────────┘                            │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔨 Build Strategy

### Phase 1: Build & Test (GitHub Actions)

**Responsibilities:**
1. Run tests (unit, integration, e2e)
2. Build Go binary
3. Generate Swagger docs
4. Build Docker image
5. Push to Docker Registry
6. Package Helm chart
7. Publish Helm chart to GitHub Pages

**Triggers:**
- `push` to `main` branch → Build + Deploy to Staging
- `push` to `develop` branch → Build only (no deploy)
- `tag` matching `v*` → Build + Deploy to Production
- `pull_request` → Build + Test only

**Artifacts Generated:**
```
leonardocaloi/ventros-crm:0.1.0          ← Docker image
leonardocaloi/ventros-crm:latest          ← Docker image (main)
leonardocaloi/ventros-crm:pr-123          ← Docker image (PR)
ventros/ventros-crm-0.1.0.tgz            ← Helm chart
```

### Phase 2: Deploy (AWX)

**Responsibilities:**
1. Pull configuration from Git
2. Execute Ansible playbook
3. Deploy to Kubernetes via Helm
4. Verify deployment health
5. Send notifications

**Triggers:**
- GitHub Actions webhook (after successful build)
- Manual execution (via AWX UI)
- Scheduled deployments (e.g., maintenance windows)

**Configuration Source:**
- `.deploy/ansible/global_vars.yml` (git repository)
- AWX Credentials (Ansible Vault)
- AWX Survey (runtime overrides)

---

## 🔄 Workflow Design

### Workflow 1: Build & Publish (Continuous Integration)

**File**: `.github/workflows/build-and-publish.yaml`

**Flow:**
1. Checkout code
2. Setup Go 1.25.1
3. Run `make test-unit` (fast, no dependencies)
4. Build Docker image (multi-stage)
5. Run `make test-integration` (with Docker Compose)
6. Push Docker image to DockerHub
7. Package Helm chart
8. Publish Helm chart to GitHub Pages
9. Create GitHub Release (if tag)

**Environment Variables:**
- `DOCKER_USERNAME`: DockerHub username
- `DOCKER_PASSWORD`: DockerHub token
- `HELM_REPO`: GitHub Pages URL

### Workflow 2: Deploy Staging (Continuous Deployment)

**File**: `.github/workflows/deploy-staging.yaml`

**Flow:**
1. Wait for "Build & Publish" to complete
2. Call AWX API to trigger deployment
3. Pass image tag and chart version
4. Wait for AWX job completion
5. Verify deployment health
6. Send Slack notification

**Trigger**: Automatic after `push` to `main`

### Workflow 3: Deploy Production (Manual)

**File**: `.github/workflows/deploy-production.yaml`

**Flow:**
1. Manual approval required (GitHub Environment)
2. Call AWX API with production credentials
3. Deploy using stable tag (not `latest`)
4. Run smoke tests
5. Send notifications

**Trigger**: Manual via GitHub UI or tag push

---

## 📝 Variable Structure

### Current Structure (Keep This!)

```
.deploy/ansible/
├── global_vars.yml              ← Actual values (git-tracked)
└── ventros_crm/
    ├── tasks/
    │   └── main.yml             ← Playbook tasks
    ├── vars/
    │   └── main.yml             ← Variable templates with defaults
    └── templates/
        └── values.yml.j2        ← Helm values template
```

### Variable Flow

```
GitHub Actions                    AWX                          Kubernetes
─────────────────                ─────                        ──────────
IMAGE_TAG=0.1.0    ────────────▶ Extra Vars    ────────────▶ Helm Values
CHART_VERSION=0.1.0              (Runtime)                    (Rendered)
                                      │
                                      ▼
                                global_vars.yml ────────────▶ Jinja2
                                (Git Tracked)                 Template
                                      │
                                      ▼
                                vars/main.yml
                                (Defaults)
```

### Enhanced global_vars.yml

```yaml
---
# ============================================================================
# VENTROS CRM - Global Configuration
# ============================================================================
# This file is git-tracked and read by AWX on every deployment.
# Override values at runtime using AWX Extra Variables or Survey.
# ============================================================================

ventros_crm:
  namespace: "ventros-crm"

  # CI/CD Metadata (Updated by GitHub Actions)
  metadata:
    git_commit: "{{ git_commit | default('unknown') }}"
    build_number: "{{ build_number | default('0') }}"
    deployed_by: "{{ deployed_by | default('awx') }}"
    deployed_at: "{{ ansible_date_time.iso8601 }}"

  # Helm configuration
  helm:
    repo_name: "ventros"
    repo_url: "https://leonardocaloi.github.io/ventros-crm/charts/"
    chart_ref: "ventros/ventros-crm"
    # ⚠️ Override this via AWX Extra Vars (from GitHub Actions)
    chart_version: "{{ chart_version | default('0.1.0') }}"

  # Image configuration
  image:
    repository: "leonardocaloi/ventros-crm"
    # ⚠️ Override this via AWX Extra Vars (from GitHub Actions)
    tag: "{{ image_tag | default('0.1.0') }}"
    pull_policy: "IfNotPresent"

  # Environment-specific overrides
  environment: "{{ deploy_environment | default('production') }}"

  # Replicas (can vary by environment)
  replicas: "{{ replicas | default(1) }}"

  # Autoscaling
  autoscaling:
    enabled: "{{ autoscaling_enabled | default(false) }}"
    min_replicas: 1
    max_replicas: 5
    target_cpu: 70

  # Resources
  resources:
    requests:
      cpu: "{{ resources_requests_cpu | default('10m') }}"
      memory: "{{ resources_requests_memory | default('128Mi') }}"
    limits:
      cpu: "{{ resources_limits_cpu | default('500m') }}"
      memory: "{{ resources_limits_memory | default('512Mi') }}"

  # Ingress configuration
  ingress:
    enabled: true
    class_name: "nginx"
    host: "{{ ingress_host | default('api.crm.ventros.cloud') }}"
    cert_issuer: "letsencrypt-clusterissuer"
    tls:
      enabled: true
      secret_name: "ventros-crm-tls"
    annotations:
      proxy_body_size: "0"
      proxy_buffer_size: "128k"
      proxy_buffers_number: "8"
      proxy_busy_buffers_size: "256k"
      proxy_connect_timeout: "300"
      proxy_send_timeout: "300"
      proxy_read_timeout: "300"
      proxy_request_buffering: "off"
      proxy_buffering: "off"
      limit_rps: "1000"
      limit_connections: "100"

  # PostgreSQL (Zalando operator)
  postgresql:
    enabled: true
    team_id: "ventros"
    instances: 1
    version: "15"
    storage:
      size: "20Gi"
      class: "longhorn"
    resources:
      requests:
        cpu: "10m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "1Gi"

  # RabbitMQ configuration
  rabbitmq:
    enabled: true
    replicas: 1
    storage:
      size: "10Gi"
      class: "longhorn"
    resources:
      requests:
        cpu: "10m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
    ingress:
      enabled: true
      host: "rabbitmq.crm.ventros.cloud"
      cert_issuer: "letsencrypt-clusterissuer"
      tls:
        enabled: true
        secret_name: "rabbitmq-crm-tls"

  # Redis configuration
  redis:
    enabled: true
    architecture: "standalone"
    auth:
      enabled: true
      password: "{{ vault_ventros_crm_redis_password | default('CHANGE_ME') }}"
    storage:
      size: "2Gi"
      class: "longhorn"
    resources:
      requests:
        cpu: "10m"
        memory: "64Mi"
      limits:
        cpu: "200m"
        memory: "256Mi"

  # Temporal configuration
  temporal:
    enabled: true
    server:
      replicas: 1
    frontend:
      replicas: 1
      resources:
        requests:
          cpu: "200m"
          memory: "256Mi"
        limits:
          cpu: "500m"
          memory: "512Mi"
    history:
      replicas: 1
      resources:
        requests:
          cpu: "200m"
          memory: "256Mi"
        limits:
          cpu: "500m"
          memory: "512Mi"
    matching:
      replicas: 1
      resources:
        requests:
          cpu: "200m"
          memory: "256Mi"
        limits:
          cpu: "500m"
          memory: "512Mi"
    worker:
      replicas: 1
      resources:
        requests:
          cpu: "200m"
          memory: "256Mi"
        limits:
          cpu: "500m"
          memory: "512Mi"
    web:
      enabled: true
      replicas: 1
      resources:
        requests:
          cpu: "200m"
          memory: "256Mi"
        limits:
          cpu: "500m"
          memory: "512Mi"
      ingress:
        enabled: true
        host: "temporal.crm.ventros.cloud"
        cert_issuer: "letsencrypt-clusterissuer"
        tls:
          enabled: true
          secret_name: "temporal-crm-tls"

  # Environment variables
  env:
    log_level: "{{ log_level | default('info') }}"
    gin_mode: "{{ gin_mode | default('release') }}"
    environment: "{{ environment | default('production') }}"

  # Secrets (from Ansible Vault)
  secrets:
    jwt_secret: "{{ vault_ventros_crm_jwt_secret | default('CHANGE_ME_IN_PRODUCTION') }}"
    api_key_secret: "{{ vault_ventros_crm_api_key_secret | default('CHANGE_ME_IN_PRODUCTION') }}"

  # Node affinity
  node_affinity:
    enabled: false
```

### AWX Extra Variables (Passed from GitHub Actions)

```yaml
---
# Runtime variables passed from GitHub Actions
image_tag: "0.1.5"                    # Built by GitHub Actions
chart_version: "0.1.5"                 # Published by GitHub Actions
deploy_environment: "staging"          # staging | production
git_commit: "abc123def"                # Git SHA
build_number: "42"                     # GitHub run number
deployed_by: "github-actions"          # Who triggered
```

---

## ⚙️ GitHub Actions Workflows

### Workflow 1: Build & Publish

**File**: `.github/workflows/build-and-publish.yaml`

```yaml
name: Build & Publish

on:
  push:
    branches: [main, develop]
    tags: ['v*']
  pull_request:
    branches: [main]

env:
  GO_VERSION: '1.25.1'
  DOCKER_IMAGE: leonardocaloi/ventros-crm
  HELM_CHART_NAME: ventros-crm

jobs:
  test-unit:
    name: Unit Tests
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run unit tests
        run: make test-unit

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out

  build-image:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: test-unit
    outputs:
      image-tag: ${{ steps.meta.outputs.version }}
      image-full: ${{ steps.meta.outputs.tags }}
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to DockerHub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Docker meta
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.DOCKER_IMAGE }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix={{branch}}-
            type=raw,value=latest,enable={{is_default_branch}}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: .deploy/container/Containerfile
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            VERSION=${{ steps.meta.outputs.version }}
            COMMIT_SHA=${{ github.sha }}
            BUILD_DATE=${{ github.event.head_commit.timestamp }}

  test-integration:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: build-image
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: ventros_crm_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      rabbitmq:
        image: rabbitmq:3.12-management-alpine
        options: >-
          --health-cmd "rabbitmq-diagnostics -q ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run integration tests
        run: make test-integration
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/ventros_crm_test?sslmode=disable
          RABBITMQ_URL: amqp://guest:guest@localhost:5672/
          REDIS_URL: redis://localhost:6379

  publish-helm:
    name: Publish Helm Chart
    runs-on: ubuntu-latest
    needs: [build-image, test-integration]
    if: github.event_name != 'pull_request'
    outputs:
      chart-version: ${{ steps.version.outputs.version }}
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install Helm
        uses: azure/setup-helm@v4

      - name: Determine version
        id: version
        run: |
          if [[ $GITHUB_REF == refs/tags/v* ]]; then
            VERSION=${GITHUB_REF#refs/tags/v}
          else
            VERSION="0.0.0-${GITHUB_SHA::7}"
          fi
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          echo "VERSION=$VERSION" >> $GITHUB_ENV

      - name: Update Chart.yaml
        run: |
          sed -i "s/^version:.*/version: ${{ steps.version.outputs.version }}/" .deploy/helm/ventros-crm/Chart.yaml
          sed -i "s/^appVersion:.*/appVersion: ${{ steps.version.outputs.version }}/" .deploy/helm/ventros-crm/Chart.yaml

      - name: Package Helm chart
        run: |
          helm package .deploy/helm/ventros-crm --destination .deploy/helm/packages/

      - name: Checkout gh-pages
        uses: actions/checkout@v4
        with:
          ref: gh-pages
          path: gh-pages

      - name: Update Helm repository
        run: |
          cp .deploy/helm/packages/*.tgz gh-pages/charts/
          helm repo index gh-pages/charts --url https://leonardocaloi.github.io/ventros-crm/charts/

      - name: Publish to GitHub Pages
        run: |
          cd gh-pages
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add charts/
          git commit -m "Publish Helm chart ${{ steps.version.outputs.version }}"
          git push

  trigger-deployment:
    name: Trigger AWX Deployment
    runs-on: ubuntu-latest
    needs: [build-image, publish-helm]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    steps:
      - name: Trigger AWX Staging Deployment
        run: |
          curl -X POST "${{ secrets.AWX_URL }}/api/v2/job_templates/${{ secrets.AWX_JOB_TEMPLATE_ID }}/launch/" \
            -H "Authorization: Bearer ${{ secrets.AWX_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{
              "extra_vars": {
                "image_tag": "${{ needs.build-image.outputs.image-tag }}",
                "chart_version": "${{ needs.publish-helm.outputs.chart-version }}",
                "deploy_environment": "staging",
                "git_commit": "${{ github.sha }}",
                "build_number": "${{ github.run_number }}",
                "deployed_by": "github-actions"
              }
            }'

      - name: Wait for AWX job
        run: |
          # Poll AWX job status (implementation depends on AWX API)
          echo "Waiting for AWX job to complete..."
          # TODO: Add polling logic

      - name: Notify Slack
        if: always()
        uses: slackapi/slack-github-action@v1
        with:
          webhook-url: ${{ secrets.SLACK_WEBHOOK_URL }}
          payload: |
            {
              "text": "Deployment to Staging: ${{ job.status }}",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Ventros CRM Deployment*\n*Environment:* Staging\n*Status:* ${{ job.status }}\n*Version:* ${{ needs.build-image.outputs.image-tag }}\n*Commit:* ${{ github.sha }}"
                  }
                }
              ]
            }
```

### Workflow 2: Deploy to Production

**File**: `.github/workflows/deploy-production.yaml`

```yaml
name: Deploy to Production

on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to deploy (e.g., 0.1.0)'
        required: true
      replicas:
        description: 'Number of replicas'
        required: false
        default: '3'

jobs:
  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://api.crm.ventros.cloud
    steps:
      - name: Validate version format
        run: |
          if [[ ! "${{ github.event.inputs.version }}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo "Invalid version format. Must be semver (e.g., 0.1.0)"
            exit 1
          fi

      - name: Trigger AWX Production Deployment
        id: deploy
        run: |
          RESPONSE=$(curl -s -X POST "${{ secrets.AWX_URL }}/api/v2/job_templates/${{ secrets.AWX_JOB_TEMPLATE_PROD_ID }}/launch/" \
            -H "Authorization: Bearer ${{ secrets.AWX_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{
              "extra_vars": {
                "image_tag": "${{ github.event.inputs.version }}",
                "chart_version": "${{ github.event.inputs.version }}",
                "deploy_environment": "production",
                "replicas": ${{ github.event.inputs.replicas }},
                "deployed_by": "${{ github.actor }}",
                "deployment_type": "manual"
              }
            }')

          JOB_ID=$(echo $RESPONSE | jq -r .id)
          echo "job-id=$JOB_ID" >> $GITHUB_OUTPUT

      - name: Wait for deployment
        run: |
          # Poll AWX job until completion
          echo "Monitoring AWX job ${{ steps.deploy.outputs.job-id }}..."
          # TODO: Add polling with timeout

      - name: Run smoke tests
        run: |
          # Basic health checks
          curl -f https://api.crm.ventros.cloud/health || exit 1
          echo "Smoke tests passed!"

      - name: Notify team
        if: always()
        uses: slackapi/slack-github-action@v1
        with:
          webhook-url: ${{ secrets.SLACK_WEBHOOK_URL }}
          payload: |
            {
              "text": "🚀 Production Deployment: ${{ job.status }}",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Ventros CRM - Production Deployment*\n*Version:* ${{ github.event.inputs.version }}\n*Replicas:* ${{ github.event.inputs.replicas }}\n*Deployed by:* ${{ github.actor }}\n*Status:* ${{ job.status }}"
                  }
                }
              ]
            }
```

---

## 🔗 AWX Integration

### AWX Project Configuration

**Project Name**: `ventros-crm-deploy`

**SCM Type**: Git

**SCM URL**: `https://github.com/ventros/crm.git`

**SCM Branch**: `main`

**SCM Update Options**:
- ✅ Clean
- ✅ Delete on Update
- ✅ Update Revision on Launch

**Playbook Path**: `.deploy/ansible/deploy.yml`

### AWX Job Template: Deploy Ventros CRM (Staging)

```yaml
name: Deploy Ventros CRM - Staging
job_type: run
inventory: Kubernetes Staging
project: ventros-crm-deploy
playbook: .deploy/ansible/deploy.yml
credentials:
  - Kubernetes Staging Credentials
  - Ansible Vault Credentials
extra_vars:
  deploy_environment: staging
  ingress_host: api.staging.crm.ventros.cloud
survey_enabled: true
survey_spec:
  name: Deployment Options
  description: Runtime deployment configuration
  spec:
    - question_name: Docker Image Tag
      question_description: Tag of Docker image to deploy
      required: true
      type: text
      variable: image_tag
      default: latest
    - question_name: Helm Chart Version
      question_description: Version of Helm chart to use
      required: true
      type: text
      variable: chart_version
      default: 0.1.0
    - question_name: Number of Replicas
      question_description: Application replicas
      required: false
      type: integer
      variable: replicas
      default: 1
      min: 1
      max: 10
```

### AWX Job Template: Deploy Ventros CRM (Production)

```yaml
name: Deploy Ventros CRM - Production
job_type: run
inventory: Kubernetes Production
project: ventros-crm-deploy
playbook: .deploy/ansible/deploy.yml
credentials:
  - Kubernetes Production Credentials
  - Ansible Vault Credentials
extra_vars:
  deploy_environment: production
  ingress_host: api.crm.ventros.cloud
  replicas: 3
survey_enabled: true
survey_spec:
  name: Production Deployment
  description: Production deployment options
  spec:
    - question_name: Docker Image Tag
      question_description: Stable version tag (no 'latest')
      required: true
      type: text
      variable: image_tag
      default: 0.1.0
    - question_name: Helm Chart Version
      question_description: Version of Helm chart to use
      required: true
      type: text
      variable: chart_version
      default: 0.1.0
```

### Main Playbook

**File**: `.deploy/ansible/deploy.yml`

```yaml
---
- name: Deploy Ventros CRM
  hosts: localhost
  connection: local
  gather_facts: true
  vars_files:
    - global_vars.yml
  roles:
    - ventros_crm
  tasks:
    - name: Display deployment summary
      debug:
        msg:
          - "=== Deployment Summary ==="
          - "Environment: {{ deploy_environment | default('unknown') }}"
          - "Namespace: {{ ventros_crm.namespace }}"
          - "Docker Image: {{ ventros_crm.image.repository }}:{{ image_tag | default(ventros_crm.image.tag) }}"
          - "Helm Chart: {{ ventros_crm.helm.chart_ref }} v{{ chart_version | default(ventros_crm.helm.chart_version) }}"
          - "Replicas: {{ replicas | default(ventros_crm.replicas) }}"
          - "Ingress: {{ ventros_crm.ingress.host }}"
          - "Git Commit: {{ git_commit | default('unknown') }}"
          - "Deployed By: {{ deployed_by | default('unknown') }}"

    - name: Wait for deployment to be ready
      kubernetes.core.k8s_info:
        kind: Deployment
        namespace: "{{ ventros_crm.namespace }}"
        name: ventros-crm
      register: deployment
      until: deployment.resources[0].status.readyReplicas | default(0) >= (replicas | default(ventros_crm.replicas) | int)
      retries: 30
      delay: 10

    - name: Verify health endpoint
      uri:
        url: "https://{{ ventros_crm.ingress.host }}/health"
        status_code: 200
        validate_certs: true
      retries: 10
      delay: 5
```

---

## 🔐 Security & Secrets

### GitHub Secrets (Actions)

```bash
# DockerHub
DOCKER_USERNAME=leonardocaloi
DOCKER_PASSWORD=<token>

# AWX
AWX_URL=https://awx.ventros.cloud
AWX_TOKEN=<api-token>
AWX_JOB_TEMPLATE_ID=123          # Staging
AWX_JOB_TEMPLATE_PROD_ID=456     # Production

# Notifications
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
```

### AWX Credentials

**Kubernetes Staging**:
- Type: Kubernetes
- Kubeconfig: (staging cluster)

**Kubernetes Production**:
- Type: Kubernetes
- Kubeconfig: (production cluster)

**Ansible Vault**:
- Type: Vault
- Vault Password: (encrypted secrets)

### Ansible Vault Secrets

**File**: `.deploy/ansible/vault.yml` (encrypted)

```yaml
---
# Application secrets
vault_ventros_crm_jwt_secret: "super-secret-jwt-key-production"
vault_ventros_crm_api_key_secret: "super-secret-api-key"

# Redis
vault_ventros_crm_redis_password: "redis-strong-password"

# PostgreSQL
vault_ventros_crm_postgres_password: "postgres-strong-password"

# RabbitMQ
vault_ventros_crm_rabbitmq_password: "rabbitmq-strong-password"
```

**Encrypt vault**:
```bash
ansible-vault encrypt .deploy/ansible/vault.yml
```

**Update AWX with vault password**:
AWX → Credentials → "Ansible Vault Credentials" → Vault Password

---

## 🌍 Deployment Environments

### Staging

**Purpose**: Pre-production testing

**Infrastructure**:
- Namespace: `ventros-crm-staging`
- Ingress: `api.staging.crm.ventros.cloud`
- Replicas: 1
- Resources: Minimal (CPU: 10m, Memory: 128Mi)

**Database**: Shared PostgreSQL cluster (separate database)

**Deployment Trigger**: Automatic on `push` to `main`

**Rollback**: Automatic on health check failure

### Production

**Purpose**: Live customer traffic

**Infrastructure**:
- Namespace: `ventros-crm`
- Ingress: `api.crm.ventros.cloud`
- Replicas: 3 (HA)
- Resources: Production-grade (CPU: 500m, Memory: 512Mi)
- Autoscaling: Enabled (min: 3, max: 10)

**Database**: Dedicated PostgreSQL cluster with automated backups

**Deployment Trigger**: Manual approval or Git tag

**Rollback**: Manual via AWX with previous version

---

## ↩️ Rollback Strategy

### Automatic Rollback (Staging)

GitHub Actions monitors deployment health:

```yaml
- name: Verify deployment
  run: |
    for i in {1..10}; do
      if curl -f https://api.staging.crm.ventros.cloud/health; then
        echo "Health check passed"
        exit 0
      fi
      sleep 10
    done
    echo "Health check failed, triggering rollback"
    exit 1

- name: Rollback on failure
  if: failure()
  run: |
    # Trigger AWX rollback job
    curl -X POST "${{ secrets.AWX_URL }}/api/v2/job_templates/${{ secrets.AWX_ROLLBACK_TEMPLATE_ID }}/launch/" \
      -H "Authorization: Bearer ${{ secrets.AWX_TOKEN }}"
```

### Manual Rollback (Production)

**Via AWX**:
1. Go to AWX → Job Templates → "Rollback Ventros CRM"
2. Select previous stable version from dropdown
3. Launch job

**Via Helm**:
```bash
# List releases
helm list -n ventros-crm

# Rollback to previous
helm rollback ventros-crm -n ventros-crm

# Rollback to specific revision
helm rollback ventros-crm 3 -n ventros-crm
```

### Rollback Playbook

**File**: `.deploy/ansible/rollback.yml`

```yaml
---
- name: Rollback Ventros CRM
  hosts: localhost
  connection: local
  vars_files:
    - global_vars.yml
  tasks:
    - name: Get current Helm release
      kubernetes.core.helm_info:
        name: ventros-crm
        namespace: "{{ ventros_crm.namespace }}"
      register: current_release

    - name: Display current version
      debug:
        msg: "Current version: {{ current_release.status.version }}"

    - name: Confirm rollback
      pause:
        prompt: "Rollback to previous version? (yes/no)"
      register: confirm

    - name: Rollback Helm release
      kubernetes.core.helm:
        name: ventros-crm
        namespace: "{{ ventros_crm.namespace }}"
        state: rollback
      when: confirm.user_input | bool

    - name: Verify rollback
      kubernetes.core.k8s_info:
        kind: Deployment
        namespace: "{{ ventros_crm.namespace }}"
        name: ventros-crm
      register: deployment
      until: deployment.resources[0].status.readyReplicas >= 1
      retries: 20
      delay: 10
```

---

## 📊 Monitoring & Observability

### Deployment Metrics

**GitHub Actions**:
- Build time
- Test duration
- Image size
- Deployment success rate

**AWX**:
- Deployment duration
- Success/failure rate
- Resource utilization during deploy

**Kubernetes**:
- Pod restart count
- Deployment rollout status
- Resource consumption

### Alerting

**Slack Notifications**:
- ✅ Build success/failure
- ✅ Deployment started
- ✅ Deployment completed
- ❌ Health check failures
- ⚠️ Rollback triggered

**Example Slack Message**:
```
🚀 Deployment to Production
━━━━━━━━━━━━━━━━━━━━
✅ Status: Success
📦 Version: 0.1.5
🏷️ Commit: abc123d
👤 Deployed by: @leonardo
⏱️ Duration: 3m 42s
🔗 https://api.crm.ventros.cloud
```

### Health Checks

**Kubernetes Probes**:
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

**Smoke Tests** (post-deployment):
```bash
# Health endpoint
curl -f https://api.crm.ventros.cloud/health

# Authentication
curl -f https://api.crm.ventros.cloud/api/v1/auth/health

# Database connectivity
curl -f https://api.crm.ventros.cloud/health/db

# RabbitMQ connectivity
curl -f https://api.crm.ventros.cloud/health/rabbitmq
```

---

## 📋 Implementation Checklist

### Phase 1: Setup GitHub Actions
- [ ] Create `.github/workflows/build-and-publish.yaml`
- [ ] Create `.github/workflows/deploy-production.yaml`
- [ ] Add GitHub Secrets (DOCKER_*, AWX_*, SLACK_*)
- [ ] Create `gh-pages` branch for Helm repository
- [ ] Test build workflow on feature branch

### Phase 2: Configure AWX
- [ ] Create AWX Project pointing to git repo
- [ ] Create Job Template "Deploy Ventros CRM - Staging"
- [ ] Create Job Template "Deploy Ventros CRM - Production"
- [ ] Create Job Template "Rollback Ventros CRM"
- [ ] Add Kubernetes credentials (staging + production)
- [ ] Add Ansible Vault credentials
- [ ] Test manual deployment from AWX UI

### Phase 3: Update Ansible Role
- [ ] Enhance `.deploy/ansible/global_vars.yml` with metadata fields
- [ ] Create `.deploy/ansible/deploy.yml` playbook
- [ ] Create `.deploy/ansible/rollback.yml` playbook
- [ ] Test variable substitution with AWX extra vars
- [ ] Document variable override mechanism

### Phase 4: Testing
- [ ] Test full workflow: git push → build → staging deploy
- [ ] Test production deployment with manual approval
- [ ] Test rollback scenario (staging)
- [ ] Test rollback scenario (production)
- [ ] Verify health checks and smoke tests
- [ ] Test Slack notifications

### Phase 5: Documentation
- [ ] Update README.md with CI/CD section
- [ ] Document rollback procedures
- [ ] Document emergency procedures
- [ ] Create runbook for common issues
- [ ] Train team on new workflow

### Phase 6: Monitoring
- [ ] Setup Slack channel for deployment notifications
- [ ] Configure Grafana dashboards (deployment metrics)
- [ ] Setup PagerDuty for critical failures
- [ ] Document on-call procedures

---

## 🎯 Success Criteria

### Build Phase
- ✅ All tests pass (unit + integration)
- ✅ Docker image builds successfully
- ✅ Image size < 100MB
- ✅ Build time < 5 minutes
- ✅ Helm chart published

### Deploy Phase
- ✅ Deployment completes in < 5 minutes
- ✅ Zero downtime (rolling update)
- ✅ Health checks pass
- ✅ No pod restarts after deployment
- ✅ Application responds to requests

### Quality Gates
- ✅ Test coverage > 80%
- ✅ No high/critical security vulnerabilities
- ✅ All smoke tests pass
- ✅ Logs show no errors in first 5 minutes

---

## 📞 Support & Troubleshooting

### Common Issues

**Build fails with "tests timeout"**:
```bash
# Increase timeout in workflow
timeout: 10m
```

**AWX job fails with "connection refused"**:
```bash
# Check Kubernetes credentials in AWX
# Verify kubeconfig context
```

**Helm deployment fails with "version mismatch"**:
```bash
# Ensure chart_version matches image_tag
# Check Helm repository index.yaml
```

**Health checks fail after deployment**:
```bash
# Check pod logs
kubectl logs -n ventros-crm -l app=ventros-crm

# Check events
kubectl get events -n ventros-crm --sort-by='.lastTimestamp'

# Trigger rollback if critical
```

### Emergency Contacts

- **Build Issues**: GitHub Actions logs
- **Deployment Issues**: AWX logs + K8s events
- **Application Issues**: Application logs + Grafana

---

## 📝 Next Steps

1. **Immediate** (Week 1):
   - Implement GitHub Actions workflows
   - Configure AWX job templates
   - Test on staging environment

2. **Short-term** (Week 2-3):
   - Add comprehensive smoke tests
   - Setup monitoring dashboards
   - Document procedures

3. **Long-term** (Month 2+):
   - Implement blue-green deployments
   - Add canary deployments
   - Automate database migrations
   - Add performance testing in pipeline

---

**Version**: 1.0
**Last Updated**: 2025-10-12
**Status**: ✅ Ready for Implementation
**Reviewed By**: DevOps Team
**Approved By**: Tech Lead
