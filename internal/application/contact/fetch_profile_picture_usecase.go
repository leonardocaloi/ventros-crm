package contact

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/infrastructure/messaging"
	"github.com/caloi/ventros-crm/internal/domain/contact"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FetchProfilePictureUseCase busca e atualiza a foto de perfil de um contato
type FetchProfilePictureUseCase struct {
	contactRepo contact.Repository
	wahaService *waha.ProfileService
	eventBus    *messaging.DomainEventBus
	logger      *zap.Logger
}

// NewFetchProfilePictureUseCase cria uma nova instância do use case
func NewFetchProfilePictureUseCase(
	contactRepo contact.Repository,
	wahaService *waha.ProfileService,
	eventBus *messaging.DomainEventBus,
	logger *zap.Logger,
) *FetchProfilePictureUseCase {
	return &FetchProfilePictureUseCase{
		contactRepo: contactRepo,
		wahaService: wahaService,
		eventBus:    eventBus,
		logger:      logger,
	}
}

// FetchProfilePictureCommand comando para buscar foto de perfil
type FetchProfilePictureCommand struct {
	ContactID uuid.UUID
	Phone     string
	Session   string
}

// Execute executa o use case
func (uc *FetchProfilePictureUseCase) Execute(ctx context.Context, cmd FetchProfilePictureCommand) error {
	// 1. Buscar contato
	contactEntity, err := uc.contactRepo.FindByID(ctx, cmd.ContactID)
	if err != nil {
		uc.logger.Error("Failed to find contact",
			zap.String("contact_id", cmd.ContactID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to find contact: %w", err)
	}

	// 2. Buscar foto de perfil via WAHA
	uc.logger.Info("Fetching profile picture from WAHA",
		zap.String("contact_id", cmd.ContactID.String()),
		zap.String("phone", cmd.Phone),
		zap.String("session", cmd.Session))

	profilePictureURL, err := uc.wahaService.FetchAndUpdateContactProfilePicture(ctx, cmd.Phone, cmd.Session)
	if err != nil {
		uc.logger.Error("Failed to fetch profile picture from WAHA",
			zap.String("contact_id", cmd.ContactID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to fetch profile picture: %w", err)
	}

	// 3. Se não tem foto, não faz nada
	if profilePictureURL == "" {
		uc.logger.Debug("Contact has no profile picture",
			zap.String("contact_id", cmd.ContactID.String()))
		return nil
	}

	// 4. Atualizar contato com a URL da foto usando o método do domínio
	contactEntity.SetProfilePicture(profilePictureURL)

	if err := uc.contactRepo.Save(ctx, contactEntity); err != nil {
		uc.logger.Error("Failed to update contact with profile picture",
			zap.String("contact_id", cmd.ContactID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to update contact: %w", err)
	}

	// 5. Disparar evento de domínio
	event := contact.ContactProfilePictureUpdatedEvent{
		ContactID:         cmd.ContactID,
		TenantID:          contactEntity.TenantID(),
		ProfilePictureURL: profilePictureURL,
		FetchedAt:         time.Now(),
	}

	if err := uc.eventBus.Publish(ctx, event); err != nil {
		uc.logger.Error("Failed to publish ContactProfilePictureUpdated event",
			zap.String("contact_id", cmd.ContactID.String()),
			zap.Error(err))
		// Não retorna erro aqui, pois o contato já foi atualizado
	}

	uc.logger.Info("Profile picture updated successfully",
		zap.String("contact_id", cmd.ContactID.String()),
		zap.String("url", profilePictureURL))

	return nil
}
