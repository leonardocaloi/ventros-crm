package contact

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/crm/contact"
)

// UpdateContactHandler handler para o comando UpdateContact
type UpdateContactHandler struct {
	repository contact.Repository
	logger     *logrus.Logger
}

// NewUpdateContactHandler cria uma nova instância do handler
func NewUpdateContactHandler(repository contact.Repository, logger *logrus.Logger) *UpdateContactHandler {
	return &UpdateContactHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de atualização de contato
func (h *UpdateContactHandler) Handle(ctx context.Context, cmd UpdateContactCommand) (*contact.Contact, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid UpdateContact command")
		return nil, err
	}

	// Find contact
	domainContact, err := h.repository.FindByID(ctx, cmd.ContactID)
	if err != nil {
		h.logger.WithError(err).WithField("contact_id", cmd.ContactID).Error("Contact not found")
		return nil, fmt.Errorf("%w: %v", ErrContactNotFound, err)
	}

	if domainContact == nil {
		h.logger.WithField("contact_id", cmd.ContactID).Warn("Contact not found (nil returned)")
		return nil, ErrContactNotFound
	}

	// Check tenant ownership
	if domainContact.TenantID() != cmd.TenantID {
		h.logger.WithFields(logrus.Fields{
			"contact_id": cmd.ContactID,
			"tenant_id":  cmd.TenantID,
		}).Warn("Access denied: tenant mismatch")
		return nil, ErrAccessDenied
	}

	// Update name if provided
	if cmd.Name != nil {
		domainContact.UpdateName(*cmd.Name)
	}

	// Update email if provided
	if cmd.Email != nil {
		if err := domainContact.SetEmail(*cmd.Email); err != nil {
			h.logger.WithError(err).WithField("email", *cmd.Email).Error("Failed to set email")
			return nil, fmt.Errorf("%w: %v", ErrInvalidEmail, err)
		}
	}

	// Update phone if provided
	if cmd.Phone != nil {
		if err := domainContact.SetPhone(*cmd.Phone); err != nil {
			h.logger.WithError(err).WithField("phone", *cmd.Phone).Error("Failed to set phone")
			return nil, fmt.Errorf("%w: %v", ErrInvalidPhone, err)
		}
	}

	// Update external ID if provided
	if cmd.ExternalID != nil {
		domainContact.SetExternalID(*cmd.ExternalID)
	}

	// Update source channel if provided
	if cmd.SourceChannel != nil {
		domainContact.SetSourceChannel(*cmd.SourceChannel)
	}

	// Update language if provided
	if cmd.Language != nil {
		domainContact.SetLanguage(*cmd.Language)
	}

	// Update timezone if provided
	if cmd.Timezone != nil {
		domainContact.SetTimezone(*cmd.Timezone)
	}

	// Update tags if provided (replace all tags)
	if len(cmd.Tags) > 0 {
		domainContact.ClearTags()
		for _, tag := range cmd.Tags {
			domainContact.AddTag(tag)
		}
	}

	// TODO: Update custom fields support
	// Custom fields are separate entities (ContactCustomField) and require their own repository
	// For now, we skip this until the infrastructure is in place
	_ = cmd.CustomFields

	// Save to repository
	if err := h.repository.Save(ctx, domainContact); err != nil {
		h.logger.WithError(err).Error("Failed to save updated contact")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"contact_id": domainContact.ID(),
		"tenant_id":  domainContact.TenantID(),
		"name":       domainContact.Name(),
	}).Info("Contact updated successfully")

	return domainContact, nil
}
