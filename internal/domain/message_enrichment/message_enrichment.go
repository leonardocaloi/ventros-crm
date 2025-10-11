package message_enrichment

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MessageEnrichment representa o enriquecimento de uma mensagem com mídia
// É um agregado que contém o resultado do processamento de IA (transcrição, OCR, parsing, etc)
type MessageEnrichment struct {
	id             uuid.UUID
	messageID      uuid.UUID
	messageGroupID uuid.UUID
	contentType    EnrichmentContentType
	provider       EnrichmentProvider
	mediaURL       string
	status         EnrichmentStatus
	extractedText  *string
	metadata       map[string]interface{}
	processingTime *time.Duration
	error          *string
	context        *string            // Contexto de processamento (chat_message, profile_picture, etc)
	createdAt      time.Time
	processedAt    *time.Time
}

// NewMessageEnrichment cria um novo enriquecimento de mensagem
func NewMessageEnrichment(
	messageID uuid.UUID,
	messageGroupID uuid.UUID,
	contentType EnrichmentContentType,
	provider EnrichmentProvider,
	mediaURL string,
	context *string,
) (*MessageEnrichment, error) {
	// Validações
	if messageID == uuid.Nil {
		return nil, fmt.Errorf("message ID cannot be nil")
	}
	if messageGroupID == uuid.Nil {
		return nil, fmt.Errorf("message group ID cannot be nil")
	}
	if !contentType.IsValid() {
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}
	if !provider.IsValid() {
		return nil, fmt.Errorf("invalid provider: %s", provider)
	}
	if mediaURL == "" {
		return nil, fmt.Errorf("media URL cannot be empty")
	}

	return &MessageEnrichment{
		id:             uuid.New(),
		messageID:      messageID,
		messageGroupID: messageGroupID,
		contentType:    contentType,
		provider:       provider,
		mediaURL:       mediaURL,
		status:         StatusPending,
		metadata:       make(map[string]interface{}),
		context:        context,
		createdAt:      time.Now(),
	}, nil
}

// Getters
func (e *MessageEnrichment) ID() uuid.UUID                         { return e.id }
func (e *MessageEnrichment) MessageID() uuid.UUID                  { return e.messageID }
func (e *MessageEnrichment) MessageGroupID() uuid.UUID             { return e.messageGroupID }
func (e *MessageEnrichment) ContentType() EnrichmentContentType    { return e.contentType }
func (e *MessageEnrichment) Provider() EnrichmentProvider          { return e.provider }
func (e *MessageEnrichment) MediaURL() string                      { return e.mediaURL }
func (e *MessageEnrichment) Status() EnrichmentStatus              { return e.status }
func (e *MessageEnrichment) ExtractedText() *string                { return e.extractedText }
func (e *MessageEnrichment) Metadata() map[string]interface{}      { return e.metadata }
func (e *MessageEnrichment) ProcessingTime() *time.Duration        { return e.processingTime }
func (e *MessageEnrichment) Error() *string                        { return e.error }
func (e *MessageEnrichment) Context() *string                      { return e.context }
func (e *MessageEnrichment) CreatedAt() time.Time                  { return e.createdAt }
func (e *MessageEnrichment) ProcessedAt() *time.Time               { return e.processedAt }

// MarkAsProcessing marca o enriquecimento como em processamento
func (e *MessageEnrichment) MarkAsProcessing() error {
	if e.status != StatusPending {
		return fmt.Errorf("can only mark pending enrichments as processing, current status: %s", e.status)
	}

	e.status = StatusProcessing
	return nil
}

// MarkAsCompleted marca o enriquecimento como concluído com sucesso
func (e *MessageEnrichment) MarkAsCompleted(
	extractedText string,
	metadata map[string]interface{},
	processingTime time.Duration,
) error {
	if e.status != StatusProcessing {
		return fmt.Errorf("can only mark processing enrichments as completed, current status: %s", e.status)
	}

	e.status = StatusCompleted
	e.extractedText = &extractedText
	e.metadata = metadata
	e.processingTime = &processingTime
	now := time.Now()
	e.processedAt = &now

	return nil
}

// MarkAsFailed marca o enriquecimento como falho
func (e *MessageEnrichment) MarkAsFailed(errorMsg string) error {
	if e.status != StatusProcessing {
		return fmt.Errorf("can only mark processing enrichments as failed, current status: %s", e.status)
	}

	e.status = StatusFailed
	e.error = &errorMsg
	now := time.Now()
	e.processedAt = &now

	return nil
}

// IsPending retorna true se o enriquecimento está pendente
func (e *MessageEnrichment) IsPending() bool {
	return e.status == StatusPending
}

// IsProcessing retorna true se o enriquecimento está em processamento
func (e *MessageEnrichment) IsProcessing() bool {
	return e.status == StatusProcessing
}

// IsCompleted retorna true se o enriquecimento foi concluído
func (e *MessageEnrichment) IsCompleted() bool {
	return e.status == StatusCompleted
}

// IsFailed retorna true se o enriquecimento falhou
func (e *MessageEnrichment) IsFailed() bool {
	return e.status == StatusFailed
}

// IsFinal retorna true se o enriquecimento está em estado final
func (e *MessageEnrichment) IsFinal() bool {
	return e.status.IsFinal()
}

// Priority retorna a prioridade do enriquecimento para processamento
func (e *MessageEnrichment) Priority() uint8 {
	return e.contentType.Priority()
}

// AddMetadata adiciona metadados ao enriquecimento
func (e *MessageEnrichment) AddMetadata(key string, value interface{}) {
	if e.metadata == nil {
		e.metadata = make(map[string]interface{})
	}
	e.metadata[key] = value
}

// GetMetadata retorna um metadado específico
func (e *MessageEnrichment) GetMetadata(key string) (interface{}, bool) {
	if e.metadata == nil {
		return nil, false
	}
	val, exists := e.metadata[key]
	return val, exists
}

// Reconstitute reconstitui um MessageEnrichment a partir do banco de dados
// Este método é usado apenas pela camada de infraestrutura (repository)
func Reconstitute(
	id uuid.UUID,
	messageID uuid.UUID,
	messageGroupID uuid.UUID,
	contentType EnrichmentContentType,
	provider EnrichmentProvider,
	mediaURL string,
	status EnrichmentStatus,
	extractedText *string,
	metadata map[string]interface{},
	processingTime *time.Duration,
	error *string,
	context *string,
	createdAt time.Time,
	processedAt *time.Time,
) (*MessageEnrichment, error) {
	// Validações básicas
	if id == uuid.Nil {
		return nil, fmt.Errorf("id cannot be nil")
	}
	if messageID == uuid.Nil {
		return nil, fmt.Errorf("message ID cannot be nil")
	}
	if messageGroupID == uuid.Nil {
		return nil, fmt.Errorf("message group ID cannot be nil")
	}
	if !contentType.IsValid() {
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}
	if !provider.IsValid() {
		return nil, fmt.Errorf("invalid provider: %s", provider)
	}
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", status)
	}
	if mediaURL == "" {
		return nil, fmt.Errorf("media URL cannot be empty")
	}

	// Inicializa metadata se for nil
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &MessageEnrichment{
		id:             id,
		messageID:      messageID,
		messageGroupID: messageGroupID,
		contentType:    contentType,
		provider:       provider,
		mediaURL:       mediaURL,
		status:         status,
		extractedText:  extractedText,
		metadata:       metadata,
		processingTime: processingTime,
		error:          error,
		context:        context,
		createdAt:      createdAt,
		processedAt:    processedAt,
	}, nil
}
