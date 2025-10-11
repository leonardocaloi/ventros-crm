# 🏗️ Ventros - User-Centric Architecture

**Data**: 2025-10-11
**Status**: ✅ Implementado
**Modelo**: User-Centric (Notion-like), não Customer-Centric

---

## 🎯 Conceito

Ventros é uma plataforma **User-Centric** onde:
- **User** (pessoa física) é o top-level
- Cada User pode criar múltiplos **Projects** (workspaces/tenants)
- Cada Project pode habilitar **Products** (CRM, BI, Automation)
- **BillingAccount** (Stripe) paga por múltiplos Projects

**NÃO é multi-customer**: não existe "Customer" ou "super admin gerenciando várias contas".

---

## 📊 Hierarquia

```
User (pessoa física - conta Google/Email)
├── Project 1 "Minha Loja"
│   ├── billing_account_id: BillingAccount 1
│   ├── tenant_id: "user123-loja" (RLS)
│   └── Products habilitados
│       ├── CRM (enabled: true)
│       ├── BI (enabled: false)
│       └── Automation (enabled: false)
│
├── Project 2 "Empresa do Cliente"
│   ├── billing_account_id: BillingAccount 2
│   ├── tenant_id: "user123-cliente" (RLS)
│   └── Products habilitados
│       └── CRM (enabled: true)
│
└── BillingAccounts (Stripe customers)
    ├── BillingAccount 1 (paga Project 1 + Project 3)
    │   ├── user_id: user123
    │   ├── stripe_customer_id: "cus_xxx"
    │   └── projects: [Project 1, Project 3]
    │
    └── BillingAccount 2 (paga Project 2)
        ├── user_id: user456
        ├── stripe_customer_id: "cus_yyy"
        └── projects: [Project 2]
```

---

## 🗄️ Database Schema

### **users** (top-level - pessoa física)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('admin','user','manager','readonly')),
    settings JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
```

### **billing_accounts** (Stripe customers)
```sql
CREATE TABLE billing_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    payment_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_methods JSONB,
    billing_email VARCHAR(255) NOT NULL,
    suspended BOOLEAN DEFAULT false,
    suspended_at TIMESTAMP,
    suspension_reason TEXT,
    stripe_customer_id VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_billing_accounts_user ON billing_accounts(user_id);
CREATE INDEX idx_billing_accounts_status ON billing_accounts(payment_status);
CREATE INDEX idx_billing_accounts_suspended ON billing_accounts(suspended);
```

### **projects** (workspaces/tenants)
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    billing_account_id UUID NOT NULL REFERENCES billing_accounts(id) ON DELETE RESTRICT,
    tenant_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    configuration JSONB, -- Pode conter enabled_products
    active BOOLEAN DEFAULT true,
    session_timeout_minutes INT DEFAULT 30,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_projects_tenant ON projects(tenant_id);
CREATE INDEX idx_projects_user ON projects(user_id);
CREATE INDEX idx_projects_billing ON projects(billing_account_id);
CREATE INDEX idx_projects_active ON projects(active);
```

### **project_products** (produtos habilitados por project) - OPCIONAL
```sql
CREATE TABLE project_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    product_type VARCHAR(50) NOT NULL, -- 'crm', 'bi', 'automation'
    enabled BOOLEAN DEFAULT true,
    settings JSONB,
    enabled_at TIMESTAMP,
    disabled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(project_id, product_type)
);

CREATE INDEX idx_project_products_project ON project_products(project_id);
CREATE INDEX idx_project_products_enabled ON project_products(project_id, enabled);
```

### **Todas as entidades CRM** (scoped por tenant_id)
```sql
-- Contacts, Messages, Sessions, Channels, Pipelines, etc.
-- Todas têm tenant_id (RLS)

CREATE TABLE contacts (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL, -- FK para projects.tenant_id
    name VARCHAR(255),
    phone VARCHAR(50),
    ...
);

-- Políticas RLS (Row-Level Security)
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;

CREATE POLICY contacts_tenant_isolation ON contacts
    USING (tenant_id = current_setting('app.current_tenant_id')::text);
```

---

## 🔗 Relacionamentos

### **User → Projects (1:N)**
- Um User pode criar múltiplos Projects
- Cada Project pertence a um único User (creator)

```go
type UserEntity struct {
    ID       uuid.UUID
    Projects []ProjectEntity `gorm:"foreignKey:UserID"`
}
```

### **User → BillingAccounts (1:N)**
- Um User pode ter múltiplas BillingAccounts
- Cada BillingAccount pertence a um único User (owner)

```go
type UserEntity struct {
    ID              uuid.UUID
    BillingAccounts []BillingAccountEntity `gorm:"foreignKey:UserID"`
}
```

