package sequence

import (
	"time"

	"github.com/google/uuid"
)

// SequenceStep representa um passo na sequência
type SequenceStep struct {
	ID              uuid.UUID
	Order           int             // Ordem do step (0, 1, 2...)
	Name            string          // Nome do step (ex: "Mensagem de boas-vindas")
	DelayAmount     int             // Quantidade de tempo para esperar
	DelayUnit       DelayUnit       // Unidade de tempo (minutes, hours, days)
	MessageTemplate MessageTemplate // Template da mensagem
	Conditions      []StepCondition // Condições para enviar este step
	CreatedAt       time.Time
}

type DelayUnit string

const (
	DelayUnitMinutes DelayUnit = "minutes"
	DelayUnitHours   DelayUnit = "hours"
	DelayUnitDays    DelayUnit = "days"
)

// MessageTemplate template da mensagem com variáveis
type MessageTemplate struct {
	Type       string            `json:"type"` // text, template, media
	Content    string            `json:"content"`
	TemplateID *string           `json:"template_id,omitempty"`
	Variables  map[string]string `json:"variables,omitempty"`
	MediaURL   *string           `json:"media_url,omitempty"`
}

// StepCondition condição para enviar o step
type StepCondition struct {
	Type     ConditionType `json:"type"`
	Operator string        `json:"operator"` // equals, contains, greater_than, etc
	Value    string        `json:"value"`
}

type ConditionType string

const (
	ConditionTypeTag            ConditionType = "tag"             // Tem tag específica
	ConditionTypeCustomField    ConditionType = "custom_field"    // Campo customizado
	ConditionTypePipelineStatus ConditionType = "pipeline_status" // Status no pipeline
	ConditionTypeLastActivity   ConditionType = "last_activity"   // Última atividade
)

// GetDelayDuration returns the delay as a time.Duration
func (s *SequenceStep) GetDelayDuration() time.Duration {
	switch s.DelayUnit {
	case DelayUnitMinutes:
		return time.Duration(s.DelayAmount) * time.Minute
	case DelayUnitHours:
		return time.Duration(s.DelayAmount) * time.Hour
	case DelayUnitDays:
		return time.Duration(s.DelayAmount) * 24 * time.Hour
	default:
		return 0
	}
}

// NewSequenceStep creates a new sequence step
func NewSequenceStep(
	order int,
	name string,
	delayAmount int,
	delayUnit DelayUnit,
	messageTemplate MessageTemplate,
) SequenceStep {
	return SequenceStep{
		ID:              uuid.New(),
		Order:           order,
		Name:            name,
		DelayAmount:     delayAmount,
		DelayUnit:       delayUnit,
		MessageTemplate: messageTemplate,
		Conditions:      []StepCondition{},
		CreatedAt:       time.Now(),
	}
}
