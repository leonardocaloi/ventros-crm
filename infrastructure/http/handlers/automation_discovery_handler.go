package handlers

import (
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/http/dto"
	"github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/gin-gonic/gin"
)

// AutomationDiscoveryHandler lida com endpoints de discovery de automação
type AutomationDiscoveryHandler struct {
	triggerRegistry *pipeline.TriggerRegistry
}

// NewAutomationDiscoveryHandler cria novo handler
func NewAutomationDiscoveryHandler(triggerRegistry *pipeline.TriggerRegistry) *AutomationDiscoveryHandler {
	return &AutomationDiscoveryHandler{
		triggerRegistry: triggerRegistry,
	}
}

// GetAutomationTypes lista todos os tipos de automação disponíveis
//
//	@Summary		Lista tipos de automação
//	@Description	Retorna todos os tipos de automação disponíveis (follow-up, event, scheduled, etc)
//	@Tags			Automation
//	@Produce		json
//	@Success		200	{array}	dto.AutomationTypeResponse
//	@Router			/api/v1/automation/types [get]
func (h *AutomationDiscoveryHandler) GetAutomationTypes(c *gin.Context) {
	types := dto.GetAutomationTypes()
	c.JSON(http.StatusOK, types)
}

// GetTriggers lista todos os triggers disponíveis
//
//	@Summary		Lista triggers disponíveis
//	@Description	Retorna todos os triggers do sistema e customizados
//	@Tags			Automation
//	@Produce		json
//	@Param			category	query		string	false	"Filtrar por categoria (session, message, pipeline, temporal, transaction, behavior)"
//	@Success		200			{object}	object{system_triggers=[]dto.TriggerResponse,custom_triggers=[]dto.TriggerResponse}
//	@Router			/api/v1/automation/triggers [get]
func (h *AutomationDiscoveryHandler) GetTriggers(c *gin.Context) {
	category := c.Query("category")

	var systemTriggers, customTriggers []pipeline.TriggerMetadata

	if category != "" {
		// Filtrar por categoria
		cat := pipeline.TriggerCategory(category)
		allTriggers := h.triggerRegistry.ListTriggersByCategory(cat)

		for _, t := range allTriggers {
			if t.IsSystem {
				systemTriggers = append(systemTriggers, t)
			} else {
				customTriggers = append(customTriggers, t)
			}
		}
	} else {
		// Todos os triggers
		systemTriggers = h.triggerRegistry.ListSystemTriggers()
		customTriggers = h.triggerRegistry.ListCustomTriggers()
	}

	// Converter para DTOs
	systemDTO := make([]dto.TriggerResponse, len(systemTriggers))
	for i, t := range systemTriggers {
		systemDTO[i] = h.toTriggerResponse(t)
	}

	customDTO := make([]dto.TriggerResponse, len(customTriggers))
	for i, t := range customTriggers {
		customDTO[i] = h.toTriggerResponse(t)
	}

	c.JSON(http.StatusOK, gin.H{
		"system_triggers": systemDTO,
		"custom_triggers": customDTO,
	})
}

// GetTriggerDetails retorna detalhes de um trigger específico
//
//	@Summary		Detalhes de um trigger
//	@Description	Retorna metadados completos de um trigger incluindo parâmetros disponíveis
//	@Tags			Automation
//	@Produce		json
//	@Param			code	path		string	true	"Código do trigger"
//	@Success		200		{object}	dto.TriggerResponse
//	@Failure		404		{object}	object{error=string}
//	@Router			/api/v1/automation/triggers/{code} [get]
func (h *AutomationDiscoveryHandler) GetTriggerDetails(c *gin.Context) {
	code := c.Param("code")

	trigger, err := h.triggerRegistry.GetTrigger(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trigger not found"})
		return
	}

	c.JSON(http.StatusOK, h.toTriggerResponse(trigger))
}

// GetActions lista todas as ações disponíveis
//
//	@Summary		Lista ações disponíveis
//	@Description	Retorna todas as ações que podem ser executadas nas automações
//	@Tags			Automation
//	@Produce		json
//	@Param			category	query	string	false	"Filtrar por categoria (messaging, pipeline, assignment, tasks, integration, organization, data, workflow)"
//	@Success		200			{array}	dto.ActionResponse
//	@Router			/api/v1/automation/actions [get]
func (h *AutomationDiscoveryHandler) GetActions(c *gin.Context) {
	category := c.Query("category")

	actions := pipeline.GetAvailableActions()

	// Filtrar por categoria se especificado
	if category != "" {
		filtered := make([]pipeline.ActionMetadata, 0)
		for _, action := range actions {
			if action.Category == category {
				filtered = append(filtered, action)
			}
		}
		actions = filtered
	}

	// Converter para DTOs
	actionsDTO := make([]dto.ActionResponse, len(actions))
	for i, action := range actions {
		actionsDTO[i] = dto.ToActionResponse(action)
	}

	c.JSON(http.StatusOK, actionsDTO)
}

// GetConditionOperators lista todos os operadores de condição
//
//	@Summary		Lista operadores de condição
//	@Description	Retorna todos os operadores disponíveis para condições (eq, gt, lt, contains, etc)
//	@Tags			Automation
//	@Produce		json
//	@Success		200	{array}	dto.ConditionOperatorResponse
//	@Router			/api/v1/automation/conditions/operators [get]
func (h *AutomationDiscoveryHandler) GetConditionOperators(c *gin.Context) {
	operators := pipeline.GetAvailableOperators()

	operatorsDTO := make([]dto.ConditionOperatorResponse, len(operators))
	for i, op := range operators {
		operatorsDTO[i] = dto.ToConditionOperatorResponse(op)
	}

	c.JSON(http.StatusOK, operatorsDTO)
}

