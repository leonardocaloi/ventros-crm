package shared

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent é a interface base para todos os eventos de domínio.
// Todos os eventos devem implementar esta interface para garantir
// rastreabilidade, idempotência e ordenação temporal.
type DomainEvent interface {
	// EventName retorna o nome do evento no formato "resource.action".
	// Exemplos: "contact.created", "session.started", "message.delivered"
	EventName() string

	// EventID retorna o identificador único do evento.
	// Usado para idempotência (prevenir processamento duplicado) e rastreabilidade.
	EventID() uuid.UUID

	// EventVersion retorna a versão do schema do evento.
	// Usado para schema evolution e compatibilidade entre versões.
	// Formato: "v1", "v2", etc.
	EventVersion() string

	// OccurredAt retorna o timestamp de quando o evento ocorreu.
	// É usado para ordenação cronológica e auditoria.
	OccurredAt() time.Time
}

// BaseEvent é uma struct base que pode ser embarcada em eventos de domínio
// para fornecer implementação padrão dos métodos comuns.
type BaseEvent struct {
	eventID      uuid.UUID
	eventName    string
	eventVersion string
	occurredAt   time.Time
}

// NewBaseEvent cria um novo BaseEvent com ID único gerado automaticamente.
func NewBaseEvent(eventName string, occurredAt time.Time) BaseEvent {
	return BaseEvent{
		eventID:      uuid.New(), // Gera ID único automaticamente
		eventName:    eventName,
		eventVersion: "v1", // Versão padrão v1
		occurredAt:   occurredAt,
	}
}

// NewBaseEventWithVersion cria um novo BaseEvent com versão específica.
// Use este construtor quando precisar criar eventos com versões diferentes de v1.
func NewBaseEventWithVersion(eventName string, version string, occurredAt time.Time) BaseEvent {
	return BaseEvent{
		eventID:      uuid.New(),
		eventName:    eventName,
		eventVersion: version,
		occurredAt:   occurredAt,
	}
}

func (e BaseEvent) EventID() uuid.UUID    { return e.eventID }
func (e BaseEvent) EventName() string     { return e.eventName }
func (e BaseEvent) EventVersion() string  { return e.eventVersion }
func (e BaseEvent) OccurredAt() time.Time { return e.occurredAt }
