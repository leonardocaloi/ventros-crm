package credential

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Credential struct {
	id             uuid.UUID
	tenantID       string
	projectID      *uuid.UUID
	credentialType CredentialType
	name           string
	description    string

	encryptedValue EncryptedValue

	oauthToken *OAuthToken

	metadata   map[string]interface{}
	isActive   bool
	expiresAt  *time.Time
	lastUsedAt *time.Time

	createdAt time.Time
	updatedAt time.Time

	events []DomainEvent
}

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
	if !credentialType.IsValid() {
		return nil, errors.New("invalid credential type")
	}

	encryptedValue, err := encryptor.Encrypt(plainValue)
	if err != nil {
		return nil, err
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

func ReconstructCredential(
	id uuid.UUID,
	tenantID string,
	projectID *uuid.UUID,
	credentialType CredentialType,
	name string,
	description string,
	encryptedValue EncryptedValue,
	oauthToken *OAuthToken,
	metadata map[string]interface{},
	isActive bool,
	expiresAt *time.Time,
	lastUsedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Credential {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &Credential{
		id:             id,
		tenantID:       tenantID,
		projectID:      projectID,
		credentialType: credentialType,
		name:           name,
		description:    description,
		encryptedValue: encryptedValue,
		oauthToken:     oauthToken,
		metadata:       metadata,
		isActive:       isActive,
		expiresAt:      expiresAt,
		lastUsedAt:     lastUsedAt,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		events:         []DomainEvent{},
	}
}

func (c *Credential) SetOAuthToken(
	accessToken string,
	refreshToken string,
	expiresIn int,
	encryptor Encryptor,
) error {
	if !c.credentialType.RequiresOAuth() {
		return errors.New("OAuth tokens only valid for OAuth credential types")
	}

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	token, err := NewOAuthToken(accessToken, refreshToken, expiresAt, encryptor)
	if err != nil {
		return err
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

func (c *Credential) UpdateValue(plainValue string, encryptor Encryptor) error {
	if plainValue == "" {
		return errors.New("value cannot be empty")
	}

	encryptedValue, err := encryptor.Encrypt(plainValue)
	if err != nil {
		return err
	}

	c.encryptedValue = encryptedValue
	c.updatedAt = time.Now()

	c.addEvent(CredentialUpdatedEvent{
		CredentialID: c.id,
		UpdatedAt:    c.updatedAt,
	})

	return nil
}

func (c *Credential) UpdateDescription(description string) {
	c.description = description
	c.updatedAt = time.Now()
}

func (c *Credential) SetProjectID(projectID uuid.UUID) {
	c.projectID = &projectID
	c.updatedAt = time.Now()
}

func (c *Credential) IsExpired() bool {
	if c.expiresAt == nil {
		return false
	}
	return time.Now().After(*c.expiresAt)
}

func (c *Credential) NeedsRefresh() bool {
	if c.expiresAt == nil || c.oauthToken == nil {
		return false
	}
	return c.oauthToken.NeedsRefresh()
}

func (c *Credential) Decrypt(encryptor Encryptor) (string, error) {
	return encryptor.Decrypt(c.encryptedValue)
}

func (c *Credential) GetAccessToken(encryptor Encryptor) (string, error) {
	if c.oauthToken == nil {
		return "", errors.New("no OAuth token available")
	}
	return c.oauthToken.GetAccessToken(encryptor)
}

func (c *Credential) GetRefreshToken(encryptor Encryptor) (string, error) {
	if c.oauthToken == nil {
		return "", errors.New("no OAuth token available")
	}
	return c.oauthToken.GetRefreshToken(encryptor)
}

func (c *Credential) MarkAsUsed() {
	now := time.Now()
	c.lastUsedAt = &now
	c.updatedAt = now

	c.addEvent(CredentialUsedEvent{
		CredentialID: c.id,
		UsedAt:       now,
	})
}

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

func (c *Credential) SetMetadata(key string, value interface{}) {
	c.metadata[key] = value
	c.updatedAt = time.Now()
}

func (c *Credential) GetMetadata(key string) (interface{}, bool) {
	val, exists := c.metadata[key]
	return val, exists
}

func (c *Credential) ID() uuid.UUID                  { return c.id }
func (c *Credential) TenantID() string               { return c.tenantID }
func (c *Credential) ProjectID() *uuid.UUID          { return c.projectID }
func (c *Credential) Type() CredentialType           { return c.credentialType }
func (c *Credential) Name() string                   { return c.name }
func (c *Credential) Description() string            { return c.description }
func (c *Credential) IsActive() bool                 { return c.isActive }
func (c *Credential) ExpiresAt() *time.Time          { return c.expiresAt }
func (c *Credential) LastUsedAt() *time.Time         { return c.lastUsedAt }
func (c *Credential) CreatedAt() time.Time           { return c.createdAt }
func (c *Credential) UpdatedAt() time.Time           { return c.updatedAt }
func (c *Credential) EncryptedValue() EncryptedValue { return c.encryptedValue }
func (c *Credential) OAuthToken() *OAuthToken        { return c.oauthToken }
func (c *Credential) Metadata() map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range c.metadata {
		copy[k] = v
	}
	return copy
}

func (c *Credential) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, c.events...)
}

func (c *Credential) ClearEvents() {
	c.events = []DomainEvent{}
}

func (c *Credential) addEvent(event DomainEvent) {
	c.events = append(c.events, event)
}
