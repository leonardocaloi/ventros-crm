package campaign

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/automation/campaign"
)

// ActivateCampaignHandler handler para o comando ActivateCampaign
type ActivateCampaignHandler struct {
	repository campaign.Repository
	logger     *logrus.Logger
}

// NewActivateCampaignHandler cria uma nova instância do handler
func NewActivateCampaignHandler(repository campaign.Repository, logger *logrus.Logger) *ActivateCampaignHandler {
	return &ActivateCampaignHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de ativação de campanha
func (h *ActivateCampaignHandler) Handle(ctx context.Context, cmd ActivateCampaignCommand) (*campaign.Campaign, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid ActivateCampaign command")
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

	// Activate campaign
	if err := camp.Activate(); err != nil {
		h.logger.WithError(err).Error("Failed to activate campaign")
		return nil, fmt.Errorf("%w: %v", ErrCampaignActivateFailed, err)
	}

	// Save to repository
	if err := h.repository.Save(camp); err != nil {
		h.logger.WithError(err).Error("Failed to save activated campaign")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithField("campaign_id", camp.ID()).Info("Campaign activated successfully")

	return camp, nil
}

// PauseCampaignHandler handler para o comando PauseCampaign
type PauseCampaignHandler struct {
	repository campaign.Repository
	logger     *logrus.Logger
}

// NewPauseCampaignHandler cria uma nova instância do handler
func NewPauseCampaignHandler(repository campaign.Repository, logger *logrus.Logger) *PauseCampaignHandler {
	return &PauseCampaignHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de pause de campanha
func (h *PauseCampaignHandler) Handle(ctx context.Context, cmd PauseCampaignCommand) (*campaign.Campaign, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid PauseCampaign command")
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

	// Pause campaign
	if err := camp.Pause(); err != nil {
		h.logger.WithError(err).Error("Failed to pause campaign")
		return nil, fmt.Errorf("%w: %v", ErrCampaignPauseFailed, err)
	}

	// Save to repository
	if err := h.repository.Save(camp); err != nil {
		h.logger.WithError(err).Error("Failed to save paused campaign")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithField("campaign_id", camp.ID()).Info("Campaign paused successfully")

	return camp, nil
}

// CompleteCampaignHandler handler para o comando CompleteCampaign
type CompleteCampaignHandler struct {
	repository campaign.Repository
	logger     *logrus.Logger
}

// NewCompleteCampaignHandler cria uma nova instância do handler
func NewCompleteCampaignHandler(repository campaign.Repository, logger *logrus.Logger) *CompleteCampaignHandler {
	return &CompleteCampaignHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de conclusão de campanha
func (h *CompleteCampaignHandler) Handle(ctx context.Context, cmd CompleteCampaignCommand) (*campaign.Campaign, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid CompleteCampaign command")
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

	// Complete campaign
	if err := camp.Complete(); err != nil {
		h.logger.WithError(err).Error("Failed to complete campaign")
		return nil, fmt.Errorf("%w: %v", ErrCampaignCompleteFailed, err)
	}

	// Save to repository
	if err := h.repository.Save(camp); err != nil {
		h.logger.WithError(err).Error("Failed to save completed campaign")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithField("campaign_id", camp.ID()).Info("Campaign completed successfully")

	return camp, nil
}
