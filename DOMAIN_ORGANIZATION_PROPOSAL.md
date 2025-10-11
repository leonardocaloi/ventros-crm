# 🏗️ Domain Organization - Proposal

**Date**: 2025-10-11
**Status**: ✅ Proposal for Review

---

## 🎯 Current Problem

All domain aggregates are in flat structure:
```
internal/domain/
├── agent/
├── agent_session/
├── broadcast/
├── channel/
├── contact/
├── message/
├── pipeline/
├── session/
├── tracking/
├── webhook/
├── ... (20+ folders)
```

**Problem**: Hard to know which belong to CRM, BI, Automation products.

---

## ✅ Proposed Solution: Organize by Product

### **Option 1: Subfolders by Product** ⭐ RECOMMENDED

```
internal/domain/
├── core/                    # Shared across products
│   ├── user/
│   ├── project/
│   ├── billing_account/
│   └── shared/             # Value objects, errors
│
├── crm/                    # CRM Product bounded context
│   ├── agent/
│   ├── agent_session/
│   ├── channel/
│   ├── channel_type/
│   ├── chat/
│   ├── contact/
│   ├── contact_event/
│   ├── contact_list/
│   ├── credential/
│   ├── event/
│   ├── message/
│   ├── message_enrichment/
│   ├── message_group/
│   ├── note/
│   ├── pipeline/
│   ├── session/
│   ├── tracking/
│   └── webhook/
│
├── bi/                     # BI Product (future)
│   ├── dashboard/
│   ├── report/
│   └── metric/
│
└── automation/             # Automation Product (future)
    ├── workflow/
    ├── trigger/
    └── action/
```

**Advantages:**
- ✅ Clear product boundaries
- ✅ Easy to understand what belongs where
- ✅ Prepared for multi-product
- ✅ Can compile products separately (future)

**Code changes:**
```go
// Before
import "github.com/caloi/ventros-crm/internal/domain/contact"

// After
import "github.com/caloi/ventros-crm/internal/domain/crm/contact"
```

---

### **Option 2: Tags in Comments** ❌ NOT RECOMMENDED

Keep flat structure, add comments:
```go
// Package contact - CRM Product
package contact
```

**Disadvantages:**
- ❌ No enforced boundaries
- ❌ Still confusing
- ❌ Can't compile separately

---

## 📋 Product Types Definition

Create a central definition of product types:

```go
// internal/domain/core/product/product_type.go
package product

type ProductType string

const (
	ProductTypeCRM        ProductType = "crm"
	ProductTypeBI         ProductType = "bi"
	ProductTypeAutomation ProductType = "automation"
)

// All returns all valid product types
func All() []ProductType {
	return []ProductType{
		ProductTypeCRM,
		ProductTypeBI,
		ProductTypeAutomation,
	}
}

// IsValid checks if product type is valid
func (p ProductType) IsValid() bool {
	switch p {
	case ProductTypeCRM, ProductTypeBI, ProductTypeAutomation:
		return true
	default:
		return false
	}
}

// String returns string representation
func (p ProductType) String() string {
	return string(p)
}
```

---

## 🗺️ Migration Plan

### Phase 1: Create Structure (No Breaking Changes)

```bash
# 1. Create product folders
mkdir -p internal/domain/core
mkdir -p internal/domain/crm
mkdir -p internal/domain/bi
mkdir -p internal/domain/automation

# 2. Move shared to core
mv internal/domain/shared internal/domain/core/shared
mv internal/domain/user internal/domain/core/user
mv internal/domain/project internal/domain/core/project
mv internal/domain/billing_account internal/domain/core/billing_account

# 3. Move CRM aggregates
mv internal/domain/agent internal/domain/crm/agent
mv internal/domain/agent_session internal/domain/crm/agent_session
mv internal/domain/channel internal/domain/crm/channel
mv internal/domain/contact internal/domain/crm/contact
mv internal/domain/message internal/domain/crm/message
mv internal/domain/pipeline internal/domain/crm/pipeline
mv internal/domain/session internal/domain/crm/session
mv internal/domain/tracking internal/domain/crm/tracking
mv internal/domain/webhook internal/domain/crm/webhook
mv internal/domain/note internal/domain/crm/note
mv internal/domain/chat internal/domain/crm/chat
mv internal/domain/contact_event internal/domain/crm/contact_event
mv internal/domain/contact_list internal/domain/crm/contact_list
mv internal/domain/credential internal/domain/crm/credential
mv internal/domain/event internal/domain/crm/event
mv internal/domain/message_enrichment internal/domain/crm/message_enrichment
mv internal/domain/message_group internal/domain/crm/message_group
mv internal/domain/channel_type internal/domain/crm/channel_type

# 4. Remove customer (confirmed removed)
rm -rf internal/domain/customer
```

