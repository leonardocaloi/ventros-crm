package messaging

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	messageports "github.com/ventros/crm/internal/application/message"
)

// MessageSenderFactory implementa a factory para criação de message senders
// Seguindo Factory Pattern e Dependency Inversion Principle (DIP)
type MessageSenderFactory struct {
	senders map[string]messageports.ChannelMessageSender
	configs map[string]ChannelConfig
	mutex   sync.RWMutex
}

// ChannelConfig representa a configuração de um canal
type ChannelConfig struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
}

// NewMessageSenderFactory cria uma nova factory
func NewMessageSenderFactory() *MessageSenderFactory {
	return &MessageSenderFactory{
		senders: make(map[string]messageports.ChannelMessageSender),
		configs: make(map[string]ChannelConfig),
	}
}

// RegisterSender registra um sender para um tipo de canal
// Seguindo Open/Closed Principle (OCP) - permite extensão sem modificação
func (f *MessageSenderFactory) RegisterSender(channelType string, sender messageports.ChannelMessageSender, config ChannelConfig) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.senders[channelType] = sender
	f.configs[channelType] = config
}

// CreateSender cria um sender para o tipo de canal especificado
func (f *MessageSenderFactory) CreateSender(channelType string) (messageports.ChannelMessageSender, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	sender, exists := f.senders[channelType]
	if !exists {
		return nil, fmt.Errorf("no sender registered for channel type: %s", channelType)
	}

	config, configExists := f.configs[channelType]
	if configExists && !config.Enabled {
		return nil, fmt.Errorf("sender for channel type %s is disabled", channelType)
	}

	return sender, nil
}

// GetAvailableSenders retorna lista de senders disponíveis
func (f *MessageSenderFactory) GetAvailableSenders() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	var senders []string
	for channelType, config := range f.configs {
		if config.Enabled {
			senders = append(senders, channelType)
		}
	}

	return senders
}

// GetSenderConfig retorna a configuração de um sender
func (f *MessageSenderFactory) GetSenderConfig(channelType string) (*ChannelConfig, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	config, exists := f.configs[channelType]
	if !exists {
		return nil, fmt.Errorf("no config found for channel type: %s", channelType)
	}

	return &config, nil
}

// UpdateSenderConfig atualiza a configuração de um sender
func (f *MessageSenderFactory) UpdateSenderConfig(channelType string, config ChannelConfig) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, exists := f.senders[channelType]; !exists {
		return fmt.Errorf("no sender registered for channel type: %s", channelType)
	}

	f.configs[channelType] = config
	return nil
}

// EnableSender habilita um sender
func (f *MessageSenderFactory) EnableSender(channelType string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	config, exists := f.configs[channelType]
	if !exists {
		return fmt.Errorf("no config found for channel type: %s", channelType)
	}

	config.Enabled = true
	f.configs[channelType] = config
	return nil
}

// DisableSender desabilita um sender
func (f *MessageSenderFactory) DisableSender(channelType string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	config, exists := f.configs[channelType]
	if !exists {
		return fmt.Errorf("no config found for channel type: %s", channelType)
	}

	config.Enabled = false
	f.configs[channelType] = config
	return nil
}

// GetSenderCapabilities retorna as capacidades de um sender
func (f *MessageSenderFactory) GetSenderCapabilities(channelType string) (*messageports.ChannelCapabilities, error) {
	sender, err := f.CreateSender(channelType)
	if err != nil {
		return nil, err
	}

	// Para obter capacidades, precisamos de um channelID dummy
	// TODO: Melhorar isso para não precisar de channelID
	return sender.GetChannelCapabilities(uuid.Nil)
}

// ValidateChannelType valida se um tipo de canal é suportado
func (f *MessageSenderFactory) ValidateChannelType(channelType string) error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if _, exists := f.senders[channelType]; !exists {
		return fmt.Errorf("unsupported channel type: %s", channelType)
	}

	return nil
}

// GetSenderStats retorna estatísticas dos senders
func (f *MessageSenderFactory) GetSenderStats() map[string]interface{} {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	stats := make(map[string]interface{})

	for channelType, config := range f.configs {
		stats[channelType] = map[string]interface{}{
			"enabled":     config.Enabled,
			"name":        config.Name,
			"description": config.Description,
		}
	}

	return stats
}

// InitializeDefaultSenders inicializa os senders padrão
func (f *MessageSenderFactory) InitializeDefaultSenders() error {
	// Registrar WAHA sender
	wahaConfig := ChannelConfig{
		Type:        "waha",
		Name:        "WAHA WhatsApp",
		Description: "WhatsApp HTTP API Multi-device",
		Config: map[string]interface{}{
			"base_url":     "http://localhost:3000",
			"session_name": "default",
		},
		Enabled: true,
	}

	wahaSender := NewWAHAMessageSender(
		wahaConfig.Config["base_url"].(string),
		"", // API key será configurado via environment
		wahaConfig.Config["session_name"].(string),
	)

	f.RegisterSender("waha", wahaSender, wahaConfig)

	// TODO: Registrar outros senders (Telegram, Email, SMS, etc.)

	return nil
}
