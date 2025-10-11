package credential

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type CredentialCreatedEvent struct {
	CredentialID   uuid.UUID
	TenantID       string
	CredentialType CredentialType
	Name           string
	CreatedAt      time.Time
}

func (e CredentialCreatedEvent) EventName() string     { return "credential.created" }
func (e CredentialCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type CredentialUpdatedEvent struct {
	CredentialID uuid.UUID
	UpdatedAt    time.Time
}

func (e CredentialUpdatedEvent) EventName() string     { return "credential.updated" }
func (e CredentialUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

type OAuthTokenRefreshedEvent struct {
	CredentialID uuid.UUID
	ExpiresAt    time.Time
	RefreshedAt  time.Time
}

func (e OAuthTokenRefreshedEvent) EventName() string     { return "credential.oauth_refreshed" }
func (e OAuthTokenRefreshedEvent) OccurredAt() time.Time { return e.RefreshedAt }

type CredentialActivatedEvent struct {
	CredentialID uuid.UUID
	ActivatedAt  time.Time
}

func (e CredentialActivatedEvent) EventName() string     { return "credential.activated" }
func (e CredentialActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

type CredentialDeactivatedEvent struct {
	CredentialID  uuid.UUID
	DeactivatedAt time.Time
}

func (e CredentialDeactivatedEvent) EventName() string     { return "credential.deactivated" }
func (e CredentialDeactivatedEvent) OccurredAt() time.Time { return e.DeactivatedAt }

type CredentialUsedEvent struct {
	CredentialID uuid.UUID
	UsedAt       time.Time
}

func (e CredentialUsedEvent) EventName() string     { return "credential.used" }
func (e CredentialUsedEvent) OccurredAt() time.Time { return e.UsedAt }

type CredentialExpiredEvent struct {
	CredentialID uuid.UUID
	ExpiredAt    time.Time
}

func (e CredentialExpiredEvent) EventName() string     { return "credential.expired" }
func (e CredentialExpiredEvent) OccurredAt() time.Time { return e.ExpiredAt }
