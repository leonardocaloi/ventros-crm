package pipeline

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type Pipeline struct {
	id                      uuid.UUID
	version                 int // Optimistic locking - prevents lost updates
	projectID               uuid.UUID
	tenantID                string
	name                    string
	description             string
	color                   string
	position                int
	active                  bool
	sessionTimeoutMinutes   *int
	leadQualificationConfig *LeadQualificationConfig // Qualificação automática por foto
	statuses                []*Status
	createdAt               time.Time
	updatedAt               time.Time

	events []shared.DomainEvent
}

func NewPipeline(projectID uuid.UUID, tenantID, name string) (*Pipeline, error) {
	if projectID == uuid.Nil {
		return nil, errors.New("projectID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	now := time.Now()
	pipeline := &Pipeline{
		id:                      uuid.New(),
		version:                 1, // Start with version 1 for new aggregates
		projectID:               projectID,
		tenantID:                tenantID,
		name:                    name,
		position:                0,
		active:                  true,
		sessionTimeoutMinutes:   nil,
		leadQualificationConfig: nil, // Desabilitado por padrão
		statuses:                []*Status{},
		createdAt:               now,
		updatedAt:               now,
		events:                  []shared.DomainEvent{},
	}

	pipeline.addEvent(NewPipelineCreatedEvent(pipeline.id, projectID, tenantID, name))

	return pipeline, nil
}

func ReconstructPipeline(
	id, projectID uuid.UUID,
	version int, // Optimistic locking version
	tenantID, name, description, color string,
	position int,
	active bool,
	sessionTimeoutMinutes *int,
	leadQualificationConfig *LeadQualificationConfig,
	createdAt, updatedAt time.Time,
) *Pipeline {
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}

	return &Pipeline{
		id:                      id,
		version:                 version,
		projectID:               projectID,
		tenantID:                tenantID,
		name:                    name,
		description:             description,
		color:                   color,
		position:                position,
		active:                  active,
		sessionTimeoutMinutes:   sessionTimeoutMinutes,
		leadQualificationConfig: leadQualificationConfig,
		statuses:                []*Status{},
		createdAt:               createdAt,
		updatedAt:               updatedAt,
		events:                  []shared.DomainEvent{},
	}
}

func (p *Pipeline) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	oldName := p.name
	p.name = name
	p.updatedAt = time.Now()

	p.addEvent(NewPipelineUpdatedEvent(p.id, "name", oldName, name))

	return nil
}

func (p *Pipeline) UpdateDescription(description string) {
	oldDescription := p.description
	p.description = description
	p.updatedAt = time.Now()

	p.addEvent(NewPipelineUpdatedEvent(p.id, "description", oldDescription, description))
}

func (p *Pipeline) UpdateColor(color string) {
	oldColor := p.color
	p.color = color
	p.updatedAt = time.Now()

	p.addEvent(NewPipelineUpdatedEvent(p.id, "color", oldColor, color))
}

func (p *Pipeline) UpdatePosition(position int) {
	oldPosition := p.position
	p.position = position
	p.updatedAt = time.Now()

	p.addEvent(NewPipelineUpdatedEvent(p.id, "position", oldPosition, position))
}

func (p *Pipeline) Activate() {
	if !p.active {
		p.active = true
		p.updatedAt = time.Now()

		p.addEvent(NewPipelineActivatedEvent(p.id))
	}
}

func (p *Pipeline) Deactivate() {
	if p.active {
		p.active = false
		p.updatedAt = time.Now()

		p.addEvent(NewPipelineDeactivatedEvent(p.id))
	}
}

func (p *Pipeline) AddStatus(status *Status) error {
	if status == nil {
		return errors.New("status cannot be nil")
	}

	for _, s := range p.statuses {
		if s.Name() == status.Name() {
			return errors.New("status with this name already exists in pipeline")
		}
	}

	p.statuses = append(p.statuses, status)
	p.updatedAt = time.Now()

	p.addEvent(NewStatusAddedToPipelineEvent(p.id, status.ID(), status.Name()))

	return nil
}

