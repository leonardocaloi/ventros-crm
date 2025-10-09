# Meta Apps: Modos, Permissões e Multi-Tenancy

## 📌 Entendendo Apps Públicos vs Development

### 1. Development Mode (Modo Desenvolvimento)

**Características:**
- ✅ **Grátis e imediato** - não precisa App Review
- ✅ Até **5 contas de teste** podem usar
- ✅ **Todas as permissões** disponíveis (whatsapp, ads, business_management)
- ❌ **NÃO funciona com usuários reais** fora da equipe do app
- ❌ Limitado a contas Admin/Developer/Tester adicionadas manualmente

**Quando usar:**
```
👨‍💻 Desenvolvimento local
🧪 Testes internos
📝 Prototipação
🎥 Gravação de screen recording para App Review
```

**Limitações:**
```go
// Apenas esses usuários podem autenticar:
- Você (desenvolvedor principal)
- Até 4 administradores adicionais
- Contas de teste criadas no Meta Dashboard

// ❌ Clientes externos NÃO conseguem conectar
```

---

### 2. Live Mode (Modo Público/Produção)

**Características:**
- ✅ **Qualquer usuário** do Facebook/WhatsApp pode conectar
- ✅ **Escalável** para milhares de clientes
- ✅ Multi-tenant pronto para SaaS
- ⚠️ **Requer App Review** para permissões avançadas
- ⚠️ **Business Verification** obrigatória

**Ativação:**
```
App Dashboard → Settings → Basic → App Mode → Switch to Live
```

**⚠️ IMPORTANTE:**
```
Você NÃO pode colocar app em Live sem App Review aprovado.
Se tentar, usuários externos verão erro:
"This app is in Development Mode"
```

---

## 🔐 Permissões: Standard vs Advanced Access

### Standard Access (Acesso Padrão)

**O que você tem SEM App Review:**

| Permissão | Standard Access | Limitações |
|-----------|-----------------|------------|
| `business_management` | ✅ Sim | Apenas para contas da sua organização |
| `whatsapp_business_messaging` | ⚠️ Limitado | Apenas contas de teste |
| `whatsapp_business_management` | ⚠️ Limitado | Apenas contas de teste |
| `ads_management` | ❌ Não | Sem acesso |
| `ads_read` | ⚠️ Limitado | Dados básicos apenas |

**Na prática:**
```go
// Com Standard Access você pode:
✅ Testar WhatsApp com suas próprias contas
✅ Ver estrutura de Ad Accounts (mas não criar ads)
✅ Acessar Business Portfolio da sua empresa

// Mas NÃO pode:
❌ Integrar clientes externos
❌ Criar campanhas de ads para terceiros
❌ Enviar mensagens WhatsApp para clientes finais
```

---

### Advanced Access (Acesso Avançado)

**O que você ganha COM App Review aprovado:**

| Permissão | Advanced Access | Capacidades |
|-----------|-----------------|-------------|
| `whatsapp_business_messaging` | ✅ Total | Enviar mensagens para QUALQUER usuário |
| `whatsapp_business_management` | ✅ Total | Gerenciar WABAs de clientes |
| `business_management` | ✅ Total | Acessar business portfolios de clientes |
| `ads_management` | ✅ Total | Criar/editar campanhas para clientes |
| `ads_read` | ✅ Total | Insights completos de anúncios |

**Depois do App Review:**
```go
// Seu app pode:
✅ Integrar QUALQUER cliente via Embedded Signup
✅ Enviar mensagens WhatsApp para usuários finais
✅ Gerenciar campanhas de ads de clientes
✅ Acessar métricas completas
✅ Escalar para milhares de clientes

// Limites de onboarding:
📊 Inicial: 10 novos clientes por 7 dias
📊 Após 10 clientes: 50 por 7 dias
📊 Após maturidade: 200+ por 7 dias
```

---

## 👥 Como Funciona para Múltiplos Usuários

