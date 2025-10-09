# 🔐 Meta WhatsApp Cloud API - Guia de Arquitetura e Implementação

**Data:** 2025-10-09
**Escopo:** Credential Management, OAuth Flow, Notifier Pattern, Meta Integration

---

## 📋 Sumário Executivo

### Decisões Arquiteturais

| Decisão | Escolha | Justificativa |
|---------|---------|---------------|
| **OutboundMessage** | ❌ Remover como aggregate | Mensagens já existem em `messages` table |
| **Outbox Types** | ✅ Adicionar `WEBHOOK`, `META_CONVERSION` | Eventos especializados para integrações |
| **Credentials** | ✅ Domain model completo | Gerenciamento seguro de tokens OAuth |
| **Notifier** | ✅ Strategy Pattern | Enviar eventos para Meta, Google Ads, etc |

---

## 🎯 1. Arquitetura Proposta

### 📊 **Fluxo de Mensagens (Atual vs Proposto)**

#### ❌ Atual (Redundante)
```
Automation → OutboundMessage (tabela separada)
                 ↓
             Worker envia
                 ↓
          Salva em Messages
```

#### ✅ Proposto (Simplificado)
```
Automation → Message (com fromMe=true)
                 ↓
             OutboxEvent (tipo: MESSAGE_DELIVERY)
                 ↓
             Worker envia via WAHA/Meta
                 ↓
             Atualiza Message (delivered/read)
```

### 🔄 **Nova Arquitetura de Eventos**

```
┌─────────────────────────────────────────────────────┐
│              OutboxEvent (Unified)                   │
├─────────────────────────────────────────────────────┤
│ Types:                                               │
│  • DOMAIN_EVENT      (contact.created, etc)         │
│  • MESSAGE_DELIVERY  (envio de mensagem)            │
│  • WEBHOOK           (notificações externas)        │
│  • META_CONVERSION   (conversão de ads)             │
│  • GOOGLE_ADS        (conversão Google)             │
└─────────────────────────────────────────────────────┘
                         ↓
            ┌────────────┴────────────┐
            │                         │
┌───────────▼──────────┐   ┌─────────▼──────────┐
│  Domain Event Bus    │   │   Notifier System  │
│  (RabbitMQ)          │   │   (External APIs)  │
└──────────────────────┘   └────────────────────┘
```

---

## 🏗️ 2. Domain Model: Credentials

### 📁 Estrutura de Arquivos

```
internal/domain/credential/
├── credential.go              # Aggregate Root
├── credential_type.go         # Value Object (Enum)
├── encrypted_value.go         # Value Object (Encrypted)
├── oauth_token.go             # Value Object
├── repository.go              # Interface
├── events.go                  # Domain Events
└── credential_test.go         # Testes unitários

infrastructure/persistence/entities/
└── credential.go              # Entity (ORM)

infrastructure/persistence/
└── gorm_credential_repository.go

infrastructure/crypto/
├── encryptor.go               # Interface
└── aes_encryptor.go           # AES-256-GCM
```

---

### 🔐 **Credential Aggregate**

