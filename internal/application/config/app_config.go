package config

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AppConfig contém configurações da aplicação carregadas do banco.
type AppConfig struct {
	ChannelTypes map[string]int // name -> id
}

// UserConfig contém configurações específicas do usuário
type UserConfig struct {
	UserID    uuid.UUID
	ProjectID uuid.UUID
	TenantID  string
}

// AppConfigService carrega configurações do banco.
type AppConfigService struct {
	db *gorm.DB
}

// NewAppConfigService cria um novo serviço de configuração.
func NewAppConfigService(db *gorm.DB) *AppConfigService {
	return &AppConfigService{
		db: db,
	}
}

// LoadConfig carrega as configurações iniciais do banco.
func (s *AppConfigService) LoadConfig(ctx context.Context) (*AppConfig, error) {
	cfg := &AppConfig{
		ChannelTypes: map[string]int{
			"waha":      1,
			"whatsapp":  2,
			"direct_ig": 3,
			"messenger": 4,
			"telegram":  5,
		},
	}
	
	return cfg, nil
}

// GetUserConfig retorna configurações específicas do usuário
func (s *AppConfigService) GetUserConfig(userID, projectID uuid.UUID, tenantID string) *UserConfig {
	return &UserConfig{
		UserID:    userID,
		ProjectID: projectID,
		TenantID:  tenantID,
	}
}

// GetChannelTypeID retorna o ID de um channel type por nome.
func (cfg *AppConfig) GetChannelTypeID(name string) (int, error) {
	id, ok := cfg.ChannelTypes[name]
	if !ok {
		return 0, fmt.Errorf("channel type '%s' not found", name)
	}
	return id, nil
}
