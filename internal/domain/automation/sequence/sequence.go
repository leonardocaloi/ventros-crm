package sequence

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

// Sequence representa uma sequência automatizada de mensagens
type Sequence struct {
	id             uuid.UUID
	version        int // Optimistic locking - prevents lost updates
	tenantID       string
	name           string
	description    string
	status         SequenceStatus
	steps          []SequenceStep // Steps ordenados

	// Entry conditions
	triggerType TriggerType            // manual, tag_added, list_joined, etc
	triggerData map[string]interface{} // Dados específicos do trigger

	// Exit conditions
	exitOnReply bool // Sai da sequence se contato responder

	// Stats
	totalEnrolled  int
	activeCount    int
	completedCount int
	exitedCount    int

	createdAt time.Time
	updatedAt time.Time

	events []shared.DomainEvent
}

type SequenceStatus string

const (
	SequenceStatusDraft    SequenceStatus = "draft"    // Rascunho
	SequenceStatusActive   SequenceStatus = "active"   // Ativa
	SequenceStatusPaused   SequenceStatus = "paused"   // Pausada
	SequenceStatusArchived SequenceStatus = "archived" // Arquivada
)

type TriggerType string

const (
	TriggerTypeManual          TriggerType = "manual"           // Entrada manual
	TriggerTypeTagAdded        TriggerType = "tag_added"        // Quando tag é adicionada
	TriggerTypeListJoined      TriggerType = "list_joined"      // Quando entra em lista
	TriggerTypeFormSubmit      TriggerType = "form_submit"      // Quando submete formulário
	TriggerTypePipelineEntered TriggerType = "pipeline_entered" // Quando entra em pipeline
)

// NewSequence creates a new sequence
func NewSequence(
	tenantID string,
	name string,
	description string,
	triggerType TriggerType,
) (*Sequence, error) {
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	now := time.Now()
	sequence := &Sequence{
		id:             uuid.New(),
		version:        1, // Start with version 1 for new aggregates
		tenantID:       tenantID,
		name:           name,
		description:    description,
		status:         SequenceStatusDraft,
		triggerType:    triggerType,
		triggerData:    make(map[string]interface{}),
		steps:          []SequenceStep{},
		exitOnReply:    true, // Default: sair se responder
		createdAt:      now,
		updatedAt:      now,
		events:         []shared.DomainEvent{},
	}

	sequence.addEvent(NewSequenceCreatedEvent(sequence.id, tenantID, name))

	return sequence, nil
}

// ReconstructSequence reconstructs a sequence from persistence
func ReconstructSequence(
	id uuid.UUID,
	version int, // Optimistic locking version
	tenantID string,
	name string,
	description string,
	status SequenceStatus,
	steps []SequenceStep,
	triggerType TriggerType,
	triggerData map[string]interface{},
	exitOnReply bool,
	totalEnrolled, activeCount, completedCount, exitedCount int,
	createdAt, updatedAt time.Time,
) *Sequence {
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}

	return &Sequence{
		id:             id,
		version:        version,
		tenantID:       tenantID,
		name:           name,
		description:    description,
		status:         status,
		steps:          steps,
		triggerType:    triggerType,
		triggerData:    triggerData,
		exitOnReply:    exitOnReply,
		totalEnrolled:  totalEnrolled,
		activeCount:    activeCount,
		completedCount: completedCount,
		exitedCount:    exitedCount,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		events:         []shared.DomainEvent{},
	}
}

// AddStep adds a new step to the sequence
func (s *Sequence) AddStep(step SequenceStep) error {
	if s.status != SequenceStatusDraft {
		return errors.New("can only add steps to draft sequences")
	}
	if step.Order < 0 {
		return errors.New("step order must be non-negative")
	}

	// Ensure unique order
	for _, existingStep := range s.steps {
		if existingStep.Order == step.Order {
			return errors.New("step order must be unique")
		}
	}

	step.ID = uuid.New()
	s.steps = append(s.steps, step)
	s.updatedAt = time.Now()

	s.addEvent(NewSequenceStepAddedEvent(s.id, step.ID, step.Order))

	return nil
}

// UpdateStep updates an existing step
func (s *Sequence) UpdateStep(stepID uuid.UUID, updatedStep SequenceStep) error {
	if s.status != SequenceStatusDraft {
		return errors.New("can only update steps in draft sequences")
	}

	for i, step := range s.steps {
		if step.ID == stepID {
			updatedStep.ID = stepID
			s.steps[i] = updatedStep
			s.updatedAt = time.Now()
			return nil
		}
	}

	return errors.New("step not found")
}

// RemoveStep removes a step from the sequence
func (s *Sequence) RemoveStep(stepID uuid.UUID) error {
	if s.status != SequenceStatusDraft {
		return errors.New("can only remove steps from draft sequences")
	}

	for i, step := range s.steps {
		if step.ID == stepID {
			s.steps = append(s.steps[:i], s.steps[i+1:]...)
			s.updatedAt = time.Now()
			return nil
		}
	}

	return errors.New("step not found")
}

// Activate activates the sequence
func (s *Sequence) Activate() error {
	if s.status == SequenceStatusActive {
		return errors.New("sequence is already active")
	}
	if len(s.steps) == 0 {
		return errors.New("cannot activate sequence without steps")
	}

	s.status = SequenceStatusActive
	s.updatedAt = time.Now()

	s.addEvent(NewSequenceActivatedEvent(s.id))

	return nil
}

