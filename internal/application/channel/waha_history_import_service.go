package channel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/channels/waha"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"github.com/ventros/crm/internal/domain/crm/contact"
	"github.com/ventros/crm/internal/domain/crm/message"
	"go.uber.org/zap"
)

// WahaHistoryImportService orchestrates WAHA history import
type WahaHistoryImportService struct {
	channelRepo   channel.Repository
	contactRepo   contact.Repository
	messageRepo   message.Repository
	historyClient *waha.HistoryClient
	logger        *zap.Logger
}

// NewWahaHistoryImportService creates a new history import service
func NewWahaHistoryImportService(
	channelRepo channel.Repository,
	contactRepo contact.Repository,
	messageRepo message.Repository,
	historyClient *waha.HistoryClient,
	logger *zap.Logger,
) *WahaHistoryImportService {
	return &WahaHistoryImportService{
		channelRepo:   channelRepo,
		contactRepo:   contactRepo,
		messageRepo:   messageRepo,
		historyClient: historyClient,
		logger:        logger,
	}
}

// ImportHistory imports chat history from WAHA
func (s *WahaHistoryImportService) ImportHistory(ctx context.Context, channelID uuid.UUID) error {
	s.logger.Info("Starting history import", zap.String("channel_id", channelID.String()))

	// 1. Get channel
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// 2. Validate channel can import
	if !ch.CanStartHistoryImport() {
		return fmt.Errorf("channel cannot start history import: status=%s, enabled=%v",
			ch.HistoryImportStatus, ch.HistoryImportEnabled)
	}

	// 3. Start import
	if err := ch.StartHistoryImport(); err != nil {
		return fmt.Errorf("failed to start history import: %w", err)
	}

	if err := s.channelRepo.Update(ch); err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	// 4. Get WAHA config
	wahaConfig, err := ch.GetWAHAConfig()
	if err != nil {
		ch.FailHistoryImport("invalid WAHA configuration")
		s.channelRepo.Update(ch)
		return fmt.Errorf("failed to get WAHA config: %w", err)
	}

	// 5. Initialize stats
	stats := channel.HistoryImportStats{
		StartedAt: time.Now(),
	}

	// 6. Fetch all chats
	chats, err := s.historyClient.FetchAllChats(ctx, wahaConfig.SessionID)
	if err != nil {
		ch.FailHistoryImport(fmt.Sprintf("failed to fetch chats: %v", err))
		s.channelRepo.Update(ch)
		return fmt.Errorf("failed to fetch chats: %w", err)
	}

	s.logger.Info("Fetched chats",
		zap.String("channel_id", channelID.String()),
		zap.Int("total_chats", len(chats)))

	// 7. Get import configuration
	fromTimestamp := ch.GetHistoryImportFromDate()
	maxMessagesPerChat := ch.GetHistoryImportMaxMessages()

	// 8. Process each chat
	for _, chat := range chats {
		// Check context cancellation
		select {
		case <-ctx.Done():
			ch.FailHistoryImport("import cancelled by context")
			s.channelRepo.Update(ch)
			return ctx.Err()
		default:
		}

		// Fetch messages for this chat
		messages, err := s.historyClient.FetchAllMessages(
			ctx,
			wahaConfig.SessionID,
			chat.ID,
			fromTimestamp,
			maxMessagesPerChat,
		)

		if err != nil {
			s.logger.Error("Failed to fetch messages for chat",
				zap.String("chat_id", chat.ID),
				zap.Error(err))
			stats.Failed++
			continue
		}

		s.logger.Debug("Processing chat messages",
			zap.String("chat_id", chat.ID),
			zap.String("chat_name", chat.Name),
			zap.Int("message_count", len(messages)))

		stats.Total += len(messages)

		// Process each message
		for _, wahaMsg := range messages {
			if err := s.processMessage(ctx, ch, wahaMsg); err != nil {
				s.logger.Error("Failed to process message",
					zap.String("message_id", wahaMsg.ID),
					zap.Error(err))
				stats.Failed++
				continue
			}
			stats.Processed++
		}
	}

	// 9. Complete import
	ch.CompleteHistoryImport(stats)
	if err := s.channelRepo.Update(ch); err != nil {
		return fmt.Errorf("failed to update channel after import: %w", err)
	}

	s.logger.Info("History import completed",
		zap.String("channel_id", channelID.String()),
		zap.Int("total", stats.Total),
		zap.Int("processed", stats.Processed),
		zap.Int("failed", stats.Failed))

	return nil
}

