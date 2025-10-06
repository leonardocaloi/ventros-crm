# 📋 TASKS - Ventros CRM v0.1.0

Arquivo de tarefas organizadas para preparação do projeto para lançamento e refatorações futuras.

**Total de grupos**: 2  
**Última atualização**: 2025-10-06  
**Versão atual**: 0.0.1-dev → **Próxima**: 0.1.0

---


prepara base de testes para preparar pra tudo isso q vai entrar antes


## 📊 RESUMO EXECUTIVO

| Grupo | Nome | Tarefas | Prioridade | Tempo Estimado |
|-------|------|---------|------------|----------------|
| 1️⃣ | **Release 0.1.0 - Organização** | 16 | 🔴 ALTA | 1-2 semanas |
| 2️⃣ | **Refatoração Arquitetural** | 19 | 🟡 MÉDIA | 6-8 semanas |

---

## 🎯 GRUPO 1: Organização para Release 0.1.0

> **Objetivo**: Deixar o projeto elegante, profissional e pronto para publicação no GitHub.  
> **Prazo**: 1-2 semanas  
> **Prioridade**: 🔴 ALTA

### ✅ Documentação Raiz (Completo!)

- [x] **README.md** - Overview profissional com badges, estrutura, quick start
- [x] **CHANGELOG.md** - Histórico de versões (Keep a Changelog format)
- [x] **CONTRIBUTING.md** - Guidelines para contribuidores
- [x] **LICENSE** - MIT License
- [x] **.gitignore** - Melhorado com seções organizadas

### 📁 Estrutura de Arquivos

- [ ] **Remover pasta `/examples`** ⚠️
  - Não está sendo importada/usada no código
  - Mover conteúdo para `/guides/code-examples/rbac_example.go`
  - Atualizar .gitignore para ignorar `/examples/`
  
- [ ] **Remover binários da raiz** ⚠️
  - Deletar `api` (69MB)
  - Deletar `ventros-crm` (69MB)
  - Já estão no .gitignore, mas limpar antes do commit inicial
  
- [ ] **Criar pasta `/guides`** para documentação educacional
  ```
  /guides/
  ├── architecture/
  │   ├── README.md (DDD, Event-Driven, SAGA)
  │   ├── diagrams/
  │   └── decisions/ (ADRs - Architecture Decision Records)
  ├── getting-started/
  │   ├── README.md (Setup local)
  │   ├── quickstart.md
  │   └── troubleshooting.md
  ├── deployment/
  │   ├── README.md (Kubernetes, Helm)
  │   ├── docker.md
  │   └── production-checklist.md
  ├── code-examples/
  │   ├── README.md
  │   ├── rbac_example.go (mover de /examples)
  │   ├── use_case_example.go
  │   └── event_handler_example.go
  └── api/
      ├── README.md (REST API docs)
      ├── authentication.md
      └── webhooks.md
  ```

- [ ] **Adicionar .editorconfig** para consistência
  ```ini
  root = true
  
  [*]
  charset = utf-8
  end_of_line = lf
  insert_final_newline = true
  trim_trailing_whitespace = true
  
  [*.go]
  indent_style = tab
  indent_size = 4
  
  [*.{yml,yaml,json}]
  indent_style = space
  indent_size = 2
  
  [Makefile]
  indent_style = tab
  ```

### 🔧 Código e Configuração

- [ ] **Adicionar versão no código**
  - Criar `internal/version/version.go`
  ```go
  package version
  
  const Version = "0.1.0"
  const BuildDate = "2025-10-06"
  ```
  - Expor em `/health` endpoint
  
- [ ] **Validar e limpar Makefile**
  - Remover comandos não usados
  - Adicionar `make setup` para setup inicial completo
  - Adicionar `make pre-commit` (test + lint)
  - Documentar todos os comandos no `make help`

- [ ] **Script de setup inicial**
  - Criar `scripts/setup.sh`
  - Verificar dependências (Go, Docker, PostgreSQL client)
  - Setup de .env
  - Iniciar infraestrutura
  - Rodar migrations

- [ ] **Validar variáveis de ambiente**
  - Revisar `.env.example`
  - Documentar todas as variáveis obrigatórias
  - Adicionar valores sensatos de default

### 📝 Documentação Adicional

- [ ] **Criar ARCHITECTURE.md** na raiz
  - Diagrama de arquitetura (DDD layers)
  - Explicação de Domain, Application, Infrastructure
  - Event flow diagram
  - Decisões arquiteturais chave

