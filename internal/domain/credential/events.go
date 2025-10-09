package credential

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent interface base para eventos de domínio
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// CredentialCreatedEvent - Credencial criada
type CredentialCreatedEvent struct {
	CredentialID   uuid.UUID
	TenantID       string
	CredentialType CredentialType
	Name           string
	CreatedAt      time.Time
}

func (e CredentialCreatedEvent) EventName() string     { return "credential.created" }
func (e CredentialCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

// CredentialUpdatedEvent - Credencial atualizada
type CredentialUpdatedEvent struct {
	CredentialID uuid.UUID
	UpdatedAt    time.Time
}

func (e CredentialUpdatedEvent) EventName() string     { return "credential.updated" }
func (e CredentialUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

// OAuthTokenRefreshedEvent - Token OAuth renovado
type OAuthTokenRefreshedEvent struct {
	CredentialID uuid.UUID
	ExpiresAt    time.Time
	RefreshedAt  time.Time
}

func (e OAuthTokenRefreshedEvent) EventName() string     { return "credential.oauth_refreshed" }
func (e OAuthTokenRefreshedEvent) OccurredAt() time.Time { return e.RefreshedAt }

// CredentialActivatedEvent - Credencial ativada
type CredentialActivatedEvent struct {
	CredentialID uuid.UUID
	ActivatedAt  time.Time
}

func (e CredentialActivatedEvent) EventName() string     { return "credential.activated" }
func (e CredentialActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

// CredentialDeactivatedEvent - Credencial desativada
type CredentialDeactivatedEvent struct {
	CredentialID  uuid.UUID
	DeactivatedAt time.Time
}

func (e CredentialDeactivatedEvent) EventName() string     { return "credential.deactivated" }
func (e CredentialDeactivatedEvent) OccurredAt() time.Time { return e.DeactivatedAt }

// CredentialUsedEvent - Credencial foi usada
type CredentialUsedEvent struct {
	CredentialID uuid.UUID
	UsedAt       time.Time
}

func (e CredentialUsedEvent) EventName() string     { return "credential.used" }
func (e CredentialUsedEvent) OccurredAt() time.Time { return e.UsedAt }

// CredentialExpiredEvent - Credencial expirou
type CredentialExpiredEvent struct {
	CredentialID uuid.UUID
	ExpiredAt    time.Time
}

func (e CredentialExpiredEvent) EventName() string     { return "credential.expired" }
func (e CredentialExpiredEvent) OccurredAt() time.Time { return e.ExpiredAt }