```go
// internal/domain/credential/credential.go
package credential

import (
	"errors"
	"time"
	"github.com/google/uuid"
)

// Credential representa credenciais criptografadas para integrações externas
type Credential struct {
	id             uuid.UUID
	tenantID       string
	projectID      *uuid.UUID       // opcional - pode ser global ao tenant
	credentialType CredentialType
	name           string
	description    string

	// Dados criptografados
	encryptedValue EncryptedValue

	// OAuth specific (quando aplicável)
	oauthToken *OAuthToken

	// Metadata
	metadata   map[string]interface{}
	isActive   bool
	expiresAt  *time.Time
	lastUsedAt *time.Time

	createdAt time.Time
	updatedAt time.Time

	events []DomainEvent
}

// NewCredential cria uma nova credencial
func NewCredential(
	tenantID string,
	credentialType CredentialType,
	name string,
	plainValue string,
	encryptor Encryptor,
) (*Credential, error) {
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if plainValue == "" {
		return nil, errors.New("value cannot be empty")
	}

	// Valida tipo de credencial
	if !credentialType.IsValid() {
		return nil, errors.New("invalid credential type")
	}

	// Criptografa o valor
	encryptedValue, err := encryptor.Encrypt(plainValue)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt credential")
	}

	now := time.Now()
	cred := &Credential{
		id:             uuid.New(),
		tenantID:       tenantID,
		credentialType: credentialType,
		name:           name,
		encryptedValue: encryptedValue,
		metadata:       make(map[string]interface{}),
		isActive:       true,
		createdAt:      now,
		updatedAt:      now,
		events:         []DomainEvent{},
	}

	cred.addEvent(CredentialCreatedEvent{
		CredentialID:   cred.id,
		TenantID:       tenantID,
		CredentialType: credentialType,
		Name:           name,
		CreatedAt:      now,
	})

	return cred, nil
}

// SetOAuthToken armazena tokens OAuth (Access + Refresh)
func (c *Credential) SetOAuthToken(
	accessToken string,
	refreshToken string,
	expiresIn int,
	encryptor Encryptor,
) error {
	if c.credentialType != CredentialTypeMetaWhatsApp &&
	   c.credentialType != CredentialTypeGoogleAds {
		return errors.New("OAuth tokens only valid for OAuth credential types")
	}

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	token, err := NewOAuthToken(accessToken, refreshToken, expiresAt, encryptor)
	if err != nil {
		return errors.Wrap(err, "failed to create OAuth token")
	}

	c.oauthToken = token
	c.expiresAt = &expiresAt
	c.updatedAt = time.Now()

	c.addEvent(OAuthTokenRefreshedEvent{
		CredentialID: c.id,
		ExpiresAt:    expiresAt,
		RefreshedAt:  c.updatedAt,
	})

	return nil
}

// RefreshOAuthToken renova o access token usando refresh token
func (c *Credential) RefreshOAuthToken(
	newAccessToken string,
	expiresIn int,
	encryptor Encryptor,
) error {
	if c.oauthToken == nil {
		return errors.New("no OAuth token to refresh")
	}

	if err := c.oauthToken.Refresh(newAccessToken, expiresIn, encryptor); err != nil {
		return err
	}

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	c.expiresAt = &expiresAt
	c.updatedAt = time.Now()

	c.addEvent(OAuthTokenRefreshedEvent{
		CredentialID: c.id,
		ExpiresAt:    expiresAt,
		RefreshedAt:  c.updatedAt,
	})

	return nil
}

// IsExpired verifica se a credencial expirou
func (c *Credential) IsExpired() bool {
	if c.expiresAt == nil {
		return false
	}
	return time.Now().After(*c.expiresAt)
}

// NeedsRefresh verifica se precisa renovar (30 min antes de expirar)
func (c *Credential) NeedsRefresh() bool {
	if c.expiresAt == nil || c.oauthToken == nil {
		return false
	}
	return time.Now().Add(30 * time.Minute).After(*c.expiresAt)
}

// Decrypt retorna o valor descriptografado
func (c *Credential) Decrypt(encryptor Encryptor) (string, error) {
	return encryptor.Decrypt(c.encryptedValue)
}

// GetAccessToken retorna o access token OAuth descriptografado
func (c *Credential) GetAccessToken(encryptor Encryptor) (string, error) {
	if c.oauthToken == nil {
		return "", errors.New("no OAuth token available")
	}
	return c.oauthToken.GetAccessToken(encryptor)
}

// MarkAsUsed atualiza lastUsedAt
func (c *Credential) MarkAsUsed() {
	now := time.Now()
	c.lastUsedAt = &now
	c.updatedAt = now
}

// Deactivate desativa a credencial
func (c *Credential) Deactivate() {
	if c.isActive {
		c.isActive = false
		c.updatedAt = time.Now()

		c.addEvent(CredentialDeactivatedEvent{
			CredentialID:  c.id,
			DeactivatedAt: c.updatedAt,
		})
	}
}

// Activate ativa a credencial
func (c *Credential) Activate() {
	if !c.isActive {
		c.isActive = true
		c.updatedAt = time.Now()

		c.addEvent(CredentialActivatedEvent{
			CredentialID: c.id,
			ActivatedAt:  c.updatedAt,
		})
	}
}

// Getters
func (c *Credential) ID() uuid.UUID            { return c.id }
func (c *Credential) TenantID() string         { return c.tenantID }
func (c *Credential) ProjectID() *uuid.UUID    { return c.projectID }
func (c *Credential) Type() CredentialType     { return c.credentialType }
func (c *Credential) Name() string             { return c.name }
func (c *Credential) Description() string      { return c.description }
func (c *Credential) IsActive() bool           { return c.isActive }
func (c *Credential) ExpiresAt() *time.Time    { return c.expiresAt }
func (c *Credential) LastUsedAt() *time.Time   { return c.lastUsedAt }
func (c *Credential) CreatedAt() time.Time     { return c.createdAt }
func (c *Credential) UpdatedAt() time.Time     { return c.updatedAt }

// Domain Events
func (c *Credential) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, c.events...)
}

func (c *Credential) ClearEvents() {
	c.events = []DomainEvent{}
}

func (c *Credential) addEvent(event DomainEvent) {
	c.events = append(c.events, event)
}
```

---

### 🎭 **Credential Types (Value Object)**

