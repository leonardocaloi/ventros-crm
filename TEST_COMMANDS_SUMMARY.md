# 🧪 Comandos de Teste - Ventros CRM

## 📋 Comandos Disponíveis

### **1. Testes Unitários** (Rápido - ~2 min)
```bash
make test-unit
```
**O que testa:**
- Domain layer (agregados, value objects, eventos)
- Application layer (use cases)
- Infrastructure utilities (crypto, channels)

**Requisitos:** ✅ Nenhum (usa mocks)

---

### **2. Testes de Integração** (Médio - ~10 min)
```bash
make infra              # Primeiro: sobe infraestrutura
make test-integration   # Depois: roda testes
```
**O que testa:**
- Repositórios com banco real
- Message queue workflows
- WebSocket connections

**Requisitos:** ⚠️ Infraestrutura rodando (`make infra`)

---

### **3. Testes E2E** (Lento - ~10 min)
```bash
make infra              # Terminal 1: infraestrutura
make api                # Terminal 2: API
make test-e2e           # Terminal 3: testes
```
**O que testa:**
- WAHA webhook processing
- Scheduled automation workers
- Fluxos completos de usuário

**Requisitos:** ⚠️ API + Infraestrutura rodando

---

### **4. Todos os Testes**
```bash
make test
```
Roda: unit + integration + e2e (requer infra + API)

---

### **5. Coverage Reports**
```bash
make test-coverage         # Todos os testes com coverage HTML
make test-coverage-unit    # Apenas unit tests com coverage
make clean-coverage        # Remove relatórios
```
**Output:** `coverage/coverage.html` (abre no browser)

---

### **6. Benchmark Tests**
```bash
make test-bench
```
Testa performance de operações críticas.

---

## 📚 Documentação Completa

| Arquivo | Tamanho | Conteúdo |
|---------|---------|----------|
| **guides/TESTING.md** | 46KB | 📖 **Guia COMPLETO** - Estratégias, padrões, exemplos |
| **tests/README.md** | 6.4KB | 📋 Estrutura de testes, fixtures, E2E |
| **README.md** | 6.3KB | 🚀 Quick start + test pyramid |

---

## 🎯 Quando usar cada comando

| Situação | Comando | Tempo |
|----------|---------|-------|
| 🏃 **Dev rápido** (antes de commit) | `make test-unit` | ~2 min |
| 🔍 **Teste feature** (com DB) | `make test-integration` | ~10 min |
| ✅ **Antes de PR** | `make test` | ~20 min |
| 📊 **Coverage report** | `make test-coverage` | ~5 min |
| 🚀 **CI/CD pipeline** | `make test` | ~20 min |

---

## 🐛 Status Atual dos Testes

✅ **Funcionando:**
- Domain layer (core, project, saga)
- Application layer (use cases)
- Infrastructure (crypto, channels)

⚠️ **Com Erros:**
- `billing_account_test.go` - Precisa atualizar assinatura da função `ReconstructBillingAccount`

---

## 📖 Leia Mais

- **Guia completo:** `guides/TESTING.md`
- **Test Pyramid:** `README.md` (seção Testing)
- **E2E Details:** `tests/README.md`

---

**💡 Dica:** Execute `make test-unit` antes de cada commit!
