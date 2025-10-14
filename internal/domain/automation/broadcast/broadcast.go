package broadcast

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

// Broadcast representa um disparo em massa
type Broadcast struct {
	id              uuid.UUID
	version         int // Optimistic locking - prevents lost updates
	tenantID        string
	name            string
	listID          uuid.UUID       // Lista de contatos alvo
	messageTemplate MessageTemplate // Template da mensagem
	status          BroadcastStatus
	scheduledFor    *time.Time // Quando disparar (nil = imediato)
	startedAt       *time.Time
	completedAt     *time.Time

	// Stats
	totalContacts int
	sentCount     int
	failedCount   int
	pendingCount  int

	// Rate limiting
	rateLimit int // mensagens por minuto (0 = sem limite)

	createdAt time.Time
	updatedAt time.Time

	events []shared.DomainEvent
}

type BroadcastStatus string

const (
	BroadcastStatusDraft     BroadcastStatus = "draft"     // Rascunho
	BroadcastStatusScheduled BroadcastStatus = "scheduled" // Agendado
	BroadcastStatusRunning   BroadcastStatus = "running"   // Em execução
	BroadcastStatusCompleted BroadcastStatus = "completed" // Concluído
	BroadcastStatusFailed    BroadcastStatus = "failed"    // Falhou
	BroadcastStatusCancelled BroadcastStatus = "cancelled" // Cancelado
)

// MessageTemplate template da mensagem com variáveis
type MessageTemplate struct {
	Type       string            `json:"type"` // text, template, media
	Content    string            `json:"content"`
	TemplateID *string           `json:"template_id,omitempty"`
	Variables  map[string]string `json:"variables,omitempty"`
	MediaURL   *string           `json:"media_url,omitempty"`
}

// NewBroadcast creates a new broadcast
func NewBroadcast(
	tenantID string,
	name string,
	listID uuid.UUID,
	messageTemplate MessageTemplate,
	rateLimit int,
) (*Broadcast, error) {
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if listID == uuid.Nil {
		return nil, errors.New("listID cannot be empty")
	}
	if messageTemplate.Type == "" {
		return nil, errors.New("messageTemplate type cannot be empty")
	}
	if messageTemplate.Content == "" && messageTemplate.TemplateID == nil {
		return nil, errors.New("messageTemplate must have either content or template_id")
	}
	if rateLimit < 0 {
		return nil, errors.New("rateLimit cannot be negative")
	}

	now := time.Now()
	broadcast := &Broadcast{
		id:              uuid.New(),
		version:         1, // Start with version 1 for new aggregates
		tenantID:        tenantID,
		name:            name,
		listID:          listID,
		messageTemplate: messageTemplate,
		status:          BroadcastStatusDraft,
		rateLimit:       rateLimit,
		createdAt:       now,
		updatedAt:       now,
		events:          []shared.DomainEvent{},
	}

	broadcast.addEvent(NewBroadcastCreatedEvent(broadcast.id, tenantID, name, listID))

	return broadcast, nil
}

// ReconstructBroadcast reconstructs a broadcast from persistence
func ReconstructBroadcast(
	id uuid.UUID,
	version int, // Optimistic locking version
	tenantID string,
	name string,
	listID uuid.UUID,
	messageTemplate MessageTemplate,
	status BroadcastStatus,
	scheduledFor *time.Time,
	startedAt *time.Time,
	completedAt *time.Time,
	totalContacts, sentCount, failedCount, pendingCount int,
	rateLimit int,
	createdAt, updatedAt time.Time,
) *Broadcast {
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}

	return &Broadcast{
		id:              id,
		version:         version,
		tenantID:        tenantID,
		name:            name,
		listID:          listID,
		messageTemplate: messageTemplate,
		status:          status,
		scheduledFor:    scheduledFor,
		startedAt:       startedAt,
		completedAt:     completedAt,
		totalContacts:   totalContacts,
		sentCount:       sentCount,
		failedCount:     failedCount,
		pendingCount:    pendingCount,
		rateLimit:       rateLimit,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		events:          []shared.DomainEvent{},
	}
}

// Schedule schedules the broadcast for a specific time
func (b *Broadcast) Schedule(scheduledFor time.Time) error {
	if b.status != BroadcastStatusDraft {
		return errors.New("can only schedule broadcasts in draft status")
	}
	if scheduledFor.Before(time.Now()) {
		return errors.New("cannot schedule for past time")
	}

	b.scheduledFor = &scheduledFor
	b.status = BroadcastStatusScheduled
	b.updatedAt = time.Now()

	b.addEvent(NewBroadcastScheduledEvent(b.id, scheduledFor))

	return nil
}

// Start starts the broadcast execution
func (b *Broadcast) Start() error {
	if b.status != BroadcastStatusDraft && b.status != BroadcastStatusScheduled {
		return errors.New("can only start broadcasts in draft or scheduled status")
	}

	now := time.Now()
	b.status = BroadcastStatusRunning
	b.startedAt = &now
	b.updatedAt = now

	b.addEvent(NewBroadcastStartedEvent(b.id))

	return nil
}

// Complete marks the broadcast as completed
func (b *Broadcast) Complete() error {
	if b.status != BroadcastStatusRunning {
		return errors.New("can only complete broadcasts that are running")
	}

	now := time.Now()
	b.status = BroadcastStatusCompleted
	b.completedAt = &now
	b.updatedAt = now

	b.addEvent(NewBroadcastCompletedEvent(b.id, b.sentCount, b.failedCount))

	return nil
}

