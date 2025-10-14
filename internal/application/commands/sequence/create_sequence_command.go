package sequence

// CreateSequenceCommand comando para criar uma sequência
type CreateSequenceCommand struct {
	TenantID    string
	Name        string
	Description string
	TriggerType string
	Steps       []SequenceStepInput
}

// SequenceStepInput representa os dados de um step
type SequenceStepInput struct {
	Order           int
	Name            string
	DelayAmount     int
	DelayUnit       string
	MessageTemplate MessageTemplateInput
}

// MessageTemplateInput representa os dados de um template de mensagem
type MessageTemplateInput struct {
	Type       string
	Content    string
	TemplateID string
	Variables  map[string]string
	MediaURL   string
}

// Validate valida o comando
func (c *CreateSequenceCommand) Validate() error {
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	if c.Name == "" {
		return ErrNameRequired
	}
	if c.TriggerType == "" {
		return ErrTriggerTypeRequired
	}
	return nil
}