### **BillingAccount → Projects (1:N)**
- **Uma BillingAccount pode pagar por VÁRIOS Projects** ✅
- **Um Project é pago por UMA BillingAccount** ✅

```go
type BillingAccountEntity struct {
    ID       uuid.UUID
    UserID   uuid.UUID
    Projects []ProjectEntity `gorm:"foreignKey:BillingAccountID"`
}

type ProjectEntity struct {
    ID               uuid.UUID
    UserID           uuid.UUID       // Quem criou
    BillingAccountID uuid.UUID       // Quem paga
    TenantID         string          // RLS
}
```

### **Project → Products (1:N)**
Dois cenários:

**Opção 1: Configuration JSONB**
```json
{
    "enabled_products": {
        "crm": {"enabled": true, "limits": {"contacts": 10000}},
        "bi": {"enabled": false},
        "automation": {"enabled": false}
    }
}
```

**Opção 2: Tabela project_products**
```go
type ProjectEntity struct {
    ID               uuid.UUID
    Products         []ProjectProductEntity `gorm:"foreignKey:ProjectID"`
}

type ProjectProductEntity struct {
    ID          uuid.UUID
    ProjectID   uuid.UUID
    ProductType string  // 'crm', 'bi', 'automation'
    Enabled     bool
}
```

---

## 🎯 Products

### **Product Types**
1. **CRM** - Customer Relationship Management
   - Contacts, Messages, Sessions
   - Channels (WhatsApp, Email, SMS)
   - Pipelines, Automations, Agents
   - Notes, Tags, Custom Fields

2. **BI** (futuro) - Business Intelligence
   - Dashboards, Reports
   - Analytics, Metrics
   - Data Export

3. **Automation** (futuro) - Integration Platform (Zapier-like)
   - Workflows, Triggers, Actions
   - API Integrations
   - Cross-product automation

### **Product Scope**
- **CRM Automations** (já existe): Automações internas do CRM
  - `internal/domain/pipeline/automation.go`
  - Triggers: `session.ended`, `message.received`
  - Actions: `change_status`, `assign_agent`, `send_message`

- **Automation Product** (futuro): Automações cross-product e APIs externas
  - Workflows visuais
  - Integrações com Salesforce, Mailchimp, Slack...
  - Orquestração entre CRM, BI e APIs externas

---

## 🔐 Multi-Tenancy (RLS)

Todas as entidades CRM usam **tenant_id** para isolamento:

```sql
-- Habilitar RLS
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;

-- Política de isolamento
CREATE POLICY tenant_isolation ON contacts
    USING (tenant_id = current_setting('app.current_tenant_id')::text);
```

**No código:**
```go
// Antes de cada query, setar o tenant_id
db.Exec("SET app.current_tenant_id = ?", tenantID)

// Todas as queries são automaticamente filtradas
contacts, _ := db.Find(&ContactEntity{}) // Só retorna contacts do tenant
```

---

## 💰 Billing (Stripe Integration)

### **Modelo de Cobrança**

**1. Por Project:**
```
BillingAccount 1 → paga por:
├── Project 1 ($99/mês)
└── Project 2 ($49/mês)
Total: $148/mês
```

**2. Por Product (futuro):**
```
Project 1:
├── CRM (enabled): $99/mês
├── BI (disabled): $0
└── Automation (disabled): $0

Project 2:
├── CRM (enabled): $49/mês
└── BI (enabled): $29/mês
Total Project 2: $78/mês
```

### **Stripe Setup**
```go
// Criar Stripe customer quando criar BillingAccount
customer, _ := stripe.Customer.New(&stripe.CustomerParams{
    Email: billingAccount.BillingEmail,
    Name:  billingAccount.Name,
})

billingAccount.StripeCustomerID = customer.ID

// Criar subscription por project
subscription, _ := stripe.Subscription.New(&stripe.SubscriptionParams{
    Customer: billingAccount.StripeCustomerID,
    Items: []*stripe.SubscriptionItemsParams{
        {Price: "price_crm_99"},
    },
})
```

---

## 📊 Exemplos de Uso

### **1. User cria novo Project**
```go
// 1. User se cadastra
user, _ := user.NewUser("João Silva", "joao@email.com", "senha123")
userRepo.Save(ctx, user)

// 2. Criar BillingAccount (Stripe)
billingAccount, _ := billing.NewBillingAccount(user.ID(), "joao@email.com")
// Integração com Stripe
stripeCustomer := createStripeCustomer(billingAccount)
billingAccount.StripeCustomerID = stripeCustomer.ID
billingRepo.Save(ctx, billingAccount)

// 3. Criar Project
project, _ := project.NewProject(
    user.ID(),                  // criador
    billingAccount.ID(),        // pagador
    "joao-loja",                // tenant_id
    "Minha Loja",               // nome
)

// 4. Habilitar CRM
project.UpdateConfiguration(map[string]interface{}{
    "enabled_products": map[string]interface{}{
        "crm": map[string]interface{}{
            "enabled": true,
            "features": []string{"contacts", "messages", "automations"},
            "limits": map[string]int{"contacts": 10000},
        },
    },
})

projectRepo.Save(ctx, project)
```

