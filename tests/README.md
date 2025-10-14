# 🧪 Testing - Ventros CRM

Este diretório contém todos os testes da aplicação, seguindo a **Test Pyramid** (Mike Cohn, 2009).

## 📋 Estrutura de Testes

```
tests/
├── integration/                           # Layer 2: Integration Tests
│   ├── waha_message_sender_test.go        # WAHA adapter tests
│   └── websocket_integration_test.go      # WebSocket tests
│
└── e2e/                                   # Layer 3: E2E Tests
    ├── api_test.go                        # General API tests
    ├── waha_webhook_test.go               # WAHA webhook flow
    ├── scheduled_automation_test.go       # Automation worker
    ├── scheduled_automation_webhook_test.go
    ├── message_send_test.go
    ├── fixtures.go                        # Test fixtures
    └── README.md                          # E2E documentation

Note: Unit tests (Layer 1) are located in internal/**/*_test.go
```

## 🎯 Test Layers

### Layer 1: Unit Tests (70% - Base)
**Location**: `internal/**/*_test.go`
- **Count**: 61 test files
- **Speed**: Fast (~2 minutes)
- **Dependencies**: None (uses mocks)

**Examples**:
- `internal/domain/crm/contact/contact_test.go` - Domain logic
- `internal/application/contact/create_contact_test.go` - Use cases
- `infrastructure/crypto/aes_encryptor_test.go` - Utilities

**Run**:
```bash
make test-unit
```

### Layer 2: Integration Tests (20% - Middle)
**Location**: `tests/integration/`
- **Count**: 2 test files ⚠️ NEEDS EXPANSION
- **Speed**: Medium (~10 minutes)
- **Dependencies**: PostgreSQL, RabbitMQ, Redis, Temporal (via testcontainers)

**Current tests**:
- Repository operations with real database
- Message queue workflows
- WebSocket connections

**Run**:
```bash
make infra              # Start infrastructure first
make test-integration
```

### Layer 3: E2E Tests (10% - Top)
**Location**: `tests/e2e/`
- **Count**: 5 test files
- **Speed**: Slow (~10 minutes)
- **Dependencies**: Full running system (API + Infrastructure)

**Current tests**:
- WAHA webhook processing (8 message types)
- Scheduled automation workers
- Complete user workflows

**Run**:
```bash
make infra              # Terminal 1
make api                # Terminal 2
make test-e2e           # Terminal 3
```

## 🚀 Como Executar

### Opção 1: Testes em Go (Recomendado)

```bash
# Terminal 1: Inicia a API
make run

# Terminal 2: Executa os testes
make test-e2e
```

### Opção 2: Script Bash (Rápido)

```bash
# Terminal 1: Inicia a API
make run

# Terminal 2: Executa o script
make test-e2e-script
```

### Limpeza Após Testes

```bash
# Limpa dados de teste
make test-e2e-cleanup
```

## 📝 O que os Testes Validam

### 1. **Criação de Usuário** ✓
- Cria usuário com email/senha
- **Valida criação automática de:**
  - ✓ Projeto padrão (`default_project_id`)
  - ✓ Pipeline padrão (`default_pipeline_id`)
  - ✓ API Key para autenticação
- Endpoint: `POST /api/v1/auth/register`

### 2. **Criação de Canal** ✓
- Cria canal WAHA (WhatsApp)
- **Valida:**
  - ✓ Canal associado ao projeto
  - ✓ Configuração WAHA aplicada
  - ✓ Canal como ponto de entrada do pipeline
- Endpoint: `POST /api/v1/channels?project_id=X`

### 3. **Ativação de Canal** ✓
- Ativa canal criado
- **Valida:**
  - ✓ Canal pronto para receber mensagens
  - ✓ Webhook configurado
- Endpoint: `POST /api/v1/channels/:id/activate`

### 4. **Criação de Contato** ✓
- Cria contato no sistema
- **Valida:**
  - ✓ Contato associado ao projeto
  - ✓ Dados de contato salvos
- Endpoint: `POST /api/v1/contacts?project_id=X`

### 5. **Listagem de Recursos** ✓
- Lista contatos e canais
- **Valida:**
  - ✓ Filtros por projeto funcionam
  - ✓ Autenticação por API Key
- Endpoints: `GET /api/v1/contacts`, `GET /api/v1/channels`

## 🔄 Fluxo de Dados

```
1. Criar Usuário
   └─→ Gera automaticamente:
       ├─→ Projeto padrão
       ├─→ Pipeline padrão
       └─→ API Key

2. Criar Canal (ponto de entrada)
   └─→ Associado ao projeto
       └─→ Conectado ao pipeline

3. Mensagens chegam via Canal
   └─→ Processadas pelo pipeline
       └─→ Criam/atualizam contatos
```

## 🧹 Cleanup Automático

Os testes incluem **cleanup automático** que:

1. Deleta canais criados
2. Deleta contatos criados
3. Remove dados temporários

**Importante:** O cleanup é executado mesmo se os testes falharem (via `TearDownSuite`).

## 📦 Dependências

```bash
# Instalar dependência testify
go get github.com/stretchr/testify/suite
go get github.com/stretchr/testify/assert
```

Ou simplesmente:
```bash
make deps
```

## 🎯 Fixtures (Dados de Teste)

Os dados de entrada estão em `fixtures.go` e podem ser facilmente modificados:

```go
// tests/e2e/fixtures.go
func GetDefaultFixtures() *TestFixtures {
    return &TestFixtures{
        Users: []UserFixture{
            {
                Name:     "Admin Teste",
                Email:    "admin.teste@ventros.local",  // ← Modificar aqui
                Password: "senha_teste_123",
                Role:     "admin",
            },
        },
        // ...
    }
}
```

## 🔍 Debug de Testes

### Ver logs detalhados:
```bash
go test -v ./tests/e2e/... -timeout 5m
```

### Executar teste específico:
```bash
go test -v ./tests/e2e/... -run TestAPITestSuite/Test1_CreateUser
```

### Manter dados após teste (sem cleanup):
Comente a linha no `api_test.go`:
```go
// func (s *APITestSuite) TearDownSuite() { ... }
```

## 📊 Relatórios

### Gerar relatório de cobertura:
```bash
go test ./tests/e2e/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🏗️ Padrão de Testes (Indústria)

Este projeto segue o padrão **Table-Driven Tests** do Go:

1. ✅ **Fixtures separados** (`fixtures.go`)
2. ✅ **Suite pattern** (`testify/suite`)
3. ✅ **Setup/Teardown automático**
4. ✅ **Rastreamento de IDs criados**
5. ✅ **Cleanup garantido** (mesmo em falhas)
6. ✅ **Testes isolados e ordenados**
7. ✅ **HTTP client reutilizável**

## 🔐 Segurança

- ✅ Emails de teste usam domínio `.local`
- ✅ Senhas simples apenas para testes
- ✅ API Keys temporárias
- ✅ Dados removidos após testes

## 📚 Referências

- [Go Testing](https://golang.org/pkg/testing/)
- [Testify Suite](https://github.com/stretchr/testify#suite-package)
- [E2E Testing Best Practices](https://martinfowler.com/articles/practical-test-pyramid.html)

---

**💡 Dica:** Execute `make test-e2e-script` para uma verificação rápida durante o desenvolvimento!
