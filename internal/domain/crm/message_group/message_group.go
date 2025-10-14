package message_group

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

// MessageGroup representa um grupo de mensagens agrupadas pelo debouncer
// Agregado: Agrupa mensagens sequenciais de um contato no mesmo canal
type MessageGroup struct {
	id          uuid.UUID
	contactID   uuid.UUID
	channelID   uuid.UUID
	sessionID   uuid.UUID
	tenantID    string
	messageIDs  []uuid.UUID // Mensagens no grupo (ordenadas)
	status      GroupStatus
	startedAt   time.Time
	completedAt *time.Time
	expiresAt   time.Time // Quando o debouncer expira
	events      []shared.DomainEvent
}

// GroupStatus representa o status do grupo de mensagens
type GroupStatus string

const (
	GroupStatusPending    GroupStatus = "pending"    // Aguardando mais mensagens
	GroupStatusProcessing GroupStatus = "processing" // Processando enriquecimentos
	GroupStatusCompleted  GroupStatus = "completed"  // Concluído e enviado para AI
	GroupStatusExpired    GroupStatus = "expired"    // Expirou sem completar
)

// NewMessageGroup cria um novo grupo de mensagens
func NewMessageGroup(
	contactID uuid.UUID,
	channelID uuid.UUID,
	sessionID uuid.UUID,
	tenantID string,
	firstMessageID uuid.UUID,
	debounceTimeout time.Duration,
) (*MessageGroup, error) {
	if contactID == uuid.Nil {
		return nil, errors.New("contact_id is required")
	}
	if channelID == uuid.Nil {
		return nil, errors.New("channel_id is required")
	}
	if debounceTimeout <= 0 {
		debounceTimeout = 15 * time.Second // Default 15s
	}

	now := time.Now()
	group := &MessageGroup{
		id:         uuid.New(),
		contactID:  contactID,
		channelID:  channelID,
		sessionID:  sessionID,
		tenantID:   tenantID,
		messageIDs: []uuid.UUID{firstMessageID},
		status:     GroupStatusPending,
		startedAt:  now,
		expiresAt:  now.Add(debounceTimeout),
		events:     []shared.DomainEvent{},
	}

	group.addEvent(NewMessageGroupCreatedEvent(group.id, contactID, channelID))

	return group, nil
}

// AddMessage adiciona mensagem ao grupo (reinicia o timer do debouncer)
func (g *MessageGroup) AddMessage(messageID uuid.UUID, debounceTimeout time.Duration) error {
	if g.status != GroupStatusPending {
		return errors.New("cannot add message to non-pending group")
	}

	// Verificar se mensagem já existe no grupo
	for _, id := range g.messageIDs {
		if id == messageID {
			return errors.New("message already in group")
		}
	}

	// Adicionar mensagem
	g.messageIDs = append(g.messageIDs, messageID)

	// Estender timeout do debouncer (reiniciar timer)
	g.expiresAt = time.Now().Add(debounceTimeout)

	g.addEvent(NewMessageAddedToGroupEvent(g.id, messageID))

	return nil
}

// MarkAsProcessing marca grupo como processando enriquecimentos
func (g *MessageGroup) MarkAsProcessing() error {
	if g.status != GroupStatusPending {
		return errors.New("only pending groups can be marked as processing")
	}

	g.status = GroupStatusProcessing

	g.addEvent(NewMessageGroupProcessingEvent(g.id, len(g.messageIDs)))

	return nil
}

// MarkAsCompleted marca grupo como concluído
func (g *MessageGroup) MarkAsCompleted() error {
	if g.status != GroupStatusProcessing {
		return errors.New("only processing groups can be marked as completed")
	}

	now := time.Now()
	g.status = GroupStatusCompleted
	g.completedAt = &now

	g.addEvent(NewMessageGroupCompletedEvent(g.id, len(g.messageIDs)))

	return nil
}

// MarkAsExpired marca grupo como expirado
func (g *MessageGroup) MarkAsExpired() {
	g.status = GroupStatusExpired
	now := time.Now()
	g.completedAt = &now

	g.addEvent(NewMessageGroupExpiredEvent(g.id))
}

// IsExpired verifica se o debouncer expirou
func (g *MessageGroup) IsExpired() bool {
	return time.Now().After(g.expiresAt)
}

// ShouldProcess verifica se deve processar o grupo
func (g *MessageGroup) ShouldProcess() bool {
	return g.status == GroupStatusPending && g.IsExpired()
}

// Getters
func (g *MessageGroup) ID() uuid.UUID           { return g.id }
func (g *MessageGroup) ContactID() uuid.UUID    { return g.contactID }
func (g *MessageGroup) ChannelID() uuid.UUID    { return g.channelID }
func (g *MessageGroup) SessionID() uuid.UUID    { return g.sessionID }
func (g *MessageGroup) TenantID() string        { return g.tenantID }
func (g *MessageGroup) MessageIDs() []uuid.UUID { return append([]uuid.UUID{}, g.messageIDs...) }
func (g *MessageGroup) MessageCount() int       { return len(g.messageIDs) }
func (g *MessageGroup) Status() GroupStatus     { return g.status }
func (g *MessageGroup) StartedAt() time.Time    { return g.startedAt }
func (g *MessageGroup) CompletedAt() *time.Time { return g.completedAt }
func (g *MessageGroup) ExpiresAt() time.Time    { return g.expiresAt }

// Domain Events
func (g *MessageGroup) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, g.events...)
}

func (g *MessageGroup) ClearEvents() {
	g.events = []shared.DomainEvent{}
}

func (g *MessageGroup) addEvent(event shared.DomainEvent) {
	g.events = append(g.events, event)
}

// ReconstructMessageGroup reconstrói grupo a partir do banco
func ReconstructMessageGroup(
	id uuid.UUID,
	contactID uuid.UUID,
	channelID uuid.UUID,
	sessionID uuid.UUID,
	tenantID string,
	messageIDs []uuid.UUID,
	status GroupStatus,
	startedAt time.Time,
	completedAt *time.Time,
	expiresAt time.Time,
) *MessageGroup {
	return &MessageGroup{
		id:          id,
		contactID:   contactID,
		channelID:   channelID,
		sessionID:   sessionID,
		tenantID:    tenantID,
		messageIDs:  messageIDs,
		status:      status,
		startedAt:   startedAt,
		completedAt: completedAt,
		expiresAt:   expiresAt,
		events:      []shared.DomainEvent{},
	}
}