- [ ] **Documentar API**
  - Exportar Postman Collection
  - Adicionar em `/docs/postman/`
  - Criar exemplos de chamadas curl

- [ ] **Badges no README**
  - Build Status (quando configurar CI)
  - Go Report Card
  - Coverage (quando configurar)
  - License

### 🧹 Limpeza de Código

- [ ] **Remover TODOs críticos** (47 encontrados)
  - Priorizar handlers que acessam repositórios diretamente
  - Remover comentários de debug
  - Completar ou remover código comentado

- [ ] **Configurar golangci-lint**
  - Criar `.golangci.yml`
  - Configurar linters: gofmt, govet, errcheck, staticcheck, etc
  - Rodar e corrigir warnings

---

## 🏗️ GRUPO 2: Refatoração Arquitetural

> **Objetivo**: Elevar arquitetura para nível enterprise com CQRS, Unit of Work, e SAGA completo.  
> **Prazo**: 6-8 semanas (pós 0.1.0)  
> **Prioridade**: 🟡 MÉDIA

### 🎯 Implementar CQRS Completo

#### Commands (Write Side)
- [ ] Criar `/internal/application/commands/contact/`
  - `create_contact_command.go`
  - `create_contact_handler.go`
  - `update_contact_command.go`
  - `update_contact_handler.go`
  - `delete_contact_command.go`
  - `delete_contact_handler.go`

- [ ] Criar `/internal/application/commands/session/`
  - `create_session_command.go`
  - `create_session_handler.go`
  - `close_session_command.go`
  - `close_session_handler.go`

- [ ] Criar `/internal/application/commands/message/`
  - `send_message_command.go`
  - `send_message_handler.go`

#### Queries (Read Side)
- [ ] Criar `/internal/application/queries/contact/`
  - `get_contact_query.go`
  - `get_contact_handler.go`
  - `list_contacts_query.go`
  - `list_contacts_handler.go`

- [ ] Criar `/internal/application/queries/session/`
  - `get_session_query.go`
  - `get_session_handler.go`
  - `list_sessions_query.go`

#### Assemblers
- [ ] Criar `/internal/application/assemblers/`
  - `contact_assembler.go` - Domain → DTO
  - `session_assembler.go`
  - `message_assembler.go`
  - `contact_assembler_test.go`

### 🔄 Refatorar Handlers

- [ ] **ContactHandler** - usar Command/Query handlers
  - Remover lógica de negócio
  - Apenas validação e delegação
  - Usar assemblers para conversão
  
- [ ] **SessionHandler** - usar Command/Query handlers
  
- [ ] **MessageHandler** - usar Command/Query handlers

- [ ] **Remover DTOs duplicados**
  - Deletar `/infrastructure/http/dto/requests.go`
  - Remover structs inline dos handlers
  - Centralizar em `/internal/application/dtos/`

### 🔐 Unit of Work Pattern

- [ ] **Criar** `infrastructure/persistence/unit_of_work.go`
  ```go
  type UnitOfWork struct {
      db     *gorm.DB
      tx     *gorm.DB
      events []shared.DomainEvent
  }
  
  func (uow *UnitOfWork) Execute(ctx context.Context, fn func(*UnitOfWork) error) error
  func (uow *UnitOfWork) ContactRepo() contact.Repository
  func (uow *UnitOfWork) SessionRepo() session.Repository
  func (uow *UnitOfWork) RegisterEvents(events ...shared.DomainEvent)
  func (uow *UnitOfWork) publishEvents(ctx context.Context) error
  ```

- [ ] **Integrar em Use Cases**
  - Refatorar `ProcessInboundMessageUseCase`
  - Refatorar `CreateContactHandler`
  - Garantir transações atômicas

- [ ] **Testes de integração**
  - Cenários de rollback
  - Eventos publicados apenas após commit

### 🔄 SAGA com Compensação (Temporal)

- [ ] **Criar** `/internal/application/sagas/process_message_saga.go`
  ```go
  func ProcessMessageSaga(ctx workflow.Context, cmd Command) error
  ```

- [ ] **Implementar Activities de Compensação**
  - `FindOrCreateContactActivity` + `CompensateContactActivity`
  - `FindOrCreateSessionActivity` + `CompensateSessionActivity`
  - `CreateMessageActivity` + `CompensateMessageActivity`

- [ ] **Testar cenários de falha**
  - Falha após criar Contact
  - Falha após criar Session
  - Verificar rollback automático

