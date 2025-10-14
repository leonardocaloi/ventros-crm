# Estratégia CI/CD - Ventros CRM

## Arquitetura de Deployment

### Camadas de Responsabilidade

#### 🔧 Makefile (Camada Aplicação - Baixo Nível)
```
Desenvolvedores/CI → Código Go
├─ make test           # testes unitários
├─ make build          # compilar binário
├─ make docker-build   # criar imagem
├─ make docker-push    # publicar imagem no registry
└─ make helm-package   # empacotar chart (helm push para ChartMuseum/Harbor)
```

#### 🎯 AWX (Camada Infra/Deployment - Alto Nível)
```
AWX Playbooks/Job Templates
├─ Provisionar nodes (RKE2)
├─ Instalar operadores (Zalando Postgres)
├─ Deploy aplicação (helm install/upgrade usando chart publicado)
├─ Gerenciar backups (Zalando backup/restore)
├─ Configurar networking/storage
└─ Disaster recovery
```

## Fluxo de CI/CD

### Pipeline Completo:
```
1. [GitHub/GitLab CI]
   └─ make test && make docker-push && make helm-push

2. [AWX Job Template Trigger via webhook/API]
   └─ Ansible playbook que faz:
      - helm repo update
      - helm upgrade --install app registry/chart:versão
      - wait readiness
      - smoke tests
```

### AWX Gerencia:
- **Deployment**: playbook chama `helm upgrade` no cluster
- **Backups Zalando**: CronJobs ou playbooks agendados
  - Zalando já tem backup automático via pgBackRest/WAL-G
  - AWX pode trigger restore quando necessário
- **Promote entre ambientes**:
  - Job Template "Deploy Staging"
  - Job Template "Deploy Production" (com approval)

## Vantagens dessa Arquitetura

✅ **Makefile**: portável, roda local e CI, não conhece infra
✅ **AWX**: single source of truth para estado da infra
✅ **Separação**: devs não precisam credenciais do cluster
✅ **Auditoria**: AWX loga tudo (quem deployou, quando, rollback)
✅ **Zalando Operator**: backups/restore declarativos via CRDs

## Integração AWX ↔ CI

### Opção 1: AWX via API (Recomendado)
No final do CI:
```bash
# make helm-push já publicou chart

# Triggerar AWX via API
curl -X POST https://awx.seudominio.com/api/v2/job_templates/ID/launch/ \
  -H "Authorization: Bearer $AWX_TOKEN" \
  -d '{"extra_vars": "{\"chart_version\": \"1.2.3\", \"environment\": \"staging\"}"}'
```

### Opção 2: AWX Poll (Simplificado)
AWX Job agendado que:
- Checa registry por novas versões do chart
- Deploy automático em staging se houver nova versão
- Notificação Slack/email para aprovar produção

## AWX Playbooks Sugeridos

### 1. deploy-app.yml
```yaml
- name: Deploy Ventros CRM
  hosts: localhost
  vars:
    environment: "{{ env }}"  # staging/prod
    chart_version: "{{ version }}"
  tasks:
    - name: Add helm repo
      kubernetes.core.helm_repository:
        name: ventros
        repo_url: https://charts.seudominio.com

    - name: Deploy/upgrade chart
      kubernetes.core.helm:
        name: ventros-crm
        chart_ref: ventros/ventros-crm
        chart_version: "{{ chart_version }}"
        release_namespace: "{{ environment }}"
        values_files:
          - "values-{{ environment }}.yaml"
        wait: true
```

### 2. backup-restore.yml
```yaml
- name: Restore Postgres backup em staging
  hosts: localhost
  tasks:
    - name: Criar clone do Postgres de produção
      kubernetes.core.k8s:
        state: present
        definition:
          apiVersion: acid.zalan.do/v1
          kind: postgresql
          metadata:
            name: ventros-staging
          spec:
            clone:
              cluster: ventros-production
              timestamp: "{{ backup_timestamp | default('latest') }}"
            # ... resto da spec

    - name: Sanitizar dados sensíveis
      kubernetes.core.k8s_exec:
        namespace: staging
        pod: ventros-staging-0
        command: psql -f /scripts/sanitize.sql
```

### 3. manage-environments.yml
```yaml
- name: Gerenciar ambientes
  hosts: localhost
  tasks:
    - name: Criar namespaces
      kubernetes.core.k8s:
        state: present
        definition:
          apiVersion: v1
          kind: Namespace
          metadata:
            name: "{{ item }}"
      loop:
        - dev
        - staging
        - production

    - name: Configurar ResourceQuotas por ambiente
      # ...
```

## Gestão de Databases com Zalando

### Backup Automático
O Zalando Postgres Operator já resolve backups:

```yaml
# No seu postgresql CRD
spec:
  enableLogicalBackup: true
  enableShmVolume: true

  # Backup contínuo WAL
  backup:
    schedule: "0 2 * * *"  # diário 2am
    retentionPolicy: "30d"

  clone:  # para staging
    cluster: "ventros-prod"
    timestamp: "2025-10-11 03:00:00+00"  # point-in-time
```

**AWX apenas**:
- Trigger restore quando precisar refresh staging
- Monitoring de jobs de backup (alertar se falhar)
- Disaster recovery (restore completo em novo cluster)

### Clone de Database para Staging

```yaml
apiVersion: "acid.zalan.do/v1"
kind: postgresql
metadata:
  name: crm-staging
spec:
  clone:
    cluster: "crm-production"  # fonte
    timestamp: "2025-01-10T15:00:00Z"  # PITR opcional
  # resto da config...
```

