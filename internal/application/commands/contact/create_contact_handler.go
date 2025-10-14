package contact

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/crm/contact"
)

// CreateContactHandler handler para o comando CreateContact
type CreateContactHandler struct {
	repository contact.Repository
	logger     *logrus.Logger
}

// NewCreateContactHandler cria uma nova instância do handler
func NewCreateContactHandler(repository contact.Repository, logger *logrus.Logger) *CreateContactHandler {
	return &CreateContactHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de criação de contato
func (h *CreateContactHandler) Handle(ctx context.Context, cmd CreateContactCommand) (*contact.Contact, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid CreateContact command")
		return nil, err
	}

	// Create domain contact
	domainContact, err := contact.NewContact(
		cmd.ProjectID,
		cmd.TenantID,
		cmd.Name,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create contact domain object")
		return nil, fmt.Errorf("%w: %v", ErrContactCreationFailed, err)
	}

	// Set optional fields
	if cmd.Email != "" {
		if err := domainContact.SetEmail(cmd.Email); err != nil {
			h.logger.WithError(err).WithField("email", cmd.Email).Error("Failed to set email")
			return nil, fmt.Errorf("%w: %v", ErrInvalidEmail, err)
		}
	}

	if cmd.Phone != "" {
		if err := domainContact.SetPhone(cmd.Phone); err != nil {
			h.logger.WithError(err).WithField("phone", cmd.Phone).Error("Failed to set phone")
			return nil, fmt.Errorf("%w: %v", ErrInvalidPhone, err)
		}
	}

	if cmd.ExternalID != "" {
		domainContact.SetExternalID(cmd.ExternalID)
	}

	if cmd.SourceChannel != "" {
		domainContact.SetSourceChannel(cmd.SourceChannel)
	}

	if cmd.Language != "" {
		domainContact.SetLanguage(cmd.Language)
	}

	if cmd.Timezone != "" {
		domainContact.SetTimezone(cmd.Timezone)
	}

	// Add tags
	for _, tag := range cmd.Tags {
		domainContact.AddTag(tag)
	}

	// TODO: Add custom fields support
	// Custom fields are separate entities (ContactCustomField) and require their own repository
	// For now, we skip this until the infrastructure is in place
	_ = cmd.CustomFields

	// Save to repository
	if err := h.repository.Save(ctx, domainContact); err != nil {
		h.logger.WithError(err).Error("Failed to save contact to repository")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"contact_id": domainContact.ID(),
		"tenant_id":  domainContact.TenantID(),
		"name":       domainContact.Name(),
	}).Info("Contact created successfully")

	return domainContact, nil
}
