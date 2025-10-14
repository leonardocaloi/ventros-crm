package activation

import (
	"fmt"

	"github.com/ventros/crm/internal/domain/crm/channel"
	"go.uber.org/zap"
)

// StrategyFactory cria strategies de ativação baseado no tipo de canal
type StrategyFactory struct {
	logger *zap.Logger
}

// NewStrategyFactory cria uma nova factory
func NewStrategyFactory(logger *zap.Logger) *StrategyFactory {
	return &StrategyFactory{
		logger: logger,
	}
}

// GetStrategy retorna a strategy apropriada para o tipo de canal
// Pattern: Factory Method
func (f *StrategyFactory) GetStrategy(channelType channel.ChannelType) (ChannelActivationStrategy, error) {
	switch channelType {
	case channel.TypeWAHA:
		return NewWAHAActivationStrategy(f.logger), nil

	case channel.TypeWhatsAppBusiness:
		// WhatsApp Business também usa WAHA internamente
		return NewWAHAActivationStrategy(f.logger), nil

	case channel.TypeTelegram:
		// TODO: Implementar TelegramActivationStrategy
		return nil, fmt.Errorf("Telegram activation strategy not implemented yet")

	case channel.TypeWhatsApp:
		// TODO: Implementar WhatsAppCloudActivationStrategy
		return nil, fmt.Errorf("WhatsApp Cloud activation strategy not implemented yet")

	case channel.TypeTwilioSMS:
		// TODO: Implementar TwilioActivationStrategy
		return nil, fmt.Errorf("Twilio activation strategy not implemented yet")

	case channel.TypeMessenger:
		// TODO: Implementar MessengerActivationStrategy
		return nil, fmt.Errorf("Messenger activation strategy not implemented yet")

	case channel.TypeInstagram:
		// TODO: Implementar InstagramActivationStrategy
		return nil, fmt.Errorf("Instagram activation strategy not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported channel type for activation: %s", channelType)
	}
}

// IsSupported verifica se um tipo de canal tem strategy implementada
func (f *StrategyFactory) IsSupported(channelType channel.ChannelType) bool {
	switch channelType {
	case channel.TypeWAHA, channel.TypeWhatsAppBusiness:
		return true
	default:
		return false
	}
}

// GetSupportedTypes retorna lista de tipos de canal com strategy implementada
func (f *StrategyFactory) GetSupportedTypes() []channel.ChannelType {
	return []channel.ChannelType{
		channel.TypeWAHA,
		channel.TypeWhatsAppBusiness,
	}
}