// Pause pauses the sequence
func (s *Sequence) Pause() error {
	if s.status != SequenceStatusActive {
		return errors.New("can only pause active sequences")
	}

	s.status = SequenceStatusPaused
	s.updatedAt = time.Now()

	s.addEvent(NewSequencePausedEvent(s.id))

	return nil
}

// Resume resumes a paused sequence
func (s *Sequence) Resume() error {
	if s.status != SequenceStatusPaused {
		return errors.New("can only resume paused sequences")
	}

	s.status = SequenceStatusActive
	s.updatedAt = time.Now()

	s.addEvent(NewSequenceResumedEvent(s.id))

	return nil
}

// Archive archives the sequence
func (s *Sequence) Archive() error {
	if s.status == SequenceStatusArchived {
		return errors.New("sequence is already archived")
	}

	s.status = SequenceStatusArchived
	s.updatedAt = time.Now()

	s.addEvent(NewSequenceArchivedEvent(s.id))

	return nil
}

// UpdateName updates the sequence name
func (s *Sequence) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	s.name = name
	s.updatedAt = time.Now()
	return nil
}

// UpdateDescription updates the sequence description
func (s *Sequence) UpdateDescription(description string) {
	s.description = description
	s.updatedAt = time.Now()
}

// UpdateExitOnReply updates the exit on reply setting
func (s *Sequence) UpdateExitOnReply(exitOnReply bool) {
	s.exitOnReply = exitOnReply
	s.updatedAt = time.Now()
}

// IncrementEnrolled increments enrolled count
func (s *Sequence) IncrementEnrolled() {
	s.totalEnrolled++
	s.activeCount++
	s.updatedAt = time.Now()
}

// MarkCompleted marks an enrollment as completed
func (s *Sequence) MarkCompleted() {
	s.activeCount--
	s.completedCount++
	s.updatedAt = time.Now()
}

// MarkExited marks an enrollment as exited
func (s *Sequence) MarkExited() {
	s.activeCount--
	s.exitedCount++
	s.updatedAt = time.Now()
}

// GetStats returns sequence statistics
func (s *Sequence) GetStats() SequenceStats {
	completionRate := 0.0
	if s.totalEnrolled > 0 {
		completionRate = float64(s.completedCount) / float64(s.totalEnrolled) * 100
	}

	return SequenceStats{
		TotalEnrolled:  s.totalEnrolled,
		ActiveCount:    s.activeCount,
		CompletedCount: s.completedCount,
		ExitedCount:    s.exitedCount,
		CompletionRate: completionRate,
	}
}

// SequenceStats represents sequence statistics
type SequenceStats struct {
	TotalEnrolled  int     `json:"total_enrolled"`
	ActiveCount    int     `json:"active_count"`
	CompletedCount int     `json:"completed_count"`
	ExitedCount    int     `json:"exited_count"`
	CompletionRate float64 `json:"completion_rate"`
}

// GetStepByOrder returns a step by its order
func (s *Sequence) GetStepByOrder(order int) (*SequenceStep, error) {
	for _, step := range s.steps {
		if step.Order == order {
			return &step, nil
		}
	}
	return nil, errors.New("step not found")
}

// GetNextStep returns the next step after the given order
func (s *Sequence) GetNextStep(currentOrder int) (*SequenceStep, error) {
	nextOrder := currentOrder + 1
	for _, step := range s.steps {
		if step.Order == nextOrder {
			return &step, nil
		}
	}
	return nil, nil // No more steps
}

// Getters
func (s *Sequence) ID() uuid.UUID                       { return s.id }
func (s *Sequence) Version() int                        { return s.version }
func (s *Sequence) TenantID() string                    { return s.tenantID }
func (s *Sequence) Name() string                        { return s.name }
func (s *Sequence) Description() string                 { return s.description }
func (s *Sequence) Status() SequenceStatus              { return s.status }
func (s *Sequence) Steps() []SequenceStep               { return append([]SequenceStep{}, s.steps...) }
func (s *Sequence) TriggerType() TriggerType            { return s.triggerType }
func (s *Sequence) TriggerData() map[string]interface{} { return s.triggerData }
func (s *Sequence) ExitOnReply() bool                   { return s.exitOnReply }
func (s *Sequence) TotalEnrolled() int                  { return s.totalEnrolled }
func (s *Sequence) ActiveCount() int                    { return s.activeCount }
func (s *Sequence) CompletedCount() int                 { return s.completedCount }
func (s *Sequence) ExitedCount() int                    { return s.exitedCount }
func (s *Sequence) CreatedAt() time.Time                { return s.createdAt }
func (s *Sequence) UpdatedAt() time.Time                { return s.updatedAt }

func (s *Sequence) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, s.events...)
}

func (s *Sequence) ClearEvents() {
	s.events = []shared.DomainEvent{}
}

func (s *Sequence) addEvent(event shared.DomainEvent) {
	s.events = append(s.events, event)
}

// Repository interface
type Repository interface {
	Save(sequence *Sequence) error
	FindByID(id uuid.UUID) (*Sequence, error)
	FindByTenantID(tenantID string) ([]*Sequence, error)
	FindActiveByTriggerType(triggerType TriggerType) ([]*Sequence, error)
	FindByStatus(status SequenceStatus) ([]*Sequence, error)
	Delete(id uuid.UUID) error
}

// Compile-time check that Sequence implements AggregateRoot interface
var _ shared.AggregateRoot = (*Sequence)(nil)
