# 🧪 Testing Quick Reference - Ventros CRM

## 🏃 Comandos Separados (execute individualmente)

### Unit Tests
```bash
make test-unit
```
- ⏱️ **Tempo:** ~2 min
- 📋 **Requisitos:** Nenhum (usa mocks)
- 🎯 **Testa:** Domain + Application layers
- 💡 **Use:** Antes de cada commit

### Integration Tests
```bash
make infra              # Primeiro: sobe infraestrutura
make test-integration   # Depois: roda testes
```
- ⏱️ **Tempo:** ~10 min
- 📋 **Requisitos:** `make infra` rodando
- 🎯 **Testa:** Repositories, Message Queue, WebSocket
- 💡 **Use:** Antes de PR, testar features com DB

### E2E Tests
```bash
make infra              # Terminal 1: infraestrutura
make api                # Terminal 2: API
make test-e2e           # Terminal 3: testes E2E
```
- ⏱️ **Tempo:** ~10 min
- 📋 **Requisitos:** `make infra` + `make api` rodando
- 🎯 **Testa:** WAHA webhooks, Automations, User flows completos
- 💡 **Use:** Validar fluxos completos, smoke tests

---

## 📦 Comandos Agrupados

| Comando | O que faz | Requisitos |
|---------|-----------|------------|
| `make test` | unit + integration + e2e | infra + API |
| `make test-coverage` | Todos com HTML report | infra + API |
| `make test-coverage-unit` | Unit tests com coverage | Nenhum |
| `make test-bench` | Performance benchmarks | Nenhum |
| `make clean-coverage` | Remove coverage reports | Nenhum |

---

## 📚 Documentação Completa

| Arquivo | Tamanho | O que contém |
|---------|---------|--------------|
| **guides/TESTING.md** | 46KB | 📖 Guia completo: estratégias, padrões, mocks, exemplos detalhados |
| **tests/README.md** | 6.4KB | 📋 Estrutura de testes, fixtures, workflows E2E |
| **README.md** | 6.3KB | 🚀 Quick start, test pyramid, métricas |

---

## 🎯 Workflows Recomendados

### Durante Desenvolvimento
```bash
# Antes de cada commit
make test-unit          # ~2 min - valida lógica de negócio
```

### Antes de Pull Request
```bash
# Terminal 1
make infra

# Terminal 2
make test-integration   # ~10 min - valida integrações
```

### CI/CD Pipeline
```bash
make infra              # Setup
make api &              # Background
make test               # Todos os testes (~20 min)
```

### Coverage Report Local
```bash
make test-coverage-unit  # Gera coverage/coverage-unit.html
# Abre automaticamente no browser
```

---

## 🐛 Debug de Testes

### Rodar teste específico
```bash
# Unit test específico
go test -v ./internal/domain/crm/contact/... -run TestContactCreation

# Integration test específico
go test -v ./tests/integration/... -run TestWAHAMessageSender

# E2E test específico
go test -v ./tests/e2e/... -run TestAPITestSuite/Test1_CreateUser
```

### Ver logs detalhados
```bash
make test-unit -v          # Verbose mode
go test -v -race ./...     # Com race detector
```

### Manter dados após teste (E2E)
Comente a linha no `tests/e2e/api_test.go`:
```go
// func (s *APITestSuite) TearDownSuite() { ... }
```

---

## 📊 Test Pyramid (Mike Cohn, 2009)

```
        /\
       /E2E\      ← 5 tests (10%) - Fluxos completos
      /----\
     /Integ.\    ← 2 tests (20%) - Com infra real
    /--------\
   /   Unit   \  ← 61 tests (70%) - Lógica de negócio
  /____________\
```

**Atual:**
- ✅ Unit: 61 tests
- ⚠️ Integration: 2 tests (precisa expandir)
- ✅ E2E: 5 tests

**Target Coverage:** 82%

---

## ⚠️ Status Atual

### ✅ Funcionando
- Domain layer (core, project, saga, contact, session)
- Application layer (use cases, commands)
- Infrastructure (crypto, channels, WAHA adapters)

### 🐛 Com Erros
- `billing_account_test.go` - Precisa atualizar assinatura `ReconstructBillingAccount`
  - Adicionar parâmetros: `version int`, `stripe_customer_id string`, `name string`

---

## 💡 Dicas

1. **Execute `make test-unit` antes de cada commit** - É rápido e pega 70% dos bugs
2. **Use `make test-coverage-unit`** para ver o que não está testado
3. **Documente testes complexos** em `guides/TESTING.md`
4. **Fixtures em `tests/e2e/fixtures.go`** - Reutilize dados de teste
5. **Table-driven tests** - Padrão Go para múltiplos cenários

---

## 🔗 Links Úteis

- [Go Testing Guide](https://golang.org/pkg/testing/)
- [Testify Suite](https://github.com/stretchr/testify#suite-package)
- [Test Pyramid - Martin Fowler](https://martinfowler.com/articles/practical-test-pyramid.html)
- [Testing Best Practices - Google](https://testing.googleblog.com/)

---

**Última atualização:** 2025-10-12
**Mantido por:** Ventros Team
