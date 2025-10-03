package config

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AppConfig contém configurações da aplicação carregadas do banco.
type AppConfig struct {
	DefaultProjectID  uuid.UUID
	DefaultCustomerID uuid.UUID
	DefaultTenantID   string
	ChannelTypes      map[string]int // name -> id
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
	// TODO: Implementar busca no banco com GORM
	// Por enquanto, usar valores hardcoded para permitir compilação
	cfg := &AppConfig{
		DefaultProjectID:  uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		DefaultCustomerID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		DefaultTenantID:   "default",
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

// GetChannelTypeID retorna o ID de um channel type por nome.
func (cfg *AppConfig) GetChannelTypeID(name string) (int, error) {
	id, ok := cfg.ChannelTypes[name]
	if !ok {
		return 0, fmt.Errorf("channel type '%s' not found", name)
	}
	return id, nil
}
