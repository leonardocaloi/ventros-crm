package importpkg

import (
	"context"

	"github.com/ventros/crm/internal/domain/crm/channel"
)

// HistoryImportStrategy define a interface para importação de histórico de canais
// Cada tipo de canal (WAHA, Telegram, Twilio, etc.) implementa sua própria strategy
// Permite que diferentes tipos tenham lógicas de importação específicas
//
// Pattern: Strategy Pattern
// Usado para encapsular algoritmos de importação diferentes por tipo de canal
type HistoryImportStrategy interface {
	// CanImport verifica pré-condições antes de iniciar importação
	// Ex: verificar se canal está ativo, se configuração está completa, etc.
	// strategy: "time_range", "full", "recent"
	CanImport(ctx context.Context, ch *channel.Channel, strategy string) error

	// Import executa a importação efetiva do histórico
	// Inicia o workflow assíncrono de importação
	// Retorna workflowID e erro se a inicialização falhar
	Import(ctx context.Context, ch *channel.Channel, params ImportParams) (string, error)

	// GetImportStatus verifica o status atual da importação
	// Usado para consultar progresso de importações em andamento
	// Retorna status atual e erro se falhar
	GetImportStatus(ctx context.Context, ch *channel.Channel) (*ImportStatusInfo, error)

	// CancelImport cancela uma importação em andamento
	// Usado para interromper importações que estão demorando muito ou com erro
	CancelImport(ctx context.Context, ch *channel.Channel, reason string) error
}

// ImportParams encapsula parâmetros para execução de importação
type ImportParams struct {
	Strategy      string // "time_range", "full", "recent"
	TimeRangeDays int    // Para strategy="time_range"
	Limit         int    // Limite de mensagens por chat
	CorrelationID string // ID para tracking (Saga Pattern)
	UserID        string // Usuário que solicitou a importação
}

// ImportStatusInfo encapsula informações de status de importação
type ImportStatusInfo struct {
	Status           string                 // "pending", "running", "completed", "failed"
	WorkflowID       string                 // ID do workflow Temporal
	CorrelationID    string                 // ID de correlação (Saga)
	MessagesImported int                    // Total de mensagens importadas
	ChatsProcessed   int                    // Total de chats processados
	StartedAt        *string                // Timestamp de início
	CompletedAt      *string                // Timestamp de conclusão
	LastError        *string                // Último erro, se houver
	Stats            map[string]interface{} // Estatísticas adicionais
}

// ImportResult encapsula o resultado de uma tentativa de importação
type ImportResult struct {
	Success       bool
	WorkflowID    string
	CorrelationID string
	Message       string
	Error         error
}
