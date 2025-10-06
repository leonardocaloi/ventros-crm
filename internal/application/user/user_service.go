package user

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUserRequest representa os dados para criar um usuário
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

// CreateUserResponse representa a resposta da criação de usuário
type CreateUserResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	APIKey    string    `json:"api_key"`
	ProjectID uuid.UUID `json:"default_project_id"`
	PipelineID uuid.UUID `json:"default_pipeline_id"`
}

// LoginRequest representa os dados de login
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse representa a resposta do login
type LoginResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	APIKey    string    `json:"api_key"`
	ProjectID uuid.UUID `json:"default_project_id"`
}

// CreateUser cria um novo usuário com projeto e pipeline default
// Se o usuário já existir, retorna os dados existentes (idempotente)
func (s *UserService) CreateUser(req CreateUserRequest) (*CreateUserResponse, error) {
	// Define e valida role
	if req.Role == "" {
		req.Role = "user"
	}
	
	// Validar se a role é válida
	if _, err := user.ParseRole(req.Role); err != nil {
		return nil, fmt.Errorf("invalid role: %s", req.Role)
	}

	// Verifica se o usuário já existe
	var existingUser entities.UserEntity
	err := s.db.Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		// Usuário já existe, retorna dados existentes
		// Busca projeto default
		var project entities.ProjectEntity
		if err := s.db.Where("user_id = ? AND active = true", existingUser.ID).First(&project).Error; err != nil {
			return nil, fmt.Errorf("no active project found for existing user")
		}

		// Busca pipeline default
		var pipeline entities.PipelineEntity
		if err := s.db.Where("project_id = ? AND active = true", project.ID).First(&pipeline).Error; err != nil {
			return nil, fmt.Errorf("no active pipeline found for existing user")
		}

		// Gera nova API key para o usuário existente
		apiKey, err := s.generateAPIKey(s.db, existingUser.ID, "Default API Key")
		if err != nil {
			return nil, fmt.Errorf("failed to generate API key for existing user: %w", err)
		}

		return &CreateUserResponse{
			UserID:     existingUser.ID,
			Name:       existingUser.Name,
			Email:      existingUser.Email,
			Role:       existingUser.Role,
			APIKey:     apiKey,
			ProjectID:  project.ID,
			PipelineID: pipeline.ID,
		}, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Usuário não existe, cria novo
	// Hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Inicia transação
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Cria o usuário
	newUser := entities.UserEntity{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		Status:       "active",
	}

	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 2. Cria conta de faturamento automaticamente
	billingAccount := entities.BillingAccountEntity{
		ID:            uuid.New(),
		UserID:        newUser.ID,
		Name:          fmt.Sprintf("Conta de %s", req.Name),
		PaymentStatus: "pending", // Começa como pending
		BillingEmail:  req.Email,
		Suspended:     false,
	}

	if err := tx.Create(&billingAccount).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create billing account: %w", err)
	}

	// FAKE: Ativar pagamento automaticamente para desenvolvimento
	// TODO: Remover quando implementar fluxo real de pagamento
	fakePaymentMethods := `[{"type":"fake_card","last_digits":"1234","is_default":true}]`
	if err := tx.Model(&billingAccount).Updates(map[string]interface{}{
		"payment_status":  "active",
		"payment_methods": []byte(fakePaymentMethods),
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to activate fake payment: %w", err)
	}

	// 3. Cria projeto default
	project := entities.ProjectEntity{
		ID:               uuid.New(),
		UserID:           newUser.ID,
		BillingAccountID: billingAccount.ID,
		TenantID:         fmt.Sprintf("user-%s", newUser.ID.String()[:8]),
		Name:             "Projeto Principal",
		Description:      "Projeto padrão criado automaticamente",
		Active:           true,
	}

	if err := tx.Create(&project).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create default project: %w", err)
	}

	// 4. Cria pipeline default
	pipeline := entities.PipelineEntity{
		ID:          uuid.New(),
		ProjectID:   project.ID,
		TenantID:    project.TenantID,
		Name:        "Pipeline Principal",
		Description: "Pipeline padrão para novos contatos",
		Color:       "#3B82F6",
		Position:    0,
		Active:      true,
	}

	if err := tx.Create(&pipeline).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create default pipeline: %w", err)
	}

	// 4. Gera API Key
	apiKey, err := s.generateAPIKey(tx, newUser.ID, "Default API Key")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &CreateUserResponse{
		UserID:     newUser.ID,
		Name:       newUser.Name,
		Email:      newUser.Email,
		Role:       newUser.Role,
		APIKey:     apiKey,
		ProjectID:  project.ID,
		PipelineID: pipeline.ID,
	}, nil
}

