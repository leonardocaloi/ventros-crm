package campaign

import (
	"github.com/google/uuid"
)

// CreateCampaignCommand comando para criar uma nova campanha
type CreateCampaignCommand struct {
	TenantID    string
	Name        string
	Description string
	GoalType    string
	GoalValue   int
	Steps       []CreateCampaignStepCommand
}

// CreateCampaignStepCommand comando para criar um step da campanha
type CreateCampaignStepCommand struct {
	Order      int
	Name       string
	Type       string
	Config     StepConfigCommand
	Conditions []StepConditionCommand
}

// StepConfigCommand configuração do step
type StepConfigCommand struct {
	BroadcastID   *uuid.UUID
	SequenceID    *uuid.UUID
	DelayAmount   *int
	DelayUnit     *string
	ConditionType *string
	ConditionData map[string]interface{}
	WaitFor       *string
	WaitTimeout   *int
	TimeoutStep   *int
}

// StepConditionCommand condição do step
type StepConditionCommand struct {
	Type     string
	Field    string
	Operator string
	Value    interface{}
	Metadata map[string]interface{}
}

// Validate valida o comando
func (c *CreateCampaignCommand) Validate() error {
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	if c.Name == "" {
		return ErrCampaignNameRequired
	}
	if c.GoalType == "" {
		return ErrGoalTypeRequired
	}
	return nil
}
