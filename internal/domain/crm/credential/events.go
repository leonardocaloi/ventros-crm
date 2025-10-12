package credential

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

type CredentialCreatedEvent struct {
	shared.BaseEvent
	CredentialID   uuid.UUID
	TenantID       string
	CredentialType CredentialType
	Name           string
}

func NewCredentialCreatedEvent(credentialID uuid.UUID, tenantID string, credentialType CredentialType, name string) CredentialCreatedEvent {
	return CredentialCreatedEvent{
		BaseEvent:      shared.NewBaseEvent("credential.created", time.Now()),
		CredentialID:   credentialID,
		TenantID:       tenantID,
		CredentialType: credentialType,
		Name:           name,
	}
}

type CredentialUpdatedEvent struct {
	shared.BaseEvent
	CredentialID uuid.UUID
}

func NewCredentialUpdatedEvent(credentialID uuid.UUID) CredentialUpdatedEvent {
	return CredentialUpdatedEvent{
		BaseEvent:    shared.NewBaseEvent("credential.updated", time.Now()),
		CredentialID: credentialID,
	}
}

type OAuthTokenRefreshedEvent struct {
	shared.BaseEvent
	CredentialID uuid.UUID
	ExpiresAt    time.Time
}

func NewOAuthTokenRefreshedEvent(credentialID uuid.UUID, expiresAt time.Time) OAuthTokenRefreshedEvent {
	return OAuthTokenRefreshedEvent{
		BaseEvent:    shared.NewBaseEvent("credential.oauth_refreshed", time.Now()),
		CredentialID: credentialID,
		ExpiresAt:    expiresAt,
	}
}

type CredentialActivatedEvent struct {
	shared.BaseEvent
	CredentialID uuid.UUID
}

func NewCredentialActivatedEvent(credentialID uuid.UUID) CredentialActivatedEvent {
	return CredentialActivatedEvent{
		BaseEvent:    shared.NewBaseEvent("credential.activated", time.Now()),
		CredentialID: credentialID,
	}
}

type CredentialDeactivatedEvent struct {
	shared.BaseEvent
	CredentialID uuid.UUID
}

func NewCredentialDeactivatedEvent(credentialID uuid.UUID) CredentialDeactivatedEvent {
	return CredentialDeactivatedEvent{
		BaseEvent:    shared.NewBaseEvent("credential.deactivated", time.Now()),
		CredentialID: credentialID,
	}
}

type CredentialUsedEvent struct {
	shared.BaseEvent
	CredentialID uuid.UUID
}

func NewCredentialUsedEvent(credentialID uuid.UUID) CredentialUsedEvent {
	return CredentialUsedEvent{
		BaseEvent:    shared.NewBaseEvent("credential.used", time.Now()),
		CredentialID: credentialID,
	}
}

type CredentialExpiredEvent struct {
	shared.BaseEvent
	CredentialID uuid.UUID
}

func NewCredentialExpiredEvent(credentialID uuid.UUID) CredentialExpiredEvent {
	return CredentialExpiredEvent{
		BaseEvent:    shared.NewBaseEvent("credential.expired", time.Now()),
		CredentialID: credentialID,
	}
}