### Cenário 1: SaaS Multi-Tenant (Ventros CRM)

**Arquitetura Recomendada:**

```
┌─────────────────────────────────────────────────────┐
│         UM ÚNICO APP META (Ventros CRM)             │
│                                                      │
│  App ID: 123456789                                  │
│  Mode: Live (após App Review)                       │
│  Permissions: Advanced Access                       │
└──────────────────┬──────────────────────────────────┘
                   │
                   │ Embedded Signup
                   │
      ┌────────────┼────────────┬────────────┐
      │            │            │            │
      ▼            ▼            ▼            ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│ Cliente A│ │ Cliente B│ │ Cliente C│ │Cliente...│
│          │ │          │ │          │ │    N     │
│ BISU     │ │ BISU     │ │ BISU     │ │ BISU     │
│ Token 1  │ │ Token 2  │ │ Token 3  │ │ Token N  │
│          │ │          │ │          │ │          │
│ WABA 1   │ │ WABA 2   │ │ WABA 3   │ │ WABA N   │
│ Ads 1    │ │ Ads 2    │ │ Ads 3    │ │ Ads N    │
└──────────┘ └──────────┘ └──────────┘ └──────────┘
```

**Como funciona:**

1. **Cliente A** clica em "Conectar Meta" no Ventros CRM
2. Abre **Embedded Signup** do seu app único
3. Cliente autentica com conta Facebook dele
4. Meta cria **BISU Token** isolado para Cliente A
5. Token salvo no banco com `tenant_id = cliente_a`
6. Cliente B repete o processo → BISU Token 2 isolado
7. Tokens são **completamente isolados** - Cliente A não acessa dados de B

**Vantagens:**
- ✅ **Um único app** para todos os clientes
- ✅ **Isolamento automático** via BISU tokens
- ✅ **Escalável** para milhares de clientes
- ✅ **Business verification** uma única vez
- ✅ **App Review** uma única vez

---

### Cenário 2: App por Cliente (NÃO Recomendado)

```
❌ Cliente A → App Meta A → Review A → Verification A
❌ Cliente B → App Meta B → Review B → Verification B
❌ Cliente C → App Meta C → Review C → Verification C
```

**Por que NÃO fazer assim:**
- ❌ Precisa **App Review** para cada cliente
- ❌ Precisa **Business Verification** para cada cliente
- ❌ Impossível escalar
- ❌ Manutenção de múltiplos apps
- ❌ Custos de compliance multiplicados

---

## 🔑 Uma Credencial que Engloba Tudo

### Opção 1: Credencial Unificada (Recomendado)

```go
// Um único tipo de credencial Meta com todas permissões
type MetaUnifiedCredential struct {
    // OAuth Token
    AccessToken  string // Criptografado
    ExpiresAt    time.Time

    // WhatsApp
    WABAID        string
    PhoneNumberID string

    // Ads
    AdAccountID   string
    PixelID       string

    // Business
    BusinessID    string
    PageID        *string

    // Permissões concedidas
    Permissions   []string
}

// credential_type = "meta_unified"
```

**Vantagens:**
- ✅ **Uma autenticação** = todas features
- ✅ **Um token** para WhatsApp + Ads + Conversions
- ✅ Simplicidade para o usuário
- ✅ Menos complexidade no código

**Uso:**
```go
// Cliente conecta UMA vez
credential := metaOAuthService.CreateUnifiedCredential(tenantID)

// Usa para WhatsApp
metaClient.SendWhatsAppMessage(credential.ID, to, message)

// Usa para Ads
metaClient.CreateAdCampaign(credential.ID, campaignData)

// Usa para Conversions
metaClient.SendConversionEvent(credential.ID, pixelID, event)
```

---

### Opção 2: Credenciais Separadas por Feature

