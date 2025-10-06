# Ventros CRM Helm Chart

Este Helm Chart instala o Ventros CRM com todas as suas dependências no Kubernetes.

## 📋 Pré-requisitos

- Kubernetes 1.23+
- Helm 3.8+
- PV provisioner support (para persistência)
- Cert-Manager (opcional, para TLS automático)


## 🚀 Instalação Rápida

### 🔒 Fase 1: Apenas PostgreSQL (Recomendado para início)

```bash
# Instalar apenas PostgreSQL Operator + Cluster
# (Aplicação desabilitada com replicaCount: 0)
helm install ventros-crm ./ventros-crm \
  --namespace ventros-crm \
  --create-namespace \
  -f values-dev.yaml
```

**O que é instalado na Fase 1:**
- ✅ PostgreSQL Operator (Zalando)
- ✅ PostgreSQL Cluster `ventros-db`
- ❌ Aplicação CRM (desabilitada)
- ❌ Redis/RabbitMQ (desabilitados)

### Desenvolvimento Completo (Minikube/Kind)

```bash
# Para habilitar a aplicação completa, edite values-dev.yaml:
# replicaCount: 1
# redis.enabled: true
# rabbitmq.enabled: true

helm upgrade ventros-crm ./ventros-crm \
  --namespace ventros-crm \
  -f values-dev.yaml
```

### Produção

```bash
# Instalar com valores de produção
helm install ventros-crm ./ventros-crm \
  --namespace ventros-crm \
  --create-namespace \
  -f values-production.yaml
```

## 📦 Componentes

O chart instala os seguintes componentes:

### Aplicação Principal
- **Ventros CRM API**: Aplicação Go com Gin framework
- **Migrations Job**: Job para executar migrações do banco

### Dependências (Opcionais)

#### PostgreSQL
- **Opção 1**: Zalando Postgres Operator (Recomendado para produção)
- **Opção 2**: PostgreSQL externo (configure em `externalPostgresql`)

#### Redis
- **Opção 1**: Redis via Bitnami Chart (interno)
- **Opção 2**: Redis externo (configure em `externalRedis`)

#### RabbitMQ
- **Opção 1**: RabbitMQ Cluster Operator (Recomendado para produção)
- **Opção 2**: RabbitMQ via Bitnami Chart (desenvolvimento)
- **Opção 3**: RabbitMQ externo (configure em `externalRabbitmq`)

#### Temporal
- **Opção 1**: Temporal via subchart (interno)
- **Opção 2**: Temporal externo (configure em `temporal.external`)

## ⚙️ Configuração

### Variáveis de Ambiente

Todas as variáveis de ambiente não-sensíveis são configuradas via **ConfigMap**:

```yaml
configMap:
  enabled: true
  data:
    app.yaml: |
      # Configurações customizadas aqui
```

O ConfigMap é automaticamente carregado via `envFrom` no deployment.

### Secrets

Secrets são configurados separadamente:

```yaml
secrets:
  # WAHA API Key
  wahaApiKey: "sua-api-key"
  
  # Admin credentials
  adminEmail: "admin@ventros.com"
  adminPassword: "senha-segura"
  adminName: "Administrator"
  
  # Ou use um secret existente
  existingSecret: "meu-secret"
```

### PostgreSQL com Zalando Operator

```yaml
postgresOperator:
  enabled: true
  createCluster: true
  cluster:
    teamId: "ventros"
    numberOfInstances: 2
    volumeSize: 10Gi
    version: "16"
    database: "ventros_crm"
    username: "ventros"
```

**Importante**: O secret do PostgreSQL é criado automaticamente pelo operador com o nome:
```
<username>.<cluster-name>.credentials.postgresql.acid.zalan.do
```

### RabbitMQ

#### Com Cluster Operator (Produção)

```yaml
rabbitmqOperator:
  enabled: true
  createCluster: true
  cluster:
    replicas: 3
    volumeSize: 10Gi
```

#### Com Bitnami Chart (Desenvolvimento)

```yaml
rabbitmq:
  enabled: true
  replicaCount: 3
  persistence:
    enabled: true
    size: 10Gi
```

### RBAC

O chart cria automaticamente Role e RoleBinding para o ServiceAccount:

```yaml
rbac:
  create: true

serviceAccount:
  create: true
  name: ""  # Usa o nome padrão do chart
```

## 🔒 Segurança

### Pod Security

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
  seccompProfile:
    type: RuntimeDefault

securityContext:
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

### Network Policy