func (p *Pipeline) RemoveStatus(statusID uuid.UUID) error {
	for i, status := range p.statuses {
		if status.ID() == statusID {
			p.statuses = append(p.statuses[:i], p.statuses[i+1:]...)
			p.updatedAt = time.Now()

			p.addEvent(NewStatusRemovedFromPipelineEvent(p.id, statusID, status.Name()))

			return nil
		}
	}

	return errors.New("status not found in pipeline")
}

func (p *Pipeline) GetStatusByID(statusID uuid.UUID) *Status {
	for _, status := range p.statuses {
		if status.ID() == statusID {
			return status
		}
	}
	return nil
}

func (p *Pipeline) GetStatusByName(name string) *Status {
	for _, status := range p.statuses {
		if status.Name() == name {
			return status
		}
	}
	return nil
}

func (p *Pipeline) ID() uuid.UUID               { return p.id }
func (p *Pipeline) Version() int                { return p.version }
func (p *Pipeline) ProjectID() uuid.UUID        { return p.projectID }
func (p *Pipeline) TenantID() string            { return p.tenantID }
func (p *Pipeline) Name() string                { return p.name }
func (p *Pipeline) Description() string         { return p.description }
func (p *Pipeline) Color() string               { return p.color }
func (p *Pipeline) Position() int               { return p.position }
func (p *Pipeline) IsActive() bool              { return p.active }
func (p *Pipeline) SessionTimeoutMinutes() *int { return p.sessionTimeoutMinutes }
func (p *Pipeline) LeadQualificationConfig() *LeadQualificationConfig {
	return p.leadQualificationConfig
}
func (p *Pipeline) Statuses() []*Status  { return append([]*Status{}, p.statuses...) }
func (p *Pipeline) CreatedAt() time.Time { return p.createdAt }
func (p *Pipeline) UpdatedAt() time.Time { return p.updatedAt }

func (p *Pipeline) SetSessionTimeout(minutes *int) error {
	if minutes != nil {
		if *minutes <= 0 {
			return errors.New("session timeout must be greater than 0")
		}
		if *minutes > 1440 {
			return errors.New("session timeout cannot exceed 1440 minutes (24 hours)")
		}
	}

	p.sessionTimeoutMinutes = minutes
	p.updatedAt = time.Now()

	return nil
}

// EnableLeadQualification ativa qualificação automática com config padrão
func (p *Pipeline) EnableLeadQualification() {
	if p.leadQualificationConfig == nil {
		p.leadQualificationConfig = NewLeadQualificationConfigWithDefaults()
	}
	p.leadQualificationConfig.Enable()
	p.updatedAt = time.Now()

	p.addEvent(NewLeadQualificationEnabledEvent(p.id))
}

// DisableLeadQualification desativa qualificação automática
func (p *Pipeline) DisableLeadQualification() {
	if p.leadQualificationConfig != nil {
		p.leadQualificationConfig.Disable()
		p.updatedAt = time.Now()

		p.addEvent(NewLeadQualificationDisabledEvent(p.id))
	}
}

// SetLeadQualificationConfig define uma config customizada
func (p *Pipeline) SetLeadQualificationConfig(config *LeadQualificationConfig) {
	p.leadQualificationConfig = config
	p.updatedAt = time.Now()

	p.addEvent(NewLeadQualificationConfigUpdatedEvent(p.id))
}

// HasLeadQualification verifica se está ativado
func (p *Pipeline) HasLeadQualification() bool {
	return p.leadQualificationConfig != nil && p.leadQualificationConfig.IsEnabled()
}

func (p *Pipeline) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, p.events...)
}

func (p *Pipeline) ClearEvents() {
	p.events = []shared.DomainEvent{}
}

func (p *Pipeline) addEvent(event shared.DomainEvent) {
	p.events = append(p.events, event)
}

// Compile-time check that Pipeline implements AggregateRoot interface
var _ shared.AggregateRoot = (*Pipeline)(nil)