### 📢 Event Subscribers

- [ ] **Criar** `/infrastructure/messaging/subscribers/`
  - `contact_event_subscriber.go`
  - `session_event_subscriber.go`
  - `message_event_subscriber.go`

- [ ] **Separar processamento assíncrono**
  - Mover lógica de event handlers
  - Usar workers dedicados
  - Garantir idempotência

### 🧪 Testes

#### Testes Unitários
- [ ] **Domain Models** (Contact, Session, Message)
  - Testar invariantes
  - Testar value objects
  - Testar domain events

#### Testes de Integração
- [ ] **Repositories**
  - CRUD completo
  - Queries complexas
  - Transações

#### Testes E2E
- [ ] **Fluxos principais**
  - Criação de contato via API
  - Processamento de mensagem completo
  - Lifecycle de sessão

#### Coverage
- [ ] **Configurar coverage**
  - Adicionar ao CI
  - Meta: mínimo 70%
  - Badge no README

---

## 📈 MÉTRICAS DE PROGRESSO

### Grupo 1 - Release 0.1.0
```
Progress: ▓▓▓▓▓░░░░░░░░░░░░░░░ 31% (5/16)
```
- ✅ **Completo**: 5 (Documentação raiz)
- 🔄 **Em progresso**: 0
- ⏳ **Pendente**: 11

### Grupo 2 - Refatoração
```
Progress: ░░░░░░░░░░░░░░░░░░░░ 0% (0/19)
```
- ✅ **Completo**: 0
- 🔄 **Em progresso**: 0
- ⏳ **Pendente**: 19

---

## 🎯 CRITÉRIOS DE ACEITAÇÃO

### ✅ Release 0.1.0 Pronta Quando:

#### Estrutura
- [ ] README.md profissional e completo
- [ ] CHANGELOG.md preenchido
- [ ] Sem binários na raiz
- [ ] Pasta `/examples` removida e conteúdo movido
- [ ] Pasta `/guides` criada com documentação

#### Código
- [ ] Versão definida no código
- [ ] Sem TODOs em código crítico
- [ ] golangci-lint passando sem warnings
- [ ] Testes básicos rodando

#### Configuração
- [ ] .gitignore completo
- [ ] .editorconfig criado
- [ ] Makefile documentado
- [ ] Script de setup funcionando

#### Documentação
- [ ] Variáveis de ambiente documentadas
- [ ] ARCHITECTURE.md criado
- [ ] Postman collection exportada
- [ ] Badges no README

---

## 🔄 WORKFLOW RECOMENDADO

### Para Release 0.1.0:
```bash
# 1. Criar branch
git checkout -b feature/prepare-0.1.0

# 2. Executar tarefas do Grupo 1
# ...

# 3. Commit atômicos
git add .
git commit -m "docs: add professional README.md"
git commit -m "chore: improve .gitignore with all patterns"
git commit -m "chore: remove examples folder and move to guides"

# 4. PR para main
git push origin feature/prepare-0.1.0
# Open PR on GitHub

# 5. Após merge, criar tag
git checkout main
git pull
git tag -a v0.1.0 -m "Release 0.1.0 - Foundation"
git push origin v0.1.0

# 6. Criar GitHub Release
# - Copiar seção do CHANGELOG.md
# - Anexar binários compilados (opcional)
```

### Para Refatoração (Grupo 2):
- Dividir em múltiplas releases (0.2.0, 0.3.0, etc)
- Cada feature em branch separada
- PRs menores e mais focados
- Garantir testes antes de merge

---

## 📝 NOTAS

### Decisões Importantes:
1. **Pasta /examples** será removida porque não é importada no código
2. **CQRS** será implementado incrementalmente (0.2.0, 0.3.0)
3. **SAGA com compensação** é prioridade alta pós-0.1.0
4. **Unit of Work** deve ser implementado antes do SAGA completo

### Dependências entre Tarefas:
- CQRS → Unit of Work → SAGA (nessa ordem)
- Handlers refatorados dependem de Commands/Queries prontos
- Testes E2E dependem de fluxos completos implementados

### Compatibilidade:
- Manter APIs atuais funcionando durante refatoração
- Deprecar endpoints antigos gradualmente
- Documentar breaking changes no CHANGELOG

---

## 🎓 REFERÊNCIAS

- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

---

**Última atualização**: 2025-10-06  
**Versão atual**: 0.0.1-dev  
**Próxima release**: 0.1.0  
**Responsável**: Carlos Loi (@caloi)
