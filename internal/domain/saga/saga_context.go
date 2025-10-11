package saga

import (
	"context"

	"github.com/google/uuid"
)

// SagaContext carrega informações de Saga através do contexto da aplicação
// Usado para propagar correlation_id, saga_type, etc sem modificar assinaturas de funções

type contextKey string

const (
	correlationIDKey contextKey = "saga_correlation_id"
	sagaTypeKey      contextKey = "saga_type"
	sagaStepKey      contextKey = "saga_step"
	sagaStepNumber   contextKey = "saga_step_number"
	tenantIDKey      contextKey = "saga_tenant_id"
)

// SagaMetadata contém metadados da Saga para rastreamento
type SagaMetadata struct {
	CorrelationID string `json:"correlation_id"` // ID único da Saga
	SagaType      string `json:"saga_type"`      // Tipo da Saga (process_inbound_message, onboard_customer, etc)
	SagaStep      string `json:"saga_step"`      // Step atual (contact_created, session_started, etc)
	StepNumber    int    `json:"step_number"`    // Número do step (1, 2, 3...)
	TenantID      string `json:"tenant_id"`      // Tenant para facilitar queries
}

// WithSaga adiciona metadados de Saga ao contexto
func WithSaga(ctx context.Context, sagaType string) context.Context {
	correlationID := uuid.New().String()
	ctx = context.WithValue(ctx, correlationIDKey, correlationID)
	ctx = context.WithValue(ctx, sagaTypeKey, sagaType)
	ctx = context.WithValue(ctx, sagaStepNumber, 1)
	return ctx
}

// WithCorrelationID adiciona apenas correlation_id (quando já existe uma Saga)
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// WithStep adiciona informação do step atual
func WithStep(ctx context.Context, step string, stepNumber int) context.Context {
	ctx = context.WithValue(ctx, sagaStepKey, step)
	ctx = context.WithValue(ctx, sagaStepNumber, stepNumber)
	return ctx
}

// NextStep avança para o próximo step da Saga (incrementa step_number automaticamente)
func NextStep(ctx context.Context, step SagaStep) context.Context {
	currentStep, _ := GetStepNumber(ctx)
	return WithStep(ctx, string(step), currentStep+1)
}

// WithTenantID adiciona tenant_id ao contexto da Saga
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// GetCorrelationID retorna o correlation_id do contexto
func GetCorrelationID(ctx context.Context) (string, bool) {
	val := ctx.Value(correlationIDKey)
	if val == nil {
		return "", false
	}
	correlationID, ok := val.(string)
	return correlationID, ok
}

// GetSagaType retorna o tipo da Saga do contexto
func GetSagaType(ctx context.Context) (string, bool) {
	val := ctx.Value(sagaTypeKey)
	if val == nil {
		return "", false
	}
	sagaType, ok := val.(string)
	return sagaType, ok
}

// GetSagaStep retorna o step atual do contexto
func GetSagaStep(ctx context.Context) (string, bool) {
	val := ctx.Value(sagaStepKey)
	if val == nil {
		return "", false
	}
	step, ok := val.(string)
	return step, ok
}

// GetStepNumber retorna o número do step
func GetStepNumber(ctx context.Context) (int, bool) {
	val := ctx.Value(sagaStepNumber)
	if val == nil {
		return 0, false
	}
	stepNum, ok := val.(int)
	return stepNum, ok
}

// GetTenantID retorna o tenant_id do contexto
func GetTenantID(ctx context.Context) (string, bool) {
	val := ctx.Value(tenantIDKey)
	if val == nil {
		return "", false
	}
	tenantID, ok := val.(string)
	return tenantID, ok
}

// GetMetadata extrai todos os metadados de Saga do contexto
func GetMetadata(ctx context.Context) *SagaMetadata {
	// Se não há correlation_id, não há Saga ativa
	correlationID, ok := GetCorrelationID(ctx)
	if !ok {
		return nil
	}

	metadata := &SagaMetadata{
		CorrelationID: correlationID,
	}

	if sagaType, ok := GetSagaType(ctx); ok {
		metadata.SagaType = sagaType
	}
	if step, ok := GetSagaStep(ctx); ok {
		metadata.SagaStep = step
	}
	if stepNum, ok := GetStepNumber(ctx); ok {
		metadata.StepNumber = stepNum
	}
	if tenantID, ok := GetTenantID(ctx); ok {
		metadata.TenantID = tenantID
	}

	return metadata
}

// MustGetCorrelationID retorna correlation_id ou panic (use quando você TEM CERTEZA que existe)
func MustGetCorrelationID(ctx context.Context) string {
	correlationID, ok := GetCorrelationID(ctx)
	if !ok {
		panic("saga: correlation_id not found in context")
	}
	return correlationID
}
