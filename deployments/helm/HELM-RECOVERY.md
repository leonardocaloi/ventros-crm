# 🔧 Guia de Recuperação do Helm

## Problema: Helm Upgrade/Install Travado

Quando o Helm fica travado em estado `pending-upgrade` ou `pending-install`, siga este guia.

---

## 🚨 Causa Raiz Identificada

O problema ocorria porque **recursos principais estavam incorretamente marcados como Helm Hooks**:

- ❌ `deployment.yaml` - estava marcado como hook (CORRIGIDO)
- ❌ `postgres-cluster.yaml` - estava marcado como hook (CORRIGIDO)  
- ❌ `rabbitmq-cluster.yaml` - estava marcado como hook (CORRIGIDO)

**Helm Hooks** são executados APÓS o upgrade, causando deadlock.

---

## ✅ Correções Aplicadas

### 1. Removidas Anotações de Hook dos Recursos Principais

```yaml
# ANTES (INCORRETO):
metadata:
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-weight": "10"

# DEPOIS (CORRETO):
metadata:
  # Sem anotações de hook
```

### 2. Adicionada Estratégia de Deployment

```yaml
# values.yaml e values-dev.yaml
deploymentStrategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1
    maxUnavailable: 0  # Zero downtime

progressDeadlineSeconds: 600  # 10 minutos timeout
revisionHistoryLimit: 3       # Manter 3 versões para rollback
```

---

## 🔄 Como Recuperar de Estado Travado

### Opção 1: Desinstalar e Reinstalar (RECOMENDADO para dev)

```bash
# 1. Desinstalar completamente
helm uninstall ventros-crm -n ventros-crm

# 2. Limpar recursos órfãos (se necessário)
kubectl delete all -l app.kubernetes.io/instance=ventros-crm -n ventros-crm

# 3. Reinstalar
helm install ventros-crm ./deployments/helm/ventros-crm \
  --namespace ventros-crm \
  --create-namespace \
  --values deployments/helm/ventros-crm/values-dev.yaml \
  --timeout 15m
```

### Opção 2: Forçar Limpeza do Estado (se uninstall falhar)

```bash
# 1. Listar releases travados
helm list -n ventros-crm --all

# 2. Deletar secrets do Helm manualmente
kubectl get secrets -n ventros-crm -l owner=helm

# 3. Deletar o secret específico do release travado
kubectl delete secret -n ventros-crm sh.helm.release.v1.ventros-crm.v<REVISION>

# 4. Tentar uninstall novamente
helm uninstall ventros-crm -n ventros-crm --no-hooks
```

### Opção 3: Rollback (se upgrade falhou)

```bash
# 1. Ver histórico
helm history ventros-crm -n ventros-crm

# 2. Rollback para última versão funcional
helm rollback ventros-crm <REVISION> -n ventros-crm

# Se rollback falhar, use --force
helm rollback ventros-crm <REVISION> -n ventros-crm --force
```

---

## 📋 Checklist Pré-Instalação

Antes de instalar/atualizar, verifique:

- [ ] Namespace existe: `kubectl get ns ventros-crm`
- [ ] CRDs instalados (se usando operators):
  ```bash
  kubectl get crd | grep -E "postgresql|rabbitmq"
  ```
- [ ] Imagem disponível (se usando local):
  ```bash
  minikube image ls | grep ventros-crm
  ```
- [ ] Valores corretos no values file
- [ ] Timeout adequado (mínimo 10m para primeira instalação)

---

## 🎯 Instalação Correta

### Desenvolvimento (Minikube)

```bash
helm upgrade ventros-crm ./deployments/helm/ventros-crm \
  --install \
  --namespace ventros-crm \
  --create-namespace \
  --values deployments/helm/ventros-crm/values-dev.yaml \
  --timeout 15m \
  --wait \
  --debug
```

### Produção

```bash
helm upgrade ventros-crm ./deployments/helm/ventros-crm \
  --install \
  --namespace ventros-crm \
  --create-namespace \
  --values deployments/helm/ventros-crm/values.yaml \
  --timeout 15m \
  --wait \
  --atomic
```

