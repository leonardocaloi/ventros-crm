package pipeline

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

type StatusType string

const (
	StatusTypeOpen   StatusType = "open"
	StatusTypeActive StatusType = "active"
	StatusTypeClosed StatusType = "closed"
)

type Status struct {
	id          uuid.UUID
	version     int // Optimistic locking - prevents lost updates
	pipelineID  uuid.UUID
	name        string
	description string
	color       string
	statusType  StatusType
	position    int
	active      bool
	createdAt   time.Time
	updatedAt   time.Time

	events []shared.DomainEvent
}

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
		version:    1, // Start with version 1 for new aggregates
		pipelineID: pipelineID,
		name:       name,
		statusType: statusType,
		position:   0,
		active:     true,
		createdAt:  now,
		updatedAt:  now,
		events:     []shared.DomainEvent{},
	}

	status.addEvent(NewStatusCreatedEvent(status.id, pipelineID, name, statusType))

	return status, nil
}

func ReconstructStatus(
	id, pipelineID uuid.UUID,
	version int, // Optimistic locking version
	name, description, color string,
	statusType StatusType,
	position int,
	active bool,
	createdAt, updatedAt time.Time,
) *Status {
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}

	return &Status{
		id:          id,
		version:     version,
		pipelineID:  pipelineID,
		name:        name,
		description: description,
		color:       color,
		statusType:  statusType,
		position:    position,
		active:      active,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      []shared.DomainEvent{},
	}
}

func (s *Status) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	oldName := s.name
	s.name = name
	s.updatedAt = time.Now()

	s.addEvent(NewStatusUpdatedEvent(s.id, "name", oldName, name))

	return nil
}

func (s *Status) UpdateDescription(description string) {
	oldDescription := s.description
	s.description = description
	s.updatedAt = time.Now()

	s.addEvent(NewStatusUpdatedEvent(s.id, "description", oldDescription, description))
}

func (s *Status) UpdateColor(color string) {
	oldColor := s.color
	s.color = color
	s.updatedAt = time.Now()

	s.addEvent(NewStatusUpdatedEvent(s.id, "color", oldColor, color))
}

func (s *Status) UpdatePosition(position int) {
	oldPosition := s.position
	s.position = position
	s.updatedAt = time.Now()

	s.addEvent(NewStatusUpdatedEvent(s.id, "position", oldPosition, position))
}

func (s *Status) UpdateType(statusType StatusType) error {
	if statusType == "" {
		return errors.New("statusType cannot be empty")
	}

	oldType := s.statusType
	s.statusType = statusType
	s.updatedAt = time.Now()

	s.addEvent(NewStatusUpdatedEvent(s.id, "status_type", string(oldType), string(statusType)))

	return nil
}

func (s *Status) Activate() {
	if !s.active {
		s.active = true
		s.updatedAt = time.Now()

		s.addEvent(NewStatusActivatedEvent(s.id))
	}
}

func (s *Status) Deactivate() {
	if s.active {
		s.active = false
		s.updatedAt = time.Now()

		s.addEvent(NewStatusDeactivatedEvent(s.id))
	}
}

func (s *Status) IsOpen() bool {
	return s.statusType == StatusTypeOpen
}

func (s *Status) IsActiveType() bool {
	return s.statusType == StatusTypeActive
}

func (s *Status) IsClosed() bool {
	return s.statusType == StatusTypeClosed
}

func (s *Status) ID() uuid.UUID          { return s.id }
func (s *Status) Version() int           { return s.version }
func (s *Status) PipelineID() uuid.UUID  { return s.pipelineID }
func (s *Status) Name() string           { return s.name }
func (s *Status) Description() string    { return s.description }
func (s *Status) Color() string          { return s.color }
func (s *Status) StatusType() StatusType { return s.statusType }
func (s *Status) Position() int          { return s.position }
func (s *Status) IsActiveStatus() bool   { return s.active }
func (s *Status) CreatedAt() time.Time   { return s.createdAt }
func (s *Status) UpdatedAt() time.Time   { return s.updatedAt }

func (s *Status) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, s.events...)
}

func (s *Status) ClearEvents() {
	s.events = []shared.DomainEvent{}
}

func (s *Status) addEvent(event shared.DomainEvent) {
	s.events = append(s.events, event)
}

// Compile-time check that Status implements AggregateRoot interface
var _ shared.AggregateRoot = (*Status)(nil)
