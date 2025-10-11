# 🏗️ Multi-Product Architecture - Plano de Refatoração

**Data**: 2025-10-11
**Status**: 📝 Proposta para Revisão
**Impacto**: 🔴 Alto (mudança arquitetural)

---

## 🎯 Problema Atual

Você tem uma **confusão conceitual** entre:
- **Customer** (não implementado)
- **Project** (implementado, mas usado como tenant_id)
- **Product** (não existe, mas deveria)

### Hierarquia ATUAL (incorreta)
```
Project (tenant_id)
├── Todas as entidades CRM
└── (sem separação de produtos)
```

### Hierarquia DESEJADA (correta)
```
Customer (Empresa Cliente - "Acme Corp")
├── Subscription/Billing
└── Products (Serviços contratados)
    ├── CRM Product
    │   ├── Workspace 1 ("Vendas")
    │   │   ├── Contacts, Messages, Sessions
    │   │   ├── Channels, Pipelines, Agents
    │   │   └── ...
    │   └── Workspace 2 ("Suporte")
    ├── BI Product (futuro)
    │   └── Workspace 1
    └── Automation Product (futuro)
        └── Workspace 1
```

---

## 🎯 Objetivos

1. ✅ **Separação de Produtos** - CRM, BI, Automation são produtos distintos
2. ✅ **Cobrança separada** - Billing por produto
3. ✅ **Organização lógica** - Bounded contexts claros
4. ✅ **Escalabilidade** - Preparar para novos produtos
5. ✅ **Sem quebrar o código existente** - Migração evolutiva

---

## 📊 Opções de Implementação

### **Opção 1: Adicionar Product + Renomear Project → Workspace**
**⭐ RECOMENDADA - Evolutiva e pragmática**

```
Customer
├── Product (CRM)
│   ├── Workspace 1 (ex: "Vendas")
│   │   ├── tenant_id: "acme-crm-vendas"
│   │   ├── Contacts, Messages, Sessions...
│   │   └── Channels, Pipelines, Agents...
│   └── Workspace 2 (ex: "Suporte")
├── Product (BI) - futuro
│   └── Workspace 1
└── Product (Automation) - futuro
    └── Workspace 1
```

**Schema:**
```sql
-- customers (já existe no domain, falta implementar)
CREATE TABLE customers (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL, -- active, suspended, inactive
    settings JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- products (NOVO)
CREATE TABLE products (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id),
    product_type VARCHAR(50) NOT NULL, -- 'crm', 'bi', 'automation'
    name VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    settings JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(customer_id, product_type)
);

-- projects (RENOMEAR → workspaces no futuro, mas manter por ora)
-- ADICIONAR product_id
ALTER TABLE projects ADD COLUMN product_id UUID REFERENCES products(id);
```

**Mudanças no código:**
```go
// internal/domain/project/project.go
type Project struct {
    id                    uuid.UUID
    customerID            uuid.UUID       // já existe
    billingAccountID      uuid.UUID       // já existe
    productID             uuid.UUID       // NOVO
    tenantID              string          // continua (RLS)
    name                  string
    // ...
}
```

**Vantagens:**
- ✅ Adiciona Product sem quebrar nada
- ✅ Mantém todo o código existente funcionando
- ✅ Permite migração gradual
- ✅ Tenant_id continua funcionando (RLS)
- ✅ Preparado para BI/Automation

**Desvantagens:**
- ⚠️ Project continua com nome confuso (mas funcional)
- ⚠️ Precisa de data migration

---

### **Opção 2: Bounded Contexts Explícitos (DDD Puro)**
**❌ NÃO RECOMENDADA AGORA - Muito disruptiva**

```
internal/domain/
├── core/
│   ├── customer/
│   └── product/
├── crm/          # Bounded Context CRM
│   ├── contact/
│   ├── message/
│   ├── session/
│   ├── channel/
│   └── ...
├── bi/           # Bounded Context BI
└── automation/   # Bounded Context Automation
```

**Vantagens:**
- ✅ Arquitetura DDD perfeita
- ✅ Separação total de contextos

