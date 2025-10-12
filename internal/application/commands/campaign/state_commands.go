package campaign

import "github.com/google/uuid"

// ActivateCampaignCommand comando para ativar uma campanha
type ActivateCampaignCommand struct {
	CampaignID uuid.UUID
	TenantID   string
}

// Validate valida o comando
func (c *ActivateCampaignCommand) Validate() error {
	if c.CampaignID == uuid.Nil {
		return ErrCampaignIDRequired
	}
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	return nil
}

// PauseCampaignCommand comando para pausar uma campanha
type PauseCampaignCommand struct {
	CampaignID uuid.UUID
	TenantID   string
}

// Validate valida o comando
func (c *PauseCampaignCommand) Validate() error {
	if c.CampaignID == uuid.Nil {
		return ErrCampaignIDRequired
	}
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	return nil
}

// CompleteCampaignCommand comando para completar uma campanha
type CompleteCampaignCommand struct {
	CampaignID uuid.UUID
	TenantID   string
}

// Validate valida o comando
func (c *CompleteCampaignCommand) Validate() error {
	if c.CampaignID == uuid.Nil {
		return ErrCampaignIDRequired
	}
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	return nil
}
