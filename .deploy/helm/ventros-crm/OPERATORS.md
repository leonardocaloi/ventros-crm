# Using Existing Operators

Este documento explica como usar operators já instalados no cluster ao invés de instalá-los novamente.

## 🎯 Conceito

**Operators** são instalados **cluster-wide** (em todo o cluster) e gerenciam **Custom Resources (CRs)**.

Se o operator já está instalado, você só precisa:
1. ✅ **Criar os Custom Resources** (postgresql, rabbitmqcluster, etc)
2. ❌ **NÃO instalar o operator novamente**

## 📦 Operators Suportados

### 1. Zalando Postgres Operator

#### Verificar se já está instalado:
```bash
kubectl get deployment -n postgres-operator postgres-operator
# ou
kubectl get crd postgresqls.acid.zalan.do
```

#### Opção A: Operator JÁ instalado no cluster
```yaml
# values-production.yaml
postgresOperator:
  enabled: true
  installOperator: false  # ← NÃO instala o operator
  createCluster: true     # ← Cria apenas o cluster PostgreSQL
  cluster:
    teamId: "ventros"
    numberOfInstances: 2
    volumeSize: 100Gi
    storageClass: "gp3"
```

#### Opção B: Instalar operator + cluster
```yaml
# values-dev.yaml
postgresOperator:
  enabled: true
  installOperator: true   # ← Instala o operator
  createCluster: true     # ← Cria o cluster PostgreSQL
```

#### Opção C: Desabilitar completamente
```yaml
postgresOperator:
  enabled: false

# Use PostgreSQL externo
externalPostgresql:
  host: "postgres.example.com"
  port: 5432
  database: "ventros_crm"
  username: "ventros"
```

### 2. RabbitMQ Cluster Operator

#### Verificar se já está instalado:
```bash
kubectl get deployment -n rabbitmq-system rabbitmq-cluster-operator
# ou
kubectl get crd rabbitmqclusters.rabbitmq.com
```

#### Opção A: Operator JÁ instalado no cluster
```yaml
# values-production.yaml
rabbitmqOperator:
  enabled: true
  installOperator: false  # ← NÃO instala o operator
  createCluster: true     # ← Cria apenas o cluster RabbitMQ
  cluster:
    replicas: 3
    volumeSize: 50Gi
    storageClass: "gp3"
```

#### Opção B: Instalar manualmente (recomendado)
```bash
# Instalar o operator manualmente (uma vez por cluster)
kubectl apply -f "https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml"

# Depois use o Helm apenas para criar o cluster
helm install ventros-crm .deploy/helm/ventros-crm/ \
  --set rabbitmqOperator.enabled=true \
  --set rabbitmqOperator.installOperator=false \
  --set rabbitmqOperator.createCluster=true
```

#### Opção C: Usar RabbitMQ Bitnami (mais simples)
```yaml
# values.yaml (padrão)
rabbitmqOperator:
  enabled: false

rabbitmq:
  enabled: true  # ← Usa subchart Bitnami (sem operator)
  replicaCount: 3
```

## 🚀 Exemplos de Instalação

### Cenário 1: Cluster novo (sem operators)
```bash
# Instala tudo (operators + clusters)
helm install ventros-crm .deploy/helm/ventros-crm/ \
  --set postgresOperator.installOperator=true \
  --set postgresOperator.createCluster=true
```

### Cenário 2: Cluster com Postgres Operator já instalado
```bash
# Usa operator existente, cria apenas o cluster
helm install ventros-crm .deploy/helm/ventros-crm/ \
  --set postgresOperator.installOperator=false \
  --set postgresOperator.createCluster=true
```

### Cenário 3: Cluster com TODOS os operators já instalados
```bash
# Usa operators existentes, cria apenas os clusters
helm install ventros-crm .deploy/helm/ventros-crm/ \
  -f values-use-existing-operators.yaml
```

**values-use-existing-operators.yaml:**
```yaml
postgresOperator:
  enabled: true
  installOperator: false
  createCluster: true

rabbitmqOperator:
  enabled: true
  installOperator: false
  createCluster: true

# Desabilita subcharts que não precisam de operator
redis:
  enabled: true
rabbitmq:
  enabled: false  # Usando operator ao invés do Bitnami
```

### Cenário 4: Produção com storageClass customizado
```bash
helm install ventros-crm .deploy/helm/ventros-crm/ \
  --set postgresOperator.installOperator=false \
  --set postgresOperator.cluster.storageClass=io2 \
  --set postgresOperator.cluster.volumeSize=200Gi \
  --set redis.master.persistence.storageClass=gp3 \
  --set redis.replica.persistence.storageClass=gp3
```

## 🔍 Verificar o que foi instalado

### Ver operators instalados:
```bash
# Postgres Operator
kubectl get deployment -A | grep postgres-operator

# RabbitMQ Operator
kubectl get deployment -A | grep rabbitmq.*operator
```

### Ver clusters criados:
```bash
# PostgreSQL clusters
kubectl get postgresql -A

# RabbitMQ clusters
kubectl get rabbitmqcluster -A
```

### Ver recursos do Ventros CRM:
```bash
# Tudo do Ventros CRM
kubectl get all -n ventros-crm

# Apenas PostgreSQL
kubectl get postgresql -n ventros-crm

# Apenas RabbitMQ
kubectl get rabbitmqcluster -n ventros-crm
```

## 📋 Fluxo de Decisão

```
┌─────────────────────────────────────┐
│ Operator já está instalado?        │
└─────────────┬───────────────────────┘
              │
      ┌───────┴───────┐
      │               │
     SIM             NÃO
      │               │
      ▼               ▼
installOperator:  installOperator:
    false             true
      │               │
      └───────┬───────┘
              │
              ▼
        createCluster:
            true
              │
              ▼
    ┌─────────────────┐
    │ Cluster criado! │
    └─────────────────┘
```

## ⚠️ Notas Importantes

1. **Operators são cluster-wide**: Um operator instalado serve TODOS os namespaces
2. **CRDs são globais**: Custom Resource Definitions são compartilhadas
3. **Não duplicar**: Instalar o mesmo operator 2x causa conflitos
4. **Verificar versão**: Certifique-se que a versão do operator é compatível
5. **Permissões**: Operators precisam de ClusterRole (admin do cluster)

## 🔧 Troubleshooting

### Erro: "CRD already exists"
```bash
# O operator já está instalado
# Solução: Use installOperator: false
```

### Erro: "no matches for kind PostgreSQL"
```bash
# O operator NÃO está instalado
# Solução: Use installOperator: true ou instale manualmente
```

### Ver logs do operator:
```bash
# Postgres Operator
kubectl logs -n postgres-operator deployment/postgres-operator -f

# RabbitMQ Operator
kubectl logs -n rabbitmq-system deployment/rabbitmq-cluster-operator -f
```

### Verificar se o cluster foi criado:
```bash
# PostgreSQL
kubectl describe postgresql ventros-crm-postgres -n ventros-crm

# RabbitMQ
kubectl describe rabbitmqcluster ventros-crm-rabbitmq -n ventros-crm
```

## 📚 Referências

- [Zalando Postgres Operator](https://github.com/zalando/postgres-operator)
- [RabbitMQ Cluster Operator](https://github.com/rabbitmq/cluster-operator)
- [Helm Conditions](https://helm.sh/docs/chart_best_practices/dependencies/#conditions-and-tags)
