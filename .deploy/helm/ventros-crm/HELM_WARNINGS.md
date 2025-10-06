# Helm Warnings - Ventros CRM

Este documento descreve os warnings conhecidos do Helm e suas soluções.

## Warnings Resolvidos ✅

### 1. `Warning: unrecognized format "int32"`

**Causa**: O CRD do PostgreSQL Operator (Zalando) usava `format: int32` que não é reconhecido em todas as versões do Kubernetes.

**Solução**: Removido o atributo `format: int32` do campo `weight` nos arquivos:
- `crds/postgresql.crd.yaml` (linha 279)
- `charts/postgres-operator/crds/postgresqls.yaml` (linha 281)

### 2. `coalesce.go:237: warning: skipped value for rabbitmq.initContainers: Not a table`

**Causa**: O chart Bitnami RabbitMQ (versão 12.x) não suporta o parâmetro `initContainers` diretamente no formato de subchart.

**Solução**: Removido `initContainers: []` das configurações do RabbitMQ. Se precisar de init containers no RabbitMQ, use o parâmetro `commonAnnotations` ou configure diretamente no RabbitmqCluster CR (quando usando o operator).

**Alternativa**: Para adicionar init containers, use:
```yaml
rabbitmq:
  extraDeploy:
    - apiVersion: v1
      kind: Pod
      # ... definição customizada
```

### 3. `Warning: spec.SessionAffinity is ignored for headless services`

**Causa**: Os subcharts Bitnami (Redis) criam headless services que, por design do Kubernetes, ignoram a configuração de SessionAffinity.

**Solução**: Configurado `sessionAffinity: ""` (vazio) no `values-dev.yaml` para desabilitar explicitamente:
```yaml
redis:
  master:
    service:
      sessionAffinity: ""
  replica:
    service:
      sessionAffinity: ""
```

**Referência**: [Kubernetes Headless Services](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services)

## Warnings Informativos ℹ️

### 4. `Warning: unknown field "configuration.kubernetes.enable_owner_references"`

**Causa**: CRD `OperatorConfiguration` antigo instalado no cluster não contém o campo `enable_owner_references`, mas o chart do PostgreSQL Operator v1.14.0 tenta configurá-lo.

**Impacto**: Mínimo. O operator funciona normalmente com o valor default.

**Solução**: 
1. Recriar o namespace (deleta o CRD antigo)
2. O Helm instalará o CRD atualizado que inclui esse campo (`crds/operatorconfiguration.crd.yaml:212-214`)

**Nota**: Os CRDs **não** são atualizados automaticamente pelo `helm upgrade`. É necessário deletá-los manualmente ou recriar o namespace.

## Comandos de Verificação

```bash
# Verificar status dos recursos criados
kubectl get postgresql,rabbitmqcluster,redis -n ventros-crm

# Verificar logs do PostgreSQL Operator
kubectl logs -n ventros-crm -l app.kubernetes.io/name=postgres-operator

# Verificar warnings do Helm
helm upgrade --install ventros-crm ./ventros-crm \
  -n ventros-crm \
  --create-namespace \
  -f values-dev.yaml \
  --dry-run --debug 2>&1 | grep -i warning
```

## Resumo

✅ **Corrigidos**: 3 warnings (int32 format, rabbitmq.initContainers, SessionAffinity)  
ℹ️ **Informativos**: 1 warning (enable_owner_references - resolve ao recriar namespace)

**Status**: Todos os warnings foram tratados. Após recriar o namespace, nenhum warning deverá aparecer.
