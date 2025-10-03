package messaging

import "time"

// DomainEventWrapper envolve eventos de domínio específicos para usar com shared.DomainEvent
type DomainEventWrapper struct {
	eventName  string
	occurredAt time.Time
	data       interface{}
}

func (w *DomainEventWrapper) EventName() string {
	return w.eventName
}

func (w *DomainEventWrapper) OccurredAt() time.Time {
	return w.occurredAt
}

func (w *DomainEventWrapper) Data() interface{} {
	return w.data
}
