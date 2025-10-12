package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/credential"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormCredentialRepository implementa credential.Repository usando GORM
type GormCredentialRepository struct {
	db *gorm.DB
}

// NewGormCredentialRepository cria um novo repositório de credenciais
func NewGormCredentialRepository(db *gorm.DB) *GormCredentialRepository {
	return &GormCredentialRepository{db: db}
}

// Save persiste uma credencial (cria ou atualiza)
func (r *GormCredentialRepository) Save(ctx context.Context, cred *credential.Credential) error {
	entity := r.toEntity(cred)

	// Verifica se já existe
	var existing entities.CredentialEntity
	err := r.db.WithContext(ctx).Where("id = ?", entity.ID).First(&existing).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create
		return r.db.WithContext(ctx).Create(entity).Error
	}

	// Update with optimistic locking
	result := r.db.WithContext(ctx).Model(&entities.CredentialEntity{}).
		Where("id = ? AND version = ?", entity.ID, existing.Version).
		Updates(map[string]interface{}{
			"version":                     existing.Version + 1, // Increment version
			"tenant_id":                   entity.TenantID,
			"project_id":                  entity.ProjectID,
			"credential_type":             entity.CredentialType,
			"name":                        entity.Name,
			"description":                 entity.Description,
			"encrypted_value_ciphertext":  entity.EncryptedValueCiphertext,
			"encrypted_value_nonce":       entity.EncryptedValueNonce,
			"metadata":                    entity.Metadata,
			"is_active":                   entity.IsActive,
			"expires_at":                  entity.ExpiresAt,
			"last_used_at":                entity.LastUsedAt,
			"oauth_access_token_ciphertext": entity.OAuthAccessTokenCiphertext,
			"oauth_access_token_nonce":    entity.OAuthAccessTokenNonce,
			"oauth_refresh_token_ciphertext": entity.OAuthRefreshTokenCiphertext,
			"oauth_refresh_token_nonce":   entity.OAuthRefreshTokenNonce,
			"oauth_token_type":            entity.OAuthTokenType,
			"oauth_expires_at":            entity.OAuthExpiresAt,
			"updated_at":                  entity.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
	if result.RowsAffected == 0 {
		return shared.NewOptimisticLockError(
			"Credential",
			entity.ID.String(),
			existing.Version,
			entity.Version,
		)
	}

	return nil
}

// FindByID busca uma credencial por ID
func (r *GormCredentialRepository) FindByID(ctx context.Context, id uuid.UUID) (*credential.Credential, error) {
	var entity entities.CredentialEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&entity), nil
}