```go
// internal/domain/credential/credential_type.go
package credential

type CredentialType string

const (
	// Meta Integrations
	CredentialTypeMetaWhatsApp     CredentialType = "meta_whatsapp_cloud"
	CredentialTypeMetaAds          CredentialType = "meta_ads"
	CredentialTypeMetaConversions  CredentialType = "meta_conversions_api"

	// Google Integrations
	CredentialTypeGoogleAds        CredentialType = "google_ads"
	CredentialTypeGoogleAnalytics  CredentialType = "google_analytics"

	// Other Integrations
	CredentialTypeWebhook          CredentialType = "webhook_auth"
	CredentialTypeAPIKey           CredentialType = "api_key"
	CredentialTypeBasicAuth        CredentialType = "basic_auth"

	// Internal
	CredentialTypeWAHA             CredentialType = "waha_instance"
)

// IsValid verifica se o tipo é válido
func (t CredentialType) IsValid() bool {
	switch t {
	case CredentialTypeMetaWhatsApp,
	     CredentialTypeMetaAds,
	     CredentialTypeMetaConversions,
	     CredentialTypeGoogleAds,
	     CredentialTypeGoogleAnalytics,
	     CredentialTypeWebhook,
	     CredentialTypeAPIKey,
	     CredentialTypeBasicAuth,
	     CredentialTypeWAHA:
		return true
	default:
		return false
	}
}

// RequiresOAuth verifica se o tipo requer OAuth
func (t CredentialType) RequiresOAuth() bool {
	switch t {
	case CredentialTypeMetaWhatsApp,
	     CredentialTypeMetaAds,
	     CredentialTypeMetaConversions,
	     CredentialTypeGoogleAds,
	     CredentialTypeGoogleAnalytics:
		return true
	default:
		return false
	}
}

// GetScopes retorna os scopes OAuth necessários
func (t CredentialType) GetScopes() []string {
	switch t {
	case CredentialTypeMetaWhatsApp:
		return []string{
			"whatsapp_business_management",
			"whatsapp_business_messaging",
		}
	case CredentialTypeMetaAds:
		return []string{
			"ads_management",
			"ads_read",
		}
	case CredentialTypeMetaConversions:
		return []string{
			"ads_management",
		}
	case CredentialTypeGoogleAds:
		return []string{
			"https://www.googleapis.com/auth/adwords",
		}
	default:
		return []string{}
	}
}
```

---

### 🔒 **Encrypted Value (Value Object)**

```go
// internal/domain/credential/encrypted_value.go
package credential

import "encoding/base64"

// EncryptedValue representa um valor criptografado
type EncryptedValue struct {
	ciphertext string // Base64 encoded
	nonce      string // Base64 encoded (for AES-GCM)
}

// NewEncryptedValue cria um valor criptografado
func NewEncryptedValue(ciphertext, nonce string) EncryptedValue {
	return EncryptedValue{
		ciphertext: ciphertext,
		nonce:      nonce,
	}
}

// Ciphertext retorna o texto cifrado
func (e EncryptedValue) Ciphertext() string {
	return e.ciphertext
}

// Nonce retorna o nonce
func (e EncryptedValue) Nonce() string {
	return e.nonce
}

// Encryptor interface para criptografia
type Encryptor interface {
	Encrypt(plaintext string) (EncryptedValue, error)
	Decrypt(encrypted EncryptedValue) (string, error)
}
```

---

### 🎫 **OAuth Token (Value Object)**

```go
// internal/domain/credential/oauth_token.go
package credential

import (
	"errors"
	"time"
)

// OAuthToken representa um token OAuth criptografado
type OAuthToken struct {
	encryptedAccessToken  EncryptedValue
	encryptedRefreshToken EncryptedValue
	expiresAt             time.Time
	tokenType             string // "Bearer"
}

// NewOAuthToken cria um novo token OAuth
func NewOAuthToken(
	accessToken string,
	refreshToken string,
	expiresAt time.Time,
	encryptor Encryptor,
) (*OAuthToken, error) {
	if accessToken == "" {
		return nil, errors.New("access token cannot be empty")
	}

	encryptedAccess, err := encryptor.Encrypt(accessToken)
	if err != nil {
		return nil, err
	}

	var encryptedRefresh EncryptedValue
	if refreshToken != "" {
		encryptedRefresh, err = encryptor.Encrypt(refreshToken)
		if err != nil {
			return nil, err
		}
	}

	return &OAuthToken{
		encryptedAccessToken:  encryptedAccess,
		encryptedRefreshToken: encryptedRefresh,
		expiresAt:             expiresAt,
		tokenType:             "Bearer",
	}, nil
}

// GetAccessToken retorna o access token descriptografado
func (t *OAuthToken) GetAccessToken(encryptor Encryptor) (string, error) {
	return encryptor.Decrypt(t.encryptedAccessToken)
}

// GetRefreshToken retorna o refresh token descriptografado
func (t *OAuthToken) GetRefreshToken(encryptor Encryptor) (string, error) {
	if t.encryptedRefreshToken.Ciphertext() == "" {
		return "", errors.New("no refresh token available")
	}
	return encryptor.Decrypt(t.encryptedRefreshToken)
}

// IsExpired verifica se o token expirou
func (t *OAuthToken) IsExpired() bool {
	return time.Now().After(t.expiresAt)
}

// Refresh atualiza o access token
func (t *OAuthToken) Refresh(
	newAccessToken string,
	expiresIn int,
	encryptor Encryptor,
) error {
	encryptedAccess, err := encryptor.Encrypt(newAccessToken)
	if err != nil {
		return err
	}

	t.encryptedAccessToken = encryptedAccess
	t.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return nil
}

// ExpiresAt retorna quando o token expira
func (t *OAuthToken) ExpiresAt() time.Time {
	return t.expiresAt
}
```

