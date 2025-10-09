package credential

import (
	"errors"
	"time"
)

// OAuthToken representa tokens OAuth criptografados
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

	// Criptografa access token
	encryptedAccess, err := encryptor.Encrypt(accessToken)
	if err != nil {
		return nil, err
	}

	// Criptografa refresh token (se fornecido)
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
	if t.encryptedRefreshToken.IsEmpty() {
		return "", errors.New("no refresh token available")
	}
	return encryptor.Decrypt(t.encryptedRefreshToken)
}

// IsExpired verifica se o token expirou
func (t *OAuthToken) IsExpired() bool {
	return time.Now().After(t.expiresAt)
}

// NeedsRefresh verifica se precisa renovar (30 min antes de expirar)
func (t *OAuthToken) NeedsRefresh() bool {
	return time.Now().Add(30 * time.Minute).After(t.expiresAt)
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

// TokenType retorna o tipo do token (geralmente "Bearer")
func (t *OAuthToken) TokenType() string {
	return t.tokenType
}

// HasRefreshToken verifica se possui refresh token
func (t *OAuthToken) HasRefreshToken() bool {
	return !t.encryptedRefreshToken.IsEmpty()
}

// ReconstructOAuthToken reconstrói um OAuthToken a partir de dados persistidos
func ReconstructOAuthToken(
	encryptedAccessToken EncryptedValue,
	encryptedRefreshToken EncryptedValue,
	expiresAt time.Time,
	tokenType *string,
) *OAuthToken {
	tt := "Bearer"
	if tokenType != nil {
		tt = *tokenType
	}

	return &OAuthToken{
		encryptedAccessToken:  encryptedAccessToken,
		encryptedRefreshToken: encryptedRefreshToken,
		expiresAt:             expiresAt,
		tokenType:             tt,
	}
}

// EncryptedAccessToken retorna o access token criptografado (para persistência)
func (t *OAuthToken) EncryptedAccessToken() EncryptedValue {
	return t.encryptedAccessToken
}

// EncryptedRefreshToken retorna o refresh token criptografado (para persistência)
func (t *OAuthToken) EncryptedRefreshToken() EncryptedValue {
	return t.encryptedRefreshToken
}
