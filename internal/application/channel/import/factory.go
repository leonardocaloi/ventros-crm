package importpkg

import (
	"fmt"

	"github.com/ventros/crm/internal/domain/crm/channel"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

// StrategyFactory cria strategies de importação baseado no tipo de canal
// Pattern: Factory Pattern
// Permite adicionar novos tipos de canal sem modificar código existente
type StrategyFactory struct {
	strategies     map[string]HistoryImportStrategy
	temporalClient client.Client
	logger         *zap.Logger
}

// NewStrategyFactory cria uma nova factory de strategies
func NewStrategyFactory(temporalClient client.Client, logger *zap.Logger) *StrategyFactory {
	factory := &StrategyFactory{
		strategies:     make(map[string]HistoryImportStrategy),
		temporalClient: temporalClient,
		logger:         logger,
	}

	// Registrar strategies conhecidas
	factory.registerStrategies()

	return factory
}

// registerStrategies registra todas as strategies disponíveis
// Para adicionar novo tipo de canal, apenas adicione aqui
func (f *StrategyFactory) registerStrategies() {
	// WAHA Strategy
	f.strategies[string(channel.TypeWAHA)] = NewWAHAImportStrategy(f.temporalClient, f.logger)

	// Adicionar novas strategies aqui:
	// f.strategies[channel.TypeTelegram] = NewTelegramImportStrategy(...)
	// f.strategies[channel.TypeTwilio] = NewTwilioImportStrategy(...)

	f.logger.Info("Import strategies registered",
		zap.Int("count", len(f.strategies)))
}

// GetStrategy retorna a strategy apropriada para o tipo de canal
// Retorna erro se tipo não for suportado
func (f *StrategyFactory) GetStrategy(channelType string) (HistoryImportStrategy, error) {
	strategy, exists := f.strategies[channelType]
	if !exists {
		f.logger.Warn("Unsupported channel type for history import",
			zap.String("channel_type", channelType))
		return nil, fmt.Errorf("history import not supported for channel type: %s", channelType)
	}

	return strategy, nil
}

// RegisterStrategy permite registrar uma nova strategy em runtime
// Útil para testes ou plugins
func (f *StrategyFactory) RegisterStrategy(channelType string, strategy HistoryImportStrategy) {
	f.strategies[channelType] = strategy
	f.logger.Info("Custom import strategy registered",
		zap.String("channel_type", channelType))
}

// SupportedTypes retorna lista de tipos de canal suportados
func (f *StrategyFactory) SupportedTypes() []string {
	types := make([]string, 0, len(f.strategies))
	for channelType := range f.strategies {
		types = append(types, channelType)
	}
	return types
}