**Desvantagens:**
- ❌ Quebra TODO o código existente
- ❌ Refatoração massiva (semanas)
- ❌ Alto risco de bugs
- ❌ Não agrega valor imediato

---

### **Opção 3: Deixar como está + Tags**
**❌ NÃO RECOMENDADA - Não resolve o problema**

Apenas documentar quais entidades são do CRM, BI, etc.

**Vantagens:**
- ✅ Zero mudanças

**Desvantagens:**
- ❌ Não resolve billing separado
- ❌ Não prepara para novos produtos
- ❌ Confusão conceitual continua

---

## 🎯 Recomendação: Opção 1 (Product + Workspace)

### Roadmap de Implementação

#### **Fase 1: Fundação (1-2 dias)** ✅ FAZER AGORA

1. **Criar entidade Product**
   - [ ] `internal/domain/product/product.go`
   - [ ] `internal/domain/product/types.go` (ProductType: CRM, BI, Automation)
   - [ ] `internal/domain/product/events.go`
   - [ ] `internal/domain/product/repository.go`

2. **Implementar Customer (já existe no domain)**
   - [ ] `infrastructure/persistence/entities/customer.go`
   - [ ] `infrastructure/persistence/gorm_customer_repository.go`
   - [ ] Migration `create_customers.up.sql`

3. **Implementar Product**
   - [ ] `infrastructure/persistence/entities/product.go`
   - [ ] `infrastructure/persistence/gorm_product_repository.go`
   - [ ] Migration `create_products.up.sql`

4. **Adicionar product_id ao Project**
   - [ ] Alterar `internal/domain/project/project.go`
   - [ ] Alterar `infrastructure/persistence/entities/project.go`
   - [ ] Migration `add_product_id_to_projects.up.sql`

#### **Fase 2: Migração de Dados (1 dia)** ⚠️ CRÍTICO

5. **Data Migration**
   ```sql
   -- Criar customer default para dados existentes
   INSERT INTO customers (id, name, email, status, created_at, updated_at)
   VALUES (gen_random_uuid(), 'Default Customer', 'admin@ventros.com', 'active', NOW(), NOW());

   -- Criar product CRM default
   INSERT INTO products (id, customer_id, product_type, name, active, created_at, updated_at)
   SELECT gen_random_uuid(), c.id, 'crm', 'Ventros CRM', true, NOW(), NOW()
   FROM customers c WHERE c.email = 'admin@ventros.com';

   -- Associar todos os projects existentes ao product CRM
   UPDATE projects
   SET product_id = (SELECT id FROM products WHERE product_type = 'crm' LIMIT 1)
   WHERE product_id IS NULL;
   ```

#### **Fase 3: Lógica de Negócio (2-3 dias)**

6. **Use Cases**
   - [ ] CreateCustomerUseCase
   - [ ] CreateProductUseCase (CRM, BI, Automation)
   - [ ] CreateWorkspaceUseCase (associar ao product)

7. **HTTP Handlers**
   - [ ] CustomerHandler (CRUD)
   - [ ] ProductHandler (CRUD)
   - [ ] Atualizar ProjectHandler (adicionar product_id)

8. **Billing Integration**
   - [ ] Billing por Product (não por Project)
   - [ ] BillingAccount → Products (one-to-many)

#### **Fase 4: Refatoração Opcional (FUTURO)**

9. **Renomear Project → Workspace** (opcional, não urgente)
   - [ ] Renomear tabela: `ALTER TABLE projects RENAME TO workspaces`
   - [ ] Renomear código: `Project` → `Workspace`
   - [ ] Atualizar docs

---

## 📋 Schema Completo Proposto

```sql
customers (top-level)
    ├── id
    ├── name
    ├── email
    ├── status (active, suspended)
    └── settings (JSONB)

products (CRM, BI, Automation)
    ├── id
    ├── customer_id → customers.id
    ├── product_type (crm, bi, automation)
    ├── name
    ├── active
    └── settings (JSONB)

projects (workspaces dentro de um product)
    ├── id
    ├── customer_id → customers.id
    ├── billing_account_id → billing_accounts.id
    ├── product_id → products.id (NOVO)
    ├── tenant_id (para RLS)
    ├── name
    └── configuration (JSONB)

billing_accounts
    ├── id
    ├── customer_id → customers.id
    └── ... (billing info)

-- Todas as entidades CRM continuam como estão
contacts, messages, sessions, channels, pipelines, agents...
    ├── tenant_id (RLS)
    └── ... (sem mudanças)
```

