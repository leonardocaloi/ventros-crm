package shared

import "time"

// DomainEvent é a interface base para todos os eventos de domínio.
// Todos os eventos devem implementar esta interface para garantir
// rastreabilidade e ordenação temporal.
type DomainEvent interface {
	// EventName retorna o nome do evento no formato "resource.action".
	// Exemplos: "contact.created", "session.started", "message.delivered"
	EventName() string

	// OccurredAt retorna o timestamp de quando o evento ocorreu.
	// É usado para ordenação cronológica e auditoria.
	OccurredAt() time.Time
}

// BaseEvent é uma struct base que pode ser embarcada em eventos de domínio
// para fornecer implementação padrão dos métodos comuns.
type BaseEvent struct {
	eventName  string
	occurredAt time.Time
}

// NewBaseEvent cria um novo BaseEvent.
func NewBaseEvent(eventName string, occurredAt time.Time) BaseEvent {
	return BaseEvent{
		eventName:  eventName,
		occurredAt: occurredAt,
	}
}

func (e BaseEvent) EventName() string     { return e.eventName }
func (e BaseEvent) OccurredAt() time.Time { return e.occurredAt }
