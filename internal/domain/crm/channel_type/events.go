package channel_type

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type ChannelTypeCreatedEvent struct {
	ChannelTypeID int
	Name          string
	Provider      string
	CreatedAt     time.Time
}

func (e ChannelTypeCreatedEvent) EventName() string     { return "channel_type.created" }
func (e ChannelTypeCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type ChannelTypeActivatedEvent struct {
	ChannelTypeID int
	ActivatedAt   time.Time
}

func (e ChannelTypeActivatedEvent) EventName() string     { return "channel_type.activated" }
func (e ChannelTypeActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

type ChannelTypeDeactivatedEvent struct {
	ChannelTypeID int
	DeactivatedAt time.Time
}

func (e ChannelTypeDeactivatedEvent) EventName() string     { return "channel_type.deactivated" }
func (e ChannelTypeDeactivatedEvent) OccurredAt() time.Time { return e.DeactivatedAt }
