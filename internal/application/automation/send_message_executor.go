package automation

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	"github.com/google/uuid"
)

// MessageSender é a interface para enviar mensagens
// Será implementada pelo serviço de mensagens existente
type MessageSender interface {
	SendMessage(ctx context.Context, req SendMessageRequest) error
}

// SendMessageRequest representa uma requisição de envio de mensagem
type SendMessageRequest struct {
	TenantID  string
	ChannelID uuid.UUID
	ContactID uuid.UUID
	Content   string
}

// SendMessageExecutor implementa a ação de enviar mensagem
type SendMessageExecutor struct {
	messageSender MessageSender
}

// NewSendMessageExecutor cria um novo executor de envio de mensagens
func NewSendMessageExecutor(messageSender MessageSender) *SendMessageExecutor {
	return &SendMessageExecutor{
		messageSender: messageSender,
	}
}

// Type retorna o tipo de ação
func (e *SendMessageExecutor) Type() pipeline.AutomationAction {
	return pipeline.ActionSendMessage
}

// Validate valida os parâmetros da ação
func (e *SendMessageExecutor) Validate(params map[string]interface{}) error {
	// Valida content
	content, ok := params["content"].(string)
	if !ok || content == "" {
		return fmt.Errorf("%w: content", pipeline.ErrMissingRequiredParam)
	}

	// channel_id é opcional (pode vir do contexto)
	if channelIDStr, ok := params["channel_id"].(string); ok && channelIDStr != "" {
		if _, err := uuid.Parse(channelIDStr); err != nil {
			return fmt.Errorf("invalid channel_id: must be a valid UUID")
		}
	}

	return nil
}

// Execute executa o envio da mensagem
func (e *SendMessageExecutor) Execute(ctx context.Context, params pipeline.ActionExecutionParams) error {
	// Extrai content
	content := params.Action.Params["content"].(string)

	// TODO: Interpolar variáveis
	// content = interpolateVariables(content, params.Variables)

	// Determina channel_id
	var channelID uuid.UUID
	if channelIDStr, ok := params.Action.Params["channel_id"].(string); ok && channelIDStr != "" {
		channelID, _ = uuid.Parse(channelIDStr)
	} else {
		// Tenta obter do contexto
		// Por exemplo, do session ou contact
		return fmt.Errorf("channel_id not provided and could not be inferred from context")
	}

	// Determina contact_id
	if params.ContactID == nil {
		return fmt.Errorf("contact_id required for send_message action")
	}

	// Envia mensagem
	req := SendMessageRequest{
		TenantID:  params.TenantID,
		ChannelID: channelID,
		ContactID: *params.ContactID,
		Content:   content,
	}

	if err := e.messageSender.SendMessage(ctx, req); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
