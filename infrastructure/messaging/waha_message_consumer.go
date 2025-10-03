package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/application/message"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
)

// WAHAMessageConsumer consome eventos de mensagem do WAHA via RabbitMQ.
type WAHAMessageConsumer struct {
	adapter        *waha.MessageAdapter
	processUseCase *message.ProcessInboundMessageUseCase
	channelTypeID  int // ID do channel type "waha"
}

// NewWAHAMessageConsumer cria um novo consumer de mensagens WAHA.
func NewWAHAMessageConsumer(
	adapter *waha.MessageAdapter,
	processUseCase *message.ProcessInboundMessageUseCase,
	channelTypeID int,
) *WAHAMessageConsumer {
	return &WAHAMessageConsumer{
		adapter:        adapter,
		processUseCase: processUseCase,
		channelTypeID:  channelTypeID,
	}
}

// ProcessMessage processa uma mensagem do RabbitMQ.
// Implementa a interface Consumer.
func (c *WAHAMessageConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	// 1. Deserializa evento do WAHA
	var wahaEvent waha.WAHAMessageEvent
	if err := json.Unmarshal(delivery.Body, &wahaEvent); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	
	// Log para debug
	fmt.Printf("Processing WAHA event: id=%s, event=%s, session=%s, from=%s\n",
		wahaEvent.ID,
		wahaEvent.Event,
		wahaEvent.Session,
		wahaEvent.Payload.From,
	)
	
	// 2. Ignora mensagens enviadas por nós (fromMe = true)
	if wahaEvent.Payload.FromMe {
		fmt.Println("Ignoring outbound message (fromMe=true)")
		return nil
	}
	
	// 3. Ignora mensagens de grupos (por enquanto)
	if wahaEvent.Payload.Data.Info.IsGroup {
		fmt.Println("Ignoring group message")
		return nil
	}
	
	// 4. Usa adapter para converter estrutura externa → domínio
	contentType, err := c.adapter.ToContentType(wahaEvent)
	if err != nil {
		return fmt.Errorf("unsupported content type: %w", err)
	}
	
	phone := c.adapter.ExtractContactPhone(wahaEvent)
	text := c.adapter.ExtractText(wahaEvent)
	mediaURL := c.adapter.ExtractMediaURL(wahaEvent)
	mimetype := c.adapter.ExtractMimeType(wahaEvent)
	tracking := c.adapter.ExtractTrackingData(wahaEvent)
	
	// 5. Converte tracking data para map[string]interface{}
	trackingInterface := make(map[string]interface{})
	for k, v := range tracking {
		trackingInterface[k] = v
	}
	
	// 6. Monta command do use case
	cmd := message.ProcessInboundMessageCommand{
		ExternalID:    wahaEvent.Payload.ID,
		Session:       wahaEvent.Session,
		ContactPhone:  phone,
		ContactName:   wahaEvent.Payload.Data.Info.PushName,
		ChannelTypeID: c.channelTypeID, // ID do canal WAHA
		ContentType:   string(contentType),
		Text:          text,
		MediaURL:      derefString(mediaURL),
		MediaMimetype: derefString(mimetype),
		TrackingData:  trackingInterface,
		ReceivedAt:    time.Unix(wahaEvent.Timestamp/1000, 0),
		
		// Metadata adicional
		Metadata: map[string]interface{}{
			"waha_event_id": wahaEvent.ID,
			"waha_session":  wahaEvent.Session,
			"customer":      wahaEvent.Metadata["customer"],
			"source":        wahaEvent.Payload.Source,
			"is_from_ad":    c.adapter.IsFromAd(wahaEvent),
		},
	}
	
	// 6. Executa use case (pode iniciar workflow Temporal)
	if err := c.processUseCase.Execute(ctx, cmd); err != nil {
		return fmt.Errorf("failed to process message: %w", err)
	}
	
	fmt.Printf("Message processed successfully: id=%s\n", wahaEvent.Payload.ID)
	return nil
}

// Start inicia o consumer.
func (c *WAHAMessageConsumer) Start(ctx context.Context, rabbitConn *RabbitMQConnection) error {
	queueName := "waha.events.message"
	consumerTag := fmt.Sprintf("ventros-crm-message-consumer-%s", uuid.New().String()[:8])
	
	fmt.Printf("Starting WAHA message consumer: queue=%s, tag=%s\n", queueName, consumerTag)
	
	return rabbitConn.StartConsumer(ctx, queueName, consumerTag, c, 10)
}

// derefString dereferences string pointer or returns empty string
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