```yaml
networkPolicy:
  enabled: true
  policyTypes:
    - Ingress
    - Egress
```

## 📊 Monitoramento

### Prometheus

```yaml
serviceMonitor:
  enabled: true
  interval: 30s
  scrapeTimeout: 10s
```

### Health Checks

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

## 🔄 Autoscaling

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

## 🌐 Ingress

```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: api.ventros.cloud
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: ventros-crm-tls
      hosts:
        - api.ventros.cloud
```

## 🔧 Comandos Úteis

### Verificar Status

```bash
# Pods
kubectl get pods -n ventros-crm

# Services
kubectl get svc -n ventros-crm

# PostgreSQL Cluster
kubectl get postgresql -n ventros-crm

# RabbitMQ Cluster
kubectl get rabbitmqcluster -n ventros-crm
```

### Logs

```bash
# Logs da aplicação
kubectl logs -n ventros-crm -l app.kubernetes.io/name=ventros-crm -f

# Logs do migration job
kubectl logs -n ventros-crm -l job-name=ventros-crm-migration
```

### Port Forward

```bash
# API
kubectl port-forward -n ventros-crm svc/ventros-crm 8080:8080

# PostgreSQL
kubectl port-forward -n ventros-crm svc/ventros-postgres 5432:5432

# RabbitMQ Management
kubectl port-forward -n ventros-crm svc/ventros-rabbitmq 15672:15672
```

### Upgrade

```bash
# Upgrade com novos valores
helm upgrade ventros-crm ./ventros-crm \
  -n ventros-crm \
  -f values-production.yaml

# Rollback
helm rollback ventros-crm -n ventros-crm
```

### Desinstalar

```bash
# Remover o release
helm uninstall ventros-crm -n ventros-crm

# Remover o namespace (cuidado!)
kubectl delete namespace ventros-crm
```

## 🐛 Troubleshooting

### Init Containers Falhando

Os init containers aguardam as dependências estarem prontas. Se falharem:

```bash
# Verificar logs do init container
kubectl logs -n ventros-crm <pod-name> -c wait-for-postgres
kubectl logs -n ventros-crm <pod-name> -c wait-for-redis
kubectl logs -n ventros-crm <pod-name> -c wait-for-rabbitmq
```

### Migrations Falhando

```bash
# Ver logs do migration job
kubectl logs -n ventros-crm -l job-name=ventros-crm-migration

# Deletar job para reexecutar
kubectl delete job -n ventros-crm ventros-crm-migration
helm upgrade ventros-crm ./ventros-crm -n ventros-crm
```

### Secrets Não Encontrados

Verifique se os secrets foram criados corretamente:

```bash
# Listar secrets
kubectl get secrets -n ventros-crm

# PostgreSQL (Zalando Operator)
kubectl get secret -n ventros-crm ventros.ventros-postgres.credentials.postgresql.acid.zalan.do

# RabbitMQ
kubectl get secret -n ventros-crm ventros-rabbitmq

# Application secrets
kubectl get secret -n ventros-crm ventros-crm-secrets
```

## 📝 Valores Importantes

### Variáveis de Ambiente Obrigatórias

O código Go espera as seguintes variáveis (configuradas via ConfigMap):

- `PORT`: Porta do servidor (padrão: 8080)
- `ENV`: Ambiente (development/production)
- `LOG_LEVEL`: Nível de log (debug/info/warn/error)
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_SSLMODE`: PostgreSQL
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_DB`: Redis
- `RABBITMQ_URL`: URL AMQP completa (formato: `amqp://user:pass@host:port/`)
- `TEMPORAL_HOST`: Host:Port do Temporal
- `TEMPORAL_NAMESPACE`: Namespace do Temporal
- `WAHA_BASE_URL`: URL base do WAHA

### Secrets Obrigatórios

- `DB_PASSWORD`: Senha do PostgreSQL
- `REDIS_PASSWORD`: Senha do Redis (se auth habilitado)
- `RABBITMQ_PASSWORD`: Senha do RabbitMQ (injetada na RABBITMQ_URL)
- `WAHA_API_KEY`: API Key do WAHA
- `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_NAME`: Credenciais do admin

## 🔗 Links Úteis

- [Zalando Postgres Operator](https://github.com/zalando/postgres-operator)
- [RabbitMQ Cluster Operator](https://www.rabbitmq.com/kubernetes/operator/operator-overview.html)
- [Temporal](https://docs.temporal.io/)
- [WAHA - WhatsApp HTTP API](https://waha.devlike.pro/)

## 📄 Licença

MIT
