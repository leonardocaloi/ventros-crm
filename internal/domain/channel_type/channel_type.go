package channel_type

import (
	"errors"
	"time"
)

// ErrChannelTypeNotFound is returned when a channel type is not found
var ErrChannelTypeNotFound = errors.New("channel type not found")

// Channel type IDs (constantes para type safety)
const (
	WAHA      = 1 // WAHA - WhatsApp HTTP API (Multi-device)
	WhatsApp  = 2 // WhatsApp Business API Official
	DirectIG  = 3 // Instagram Direct Messages
	Messenger = 4 // Facebook Messenger
	Telegram  = 5 // Telegram Bot API
)

// Names mapeamento de IDs para nomes
var Names = map[int]string{
	WAHA:      "waha",
	WhatsApp:  "whatsapp",
	DirectIG:  "direct_ig",
	Messenger: "messenger",
	Telegram:  "telegram",
}

// GetName retorna o nome de um channel type por ID
func GetName(id int) string {
	if name, ok := Names[id]; ok {
		return name
	}
	return "unknown"
}

// ChannelType é o Aggregate Root para tipos de canal de comunicação.
// Ex: WhatsApp, Instagram, Telegram, Messenger, Email, SMS.
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

// NewChannelType cria um novo tipo de canal.
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

// ReconstructChannelType reconstrói um ChannelType a partir de dados persistidos.
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

// Activate ativa o canal.
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

// Deactivate desativa o canal.
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

// UpdateConfiguration atualiza a configuração do canal.
func (ct *ChannelType) UpdateConfiguration(config map[string]interface{}) {
	ct.configuration = config
	ct.updatedAt = time.Now()
}

// GetConfiguration retorna uma configuração específica.
func (ct *ChannelType) GetConfiguration(key string) (interface{}, bool) {
	val, ok := ct.configuration[key]
	return val, ok
}

// UpdateDescription atualiza a descrição do canal.
func (ct *ChannelType) UpdateDescription(description string) {
	ct.description = description
	ct.updatedAt = time.Now()
}

// IsActive verifica se o canal está ativo.
func (ct *ChannelType) IsActive() bool {
	return ct.active
}

// IsMeta verifica se é um canal da Meta (Facebook/Instagram/WhatsApp).
func (ct *ChannelType) IsMeta() bool {
	return ct.provider == "meta"
}

// Getters
func (ct *ChannelType) ID() int             { return ct.id }
func (ct *ChannelType) Name() string        { return ct.name }
func (ct *ChannelType) Description() string { return ct.description }
func (ct *ChannelType) Provider() string    { return ct.provider }
func (ct *ChannelType) Configuration() map[string]interface{} {
	// Return copy
	copy := make(map[string]interface{})
	for k, v := range ct.configuration {
		copy[k] = v
	}
	return copy
}
func (ct *ChannelType) CreatedAt() time.Time { return ct.createdAt }
func (ct *ChannelType) UpdatedAt() time.Time { return ct.updatedAt }

// Domain Events
func (ct *ChannelType) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, ct.events...)
}

func (ct *ChannelType) ClearEvents() {
	ct.events = []DomainEvent{}
}

func (ct *ChannelType) addEvent(event DomainEvent) {
	ct.events = append(ct.events, event)
}
