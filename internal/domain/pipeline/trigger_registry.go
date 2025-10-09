package pipeline

import (
	"errors"
	"fmt"
	"sync"
)

// TriggerRegistry registra triggers disponíveis (system + custom)
type TriggerRegistry struct {
	mu             sync.RWMutex
	systemTriggers map[string]TriggerMetadata
	customTriggers map[string]TriggerMetadata
}

// TriggerMetadata metadados sobre um trigger
type TriggerMetadata struct {
	Code        string
	Name        string
	Description string
	Category    TriggerCategory
	IsSystem    bool
	Parameters  []TriggerParameter
}

// TriggerCategory categoria do trigger
type TriggerCategory string

const (
	CategorySession     TriggerCategory = "session"
	CategoryMessage     TriggerCategory = "message"
	CategoryPipeline    TriggerCategory = "pipeline"
	CategoryTemporal    TriggerCategory = "temporal"
	CategoryTransaction TriggerCategory = "transaction"
	CategoryBehavior    TriggerCategory = "behavior"
	CategoryCustom      TriggerCategory = "custom"
	CategoryWebhook     TriggerCategory = "webhook"
)

// TriggerParameter parâmetro que o trigger disponibiliza no contexto
type TriggerParameter struct {
	Name        string
	Type        string // "string", "int", "float", "bool", "uuid"
	Description string
	Example     string
}

// NewTriggerRegistry cria registry com triggers do sistema
func NewTriggerRegistry() *TriggerRegistry {
	registry := &TriggerRegistry{
		systemTriggers: make(map[string]TriggerMetadata),
		customTriggers: make(map[string]TriggerMetadata),
	}

	// Registra triggers do sistema
	registry.registerSystemTriggers()

	return registry
}

