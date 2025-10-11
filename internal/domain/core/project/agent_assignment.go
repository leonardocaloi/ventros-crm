package project

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AssignmentStrategy define como os agentes são atribuídos às sessões
type AssignmentStrategy string

const (
	// StrategyRoundRobin - Reveza entre os agentes na ordem
	StrategyRoundRobin AssignmentStrategy = "round_robin"

	// StrategyLeastSessions - Atribui ao agente com menos sessões ativas
	StrategyLeastSessions AssignmentStrategy = "least_sessions"

	// StrategyManual - Sem atribuição automática
	StrategyManual AssignmentStrategy = "manual"
)

// ReassignmentTrigger define o que dispara uma reatribuição
type ReassignmentTrigger string

const (
	// TriggerInactivity - Reatribui após tempo de inatividade
	TriggerInactivity ReassignmentTrigger = "inactivity"

	// TriggerManual - Reatribuição manual por outro agente
	TriggerManual ReassignmentTrigger = "manual"

	// TriggerNoResponse - Sem resposta do agente após mensagem do contato
	TriggerNoResponse ReassignmentTrigger = "no_response"

	// TriggerWorkloadBalance - Balanceamento de carga entre agentes
	TriggerWorkloadBalance ReassignmentTrigger = "workload_balance"
)

// ReassignmentRule define uma regra de reatribuição automática
type ReassignmentRule struct {
	// Enabled - Se a regra está ativa
	Enabled bool `json:"enabled"`

	// Trigger - O que dispara a reatribuição
	Trigger ReassignmentTrigger `json:"trigger"`

	// InactivityTimeoutMinutes - Minutos de inatividade antes de reatribuir
	// Usado quando Trigger = TriggerInactivity ou TriggerNoResponse
	InactivityTimeoutMinutes int `json:"inactivity_timeout_minutes"`

	// MaxReassignments - Número máximo de reatribuições para uma sessão
	// 0 = ilimitado
	MaxReassignments int `json:"max_reassignments"`

	// NotifyPreviousAgent - Se deve notificar o agente anterior
	NotifyPreviousAgent bool `json:"notify_previous_agent"`

	// OnlyDuringBusinessHours - Se só deve reatribuir em horário comercial
	OnlyDuringBusinessHours bool `json:"only_during_business_hours"`
}

// NewReassignmentRule cria uma nova regra de reatribuição
func NewReassignmentRule(trigger ReassignmentTrigger, timeoutMinutes int) *ReassignmentRule {
	return &ReassignmentRule{
		Enabled:                  true,
		Trigger:                  trigger,
		InactivityTimeoutMinutes: timeoutMinutes,
		MaxReassignments:         3, // Default: 3 reatribuições máximas
		NotifyPreviousAgent:      true,
		OnlyDuringBusinessHours:  false,
	}
}

// Validate valida a regra de reatribuição
func (r *ReassignmentRule) Validate() error {
	if !r.Enabled {
		return nil
	}

	switch r.Trigger {
	case TriggerInactivity, TriggerNoResponse:
		if r.InactivityTimeoutMinutes <= 0 {
			return errors.New("inactivity timeout must be greater than 0")
		}
	case TriggerManual, TriggerWorkloadBalance:
		// Não precisa de timeout
	default:
		return fmt.Errorf("invalid reassignment trigger: %s", r.Trigger)
	}

	if r.MaxReassignments < 0 {
		return errors.New("max reassignments cannot be negative")
	}

	return nil
}

// ShouldReassign verifica se deve reatribuir baseado no tempo
func (r *ReassignmentRule) ShouldReassign(lastActivityAt time.Time, reassignmentCount int) bool {
	if !r.Enabled {
		return false
	}

	// Verifica limite de reatribuições
	if r.MaxReassignments > 0 && reassignmentCount >= r.MaxReassignments {
		return false
	}

	// Para triggers baseados em tempo, verifica inatividade
	if r.Trigger == TriggerInactivity || r.Trigger == TriggerNoResponse {
		inactiveDuration := time.Since(lastActivityAt)
		timeoutDuration := time.Duration(r.InactivityTimeoutMinutes) * time.Minute
		return inactiveDuration >= timeoutDuration
	}

	return false
}

