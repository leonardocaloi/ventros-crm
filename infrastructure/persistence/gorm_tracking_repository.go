package persistence

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/crm/tracking"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormTrackingRepository implementa tracking.Repository usando GORM
type GormTrackingRepository struct {
	db *gorm.DB
}

// NewGormTrackingRepository cria uma nova instância do repository
func NewGormTrackingRepository(db *gorm.DB) *GormTrackingRepository {
	return &GormTrackingRepository{db: db}
}

// Create persiste um novo tracking
func (r *GormTrackingRepository) Create(ctx context.Context, t *tracking.Tracking) error {
	entity := r.toEntity(t)
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("failed to create tracking: %w", err)
	}
	return nil
}

// FindByID busca tracking por ID
func (r *GormTrackingRepository) FindByID(ctx context.Context, id uuid.UUID) (*tracking.Tracking, error) {
	var entity entities.TrackingEntity
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, tracking.ErrTrackingNotFound
		}
		return nil, fmt.Errorf("failed to find tracking: %w", err)
	}
	return r.toDomain(&entity)
}

// FindByContactID busca todos os trackings de um contato
func (r *GormTrackingRepository) FindByContactID(ctx context.Context, contactID uuid.UUID) ([]*tracking.Tracking, error) {
	var entities []entities.TrackingEntity
	if err := r.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		Order("created_at DESC").
		Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find trackings by contact: %w", err)
	}
	return r.toDomainList(entities)
}

// FindBySessionID busca todos os trackings de uma sessão
func (r *GormTrackingRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*tracking.Tracking, error) {
	var entities []entities.TrackingEntity
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find trackings by session: %w", err)
	}
	return r.toDomainList(entities)
}

// FindByProjectID busca todos os trackings de um projeto
func (r *GormTrackingRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*tracking.Tracking, error) {
	var entities []entities.TrackingEntity
	query := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find trackings by project: %w", err)
	}
	return r.toDomainList(entities)
}

// FindBySource busca trackings por fonte
func (r *GormTrackingRepository) FindBySource(ctx context.Context, projectID uuid.UUID, source tracking.Source, limit, offset int) ([]*tracking.Tracking, error) {
	var entities []entities.TrackingEntity
	query := r.db.WithContext(ctx).
		Where("project_id = ? AND source = ?", projectID, string(source)).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find trackings by source: %w", err)
	}
	return r.toDomainList(entities)
}

// FindByCampaign busca trackings por campanha
func (r *GormTrackingRepository) FindByCampaign(ctx context.Context, projectID uuid.UUID, campaign string, limit, offset int) ([]*tracking.Tracking, error) {
	var entities []entities.TrackingEntity
	query := r.db.WithContext(ctx).
		Where("project_id = ? AND campaign = ?", projectID, campaign).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find trackings by campaign: %w", err)
	}
	return r.toDomainList(entities)
}

// FindByClickID busca tracking por click ID (único)
func (r *GormTrackingRepository) FindByClickID(ctx context.Context, clickID string) (*tracking.Tracking, error) {
	var entity entities.TrackingEntity
	if err := r.db.WithContext(ctx).Where("click_id = ?", clickID).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, tracking.ErrTrackingNotFound
		}
		return nil, fmt.Errorf("failed to find tracking by click_id: %w", err)
	}
	return r.toDomain(&entity)
}

// Update atualiza um tracking existente
func (r *GormTrackingRepository) Update(ctx context.Context, t *tracking.Tracking) error {
	entity := r.toEntity(t)
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("failed to update tracking: %w", err)
	}
	return nil
}

// Delete remove um tracking (soft delete)
func (r *GormTrackingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.TrackingEntity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete tracking: %w", err)
	}
	return nil
}

// toEntity converte domain tracking para entity
func (r *GormTrackingRepository) toEntity(t *tracking.Tracking) *entities.TrackingEntity {
	return &entities.TrackingEntity{
		ID:             t.ID(),
		ContactID:      t.ContactID(),
		SessionID:      t.SessionID(),
		TenantID:       t.TenantID(),
		ProjectID:      t.ProjectID(),
		Source:         string(t.Source()),
		Platform:       string(t.Platform()),
		Campaign:       t.Campaign(),
		AdID:           t.AdID(),
		AdURL:          t.AdURL(),
		ClickID:        t.ClickID(),
		ConversionData: t.ConversionData(),
		UTMSource:      t.UTMSource(),
		UTMMedium:      t.UTMMedium(),
		UTMCampaign:    t.UTMCampaign(),
		UTMTerm:        t.UTMTerm(),
		UTMContent:     t.UTMContent(),
		Metadata:       t.Metadata(),
		CreatedAt:      t.CreatedAt(),
		UpdatedAt:      t.UpdatedAt(),
	}
}

// toDomain converte entity para domain tracking
func (r *GormTrackingRepository) toDomain(entity *entities.TrackingEntity) (*tracking.Tracking, error) {
	t := tracking.ReconstructTracking(
		entity.ID,
		entity.ContactID,
		entity.SessionID,
		entity.TenantID,
		entity.ProjectID,
		tracking.Source(entity.Source),
		tracking.Platform(entity.Platform),
		entity.Campaign,
		entity.AdID,
		entity.AdURL,
		entity.ClickID,
		entity.ConversionData,
		entity.UTMSource,
		entity.UTMMedium,
		entity.UTMCampaign,
		entity.UTMTerm,
		entity.UTMContent,
		entity.Metadata,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
	return t, nil
}

// toDomainList converte lista de entities para domain
func (r *GormTrackingRepository) toDomainList(entities []entities.TrackingEntity) ([]*tracking.Tracking, error) {
	trackings := make([]*tracking.Tracking, 0, len(entities))
	for _, entity := range entities {
		t, err := r.toDomain(&entity)
		if err != nil {
			return nil, err
		}
		trackings = append(trackings, t)
	}
	return trackings, nil
}
