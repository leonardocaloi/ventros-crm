package campaign

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/automation/campaign"
)

// CreateCampaignHandler handler para o comando CreateCampaign
type CreateCampaignHandler struct {
	repository campaign.Repository
	logger     *logrus.Logger
}

// NewCreateCampaignHandler cria uma nova instância do handler
func NewCreateCampaignHandler(repository campaign.Repository, logger *logrus.Logger) *CreateCampaignHandler {
	return &CreateCampaignHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de criação de campanha
func (h *CreateCampaignHandler) Handle(ctx context.Context, cmd CreateCampaignCommand) (*campaign.Campaign, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid CreateCampaign command")
		return nil, err
	}

	// Create campaign domain object
	camp, err := campaign.NewCampaign(
		cmd.TenantID,
		cmd.Name,
		cmd.Description,
		campaign.GoalType(cmd.GoalType),
		cmd.GoalValue,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create campaign domain object")
		return nil, fmt.Errorf("%w: %v", ErrCampaignCreationFailed, err)
	}

	// Add steps if provided
	for _, stepCmd := range cmd.Steps {
		if err := h.addStepToCampaign(camp, stepCmd); err != nil {
			h.logger.WithError(err).Error("Failed to add step to campaign")
			return nil, err
		}
	}

	// Save to repository
	if err := h.repository.Save(camp); err != nil {
		h.logger.WithError(err).Error("Failed to save campaign to repository")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"campaign_id": camp.ID(),
		"tenant_id":   camp.TenantID(),
		"name":        camp.Name(),
	}).Info("Campaign created successfully")

	return camp, nil
}

// addStepToCampaign adiciona um step à campanha
func (h *CreateCampaignHandler) addStepToCampaign(camp *campaign.Campaign, stepCmd CreateCampaignStepCommand) error {
	// Build step config
	config := campaign.StepConfig{
		BroadcastID:   stepCmd.Config.BroadcastID,
		SequenceID:    stepCmd.Config.SequenceID,
		DelayAmount:   stepCmd.Config.DelayAmount,
		DelayUnit:     stepCmd.Config.DelayUnit,
		ConditionType: stepCmd.Config.ConditionType,
		ConditionData: stepCmd.Config.ConditionData,
		WaitFor:       stepCmd.Config.WaitFor,
		WaitTimeout:   stepCmd.Config.WaitTimeout,
		TimeoutStep:   stepCmd.Config.TimeoutStep,
	}

	// Create step
	step := campaign.NewCampaignStep(
		stepCmd.Order,
		stepCmd.Name,
		campaign.StepType(stepCmd.Type),
		config,
	)

	// Add conditions to step
	for _, condCmd := range stepCmd.Conditions {
		step.AddCondition(campaign.StepCondition{
			Type:     condCmd.Type,
			Field:    condCmd.Field,
			Operator: condCmd.Operator,
			Value:    condCmd.Value,
			Metadata: condCmd.Metadata,
		})
	}

	// Add step to campaign
	if err := camp.AddStep(step); err != nil {
		return fmt.Errorf("%w: %v", ErrStepCreationFailed, err)
	}

	return nil
}
