# Integração de Autenticação - Frontend e Backend

## 📋 Visão Geral

Este documento descreve a integração correta entre o frontend e o backend para autenticação, seguindo as melhores práticas da indústria.

## 🏗️ Arquitetura

### Backend (Go + Gin)
- **Localização**: `infrastructure/http/handlers/auth_handler.go`
- **Autenticação**: API Keys (Bearer tokens)
- **Storage**: PostgreSQL com hash SHA-256
- **Endpoints**: `/api/v1/auth/*`

### Frontend (Next.js + TypeScript)
- **Localização**: `ventros-frontend/src/app/(auth)/`
- **Storage**: localStorage (chave: `api_key`)
- **Comunicação**: Server Actions + Client Services

## 🔐 Fluxo de Autenticação

### 1. Registro de Usuário

#### Frontend Request
```typescript
// ventros-frontend/src/app/(auth)/server-actions.ts:89-208
interface RegisterPayload {
  name: string;      // Nome completo: "FirstName LastName"
  email: string;
  password: string;
  role?: string;     // Opcional, padrão: "user"
}
```

#### Backend Response
```json
{
  "message": "User created successfully",
  "user_id": "uuid",
  "name": "João Silva",
  "email": "joao@empresa.com",
  "role": "user",
  "api_key": "64-char-hex-string",
  "default_project_id": "uuid",
  "default_pipeline_id": "uuid",
  "note": "Save this API key - it won't be shown again"
}
```

#### Validações Frontend
- ✅ Senha mínima: 8 caracteres
- ✅ Deve conter: maiúscula, minúscula, número, caractere especial
- ✅ Confirmação de senha deve coincidir
- ✅ Nome e sobrenome obrigatórios

#### Backend Behavior
```go
// internal/application/user/user_service.go:61-240
// 1. Valida role
// 2. Verifica se usuário já existe (idempotente)
// 3. Hash bcrypt da senha
// 4. Cria usuário, billing account, projeto e pipeline em transação
// 5. Gera API key (SHA-256 hash)
// 6. Retorna todos os dados incluindo API key
```

### 2. Login de Usuário

#### Frontend Request
```typescript
// ventros-frontend/src/app/(auth)/server-actions.ts:5-88
interface LoginPayload {
  email: string;
  password: string;
}
```

#### Backend Response
```json
{
  "message": "Login successful",
  "user_id": "uuid",
  "email": "joao@empresa.com",
  "role": "user",
  "api_key": "64-char-hex-string",
  "default_project_id": "uuid"
}
```

#### Backend Behavior
```go
// internal/application/user/user_service.go:242-289
// 1. Busca usuário por email e status ativo
// 2. Verifica senha com bcrypt
// 3. Retorna API key existente ou gera nova
// 4. Atualiza last_used da API key
```

### 3. Autenticação em Requisições

#### Frontend - Client Side
```typescript
// frontend/src/lib/api-client.ts
const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL + '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Interceptor adiciona automaticamente
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('api_key');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

#### Backend - Middleware
```go
// infrastructure/http/middleware/auth.go
// 1. Extrai token do header Authorization
// 2. Valida formato Bearer
// 3. Busca API key no banco (SHA-256 hash)
// 4. Valida usuário ativo
// 5. Injeta contexto no Gin com user_id, email, role, tenant_id
```

## 📁 Estrutura de Arquivos

```
ventros-crm/
├── infrastructure/
│   └── http/
│       ├── handlers/
│       │   └── auth_handler.go          # Handlers de autenticação
│       ├── middleware/
│       │   └── auth.go                  # Middleware de autenticação
│       └── routes/
│           └── routes.go                # Rotas (linha 311-324)
├── internal/
│   └── application/
│       └── user/
│           └── user_service.go          # Lógica de negócio
├── frontend/
│   └── src/
│       ├── services/
│       │   └── auth.service.ts          # Client-side API calls
│       ├── hooks/
│       │   └── useAuth.ts               # React hook
│       ├── lib/
│       │   └── api-client.ts            # Axios config
│       └── types/
│           └── auth.types.ts            # TypeScript types
└── ventros-frontend/
    └── src/
        └── app/
            └── (auth)/
                ├── server-actions.ts     # Server actions
                ├── register/
                │   └── page.tsx         # Página de registro
                └── login/
                    └── page.tsx         # Página de login
```

## 🔧 Configuração

### Variáveis de Ambiente

#### Backend (.env)
```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/ventros_crm

# Server
PORT=8080
GIN_MODE=release
```

#### Frontend (.env.local)
```bash
# API Backend URL
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080

# App URL
NEXT_PUBLIC_BASE_URL=http://localhost:3000