// AgentAssignmentConfig configura como agentes são atribuídos às sessões
type AgentAssignmentConfig struct {
	// Enabled - Se a atribuição automática está habilitada
	Enabled bool `json:"enabled"`

	// AgentIDs - Lista de agentes disponíveis para atribuição
	// Pode incluir agentes humanos, IA ou virtuais
	AgentIDs []uuid.UUID `json:"agent_ids"`

	// Strategy - Estratégia de distribuição
	Strategy AssignmentStrategy `json:"strategy"`

	// RoundRobinIndex - Índice atual para round-robin (uso interno)
	RoundRobinIndex int `json:"round_robin_index"`

	// OnlyHumanAgents - Se true, só atribui agentes humanos (exclui virtuais/IA)
	OnlyHumanAgents bool `json:"only_human_agents"`

	// ExcludeVirtualAgents - Se true, exclui agentes virtuais da atribuição
	// (agentes virtuais não podem enviar mensagens)
	ExcludeVirtualAgents bool `json:"exclude_virtual_agents"`

	// ReassignmentRules - Regras de reatribuição automática
	ReassignmentRules []*ReassignmentRule `json:"reassignment_rules"`

	// RequireAtLeastOneAgent - Se true, exige pelo menos um agente atribuído
	// Sessions não podem ser criadas sem agente quando true
	RequireAtLeastOneAgent bool `json:"require_at_least_one_agent"`
}

// NewAgentAssignmentConfig cria uma nova configuração de atribuição
func NewAgentAssignmentConfig() *AgentAssignmentConfig {
	return &AgentAssignmentConfig{
		Enabled:                false,
		AgentIDs:               []uuid.UUID{},
		Strategy:               StrategyManual,
		RoundRobinIndex:        0,
		OnlyHumanAgents:        false,
		ExcludeVirtualAgents:   true, // Default: não atribui a virtuais
		ReassignmentRules:      []*ReassignmentRule{},
		RequireAtLeastOneAgent: true, // Default: exige pelo menos um agente
	}
}

// Validate valida a configuração
func (c *AgentAssignmentConfig) Validate() error {
	if !c.Enabled {
		return nil // Se desabilitado, não precisa validar
	}

	if c.RequireAtLeastOneAgent && len(c.AgentIDs) == 0 {
		return errors.New("at least one agent must be configured when assignment is enabled")
	}

	switch c.Strategy {
	case StrategyRoundRobin, StrategyLeastSessions, StrategyManual:
		// Valid strategies
	default:
		return fmt.Errorf("invalid assignment strategy: %s", c.Strategy)
	}

	// Valida regras de reatribuição
	for i, rule := range c.ReassignmentRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid reassignment rule at index %d: %w", i, err)
		}
	}

	return nil
}

// AddAgent adiciona um agente à lista
func (c *AgentAssignmentConfig) AddAgent(agentID uuid.UUID) error {
	if agentID == uuid.Nil {
		return errors.New("agent ID cannot be nil")
	}

	// Verifica duplicação
	for _, id := range c.AgentIDs {
		if id == agentID {
			return fmt.Errorf("agent %s is already in the list", agentID)
		}
	}

	c.AgentIDs = append(c.AgentIDs, agentID)
	return nil
}

// RemoveAgent remove um agente da lista
func (c *AgentAssignmentConfig) RemoveAgent(agentID uuid.UUID) error {
	for i, id := range c.AgentIDs {
		if id == agentID {
			c.AgentIDs = append(c.AgentIDs[:i], c.AgentIDs[i+1:]...)
			// Reset round-robin index if needed
			if c.RoundRobinIndex >= len(c.AgentIDs) {
				c.RoundRobinIndex = 0
			}
			return nil
		}
	}
	return fmt.Errorf("agent %s not found in the list", agentID)
}

// HasAgent verifica se um agente está na lista
func (c *AgentAssignmentConfig) HasAgent(agentID uuid.UUID) bool {
	for _, id := range c.AgentIDs {
		if id == agentID {
			return true
		}
	}
	return false
}

