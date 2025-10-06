package channel

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/channel"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ChannelService gerencia canais de comunicação
type ChannelService struct {
	repo   channel.Repository
	logger *zap.Logger
}

// NewChannelService cria um novo serviço de canais
func NewChannelService(repo channel.Repository, logger *zap.Logger) *ChannelService {
	return &ChannelService{
		repo:   repo,
		logger: logger,
	}
}

// CreateChannelRequest representa os dados para criar um canal
type CreateChannelRequest struct {
	UserID    uuid.UUID `json:"user_id"`
	ProjectID uuid.UUID `json:"project_id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name" binding:"required"`
	Type      string    `json:"type" binding:"required"`
	
	// Configuração WAHA (opcional, apenas para canais WAHA)
	WAHAConfig *WAHAConfigRequest `json:"waha_config,omitempty"`
}

// WAHAConfigRequest representa a configuração WAHA
type WAHAConfigRequest struct {
	BaseURL    string `json:"base_url" binding:"required"`
	APIKey     string `json:"api_key"`     // Chave da API para autenticação
	Token      string `json:"token"`       // Token de acesso (alternativo à API key)
	SessionID  string `json:"session_id"`  // ID da sessão WAHA (equivale ao ExternalID)
	WebhookURL string `json:"webhook_url"`
}

// ChannelResponse representa a resposta de um canal
type ChannelResponse struct {
	ID               uuid.UUID                  `json:"id"`
	UserID           uuid.UUID                  `json:"user_id"`
	ProjectID        uuid.UUID                  `json:"project_id"`
	TenantID         string                     `json:"tenant_id"`
	Name             string                     `json:"name"`
	Type             string                     `json:"type"`
	Status           string                     `json:"status"`
	Config           map[string]interface{}     `json:"config,omitempty"`
	MessagesReceived int                        `json:"messages_received"`
	MessagesSent     int                        `json:"messages_sent"`
	LastMessageAt    *string                    `json:"last_message_at,omitempty"`
	LastErrorAt      *string                    `json:"last_error_at,omitempty"`
	LastError        string                     `json:"last_error,omitempty"`
	CreatedAt        string                     `json:"created_at"`
	UpdatedAt        string                     `json:"updated_at"`
}

// CreateChannel cria um novo canal
func (s *ChannelService) CreateChannel(ctx context.Context, req CreateChannelRequest) (*ChannelResponse, error) {
	channelType := channel.ChannelType(req.Type)
	
	var ch *channel.Channel
	var err error
	
	// Criar canal baseado no tipo
	switch channelType {
	case channel.TypeWAHA:
		if req.WAHAConfig == nil {
			return nil, fmt.Errorf("WAHA configuration is required for WAHA channels")
		}
		
		wahaConfig := channel.WAHAConfig{
			BaseURL: req.WAHAConfig.BaseURL,
			Auth: channel.WAHAAuth{
				APIKey: req.WAHAConfig.APIKey,
				Token:  req.WAHAConfig.Token,
			},
			SessionID:  req.WAHAConfig.SessionID,
			WebhookURL: req.WAHAConfig.WebhookURL,
		}
		
		ch, err = channel.NewWAHAChannel(req.UserID, req.ProjectID, req.TenantID, req.Name, wahaConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create WAHA channel: %w", err)
		}
		
	default:
		ch, err = channel.NewChannel(req.UserID, req.ProjectID, req.TenantID, req.Name, channelType)
		if err != nil {
			return nil, fmt.Errorf("failed to create channel: %w", err)
		}
	}
	
	// Persistir no banco
	if err := s.repo.Create(ch); err != nil {
		s.logger.Error("Failed to create channel",
			zap.Error(err),
			zap.String("name", req.Name),
			zap.String("type", req.Type),
			zap.String("user_id", req.UserID.String()))
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}
	
	s.logger.Info("Channel created successfully",
		zap.String("id", ch.ID.String()),
		zap.String("name", ch.Name),
		zap.String("type", string(ch.Type)),
		zap.String("user_id", ch.UserID.String()))
	
	return s.toResponse(ch), nil
}

// GetChannelsByUser retorna todos os canais de um usuário
func (s *ChannelService) GetChannelsByUser(ctx context.Context, userID uuid.UUID) ([]*ChannelResponse, error) {
	channels, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	
	responses := make([]*ChannelResponse, len(channels))
	for i, ch := range channels {
		responses[i] = s.toResponse(ch)
	}
	
	return responses, nil
}

// GetChannelsByProject retorna todos os canais de um projeto
func (s *ChannelService) GetChannelsByProject(ctx context.Context, projectID uuid.UUID) ([]*ChannelResponse, error) {
	channels, err := s.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	
	responses := make([]*ChannelResponse, len(channels))
	for i, ch := range channels {
		responses[i] = s.toResponse(ch)
	}
	
	return responses, nil
}

// GetChannel retorna um canal específico
func (s *ChannelService) GetChannel(ctx context.Context, channelID uuid.UUID) (*ChannelResponse, error) {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	
	return s.toResponse(ch), nil
}

// ActivateChannel ativa um canal
func (s *ChannelService) ActivateChannel(ctx context.Context, channelID uuid.UUID) error {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	
	ch.Activate()
	
	if err := s.repo.Update(ch); err != nil {
		return fmt.Errorf("failed to activate channel: %w", err)
	}
	
	s.logger.Info("Channel activated",
		zap.String("id", ch.ID.String()),
		zap.String("name", ch.Name))
	
	return nil
}

// DeactivateChannel desativa um canal
func (s *ChannelService) DeactivateChannel(ctx context.Context, channelID uuid.UUID) error {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	
	ch.Deactivate()
	
	if err := s.repo.Update(ch); err != nil {
		return fmt.Errorf("failed to deactivate channel: %w", err)
	}
	
	s.logger.Info("Channel deactivated",
		zap.String("id", ch.ID.String()),
		zap.String("name", ch.Name))
	
	return nil
}

// DeleteChannel deleta um canal
func (s *ChannelService) DeleteChannel(ctx context.Context, channelID uuid.UUID) error {
	if err := s.repo.Delete(channelID); err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}
	
	s.logger.Info("Channel deleted", zap.String("id", channelID.String()))
	return nil
}

// toResponse converte domínio para response
func (s *ChannelService) toResponse(ch *channel.Channel) *ChannelResponse {
	response := &ChannelResponse{
		ID:               ch.ID,
		UserID:           ch.UserID,
		ProjectID:        ch.ProjectID,
		TenantID:         ch.TenantID,
		Name:             ch.Name,
		Type:             string(ch.Type),
		Status:           string(ch.Status),
		Config:           ch.Config,
		MessagesReceived: ch.MessagesReceived,
		MessagesSent:     ch.MessagesSent,
		LastError:        ch.LastError,
		CreatedAt:        ch.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        ch.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	
	if ch.LastMessageAt != nil {
		lastMsg := ch.LastMessageAt.Format("2006-01-02T15:04:05Z07:00")
		response.LastMessageAt = &lastMsg
	}
	
	if ch.LastErrorAt != nil {
		lastErr := ch.LastErrorAt.Format("2006-01-02T15:04:05Z07:00")
		response.LastErrorAt = &lastErr
	}
	
	return response
}