// Login autentica um usuário e retorna sua API key ativa
func (s *UserService) Login(req LoginRequest) (*LoginResponse, error) {
	var user entities.UserEntity
	if err := s.db.Where("email = ? AND status = 'active'", req.Email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verifica senha
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Busca API key ativa ou cria uma nova
	var apiKeyEntity entities.UserAPIKeyEntity
	err := s.db.Where("user_id = ? AND active = true", user.ID).First(&apiKeyEntity).Error
	
	var apiKey string
	if err == gorm.ErrRecordNotFound {
		// Cria nova API key
		apiKey, err = s.generateAPIKey(s.db, user.ID, "Login API Key")
		if err != nil {
			return nil, fmt.Errorf("failed to generate API key: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	} else {
		// Usa API key existente (mas não retorna o valor real por segurança)
		// Em produção, você pode querer regenerar ou usar um sistema de refresh
		apiKey = "existing-key-hidden"
	}

	// Busca projeto default
	var project entities.ProjectEntity
	if err := s.db.Where("user_id = ? AND active = true", user.ID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("no active project found")
	}

	// Atualiza último uso da API key
	s.db.Model(&apiKeyEntity).Update("last_used", time.Now())

	return &LoginResponse{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		APIKey:    apiKey,
		ProjectID: project.ID,
	}, nil
}

// ValidateAPIKey valida uma API key e retorna o contexto do usuário
func (s *UserService) ValidateAPIKey(apiKey string) (*entities.UserEntity, *entities.ProjectEntity, error) {
	// Hash da API key para busca
	hasher := sha256.New()
	hasher.Write([]byte(apiKey))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	var apiKeyEntity entities.UserAPIKeyEntity
	if err := s.db.Where("key_hash = ? AND active = true", keyHash).First(&apiKeyEntity).Error; err != nil {
		return nil, nil, fmt.Errorf("invalid API key")
	}

	// Busca usuário
	var user entities.UserEntity
	if err := s.db.Where("id = ? AND status = 'active'", apiKeyEntity.UserID).First(&user).Error; err != nil {
		return nil, nil, fmt.Errorf("user not found or inactive")
	}

	// Busca projeto default
	var project entities.ProjectEntity
	if err := s.db.Where("user_id = ? AND active = true", user.ID).First(&project).Error; err != nil {
		return nil, nil, fmt.Errorf("no active project found")
	}

	// Atualiza último uso
	s.db.Model(&apiKeyEntity).Update("last_used", time.Now())

	return &user, &project, nil
}

// generateAPIKey gera uma nova API key para o usuário
func (s *UserService) generateAPIKey(tx *gorm.DB, userID uuid.UUID, name string) (string, error) {
	// Desativa API keys existentes (apenas 1 ativa por usuário)
	if err := tx.Model(&entities.UserAPIKeyEntity{}).Where("user_id = ?", userID).Update("active", false).Error; err != nil {
		return "", fmt.Errorf("failed to deactivate existing keys: %w", err)
	}

	// Gera nova API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	
	apiKey := hex.EncodeToString(keyBytes)
	
	// Hash da API key para armazenamento
	hasher := sha256.New()
	hasher.Write([]byte(apiKey))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	// Salva no banco
	apiKeyEntity := entities.UserAPIKeyEntity{
		ID:      uuid.New(),
		UserID:  userID,
		Name:    name,
		KeyHash: keyHash,
		Active:  true,
	}

	if err := tx.Create(&apiKeyEntity).Error; err != nil {
		return "", fmt.Errorf("failed to save API key: %w", err)
	}

	return apiKey, nil
}

// RegenerateAPIKey regenera a API key do usuário
func (s *UserService) RegenerateAPIKey(userID uuid.UUID, name string) (string, error) {
	return s.generateAPIKey(s.db, userID, name)
}