# Auth Config
NEXT_PUBLIC_AUTH_MODE=api_key
NEXT_PUBLIC_AUTH_TOKEN_KEY=api_key
```

## 🚀 Melhores Práticas Implementadas

### ✅ Segurança
1. **Senha Hasheada**: bcrypt com custo padrão
2. **API Key Hasheada**: SHA-256 no banco
3. **Validação Server-Side**: Regex de senha no backend
4. **HTTPS Only (Produção)**: Configurar reverse proxy
5. **CORS Configurado**: Apenas origens permitidas

### ✅ Experiência do Usuário
1. **Feedback Imediato**: Validação client-side antes de enviar
2. **Mensagens Claras**: Erros específicos e acionáveis
3. **Loading States**: Indicadores visuais durante requisições
4. **Auto-login**: Após registro bem-sucedido

### ✅ Manutenibilidade
1. **Tipos Fortes**: TypeScript no frontend
2. **Documentação**: Comentários com referências ao código backend
3. **Separação de Concerns**: Service layer, hooks, components
4. **Idempotência**: Registro não falha se usuário existir

### ✅ Performance
1. **API Key Caching**: Reutilização de keys existentes
2. **Conexão Persistente**: Connection pool no PostgreSQL
3. **Transações**: Operações atômicas no registro

## 🧪 Testando a Integração

### 1. Teste de Registro

```bash
# Backend deve estar rodando na porta 8080
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "João Silva",
    "email": "joao@teste.com",
    "password": "Senha@123",
    "role": "user"
  }'
```

### 2. Teste de Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "joao@teste.com",
    "password": "Senha@123"
  }'
```

### 3. Teste de Autenticação

```bash
# Usar api_key retornada do login/registro
curl http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer <api_key>"
```

## 🐛 Troubleshooting

### Erro: "Invalid credentials"
- ✅ Verificar se o usuário existe no banco
- ✅ Verificar se a senha está correta
- ✅ Verificar se o usuário está ativo (`status = 'active'`)

### Erro: "Not authenticated"
- ✅ Verificar se API key está no header Authorization
- ✅ Verificar formato: `Bearer <api_key>`
- ✅ Verificar se API key está ativa no banco

### Erro: "Connection refused"
- ✅ Verificar se backend está rodando
- ✅ Verificar porta correta (8080)
- ✅ Verificar variável NEXT_PUBLIC_API_BASE_URL

### Erro de CORS
- ✅ Configurar CORS no backend
- ✅ Adicionar origem do frontend nas origens permitidas

## 📚 Referências

- Backend Handler: `infrastructure/http/handlers/auth_handler.go`
- Backend Service: `internal/application/user/user_service.go`
- Backend Middleware: `infrastructure/http/middleware/auth.go`
- Frontend Service: `frontend/src/services/auth.service.ts`
- Frontend Hook: `frontend/src/hooks/useAuth.ts`
- Frontend Types: `frontend/src/types/auth.types.ts`
- Server Actions: `ventros-frontend/src/app/(auth)/server-actions.ts`

## 🔄 Fluxo Completo

```
┌─────────────┐
│   Browser   │
└─────┬───────┘
      │
      │ 1. Preenche formulário
      ▼
┌─────────────────────┐
│  Register Page      │
│  (Next.js)          │
└─────┬───────────────┘
      │
      │ 2. Valida client-side
      ▼
┌─────────────────────┐
│  Server Action      │
│  register()         │
└─────┬───────────────┘
      │
      │ 3. POST /api/v1/auth/register
      ▼
┌─────────────────────┐
│  Auth Handler       │
│  (Go/Gin)           │
└─────┬───────────────┘
      │
      │ 4. Valida backend
      ▼
┌─────────────────────┐
│  User Service       │
│  CreateUser()       │
└─────┬───────────────┘
      │
      │ 5. Transação DB
      ▼
┌─────────────────────┐
│  PostgreSQL         │
│  - users            │
│  - billing_accounts │
│  - projects         │
│  - pipelines        │
│  - user_api_keys    │
└─────┬───────────────┘
      │
      │ 6. Response com API key
      ▼
┌─────────────────────┐
│  Browser            │
│  - Salva API key    │
│  - Redireciona      │
└─────────────────────┘
```

## ✅ Checklist de Implementação

- [x] Backend: Endpoint de registro
- [x] Backend: Endpoint de login
- [x] Backend: Endpoint de profile
- [x] Backend: Middleware de autenticação
- [x] Backend: Validação de senha
- [x] Frontend: Server actions alinhadas
- [x] Frontend: Tipos TypeScript corretos
- [x] Frontend: Service layer atualizado
- [x] Frontend: Hook useAuth atualizado
- [x] Frontend: Validação de formulário
- [x] Frontend: Configuração de ambiente
- [x] Documentação: Arquivo de integração
- [ ] Testes: E2E do fluxo completo
- [ ] Testes: Unitários do backend
- [ ] Testes: Unitários do frontend
- [ ] Deploy: Configuração de produção

## 📝 Próximos Passos

1. **Implementar Refresh de API Keys**: Rotação automática de keys
2. **Rate Limiting**: Limitar tentativas de login
3. **2FA (Two-Factor Auth)**: Autenticação em dois fatores
4. **OAuth Integration**: Login social (Google, GitHub)
5. **Session Management**: Gerenciar múltiplas sessões
6. **Audit Logging**: Log de todas as ações de autenticação
