package agent

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AIMessage representa uma mensagem no contexto de IA
type AIMessage struct {
	Role      string // "user", "assistant", "system"
	Content   string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// AIConversationContext representa o contexto de uma conversa para IA
type AIConversationContext struct {
	SessionID       uuid.UUID
	ContactID       uuid.UUID
	ContactName     string
	ContactPhone    string
	Messages        []AIMessage
	SessionMetadata map[string]interface{}
	ProjectContext  map[string]interface{} // Informações do projeto/empresa
}

// AIResponse representa a resposta de um agente de IA
type AIResponse struct {
	Message          string
	Confidence       float64 // 0.0 a 1.0
	ShouldEscalate   bool    // Se deve escalar para humano
	SuggestedActions []string
	Metadata         map[string]interface{}
	ResponseTimeMs   int
}

// AIProvider é a interface para provedores de IA externos
// Implementações: OpenAI, Anthropic, Google AI, Azure OpenAI, etc.
type AIProvider interface {
	// GetName retorna o nome do provider (ex: "openai", "anthropic")
	GetName() string

	// GenerateResponse gera uma resposta baseada no contexto da conversa
	GenerateResponse(ctx context.Context, conversation AIConversationContext) (*AIResponse, error)

	// ValidateConfig valida a configuração do provider
	ValidateConfig(config map[string]interface{}) error

	// IsHealthy verifica se o provider está saudável/disponível
	IsHealthy(ctx context.Context) bool
}

// AIProviderFactory cria instâncias de AI providers
type AIProviderFactory interface {
	// CreateProvider cria um provider baseado na configuração
	CreateProvider(providerName string, config map[string]interface{}) (AIProvider, error)

	// ListAvailableProviders lista os providers disponíveis
	ListAvailableProviders() []string
}

// AIAgentService gerencia agentes de IA
type AIAgentService interface {
	// ProcessMessage processa uma mensagem usando um agente de IA
	ProcessMessage(ctx context.Context, agentID uuid.UUID, conversation AIConversationContext) (*AIResponse, error)

	// GetProvider retorna o provider configurado para um agente
	GetProvider(agentID uuid.UUID) (AIProvider, error)

	// UpdateProvider atualiza o provider de um agente
	UpdateProvider(agentID uuid.UUID, providerName string, config map[string]interface{}) error
}
