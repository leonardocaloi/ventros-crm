package tracking

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/crm/tracking"
	"go.uber.org/zap"
)

// GetTrackingUseCase busca um tracking por ID
type GetTrackingUseCase struct {
	repo   tracking.Repository
	logger *zap.Logger
}

// NewGetTrackingUseCase cria uma nova instância do use case
func NewGetTrackingUseCase(repo tracking.Repository, logger *zap.Logger) *GetTrackingUseCase {
	return &GetTrackingUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute busca tracking por ID
func (uc *GetTrackingUseCase) Execute(ctx context.Context, id uuid.UUID) (*TrackingDTO, error) {
	t, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == tracking.ErrTrackingNotFound {
			return nil, err
		}
		uc.logger.Error("Failed to fetch tracking",
			zap.Error(err),
			zap.String("tracking_id", id.String()),
		)
		return nil, fmt.Errorf("failed to fetch tracking: %w", err)
	}

	result := ToDTO(t)
	return &result, nil
}
