package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/pipeline"
)

// SendWebhookExecutor implementa a ação de enviar webhook
type SendWebhookExecutor struct {
	httpClient *http.Client
}

// NewSendWebhookExecutor cria um novo executor de webhook
func NewSendWebhookExecutor() *SendWebhookExecutor {
	return &SendWebhookExecutor{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Type retorna o tipo de ação
func (e *SendWebhookExecutor) Type() pipeline.AutomationAction {
	return pipeline.ActionSendWebhook
}

// Validate valida os parâmetros da ação
func (e *SendWebhookExecutor) Validate(params map[string]interface{}) error {
	// Valida URL
	url, ok := params["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("%w: url", pipeline.ErrMissingRequiredParam)
	}

	// Payload é opcional mas deve ser um objeto se fornecido
	if payload, ok := params["payload"]; ok && payload != nil {
		if _, ok := payload.(map[string]interface{}); !ok {
			return fmt.Errorf("payload must be an object")
		}
	}

	// Headers são opcionais mas devem ser um objeto se fornecidos
	if headers, ok := params["headers"]; ok && headers != nil {
		if _, ok := headers.(map[string]interface{}); !ok {
			return fmt.Errorf("headers must be an object")
		}
	}

	return nil
}

// Execute executa o envio do webhook
func (e *SendWebhookExecutor) Execute(ctx context.Context, params pipeline.ActionExecutionParams) error {
	// Extrai parâmetros
	url := params.Action.Params["url"].(string)

	// Prepara payload
	payload := make(map[string]interface{})
	if p, ok := params.Action.Params["payload"].(map[string]interface{}); ok {
		payload = p
	}

	// Adiciona dados de contexto ao payload
	payload["automation"] = map[string]interface{}{
		"rule_id":   params.RuleID.String(),
		"rule_name": params.RuleName,
		"tenant_id": params.TenantID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if params.ContactID != nil {
		payload["contact_id"] = params.ContactID.String()
	}
	if params.SessionID != nil {
		payload["session_id"] = params.SessionID.String()
	}
	if params.AgentID != nil {
		payload["agent_id"] = params.AgentID.String()
	}

	// Adiciona variáveis extras
	if len(params.Variables) > 0 {
		payload["variables"] = params.Variables
	}

	// Serializa payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Cria request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Headers padrão
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Ventros-CRM-Automation/1.0")

	// Headers customizados
	if headers, ok := params.Action.Params["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	// Envia request
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Lê resposta
	body, _ := io.ReadAll(resp.Body)

	// Verifica status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned error status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