---

## 🔄 Impacto nas Entidades CRM

**✅ ZERO IMPACTO** nas entidades CRM existentes:
- Contacts, Messages, Sessions, Channels, Pipelines, Agents, Notes, etc.
- Todas continuam usando `tenant_id` (RLS)
- Nenhuma mudança necessária

**⚠️ ÚNICO IMPACTO** no Project:
- Adicionar campo `product_id`
- Lógica de criação deve associar ao Product

---

## 📊 Exemplo de Uso

### Criar Cliente com CRM

```go
// 1. Criar Customer
customer, _ := customer.NewCustomer("Acme Corp", "admin@acme.com")
customerRepo.Save(ctx, customer)

// 2. Criar BillingAccount
billingAccount, _ := billing.NewBillingAccount(customer.ID())
billingRepo.Save(ctx, billingAccount)

// 3. Criar Product CRM
productCRM, _ := product.NewProduct(
    customer.ID(),
    product.ProductTypeCRM,
    "Acme CRM",
)
productRepo.Save(ctx, productCRM)

// 4. Criar Workspace (Project) dentro do CRM
workspace, _ := project.NewProject(
    customer.ID(),
    billingAccount.ID(),
    productCRM.ID(),           // NOVO
    "acme-crm-vendas",         // tenant_id (RLS)
    "Acme - Vendas",
)
projectRepo.Save(ctx, workspace)

// 5. Usar o CRM normalmente
// Contacts, Messages, Sessions... tudo funciona com tenant_id
```

### Futuro: Adicionar BI

```go
// 1. Criar Product BI
productBI, _ := product.NewProduct(
    customer.ID(),
    product.ProductTypeBI,
    "Acme BI",
)
productRepo.Save(ctx, productBI)

// 2. Criar Workspace (Project) dentro do BI
workspaceBI, _ := project.NewProject(
    customer.ID(),
    billingAccount.ID(),
    productBI.ID(),
    "acme-bi-analytics",
    "Acme - Analytics",
)
projectRepo.Save(ctx, workspaceBI)

// 3. Billing separado por produto
billingService.CalculateUsage(customer.ID(), product.ProductTypeCRM)
billingService.CalculateUsage(customer.ID(), product.ProductTypeBI)
```

---

## 🎯 Decisão Necessária

**Você precisa decidir:**

1. ✅ **Opção 1 (Recomendada)**: Implementar Product + adicionar product_id ao Project
2. ❌ **Opção 2**: Refatoração massiva com Bounded Contexts (não recomendado agora)
3. ❌ **Opção 3**: Deixar como está (não resolve)

**Minha sugestão**:
- ✅ Implementar **Opção 1** agora (Fase 1-2)
- ⏳ Deixar Fase 3-4 para depois (não crítico)
- 📚 Documentar bem a nova hierarquia

---

## ❓ Perguntas para Você

1. **Quando você quer ter BI e Automation prontos?**
   - Se for >6 meses: pode deixar Product como preparação
   - Se for <3 meses: melhor implementar Product agora

2. **Billing já está funcionando?**
   - Se sim: precisa migrar para billing por Product
   - Se não: implementar já no modelo novo

3. **Você quer renomear Project → Workspace?**
   - Se sim: fazer na Fase 4
   - Se não: deixar como Project (funciona igual)

4. **Quantos clientes você tem em produção?**
   - Se 0: fácil, sem data migration
   - Se >0: precisa planejar data migration

---

## 🚀 Próximos Passos

**SE VOCÊ CONCORDAR COM OPÇÃO 1:**

Eu posso implementar agora:
1. ✅ Criar entidade Product (domain)
2. ✅ Implementar Customer (persistence)
3. ✅ Implementar Product (persistence)
4. ✅ Adicionar product_id ao Project
5. ✅ Criar migrations
6. ✅ Criar data migration script

**Tempo estimado**: 4-6 horas

**O que você acha?**