### Phase 2: Update Imports (Automated)

```bash
# Use gofmt to update imports
find . -name "*.go" -exec sed -i 's|internal/domain/contact|internal/domain/crm/contact|g' {} \;
find . -name "*.go" -exec sed -i 's|internal/domain/message|internal/domain/crm/message|g' {} \;
# ... (repeat for all aggregates)
```

### Phase 3: Verify

```bash
# Build to check for errors
make build

# Run tests
make test
```

---

## 🎯 Final Structure

```
internal/domain/
├── core/                           # Shared across all products
│   ├── user/                      # User aggregate
│   ├── project/                   # Project aggregate
│   ├── billing_account/           # Billing aggregate
│   ├── product/                   # Product types definition
│   └── shared/                    # Value objects, errors
│       ├── custom_field.go
│       ├── domain_event.go
│       ├── errors.go
│       ├── hex_color.go
│       ├── money.go
│       └── tenant_id.go
│
├── crm/                           # ✅ CRM Product (18 aggregates)
│   ├── agent/
│   ├── agent_session/
│   ├── channel/
│   ├── channel_type/
│   ├── chat/
│   ├── contact/
│   ├── contact_event/
│   ├── contact_list/
│   ├── credential/
│   ├── event/
│   ├── message/
│   ├── message_enrichment/
│   ├── message_group/
│   ├── note/
│   ├── pipeline/
│   ├── session/
│   ├── tracking/
│   └── webhook/
│
├── bi/                            # 📊 BI Product (future)
│   └── .gitkeep
│
└── automation/                    # ⚙️ Automation Product (future)
    └── .gitkeep
```

---

## ✅ Benefits

### **1. Clear Boundaries**
- Easy to see what belongs to CRM
- Prepared for BI and Automation products
- New developers understand structure immediately

### **2. Scalability**
- Can compile products separately
- Can deploy products separately (microservices future)
- Can have different teams for different products

### **3. Multi-tenancy per Product**
```go
// Project can enable products
type Project struct {
    EnabledProducts []product.ProductType
    // CRM: contacts, messages, sessions...
    // BI: dashboards, reports...
    // Automation: workflows, triggers...
}
```

### **4. Billing per Product**
```sql
-- Project products (from ARCHITECTURE_USER_CENTRIC.md)
CREATE TABLE project_products (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    product_type VARCHAR(50), -- 'crm', 'bi', 'automation'
    enabled BOOLEAN,
    settings JSONB,
    UNIQUE(project_id, product_type)
);
```

---

## 🚫 What We're Removing

### ❌ Customer Aggregate
**Status**: CONFIRMED REMOVED

**Reason**: User-centric model, not customer-centric (see ARCHITECTURE_USER_CENTRIC.md)

**Hierarchy:**
```
User (person)
└── Projects (workspaces)
    └── Products (CRM, BI, Automation)
```

NOT:
```
Customer (company)  ← REMOVED
└── ...
```

---

## 📊 Summary

| Aspect | Current | Proposed |
|--------|---------|----------|
| **Structure** | Flat (20+ folders) | Organized by product |
| **CRM Aggregates** | Mixed with core | `internal/domain/crm/` |
| **Product Types** | Not defined | Enum in `core/product/` |
| **Customer** | Exists but unused | ✅ REMOVED |
| **Scalability** | Hard to add products | Easy (new folder) |
| **Understanding** | Confusing | Clear |

---

## 🎯 Decision Needed

**Do you want to proceed with Option 1 (subfolders)?**

If YES:
1. I can create the structure
2. Move files
3. Update imports
4. Verify build/tests

**Estimated time**: 2-3 hours

**Risk**: Medium (import changes, but automated)

**Benefit**: High (clarity, scalability, multi-product ready)

---

**Your call!** 🚀
