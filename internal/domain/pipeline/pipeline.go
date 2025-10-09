package pipeline

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Pipeline representa um pipeline de vendas/atendimento
type Pipeline struct {
	id                    uuid.UUID
	projectID             uuid.UUID
	tenantID              string
	name                  string
	description           string
	color                 string
	position              int
	active                bool
	sessionTimeoutMinutes *int // Timeout de inatividade para sessões (NULL = herda do channel/project)
	statuses              []*Status
	createdAt             time.Time
	updatedAt             time.Time

	// Domain Events
	events []DomainEvent
}

// NewPipeline cria um novo pipeline
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
		id:                    uuid.New(),
		projectID:             projectID,
		tenantID:              tenantID,
		name:                  name,
		position:              0,
		active:                true,
		sessionTimeoutMinutes: nil, // NULL = herda do channel/project
		statuses:              []*Status{},
		createdAt:             now,
		updatedAt:             now,
		events:                []DomainEvent{},
	}

	pipeline.addEvent(PipelineCreatedEvent{
		PipelineID: pipeline.id,
		ProjectID:  projectID,
		TenantID:   tenantID,
		Name:       name,
		CreatedAt:  now,
	})

	return pipeline, nil
}

// ReconstructPipeline reconstrói um pipeline a partir de dados persistidos
func ReconstructPipeline(
	id, projectID uuid.UUID,
	tenantID, name, description, color string,
	position int,
	active bool,
	sessionTimeoutMinutes *int,
	createdAt, updatedAt time.Time,
) *Pipeline {
	return &Pipeline{
		id:                    id,
		projectID:             projectID,
		tenantID:              tenantID,
		name:                  name,
		description:           description,
		color:                 color,
		position:              position,
		active:                active,
		sessionTimeoutMinutes: sessionTimeoutMinutes, // NULL = inherit from channel/project
		statuses:              []*Status{},
		createdAt:             createdAt,
		updatedAt:             updatedAt,
		events:                []DomainEvent{},
	}
}

// UpdateName atualiza o nome do pipeline
func (p *Pipeline) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	oldName := p.name
	p.name = name
	p.updatedAt = time.Now()

	p.addEvent(PipelineUpdatedEvent{
		PipelineID: p.id,
		Field:      "name",
		OldValue:   oldName,
		NewValue:   name,
		UpdatedAt:  p.updatedAt,
	})

	return nil
}

// UpdateDescription atualiza a descrição do pipeline
func (p *Pipeline) UpdateDescription(description string) {
	oldDescription := p.description
	p.description = description
	p.updatedAt = time.Now()

	p.addEvent(PipelineUpdatedEvent{
		PipelineID: p.id,
		Field:      "description",
		OldValue:   oldDescription,
		NewValue:   description,
		UpdatedAt:  p.updatedAt,
	})
}

// UpdateColor atualiza a cor do pipeline
func (p *Pipeline) UpdateColor(color string) {
	oldColor := p.color
	p.color = color
	p.updatedAt = time.Now()

	p.addEvent(PipelineUpdatedEvent{
		PipelineID: p.id,
		Field:      "color",
		OldValue:   oldColor,
		NewValue:   color,
		UpdatedAt:  p.updatedAt,
	})
}

// UpdatePosition atualiza a posição do pipeline
func (p *Pipeline) UpdatePosition(position int) {
	oldPosition := p.position
	p.position = position
	p.updatedAt = time.Now()

	p.addEvent(PipelineUpdatedEvent{
		PipelineID: p.id,
		Field:      "position",
		OldValue:   oldPosition,
		NewValue:   position,
		UpdatedAt:  p.updatedAt,
	})
}

// Activate ativa o pipeline
func (p *Pipeline) Activate() {
	if !p.active {
		p.active = true
		p.updatedAt = time.Now()

		p.addEvent(PipelineActivatedEvent{
			PipelineID:  p.id,
			ActivatedAt: p.updatedAt,
		})
	}
}

