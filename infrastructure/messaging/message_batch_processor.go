package messaging

import (
	"context"
	"fmt"
	"strings"
)

// MessageBatchProcessor processa batches de mensagens
// Camada DESACOPLADA - pode ser usada ou não
type MessageBatchProcessor struct {
	// Strategies opcionais
	concatenator MessageConcatenator
	validator    MessageValidator
	enricher     MessageEnricher
	sender       MessageSender
}

// MessageConcatenator concatena mensagens em formato desejado
type MessageConcatenator interface {
	Concatenate(messages []BufferedMessage) (string, error)
}

// MessageValidator valida se batch deve ser processado
type MessageValidator interface {
	Validate(messages []BufferedMessage) error
}

// MessageEnricher enriquece mensagens com contexto adicional
type MessageEnricher interface {
	Enrich(ctx context.Context, messages []BufferedMessage) (interface{}, error)
}

// MessageSender envia mensagens processadas para destino final
type MessageSender interface {
	Send(ctx context.Context, sessionKey string, content string, metadata interface{}) error
}

// NewMessageBatchProcessor cria processor com strategies opcionais
func NewMessageBatchProcessor(
	concatenator MessageConcatenator,
	validator MessageValidator,
	enricher MessageEnricher,
	sender MessageSender,
) *MessageBatchProcessor {
	return &MessageBatchProcessor{
		concatenator: concatenator,
		validator:    validator,
		enricher:     enricher,
		sender:       sender,
	}
}

// Process processa um batch completo
func (p *MessageBatchProcessor) Process(
	ctx context.Context,
	sessionKey string,
	messages []BufferedMessage,
) error {
	if len(messages) == 0 {
		return nil
	}

	fmt.Printf("🔄 [Processor] Processing batch: session=%s, count=%d\n",
		sessionKey, len(messages))

	// 1. Validação (opcional)
	if p.validator != nil {
		if err := p.validator.Validate(messages); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	// 2. Concatenação (opcional)
	var content string
	var err error
	if p.concatenator != nil {
		content, err = p.concatenator.Concatenate(messages)
		if err != nil {
			return fmt.Errorf("concatenation failed: %w", err)
		}
	} else {
		// Default: concatena texto simples
		content = DefaultConcatenate(messages)
	}

	// 3. Enrichment (opcional)
	var metadata interface{}
	if p.enricher != nil {
		metadata, err = p.enricher.Enrich(ctx, messages)
		if err != nil {
			return fmt.Errorf("enrichment failed: %w", err)
		}
	}

	// 4. Envio (opcional)
	if p.sender != nil {
		if err := p.sender.Send(ctx, sessionKey, content, metadata); err != nil {
			return fmt.Errorf("send failed: %w", err)
		}
	}

	fmt.Printf("✅ [Processor] Batch processed successfully: session=%s\n", sessionKey)

	return nil
}

// DefaultConcatenate concatenação simples (fallback)
func DefaultConcatenate(messages []BufferedMessage) string {
	var builder strings.Builder
	for i, msg := range messages {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(msg.Text)
	}
	return builder.String()
}

// SimpleConcatenator implementação básica
type SimpleConcatenator struct{}

func (SimpleConcatenator) Concatenate(messages []BufferedMessage) (string, error) {
	return DefaultConcatenate(messages), nil
}

// MediaAwareConcatenator detecta mídia e ajusta concatenação
type MediaAwareConcatenator struct{}

func (MediaAwareConcatenator) Concatenate(messages []BufferedMessage) (string, error) {
	hasMedia := false
	var builder strings.Builder

	for i, msg := range messages {
		if i > 0 {
			builder.WriteString("\n")
		}

		// Detecta mídia
		if isMediaType(msg.Type) {
			hasMedia = true
			builder.WriteString(fmt.Sprintf("[%s: %s]", strings.ToUpper(msg.Type), msg.Text))
		} else {
			builder.WriteString(msg.Text)
		}
	}

	result := builder.String()
	if hasMedia {
		result = "[CONTÉM MÍDIA]\n" + result
	}

	return result, nil
}

func isMediaType(msgType string) bool {
	mediaTypes := []string{"image", "video", "audio", "document", "sticker", "voice"}
	for _, t := range mediaTypes {
		if msgType == t {
			return true
		}
	}
	return false
}

// MinMessageValidator garante mínimo de mensagens
type MinMessageValidator struct {
	MinCount int
}

func (v MinMessageValidator) Validate(messages []BufferedMessage) error {
	if len(messages) < v.MinCount {
		return fmt.Errorf("batch too small: %d < %d", len(messages), v.MinCount)
	}
	return nil
}

// NoopValidator não valida nada (passa tudo)
type NoopValidator struct{}

func (NoopValidator) Validate(messages []BufferedMessage) error {
	return nil
}
