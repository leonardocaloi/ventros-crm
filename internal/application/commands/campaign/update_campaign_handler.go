package campaign

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/automation/campaign"
	"github.com/sirupsen/logrus"
)

// UpdateCampaignHandler handler para o comando UpdateCampaign
type UpdateCampaignHandler struct {
	repository campaign.Repository
	logger     *logrus.Logger
}

// NewUpdateCampaignHandler cria uma nova instância do handler
func NewUpdateCampaignHandler(repository campaign.Repository, logger *logrus.Logger) *UpdateCampaignHandler {
	return &UpdateCampaignHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de atualização de campanha
func (h *UpdateCampaignHandler) Handle(ctx context.Context, cmd UpdateCampaignCommand) (*campaign.Campaign, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid UpdateCampaign command")
		return nil, err
	}

	// Find campaign
	camp, err := h.repository.FindByID(cmd.CampaignID)
	if err != nil {
		h.logger.WithError(err).WithField("campaign_id", cmd.CampaignID).Error("Campaign not found")
		return nil, fmt.Errorf("%w: %v", ErrCampaignNotFound, err)
	}

	// Check tenant ownership
	if camp.TenantID() != cmd.TenantID {
		h.logger.WithFields(logrus.Fields{
			"campaign_id": cmd.CampaignID,
			"tenant_id":   cmd.TenantID,
		}).Warn("Access denied: tenant mismatch")
		return nil, ErrAccessDenied
	}

	// Update name if provided
	if cmd.Name != nil {
		if err := camp.UpdateName(*cmd.Name); err != nil {
			h.logger.WithError(err).Error("Failed to update campaign name")
			return nil, fmt.Errorf("%w: %v", ErrCampaignUpdateFailed, err)
		}
	}

	// Update description if provided
	if cmd.Description != nil {
		camp.UpdateDescription(*cmd.Description)
	}

	// Update goal if both provided
	if cmd.GoalType != nil && cmd.GoalValue != nil {
		if err := camp.UpdateGoal(campaign.GoalType(*cmd.GoalType), *cmd.GoalValue); err != nil {
			h.logger.WithError(err).Error("Failed to update campaign goal")
			return nil, fmt.Errorf("%w: %v", ErrCampaignUpdateFailed, err)
		}
	}

	// Save to repository
	if err := h.repository.Save(camp); err != nil {
		h.logger.WithError(err).Error("Failed to save updated campaign")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"campaign_id": camp.ID(),
		"tenant_id":   camp.TenantID(),
		"name":        camp.Name(),
	}).Info("Campaign updated successfully")

	return camp, nil
}