```go
// Múltiplas credenciais especializadas
type WhatsAppCredential struct {
    AccessToken   string
    WABAID        string
    PhoneNumberID string
    Permissions   []string{"whatsapp_business_messaging"}
}

type AdsCredential struct {
    AccessToken   string
    AdAccountID   string
    PixelID       string
    Permissions   []string{"ads_management"}
}

// Tenant tem múltiplas credenciais
tenant_id = "abc123"
├── credential_whatsapp  (meta_whatsapp_cloud)
└── credential_ads       (meta_ads)
```

**Desvantagens:**
- ❌ Usuário precisa conectar **2-3 vezes**
- ❌ Gerenciar múltiplos tokens
- ❌ Complexidade no código
- ❌ UX ruim

**Quando usar:**
- ⚠️ Apenas se usuário quiser conectar **apenas WhatsApp OU apenas Ads**
- ⚠️ Compliance/segurança requer separação explícita

---

## 🎯 Recomendação Final para Ventros CRM

### Estratégia Recomendada:

```
┌─────────────────────────────────────────────────────┐
│         1 APP META em LIVE MODE                     │
│                                                      │
│  ✅ Advanced Access para todas permissões           │
│  ✅ Business Verification completa                  │
│  ✅ Embedded Signup configurado                     │
└─────────────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│      CREDENCIAL UNIFICADA "meta_unified"            │
│                                                      │
│  Permissões solicitadas:                            │
│  - whatsapp_business_messaging                      │
│  - whatsapp_business_management                     │
│  - business_management                              │
│  - ads_management                                   │
│  - ads_read                                         │
│                                                      │
│  Um token = todas features                          │
└─────────────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│        MULTI-TENANT via BISU TOKENS                 │
│                                                      │
│  Cada cliente: 1 botão → 1 token → todas features  │
│  Isolamento: tenant_id no PostgreSQL                │
│  Criptografia: AES-256-GCM (já implementado)        │
└─────────────────────────────────────────────────────┘
```

### Implementação Sugerida:

**1. Frontend - Um Botão:**
```jsx
<button onClick={connectMetaAccount}>
  🔗 Conectar Meta (WhatsApp + Ads)
</button>
```

**2. Backend - Uma Credencial:**
```go
// Solicitar TODAS permissões de uma vez
scopes := []string{
    "whatsapp_business_messaging",
    "whatsapp_business_management",
    "business_management",
    "ads_management",
    "ads_read",
}

// Salvar credencial unificada
credential := NewCredential(
    tenantID,
    CredentialTypeMeta,  // "meta_unified"
    "Meta Account",
    accessToken,
    encryptor,
)

// Armazenar metadados
credential.SetMetadata("waba_id", wabaID)
credential.SetMetadata("phone_number_id", phoneNumberID)
credential.SetMetadata("ad_account_id", adAccountID)
credential.SetMetadata("pixel_id", pixelID)
credential.SetMetadata("business_id", businessID)
credential.SetMetadata("permissions", permissions)
```

**3. Uso - Mesma Credencial:**
```go
// WhatsApp
SendWhatsAppMessage(credentialID, to, message)

// Ads
CreateAdCampaign(credentialID, campaignData)

// Conversions
SendConversionEvent(credentialID, eventData)
```

---

## ⚠️ Permissões que Usuário Pode Recusar

Quando usuário conecta via Embedded Signup, ele pode **desmarcar permissões**:

```
┌────────────────────────────────────────┐
│  Conectar Meta ao Ventros CRM          │
│                                         │
│  ☑️ WhatsApp Business Messaging         │
│  ☑️ WhatsApp Business Management        │
│  ☑️ Business Management                 │
│  ☐ Ads Management  ← Usuário desmarcou │
│  ☑️ Read Ads Data                       │
│                                         │
│  [Continuar]  [Cancelar]               │
└────────────────────────────────────────┘
```

