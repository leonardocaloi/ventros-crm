# 🕐 Configuração Automática do Temporal

## 📋 Visão Geral

Este Helm Chart está configurado para instalar o **Temporal** automaticamente usando o **PostgreSQL** gerenciado pelo Zalando Postgres Operator, com **credenciais geradas dinamicamente**.

## ✅ O que Já Está Configurado

### 1️⃣ **Database e Usuário Compartilhado (Padrão Docker-Compose)**

Quando você habilita o Temporal (`temporal.enabled: true`), o Postgres Operator **cria automaticamente**:

**Usuário:**
- 👤 `ventros` - Usado pela aplicação E pelo Temporal (superuser + createdb)

**Database:**
- 📦 `ventros_crm` - Database único compartilhado
  - Aplicação usa schema `public` (tabelas normais)
  - Temporal cria schemas separados automaticamente:
    - `temporal` - Workflows, timers, histórico
    - `temporal_visibility` - Índice para queries e busca

**Configuração:** `templates/postgres-cluster.yaml`

```yaml
users:
  ventros:  # Único usuário compartilhado
    - superuser
    - createdb

databases:
  ventros_crm: ventros  # Database único
```

### 🎯 **Por que esta abordagem?**
✅ **Simplicidade:** Apenas 1 database, 1 usuário  
✅ **Compatibilidade:** Igual ao `docker-compose.yml`  
✅ **Isolamento:** Temporal usa schemas separados  
✅ **Backups:** Backup único cobre tudo  
✅ **Migrations:** Gerenciamento simplificado

### 2️⃣ **Credenciais Automáticas**

O Postgres Operator gera automaticamente **1 Kubernetes Secret** compartilhado:

**Para aplicação E Temporal:**
```
Secret: ventros.ventros-crm-postgres.credentials.postgresql.acid.zalan.do
  - username: ventros
  - password: <gerado-automaticamente>
```

✅ **Ambos usam o mesmo secret!** Seguindo o padrão do docker-compose.

### 3️⃣ **Injeção Automática no Temporal**

O subchart do Temporal (`charts/temporal`) recebe automaticamente via `values.yaml`:

```yaml
temporal:
  server:
    config:
      persistence:
        default:
          sql:
            host: ventros-crm-postgres
            port: 5432
            database: temporal
            user: temporal_user  # Usuário dedicado
            existingSecret: temporal-user.ventros-crm-postgres.credentials.postgresql.acid.zalan.do
        visibility:
          sql:
            host: ventros-crm-postgres
            database: temporal_visibility
            user: temporal_user
            existingSecret: temporal-user.ventros-crm-postgres.credentials.postgresql.acid.zalan.do
```

**Helpers Usados:** `templates/_config-helpers.tpl`

- `ventros-crm.postgresql.host` → `ventros-crm-postgres`
- `ventros-crm.postgresql.secretName` → Nome do secret gerado
- `ventros-crm.temporal.database.default` → `temporal`
- `ventros-crm.temporal.database.visibility` → `temporal_visibility`

---

## 🚀 Como Habilitar o Temporal

### **Opção 1: Development (values-dev.yaml)**

```yaml
temporal:
  enabled: true  # ← Mudar de false para true
  useInternalPostgres: true  # Já está configurado
```

### **Opção 2: Production (values.yaml)**

```yaml
temporal:
  enabled: true  # ← Mudar de false para true
  useInternalPostgres: true  # Já está configurado
```

### **Comando de Instalação**

```bash
# Development
helm upgrade ventros-crm . \
  -n ventros-crm \
  -f values-dev.yaml \
  --timeout 20m

# Production
helm install ventros-crm . \
  -n ventros-crm \
  --create-namespace \
  --timeout 20m
```

---

## 🔍 Verificação

### 1. **Checar se os databases foram criados**

```bash
# Conectar ao PostgreSQL
kubectl exec -it ventros-crm-postgres-0 -n ventros-crm -- psql -U ventros

# Listar databases
\l

# Deve mostrar:
# - postgres
# - ventros_db (aplicação)
# - temporal ✅
# - temporal_visibility ✅
```

### 2. **Verificar o Secret**

```bash
# Ver o secret do Temporal
kubectl get secret \
  temporal-user.ventros-crm-postgres.credentials.postgresql.acid.zalan.do \
  -n ventros-crm \
  -o yaml

# Pegar password do Temporal
kubectl get secret \
  temporal-user.ventros-crm-postgres.credentials.postgresql.acid.zalan.do \
  -n ventros-crm \
  -o jsonpath='{.data.password}' | base64 -d

# Ver o secret da aplicação
kubectl get secret \
  ventros.ventros-crm-postgres.credentials.postgresql.acid.zalan.do \
  -n ventros-crm \
  -o yaml
```

