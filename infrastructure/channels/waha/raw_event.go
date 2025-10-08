package waha

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WAHARawEvent representa um evento WAHA bruto recebido via webhook
// antes de qualquer parsing ou validação. Esta estrutura nunca falha.
type WAHARawEvent struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Session   string            `json:"session"`   // Query param ?session=xxx
	Body      []byte            `json:"body"`      // JSON bruto do webhook
	Headers   map[string]string `json:"headers"`   // Headers HTTP relevantes
	Source    string            `json:"source"`    // "webhook", "retry", etc
	Metadata  map[string]string `json:"metadata"`  // Dados extras para debug
}

// NewWAHARawEvent cria um novo evento raw a partir de dados do webhook
func NewWAHARawEvent(session string, body []byte, headers map[string]string) WAHARawEvent {
	return WAHARawEvent{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Session:   session,
		Body:      body,
		Headers:   headers,
		Source:    "webhook",
		Metadata:  make(map[string]string),
	}
}

// WAHAProcessedEvent representa um evento WAHA após parsing e validação
type WAHAProcessedEvent struct {
	RawEventID string                 `json:"raw_event_id"` // Referência ao evento original
	EventType  string                 `json:"event_type"`   // "message", "call", etc
	Session    string                 `json:"session"`
	ParsedAt   time.Time              `json:"parsed_at"`
	Payload    map[string]interface{} `json:"payload"`      // Payload parseado
	Metadata   map[string]interface{} `json:"metadata"`     // Metadados extras
}

// WAHAParseError representa um erro de parsing com contexto
type WAHAParseError struct {
	RawEventID string    `json:"raw_event_id"`
	Error      string    `json:"error"`
	ErrorType  string    `json:"error_type"` // "json_unmarshal", "unsupported_media", etc
	OccurredAt time.Time `json:"occurred_at"`
	RawBody    []byte    `json:"raw_body"` // Para debug
}

// ToJSON serializa o evento raw para JSON
func (e WAHARawEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializa um evento raw do JSON
func (e *WAHARawEvent) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// AddMetadata adiciona metadados ao evento
func (e *WAHARawEvent) AddMetadata(key, value string) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
}

// GetContentType retorna o Content-Type do header
func (e WAHARawEvent) GetContentType() string {
	if contentType, exists := e.Headers["Content-Type"]; exists {
		return contentType
	}
	return "application/json" // Default
}

// GetBodySize retorna o tamanho do body em bytes
func (e WAHARawEvent) GetBodySize() int {
	return len(e.Body)
}

// IsRetry verifica se é um evento de retry
func (e WAHARawEvent) IsRetry() bool {
	return e.Source == "retry"
}

// MarkAsRetry marca o evento como retry
func (e *WAHARawEvent) MarkAsRetry() {
	e.Source = "retry"
	e.AddMetadata("retry_at", time.Now().Format(time.RFC3339))
}