### **2. User cria segundo Project com mesma BillingAccount**
```go
// Reusar BillingAccount existente
project2, _ := project.NewProject(
    user.ID(),
    billingAccount.ID(),  // Mesma billing account!
    "joao-cliente-abc",
    "Cliente ABC",
)

projectRepo.Save(ctx, project2)

// Billing:
// BillingAccount 1 paga por Project 1 + Project 2
```

### **3. User habilita BI em Project**
```go
// Atualizar configuration
project.UpdateConfiguration(map[string]interface{}{
    "enabled_products": map[string]interface{}{
        "crm": {"enabled": true},
        "bi": {"enabled": true},  // ✅ Habilitado
    },
})

// Billing atualiza subscription no Stripe
stripe.Subscription.Update(subscriptionID, &stripe.SubscriptionParams{
    Items: []*stripe.SubscriptionItemsParams{
        {Price: "price_crm_99"},
        {Price: "price_bi_29"}, // Adiciona BI
    },
})
```

---

## 🚫 O que NÃO existe

### ❌ Customer
- **NÃO existe** entidade Customer
- User é o top-level, não Customer
- Não é modelo SaaS multi-customer

### ❌ Super Admin de várias contas
- Cada User gerencia seus próprios Projects
- Não existe "painel admin de todas as contas"

### ❌ Project com múltiplas BillingAccounts
- **1 Project = 1 BillingAccount** (quem paga)
- Um Project NÃO pode ter várias contas pagando

---

## ✅ O que existe

### ✅ User (top-level)
- Pessoa física que se cadastra
- Cria e gerencia Projects
- Possui BillingAccounts

### ✅ Project (workspace/tenant)
- Criado por um User
- Pago por uma BillingAccount
- Tem tenant_id (RLS)
- Habilita Products (CRM, BI, Automation)

### ✅ BillingAccount (Stripe customer)
- Pertence a um User
- Pode pagar por múltiplos Projects
- Integração com Stripe

### ✅ Multi-tenancy (RLS)
- Isolamento por tenant_id
- Row-Level Security no PostgreSQL

### ✅ CRM Automations
- Automações internas do CRM
- Já implementado em `internal/domain/pipeline/automation.go`

---

## 🗺️ Roadmap

### **Fase 1: Atual (CRM)** ✅
- [x] Users, Projects, BillingAccounts
- [x] CRM completo (Contacts, Messages, Sessions, Channels, Pipelines)
- [x] CRM Automations
- [x] Multi-tenancy (RLS)
- [ ] Stripe integration (em progresso)

### **Fase 2: Products (curto prazo)**
- [ ] Adicionar `enabled_products` ao Project configuration
- [ ] Billing por produto (Stripe subscriptions)
- [ ] UI para habilitar/desabilitar produtos

### **Fase 3: BI (médio prazo)**
- [ ] Produto BI
- [ ] Dashboards, Reports, Analytics
- [ ] Integração com CRM data

### **Fase 4: Automation Product (longo prazo)**
- [ ] Produto Automation (Zapier-like)
- [ ] Workflow builder visual
- [ ] Integrações externas (Salesforce, Mailchimp, Slack...)
- [ ] Cross-product orchestration

---

## 📚 Comparação com outros modelos

### **Ventros (User-Centric)**
```
User → Projects → Products (CRM, BI, Automation)
```
**Exemplos:** Notion, Linear, Airtable

### **SaaS Multi-Customer (Customer-Centric)**
```
Customer → Subscription → Projects
```
**Exemplos:** Salesforce, HubSpot, Zendesk

### **Single-Tenant**
```
Company → Departments → Teams
```
**Exemplos:** SAP, Oracle, sistemas enterprise

---

## 🎯 Conclusão

**Ventros é User-Centric:**
- ✅ User cria Projects (workspaces)
- ✅ BillingAccount paga por Projects
- ✅ Cada Project habilita Products (CRM, BI, Automation)
- ✅ Multi-tenancy via tenant_id (RLS)
- ❌ NÃO existe Customer (não é multi-customer SaaS)

**Modelo similar a:** Notion, Linear, Airtable (não Salesforce/HubSpot)

---

**Versão:** 2025-10-11
**Status:** ✅ Arquitetura atual documentada
