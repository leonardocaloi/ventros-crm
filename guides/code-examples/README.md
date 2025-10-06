# 💻 Code Examples

Exemplos práticos de padrões e funcionalidades do Ventros CRM.

---

## 📚 Exemplos Disponíveis

### 🛡️ RBAC (Role-Based Access Control)
**Arquivo**: `rbac_example.go`

Exemplo completo de como implementar controle de acesso baseado em roles nas rotas da API.

**Features demonstradas**:
- Middleware de autenticação
- Middleware RBAC
- Definição de recursos e operações
- Verificação de permissões
- Uso programático de roles

**Roles disponíveis**:
- `Admin` - Acesso total
- `Manager` - Gerenciamento de equipe e analytics
- `User` - CRUD de recursos próprios
- `ReadOnly` - Apenas leitura

**Exemplo de uso**:
```go
// Requer permissão específica
webhooks.POST("", 
    rbacMiddleware.RequirePermission(user.ResourceWebhook, user.OperationCreate),
    handler,
)

// Requer role específica
users.POST("", 
    rbacMiddleware.RequireRole(user.RoleAdmin),
    handler,
)

// Requer uma das roles
analytics.GET("/export", 
    rbacMiddleware.RequireAnyRole(user.RoleAdmin, user.RoleManager),
    handler,
)
```

---

## 🎯 Como Usar

1. **Leia o código**: Cada exemplo está bem comentado
2. **Adapte ao seu caso**: Copie e modifique conforme necessário
3. **Teste**: Sempre teste suas implementações

---

## 📖 Mais Exemplos em Breve

- `use_case_example.go` - Padrão de Use Cases
- `event_handler_example.go` - Domain Events e choreography
- `repository_example.go` - Repository pattern
- `temporal_workflow_example.go` - Workflows duráveis

---

## 🤝 Contribua

Tem um exemplo útil? Adicione aqui seguindo o mesmo padrão!

Ver [CONTRIBUTING.md](../../CONTRIBUTING.md) para guidelines.
