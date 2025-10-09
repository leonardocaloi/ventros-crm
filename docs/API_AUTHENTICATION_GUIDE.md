# 🔐 Guia de Autenticação da API

## 📋 Visão Geral

Este documento descreve como a autenticação funciona em **todas as rotas da API** e como o frontend está configurado para usá-las.

## 🛡️ Status de Autenticação das Rotas

### ✅ Rotas Públicas (SEM autenticação)

```
POST   /api/v1/auth/register          # Criar nova conta
POST   /api/v1/auth/login             # Login
GET    /api/v1/auth/info              # Info de autenticação (dev)
GET    /health                        # Health check geral
GET    /ready                         # Readiness probe
GET    /live                          # Liveness probe
GET    /health/database               # Check database
GET    /health/migrations             # Check migrations
GET    /health/redis                  # Check Redis
GET    /health/rabbitmq               # Check RabbitMQ
GET    /health/temporal               # Check Temporal
GET    /swagger/*                     # Documentação Swagger
POST   /api/v1/webhooks/waha/:session # Webhook WAHA (recebe do WhatsApp)
GET    /api/v1/webhooks/waha          # Info webhook WAHA
```

### 🔒 Rotas Protegidas (REQUEREM autenticação)

#### Auth (Perfil e API Keys)
```
GET    /api/v1/auth/profile           # Buscar perfil do usuário
POST   /api/v1/auth/api-key           # Gerar nova API key
```

#### Channels
```
GET    /api/v1/channels               # Listar canais
POST   /api/v1/channels               # Criar canal
GET    /api/v1/channels/:id           # Buscar canal
POST   /api/v1/channels/:id/activate  # Ativar canal
POST   /api/v1/channels/:id/deactivate # Desativar canal
DELETE /api/v1/channels/:id           # Deletar canal
GET    /api/v1/channels/:id/webhook-url # URL do webhook
POST   /api/v1/channels/:id/configure-webhook # Configurar webhook
GET    /api/v1/channels/:id/webhook-info # Info do webhook
POST   /api/v1/channels/:id/activate-waha # Ativar canal WAHA
POST   /api/v1/channels/:id/import-history # Importar histórico WAHA
GET    /api/v1/channels/:id/sessions  # Listar sessões do canal
GET    /api/v1/channels/:id/sessions/:session_id # Buscar sessão
```

#### Contacts
```
GET    /api/v1/contacts               # Listar contatos
POST   /api/v1/contacts               # Criar contato
GET    /api/v1/contacts/:id           # Buscar contato
PUT    /api/v1/contacts/:id           # Atualizar contato
DELETE /api/v1/contacts/:id           # Deletar contato
GET    /api/v1/contacts/:id/sessions  # Listar sessões do contato
GET    /api/v1/contacts/:id/sessions/:session_id # Buscar sessão
PUT    /api/v1/contacts/:id/pipelines/:pipeline_id/status # Mudar status
GET    /api/v1/contacts/:contact_id/trackings # Buscar trackings
```

#### Projects
```
GET    /api/v1/projects               # Listar projetos
POST   /api/v1/projects               # Criar projeto
GET    /api/v1/projects/:id           # Buscar projeto
PUT    /api/v1/projects/:id           # Atualizar projeto
DELETE /api/v1/projects/:id           # Deletar projeto
```

#### Pipelines
```
GET    /api/v1/pipelines              # Listar pipelines
POST   /api/v1/pipelines              # Criar pipeline
GET    /api/v1/pipelines/:id          # Buscar pipeline
POST   /api/v1/pipelines/:id/statuses # Criar status
PUT    /api/v1/pipelines/:id/contacts/:contact_id/status # Mudar status do contato
GET    /api/v1/pipelines/:id/contacts/:contact_id/status # Buscar status do contato
```

#### Sessions
```
GET    /api/v1/sessions               # Listar sessões (requer ?contact_id ou ?channel_id)
GET    /api/v1/sessions/:id           # Buscar sessão
POST   /api/v1/sessions/:id/close     # Fechar sessão
GET    /api/v1/sessions/stats         # Estatísticas de sessões
```