// GetNextAgentRoundRobin retorna o próximo agente usando round-robin
func (c *AgentAssignmentConfig) GetNextAgentRoundRobin() (uuid.UUID, error) {
	if len(c.AgentIDs) == 0 {
		return uuid.Nil, errors.New("no agents configured")
	}

	if c.RoundRobinIndex >= len(c.AgentIDs) {
		c.RoundRobinIndex = 0
	}

	agentID := c.AgentIDs[c.RoundRobinIndex]
	c.RoundRobinIndex++

	return agentID, nil
}

// GetAgentCount retorna o número de agentes configurados
func (c *AgentAssignmentConfig) GetAgentCount() int {
	return len(c.AgentIDs)
}

// Clear limpa todos os agentes
func (c *AgentAssignmentConfig) Clear() {
	c.AgentIDs = []uuid.UUID{}
	c.RoundRobinIndex = 0
}

// SetStrategy define a estratégia de atribuição
func (c *AgentAssignmentConfig) SetStrategy(strategy AssignmentStrategy) error {
	switch strategy {
	case StrategyRoundRobin, StrategyLeastSessions, StrategyManual:
		c.Strategy = strategy
		return nil
	default:
		return fmt.Errorf("invalid assignment strategy: %s", strategy)
	}
}

// Enable habilita a atribuição automática
func (c *AgentAssignmentConfig) Enable() {
	c.Enabled = true
}

// Disable desabilita a atribuição automática
func (c *AgentAssignmentConfig) Disable() {
	c.Enabled = false
}

// IsEnabled retorna se a atribuição automática está habilitada
func (c *AgentAssignmentConfig) IsEnabled() bool {
	return c.Enabled
}

// ShouldAutoAssign retorna se deve fazer atribuição automática
func (c *AgentAssignmentConfig) ShouldAutoAssign() bool {
	return c.Enabled && len(c.AgentIDs) > 0 && c.Strategy != StrategyManual
}

// ===== Reassignment Rule Management =====

// AddReassignmentRule adiciona uma regra de reatribuição
func (c *AgentAssignmentConfig) AddReassignmentRule(rule *ReassignmentRule) error {
	if rule == nil {
		return errors.New("reassignment rule cannot be nil")
	}

	if err := rule.Validate(); err != nil {
		return fmt.Errorf("invalid reassignment rule: %w", err)
	}

	c.ReassignmentRules = append(c.ReassignmentRules, rule)
	return nil
}

// RemoveReassignmentRule remove uma regra de reatribuição por índice
func (c *AgentAssignmentConfig) RemoveReassignmentRule(index int) error {
	if index < 0 || index >= len(c.ReassignmentRules) {
		return errors.New("invalid reassignment rule index")
	}

	c.ReassignmentRules = append(c.ReassignmentRules[:index], c.ReassignmentRules[index+1:]...)
	return nil
}

// GetActiveReassignmentRules retorna apenas as regras ativas
func (c *AgentAssignmentConfig) GetActiveReassignmentRules() []*ReassignmentRule {
	active := make([]*ReassignmentRule, 0)
	for _, rule := range c.ReassignmentRules {
		if rule.Enabled {
			active = append(active, rule)
		}
	}
	return active
}

// ShouldReassignSession verifica se uma sessão deve ser reatribuída
// baseado em todas as regras ativas
func (c *AgentAssignmentConfig) ShouldReassignSession(lastActivityAt time.Time, reassignmentCount int) (bool, *ReassignmentRule) {
	for _, rule := range c.ReassignmentRules {
		if rule.ShouldReassign(lastActivityAt, reassignmentCount) {
			return true, rule
		}
	}
	return false, nil
}

// ClearReassignmentRules remove todas as regras de reatribuição
func (c *AgentAssignmentConfig) ClearReassignmentRules() {
	c.ReassignmentRules = []*ReassignmentRule{}
}

// HasReassignmentRules retorna se há regras de reatribuição configuradas
func (c *AgentAssignmentConfig) HasReassignmentRules() bool {
	return len(c.ReassignmentRules) > 0
}

// GetReassignmentRule retorna uma regra específica por índice
func (c *AgentAssignmentConfig) GetReassignmentRule(index int) (*ReassignmentRule, error) {
	if index < 0 || index >= len(c.ReassignmentRules) {
		return nil, errors.New("invalid reassignment rule index")
	}
	return c.ReassignmentRules[index], nil
}