// GetLogicOperators lista operadores lógicos (AND/OR)
//
//	@Summary		Lista operadores lógicos
//	@Description	Retorna operadores lógicos disponíveis para combinar condições
//	@Tags			Automation
//	@Produce		json
//	@Success		200	{array}	dto.LogicOperatorResponse
//	@Router			/api/v1/automation/logic-operators [get]
func (h *AutomationDiscoveryHandler) GetLogicOperators(c *gin.Context) {
	operators := dto.GetLogicOperators()
	c.JSON(http.StatusOK, operators)
}

// GetFullDiscovery retorna todos os metadados de automação em uma única chamada
//
//	@Summary		Discovery completo de automação
//	@Description	Retorna tipos, triggers, ações, operadores e lógica em uma única resposta
//	@Tags			Automation
//	@Produce		json
//	@Success		200	{object}	dto.AutomationDiscoveryResponse
//	@Router			/api/v1/automation/discovery [get]
func (h *AutomationDiscoveryHandler) GetFullDiscovery(c *gin.Context) {
	// Tipos
	types := dto.GetAutomationTypes()

	// Triggers
	systemTriggers := h.triggerRegistry.ListSystemTriggers()
	customTriggers := h.triggerRegistry.ListCustomTriggers()

	systemDTO := make([]dto.TriggerResponse, len(systemTriggers))
	for i, t := range systemTriggers {
		systemDTO[i] = h.toTriggerResponse(t)
	}

	customDTO := make([]dto.TriggerResponse, len(customTriggers))
	for i, t := range customTriggers {
		customDTO[i] = h.toTriggerResponse(t)
	}

	allTriggersDTO := append(systemDTO, customDTO...)

	// Actions
	actions := pipeline.GetAvailableActions()
	actionsDTO := make([]dto.ActionResponse, len(actions))
	for i, action := range actions {
		actionsDTO[i] = dto.ToActionResponse(action)
	}

	// Operators
	operators := pipeline.GetAvailableOperators()
	operatorsDTO := make([]dto.ConditionOperatorResponse, len(operators))
	for i, op := range operators {
		operatorsDTO[i] = dto.ToConditionOperatorResponse(op)
	}

	// Logic
	logicOps := dto.GetLogicOperators()

	response := dto.AutomationDiscoveryResponse{
		Types:      types,
		Triggers:   allTriggersDTO,
		Actions:    actionsDTO,
		Operators:  operatorsDTO,
		LogicTypes: logicOps,
	}

	c.JSON(http.StatusOK, response)
}

// toTriggerResponse converte TriggerMetadata para DTO
func (h *AutomationDiscoveryHandler) toTriggerResponse(t pipeline.TriggerMetadata) dto.TriggerResponse {
	params := make([]dto.TriggerParameter, len(t.Parameters))
	for i, p := range t.Parameters {
		params[i] = dto.TriggerParameter{
			Name:        p.Name,
			Type:        p.Type,
			Description: p.Description,
			Example:     p.Example,
		}
	}

	return dto.TriggerResponse{
		Code:        t.Code,
		Name:        t.Name,
		Description: t.Description,
		Category:    string(t.Category),
		IsSystem:    t.IsSystem,
		Parameters:  params,
	}
}

// RegisterCustomTrigger permite registrar trigger customizado via API
//
//	@Summary		Registrar trigger customizado
//	@Description	Permite que admins registrem triggers customizados com prefixo 'custom.'
//	@Tags			Automation
//	@Accept			json
//	@Produce		json
//	@Param			trigger	body		object{code=string,name=string,description=string,parameters=[]object{name=string,type=string,description=string}}	true	"Metadados do trigger"
//	@Success		201		{object}	object{message=string,trigger=dto.TriggerResponse}
//	@Failure		400		{object}	object{error=string}
//	@Router			/api/v1/automation/triggers/custom [post]
func (h *AutomationDiscoveryHandler) RegisterCustomTrigger(c *gin.Context) {
	var req struct {
		Code        string `json:"code" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Parameters  []struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			Description string `json:"description"`
			Example     string `json:"example"`
		} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Converte parâmetros
	params := make([]pipeline.TriggerParameter, len(req.Parameters))
	for i, p := range req.Parameters {
		params[i] = pipeline.TriggerParameter{
			Name:        p.Name,
			Type:        p.Type,
			Description: p.Description,
			Example:     p.Example,
		}
	}

	// Cria metadata
	trigger := pipeline.TriggerMetadata{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Parameters:  params,
	}

	// Registra
	if err := h.triggerRegistry.RegisterCustomTrigger(trigger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "custom trigger registered successfully",
		"trigger": h.toTriggerResponse(trigger),
	})
}

// UnregisterCustomTrigger remove trigger customizado
//
//	@Summary		Remover trigger customizado
//	@Description	Remove trigger customizado previamente registrado
//	@Tags			Automation
//	@Param			code	path		string	true	"Código do trigger customizado"
//	@Success		200		{object}	object{message=string}
//	@Failure		400		{object}	object{error=string}
//	@Router			/api/v1/automation/triggers/custom/{code} [delete]
func (h *AutomationDiscoveryHandler) UnregisterCustomTrigger(c *gin.Context) {
	code := c.Param("code")

	if err := h.triggerRegistry.UnregisterCustomTrigger(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "custom trigger unregistered successfully"})
}