// FindByTenantAndType busca credenciais por tenant e tipo
func (r *GormCredentialRepository) FindByTenantAndType(
	ctx context.Context,
	tenantID string,
	credType credential.CredentialType,
) ([]*credential.Credential, error) {
	var entities []entities.CredentialEntity
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND credential_type = ?", tenantID, credType.String()).
		Order("created_at DESC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return r.toDomainList(entities), nil
}

// FindByTenantAndName busca uma credencial específica por nome
func (r *GormCredentialRepository) FindByTenantAndName(
	ctx context.Context,
	tenantID string,
	name string,
) (*credential.Credential, error) {
	var entity entities.CredentialEntity
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND name = ?", tenantID, name).
		First(&entity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&entity), nil
}

// FindByProjectAndType busca credenciais de um projeto específico
func (r *GormCredentialRepository) FindByProjectAndType(
	ctx context.Context,
	projectID uuid.UUID,
	credType credential.CredentialType,
) ([]*credential.Credential, error) {
	var entities []entities.CredentialEntity
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND credential_type = ?", projectID, credType.String()).
		Order("created_at DESC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return r.toDomainList(entities), nil
}

// FindActiveByTenant busca todas as credenciais ativas de um tenant
func (r *GormCredentialRepository) FindActiveByTenant(
	ctx context.Context,
	tenantID string,
) ([]*credential.Credential, error) {
	var entities []entities.CredentialEntity
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("created_at DESC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return r.toDomainList(entities), nil
}

// FindExpiring busca credenciais que expiram em breve (para renovação)
func (r *GormCredentialRepository) FindExpiring(
	ctx context.Context,
	withinMinutes int,
) ([]*credential.Credential, error) {
	threshold := time.Now().Add(time.Duration(withinMinutes) * time.Minute)

	var entities []entities.CredentialEntity
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND oauth_expires_at IS NOT NULL AND oauth_expires_at <= ?", true, threshold).
		Order("oauth_expires_at ASC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return r.toDomainList(entities), nil
}

// Delete remove uma credencial (marca como inativa)
func (r *GormCredentialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.CredentialEntity{}).
		Where("id = ?", id).
		Update("is_active", false).
		Error
}

// toEntity converte domínio para entidade de persistência
func (r *GormCredentialRepository) toEntity(cred *credential.Credential) *entities.CredentialEntity {
	entity := &entities.CredentialEntity{
		ID:                       cred.ID(),
		Version:                  cred.Version(),
		TenantID:                 cred.TenantID(),
		ProjectID:                cred.ProjectID(),
		CredentialType:           cred.Type().String(),
		Name:                     cred.Name(),
		Description:              cred.Description(),
		EncryptedValueCiphertext: cred.EncryptedValue().Ciphertext(),
		EncryptedValueNonce:      cred.EncryptedValue().Nonce(),
		Metadata:                 entities.MetadataJSON(cred.Metadata()),
		IsActive:                 cred.IsActive(),
		ExpiresAt:                cred.ExpiresAt(),
		LastUsedAt:               cred.LastUsedAt(),
		CreatedAt:                cred.CreatedAt(),
		UpdatedAt:                cred.UpdatedAt(),
	}

	// OAuth Token (se existir)
	if token := cred.OAuthToken(); token != nil {
		// Extrai os valores criptografados do access token
		accessEncrypted := token.EncryptedAccessToken()
		accessCiphertext := accessEncrypted.Ciphertext()
		accessNonce := accessEncrypted.Nonce()
		tokenType := token.TokenType()

		entity.OAuthAccessTokenCiphertext = &accessCiphertext
		entity.OAuthAccessTokenNonce = &accessNonce
		entity.OAuthTokenType = &tokenType
		expiresAt := token.ExpiresAt()
		entity.OAuthExpiresAt = &expiresAt

		// Se tiver refresh token
		if token.HasRefreshToken() {
			refreshEncrypted := token.EncryptedRefreshToken()
			refreshCiphertext := refreshEncrypted.Ciphertext()
			refreshNonce := refreshEncrypted.Nonce()
			entity.OAuthRefreshTokenCiphertext = &refreshCiphertext
			entity.OAuthRefreshTokenNonce = &refreshNonce
		}
	}

	return entity
}

// toDomain converte entidade de persistência para domínio
func (r *GormCredentialRepository) toDomain(entity *entities.CredentialEntity) *credential.Credential {
	encryptedValue := credential.NewEncryptedValue(
		entity.EncryptedValueCiphertext,
		entity.EncryptedValueNonce,
	)

	// OAuth Token (se existir)
	var oauthToken *credential.OAuthToken
	if entity.OAuthAccessTokenCiphertext != nil && entity.OAuthAccessTokenNonce != nil {
		accessToken := credential.NewEncryptedValue(
			*entity.OAuthAccessTokenCiphertext,
			*entity.OAuthAccessTokenNonce,
		)

		var refreshToken credential.EncryptedValue
		if entity.OAuthRefreshTokenCiphertext != nil && entity.OAuthRefreshTokenNonce != nil {
			refreshToken = credential.NewEncryptedValue(
				*entity.OAuthRefreshTokenCiphertext,
				*entity.OAuthRefreshTokenNonce,
			)
		}

		// Reconstrói o OAuthToken usando ReconstructOAuthToken
		expiresAt := time.Time{}
		if entity.OAuthExpiresAt != nil {
			expiresAt = *entity.OAuthExpiresAt
		}

		oauthToken = credential.ReconstructOAuthToken(
			accessToken,
			refreshToken,
			expiresAt,
			entity.OAuthTokenType,
		)
	}

	metadata := make(map[string]interface{})
	if entity.Metadata != nil {
		metadata = map[string]interface{}(entity.Metadata)
	}

	return credential.ReconstructCredential(
		entity.ID,
		entity.Version,
		entity.TenantID,
		entity.ProjectID,
		credential.CredentialType(entity.CredentialType),
		entity.Name,
		entity.Description,
		encryptedValue,
		oauthToken,
		metadata,
		entity.IsActive,
		entity.ExpiresAt,
		entity.LastUsedAt,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
}

// toDomainList converte lista de entidades para domínio
func (r *GormCredentialRepository) toDomainList(entities []entities.CredentialEntity) []*credential.Credential {
	credentials := make([]*credential.Credential, len(entities))
	for i, entity := range entities {
		credentials[i] = r.toDomain(&entity)
	}
	return credentials
}
