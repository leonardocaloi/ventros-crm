package contact_event

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ContactEvent é o Aggregate Root para eventos de contato em tempo real.
// Esses eventos são visíveis para clientes e podem ser transmitidos via streaming.
type ContactEvent struct {
	// Identidade
	id        uuid.UUID
	contactID uuid.UUID
	sessionID *uuid.UUID
	tenantID  string

	// Detalhes do evento
	eventType string
	category  Category
	priority  Priority

	// Conteúdo
	title       *string
	description *string
	payload     map[string]interface{}
	metadata    map[string]interface{}

	// Origem e rastreamento
	source            Source
	triggeredBy       *uuid.UUID // Agent ID
	integrationSource *string

	// Entrega em tempo real
	isRealtime  bool
	delivered   bool
	deliveredAt *time.Time
	read        bool
	readAt      *time.Time

	// Visibilidade
	visibleToClient bool
	visibleToAgent  bool
	expiresAt       *time.Time

	// Timestamps
	occurredAt time.Time
	createdAt  time.Time
}

// NewContactEvent cria um novo evento de contato.
func NewContactEvent(
	contactID uuid.UUID,
	tenantID string,
	eventType string,
	category Category,
	priority Priority,
	source Source,
) (*ContactEvent, error) {
	// Validações
	if contactID == uuid.Nil {
		return nil, errors.New("contactID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if eventType == "" {
		return nil, errors.New("eventType cannot be empty")
	}
	if !category.IsValid() {
		return nil, errors.New("invalid category")
	}
	if !priority.IsValid() {
		return nil, errors.New("invalid priority")
	}
	if !source.IsValid() {
		return nil, errors.New("invalid source")
	}

	now := time.Now()
	return &ContactEvent{
		id:              uuid.New(),
		contactID:       contactID,
		tenantID:        tenantID,
		eventType:       eventType,
		category:        category,
		priority:        priority,
		source:          source,
		payload:         make(map[string]interface{}),
		metadata:        make(map[string]interface{}),
		isRealtime:      true,
		delivered:       false,
		read:            false,
		visibleToClient: true,
		visibleToAgent:  true,
		occurredAt:      now,
		createdAt:       now,
	}, nil
}

// ReconstructContactEvent reconstrói um ContactEvent a partir de dados persistidos.
func ReconstructContactEvent(
	id uuid.UUID,
	contactID uuid.UUID,
	sessionID *uuid.UUID,
	tenantID string,
	eventType string,
	category Category,
	priority Priority,
	title *string,
	description *string,
	payload map[string]interface{},
	metadata map[string]interface{},
	source Source,
	triggeredBy *uuid.UUID,
	integrationSource *string,
	isRealtime bool,
	delivered bool,
	deliveredAt *time.Time,
	read bool,
	readAt *time.Time,
	visibleToClient bool,
	visibleToAgent bool,
	expiresAt *time.Time,
	occurredAt time.Time,
	createdAt time.Time,
) *ContactEvent {
	if payload == nil {
		payload = make(map[string]interface{})
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &ContactEvent{
		id:                id,
		contactID:         contactID,
		sessionID:         sessionID,
		tenantID:          tenantID,
		eventType:         eventType,
		category:          category,
		priority:          priority,
		title:             title,
		description:       description,
		payload:           payload,
		metadata:          metadata,
		source:            source,
		triggeredBy:       triggeredBy,
		integrationSource: integrationSource,
		isRealtime:        isRealtime,
		delivered:         delivered,
		deliveredAt:       deliveredAt,
		read:              read,
		readAt:            readAt,
		visibleToClient:   visibleToClient,
		visibleToAgent:    visibleToAgent,
		expiresAt:         expiresAt,
		occurredAt:        occurredAt,
		createdAt:         createdAt,
	}
}

// AttachToSession associa o evento a uma sessão.
func (e *ContactEvent) AttachToSession(sessionID uuid.UUID) error {
	if sessionID == uuid.Nil {
		return errors.New("sessionID cannot be nil")
	}
	e.sessionID = &sessionID
	return nil
}

// SetTitle define o título do evento.
func (e *ContactEvent) SetTitle(title string) {
	e.title = &title
}

// SetDescription define a descrição do evento.
func (e *ContactEvent) SetDescription(description string) {
	e.description = &description
}

// AddPayloadField adiciona um campo ao payload.
func (e *ContactEvent) AddPayloadField(key string, value interface{}) {
	e.payload[key] = value
}

// AddMetadataField adiciona um campo ao metadata.
func (e *ContactEvent) AddMetadataField(key string, value interface{}) {
	e.metadata[key] = value
}

// SetTriggeredBy define o agente que disparou o evento.
func (e *ContactEvent) SetTriggeredBy(agentID uuid.UUID) error {
	if agentID == uuid.Nil {
		return errors.New("agentID cannot be nil")
	}
	e.triggeredBy = &agentID
	return nil
}

// SetIntegrationSource define a fonte de integração.
func (e *ContactEvent) SetIntegrationSource(source string) {
	e.integrationSource = &source
}

// MarkAsDelivered marca o evento como entregue.
func (e *ContactEvent) MarkAsDelivered() {
	if !e.delivered {
		now := time.Now()
		e.delivered = true
		e.deliveredAt = &now
	}
}

// MarkAsRead marca o evento como lido.
func (e *ContactEvent) MarkAsRead() {
	if !e.read {
		now := time.Now()
		e.read = true
		e.readAt = &now
	}
}

// SetRealtimeDelivery define se o evento deve ser entregue em tempo real.
func (e *ContactEvent) SetRealtimeDelivery(enabled bool) {
	e.isRealtime = enabled
}

// SetVisibility define a visibilidade do evento.
func (e *ContactEvent) SetVisibility(visibleToClient, visibleToAgent bool) {
	e.visibleToClient = visibleToClient
	e.visibleToAgent = visibleToAgent
}

// SetExpiresAt define quando o evento expira.
func (e *ContactEvent) SetExpiresAt(expiresAt time.Time) error {
	if expiresAt.Before(time.Now()) {
		return errors.New("expiresAt cannot be in the past")
	}
	e.expiresAt = &expiresAt
	return nil
}

// IsExpired verifica se o evento expirou.
func (e *ContactEvent) IsExpired() bool {
	if e.expiresAt == nil {
		return false
	}
	return time.Now().After(*e.expiresAt)
}

// IsDelivered verifica se o evento foi entregue.
func (e *ContactEvent) IsDelivered() bool {
	return e.delivered
}

// IsRead verifica se o evento foi lido.
func (e *ContactEvent) IsRead() bool {
	return e.read
}

// ShouldBeDeliveredInRealtime verifica se deve ser entregue em tempo real.
func (e *ContactEvent) ShouldBeDeliveredInRealtime() bool {
	return e.isRealtime && !e.delivered && !e.IsExpired()
}

// IsVisibleToClient verifica se é visível para o cliente.
func (e *ContactEvent) IsVisibleToClient() bool {
	return e.visibleToClient && !e.IsExpired()
}

// IsVisibleToAgent verifica se é visível para o agente.
func (e *ContactEvent) IsVisibleToAgent() bool {
	return e.visibleToAgent && !e.IsExpired()
}

// HasSession verifica se o evento está associado a uma sessão.
func (e *ContactEvent) HasSession() bool {
	return e.sessionID != nil
}

// IsHighPriority verifica se é alta prioridade.
func (e *ContactEvent) IsHighPriority() bool {
	return e.priority == PriorityHigh || e.priority == PriorityUrgent
}

// IsSystemGenerated verifica se foi gerado pelo sistema.
func (e *ContactEvent) IsSystemGenerated() bool {
	return e.source == SourceSystem
}

// Getters

func (e *ContactEvent) ID() uuid.UUID              { return e.id }
func (e *ContactEvent) ContactID() uuid.UUID       { return e.contactID }
func (e *ContactEvent) SessionID() *uuid.UUID      { return e.sessionID }
func (e *ContactEvent) TenantID() string           { return e.tenantID }
func (e *ContactEvent) EventType() string          { return e.eventType }
func (e *ContactEvent) Category() Category         { return e.category }
func (e *ContactEvent) Priority() Priority         { return e.priority }
func (e *ContactEvent) Title() *string             { return e.title }
func (e *ContactEvent) Description() *string       { return e.description }
func (e *ContactEvent) Source() Source             { return e.source }
func (e *ContactEvent) TriggeredBy() *uuid.UUID    { return e.triggeredBy }
func (e *ContactEvent) IntegrationSource() *string { return e.integrationSource }
func (e *ContactEvent) DeliveredAt() *time.Time    { return e.deliveredAt }
func (e *ContactEvent) ReadAt() *time.Time         { return e.readAt }
func (e *ContactEvent) ExpiresAt() *time.Time      { return e.expiresAt }
func (e *ContactEvent) OccurredAt() time.Time      { return e.occurredAt }
func (e *ContactEvent) CreatedAt() time.Time       { return e.createdAt }

// Payload retorna uma cópia do payload.
func (e *ContactEvent) Payload() map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range e.payload {
		copy[k] = v
	}
	return copy
}

// Metadata retorna uma cópia do metadata.
func (e *ContactEvent) Metadata() map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range e.metadata {
		copy[k] = v
	}
	return copy
}