// registerSystemTriggers registra os 10 triggers hard-coded
func (r *TriggerRegistry) registerSystemTriggers() {
	systemTriggers := []TriggerMetadata{
		{
			Code:        string(TriggerSessionEnded),
			Name:        "Sessão Encerrada",
			Description: "Disparado quando uma sessão é encerrada normalmente",
			Category:    CategorySession,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "session_id", Type: "uuid", Description: "ID da sessão"},
				{Name: "contact_id", Type: "uuid", Description: "ID do contato"},
				{Name: "session_duration_minutes", Type: "float", Description: "Duração da sessão em minutos"},
				{Name: "message_count", Type: "int", Description: "Total de mensagens na sessão"},
				{Name: "resolved", Type: "bool", Description: "Se sessão foi resolvida"},
			},
		},
		{
			Code:        string(TriggerSessionTimeout),
			Name:        "Sessão Expirou",
			Description: "Disparado quando uma sessão expira por inatividade",
			Category:    CategorySession,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "session_id", Type: "uuid"},
				{Name: "contact_id", Type: "uuid"},
				{Name: "hours_since_last_message", Type: "float", Description: "Horas desde última mensagem"},
				{Name: "last_message_at", Type: "timestamp"},
			},
		},
		{
			Code:        string(TriggerSessionResolved),
			Name:        "Sessão Resolvida",
			Description: "Disparado quando uma sessão é marcada como resolvida",
			Category:    CategorySession,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "session_id", Type: "uuid"},
				{Name: "contact_id", Type: "uuid"},
				{Name: "agent_id", Type: "uuid", Description: "Agente que resolveu"},
				{Name: "resolution_time_minutes", Type: "float"},
			},
		},
		{
			Code:        string(TriggerSessionEscalated),
			Name:        "Sessão Escalada",
			Description: "Disparado quando uma sessão é escalada para outro nível",
			Category:    CategorySession,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "session_id", Type: "uuid"},
				{Name: "contact_id", Type: "uuid"},
				{Name: "from_queue_id", Type: "uuid"},
				{Name: "to_queue_id", Type: "uuid"},
				{Name: "escalation_reason", Type: "string"},
			},
		},
		{
			Code:        string(TriggerNoResponse),
			Name:        "Sem Resposta",
			Description: "Disparado quando cliente não responde há X tempo",
			Category:    CategoryMessage,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "session_id", Type: "uuid"},
				{Name: "contact_id", Type: "uuid"},
				{Name: "hours_since_last_message", Type: "float"},
				{Name: "message_count", Type: "int"},
			},
		},
		{
			Code:        string(TriggerMessageReceived),
			Name:        "Mensagem Recebida",
			Description: "Disparado quando uma nova mensagem é recebida",
			Category:    CategoryMessage,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "session_id", Type: "uuid"},
				{Name: "contact_id", Type: "uuid"},
				{Name: "message_id", Type: "uuid"},
				{Name: "message_count", Type: "int", Description: "Total de mensagens na sessão"},
			},
		},
		{
			Code:        string(TriggerStatusChanged),
			Name:        "Status Mudou",
			Description: "Disparado quando status do contato muda no pipeline",
			Category:    CategoryPipeline,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "pipeline_id", Type: "uuid"},
				{Name: "old_status_id", Type: "uuid"},
				{Name: "new_status_id", Type: "uuid"},
			},
		},
		{
			Code:        string(TriggerStageCompleted),
			Name:        "Etapa Concluída",
			Description: "Disparado quando uma etapa do pipeline é concluída",
			Category:    CategoryPipeline,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "pipeline_id", Type: "uuid"},
				{Name: "stage_id", Type: "uuid"},
				{Name: "completion_time_minutes", Type: "float"},
			},
		},
		{
			Code:        string(TriggerAfterDelay),
			Name:        "Após Delay",
			Description: "Disparado após um delay específico desde um evento",
			Category:    CategoryTemporal,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "original_event", Type: "string"},
				{Name: "delay_minutes", Type: "int"},
			},
		},
		{
			Code:        string(TriggerScheduled),
			Name:        "Agendado",
			Description: "Disparado em horários agendados (cron, recorrente)",
			Category:    CategoryTemporal,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "scheduled_at", Type: "timestamp"},
				{Name: "schedule_type", Type: "string", Description: "once, daily, weekly, monthly, cron"},
			},
		},
		{
			Code:        string(TriggerPurchaseCompleted),
			Name:        "Compra Concluída",
			Description: "Disparado quando uma compra é finalizada com sucesso",
			Category:    CategoryTransaction,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "order_id", Type: "uuid"},
				{Name: "amount", Type: "float", Description: "Valor da compra"},
				{Name: "payment_method", Type: "string", Description: "Método de pagamento utilizado"},
				{Name: "items_count", Type: "int", Description: "Quantidade de itens"},
			},
		},
		{
			Code:        string(TriggerPaymentReceived),
			Name:        "Pagamento Recebido",
			Description: "Disparado quando o pagamento é confirmado",
			Category:    CategoryTransaction,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "payment_id", Type: "uuid"},
				{Name: "amount", Type: "float"},
				{Name: "payment_method", Type: "string"},
			},
		},
		{
			Code:        string(TriggerRefundIssued),
			Name:        "Reembolso Emitido",
			Description: "Disparado quando um reembolso é processado",
			Category:    CategoryTransaction,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "refund_id", Type: "uuid"},
				{Name: "amount", Type: "float"},
				{Name: "reason", Type: "string"},
			},
		},
		{
			Code:        string(TriggerCartAbandoned),
			Name:        "Carrinho Abandonado",
			Description: "Disparado quando carrinho é abandonado sem finalizar compra",
			Category:    CategoryTransaction,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "cart_id", Type: "uuid"},
				{Name: "cart_value", Type: "float"},
				{Name: "items_count", Type: "int"},
				{Name: "hours_since_abandonment", Type: "float"},
			},
		},
		{
			Code:        string(TriggerOrderShipped),
			Name:        "Pedido Enviado",
			Description: "Disparado quando pedido é despachado para entrega",
			Category:    CategoryTransaction,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "order_id", Type: "uuid"},
				{Name: "tracking_code", Type: "string"},
				{Name: "carrier", Type: "string"},
			},
		},
		{
			Code:        string(TriggerPageVisited),
			Name:        "Página Visitada",
			Description: "Disparado quando contato visita página específica",
			Category:    CategoryBehavior,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "page_url", Type: "string"},
				{Name: "page_title", Type: "string"},
				{Name: "visit_count", Type: "int", Description: "Número de visitas"},
			},
		},
		{
			Code:        string(TriggerFormSubmitted),
			Name:        "Formulário Enviado",
			Description: "Disparado quando contato submete formulário",
			Category:    CategoryBehavior,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "form_id", Type: "string"},
				{Name: "form_name", Type: "string"},
				{Name: "form_data", Type: "object"},
			},
		},
		{
			Code:        string(TriggerFileDownloaded),
			Name:        "Arquivo Baixado",
			Description: "Disparado quando contato baixa arquivo/recurso",
			Category:    CategoryBehavior,
			IsSystem:    true,
			Parameters: []TriggerParameter{
				{Name: "contact_id", Type: "uuid"},
				{Name: "file_name", Type: "string"},
				{Name: "file_type", Type: "string"},
				{Name: "file_size_mb", Type: "float"},
			},
		},
	}

	for _, trigger := range systemTriggers {
		r.systemTriggers[trigger.Code] = trigger
	}
}