#### Trackings
```
GET    /api/v1/trackings/enums        # Buscar enums de tracking
POST   /api/v1/trackings              # Criar tracking
GET    /api/v1/trackings/:id          # Buscar tracking
```

#### Webhook Subscriptions
```
GET    /api/v1/webhook-subscriptions/available-events # Eventos disponíveis
POST   /api/v1/webhook-subscriptions  # Criar subscription
GET    /api/v1/webhook-subscriptions  # Listar subscriptions
GET    /api/v1/webhook-subscriptions/:id # Buscar subscription
PUT    /api/v1/webhook-subscriptions/:id # Atualizar subscription
DELETE /api/v1/webhook-subscriptions/:id # Deletar subscription
```

## 🔑 Como Funciona a Autenticação

### Backend (Go)

O middleware de autenticação (`infrastructure/http/middleware/auth.go`) aceita **3 formas**:

#### 1. Authorization Bearer (PRODUÇÃO - Recomendado)
```bash
curl http://localhost:8080/api/v1/contacts \
  -H "Authorization: Bearer <sua-api-key>"
```

#### 2. Header X-Dev-User-ID (DESENVOLVIMENTO)
```bash
curl http://localhost:8080/api/v1/contacts \
  -H "X-Dev-User-ID: 123e4567-e89b-12d3-a456-426614174000" \
  -H "X-Dev-Email: dev@example.com" \
  -H "X-Dev-Role: admin" \
  -H "X-Dev-Tenant-ID: dev-tenant"
```

#### 3. Dev Keys (DESENVOLVIMENTO)
```bash
# Admin
curl http://localhost:8080/api/v1/contacts \
  -H "Authorization: Bearer dev-admin-key"

# User
curl http://localhost:8080/api/v1/contacts \
  -H "Authorization: Bearer dev-user-key"
```

### Frontend (TypeScript)

O `api-client.ts` **automaticamente** adiciona o header `Authorization: Bearer` em todas as requisições:

```typescript
// frontend/src/lib/api-client.ts:47-62
apiClient.interceptors.request.use((config) => {
  const token = getAuthToken(); // Busca do localStorage

  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }

  return config;
});
```

## 📦 Services Implementados no Frontend

Todos os services estão em `frontend/src/services/` e **já estão configurados** para usar autenticação automaticamente:

### ✅ Services Disponíveis

| Service | Arquivo | Rotas Cobertas |
|---------|---------|----------------|
| **Auth** | `auth.service.ts` | login, register, profile, api-key |
| **Agent** | `agent.service.ts` | CRUD de agentes |
| **Channel** | `channel.service.ts` | CRUD de canais, webhook config |
| **Contact** | `contact.service.ts` | CRUD de contatos |
| **Message** | `message.service.ts` | CRUD de mensagens |
| **Pipeline** | `pipeline.service.ts` | CRUD de pipelines, status |
| **Project** | `project.service.ts` | CRUD de projetos |
| **Session** | `session.service.ts` | Listar sessões, close |
| **Tracking** | `tracking.service.ts` | CRUD de trackings |
| **Webhook** | `webhook.service.ts` | CRUD de webhook subscriptions |

### Exemplo de Uso

```typescript
import { contactService } from '@/services';

// Buscar contatos (autenticação automática via interceptor)
const contacts = await contactService.list({
  page: 1,
  page_size: 10
});

// Criar contato
const newContact = await contactService.create({
  name: 'João Silva',
  email: 'joao@example.com',
  phone: '+5511999999999'
});
```

## 🔄 Fluxo Completo de Autenticação

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │
       │ 1. Login/Register
       ▼
┌─────────────────────┐
│  Auth Service       │
│  auth.service.ts    │
└──────┬──────────────┘
       │
       │ 2. POST /api/v1/auth/login
       ▼
┌─────────────────────┐
│  Backend Go         │
│  auth_handler.go    │
└──────┬──────────────┘
       │
       │ 3. Valida credenciais
       │ 4. Retorna api_key
       ▼
┌─────────────────────┐
│  localStorage       │
│  api_key: "abc..."  │
└──────┬──────────────┘
       │
       │ 5. Todas as requisições posteriores
       ▼
┌─────────────────────┐
│  API Client         │
│  Interceptor        │
│  + Authorization    │
└──────┬──────────────┘
       │
       │ 6. Headers adicionados automaticamente
       ▼
