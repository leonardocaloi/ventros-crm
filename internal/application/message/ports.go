package message

import (
	"context"
	"time"

	"github.com/google/uuid"
	domainMessage "github.com/ventros/crm/internal/domain/crm/message"
)

// MessageType define os tipos de mensagem suportados
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeDocument MessageType = "document"
	MessageTypeLocation MessageType = "location"
	MessageTypeContact  MessageType = "contact"
	MessageTypeTemplate MessageType = "template"
)

// MessagePriority define a prioridade da mensagem
type MessagePriority string

const (
	PriorityLow    MessagePriority = "low"
	PriorityNormal MessagePriority = "normal"
	PriorityHigh   MessagePriority = "high"
	PriorityUrgent MessagePriority = "urgent"
)

// OutboundMessage representa uma mensagem a ser enviada
type OutboundMessage struct {
	ID           uuid.UUID              `json:"id"`
	ChannelID    uuid.UUID              `json:"channel_id"`
	ContactID    uuid.UUID              `json:"contact_id"`
	SessionID    *uuid.UUID             `json:"session_id,omitempty"`
	AgentID      uuid.UUID              `json:"agent_id"` // OBRIGATÓRIO: ID do agente (humano ou sistema)
	Source       domainMessage.Source   `json:"source"`   // OBRIGATÓRIO: Origem da mensagem (manual, broadcast, etc)
	Type         MessageType            `json:"type"`
	Content      string                 `json:"content"`
	MediaURL     *string                `json:"media_url,omitempty"`
	MediaType    *string                `json:"media_type,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Priority     MessagePriority        `json:"priority"`
	ScheduledAt  *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	ReplyToID    *uuid.UUID             `json:"reply_to_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// SendResult representa o resultado do envio de uma mensagem
type SendResult struct {
	MessageID   uuid.UUID              `json:"message_id"`
	ExternalID  *string                `json:"external_id,omitempty"`
	Status      string                 `json:"status"`
	DeliveredAt *time.Time             `json:"delivered_at,omitempty"`
	Error       *string                `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MessageSender interface define o contrato para envio de mensagens
// Port da Application Layer - implementado pela Infrastructure Layer
type MessageSender interface {
	SendMessage(ctx context.Context, message *OutboundMessage) (*SendResult, error)
	SendBulkMessages(ctx context.Context, messages []*OutboundMessage) ([]*SendResult, error)
	GetSupportedTypes() []MessageType
	ValidateMessage(message *OutboundMessage) error
}

// ChannelMessageSender interface específica para cada tipo de canal
type ChannelMessageSender interface {
	MessageSender
	GetChannelType() string
	IsChannelSupported(channelID uuid.UUID) bool
	GetChannelCapabilities(channelID uuid.UUID) (*ChannelCapabilities, error)
}

// ChannelCapabilities define as capacidades de um canal específico
type ChannelCapabilities struct {
	SupportedTypes   []MessageType   `json:"supported_types"`
	MaxContentLength int             `json:"max_content_length"`
	MaxMediaSize     int64           `json:"max_media_size"`
	SupportedFormats []string        `json:"supported_formats"`
	Features         map[string]bool `json:"features"`
	RateLimits       map[string]int  `json:"rate_limits"`
}

// MessageSenderFactory interface para criação de senders
type MessageSenderFactory interface {
	CreateSender(channelType string) (ChannelMessageSender, error)
	GetAvailableSenders() []string
}

// MessageQueue interface para enfileiramento de mensagens
type MessageQueue interface {
	Enqueue(ctx context.Context, message *OutboundMessage) error
	Dequeue(ctx context.Context) (*OutboundMessage, error)
	GetQueueSize(ctx context.Context) (int, error)
	GetPendingMessages(ctx context.Context, limit int) ([]*OutboundMessage, error)
}

// MessageRepository interface para persistência
type MessageRepository interface {
	SaveOutboundMessage(ctx context.Context, message *OutboundMessage) error
	UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, result *SendResult) error
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*OutboundMessage, error)
	GetMessagesBySession(ctx context.Context, sessionID uuid.UUID) ([]*OutboundMessage, error)
	GetPendingMessages(ctx context.Context, limit int) ([]*OutboundMessage, error)
}

// MessageValidator interface para validação de mensagens
type MessageValidator interface {
	ValidateContent(messageType MessageType, content string) error
	ValidateMedia(mediaURL string, mediaType string) error
	ValidateTemplate(templateData map[string]interface{}) error
}

// MessageScheduler interface para agendamento de mensagens
type MessageScheduler interface {
	ScheduleMessage(ctx context.Context, message *OutboundMessage) error
	GetScheduledMessages(ctx context.Context, before time.Time) ([]*OutboundMessage, error)
	CancelScheduledMessage(ctx context.Context, messageID uuid.UUID) error
}

// MessageMetrics interface para métricas de mensagens
type MessageMetrics interface {
	RecordMessageSent(channelType string, messageType MessageType)
	RecordMessageFailed(channelType string, messageType MessageType, reason string)
	RecordDeliveryTime(channelType string, duration time.Duration)
	GetMessageStats(ctx context.Context, channelType string) (*MessageStats, error)
}

// MessageStats representa estatísticas de mensagens
type MessageStats struct {
	TotalSent      int64                 `json:"total_sent"`
	TotalFailed    int64                 `json:"total_failed"`
	SuccessRate    float64               `json:"success_rate"`
	AverageLatency time.Duration         `json:"average_latency"`
	TypeBreakdown  map[MessageType]int64 `json:"type_breakdown"`
}