// RegisterCustomTrigger registra trigger customizado
func (r *TriggerRegistry) RegisterCustomTrigger(trigger TriggerMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Valida
	if trigger.Code == "" {
		return errors.New("trigger code cannot be empty")
	}

	// Não pode sobrescrever trigger do sistema
	if _, exists := r.systemTriggers[trigger.Code]; exists {
		return fmt.Errorf("cannot override system trigger: %s", trigger.Code)
	}

	// Prefixo obrigatório para custom triggers
	if len(trigger.Code) < 7 || trigger.Code[:7] != "custom." {
		return errors.New("custom triggers must start with 'custom.' prefix")
	}

	trigger.IsSystem = false
	trigger.Category = CategoryCustom

	r.customTriggers[trigger.Code] = trigger
	return nil
}

// UnregisterCustomTrigger remove trigger customizado
func (r *TriggerRegistry) UnregisterCustomTrigger(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Não pode remover trigger do sistema
	if _, exists := r.systemTriggers[code]; exists {
		return errors.New("cannot unregister system trigger")
	}

	delete(r.customTriggers, code)
	return nil
}

// IsValidTrigger verifica se trigger é válido (system ou custom)
func (r *TriggerRegistry) IsValidTrigger(code string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, isSystem := r.systemTriggers[code]
	_, isCustom := r.customTriggers[code]

	return isSystem || isCustom
}

// GetTrigger busca metadados de um trigger
func (r *TriggerRegistry) GetTrigger(code string) (TriggerMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if trigger, exists := r.systemTriggers[code]; exists {
		return trigger, nil
	}

	if trigger, exists := r.customTriggers[code]; exists {
		return trigger, nil
	}

	return TriggerMetadata{}, fmt.Errorf("trigger not found: %s", code)
}

// ListSystemTriggers lista todos os triggers do sistema
func (r *TriggerRegistry) ListSystemTriggers() []TriggerMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	triggers := make([]TriggerMetadata, 0, len(r.systemTriggers))
	for _, trigger := range r.systemTriggers {
		triggers = append(triggers, trigger)
	}
	return triggers
}

// ListCustomTriggers lista triggers customizados
func (r *TriggerRegistry) ListCustomTriggers() []TriggerMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	triggers := make([]TriggerMetadata, 0, len(r.customTriggers))
	for _, trigger := range r.customTriggers {
		triggers = append(triggers, trigger)
	}
	return triggers
}

// ListAllTriggers lista todos (system + custom)
func (r *TriggerRegistry) ListAllTriggers() []TriggerMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	triggers := make([]TriggerMetadata, 0, len(r.systemTriggers)+len(r.customTriggers))

	for _, trigger := range r.systemTriggers {
		triggers = append(triggers, trigger)
	}

	for _, trigger := range r.customTriggers {
		triggers = append(triggers, trigger)
	}

	return triggers
}

// ListTriggersByCategory filtra por categoria
func (r *TriggerRegistry) ListTriggersByCategory(category TriggerCategory) []TriggerMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	triggers := make([]TriggerMetadata, 0)

	for _, trigger := range r.systemTriggers {
		if trigger.Category == category {
			triggers = append(triggers, trigger)
		}
	}

	for _, trigger := range r.customTriggers {
		if trigger.Category == category {
			triggers = append(triggers, trigger)
		}
	}

	return triggers
}

// GetParametersForTrigger retorna parâmetros disponíveis para um trigger
func (r *TriggerRegistry) GetParametersForTrigger(code string) ([]TriggerParameter, error) {
	trigger, err := r.GetTrigger(code)
	if err != nil {
		return nil, err
	}
	return trigger.Parameters, nil
}
