package agent

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AIMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

type AIConversationContext struct {
	SessionID       uuid.UUID
	ContactID       uuid.UUID
	ContactName     string
	ContactPhone    string
	Messages        []AIMessage
	SessionMetadata map[string]interface{}
	ProjectContext  map[string]interface{}
}

type AIResponse struct {
	Message          string
	Confidence       float64
	ShouldEscalate   bool
	SuggestedActions []string
	Metadata         map[string]interface{}
	ResponseTimeMs   int
}

type AIProvider interface {
	GetName() string

	GenerateResponse(ctx context.Context, conversation AIConversationContext) (*AIResponse, error)

	ValidateConfig(config map[string]interface{}) error

	IsHealthy(ctx context.Context) bool
}

type AIProviderFactory interface {
	CreateProvider(providerName string, config map[string]interface{}) (AIProvider, error)

	ListAvailableProviders() []string
}

type AIAgentService interface {
	ProcessMessage(ctx context.Context, agentID uuid.UUID, conversation AIConversationContext) (*AIResponse, error)

	GetProvider(agentID uuid.UUID) (AIProvider, error)

	UpdateProvider(agentID uuid.UUID, providerName string, config map[string]interface{}) error
}