---

## 🔐 3. Implementação de Criptografia

### 🛡️ **AES-256-GCM Encryptor**

```go
// infrastructure/crypto/aes_encryptor.go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/cawolfkreo/ventros-crm/internal/domain/credential"
)

// AESEncryptor implementa criptografia AES-256-GCM
type AESEncryptor struct {
	key []byte // 32 bytes para AES-256
}

// NewAESEncryptor cria um novo encryptor
// IMPORTANTE: A key DEVE vir de variável de ambiente
// NUNCA hardcode a chave!
func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
	if len(key) != 32 {
		return nil, errors.New("AES-256 requires a 32-byte key")
	}

	return &AESEncryptor{key: key}, nil
}

// Encrypt criptografa um plaintext
func (e *AESEncryptor) Encrypt(plaintext string) (credential.EncryptedValue, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return credential.EncryptedValue{}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return credential.EncryptedValue{}, err
	}

	// Gera nonce aleatório
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return credential.EncryptedValue{}, err
	}

	// Criptografa
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	return credential.NewEncryptedValue(
		base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(nonce),
	), nil
}

// Decrypt descriptografa um valor criptografado
func (e *AESEncryptor) Decrypt(encrypted credential.EncryptedValue) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Decodifica base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.Ciphertext())
	if err != nil {
		return "", err
	}

	nonce, err := base64.StdEncoding.DecodeString(encrypted.Nonce())
	if err != nil {
		return "", err
	}

	// Descriptografa
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
```

### 🔑 **Gerenciamento de Chave**

```go
// infrastructure/config/encryption.go
package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
)

// GetEncryptionKey obtém a chave de criptografia
// ORDEM DE PRIORIDADE:
// 1. Variável de ambiente ENCRYPTION_KEY
// 2. AWS Secrets Manager / GCP Secret Manager
// 3. Arquivo .env (APENAS DESENVOLVIMENTO)
func GetEncryptionKey() ([]byte, error) {
	// 1. Tenta variável de ambiente
	if keyBase64 := os.Getenv("ENCRYPTION_KEY"); keyBase64 != "" {
		key, err := base64.StdEncoding.DecodeString(keyBase64)
		if err != nil {
			return nil, errors.New("invalid ENCRYPTION_KEY format")
		}
		if len(key) != 32 {
			return nil, errors.New("ENCRYPTION_KEY must be 32 bytes (AES-256)")
		}
		return key, nil
	}

	// 2. TODO: Implementar AWS Secrets Manager
	// key, err := getFromAWSSecretsManager("ventros-crm/encryption-key")

	// 3. Se estiver em desenvolvimento, gera chave temporária
	if os.Getenv("ENV") == "development" {
		return generateTempKey()
	}

	return nil, errors.New("ENCRYPTION_KEY not configured")
}

// GenerateNewKey gera uma nova chave AES-256 (use apenas para setup inicial)
func GenerateNewKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

func generateTempKey() ([]byte, error) {
	log.Println("⚠️  WARNING: Using temporary encryption key (DEVELOPMENT ONLY)")
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
```

---

## 🚀 4. Meta WhatsApp Cloud API - OAuth Flow

### 📋 **OAuth Flow Completo**