// Cancel cancels the broadcast
func (b *Broadcast) Cancel() error {
	if b.status == BroadcastStatusCompleted || b.status == BroadcastStatusCancelled {
		return errors.New("cannot cancel completed or already cancelled broadcasts")
	}

	now := time.Now()
	b.status = BroadcastStatusCancelled
	b.updatedAt = now

	b.addEvent(NewBroadcastCancelledEvent(b.id))

	return nil
}

// Fail marks the broadcast as failed
func (b *Broadcast) Fail(reason string) {
	now := time.Now()
	b.status = BroadcastStatusFailed
	b.updatedAt = now

	b.addEvent(NewBroadcastFailedEvent(b.id, reason))
}

// UpdateTotalContacts updates the total number of contacts
func (b *Broadcast) UpdateTotalContacts(total int) {
	b.totalContacts = total
	b.pendingCount = total
	b.updatedAt = time.Now()
}

// IncrementSent increments sent count
func (b *Broadcast) IncrementSent() {
	b.sentCount++
	b.pendingCount--
	b.updatedAt = time.Now()
}

// IncrementFailed increments failed count
func (b *Broadcast) IncrementFailed() {
	b.failedCount++
	b.pendingCount--
	b.updatedAt = time.Now()
}

// UpdateName updates the broadcast name
func (b *Broadcast) UpdateName(name string) error {
	if b.status != BroadcastStatusDraft {
		return errors.New("can only update name in draft status")
	}
	if name == "" {
		return errors.New("name cannot be empty")
	}

	b.name = name
	b.updatedAt = time.Now()
	return nil
}

// UpdateMessageTemplate updates the message template
func (b *Broadcast) UpdateMessageTemplate(template MessageTemplate) error {
	if b.status != BroadcastStatusDraft {
		return errors.New("can only update template in draft status")
	}
	if template.Type == "" {
		return errors.New("template type cannot be empty")
	}
	if template.Content == "" && template.TemplateID == nil {
		return errors.New("template must have either content or template_id")
	}

	b.messageTemplate = template
	b.updatedAt = time.Now()
	return nil
}

// UpdateRateLimit updates the rate limit
func (b *Broadcast) UpdateRateLimit(rateLimit int) error {
	if rateLimit < 0 {
		return errors.New("rateLimit cannot be negative")
	}

	b.rateLimit = rateLimit
	b.updatedAt = time.Now()
	return nil
}

// GetStats returns broadcast statistics
func (b *Broadcast) GetStats() BroadcastStats {
	progress := 0.0
	if b.totalContacts > 0 {
		progress = float64(b.sentCount+b.failedCount) / float64(b.totalContacts) * 100
	}

	return BroadcastStats{
		TotalContacts: b.totalContacts,
		SentCount:     b.sentCount,
		FailedCount:   b.failedCount,
		PendingCount:  b.pendingCount,
		Progress:      progress,
	}
}

// BroadcastStats represents broadcast statistics
type BroadcastStats struct {
	TotalContacts int     `json:"total_contacts"`
	SentCount     int     `json:"sent_count"`
	FailedCount   int     `json:"failed_count"`
	PendingCount  int     `json:"pending_count"`
	Progress      float64 `json:"progress"`
}

// IsReadyToStart checks if broadcast is ready to start
func (b *Broadcast) IsReadyToStart() bool {
	if b.status != BroadcastStatusScheduled {
		return false
	}
	if b.scheduledFor == nil {
		return false
	}
	return b.scheduledFor.Before(time.Now()) || b.scheduledFor.Equal(time.Now())
}

// Getters
func (b *Broadcast) ID() uuid.UUID                    { return b.id }
func (b *Broadcast) Version() int                     { return b.version }
func (b *Broadcast) TenantID() string                 { return b.tenantID }
func (b *Broadcast) Name() string                     { return b.name }
func (b *Broadcast) ListID() uuid.UUID                { return b.listID }
func (b *Broadcast) MessageTemplate() MessageTemplate { return b.messageTemplate }
func (b *Broadcast) Status() BroadcastStatus          { return b.status }
func (b *Broadcast) ScheduledFor() *time.Time         { return b.scheduledFor }
func (b *Broadcast) StartedAt() *time.Time            { return b.startedAt }
func (b *Broadcast) CompletedAt() *time.Time          { return b.completedAt }
func (b *Broadcast) TotalContacts() int               { return b.totalContacts }
func (b *Broadcast) SentCount() int                   { return b.sentCount }
func (b *Broadcast) FailedCount() int                 { return b.failedCount }
func (b *Broadcast) PendingCount() int                { return b.pendingCount }
func (b *Broadcast) RateLimit() int                   { return b.rateLimit }
func (b *Broadcast) CreatedAt() time.Time             { return b.createdAt }
func (b *Broadcast) UpdatedAt() time.Time             { return b.updatedAt }

func (b *Broadcast) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, b.events...)
}

func (b *Broadcast) ClearEvents() {
	b.events = []shared.DomainEvent{}
}

func (b *Broadcast) addEvent(event shared.DomainEvent) {
	b.events = append(b.events, event)
}

// Repository interface
type Repository interface {
	Save(broadcast *Broadcast) error
	FindByID(id uuid.UUID) (*Broadcast, error)
	FindByTenantID(tenantID string) ([]*Broadcast, error)
	FindScheduledReady() ([]*Broadcast, error)
	FindByStatus(status BroadcastStatus) ([]*Broadcast, error)
	Delete(id uuid.UUID) error
}

// Compile-time check that Broadcast implements AggregateRoot interface
var _ shared.AggregateRoot = (*Broadcast)(nil)
