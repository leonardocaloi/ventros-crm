package campaign

import (
	"time"

	"github.com/google/uuid"
)

// CampaignStep represents a single step in a campaign
type CampaignStep struct {
	ID         uuid.UUID
	Order      int
	Name       string
	Type       StepType
	Config     StepConfig
	Conditions []StepCondition
	CreatedAt  time.Time
}

type StepType string

const (
	StepTypeBroadcast StepType = "broadcast"
	StepTypeSequence  StepType = "sequence"
	StepTypeDelay     StepType = "delay"
	StepTypeCondition StepType = "condition"
	StepTypeWait      StepType = "wait"
)

// StepConfig holds configuration for different step types
type StepConfig struct {
	// For broadcast steps
	BroadcastID *uuid.UUID `json:"broadcast_id,omitempty"`

	// For sequence steps
	SequenceID *uuid.UUID `json:"sequence_id,omitempty"`

	// For delay/wait steps
	DelayAmount *int    `json:"delay_amount,omitempty"`
	DelayUnit   *string `json:"delay_unit,omitempty"` // minutes, hours, days

	// For condition steps
	ConditionType *string                `json:"condition_type,omitempty"` // tag_has, field_equals, etc.
	ConditionData map[string]interface{} `json:"condition_data,omitempty"`

	// For wait steps (wait for user action)
	WaitFor     *string `json:"wait_for,omitempty"`     // reply, click, open
	WaitTimeout *int    `json:"wait_timeout,omitempty"` // timeout in hours
	TimeoutStep *int    `json:"timeout_step,omitempty"` // step to jump to on timeout
}

// StepCondition represents a condition that must be met for the step to execute
type StepCondition struct {
	Type     string                 `json:"type"`     // tag_has, field_equals, pipeline_status, etc.
	Field    string                 `json:"field"`    // field to check
	Operator string                 `json:"operator"` // equals, contains, greater_than, etc.
	Value    interface{}            `json:"value"`    // value to compare against
	Metadata map[string]interface{} `json:"metadata"` // additional metadata
}

// NewCampaignStep creates a new campaign step
func NewCampaignStep(order int, name string, stepType StepType, config StepConfig) CampaignStep {
	return CampaignStep{
		ID:         uuid.New(),
		Order:      order,
		Name:       name,
		Type:       stepType,
		Config:     config,
		Conditions: []StepCondition{},
		CreatedAt:  time.Now(),
	}
}

// AddCondition adds a condition to the step
func (s *CampaignStep) AddCondition(condition StepCondition) {
	s.Conditions = append(s.Conditions, condition)
}

// GetDelayDuration calculates the delay duration for delay/wait steps
func (s *CampaignStep) GetDelayDuration() time.Duration {
	if s.Type != StepTypeDelay && s.Type != StepTypeWait {
		return 0
	}

	if s.Config.DelayAmount == nil || s.Config.DelayUnit == nil {
		return 0
	}

	amount := *s.Config.DelayAmount
	unit := *s.Config.DelayUnit

	switch unit {
	case "minutes":
		return time.Duration(amount) * time.Minute
	case "hours":
		return time.Duration(amount) * time.Hour
	case "days":
		return time.Duration(amount) * 24 * time.Hour
	default:
		return 0
	}
}

// Validate validates the step configuration
func (s *CampaignStep) Validate() error {
	switch s.Type {
	case StepTypeBroadcast:
		if s.Config.BroadcastID == nil {
			return ErrInvalidStepConfig
		}
	case StepTypeSequence:
		if s.Config.SequenceID == nil {
			return ErrInvalidStepConfig
		}
	case StepTypeDelay, StepTypeWait:
		if s.Config.DelayAmount == nil || s.Config.DelayUnit == nil {
			return ErrInvalidStepConfig
		}
	case StepTypeCondition:
		if s.Config.ConditionType == nil {
			return ErrInvalidStepConfig
		}
	}
	return nil
}
