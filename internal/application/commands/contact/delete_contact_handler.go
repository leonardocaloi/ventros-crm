package contact

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/crm/contact"
)

// DeleteContactHandler handler para o comando DeleteContact
type DeleteContactHandler struct {
	repository contact.Repository
	logger     *logrus.Logger
}

// NewDeleteContactHandler cria uma nova instância do handler
func NewDeleteContactHandler(repository contact.Repository, logger *logrus.Logger) *DeleteContactHandler {
	return &DeleteContactHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de deleção de contato (soft delete)
func (h *DeleteContactHandler) Handle(ctx context.Context, cmd DeleteContactCommand) error {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid DeleteContact command")
		return err
	}

	// Find contact
	domainContact, err := h.repository.FindByID(ctx, cmd.ContactID)
	if err != nil {
		h.logger.WithError(err).WithField("contact_id", cmd.ContactID).Error("Contact not found")
		return fmt.Errorf("%w: %v", ErrContactNotFound, err)
	}

	if domainContact == nil {
		h.logger.WithField("contact_id", cmd.ContactID).Warn("Contact not found (nil returned)")
		return ErrContactNotFound
	}

	// Check tenant ownership
	if domainContact.TenantID() != cmd.TenantID {
		h.logger.WithFields(logrus.Fields{
			"contact_id": cmd.ContactID,
			"tenant_id":  cmd.TenantID,
		}).Warn("Access denied: tenant mismatch")
		return ErrAccessDenied
	}

	// Soft delete
	domainContact.Delete()

	// Save to repository
	if err := h.repository.Save(ctx, domainContact); err != nil {
		h.logger.WithError(err).Error("Failed to save deleted contact")
		return fmt.Errorf("%w: %v", ErrContactDeleteFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"contact_id": domainContact.ID(),
		"tenant_id":  domainContact.TenantID(),
	}).Info("Contact deleted successfully")

	return nil
}