**Como lidar:**
```go
// Após callback OAuth, verificar permissões concedidas
tokenResp := metaOAuthService.ExchangeCodeForToken(code)

grantedPermissions := tokenResp.Permissions
// ["whatsapp_business_messaging", "business_management"]

// Verificar se tem permissões mínimas
requiredPermissions := []string{
    "whatsapp_business_messaging",
    "business_management",
}

if !hasAllPermissions(grantedPermissions, requiredPermissions) {
    return errors.New("Permissões mínimas não concedidas")
}

// Salvar permissões concedidas no metadata
credential.SetMetadata("permissions", grantedPermissions)

// Frontend verifica e desabilita features não autorizadas
if !hasPermission(credential, "ads_management") {
    // Esconder botão "Criar Campanha"
}
```

---

## 📊 Comparação de Cenários

### Cenário A: App em Development

```
✅ Rápido para testar
✅ Todas permissões disponíveis
✅ Sem custos

❌ Máximo 5 usuários
❌ Apenas contas de teste
❌ NÃO serve para produção SaaS
```

### Cenário B: App Live + Standard Access (SEM Review)

```
✅ Qualquer usuário pode conectar
✅ Botão de login funciona

❌ Permissões limitadas
❌ NÃO pode enviar WhatsApp para usuários finais
❌ NÃO pode criar ads para clientes
❌ Serve apenas para "login social"
```

### Cenário C: App Live + Advanced Access (COM Review) ⭐

```
✅ Qualquer usuário pode conectar
✅ TODAS permissões funcionam
✅ Enviar WhatsApp para usuários finais
✅ Criar/gerenciar ads de clientes
✅ Multi-tenant pronto
✅ Escalável para milhares de clientes

⚠️ Requer App Review (2-15 dias)
⚠️ Requer Business Verification
⚠️ Requer screen recordings e documentação
```

---

## 🚀 Roadmap de Implementação

### Fase 1: Development Mode (Semana 1-2)
```
1. Criar app Meta em Development
2. Implementar Embedded Signup (código Go)
3. Testar com contas de teste
4. Validar fluxo completo
5. Gravar screen recordings para App Review
```

### Fase 2: App Review (Semana 3)
```
1. Completar Business Verification
2. Submeter App Review
3. Upload screen recordings
4. Aguardar aprovação (2-15 dias)
```

### Fase 3: Production (Semana 4-5)
```
1. Receber aprovação
2. Colocar app em Live Mode
3. Testar com primeiros clientes reais
4. Monitorar limites de onboarding
5. Escalar gradualmente
```

---

## 🤔 Perguntas Frequentes

### 1. Preciso de um app para cada cliente?
**Não!** Um único app serve todos os clientes via BISU tokens.

### 2. Como funciona isolamento entre clientes?
BISU tokens são **scoped por cliente** automaticamente. Cliente A nunca acessa dados de Cliente B.

### 3. Posso usar o mesmo token para WhatsApp e Ads?
**Sim!** Um único token Meta pode ter múltiplas permissões.

### 4. Quanto tempo leva o App Review?
Normalmente **2-15 dias úteis**. Pode ser mais rápido se documentação estiver perfeita.

### 5. App em Development serve para produção?
**Não!** Máximo 5 usuários. Para SaaS real, precisa estar em Live Mode.

### 6. Posso testar antes do App Review?
**Sim!** Use Development Mode com contas de teste para validar tudo.

### 7. E se usuário revogar permissões depois?
Configure **webhook de deauthorization** para detectar e desativar credencial automaticamente.

### 8. Tokens expiram?
BISU tokens duram **60 dias**. Implemente refresh automático antes de expirar.

---

## 📚 Recursos Adicionais

- [Meta App Modes](https://developers.facebook.com/docs/development/build-and-test/app-modes/)
- [Permission Reference](https://developers.facebook.com/docs/permissions/reference)
- [App Review Process](https://developers.facebook.com/docs/app-review/)
- [Business Verification](https://www.facebook.com/business/help/2058515294227817)
- [Embedded Signup](https://developers.facebook.com/docs/whatsapp/embedded-signup/)
