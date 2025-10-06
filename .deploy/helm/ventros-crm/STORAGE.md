# Storage Configuration

Este documento descreve como configurar o `storageClass` para todos os componentes que persistem dados no Ventros CRM.

**Nota:** Por padrão, todos os componentes usam o storageClass default do cluster (deixando vazio `""`). Você pode customizar conforme necessário.

## 📦 Componentes com Persistência

### 1. PostgreSQL (Zalando Operator)

```yaml
postgresOperator:
  cluster:
    volumeSize: 10Gi
    storageClass: ""  # Vazio = usa default do cluster
```

**Exemplo com storageClass customizado:**
```yaml
postgresOperator:
  cluster:
    volumeSize: 50Gi
    storageClass: "seu-storage-class"
```

### 2. Redis (Master)

```yaml
redis:
  master:
    persistence:
      enabled: true
      size: 8Gi
      storageClass: ""  # Vazio = usa default do cluster
```

**Exemplo customizado:**
```yaml
redis:
  master:
    persistence:
      size: 20Gi
      storageClass: "seu-storage-class"
```

### 3. Redis (Replicas)

```yaml
redis:
  replica:
    persistence:
      enabled: true
      size: 8Gi
      storageClass: ""  # Vazio = usa default do cluster
```

### 4. RabbitMQ (Bitnami)

```yaml
rabbitmq:
  persistence:
    enabled: true
    size: 10Gi
    storageClass: ""  # Vazio = usa default do cluster
```

**Exemplo customizado:**
```yaml
rabbitmq:
  persistence:
    size: 30Gi
    storageClass: "seu-storage-class"
```

### 5. RabbitMQ (Cluster Operator)

```yaml
rabbitmqOperator:
  cluster:
    volumeSize: 10Gi
    storageClass: ""  # Vazio = usa default do cluster
```

## 🎯 StorageClasses Comuns

### AWS EKS
- `gp2` - General Purpose SSD (padrão antigo)
- `gp3` - General Purpose SSD (recomendado)
- `io1` / `io2` - Provisioned IOPS SSD (alta performance)
- `st1` - Throughput Optimized HDD
- `sc1` - Cold HDD

### Google GKE
- `standard` - Standard persistent disk (padrão)
- `standard-rwo` - Standard RWO
- `premium-rwo` - SSD persistent disk

### Azure AKS
- `default` - Azure Disk (padrão)
- `managed-premium` - Premium SSD
- `azurefile` - Azure Files
- `azurefile-premium` - Azure Files Premium

### On-Premise / Minikube
- `standard` - hostPath (padrão)
- `local-path` - Local Path Provisioner
- Ou seu provisioner customizado (Longhorn, Rook-Ceph, etc)

## 📋 Exemplo Completo

```yaml
# values-custom.yaml

# Customizar storageClass para todos os componentes
postgresOperator:
  cluster:
    volumeSize: 100Gi
    storageClass: "seu-storage-class"

redis:
  master:
    persistence:
      size: 20Gi
      storageClass: "seu-storage-class"
  replica:
    persistence:
      size: 20Gi
      storageClass: "seu-storage-class"

rabbitmq:
  persistence:
    size: 50Gi
    storageClass: "seu-storage-class"
```

## 🔍 Verificar StorageClasses Disponíveis

```bash
# Listar storageClasses no cluster
kubectl get storageclass

# Ver detalhes de um storageClass
kubectl describe storageclass <nome-do-storage-class>

# Ver qual é o default
kubectl get storageclass | grep "(default)"
```

## ⚙️ Instalação com StorageClass Customizado

```bash
# Opção 1: Via --set
helm install ventros-crm .deploy/helm/ventros-crm/ \
  --set postgresOperator.cluster.storageClass=seu-storage-class \
  --set redis.master.persistence.storageClass=seu-storage-class \
  --set redis.replica.persistence.storageClass=seu-storage-class \
  --set rabbitmq.persistence.storageClass=seu-storage-class

# Opção 2: Via values file
helm install ventros-crm .deploy/helm/ventros-crm/ \
  -f values-custom.yaml
```

## 🚨 Notas Importantes

1. **Default StorageClass**: Se deixar vazio (`""`), o Kubernetes usa o storageClass marcado como default no cluster
2. **Migração**: Mudar storageClass requer recriar os PVCs (faça backup antes!)
3. **Binding Mode**: Verifique se o storageClass usa `WaitForFirstConsumer` ou `Immediate`
4. **Disponibilidade**: Use `kubectl get storageclass` para ver as opções disponíveis no seu cluster