```
┌──────────────────────────────────────────────────────────────┐
│                  Meta OAuth Flow (PKCE)                       │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  1. User → Frontend: Clica "Conectar WhatsApp"              │
│                ↓                                              │
│  2. Frontend → Backend: GET /api/oauth/meta/authorize        │
│                ↓                                              │
│  3. Backend gera:                                             │
│      • state (random token)                                   │
│      • code_verifier (PKCE)                                   │
│      • code_challenge (SHA256(code_verifier))                 │
│                ↓                                              │
│  4. Backend → Frontend: Redirect URL                          │
│                ↓                                              │
│  5. Frontend → Meta: Redirect to Facebook Login              │
│      https://www.facebook.com/v18.0/dialog/oauth?             │
│        client_id={APP_ID}                                     │
│        &redirect_uri={REDIRECT_URI}                           │
│        &scope=whatsapp_business_messaging,                    │
│               whatsapp_business_management                    │
│        &state={STATE}                                         │
│        &code_challenge={CODE_CHALLENGE}                       │
│        &code_challenge_method=S256                            │
│                ↓                                              │
│  6. User aprova permissões na Meta                            │
│                ↓                                              │
│  7. Meta → Frontend: Redirect de volta com code               │
│      {REDIRECT_URI}?code={AUTH_CODE}&state={STATE}           │
│                ↓                                              │
│  8. Frontend → Backend: POST /api/oauth/meta/callback         │
│      { code, state }                                          │
│                ↓                                              │
│  9. Backend valida state                                      │
│                ↓                                              │
│ 10. Backend → Meta: Exchange code for token                   │
│      POST https://graph.facebook.com/v18.0/oauth/access_token │
│      { code, client_id, client_secret, code_verifier }        │
│                ↓                                              │
│ 11. Meta → Backend: Returns tokens                            │
│      { access_token, refresh_token, expires_in }              │
│                ↓                                              │
│ 12. Backend:                                                  │
│      • Cria Credential (criptografado)                        │
│      • Salva no banco                                         │
│      • Retorna sucesso                                        │
│                ↓                                              │
│ 13. Frontend: Mostra "Conectado com sucesso!"                │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

### 🔧 **Implementação em Go**

```go
// internal/application/oauth/meta_oauth_service.go
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cawolfkreo/ventros-crm/internal/domain/credential"
)

type MetaOAuthService struct {
	appID          string
	appSecret      string
	redirectURI    string
	credentialRepo credential.Repository
	encryptor      credential.Encryptor
	stateStore     StateStore // Redis ou in-memory
}

// GenerateAuthURL gera URL de autorização
func (s *MetaOAuthService) GenerateAuthURL(tenantID string) (string, error) {
	// Gera state (anti-CSRF)
	state, err := generateRandomString(32)
	if err != nil {
		return "", err
	}

	// Gera code_verifier (PKCE)
	codeVerifier, err := generateRandomString(64)
	if err != nil {
		return "", err
	}

	// Gera code_challenge
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Salva state e code_verifier temporariamente (5 min)
	if err := s.stateStore.Save(state, OAuthState{
		TenantID:     tenantID,
		CodeVerifier: codeVerifier,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
	}); err != nil {
		return "", err
	}

	// Monta URL
	params := url.Values{}
	params.Set("client_id", s.appID)
	params.Set("redirect_uri", s.redirectURI)
	params.Set("scope", "whatsapp_business_messaging,whatsapp_business_management")
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	authURL := fmt.Sprintf(
		"https://www.facebook.com/v18.0/dialog/oauth?%s",
		params.Encode(),
	)

	return authURL, nil
}

// HandleCallback processa callback do OAuth
func (s *MetaOAuthService) HandleCallback(
	ctx context.Context,
	code string,
	state string,
) (*credential.Credential, error) {
	// 1. Valida state
	oauthState, err := s.stateStore.Get(state)
	if err != nil {
		return nil, errors.New("invalid or expired state")
	}

	// 2. Troca code por token
	tokens, err := s.exchangeCodeForToken(code, oauthState.CodeVerifier)
	if err != nil {
		return nil, err
	}

	// 3. Cria credencial criptografada
	cred, err := credential.NewCredential(
		oauthState.TenantID,
		credential.CredentialTypeMetaWhatsApp,
		"Meta WhatsApp Cloud API",
		s.appSecret, // valor base (não usado diretamente)
		s.encryptor,
	)
	if err != nil {
		return nil, err
	}

	// 4. Adiciona tokens OAuth
	err = cred.SetOAuthToken(
		tokens.AccessToken,
		tokens.RefreshToken,
		tokens.ExpiresIn,
		s.encryptor,
	)
	if err != nil {
		return nil, err
	}

	// 5. Salva no banco
	if err := s.credentialRepo.Save(cred); err != nil {
		return nil, err
	}

	// 6. Remove state
	s.stateStore.Delete(state)

	return cred, nil
}

