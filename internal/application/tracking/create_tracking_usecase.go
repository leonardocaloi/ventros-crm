package tracking

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/tracking"
	"go.uber.org/zap"
)

// EventBus interface for publishing domain events
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
}

// TransactionManager gerencia transações de banco de dados.
type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CreateTrackingUseCase cria um novo tracking de conversão
type CreateTrackingUseCase struct {
	repo      tracking.Repository
	eventBus  EventBus
	logger    *zap.Logger
	txManager TransactionManager
}

// NewCreateTrackingUseCase cria uma nova instância do use case
func NewCreateTrackingUseCase(repo tracking.Repository, eventBus EventBus, logger *zap.Logger, txManager TransactionManager) *CreateTrackingUseCase {
	return &CreateTrackingUseCase{
		repo:      repo,
		eventBus:  eventBus,
		logger:    logger,
		txManager: txManager,
	}
}

// Execute cria um novo tracking
func (uc *CreateTrackingUseCase) Execute(ctx context.Context, dto CreateTrackingDTO) (*TrackingDTO, error) {
	// Valida source
	source := tracking.Source(dto.Source)
	if !isValidSource(source) {
		return nil, fmt.Errorf("invalid source: %s", dto.Source)
	}

	// Valida platform
	platform := tracking.Platform(dto.Platform)
	if !isValidPlatform(platform) {
		return nil, fmt.Errorf("invalid platform: %s", dto.Platform)
	}

	// Cria o aggregate
	t, err := tracking.NewTracking(
		dto.ContactID,
		dto.SessionID,
		dto.TenantID,
		dto.ProjectID,
		source,
		platform,
	)
	if err != nil {
		uc.logger.Error("Failed to create tracking aggregate",
			zap.Error(err),
			zap.String("contact_id", dto.ContactID.String()),
		)
		return nil, fmt.Errorf("failed to create tracking: %w", err)
	}

	// Define dados adicionais
	if dto.Campaign != "" {
		t.SetCampaign(dto.Campaign)
	}
	if dto.AdID != "" || dto.AdURL != "" {
		t.SetAdInfo(dto.AdID, dto.AdURL)
	}
	if dto.ClickID != "" {
		t.SetClickID(dto.ClickID)
	}
	if dto.ConversionData != "" {
		t.SetConversionData(dto.ConversionData)
	}
	if dto.UTMSource != "" || dto.UTMMedium != "" || dto.UTMCampaign != "" {
		t.SetUTMParameters(dto.UTMSource, dto.UTMMedium, dto.UTMCampaign, dto.UTMTerm, dto.UTMContent)
	}
	if dto.Metadata != nil && len(dto.Metadata) > 0 {
		t.SetMetadata(dto.Metadata)
	}

	// ✅ TRANSAÇÃO ATÔMICA: Save + Publish juntos
	err = uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Persiste (usa transação do contexto)
		if err := uc.repo.Create(txCtx, t); err != nil {
			uc.logger.Error("Failed to persist tracking",
				zap.Error(err),
				zap.String("tracking_id", t.ID().String()),
			)
			return fmt.Errorf("failed to save tracking: %w", err)
		}

		// Publica eventos de domínio (usa mesma transação)
		events := t.DomainEvents()
		for _, event := range events {
			if err := uc.eventBus.Publish(txCtx, event); err != nil {
				uc.logger.Error("Failed to publish domain event",
					zap.Error(err),
					zap.String("tracking_id", t.ID().String()),
				)
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	t.ClearEvents()

	uc.logger.Info("Tracking created",
		zap.String("tracking_id", t.ID().String()),
		zap.String("contact_id", dto.ContactID.String()),
		zap.String("source", dto.Source),
	)

	result := ToDTO(t)
	return &result, nil
}

func isValidSource(source tracking.Source) bool {
	validSources := []tracking.Source{
		tracking.SourceMetaAds,
		tracking.SourceGoogleAds,
		tracking.SourceTikTokAds,
		tracking.SourceLinkedIn,
		tracking.SourceOrganic,
		tracking.SourceDirect,
		tracking.SourceReferral,
		tracking.SourceOther,
	}
	for _, s := range validSources {
		if source == s {
			return true
		}
	}
	return false
}

func isValidPlatform(platform tracking.Platform) bool {
	validPlatforms := []tracking.Platform{
		tracking.PlatformInstagram,
		tracking.PlatformFacebook,
		tracking.PlatformGoogle,
		tracking.PlatformTikTok,
		tracking.PlatformLinkedIn,
		tracking.PlatformWhatsApp,
		tracking.PlatformOther,
	}
	for _, p := range validPlatforms {
		if platform == p {
			return true
		}
	}
	return false
}
