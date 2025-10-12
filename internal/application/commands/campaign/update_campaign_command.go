package campaign

import "github.com/google/uuid"

// UpdateCampaignCommand comando para atualizar uma campanha
type UpdateCampaignCommand struct {
	CampaignID  uuid.UUID
	TenantID    string
	Name        *string
	Description *string
	GoalType    *string
	GoalValue   *int
}

// Validate valida o comando
func (c *UpdateCampaignCommand) Validate() error {
	if c.CampaignID == uuid.Nil {
		return ErrCampaignIDRequired
	}
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	// At least one field must be provided for update
	if c.Name == nil && c.Description == nil && c.GoalType == nil && c.GoalValue == nil {
		return ErrNoFieldsToUpdate
	}
	return nil
}
