package chat

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Create saves a new chat
	Create(ctx context.Context, chat *Chat) error

	// FindByID finds a chat by ID
	FindByID(ctx context.Context, id uuid.UUID) (*Chat, error)

	// FindByExternalID finds a chat by external ID (WhatsApp group ID, etc)
	FindByExternalID(ctx context.Context, externalID string) (*Chat, error)

	// FindByProject finds all chats for a project
	FindByProject(ctx context.Context, projectID uuid.UUID) ([]*Chat, error)

	// FindByTenant finds all chats for a tenant
	FindByTenant(ctx context.Context, tenantID string) ([]*Chat, error)

	// FindByContact finds all chats where contact is a participant
	FindByContact(ctx context.Context, contactID uuid.UUID) ([]*Chat, error)

	// FindActiveByProject finds all active chats (not archived, not closed) for a project
	FindActiveByProject(ctx context.Context, projectID uuid.UUID) ([]*Chat, error)

	// FindIndividualByContact finds an individual chat for a contact in a project
	FindIndividualByContact(ctx context.Context, contactID uuid.UUID, projectID uuid.UUID) (*Chat, error)

	// Update updates an existing chat
	Update(ctx context.Context, chat *Chat) error

	// Delete soft deletes a chat
	Delete(ctx context.Context, id uuid.UUID) error

	// SearchBySubject searches chats by subject
	SearchBySubject(ctx context.Context, tenantID string, subject string) ([]*Chat, error)
}
