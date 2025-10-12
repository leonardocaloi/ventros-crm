package entities

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CredentialEntity representa uma credencial no banco de dados
type CredentialEntity struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key"`
	Version   int        `gorm:"default:1;not null"` // Optimistic locking
	TenantID  string     `gorm:"type:varchar(255);not null;index:idx_credentials_tenant"`
	ProjectID *uuid.UUID `gorm:"type:uuid;index:idx_credentials_project"`

	// Tipo e identificação
	CredentialType string `gorm:"type:varchar(50);not null;index:idx_credentials_type"`
	Name           string `gorm:"type:varchar(255);not null"`
	Description    string `gorm:"type:text"`

	// Valor criptografado (AES-256-GCM)
	EncryptedValueCiphertext string `gorm:"type:text;not null;column:encrypted_value_ciphertext"`
	EncryptedValueNonce      string `gorm:"type:text;not null;column:encrypted_value_nonce"`

	// OAuth tokens (quando aplicável)
	OAuthAccessTokenCiphertext  *string    `gorm:"type:text;column:oauth_access_token_ciphertext"`
	OAuthAccessTokenNonce       *string    `gorm:"type:text;column:oauth_access_token_nonce"`
	OAuthRefreshTokenCiphertext *string    `gorm:"type:text;column:oauth_refresh_token_ciphertext"`
	OAuthRefreshTokenNonce      *string    `gorm:"type:text;column:oauth_refresh_token_nonce"`
	OAuthTokenType              *string    `gorm:"type:varchar(20);column:oauth_token_type"`
	OAuthExpiresAt              *time.Time `gorm:"column:oauth_expires_at"`

	// Metadata adicional
	Metadata MetadataJSON `gorm:"type:jsonb;default:'{}'"`

	// Status e lifecycle
	IsActive   bool       `gorm:"not null;default:true;index:idx_credentials_active"`
	ExpiresAt  *time.Time `gorm:"column:expires_at;index:idx_credentials_expires_at"`
	LastUsedAt *time.Time `gorm:"column:last_used_at"`

	// Auditoria
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName especifica o nome da tabela
func (CredentialEntity) TableName() string {
	return "credentials"
}

// MetadataJSON é um tipo customizado para metadata JSONB
type MetadataJSON map[string]interface{}

// Scan implementa sql.Scanner
func (m *MetadataJSON) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		*m = make(map[string]interface{})
		return nil
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*m = result
	return nil
}

// Value implementa driver.Valuer
func (m MetadataJSON) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(make(map[string]interface{}))
	}
	return json.Marshal(m)
}