// processMessage processes a single WAHA message
func (s *WahaHistoryImportService) processMessage(
	ctx context.Context,
	ch *channel.Channel,
	wahaMsg waha.HistoryMessage,
) error {
	// 1. Check if message already exists (deduplication by channel message ID)
	existing, err := s.messageRepo.FindByChannelMessageID(ctx, wahaMsg.ID)
	if err == nil && existing != nil {
		s.logger.Debug("Message already exists, skipping",
			zap.String("message_id", wahaMsg.ID))
		return nil // Skip duplicate
	}

	// 2. Get or create contact
	contactEntity, err := s.getOrCreateContact(ctx, ch, wahaMsg)
	if err != nil {
		return fmt.Errorf("failed to get/create contact: %w", err)
	}

	// 3. Parse tenant ID (assuming it's a UUID string)
	tenantUUID, err := uuid.Parse(ch.TenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	// 4. Map message type
	msgType := s.mapMessageType(wahaMsg.Type)

	// 5. Determine fromMe
	fromMe := wahaMsg.FromMe

	// 6. Create message
	msg, err := message.NewMessage(
		contactEntity.ID(),
		ch.ProjectID,
		tenantUUID,
		msgType,
		fromMe,
	)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// 7. Set text content
	if wahaMsg.Body != "" {
		if err := msg.SetText(wahaMsg.Body); err != nil {
			s.logger.Warn("Failed to set text", zap.Error(err))
		}
	}

	// 8. Set media content
	if wahaMsg.MediaURL != "" {
		if err := msg.SetMediaContent(wahaMsg.MediaURL, wahaMsg.MimeType); err != nil {
			s.logger.Warn("Failed to set media content", zap.Error(err))
		}
	}

	// 9. Assign to channel
	msg.AssignToChannel(ch.ID, nil)

	// 10. Set channel message ID (external ID from WAHA)
	msg.SetChannelMessageID(wahaMsg.ID)

	// 11. Update status based on ack
	s.updateMessageStatus(msg, wahaMsg.Ack)

	// 12. Save message
	if err := s.messageRepo.Save(ctx, msg); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	s.logger.Debug("Message imported",
		zap.String("message_id", wahaMsg.ID),
		zap.String("contact_id", contactEntity.ID().String()),
		zap.Bool("from_me", fromMe),
		zap.String("type", string(msgType)))

	// 9. Trigger enrichment for media messages (via domain events)
	// The message aggregate will publish enrichment events if needed
	// This will be handled by the message enrichment consumer

	return nil
}

// getOrCreateContact gets an existing contact or creates a new one
func (s *WahaHistoryImportService) getOrCreateContact(
	ctx context.Context,
	ch *channel.Channel,
	wahaMsg waha.HistoryMessage,
) (*contact.Contact, error) {
	// Extract phone from chat ID (format: "5511999999999@c.us" or "5511999999999@s.whatsapp.net")
	phone := extractPhoneFromChatID(wahaMsg.ChatID)
	if phone == "" {
		return nil, fmt.Errorf("invalid chat ID format: %s", wahaMsg.ChatID)
	}

	// Try to find existing contact by phone
	existing, err := s.contactRepo.FindByPhone(ctx, ch.ProjectID, phone)
	if err == nil && existing != nil {
		s.logger.Debug("Contact found",
			zap.String("contact_id", existing.ID().String()),
			zap.String("phone", phone))
		return existing, nil
	}

	// Create new contact
	newContact, err := contact.NewContact(
		ch.ProjectID,
		ch.TenantID,
		phone, // name (will be updated later with actual name)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	// Set phone
	if err := newContact.SetPhone(phone); err != nil {
		s.logger.Warn("Failed to set phone", zap.Error(err))
	}

	// Save contact
	if err := s.contactRepo.Save(ctx, newContact); err != nil {
		return nil, fmt.Errorf("failed to save contact: %w", err)
	}

	s.logger.Info("Contact created from history import",
		zap.String("contact_id", newContact.ID().String()),
		zap.String("phone", phone),
		zap.String("channel_id", ch.ID.String()))

	return newContact, nil
}

// updateMessageStatus updates message status based on WAHA ack
func (s *WahaHistoryImportService) updateMessageStatus(msg *message.Message, ack int) {
	switch ack {
	case 0:
		// Pending - default status
	case 1:
		// Sent - default status
	case 2:
		msg.MarkAsDelivered()
	case 3:
		msg.MarkAsRead()
	case 4:
		msg.MarkAsPlayed() // Only for voice messages
	}
}

// mapMessageType maps WAHA message type to our message content type
func (s *WahaHistoryImportService) mapMessageType(wahaType string) message.ContentType {
	switch wahaType {
	case "chat", "text":
		return message.ContentTypeText
	case "image":
		return message.ContentTypeImage
	case "video":
		return message.ContentTypeVideo
	case "audio", "ptt":
		return message.ContentTypeAudio
	case "document":
		return message.ContentTypeDocument
	case "location":
		return message.ContentTypeLocation
	case "vcard", "contact":
		return message.ContentTypeContact
	default:
		return message.ContentTypeText
	}
}

// extractPhoneFromChatID extracts phone number from WAHA chat ID
// Formats supported:
// - "5511999999999@c.us" (individual chat)
// - "5511999999999@s.whatsapp.net" (individual chat)
// - "5511999999999-1234567890@g.us" (group chat - extracts first number)
func extractPhoneFromChatID(chatID string) string {
	// Remove @ suffix
	parts := strings.Split(chatID, "@")
	if len(parts) == 0 {
		return ""
	}

	// Get the number part
	numberPart := parts[0]

	// For groups, extract the first number before the dash
	if strings.Contains(numberPart, "-") {
		groupParts := strings.Split(numberPart, "-")
		if len(groupParts) > 0 {
			return groupParts[0]
		}
	}

	return numberPart
}
