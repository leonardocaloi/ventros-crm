package credential

import (
	"errors"
	"time"
)

type OAuthToken struct {
	encryptedAccessToken  EncryptedValue
	encryptedRefreshToken EncryptedValue
	expiresAt             time.Time
	tokenType             string
}

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

func (t *OAuthToken) GetAccessToken(encryptor Encryptor) (string, error) {
	return encryptor.Decrypt(t.encryptedAccessToken)
}

func (t *OAuthToken) GetRefreshToken(encryptor Encryptor) (string, error) {
	if t.encryptedRefreshToken.IsEmpty() {
		return "", errors.New("no refresh token available")
	}
	return encryptor.Decrypt(t.encryptedRefreshToken)
}

func (t *OAuthToken) IsExpired() bool {
	return time.Now().After(t.expiresAt)
}

func (t *OAuthToken) NeedsRefresh() bool {
	return time.Now().Add(30 * time.Minute).After(t.expiresAt)
}

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

func (t *OAuthToken) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t *OAuthToken) TokenType() string {
	return t.tokenType
}

func (t *OAuthToken) HasRefreshToken() bool {
	return !t.encryptedRefreshToken.IsEmpty()
}

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

func (t *OAuthToken) EncryptedAccessToken() EncryptedValue {
	return t.encryptedAccessToken
}

func (t *OAuthToken) EncryptedRefreshToken() EncryptedValue {
	return t.encryptedRefreshToken
}
