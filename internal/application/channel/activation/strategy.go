package activation

import (
	"context"

	"github.com/ventros/crm/internal/domain/crm/channel"
)

// ChannelActivationStrategy define a interface para ativação de canais
// Cada tipo de canal (WAHA, Telegram, Twilio, etc.) implementa sua própria strategy
// Permite que diferentes tipos tenham lógicas de validação e ativação específicas
//
// Pattern: Strategy Pattern
// Usado para encapsular algoritmos de ativação diferentes por tipo de canal
type ChannelActivationStrategy interface {
	// CanActivate verifica pré-condições antes de tentar ativar
	// Ex: verificar se configuração está completa, se credenciais existem, etc.
	CanActivate(ctx context.Context, ch *channel.Channel) error

	// Activate executa a ativação efetiva do canal
	// Ex: fazer health check com API externa, configurar webhook, etc.
	// Retorna erro se a ativação falhar
	Activate(ctx context.Context, ch *channel.Channel) error

	// HealthCheck verifica o status atual do canal
	// Usado pelo polling worker para verificar canais em "activating"
	// Retorna: (isHealthy bool, statusMessage string, error)
	HealthCheck(ctx context.Context, ch *channel.Channel) (bool, string, error)

	// Compensate executa ações de compensação quando ativação falha
	// Ex: remover webhook configurado, limpar recursos criados, etc.
	Compensate(ctx context.Context, ch *channel.Channel) error
}

// ActivationResult encapsula o resultado de uma tentativa de ativação
type ActivationResult struct {
	Success     bool
	Message     string
	ShouldRetry bool // Se true, o worker pode tentar novamente depois
	RetryAfter  int  // Segundos para aguardar antes de retry
	Error       error
}

// HealthCheckResult encapsula o resultado de um health check
type HealthCheckResult struct {
	IsHealthy     bool
	Status        string // Ex: "WORKING", "SCAN_QR_CODE", "FAILED"
	Message       string
	LastCheckedAt string
}
