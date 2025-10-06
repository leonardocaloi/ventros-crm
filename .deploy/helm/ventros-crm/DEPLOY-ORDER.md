# 📋 Ordem de Deploy do Helm Chart

Este documento descreve a ordem de criação dos recursos no Kubernetes usando Helm Hooks.

## 🔄 Sequência de Deploy

### Fase 1: Recursos Base (Automático - sem hooks)
- **ServiceAccount** - Criado primeiro automaticamente
- **ConfigMaps** - Configurações da aplicação
- **Secrets** - Credenciais e senhas
- **Operators** (PostgreSQL, RabbitMQ) - Instalados via subcharts

### Fase 2: Infraestrutura de Dados (Hooks post-install/post-upgrade)

#### Weight: -10 (PostgreSQL Cluster)
```yaml
annotations:
  "helm.sh/hook": post-install,post-upgrade
  "helm.sh/hook-weight": "-10"
```
- **PostgreSQL Cluster** - Banco de dados principal
- Aguarda até o cluster estar `Running` e `Ready`
- Cria databases e usuários automaticamente

#### Weight: -5 (RabbitMQ Cluster)  
```yaml
annotations:
  "helm.sh/hook": post-install,post-upgrade
  "helm.sh/hook-weight": "-5"
```
- **RabbitMQ Cluster** - Message broker
- Aguarda até o cluster estar pronto
- `deletionPolicy: WaitForMessages` para cleanup seguro

### Fase 3: Temporal (Weight: 0 - padrão subchart)
- **Temporal** - Workflow engine
- Depende do PostgreSQL estar pronto
- Init containers aguardam PostgreSQL

### Fase 4: Database Migration (Weight: 5)
```yaml
annotations:
  "helm.sh/hook": post-install,post-upgrade
  "helm.sh/hook-weight": "5"
```
- **Migration Job** - Executa `./migrate-auth`
- Init container aguarda PostgreSQL estar disponível
- Timeout de 300s para aguardar o cluster
- `backoffLimit: 3` - até 3 tentativas

### Fase 5: Aplicação (Weight: 10 - último)
```yaml
annotations:
  "helm.sh/hook": post-install,post-upgrade
  "helm.sh/hook-weight": "10"
```
- **Deployment** - Aplicação Ventros CRM
- Aguarda **TUDO** estar pronto:
  - PostgreSQL cluster
  - Redis
  - RabbitMQ
  - Temporal
  - Migrations executadas
- Init containers verificam conectividade de cada serviço

## ✅ Benefícios desta Abordagem

1. **Ordem Garantida**: Cada recurso só é criado quando suas dependências estão prontas
2. **Sem Race Conditions**: Migration roda uma vez, antes da aplicação
3. **Temporal Otimizado**: Só tenta conectar quando PostgreSQL está pronto
4. **Troubleshooting Fácil**: Falhas são isoladas por fase
5. **Rollback Seguro**: Helm gerencia todo o ciclo de vida

## 🔍 Verificação da Ordem

Para ver a ordem de criação dos recursos:

```bash
# Ver todos os recursos com suas annotations
kubectl get all,postgresql,rabbitmqcluster,job -n ventros-crm -o yaml | grep -A 2 "helm.sh/hook"

# Ver ordem de criação por timestamp
kubectl get events -n ventros-crm --sort-by='.lastTimestamp'
```

## 🚨 Troubleshooting

### Migration Job Falha
```bash
# Ver logs do job
kubectl logs job/ventros-crm-migration -n ventros-crm

# Ver eventos
kubectl describe job ventros-crm-migration -n ventros-crm
```

### Temporal não conecta
```bash
# Verificar se PostgreSQL está pronto
kubectl get postgresql -n ventros-crm

# Ver logs do Temporal
kubectl logs -n ventros-crm -l app.kubernetes.io/component=frontend
```

### App em CrashLoopBackOff
```bash
# Verificar init containers
kubectl describe pod -n ventros-crm -l app.kubernetes.io/name=ventros-crm

# Ver qual serviço não está disponível
kubectl logs -n ventros-crm <pod-name> -c wait-for-<service>
```

## 📝 Notas Importantes

- **Hooks são síncronos**: Helm aguarda cada fase completar antes de prosseguir
- **Timeout padrão**: 5 minutos (ajustável com `--timeout`)
- **Delete policy**: Jobs de migration são deletados antes de nova instalação (`before-hook-creation`)
- **Idempotência**: Migrations devem ser idempotentes para suportar re-runs