// Deactivate desativa o pipeline
func (p *Pipeline) Deactivate() {
	if p.active {
		p.active = false
		p.updatedAt = time.Now()

		p.addEvent(PipelineDeactivatedEvent{
			PipelineID:    p.id,
			DeactivatedAt: p.updatedAt,
		})
	}
}

// AddStatus adiciona um status ao pipeline
func (p *Pipeline) AddStatus(status *Status) error {
	if status == nil {
		return errors.New("status cannot be nil")
	}

	// Verifica se já existe um status com o mesmo nome
	for _, s := range p.statuses {
		if s.Name() == status.Name() {
			return errors.New("status with this name already exists in pipeline")
		}
	}

	p.statuses = append(p.statuses, status)
	p.updatedAt = time.Now()

	p.addEvent(StatusAddedToPipelineEvent{
		PipelineID: p.id,
		StatusID:   status.ID(),
		StatusName: status.Name(),
		AddedAt:    p.updatedAt,
	})

	return nil
}

// RemoveStatus remove um status do pipeline
func (p *Pipeline) RemoveStatus(statusID uuid.UUID) error {
	for i, status := range p.statuses {
		if status.ID() == statusID {
			p.statuses = append(p.statuses[:i], p.statuses[i+1:]...)
			p.updatedAt = time.Now()

			p.addEvent(StatusRemovedFromPipelineEvent{
				PipelineID: p.id,
				StatusID:   statusID,
				StatusName: status.Name(),
				RemovedAt:  p.updatedAt,
			})

			return nil
		}
	}

	return errors.New("status not found in pipeline")
}

// GetStatusByID busca um status por ID
func (p *Pipeline) GetStatusByID(statusID uuid.UUID) *Status {
	for _, status := range p.statuses {
		if status.ID() == statusID {
			return status
		}
	}
	return nil
}

// GetStatusByName busca um status por nome
func (p *Pipeline) GetStatusByName(name string) *Status {
	for _, status := range p.statuses {
		if status.Name() == name {
			return status
		}
	}
	return nil
}

// Getters
func (p *Pipeline) ID() uuid.UUID              { return p.id }
func (p *Pipeline) ProjectID() uuid.UUID       { return p.projectID }
func (p *Pipeline) TenantID() string           { return p.tenantID }
func (p *Pipeline) Name() string               { return p.name }
func (p *Pipeline) Description() string        { return p.description }
func (p *Pipeline) Color() string              { return p.color }
func (p *Pipeline) Position() int              { return p.position }
func (p *Pipeline) IsActive() bool             { return p.active }
func (p *Pipeline) SessionTimeoutMinutes() *int { return p.sessionTimeoutMinutes }
func (p *Pipeline) Statuses() []*Status        { return append([]*Status{}, p.statuses...) } // Copy
func (p *Pipeline) CreatedAt() time.Time       { return p.createdAt }
func (p *Pipeline) UpdatedAt() time.Time       { return p.updatedAt }

// SetSessionTimeout atualiza o timeout de sessão (em minutos)
// Pass nil to inherit from channel/project, or a value to override
func (p *Pipeline) SetSessionTimeout(minutes *int) error {
	if minutes != nil {
		if *minutes <= 0 {
			return errors.New("session timeout must be greater than 0")
		}
		if *minutes > 1440 { // Máximo 24 horas
			return errors.New("session timeout cannot exceed 1440 minutes (24 hours)")
		}
	}

	p.sessionTimeoutMinutes = minutes
	p.updatedAt = time.Now()

	return nil
}

// DomainEvents retorna os eventos de domínio
func (p *Pipeline) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, p.events...)
}

// ClearEvents limpa os eventos (após publicação)
func (p *Pipeline) ClearEvents() {
	p.events = []DomainEvent{}
}

// addEvent adiciona um evento de domínio
func (p *Pipeline) addEvent(event DomainEvent) {
	p.events = append(p.events, event)
}