┌─────────────────────┐
│  Backend Go         │
│  Auth Middleware    │
│  Valida api_key     │
└─────────────────────┘
```

## 🧪 Testando Autenticação

### 1. Registrar e Obter API Key

```bash
# Registrar
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "Test@123"
  }'

# Response:
{
  "message": "User created successfully",
  "user_id": "a1b2c3d4-...",
  "api_key": "abc123def456..."  # ← GUARDAR ISSO!
}
```

### 2. Usar API Key nas Requisições

```bash
# Salvar em variável
export API_KEY="abc123def456..."

# Listar contatos (autenticado)
curl http://localhost:8080/api/v1/contacts \
  -H "Authorization: Bearer $API_KEY"

# Criar canal (autenticado)
curl -X POST http://localhost:8080/api/v1/channels \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WhatsApp Principal",
    "type": "waha"
  }'
```

### 3. Testar Erro de Autenticação

```bash
# Sem Authorization (deve falhar)
curl http://localhost:8080/api/v1/contacts

# Response: 401 Unauthorized
{
  "error": "Authentication required",
  "hint": "Use X-Dev-User-ID header in dev mode or Authorization: Bearer <api_key>"
}
```

## 🔒 Segurança

### ✅ Implementado

1. **API Keys Hasheadas**: SHA-256 no banco
2. **Senhas Hasheadas**: bcrypt com custo padrão
3. **Validação Server-Side**: Todas as rotas protegidas
4. **Auto-logout em 401**: Frontend remove token automaticamente
5. **Token em localStorage**: Conveniente para desenvolvimento
6. **Bearer Token Standard**: Segue RFC 6750

### ⚠️ Para Produção

1. **HTTPS Obrigatório**: Nunca usar HTTP em produção
2. **httpOnly Cookies**: Migrar de localStorage
3. **CSRF Tokens**: Proteção contra CSRF
4. **Rate Limiting**: Limitar tentativas de autenticação
5. **API Key Rotation**: Implementar renovação periódica
6. **Audit Logs**: Registrar todos os acessos

## 🐛 Troubleshooting

### Erro: "Authentication required"

**Causa**: API key não está sendo enviada ou é inválida

**Solução**:
```typescript
// Verificar se tem API key no localStorage
console.log(localStorage.getItem('api_key'));

// Se não tiver, fazer login novamente
import { authService } from '@/services';
await authService.login({ email: 'seu@email.com', password: 'senha' });
```

### Erro: "Invalid API key"

**Causa**: API key expirou ou foi revogada

**Solução**: Fazer novo login para obter nova API key

### Frontend não envia Authorization header

**Causa**: `api-client.ts` não configurado corretamente

**Solução**: Verificar se o interceptor está adicionando o header:
```typescript
// Deve ter isso em api-client.ts
config.headers['Authorization'] = `Bearer ${token}`;
```

## 📚 Referências

### Backend
- Middleware: `infrastructure/http/middleware/auth.go`
- Routes: `infrastructure/http/routes/routes.go`
- Auth Handler: `infrastructure/http/handlers/auth_handler.go`
- User Service: `internal/application/user/user_service.go`

### Frontend
- API Client: `frontend/src/lib/api-client.ts`
- Auth Service: `frontend/src/services/auth.service.ts`
- Auth Hook: `frontend/src/hooks/useAuth.ts`
- Auth Types: `frontend/src/types/auth.types.ts`

## 📝 Checklist de Integração

- [x] Backend: Middleware de autenticação implementado
- [x] Backend: Rotas protegidas configuradas
- [x] Backend: API Keys funcionando
- [x] Frontend: API client com interceptor
- [x] Frontend: Services implementados
- [x] Frontend: Hooks de autenticação
- [x] Frontend: Tipos TypeScript
- [x] Frontend: Auto-logout em 401
- [x] Documentação: Guia completo
- [ ] Testes: E2E de autenticação
- [ ] Produção: HTTPS configurado
- [ ] Produção: Rate limiting
- [ ] Produção: Audit logs

---

**Status**: ✅ **Autenticação Totalmente Funcional**
**Última Atualização**: 2025-10-09
