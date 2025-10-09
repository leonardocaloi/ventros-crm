# Guia de Implementação: Meta OAuth 2.0 para WhatsApp Cloud API e Facebook Ads

## 📋 Índice

1. [Visão Geral](#visão-geral)
2. [Arquitetura da Solução](#arquitetura-da-solução)
3. [Tipos de Tokens e Quando Usar](#tipos-de-tokens-e-quando-usar)
4. [Permissões Necessárias](#permissões-necessárias)
5. [Fluxo de Autenticação (Embedded Signup)](#fluxo-de-autenticação-embedded-signup)
6. [Implementação em Go](#implementação-em-go)
7. [App Review e Permissões Avançadas](#app-review-e-permissões-avançadas)
8. [Webhooks de Deauthorization](#webhooks-de-deauthorization)
9. [Multi-Tenancy e Gestão de Credenciais](#multi-tenancy-e-gestão-de-credenciais)
10. [Renovação Automática de Tokens](#renovação-automática-de-tokens)
11. [Segurança e Boas Práticas](#segurança-e-boas-práticas)

---

## Visão Geral

Este documento descreve a implementação de autenticação OAuth 2.0 da Meta para permitir que o **Ventros CRM** gerencie:

- **WhatsApp Cloud API**: Envio e recebimento de mensagens
- **Facebook Ads Management**: Criação e gerenciamento de campanhas
- **Meta Conversions API**: Envio de eventos de conversão
- **Business Management**: Gerenciamento de ativos empresariais

### Cenário de Uso

O sistema permitirá que **múltiplos usuários/clientes** conectem suas contas Meta ao CRM através de um botão de login OAuth, sem necessidade de configuração manual. Cada cliente terá suas próprias credenciais isoladas (multi-tenant).

---

## Arquitetura da Solução

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend (React/Vue)                     │
│                                                               │
│  ┌─────────────────────┐   ┌─────────────────────┐          │
│  │  "Conectar Meta"    │   │  "Gerenciar         │          │
│  │  OAuth Button       │   │  Permissões"        │          │
│  └──────────┬──────────┘   └─────────────────────┘          │
└─────────────┼───────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend (Go - API Layer)                  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  OAuth Handler                                       │   │
│  │  - /api/v1/integrations/meta/auth/start              │   │
│  │  - /api/v1/integrations/meta/auth/callback           │   │
│  │  - /api/v1/integrations/meta/auth/refresh            │   │
│  │  - /api/v1/integrations/meta/webhooks/deauth         │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Application Service                                 │   │
│  │  - MetaOAuthService                                  │   │
│  │  - CredentialRefreshService                          │   │
│  │  - MetaAPIClient                                     │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Domain Layer                                        │   │
│  │  - Credential Aggregate (já existe)                  │   │
│  │  - MetaCredential Value Object                       │   │
│  │  - OAuthToken Value Object (já existe)               │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Infrastructure                                      │   │
│  │  - AES-256-GCM Encryption                            │   │
│  │  - PostgreSQL (credentials table)                    │   │
│  │  - RabbitMQ (refresh events)                         │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────┬───────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Meta APIs                                │
│  - Graph API (OAuth, Business Management)                   │
│  - WhatsApp Cloud API                                        │
│  - Marketing API (Ads)                                       │
│  - Conversions API                                           │
└─────────────────────────────────────────────────────────────┘
```

---

## Tipos de Tokens e Quando Usar

### 1. **User Access Token** (Token de Usuário)

**Características:**
- Curta duração (algumas horas)
- Representa um usuário individual do Facebook
- Expira rapidamente

**Quando usar:**
- ❌ **NÃO use em produção**
- ✅ Apenas para desenvolvimento inicial e testes

### 2. **System User Access Token** (Token de Usuário do Sistema)

**Características:**
- Longa duração (pode ser permanente)
- Representa sua organização/empresa
- Dois subtipos:
  - **Admin System User**: Acesso total a todos os WABAs e ativos
  - **Employee System User**: Acesso restrito a WABAs específicos

**Quando usar:**
- ✅ **Desenvolvedor direto** (Direct Developer) acessando seus próprios dados
- ✅ **Solution Partner** compartilhando linha de crédito
- ❌ **NÃO use para multi-tenancy com clientes externos**

**Criação Manual:**
1. Business Settings → Users → System Users
2. Create System User (Admin ou Employee)
3. Assign App + Permissions
4. Generate Token

### 3. **Business Integration System User Access Token** (Token BISU) ⭐

**Características:**
- Longa duração
- **Escopo por cliente** (customer-scoped)
- Gerado via **Embedded Signup** OAuth flow
- Isolamento automático entre clientes (multi-tenant)

**Quando usar:**
- ✅ **Multi-tenant SaaS** (como o Ventros CRM)
- ✅ **Tech Provider** acessando dados de clientes onboarded
- ✅ **Solution Partner** para operações que não sejam compartilhamento de crédito

**Geração:**
```
Implementar Embedded Signup → Usuário autentica →
Recebe código → Troca por BISU Token
```

### 🎯 **Recomendação para o Ventros CRM**

Use **Business Integration System User Access Token** via **Embedded Signup**:

- Cada cliente terá seu próprio token isolado
- Gerenciamento automático de permissões
- Não precisa criar System Users manualmente
- Escalável para centenas/milhares de clientes

---

## Permissões Necessárias

### WhatsApp Cloud API

```go
// Permissões básicas
var whatsappScopes = []string{
    "whatsapp_business_messaging",    // Enviar/receber mensagens
    "whatsapp_business_management",   // Gerenciar WABAs, templates
    "business_management",            // Acessar business portfolio
}
```

**Capacidades:**
- Enviar mensagens (texto, mídia, templates)
- Receber webhooks de mensagens
- Gerenciar templates de mensagem
- Configurar perfil comercial
- Registrar números de telefone

### Facebook Ads Management

```go
// Permissões para anúncios
var adsScopes = []string{
    "ads_management",         // Criar e gerenciar campanhas
    "ads_read",              // Ler insights e métricas
    "business_management",    // Gerenciar ad accounts
}
```

**Capacidades:**
- Criar campanhas publicitárias
- Gerenciar conjuntos de anúncios
- Ler métricas e insights
- Configurar pixels de conversão

### Meta Conversions API

```go
// Permissões para eventos
var conversionsScopes = []string{
    "ads_management",         // Necessário para Conversions API
    "business_management",
}
```

### Permissões Combinadas (Todas as Features)

```go
// Todas as permissões para um sistema completo
var allMetaScopes = []string{
    "whatsapp_business_messaging",
    "whatsapp_business_management",
    "business_management",
    "ads_management",
    "ads_read",
    "pages_messaging",           // Opcional: mensagens de páginas
    "pages_manage_metadata",     // Opcional: gerenciar páginas
    "pages_read_engagement",     // Opcional: analytics de páginas
}
```

---

## Fluxo de Autenticação (Embedded Signup)

### Passo 1: Configuração Inicial no Meta Developers

1. Criar **Business App** em https://developers.facebook.com/apps
2. Adicionar produtos:
   - **Facebook Login for Business**
   - **WhatsApp Business Platform**
   - **Marketing API** (para ads)
3. Configurar **Embedded Signup**:
   - Obter `Configuration ID`
   - Configurar `Redirect URI`: `https://seu-dominio.com/api/v1/integrations/meta/auth/callback`
4. Configurar **Webhooks**:
   - Callback URL: `https://seu-dominio.com/api/v1/integrations/meta/webhooks`
   - Subscribe to: `messages`, `messaging_postbacks`, etc.

### Passo 2: Iniciar Fluxo OAuth (Frontend)

```javascript
// Frontend - React/Vue
async function connectMetaAccount() {
  // 1. Configurar Facebook SDK
  window.fbAsyncInit = function() {
    FB.init({
      appId: process.env.REACT_APP_META_APP_ID,
      cookie: true,
      xfbml: true,
      version: 'v18.0'
    });
  };

  // 2. Chamar Embedded Signup
  FB.login(function(response) {
    if (response.authResponse) {
      const { code } = response.authResponse;

      // 3. Enviar code para backend
      fetch('/api/v1/integrations/meta/auth/callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code: code,
          configuration_id: process.env.REACT_APP_META_CONFIG_ID
        })
      });
    }
  }, {
    config_id: process.env.REACT_APP_META_CONFIG_ID,
    response_type: 'code',
    override_default_response_type: true,
    extras: {
      setup: {
        // Embedded Signup para WhatsApp
      }
    }
  });
}
```

### Passo 3: Backend Troca Code por Token

```go
// Backend - Go Handler
func (h *MetaOAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Code            string `json:"code"`
        ConfigurationID string `json:"configuration_id"`
    }

    json.NewDecoder(r.Body).Decode(&req)

    // 1. Trocar code por token
    tokenResp, err := h.metaOAuthService.ExchangeCodeForToken(
        r.Context(),
        req.Code,
    )
    if err != nil {
        http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
        return
    }

    // 2. Salvar credencial criptografada
    tenantID := r.Context().Value("tenant_id").(string)

    credential, err := h.credentialService.CreateMetaCredential(
        r.Context(),
        tenantID,
        tokenResp.AccessToken,
        tokenResp.RefreshToken,
        tokenResp.ExpiresIn,
        tokenResp.WABAID,
        tokenResp.PhoneNumberID,
    )

    // 3. Retornar sucesso
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "credential_id": credential.ID(),
        "waba_id": tokenResp.WABAID,
    })
}
```

---

## Implementação em Go

### Estrutura de Diretórios

```
internal/
├── domain/
│   └── credential/
│       ├── credential.go              # Já existe
│       ├── credential_type.go         # Já existe - adicionar novos tipos
│       ├── oauth_token.go             # Já existe
│       └── meta_credential.go         # NOVO - Value Object para Meta
│
├── application/
│   └── integration/
│       ├── meta_oauth_service.go      # NOVO - Lógica OAuth
│       ├── meta_api_client.go         # NOVO - Cliente Graph API
│       └── credential_refresh_service.go  # NOVO - Refresh automático
│
└── infrastructure/
    ├── http/
    │   └── handlers/
    │       ├── meta_oauth_handler.go  # NOVO - Endpoints OAuth
    │       └── meta_webhook_handler.go  # NOVO - Webhooks Meta
    │
    └── meta/
        ├── oauth_client.go            # NOVO - HTTP client Meta
        └── embedded_signup.go         # NOVO - Embedded Signup logic
```

### 1. Atualizar CredentialType

```go
// internal/domain/credential/credential_type.go

const (
    // ... tipos existentes ...

    // Meta Integrations - UNIFIED
    CredentialTypeMeta CredentialType = "meta_unified"  // Todas permissões Meta

    // Meta Integrations - LEGACY (manter para backward compatibility)
    CredentialTypeMetaWhatsApp    CredentialType = "meta_whatsapp_cloud"
    CredentialTypeMetaAds         CredentialType = "meta_ads"
    CredentialTypeMetaConversions CredentialType = "meta_conversions_api"
)

func (t CredentialType) GetScopes() []string {
    switch t {
    case CredentialTypeMeta:
        return []string{
            "whatsapp_business_messaging",
            "whatsapp_business_management",
            "business_management",
            "ads_management",
            "ads_read",
        }
    case CredentialTypeMetaWhatsApp:
        return []string{
            "whatsapp_business_messaging",
            "whatsapp_business_management",
            "business_management",
        }
    case CredentialTypeMetaAds, CredentialTypeMetaConversions:
        return []string{
            "ads_management",
            "ads_read",
            "business_management",
        }
    // ... outros casos ...
    }
}

func (t CredentialType) GetOAuthEndpoint() string {
    switch t {
    case CredentialTypeMeta, CredentialTypeMetaWhatsApp,
         CredentialTypeMetaAds, CredentialTypeMetaConversions:
        return "https://www.facebook.com/v18.0/dialog/oauth"
    // ... outros casos ...
    }
}

func (t CredentialType) GetTokenEndpoint() string {
    switch t {
    case CredentialTypeMeta, CredentialTypeMetaWhatsApp,
         CredentialTypeMetaAds, CredentialTypeMetaConversions:
        return "https://graph.facebook.com/v18.0/oauth/access_token"
    // ... outros casos ...
    }
}
```

### 2. Criar MetaCredential Value Object

```go
// internal/domain/credential/meta_credential.go

package credential

import "github.com/google/uuid"

// MetaCredential contém metadados específicos da Meta
type MetaCredential struct {
    WABAID        string      // WhatsApp Business Account ID
    PhoneNumberID string      // Phone Number ID
    BusinessID    string      // Business Portfolio ID
    AdAccountID   *string     // Ad Account ID (opcional)
    PixelID       *string     // Pixel ID (opcional)
    Permissions   []string    // Permissões concedidas
}

// ToMetadata converte para map para salvar em Credential.metadata
func (m MetaCredential) ToMetadata() map[string]interface{} {
    metadata := map[string]interface{}{
        "waba_id":         m.WABAID,
        "phone_number_id": m.PhoneNumberID,
        "business_id":     m.BusinessID,
        "permissions":     m.Permissions,
    }

    if m.AdAccountID != nil {
        metadata["ad_account_id"] = *m.AdAccountID
    }

    if m.PixelID != nil {
        metadata["pixel_id"] = *m.PixelID
    }

    return metadata
}

// FromMetadata reconstrói a partir de metadata
func MetaCredentialFromMetadata(metadata map[string]interface{}) MetaCredential {
    mc := MetaCredential{}

    if val, ok := metadata["waba_id"].(string); ok {
        mc.WABAID = val
    }
    if val, ok := metadata["phone_number_id"].(string); ok {
        mc.PhoneNumberID = val
    }
    if val, ok := metadata["business_id"].(string); ok {
        mc.BusinessID = val
    }
    if val, ok := metadata["permissions"].([]interface{}); ok {
        for _, p := range val {
            if pStr, ok := p.(string); ok {
                mc.Permissions = append(mc.Permissions, pStr)
            }
        }
    }

    if val, ok := metadata["ad_account_id"].(string); ok {
        mc.AdAccountID = &val
    }

    if val, ok := metadata["pixel_id"].(string); ok {
        mc.PixelID = &val
    }

    return mc
}
```

### 3. Criar MetaOAuthService

```go
// internal/application/integration/meta_oauth_service.go

package integration

import (
    "context"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "encoding/json"
    "io"

    "github.com/google/uuid"
    "ventros-crm/internal/domain/credential"
)

type MetaOAuthService struct {
    credentialRepo credential.Repository
    encryptor      credential.Encryptor
    appID          string
    appSecret      string
    redirectURI    string
}

// TokenResponse representa a resposta do Meta após troca de código
type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`

    // Dados do Embedded Signup
    WABAID        string   `json:"waba_id"`
    PhoneNumberID string   `json:"phone_number_id"`
    BusinessID    string   `json:"business_id"`
    Permissions   []string `json:"granted_scopes"`
}

func NewMetaOAuthService(
    credentialRepo credential.Repository,
    encryptor credential.Encryptor,
    appID, appSecret, redirectURI string,
) *MetaOAuthService {
    return &MetaOAuthService{
        credentialRepo: credentialRepo,
        encryptor:      encryptor,
        appID:          appID,
        appSecret:      appSecret,
        redirectURI:    redirectURI,
    }
}

// GetAuthorizationURL gera a URL de autorização OAuth
func (s *MetaOAuthService) GetAuthorizationURL(
    state string,
    scopes []string,
) string {
    params := url.Values{}
    params.Add("client_id", s.appID)
    params.Add("redirect_uri", s.redirectURI)
    params.Add("state", state)
    params.Add("scope", scopesToString(scopes))
    params.Add("response_type", "code")

    return fmt.Sprintf(
        "https://www.facebook.com/v18.0/dialog/oauth?%s",
        params.Encode(),
    )
}

// ExchangeCodeForToken troca o código de autorização por access token
func (s *MetaOAuthService) ExchangeCodeForToken(
    ctx context.Context,
    code string,
) (*TokenResponse, error) {
    // 1. Construir request
    tokenURL := "https://graph.facebook.com/v18.0/oauth/access_token"

    params := url.Values{}
    params.Add("client_id", s.appID)
    params.Add("client_secret", s.appSecret)
    params.Add("redirect_uri", s.redirectURI)
    params.Add("code", code)

    // 2. Fazer request
    resp, err := http.Get(tokenURL + "?" + params.Encode())
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("token exchange failed: %s", string(body))
    }

    // 3. Decodificar resposta
    var tokenResp TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return nil, fmt.Errorf("failed to decode token response: %w", err)
    }

    // 4. Buscar informações adicionais (WABA ID, Business ID, etc)
    if err := s.enrichTokenResponse(ctx, &tokenResp); err != nil {
        return nil, fmt.Errorf("failed to enrich token: %w", err)
    }

    return &tokenResp, nil
}

// enrichTokenResponse busca informações adicionais do usuário/negócio
func (s *MetaOAuthService) enrichTokenResponse(
    ctx context.Context,
    tokenResp *TokenResponse,
) error {
    // Buscar informações do usuário/negócio via Graph API
    // GET /me?fields=id,businesses{id,name},whatsapp_business_accounts

    debugURL := fmt.Sprintf(
        "https://graph.facebook.com/v18.0/debug_token?input_token=%s&access_token=%s|%s",
        tokenResp.AccessToken,
        s.appID,
        s.appSecret,
    )

    resp, err := http.Get(debugURL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var debugResp struct {
        Data struct {
            Scopes []string `json:"scopes"`
        } `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&debugResp); err != nil {
        return err
    }

    tokenResp.Permissions = debugResp.Data.Scopes

    return nil
}

// CreateMetaCredential cria uma credencial Meta no domain
func (s *MetaOAuthService) CreateMetaCredential(
    ctx context.Context,
    tenantID string,
    projectID *uuid.UUID,
    tokenResp *TokenResponse,
) (*credential.Credential, error) {
    // 1. Criar credencial base
    cred, err := credential.NewCredential(
        tenantID,
        credential.CredentialTypeMeta,
        "Meta Unified Account",
        tokenResp.AccessToken, // Será criptografado
        s.encryptor,
    )
    if err != nil {
        return nil, err
    }

    // 2. Adicionar OAuth token
    // Nota: Meta não fornece refresh token no Embedded Signup
    // Tokens de BISU são long-lived (60 dias)
    if err := cred.SetOAuthToken(
        tokenResp.AccessToken,
        "", // Sem refresh token
        tokenResp.ExpiresIn,
        s.encryptor,
    ); err != nil {
        return nil, err
    }

    // 3. Adicionar metadados Meta
    metaCred := credential.MetaCredential{
        WABAID:        tokenResp.WABAID,
        PhoneNumberID: tokenResp.PhoneNumberID,
        BusinessID:    tokenResp.BusinessID,
        Permissions:   tokenResp.Permissions,
    }

    for key, value := range metaCred.ToMetadata() {
        cred.SetMetadata(key, value)
    }

    // 4. Associar ao projeto (se fornecido)
    if projectID != nil {
        cred.SetProjectID(*projectID)
    }

    // 5. Salvar no repositório
    if err := s.credentialRepo.Save(cred); err != nil {
        return nil, err
    }

    return cred, nil
}

// RefreshAccessToken renova o access token (se aplicável)
// Nota: BISU tokens da Meta não têm refresh token tradicional
// Eles são long-lived (60 dias) e precisam ser renovados via re-autenticação
func (s *MetaOAuthService) RefreshAccessToken(
    ctx context.Context,
    credentialID uuid.UUID,
) error {
    // 1. Buscar credencial
    cred, err := s.credentialRepo.FindByID(credentialID)
    if err != nil {
        return err
    }

    // 2. Verificar se precisa renovar
    if !cred.NeedsRefresh() {
        return nil // Ainda válido
    }

    // 3. Para Meta BISU tokens, renovar significa trocar long-lived por long-lived
    // GET /oauth/access_token?grant_type=fb_exchange_token&
    //     client_id={app-id}&client_secret={app-secret}&
    //     fb_exchange_token={short-lived-token}

    currentToken, err := cred.GetAccessToken(s.encryptor)
    if err != nil {
        return err
    }

    exchangeURL := fmt.Sprintf(
        "https://graph.facebook.com/v18.0/oauth/access_token?grant_type=fb_exchange_token&client_id=%s&client_secret=%s&fb_exchange_token=%s",
        s.appID,
        s.appSecret,
        currentToken,
    )

    resp, err := http.Get(exchangeURL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var tokenResp struct {
        AccessToken string `json:"access_token"`
        TokenType   string `json:"token_type"`
        ExpiresIn   int    `json:"expires_in"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return err
    }

    // 4. Atualizar credencial
    if err := cred.RefreshOAuthToken(
        tokenResp.AccessToken,
        tokenResp.ExpiresIn,
        s.encryptor,
    ); err != nil {
        return err
    }

    // 5. Salvar
    return s.credentialRepo.Save(cred)
}

func scopesToString(scopes []string) string {
    result := ""
    for i, scope := range scopes {
        if i > 0 {
            result += ","
        }
        result += scope
    }
    return result
}
```

### 4. Criar MetaAPIClient

```go
// internal/application/integration/meta_api_client.go

package integration

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "bytes"

    "github.com/google/uuid"
    "ventros-crm/internal/domain/credential"
)

// MetaAPIClient é um cliente para interagir com as APIs da Meta
type MetaAPIClient struct {
    credentialRepo credential.Repository
    encryptor      credential.Encryptor
    httpClient     *http.Client
    baseURL        string
}

func NewMetaAPIClient(
    credentialRepo credential.Repository,
    encryptor credential.Encryptor,
) *MetaAPIClient {
    return &MetaAPIClient{
        credentialRepo: credentialRepo,
        encryptor:      encryptor,
        httpClient:     &http.Client{},
        baseURL:        "https://graph.facebook.com/v18.0",
    }
}

// SendWhatsAppMessage envia mensagem via WhatsApp Cloud API
func (c *MetaAPIClient) SendWhatsAppMessage(
    ctx context.Context,
    credentialID uuid.UUID,
    to string,
    message string,
) error {
    // 1. Buscar credencial
    cred, err := c.credentialRepo.FindByID(credentialID)
    if err != nil {
        return err
    }

    // 2. Obter access token
    accessToken, err := cred.GetAccessToken(c.encryptor)
    if err != nil {
        return err
    }

    // 3. Obter phone number ID do metadata
    phoneNumberID, _ := cred.GetMetadata("phone_number_id")

    // 4. Montar payload
    payload := map[string]interface{}{
        "messaging_product": "whatsapp",
        "to":                to,
        "type":              "text",
        "text": map[string]string{
            "body": message,
        },
    }

    payloadBytes, _ := json.Marshal(payload)

    // 5. Fazer request
    url := fmt.Sprintf("%s/%s/messages", c.baseURL, phoneNumberID)

    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        url,
        bytes.NewReader(payloadBytes),
    )
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("WhatsApp API error: %s", string(body))
    }

    // 6. Marcar credencial como usada
    cred.MarkAsUsed()
    c.credentialRepo.Save(cred)

    return nil
}

// GetAdAccounts lista ad accounts disponíveis
func (c *MetaAPIClient) GetAdAccounts(
    ctx context.Context,
    credentialID uuid.UUID,
) ([]map[string]interface{}, error) {
    cred, err := c.credentialRepo.FindByID(credentialID)
    if err != nil {
        return nil, err
    }

    accessToken, err := cred.GetAccessToken(c.encryptor)
    if err != nil {
        return nil, err
    }

    businessID, _ := cred.GetMetadata("business_id")

    url := fmt.Sprintf(
        "%s/%s/owned_ad_accounts?access_token=%s",
        c.baseURL,
        businessID,
        accessToken,
    )

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Data []map[string]interface{} `json:"data"`
    }

    json.NewDecoder(resp.Body).Decode(&result)

    return result.Data, nil
}

// SendConversionEvent envia evento para Conversions API
func (c *MetaAPIClient) SendConversionEvent(
    ctx context.Context,
    credentialID uuid.UUID,
    pixelID string,
    eventName string,
    eventData map[string]interface{},
) error {
    cred, err := c.credentialRepo.FindByID(credentialID)
    if err != nil {
        return err
    }

    accessToken, err := cred.GetAccessToken(c.encryptor)
    if err != nil {
        return err
    }

    payload := map[string]interface{}{
        "data": []map[string]interface{}{
            {
                "event_name": eventName,
                "event_time": eventData["event_time"],
                "user_data":  eventData["user_data"],
                "custom_data": eventData["custom_data"],
            },
        },
    }

    payloadBytes, _ := json.Marshal(payload)

    url := fmt.Sprintf(
        "%s/%s/events?access_token=%s",
        c.baseURL,
        pixelID,
        accessToken,
    )

    resp, err := http.Post(url, "application/json", bytes.NewReader(payloadBytes))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("Conversions API error: %s", string(body))
    }

    return nil
}
```

### 5. Criar Handler HTTP

```go
// infrastructure/http/handlers/meta_oauth_handler.go

package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/google/uuid"
    "ventros-crm/internal/application/integration"
    "ventros-crm/internal/domain/credential"
)

type MetaOAuthHandler struct {
    oauthService *integration.MetaOAuthService
}

func NewMetaOAuthHandler(
    oauthService *integration.MetaOAuthService,
) *MetaOAuthHandler {
    return &MetaOAuthHandler{
        oauthService: oauthService,
    }
}

// StartOAuthFlow inicia o fluxo OAuth
// GET /api/v1/integrations/meta/auth/start
func (h *MetaOAuthHandler) StartOAuthFlow(w http.ResponseWriter, r *http.Request) {
    // Gerar state único para CSRF protection
    state := uuid.New().String()

    // Salvar state em sessão/redis (implementar)
    // session.Set("oauth_state", state)

    // Obter URL de autorização
    scopes := credential.CredentialTypeMeta.GetScopes()
    authURL := h.oauthService.GetAuthorizationURL(state, scopes)

    // Retornar URL
    json.NewEncoder(w).Encode(map[string]string{
        "authorization_url": authURL,
        "state": state,
    })
}

// HandleCallback processa o callback OAuth
// POST /api/v1/integrations/meta/auth/callback
func (h *MetaOAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Code  string `json:"code"`
        State string `json:"state"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // 1. Validar state (CSRF protection)
    // savedState := session.Get("oauth_state")
    // if savedState != req.State {
    //     http.Error(w, "Invalid state", http.StatusBadRequest)
    //     return
    // }

    // 2. Trocar code por token
    tokenResp, err := h.oauthService.ExchangeCodeForToken(
        r.Context(),
        req.Code,
    )
    if err != nil {
        http.Error(w, "Token exchange failed", http.StatusInternalServerError)
        return
    }

    // 3. Obter tenant ID do contexto (middleware)
    tenantID := r.Context().Value("tenant_id").(string)

    // 4. Criar credencial
    cred, err := h.oauthService.CreateMetaCredential(
        r.Context(),
        tenantID,
        nil, // ProjectID pode ser passado no request
        tokenResp,
    )
    if err != nil {
        http.Error(w, "Failed to save credential", http.StatusInternalServerError)
        return
    }

    // 5. Retornar sucesso
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":        true,
        "credential_id":  cred.ID(),
        "waba_id":        tokenResp.WABAID,
        "phone_number_id": tokenResp.PhoneNumberID,
        "permissions":    tokenResp.Permissions,
    })
}

// RefreshToken renova um access token
// POST /api/v1/integrations/meta/auth/refresh/:credential_id
func (h *MetaOAuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
    credentialID := r.URL.Query().Get("credential_id")

    id, err := uuid.Parse(credentialID)
    if err != nil {
        http.Error(w, "Invalid credential ID", http.StatusBadRequest)
        return
    }

    if err := h.oauthService.RefreshAccessToken(r.Context(), id); err != nil {
        http.Error(w, "Refresh failed", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]bool{
        "success": true,
    })
}
```

### 6. Adicionar Rotas

```go
// infrastructure/http/routes/routes.go

// Adicionar rotas Meta OAuth
metaOAuthHandler := handlers.NewMetaOAuthHandler(metaOAuthService)

router.HandleFunc("/api/v1/integrations/meta/auth/start",
    authMiddleware(metaOAuthHandler.StartOAuthFlow)).Methods("GET")

router.HandleFunc("/api/v1/integrations/meta/auth/callback",
    authMiddleware(metaOAuthHandler.HandleCallback)).Methods("POST")

router.HandleFunc("/api/v1/integrations/meta/auth/refresh",
    authMiddleware(metaOAuthHandler.RefreshToken)).Methods("POST")
```

---

## App Review e Permissões Avançadas

### Processo de App Review

Para usar as permissões em produção com múltiplos clientes, você **DEVE** passar pelo App Review da Meta.

#### 1. Pré-requisitos

- ✅ **Business Verification** completa
- ✅ App em modo **Development** funcionando
- ✅ Testar fluxo completo com conta de teste
- ✅ Privacy Policy URL configurada
- ✅ Terms of Service URL configurada

#### 2. Permissões que Requerem Review

| Permissão | Nível Padrão | Review Necessário |
|-----------|--------------|-------------------|
| `whatsapp_business_messaging` | Standard | ✅ Advanced |
| `whatsapp_business_management` | Standard | ✅ Advanced |
| `business_management` | Standard | ✅ Advanced |
| `ads_management` | None | ✅ Advanced |
| `ads_read` | Standard | ✅ Advanced |

#### 3. Documentação Necessária

**Para `whatsapp_business_messaging`:**

1. **Screen Recording** (5-15 minutos):
   - Mostrar login no seu app
   - Criar template de mensagem
   - Enviar mensagem via API
   - Mostrar mensagem recebida no WhatsApp

2. **Descrição do Uso**:
```
Nossa plataforma permite que empresas gerenciem comunicação com clientes
via WhatsApp Cloud API. Usuários podem:
- Enviar mensagens de texto, imagens e templates
- Receber mensagens de clientes
- Gerenciar templates de mensagem
- Visualizar histórico de conversas
```

**Para `ads_management`:**

1. **Screen Recording**:
   - Criar campanha publicitária via app
   - Configurar conjunto de anúncios
   - Visualizar métricas

2. **Descrição do Uso**:
```
Nossa plataforma permite gerenciamento de campanhas publicitárias do Facebook.
Usuários podem criar, editar e monitorar anúncios diretamente pelo CRM.
```

#### 4. Limites de Onboarding

| Nível | Clientes por 7 dias | Como Aumentar |
|-------|---------------------|---------------|
| Inicial | 10 novos clientes | Automático após primeiros clientes |
| Intermediário | 50 novos clientes | Manter compliance e qualidade |
| Avançado | 200+ novos clientes | Solicitar aumento via suporte |

#### 5. Submeter App Review

```bash
# Passo a passo
1. Meta App Dashboard → App Review → Permissions and Features
2. Selecionar cada permissão (whatsapp_business_messaging, etc)
3. Click "Request Advanced Access"
4. Upload screen recording
5. Preencher formulário de uso
6. Submit for Review

# Tempo de Review: 2-15 dias úteis
```

#### 6. Checklist Pré-Submissão

- [ ] App funcionando em Development mode
- [ ] Testar Embedded Signup flow completo
- [ ] Verificar que webhooks estão recebendo eventos
- [ ] Screen recording de alta qualidade (1080p mínimo)
- [ ] Privacy Policy publicada e acessível
- [ ] Terms of Service publicados
- [ ] Business verification completa
- [ ] Descrição clara de uso de cada permissão

---

## Webhooks de Deauthorization

Quando um usuário remove a integração Meta, você precisa ser notificado e desativar a credencial.

### 1. Configurar Webhook no Meta Dashboard

```
App Dashboard → Products → Webhooks → Subscribe to:
- permissions (user revoked permissions)
- whatsapp_business_account (account deleted)
```

### 2. Implementar Handler

```go
// infrastructure/http/handlers/meta_webhook_handler.go

package handlers

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "io"
    "net/http"

    "ventros-crm/internal/domain/credential"
)

type MetaWebhookHandler struct {
    credentialRepo credential.Repository
    appSecret      string
}

// HandleWebhook processa webhooks da Meta
// POST /api/v1/integrations/meta/webhooks
func (h *MetaWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    // 1. Verificar assinatura
    signature := r.Header.Get("X-Hub-Signature-256")

    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Invalid body", http.StatusBadRequest)
        return
    }

    if !h.verifySignature(signature, body) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // 2. Processar webhook
    var webhook struct {
        Object string `json:"object"`
        Entry  []struct {
            ID      string `json:"id"`
            Time    int64  `json:"time"`
            Changes []struct {
                Field string                 `json:"field"`
                Value map[string]interface{} `json:"value"`
            } `json:"changes"`
        } `json:"entry"`
    }

    if err := json.Unmarshal(body, &webhook); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // 3. Processar cada evento
    for _, entry := range webhook.Entry {
        for _, change := range entry.Changes {
            h.handleChange(change.Field, change.Value)
        }
    }

    w.WriteHeader(http.StatusOK)
}

func (h *MetaWebhookHandler) handleChange(field string, value map[string]interface{}) {
    switch field {
    case "permissions":
        // Usuário revogou permissões
        h.handlePermissionRevoked(value)

    case "whatsapp_business_account":
        // WABA foi deletado
        h.handleWABADeleted(value)
    }
}

func (h *MetaWebhookHandler) handlePermissionRevoked(value map[string]interface{}) {
    // Extrair user ID ou business ID
    // Buscar credenciais associadas
    // Desativar credenciais

    // Exemplo simplificado:
    businessID := value["business_id"].(string)

    // Buscar todas credenciais com esse business_id
    // credentials := h.credentialRepo.FindByMetadata("business_id", businessID)

    // for _, cred := range credentials {
    //     cred.Deactivate()
    //     h.credentialRepo.Save(cred)
    // }
}

func (h *MetaWebhookHandler) handleWABADeleted(value map[string]interface{}) {
    wabaID := value["waba_id"].(string)

    // Desativar credenciais com esse WABA
    // credentials := h.credentialRepo.FindByMetadata("waba_id", wabaID)

    // for _, cred := range credentials {
    //     cred.Deactivate()
    //     h.credentialRepo.Save(cred)
    // }
}

func (h *MetaWebhookHandler) verifySignature(signature string, body []byte) bool {
    // Remove "sha256=" prefix
    if len(signature) > 7 {
        signature = signature[7:]
    }

    mac := hmac.New(sha256.New, []byte(h.appSecret))
    mac.Write(body)
    expectedMAC := hex.EncodeToString(mac.Sum(nil))

    return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// VerifyWebhook verifica webhook durante configuração
// GET /api/v1/integrations/meta/webhooks
func (h *MetaWebhookHandler) VerifyWebhook(w http.ResponseWriter, r *http.Request) {
    mode := r.URL.Query().Get("hub.mode")
    token := r.URL.Query().Get("hub.verify_token")
    challenge := r.URL.Query().Get("hub.challenge")

    // Verificar token (deve ser configurado no Meta Dashboard)
    expectedToken := "your_verify_token_12345" // Guardar em config

    if mode == "subscribe" && token == expectedToken {
        w.Write([]byte(challenge))
        return
    }

    http.Error(w, "Forbidden", http.StatusForbidden)
}
```

---

## Multi-Tenancy e Gestão de Credenciais

### Isolamento por Tenant

O sistema já possui isolamento por `tenant_id` na tabela `credentials`. Cada cliente terá:

- **Credenciais isoladas** por `tenant_id`
- **Tokens criptografados** individualmente
- **Metadata específica** (WABA ID, Business ID, etc)

### Listar Credenciais de um Tenant

```go
// Buscar todas credenciais Meta de um tenant
func (r *GormCredentialRepository) FindByTenantAndType(
    tenantID string,
    credType credential.CredentialType,
) ([]*credential.Credential, error) {
    var entities []entities.CredentialEntity

    err := r.db.Where(
        "tenant_id = ? AND credential_type = ? AND is_active = true",
        tenantID,
        credType,
    ).Find(&entities).Error

    if err != nil {
        return nil, err
    }

    // Converter para domain
    credentials := make([]*credential.Credential, len(entities))
    for i, entity := range entities {
        credentials[i] = r.toDomain(entity)
    }

    return credentials, nil
}
```

### Selecionar Credencial por Canal

```go
// Quando criar um channel WhatsApp Cloud API, vincular à credencial
type ChannelEntity struct {
    ID           uuid.UUID
    TenantID     string
    ChannelType  string  // "whatsapp_cloud"
    CredentialID *uuid.UUID  // FK para credentials
    // ...
}

// Ao enviar mensagem, usar credencial do canal
func (s *MessageService) SendMessage(channelID uuid.UUID, message string) error {
    channel := s.channelRepo.FindByID(channelID)

    if channel.CredentialID == nil {
        return errors.New("channel has no credential")
    }

    return s.metaClient.SendWhatsAppMessage(
        ctx,
        *channel.CredentialID,
        recipient,
        message,
    )
}
```

---

## Renovação Automática de Tokens

### Background Worker com Temporal

```go
// infrastructure/workflow/credential_refresh_worker.go

package workflow

import (
    "context"
    "time"

    "go.temporal.io/sdk/workflow"
    "ventros-crm/internal/application/integration"
)

// CredentialRefreshWorkflow renova tokens automaticamente
func CredentialRefreshWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)

    // Configurar timer para rodar diariamente
    for {
        // 1. Buscar credenciais que expiram em 7 dias
        var credentials []string

        err := workflow.ExecuteActivity(
            ctx,
            "FindExpiringCredentials",
            7, // dias
        ).Get(ctx, &credentials)

        if err != nil {
            logger.Error("Failed to find expiring credentials", "error", err)
            continue
        }

        // 2. Renovar cada credencial
        for _, credID := range credentials {
            workflow.ExecuteActivity(
                ctx,
                "RefreshCredential",
                credID,
            )
        }

        // 3. Aguardar 24h
        workflow.Sleep(ctx, 24*time.Hour)
    }
}

// FindExpiringCredentials activity
func FindExpiringCredentials(ctx context.Context, daysUntilExpiry int) ([]string, error) {
    // Query credentials expirando em X dias
    // SELECT id FROM credentials
    // WHERE expires_at BETWEEN NOW() AND NOW() + INTERVAL 'X days'
    // AND is_active = true

    return []string{}, nil
}

// RefreshCredential activity
func RefreshCredential(ctx context.Context, credentialID string) error {
    // Chamar metaOAuthService.RefreshAccessToken(credentialID)
    return nil
}
```

### Cron Job Alternativo (Simples)

```go
// cmd/credential-refresher/main.go

package main

import (
    "context"
    "log"
    "time"
)

func main() {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        refreshExpiringCredentials()
    }
}

func refreshExpiringCredentials() {
    ctx := context.Background()

    // 1. Buscar credenciais que expiram em 7 dias
    credentials := findExpiringCredentials(ctx, 7)

    // 2. Renovar cada uma
    for _, cred := range credentials {
        if err := metaOAuthService.RefreshAccessToken(ctx, cred.ID); err != nil {
            log.Printf("Failed to refresh credential %s: %v", cred.ID, err)
        }
    }
}
```

---

## Segurança e Boas Práticas

### 1. Armazenamento Seguro

✅ **Já implementado no sistema:**

- Tokens armazenados criptografados (AES-256-GCM)
- Nonce único por valor
- Chave de criptografia em variável de ambiente

```go
// Nunca logar tokens em logs
logger.Info("Credential created",
    "credential_id", cred.ID(),
    // ❌ NÃO: "access_token", token
)

// Sempre usar encryptor
accessToken, err := cred.GetAccessToken(encryptor)
```

### 2. HTTPS Obrigatório

```go
// Forçar HTTPS em produção
if os.Getenv("ENV") == "production" {
    router.Use(middleware.ForceHTTPS())
}
```

### 3. Validação de State (CSRF Protection)

```go
// Sempre validar state no callback OAuth
func (h *MetaOAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    state := r.FormValue("state")

    savedState := session.Get("oauth_state")
    if state != savedState {
        http.Error(w, "Invalid state parameter", http.StatusBadRequest)
        return
    }

    // Limpar state usado
    session.Delete("oauth_state")

    // ... continuar
}
```

### 4. Rate Limiting

```go
// Proteger endpoints OAuth contra abuse
router.Use(middleware.RateLimit(
    "oauth",
    10,  // 10 requests
    time.Minute, // por minuto
))
```

### 5. Auditoria

```go
// Logar todas operações com credenciais
type CredentialAuditLog struct {
    CredentialID uuid.UUID
    Action       string // "created", "used", "refreshed", "deactivated"
    TenantID     string
    UserID       *uuid.UUID
    IPAddress    string
    Timestamp    time.Time
}

// Salvar em tabela separada
func (s *MetaOAuthService) CreateMetaCredential(...) {
    // ... criar credencial ...

    auditLog := CredentialAuditLog{
        CredentialID: cred.ID(),
        Action:       "created",
        TenantID:     tenantID,
        Timestamp:    time.Now(),
    }

    s.auditRepo.Save(auditLog)
}
```

### 6. Rotação de Secrets

```go
// App Secret deve ser rotacionado periodicamente
// Manter versão antiga por 30 dias durante transição

type AppSecretRotation struct {
    CurrentSecret  string
    PreviousSecret string
    RotatedAt      time.Time
}

func verifySignature(signature string, body []byte, secrets AppSecretRotation) bool {
    // Tentar com secret atual
    if verify(signature, body, secrets.CurrentSecret) {
        return true
    }

    // Se falhar, tentar com secret anterior (durante período de transição)
    if time.Since(secrets.RotatedAt) < 30*24*time.Hour {
        return verify(signature, body, secrets.PreviousSecret)
    }

    return false
}
```

---

## Configuração (Environment Variables)

```bash
# .env

# Meta OAuth Configuration
META_APP_ID=your_app_id_here
META_APP_SECRET=your_app_secret_here
META_REDIRECT_URI=https://ventros-crm.com/api/v1/integrations/meta/auth/callback
META_CONFIGURATION_ID=your_embedded_signup_config_id
META_WEBHOOK_VERIFY_TOKEN=random_secure_token_12345

# Encryption
CREDENTIAL_ENCRYPTION_KEY=base64_encoded_32_byte_key

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/ventros_crm

# API Base URLs
META_GRAPH_API_VERSION=v18.0
META_GRAPH_API_BASE_URL=https://graph.facebook.com
```

---

## Resumo da Implementação

### ✅ O que será desenvolvido:

1. **Domain Layer:**
   - ✅ Credential aggregate (já existe)
   - ✅ MetaCredential value object (NOVO)
   - ✅ Tipos de credencial Meta atualizados

2. **Application Layer:**
   - ✅ MetaOAuthService (NOVO)
   - ✅ MetaAPIClient (NOVO)
   - ✅ CredentialRefreshService (NOVO)

3. **Infrastructure Layer:**
   - ✅ MetaOAuthHandler (NOVO)
   - ✅ MetaWebhookHandler (NOVO)
   - ✅ Rotas OAuth (NOVO)

4. **Frontend:**
   - ✅ Botão "Conectar Meta"
   - ✅ Integração Facebook SDK
   - ✅ Tela de gerenciamento de credenciais

5. **DevOps:**
   - ✅ Variáveis de ambiente
   - ✅ Webhook endpoint público (HTTPS)
   - ✅ Background worker para refresh

### 🎯 Próximos Passos:

1. Criar app no Meta Developers
2. Implementar domain objects
3. Implementar application services
4. Implementar handlers HTTP
5. Adicionar rotas
6. Desenvolver frontend
7. Testar em development
8. Submeter App Review
9. Deploy em produção

---

## Referências

- [Meta Graph API Documentation](https://developers.facebook.com/docs/graph-api/)
- [WhatsApp Cloud API](https://developers.facebook.com/docs/whatsapp/cloud-api/)
- [Embedded Signup](https://developers.facebook.com/docs/whatsapp/embedded-signup/)
- [Facebook Login for Business](https://developers.facebook.com/docs/facebook-login/facebook-login-for-business/)
- [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2)
- [App Review Guidelines](https://developers.facebook.com/docs/app-review/)