### 3. **Verificar Pods do Temporal**

```bash
kubectl get pods -n ventros-crm -l app.kubernetes.io/name=temporal

# Deve mostrar:
# - temporal-frontend-xxxxx
# - temporal-history-xxxxx
# - temporal-matching-xxxxx
# - temporal-worker-xxxxx
# - temporal-web-xxxxx (UI)
```

### 4. **Acessar Temporal Web UI**

```bash
# Port-forward
kubectl port-forward -n ventros-crm svc/ventros-crm-temporal-web 8080:8080

# Abrir no browser
open http://localhost:8080
```

---

## 🏗️ Schema Setup

O Temporal precisa inicializar os schemas dos databases. Isso é feito automaticamente pelo **schema setup job**:

```yaml
temporal:
  schema:
    setup:
      enabled: true  # Job roda no install/upgrade
      backoffLimit: 10
    update:
      enabled: true  # Job roda updates de schema
```

### Jobs Executados

```bash
# Ver jobs
kubectl get jobs -n ventros-crm | grep temporal

# Logs do setup
kubectl logs -n ventros-crm job/ventros-crm-temporal-schema-setup

# Logs do update
kubectl logs -n ventros-crm job/ventros-crm-temporal-schema-update
```

---

## ⚙️ Configuração Avançada

### Ajustar Recursos

```yaml
temporal:
  server:
    resources:
      requests:
        cpu: 500m      # ← Ajustar conforme carga
        memory: 512Mi
      limits:
        cpu: 1000m
        memory: 1Gi
  
  web:
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
```

### Connection Pooling

```yaml
temporal:
  server:
    config:
      persistence:
        default:
          sql:
            maxConns: 20         # ← Max conexões ao DB
            maxIdleConns: 20     # ← Conexões idle
            maxConnLifetime: "1h" # ← Tempo de vida
```

---

## 🐛 Troubleshooting

### Temporal não conecta ao PostgreSQL

**Sintoma:** Pods do Temporal em `CrashLoopBackOff`

**Solução:**

```bash
# 1. Verificar se o PostgreSQL está rodando
kubectl get pods -n ventros-crm -l application=spilo

# 2. Testar conectividade
kubectl exec -it ventros-crm-postgres-0 -n ventros-crm -- \
  psql -U ventros -d temporal -c "SELECT 1;"

# 3. Ver logs do Temporal
kubectl logs -n ventros-crm <temporal-pod-name>
```

### Schema não foi inicializado

**Sintoma:** Erro `relation "executions" does not exist`

**Solução:**

```bash
# 1. Verificar se o setup job rodou
kubectl get jobs -n ventros-crm | grep schema-setup

# 2. Ver logs
kubectl logs -n ventros-crm job/ventros-crm-temporal-schema-setup

# 3. Re-rodar manualmente se necessário
kubectl delete job -n ventros-crm ventros-crm-temporal-schema-setup
helm upgrade ventros-crm . -n ventros-crm -f values-dev.yaml
```

### Secret não encontrado

**Sintoma:** `existingSecret not found`

**Solução:**

```bash
# 1. Verificar se o Postgres Operator criou o secret
kubectl get secrets -n ventros-crm | grep credentials

# 2. Se não existe, o PostgreSQL pode não estar pronto
kubectl get postgresql -n ventros-crm

# 3. Aguardar o PostgreSQL ficar Ready
kubectl wait --for=condition=Running \
  postgresql/ventros-crm-postgres \
  -n ventros-crm \
  --timeout=300s
```

---

## 📚 Referências

- [Temporal Helm Chart](https://github.com/temporalio/helm-charts)
- [Zalando Postgres Operator](https://github.com/zalando/postgres-operator)
- [Temporal Documentation](https://docs.temporal.io/)
- [PostgreSQL Configuration](https://www.postgresql.org/docs/current/runtime-config.html)

---

## ✨ Resumo

✅ **Databases** criados automaticamente pelo Postgres Operator  
✅ **Credenciais** geradas automaticamente via Kubernetes Secret  
✅ **Temporal** configurado para usar PostgreSQL interno  
✅ **Schema** inicializado automaticamente via setup job  
✅ **Web UI** acessível via port-forward  

**Para habilitar:** Mude `temporal.enabled: false` para `true` e rode `helm upgrade`! 🚀