// exchangeCodeForToken troca authorization code por tokens
func (s *MetaOAuthService) exchangeCodeForToken(
	code string,
	codeVerifier string,
) (*TokenResponse, error) {
	params := url.Values{}
	params.Set("client_id", s.appID)
	params.Set("client_secret", s.appSecret)
	params.Set("code", code)
	params.Set("redirect_uri", s.redirectURI)
	params.Set("code_verifier", codeVerifier)

	tokenURL := "https://graph.facebook.com/v18.0/oauth/access_token"

	resp, err := http.PostForm(tokenURL, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// RefreshToken renova access token usando refresh token
func (s *MetaOAuthService) RefreshToken(
	ctx context.Context,
	credentialID uuid.UUID,
) error {
	// 1. Busca credencial
	cred, err := s.credentialRepo.FindByID(credentialID)
	if err != nil {
		return err
	}

	// 2. Obtém refresh token
	refreshToken, err := cred.GetRefreshToken(s.encryptor)
	if err != nil {
		return err
	}

	// 3. Chama Meta API para renovar
	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("refresh_token", refreshToken)
	params.Set("client_id", s.appID)
	params.Set("client_secret", s.appSecret)

	tokenURL := "https://graph.facebook.com/v18.0/oauth/access_token"

	resp, err := http.PostForm(tokenURL, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	// 4. Atualiza credencial
	err = cred.RefreshOAuthToken(
		tokenResp.AccessToken,
		tokenResp.ExpiresIn,
		s.encryptor,
	)
	if err != nil {
		return err
	}

	// 5. Salva
	return s.credentialRepo.Save(cred)
}

// Helpers

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type OAuthState struct {
	TenantID     string
	CodeVerifier string
	ExpiresAt    time.Time
}
```

---

## 🔔 5. Notifier Pattern - Conversões Meta

### 🎯 **Arquitetura de Notifiers**

```
┌─────────────────────────────────────────────────────────────┐
│                   Notifier System                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  OutboxEvent (tipo: META_CONVERSION)                         │
│       ↓                                                      │
│  NotifierRegistry                                            │
│       ↓                                                      │
│  ┌──────────────┬──────────────┬──────────────┐             │
│  │ MetaNotifier │ GoogleNotifier│ CustomNotifier│            │
│  └──────┬───────┴──────┬───────┴──────┬────────┘             │
│         ↓              ↓              ↓                      │
│   Meta Conv API   Google Ads API  Webhook                    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 📝 **Interface e Implementação**

```go
// internal/domain/notifier/notifier.go
package notifier

import (
	"context"
	"github.com/google/uuid"
)

// Notifier interface para envio de eventos externos
type Notifier interface {
	// Notify envia evento para sistema externo
	Notify(ctx context.Context, event NotificationEvent) error

	// Type retorna o tipo de notifier
	Type() NotifierType

	// Validate valida se o evento pode ser enviado
	Validate(event NotificationEvent) error
}

// NotificationEvent representa um evento a ser notificado
type NotificationEvent struct {
	EventID     uuid.UUID
	TenantID    string
	EventType   string                 // "conversion", "lead", "purchase"
	EventData   map[string]interface{} // dados do evento
	CredentialID uuid.UUID             // credencial a usar
}

// NotifierType tipos de notifiers
type NotifierType string

const (
	NotifierTypeMeta   NotifierType = "meta_conversions"
	NotifierTypeGoogle NotifierType = "google_ads"
	NotifierTypeWebhook NotifierType = "webhook"
)
```

```go
// internal/application/notifier/meta_notifier.go
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cawolfkreo/ventros-crm/internal/domain/credential"
	"github.com/cawolfkreo/ventros-crm/internal/domain/notifier"
)

// MetaConversionsNotifier envia conversões para Meta Conversions API
type MetaConversionsNotifier struct {
	credentialRepo credential.Repository
	encryptor      credential.Encryptor
	pixelID        string // Meta Pixel ID
	httpClient     *http.Client
}

func NewMetaConversionsNotifier(
	credentialRepo credential.Repository,
	encryptor credential.Encryptor,
	pixelID string,
) *MetaConversionsNotifier {
	return &MetaConversionsNotifier{
		credentialRepo: credentialRepo,
		encryptor:      encryptor,
		pixelID:        pixelID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (n *MetaConversionsNotifier) Type() notifier.NotifierType {
	return notifier.NotifierTypeMeta
}

func (n *MetaConversionsNotifier) Validate(event notifier.NotificationEvent) error {
	// Valida campos obrigatórios
	required := []string{"event_name", "event_time", "user_data"}
	for _, field := range required {
		if _, ok := event.EventData[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

func (n *MetaConversionsNotifier) Notify(
	ctx context.Context,
	event notifier.NotificationEvent,
) error {
	// 1. Busca credencial
	cred, err := n.credentialRepo.FindByID(event.CredentialID)
	if err != nil {
		return fmt.Errorf("failed to get credential: %w", err)
	}

	// 2. Obtém access token
	accessToken, err := cred.GetAccessToken(n.encryptor)
	if err != nil {
		return fmt.Errorf("failed to decrypt access token: %w", err)
	}

	// 3. Monta payload da Conversions API
	payload := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"event_name":      event.EventData["event_name"],
				"event_time":      event.EventData["event_time"],
				"event_id":        event.EventID.String(),
				"event_source_url": event.EventData["event_source_url"],
				"user_data":       event.EventData["user_data"],
				"custom_data":     event.EventData["custom_data"],
				"action_source":   "website", // ou "app", "phone_call", etc
			},
		},
	}

	// 4. Envia para Meta
	url := fmt.Sprintf(
		"https://graph.facebook.com/v18.0/%s/events?access_token=%s",
		n.pixelID,
		accessToken,
	)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send conversion: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Meta API returned status %d", resp.StatusCode)
	}

	// 5. Marca credencial como usada
	cred.MarkAsUsed()
	_ = n.credentialRepo.Save(cred)

	return nil
}
```

---

## 📊 6. OutboxEvent Types (Expandido)

### 🔧 **Novos Tipos de Outbox**

```go
// internal/domain/outbox/event_type.go
package outbox

type EventType string

const (
	// Domain Events (existentes)
	EventTypeDomain          EventType = "domain_event"

	// Message Delivery (novo)
	EventTypeMessageDelivery EventType = "message_delivery"

	// External Notifications (novo)
	EventTypeWebhook         EventType = "webhook"
	EventTypeMetaConversion  EventType = "meta_conversion"
	EventTypeGoogleAds       EventType = "google_ads_conversion"

	// Custom (novo)
	EventTypeCustom          EventType = "custom"
)

// GetNotifierType retorna o tipo de notifier para um event type
func (t EventType) GetNotifierType() (notifier.NotifierType, bool) {
	switch t {
	case EventTypeMetaConversion:
		return notifier.NotifierTypeMeta, true
	case EventTypeGoogleAds:
		return notifier.NotifierTypeGoogle, true
	case EventTypeWebhook:
		return notifier.NotifierTypeWebhook, true
	default:
		return "", false
	}
}
```

---

## 🎯 7. Migration: Adicionar Tipos ao Outbox

```sql
-- infrastructure/database/migrations/000022_add_outbox_event_types.up.sql

-- Adiciona coluna event_category
ALTER TABLE outbox_events
ADD COLUMN IF NOT EXISTS event_category VARCHAR(50) NOT NULL DEFAULT 'domain_event';

-- Cria índice
CREATE INDEX IF NOT EXISTS idx_outbox_event_category
ON outbox_events(event_category, status)
WHERE status IN ('pending', 'failed');

-- Adiciona coluna credential_id (para notifiers)
ALTER TABLE outbox_events
ADD COLUMN IF NOT EXISTS credential_id UUID;

CREATE INDEX IF NOT EXISTS idx_outbox_credential
ON outbox_events(credential_id)
WHERE credential_id IS NOT NULL;

-- Comentários
COMMENT ON COLUMN outbox_events.event_category IS
'Categoria do evento: domain_event, message_delivery, webhook, meta_conversion, google_ads_conversion, custom';

COMMENT ON COLUMN outbox_events.credential_id IS
'ID da credencial usada para notifiers externos (Meta, Google Ads, etc)';
```

---

## 🚀 8. Exemplo de Uso Completo

### 📝 **Cenário: Conversão de Lead vindo do Meta Ads**

```go
// internal/application/contact/handle_lead_conversion.go
package contact

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cawolfkreo/ventros-crm/internal/domain/contact"
	"github.com/cawolfkreo/ventros-crm/internal/domain/outbox"
)

type HandleLeadConversionUseCase struct {
	contactRepo contact.Repository
	outboxRepo  outbox.Repository
}

func (uc *HandleLeadConversionUseCase) Execute(
	ctx context.Context,
	input LeadConversionInput,
) error {
	// 1. Cria/atualiza contato
	contact, err := uc.contactRepo.FindByPhone(input.Phone)
	if err != nil {
		// Cria novo contato
		contact, err = contact.NewContact(...)
		if err != nil {
			return err
		}
	}

	// 2. Adiciona conversão de ad
	contact.AddAdConversion(
		input.AdID,
		input.CampaignID,
		input.AdsetID,
		"meta",
		input.ConversionValue,
	)

	// 3. Salva contato
	if err := uc.contactRepo.Save(contact); err != nil {
		return err
	}

	// 4. Cria evento de conversão para Meta
	conversionData := map[string]interface{}{
		"event_name": "Lead",
		"event_time": time.Now().Unix(),
		"event_source_url": input.SourceURL,
		"user_data": map[string]interface{}{
			"ph": hashPhone(input.Phone), // SHA256
			"em": hashEmail(input.Email),  // SHA256
			"fn": hashName(input.FirstName),
			"ln": hashName(input.LastName),
			"country": input.Country,
		},
		"custom_data": map[string]interface{}{
			"value":    input.ConversionValue,
			"currency": "BRL",
			"content_name": "Lead Qualification",
		},
	}

	eventData, _ := json.Marshal(conversionData)

	// 5. Cria OutboxEvent para enviar conversão
	outboxEvent := &outbox.OutboxEvent{
		EventID:       uuid.New(),
		AggregateID:   contact.ID(),
		AggregateType: "contact",
		EventType:     string(outbox.EventTypeMetaConversion),
		EventData:     eventData,
		TenantID:      &input.TenantID,
		CredentialID:  &input.CredentialID, // credencial Meta
		Status:        outbox.StatusPending,
		CreatedAt:     time.Now(),
	}

	// 6. Salva no outbox (MESMA TRANSAÇÃO do contact)
	if err := uc.outboxRepo.Save(ctx, outboxEvent); err != nil {
		return err
	}

	// 7. Worker vai processar e enviar para Meta

	return nil
}
```

---

## ✅ 9. Checklist de Implementação

### 🔐 **Credentials & Security**

- [ ] Criar domain model `Credential`
- [ ] Implementar `AESEncryptor` (AES-256-GCM)
- [ ] Configurar variável de ambiente `ENCRYPTION_KEY`
- [ ] Implementar integração com AWS Secrets Manager (produção)
- [ ] Criar migrations para tabela `credentials`
- [ ] Implementar `CredentialRepository`
- [ ] Criar testes unitários para `Credential` aggregate
- [ ] Implementar rotação automática de encryption key

### 🔑 **Meta OAuth Flow**

- [ ] Criar `MetaOAuthService`
- [ ] Implementar PKCE (code_verifier, code_challenge)
- [ ] Implementar state validation (anti-CSRF)
- [ ] Criar endpoints `/api/oauth/meta/authorize` e `/callback`
- [ ] Implementar refresh token automático
- [ ] Adicionar worker para renovar tokens expirados
- [ ] Testar fluxo completo end-to-end
- [ ] Documentar permissões necessárias

### 🔔 **Notifier System**

- [ ] Criar interface `Notifier`
- [ ] Implementar `MetaConversionsNotifier`
- [ ] Implementar `GoogleAdsNotifier`
- [ ] Implementar `WebhookNotifier`
- [ ] Criar `NotifierRegistry` (strategy pattern)
- [ ] Adicionar tipos de evento ao `OutboxEvent`
- [ ] Criar worker para processar notifiers
- [ ] Implementar retry logic com exponential backoff

### 📊 **OutboxEvent Types**

- [ ] Adicionar campo `event_category` ao outbox
- [ ] Adicionar campo `credential_id` ao outbox
- [ ] Criar migration 000022
- [ ] Atualizar worker para processar novos tipos
- [ ] Implementar roteamento por tipo (domain event vs notifier)

### 🧪 **Testes**

- [ ] Testes unitários: `Credential` aggregate
- [ ] Testes unitários: `AESEncryptor`
- [ ] Testes unitários: `MetaOAuthService`
- [ ] Testes unitários: `MetaConversionsNotifier`
- [ ] Testes de integração: OAuth flow completo
- [ ] Testes de integração: Envio de conversão
- [ ] Testes E2E: Fluxo completo de lead → conversão

---

## 🎯 10. Próximos Passos Recomendados

### 🔴 **Prioridade ALTA (Fazer AGORA)**

1. **Remover `OutboundMessage` como aggregate**
   - Mensagens já existem em `messages` table
   - Use `OutboxEvent` com tipo `MESSAGE_DELIVERY`

2. **Implementar `Credential` domain model**
   - Seguir estrutura proposta acima
   - Usar AES-256-GCM

3. **Implementar Meta OAuth flow**
   - Com PKCE
   - Com refresh token automático

### 🟡 **Prioridade MÉDIA (Próximas 2 semanas)**

4. **Implementar Meta Conversions Notifier**
5. **Adicionar tipos de evento ao Outbox**
6. **Criar worker para notifiers**
7. **Adicionar testes unitários (60% coverage)**

### 🟢 **Prioridade BAIXA (Próximo mês)**

8. **Implementar Google Ads Notifier**
9. **Adicionar suporte a múltiplas credenciais por tenant**
10. **Implementar auditoria de uso de credentials**
11. **Dashboard de integrações ativas**

---

## 📚 Referências

- [Meta WhatsApp Cloud API](https://developers.facebook.com/docs/whatsapp/cloud-api/)
- [Meta Conversions API](https://developers.facebook.com/docs/marketing-api/conversions-api/)
- [OAuth 2.0 PKCE](https://oauth.net/2/pkce/)
- [AES-GCM Encryption](https://en.wikipedia.org/wiki/Galois/Counter_Mode)
- [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/)

---

**Fim do Guia** 🎉

Quer que eu implemente alguma parte específica agora?
