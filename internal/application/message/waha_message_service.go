package message

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/application/config"
	"github.com/caloi/ventros-crm/internal/domain/channel"
	"go.uber.org/zap"
)

// WAHAMessageService encapsula a lógica de processamento de mensagens WAHA.
// Remove responsabilidade excessiva dos handlers/consumers.
type WAHAMessageService struct {
	logger              *zap.Logger
	channelRepo         channel.Repository
	processMessageUC    *ProcessInboundMessageUseCase
	appConfig           *config.AppConfig
	messageAdapter      *waha.MessageAdapter
}

// NewWAHAMessageService cria um novo serviço de mensagens WAHA.
func NewWAHAMessageService(
	logger *zap.Logger,
	channelRepo channel.Repository,
	processMessageUC *ProcessInboundMessageUseCase,
	appConfig *config.AppConfig,
	messageAdapter *waha.MessageAdapter,
) *WAHAMessageService {
	return &WAHAMessageService{
		logger:           logger,
		channelRepo:      channelRepo,
		processMessageUC: processMessageUC,
		appConfig:        appConfig,
		messageAdapter:   messageAdapter,
	}
}

// ProcessWAHAMessage processa uma mensagem do WAHA (vinda de webhook ou RabbitMQ).
func (s *WAHAMessageService) ProcessWAHAMessage(ctx context.Context, event waha.WAHAMessageEvent) error {
	// 1. Validações iniciais
	if event.Payload.FromMe {
		s.logger.Debug("Ignoring outbound message", zap.String("message_id", event.Payload.ID))
		return nil
	}
	
	if event.Payload.Data.Info.IsGroup {
		s.logger.Debug("Ignoring group message", zap.String("message_id", event.Payload.ID))
		return nil
	}
	
	// 2. Buscar canal pelo ExternalID (WAHA session)
	ch, err := s.channelRepo.GetByExternalID(event.Session)
	if err != nil {
		return fmt.Errorf("channel not found for WAHA session '%s': %w", event.Session, err)
	}
	
	// 3. Validar se canal está ativo
	if !ch.IsActive() {
		return fmt.Errorf("channel %s is not active (status: %s)", ch.ID.String(), ch.Status)
	}
	
	// 4. Obter ChannelTypeID do AppConfig (remove hardcoded!)
	channelTypeID, err := s.appConfig.GetChannelTypeID(string(ch.Type))
	if err != nil {
		return fmt.Errorf("failed to get channel type ID for '%s': %w", ch.Type, err)
	}
	
	// 5. Extrair dados da mensagem usando adapter
	contentType, err := s.messageAdapter.ToContentType(event)
	if err != nil {
		return fmt.Errorf("unsupported content type: %w", err)
	}
	
	phone := s.messageAdapter.ExtractContactPhone(event)
	text := s.messageAdapter.ExtractText(event)
	mediaURL := s.messageAdapter.ExtractMediaURL(event)
	mimetype := s.messageAdapter.ExtractMimeType(event)
	tracking := s.messageAdapter.ExtractTrackingData(event)
	
	// 6. Converter tracking data para map[string]interface{}
	trackingInterface := make(map[string]interface{})
	for k, v := range tracking {
		trackingInterface[k] = v
	}
	
	// 7. Montar command
	cmd := ProcessInboundMessageCommand{
		MessageID:     event.Payload.ID,
		ContactPhone:  phone,
		ContactName:   event.Payload.Data.Info.PushName,
		// Dados do canal (OBRIGATÓRIO)
		ChannelID:     ch.ID,
		ProjectID:     ch.ProjectID,
		CustomerID:    ch.UserID,
		TenantID:      ch.TenantID,
		ChannelTypeID: channelTypeID, // ✅ Obtido do AppConfig
		ContentType:   string(contentType),
		Text:          text,
		MediaURL:      derefString(mediaURL),
		MediaMimetype: derefString(mimetype),
		TrackingData:  trackingInterface,
		ReceivedAt:    time.Unix(event.Payload.Timestamp/1000, 0), // Converter milliseconds para time.Time
		Metadata: map[string]interface{}{
			"waha_event_id": event.ID,
			"waha_session":  event.Session, // ExternalID do canal WAHA
			"channel_id":    ch.ID.String(),
			"channel_name":  ch.Name,
			"source":        event.Payload.Source,
			"is_from_ad":    s.messageAdapter.IsFromAd(event),
		},
	}
	
	// 8. Executar use case
	if err := s.processMessageUC.Execute(ctx, cmd); err != nil {
		return fmt.Errorf("failed to process message: %w", err)
	}
	
	// 9. Atualizar estatísticas do canal
	ch.IncrementMessagesReceived()
	if err := s.channelRepo.Update(ch); err != nil {
		s.logger.Warn("Failed to update channel statistics",
			zap.String("channel_id", ch.ID.String()),
			zap.Error(err))
	}
	
	s.logger.Info("WAHA message processed successfully",
		zap.String("message_id", event.Payload.ID),
		zap.String("from", phone),
		zap.String("waha_session", event.Session),
		zap.String("channel_id", ch.ID.String()),
		zap.String("project_id", ch.ProjectID.String()))
	
	return nil
}

// derefString dereferences string pointer or returns empty string
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