### CronJob para Refresh Staging

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: refresh-staging-db
spec:
  schedule: "0 2 * * *"  # 2h da manhã
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: refresh
            image: bitnami/kubectl
            command:
            - /bin/sh
            - -c
            - |
              # Delete staging cluster
              kubectl delete postgresql crm-staging -n staging
              # Recria do clone de prod
              kubectl apply -f /configs/staging-clone.yaml
              # Aguarda ficar ready
              kubectl wait --for=condition=Ready postgresql/crm-staging
              # Roda sanitização
              kubectl exec crm-staging-0 -- psql -f /scripts/sanitize.sql
```

### Script de Sanitização (ConfigMap)

```sql
-- sanitize.sql
UPDATE contacts SET
  phone = CONCAT('5511', LPAD((random()*100000000)::int::text, 9, '0')),
  email = CONCAT('user', id, '@example.com');

UPDATE messages SET
  content = 'REDACTED'
WHERE content LIKE '%senha%' OR content LIKE '%cartão%';

-- Outros campos sensíveis...
```

## Workflow Completo

### No Makefile:

```makefile
.PHONY: deploy-staging-full
deploy-staging-full:
	@echo "🔄 Refreshing staging database from production..."
	kubectl apply -f k8s/staging-db-refresh-job.yaml
	kubectl wait --for=condition=complete job/staging-db-refresh --timeout=30m
	@echo "🚀 Deploying application to staging..."
	make helm-upgrade-staging
	@echo "🧪 Running integration tests..."
	make integration-test

.PHONY: deploy-staging-quick
deploy-staging-quick:
	@echo "🚀 Quick deploy (sem refresh de DB)..."
	make helm-upgrade-staging
```

### No CI/CD (GitHub Actions exemplo):

```yaml
# Merge na main
- name: Deploy to Staging (Full)
  run: make deploy-staging-full

- name: Wait for smoke tests
  run: make smoke-test-staging

- name: Manual approval for production
  uses: trstringer/manual-approval@v1

- name: Deploy to Production
  run: make deploy-prod
```

## Ambientes Recomendados

### Mínimo viável:
1. **Development (local)**: kind/k3d local ou namespace dev no cluster
2. **Staging/Homologação**: ambiente que imita produção
3. **Production**: real

### Homologação é crucial porque:
- Testa migrations reais antes de produção
- Valida integrações (Temporal, RabbitMQ, Keycloak)
- Permite testar rollbacks
- Dados reais (sanitizados) para performance testing

## Vantagens do Clone de Produção

✅ **Dados reais**: testa com volume e complexidade de produção
✅ **Migrations seguras**: valida antes de prod
✅ **Performance testing**: queries com dados reais
✅ **LGPD compliance**: dados sanitizados
✅ **Automático**: CronJob mantém staging atualizado
✅ **Fast feedback**: clone do Zalando é rápido (usa snapshots)

## Estrutura Recomendada

```
ventros-crm/              # repo Go
├─ Makefile              # build/test/publish
├─ helm/
│  └─ ventros-crm/      # chart da aplicação
├─ .github/workflows/   # ou .gitlab-ci.yml
│  └─ ci.yml            # roda make + trigger AWX
└─ README.md

infrastructure/          # repo Ansible (separado ou monorepo)
├─ inventories/
│  ├─ dev/
│  ├─ staging/
│  └─ production/
├─ playbooks/
│  ├─ deploy-app.yml
│  ├─ backup-restore.yml
│  ├─ provision-cluster.yml
│  └─ manage-operators.yml
└─ awx_configs/         # exportar configs AWX como código
```

## Extras

### Para RabbitMQ:
Zalando não gerencia, mas você pode:
```bash
# Backup definitions
rabbitmqctl export_definitions staging-backup.json

# Restore
rabbitmqctl import_definitions staging-backup.json
```

### Para Temporal:
- Temporal tem seu próprio persistence (Postgres gerenciado pelo Zalando também)
- Clone do Postgres já inclui dados do Temporal
- **Cuidado**: workflows em execução não migram, só history

### Monitoring:
Zalando exporta métricas Prometheus out-of-the-box:
- `pg_up`
- `pg_database_size_bytes`
- `pg_replication_lag`

## Resumo da Estratégia

| Camada | Ferramenta | Responsabilidade |
|--------|------------|------------------|
| **Código** | Make | Build, test, publish artefatos |
| **CI** | GitHub Actions/GitLab | Rodar make + trigger AWX |
| **Infra** | AWX/Ansible | Deploy, backup, infra, operators |
| **Cluster** | Kubernetes/Helm | Rodar aplicação |
| **Dados** | Zalando Operator | Postgres HA + backups |

## Makefile - Targets Sugeridos

```makefile
# Build & Test
make test              # go test
make lint              # golangci-lint
make build             # go build
make docker-build      # build imagem
make docker-push       # push registry

# Helm
make helm-package      # helm package
make helm-push         # push helm chart

# Deploy
make deploy-dev        # helm install/upgrade no dev
make deploy-staging    # helm install/upgrade staging
make deploy-prod       # helm install/upgrade prod

# Database
make db-backup         # backup produção
make db-restore-staging   # restore sanitizado em staging

# Testing
make integration-test     # testes contra staging
make smoke-test          # smoke tests

# Rollback
make rollback-staging    # helm rollback staging
make rollback-prod       # helm rollback prod
```

## Conclusão

Makefiles são **totalmente válidos e amplamente usados pela indústria** (Google, Docker, Kubernetes, HashiCorp). Com AWX orquestrando a infra e Zalando gerenciando databases, você tem uma arquitetura profissional e escalável.

A separação é perfeita:
- **Makefile embaixo**: portável, não conhece infra
- **AWX em cima**: orquestra tudo, auditoria, RBAC, visibilidade

Você está no caminho certo! 🎯
