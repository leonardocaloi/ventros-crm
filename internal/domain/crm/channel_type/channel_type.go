package channel_type

import (
	"errors"
	"time"
)

var ErrChannelTypeNotFound = errors.New("channel type not found")

const (
	WAHA      = 1
	WhatsApp  = 2
	DirectIG  = 3
	Messenger = 4
	Telegram  = 5
)

var Names = map[int]string{
	WAHA:      "waha",
	WhatsApp:  "whatsapp",
	DirectIG:  "direct_ig",
	Messenger: "messenger",
	Telegram:  "telegram",
}

func GetName(id int) string {
	if name, ok := Names[id]; ok {
		return name
	}
	return "unknown"
}

type ChannelType struct {
	id            int
	name          string
	description   string
	provider      string
	configuration map[string]interface{}
	active        bool
	createdAt     time.Time
	updatedAt     time.Time

	events []DomainEvent
}

func NewChannelType(
	id int,
	name string,
	provider string,
	description string,
) (*ChannelType, error) {
	if id <= 0 {
		return nil, errors.New("id must be positive")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if provider == "" {
		return nil, errors.New("provider cannot be empty")
	}

	now := time.Now()
	ct := &ChannelType{
		id:            id,
		name:          name,
		description:   description,
		provider:      provider,
		configuration: make(map[string]interface{}),
		active:        true,
		createdAt:     now,
		updatedAt:     now,
		events:        []DomainEvent{},
	}

	ct.addEvent(ChannelTypeCreatedEvent{
		ChannelTypeID: id,
		Name:          name,
		Provider:      provider,
		CreatedAt:     now,
	})

	return ct, nil
}

func ReconstructChannelType(
	id int,
	name string,
	description string,
	provider string,
	configuration map[string]interface{},
	active bool,
	createdAt time.Time,
	updatedAt time.Time,
) *ChannelType {
	if configuration == nil {
		configuration = make(map[string]interface{})
	}

	return &ChannelType{
		id:            id,
		name:          name,
		description:   description,
		provider:      provider,
		configuration: configuration,
		active:        active,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		events:        []DomainEvent{},
	}
}

func (ct *ChannelType) Activate() error {
	if ct.active {
		return errors.New("channel type is already active")
	}

	ct.active = true
	ct.updatedAt = time.Now()

	ct.addEvent(ChannelTypeActivatedEvent{
		ChannelTypeID: ct.id,
		ActivatedAt:   ct.updatedAt,
	})

	return nil
}

func (ct *ChannelType) Deactivate() error {
	if !ct.active {
		return errors.New("channel type is already inactive")
	}

	ct.active = false
	ct.updatedAt = time.Now()

	ct.addEvent(ChannelTypeDeactivatedEvent{
		ChannelTypeID: ct.id,
		DeactivatedAt: ct.updatedAt,
	})

	return nil
}

func (ct *ChannelType) UpdateConfiguration(config map[string]interface{}) {
	ct.configuration = config
	ct.updatedAt = time.Now()
}

func (ct *ChannelType) GetConfiguration(key string) (interface{}, bool) {
	val, ok := ct.configuration[key]
	return val, ok
}

func (ct *ChannelType) UpdateDescription(description string) {
	ct.description = description
	ct.updatedAt = time.Now()
}

func (ct *ChannelType) IsActive() bool {
	return ct.active
}

func (ct *ChannelType) IsMeta() bool {
	return ct.provider == "meta"
}

func (ct *ChannelType) ID() int             { return ct.id }
func (ct *ChannelType) Name() string        { return ct.name }
func (ct *ChannelType) Description() string { return ct.description }
func (ct *ChannelType) Provider() string    { return ct.provider }
func (ct *ChannelType) Configuration() map[string]interface{} {

	copy := make(map[string]interface{})
	for k, v := range ct.configuration {
		copy[k] = v
	}
	return copy
}
func (ct *ChannelType) CreatedAt() time.Time { return ct.createdAt }
func (ct *ChannelType) UpdatedAt() time.Time { return ct.updatedAt }

func (ct *ChannelType) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, ct.events...)
}

func (ct *ChannelType) ClearEvents() {
	ct.events = []DomainEvent{}
}

func (ct *ChannelType) addEvent(event DomainEvent) {
	ct.events = append(ct.events, event)
}