**Flags importantes:**
- `--timeout 15m`: Tempo máximo de espera
- `--wait`: Espera todos os recursos ficarem prontos
- `--atomic`: Rollback automático se falhar
- `--debug`: Mostra detalhes do processo

---

## 🔍 Diagnóstico de Problemas

### Ver status do release

```bash
helm status ventros-crm -n ventros-crm
```

### Ver logs do deployment

```bash
kubectl logs -n ventros-crm deployment/ventros-crm --tail=100 -f
```

### Ver eventos do namespace

```bash
kubectl get events -n ventros-crm --sort-by='.lastTimestamp'
```

### Ver pods com problemas

```bash
kubectl get pods -n ventros-crm
kubectl describe pod <POD_NAME> -n ventros-crm
```

### Verificar Jobs de Hook (se houver)

```bash
# Listar jobs
kubectl get jobs -n ventros-crm

# Ver logs de job específico
kubectl logs -n ventros-crm job/ventros-crm-migration
```

---

## ⚠️ Problemas Comuns

### 1. "another operation is in progress"

**Causa:** Release travado em estado pendente

**Solução:** Use Opção 2 (Forçar Limpeza) acima

### 2. "context deadline exceeded"

**Causa:** Timeout muito curto ou recursos demorando para iniciar

**Solução:** 
- Aumente `--timeout` para 15m ou 20m
- Verifique se initContainers estão travados
- Verifique se dependências (PostgreSQL, Redis, etc) estão prontas

### 3. "ImagePullBackOff"

**Causa:** Imagem não encontrada

**Solução (Minikube):**
```bash
# Build e load da imagem
docker build -t ventros-crm:latest .
minikube image load ventros-crm:latest
```

### 4. Pods em CrashLoopBackOff

**Causa:** Aplicação falhando ao iniciar

**Solução:**
```bash
# Ver logs
kubectl logs -n ventros-crm deployment/ventros-crm --previous

# Verificar probes
kubectl describe pod -n ventros-crm <POD_NAME> | grep -A 10 Liveness
```

---

## 📊 Monitoramento Durante Instalação

Em um terminal separado, monitore:

```bash
# Watch pods
watch kubectl get pods -n ventros-crm

# Watch events
kubectl get events -n ventros-crm --watch

# Watch helm status
watch helm list -n ventros-crm
```

---

## 🎉 Verificação Pós-Instalação

```bash
# 1. Verificar release
helm list -n ventros-crm

# 2. Verificar pods
kubectl get pods -n ventros-crm

# 3. Verificar services
kubectl get svc -n ventros-crm

# 4. Testar aplicação
kubectl port-forward -n ventros-crm svc/ventros-crm 8080:8080
curl http://localhost:8080/health
```

---

## 📝 Notas Importantes

1. **Hooks vs Recursos Normais:**
   - Hooks: Executam APÓS install/upgrade (Jobs temporários)
   - Recursos Normais: Gerenciados durante install/upgrade (Deployments, Services, etc)

2. **Progress Deadline:**
   - Define quanto tempo o Kubernetes espera o deployment completar
   - Padrão: 600s (10 minutos)
   - Ajuste conforme necessário

3. **Rolling Update Strategy:**
   - `maxSurge: 1`: Permite 1 pod extra durante update
   - `maxUnavailable: 0`: Garante zero downtime
   - Pods novos só substituem antigos quando prontos

4. **Revision History:**
   - Mantém 3 versões antigas por padrão
   - Permite rollback rápido se necessário
   - Limpa automaticamente versões antigas

---

## 🆘 Suporte

Se o problema persistir:

1. Capture logs completos:
   ```bash
   kubectl logs -n ventros-crm deployment/ventros-crm > app.log
   kubectl describe deployment -n ventros-crm ventros-crm > deployment.yaml
   helm get values ventros-crm -n ventros-crm > current-values.yaml
   ```

2. Verifique configurações:
   ```bash
   helm get manifest ventros-crm -n ventros-crm > manifest.yaml
   ```

3. Revise este guia e as correções aplicadas
