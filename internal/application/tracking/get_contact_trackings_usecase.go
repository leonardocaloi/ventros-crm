package tracking

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/crm/tracking"
	"go.uber.org/zap"
)

// GetContactTrackingsUseCase busca todos os trackings de um contato
type GetContactTrackingsUseCase struct {
	repo   tracking.Repository
	logger *zap.Logger
}

// NewGetContactTrackingsUseCase cria uma nova instância do use case
func NewGetContactTrackingsUseCase(repo tracking.Repository, logger *zap.Logger) *GetContactTrackingsUseCase {
	return &GetContactTrackingsUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute busca trackings por contact ID
func (uc *GetContactTrackingsUseCase) Execute(ctx context.Context, contactID uuid.UUID) ([]TrackingDTO, error) {
	trackings, err := uc.repo.FindByContactID(ctx, contactID)
	if err != nil {
		uc.logger.Error("Failed to fetch trackings by contact",
			zap.Error(err),
			zap.String("contact_id", contactID.String()),
		)
		return nil, fmt.Errorf("failed to fetch trackings: %w", err)
	}

	return ToDTOList(trackings), nil
}
