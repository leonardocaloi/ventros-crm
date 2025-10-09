package pipeline

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// StatusType define os tipos de status
type StatusType string

const (
	StatusTypeOpen   StatusType = "open"   // Status inicial/aberto
	StatusTypeActive StatusType = "active" // Status ativo/em progresso
	StatusTypeClosed StatusType = "closed" // Status fechado/finalizado
)

// Status representa um status dentro de um pipeline
type Status struct {
	id          uuid.UUID
	pipelineID  uuid.UUID
	name        string
	description string
	color       string
	statusType  StatusType
	position    int
	active      bool
	createdAt   time.Time
	updatedAt   time.Time

	// Domain Events
	events []DomainEvent
}

// NewStatus cria um novo status
func NewStatus(pipelineID uuid.UUID, name string, statusType StatusType) (*Status, error) {
	if pipelineID == uuid.Nil {
		return nil, errors.New("pipelineID cannot be nil")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if statusType == "" {
		return nil, errors.New("statusType cannot be empty")
	}

	now := time.Now()
	status := &Status{
		id:         uuid.New(),
		pipelineID: pipelineID,
		name:       name,
		statusType: statusType,
		position:   0,
		active:     true,
		createdAt:  now,
		updatedAt:  now,
		events:     []DomainEvent{},
	}

	status.addEvent(StatusCreatedEvent{
		StatusID:   status.id,
		PipelineID: pipelineID,
		Name:       name,
		StatusType: statusType,
		CreatedAt:  now,
	})

	return status, nil
}

// ReconstructStatus reconstrói um status a partir de dados persistidos
func ReconstructStatus(
	id, pipelineID uuid.UUID,
	name, description, color string,
	statusType StatusType,
	position int,
	active bool,
	createdAt, updatedAt time.Time,
) *Status {
	return &Status{
		id:          id,
		pipelineID:  pipelineID,
		name:        name,
		description: description,
		color:       color,
		statusType:  statusType,
		position:    position,
		active:      active,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      []DomainEvent{},
	}
}

// UpdateName atualiza o nome do status
func (s *Status) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	oldName := s.name
	s.name = name
	s.updatedAt = time.Now()

	s.addEvent(StatusUpdatedEvent{
		StatusID:  s.id,
		Field:     "name",
		OldValue:  oldName,
		NewValue:  name,
		UpdatedAt: s.updatedAt,
	})

	return nil
}

// UpdateDescription atualiza a descrição do status
func (s *Status) UpdateDescription(description string) {
	oldDescription := s.description
	s.description = description
	s.updatedAt = time.Now()

	s.addEvent(StatusUpdatedEvent{
		StatusID:  s.id,
		Field:     "description",
		OldValue:  oldDescription,
		NewValue:  description,
		UpdatedAt: s.updatedAt,
	})
}

// UpdateColor atualiza a cor do status
func (s *Status) UpdateColor(color string) {
	oldColor := s.color
	s.color = color
	s.updatedAt = time.Now()

	s.addEvent(StatusUpdatedEvent{
		StatusID:  s.id,
		Field:     "color",
		OldValue:  oldColor,
		NewValue:  color,
		UpdatedAt: s.updatedAt,
	})
}

// UpdatePosition atualiza a posição do status
func (s *Status) UpdatePosition(position int) {
	oldPosition := s.position
	s.position = position
	s.updatedAt = time.Now()

	s.addEvent(StatusUpdatedEvent{
		StatusID:  s.id,
		Field:     "position",
		OldValue:  oldPosition,
		NewValue:  position,
		UpdatedAt: s.updatedAt,
	})
}

// UpdateType atualiza o tipo do status
func (s *Status) UpdateType(statusType StatusType) error {
	if statusType == "" {
		return errors.New("statusType cannot be empty")
	}

	oldType := s.statusType
	s.statusType = statusType
	s.updatedAt = time.Now()

	s.addEvent(StatusUpdatedEvent{
		StatusID:  s.id,
		Field:     "status_type",
		OldValue:  string(oldType),
		NewValue:  string(statusType),
		UpdatedAt: s.updatedAt,
	})

	return nil
}

// Activate ativa o status
func (s *Status) Activate() {
	if !s.active {
		s.active = true
		s.updatedAt = time.Now()

		s.addEvent(StatusActivatedEvent{
			StatusID:    s.id,
			ActivatedAt: s.updatedAt,
		})
	}
}

// Deactivate desativa o status
func (s *Status) Deactivate() {
	if s.active {
		s.active = false
		s.updatedAt = time.Now()

		s.addEvent(StatusDeactivatedEvent{
			StatusID:      s.id,
			DeactivatedAt: s.updatedAt,
		})
	}
}

// IsOpen verifica se é um status de abertura
func (s *Status) IsOpen() bool {
	return s.statusType == StatusTypeOpen
}

// IsActive verifica se é um status ativo
func (s *Status) IsActiveType() bool {
	return s.statusType == StatusTypeActive
}

// IsClosed verifica se é um status de fechamento
func (s *Status) IsClosed() bool {
	return s.statusType == StatusTypeClosed
}

// Getters
func (s *Status) ID() uuid.UUID          { return s.id }
func (s *Status) PipelineID() uuid.UUID  { return s.pipelineID }
func (s *Status) Name() string           { return s.name }
func (s *Status) Description() string    { return s.description }
func (s *Status) Color() string          { return s.color }
func (s *Status) StatusType() StatusType { return s.statusType }
func (s *Status) Position() int          { return s.position }
func (s *Status) IsActiveStatus() bool   { return s.active }
func (s *Status) CreatedAt() time.Time   { return s.createdAt }
func (s *Status) UpdatedAt() time.Time   { return s.updatedAt }

// DomainEvents retorna os eventos de domínio
func (s *Status) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, s.events...)
}

// ClearEvents limpa os eventos (após publicação)
func (s *Status) ClearEvents() {
	s.events = []DomainEvent{}
}

// addEvent adiciona um evento de domínio
func (s *Status) addEvent(event DomainEvent) {
	s.events = append(s.events, event)
}
