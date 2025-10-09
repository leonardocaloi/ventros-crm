package session

import (
	"context"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/channel"
	"github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/google/uuid"
)

// SessionTimeoutResolver resolve o timeout correto baseado na hierarquia:
// 1. Pipeline.SessionTimeoutMinutes (se Channel tem pipeline associado)
// 2. Channel.DefaultSessionTimeoutMinutes (se não tem pipeline)
// 3. Sistema Default: 30 minutos (fallback)
type SessionTimeoutResolver struct {
	channelRepo   channel.Repository
	pipelineRepo  pipeline.Repository
	systemDefault time.Duration
}

// NewSessionTimeoutResolver cria um novo resolver
func NewSessionTimeoutResolver(
	channelRepo channel.Repository,
	pipelineRepo pipeline.Repository,
) *SessionTimeoutResolver {
	return &SessionTimeoutResolver{
		channelRepo:   channelRepo,
		pipelineRepo:  pipelineRepo,
		systemDefault: 30 * time.Minute, // fallback do sistema
	}
}

// ResolveForChannel resolve timeout para um canal específico
func (r *SessionTimeoutResolver) ResolveForChannel(
	ctx context.Context,
	channelID uuid.UUID,
) (time.Duration, *uuid.UUID, error) {
	// 1. Busca canal
	ch, err := r.channelRepo.GetByID(channelID)
	if err != nil {
		// Se não encontrou canal, usa default do sistema
		return r.systemDefault, nil, nil
	}

	// 2. Se canal tem pipeline associado, usa timeout do pipeline
	if ch.PipelineID != nil && *ch.PipelineID != uuid.Nil {
		pipe, err := r.pipelineRepo.FindPipelineByID(ctx, *ch.PipelineID)
		if err == nil && pipe != nil && pipe.SessionTimeoutMinutes() != nil {
			timeout := time.Duration(*pipe.SessionTimeoutMinutes()) * time.Minute
			return timeout, ch.PipelineID, nil
		}
		// Se falhou ao buscar pipeline ou timeout é NULL, continua para próxima opção
	}

	// 3. Usa timeout do canal
	if ch.DefaultSessionTimeoutMinutes > 0 {
		timeout := time.Duration(ch.DefaultSessionTimeoutMinutes) * time.Minute
		return timeout, nil, nil
	}

	// 4. Fallback: default do sistema
	return r.systemDefault, nil, nil
}

// ResolveForContact resolve timeout baseado no contact e channel
// Útil quando você tem contactID mas não tem session ainda
func (r *SessionTimeoutResolver) ResolveForContact(
	ctx context.Context,
	contactID uuid.UUID,
	channelID uuid.UUID,
) (time.Duration, *uuid.UUID, error) {
	// Por enquanto, apenas delega para ResolveForChannel
	// No futuro, pode considerar preferências do contato
	return r.ResolveForChannel(ctx, channelID)
}

// ResolveWithFallback resolve com valor de fallback customizado
func (r *SessionTimeoutResolver) ResolveWithFallback(
	ctx context.Context,
	channelID uuid.UUID,
	fallback time.Duration,
) (time.Duration, *uuid.UUID, error) {
	timeout, pipelineID, err := r.ResolveForChannel(ctx, channelID)
	if err != nil || timeout == r.systemDefault {
		// Se houve erro ou retornou default do sistema, usa fallback customizado
		return fallback, pipelineID, err
	}
	return timeout, pipelineID, nil
}

// GetEffectiveTimeout retorna o timeout efetivo de um canal
// Versão simplificada que retorna apenas a duração
func (r *SessionTimeoutResolver) GetEffectiveTimeout(
	ctx context.Context,
	channelID uuid.UUID,
) time.Duration {
	timeout, _, _ := r.ResolveForChannel(ctx, channelID)
	return timeout
}

// TimeoutInfo contém informações detalhadas sobre o timeout resolvido
type TimeoutInfo struct {
	Duration     time.Duration
	Source       TimeoutSource
	PipelineID   *uuid.UUID
	PipelineName string
	ChannelName  string
}

// TimeoutSource indica de onde veio o timeout
type TimeoutSource string

const (
	TimeoutSourcePipeline      TimeoutSource = "pipeline"
	TimeoutSourceChannel       TimeoutSource = "channel"
	TimeoutSourceSystemDefault TimeoutSource = "system_default"
)

// ResolveWithDetails resolve e retorna informações detalhadas
func (r *SessionTimeoutResolver) ResolveWithDetails(
	ctx context.Context,
	channelID uuid.UUID,
) (*TimeoutInfo, error) {
	info := &TimeoutInfo{}

	// 1. Busca canal
	ch, err := r.channelRepo.GetByID(channelID)
	if err != nil {
		info.Duration = r.systemDefault
		info.Source = TimeoutSourceSystemDefault
		return info, nil
	}

	info.ChannelName = ch.Name

	// 2. Tenta usar pipeline
	if ch.PipelineID != nil && *ch.PipelineID != uuid.Nil {
		pipe, err := r.pipelineRepo.FindPipelineByID(ctx, *ch.PipelineID)
		if err == nil && pipe != nil && pipe.SessionTimeoutMinutes() != nil {
			info.Duration = time.Duration(*pipe.SessionTimeoutMinutes()) * time.Minute
			info.Source = TimeoutSourcePipeline
			info.PipelineID = ch.PipelineID
			info.PipelineName = pipe.Name()
			return info, nil
		}
	}

	// 3. Usa timeout do canal
	if ch.DefaultSessionTimeoutMinutes > 0 {
		info.Duration = time.Duration(ch.DefaultSessionTimeoutMinutes) * time.Minute
		info.Source = TimeoutSourceChannel
		return info, nil
	}

	// 4. Fallback
	info.Duration = r.systemDefault
	info.Source = TimeoutSourceSystemDefault
	return info, nil
}

// SetSystemDefault altera o default do sistema (útil para testes)
func (r *SessionTimeoutResolver) SetSystemDefault(duration time.Duration) {
	r.systemDefault = duration
}
