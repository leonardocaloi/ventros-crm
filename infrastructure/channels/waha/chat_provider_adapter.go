package waha

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ventros/crm/internal/domain/crm/channel"
	"go.uber.org/zap"
)

// ChatProviderAdapter implements domain.ChatProvider for WAHA
//
// This adapter manages chat operations.
// Used by both channel types:
// - TypeWAHA (manual mode)
// - TypeWhatsAppBusiness (auto mode)
type ChatProviderAdapter struct {
	client      *Client
	sessionName string
	logger      *zap.Logger
}

// NewChatProviderAdapter creates a new chat provider adapter
func NewChatProviderAdapter(client *Client, sessionName string, logger *zap.Logger) *ChatProviderAdapter {
	return &ChatProviderAdapter{
		client:      client,
		sessionName: sessionName,
		logger:      logger,
	}
}

// GetChatsOverview returns overview of all chats
//
// Includes: chat id, name, picture, last message
// Sorted by last message timestamp
func (a *ChatProviderAdapter) GetChatsOverview(ctx context.Context, limit, offset int, chatIDs []string) ([]channel.ChatOverview, error) {
	a.logger.Info("Getting chats overview",
		zap.String("session_name", a.sessionName),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.Strings("chat_ids", chatIDs))

	path := fmt.Sprintf("/api/%s/chats", a.sessionName)

	// Build query parameters
	params := []string{}
	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}
	if offset > 0 {
		params = append(params, fmt.Sprintf("offset=%d", offset))
	}
	if len(chatIDs) > 0 {
		for _, chatID := range chatIDs {
			params = append(params, fmt.Sprintf("chatIds=%s", chatID))
		}
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		a.logger.Error("Failed to get chats overview",
			zap.String("session_name", a.sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get chats overview: %w", err)
	}

	var wahaChats []WAHAChatOverview
	if err := a.client.ParseResponse(resp, &wahaChats); err != nil {
		return nil, fmt.Errorf("failed to parse chats response: %w", err)
	}

	// Convert WAHA chats to domain chats
	chats := make([]channel.ChatOverview, len(wahaChats))
	for i, wc := range wahaChats {
		chats[i] = mapWAHAChatOverviewToDomain(wc)
	}

	a.logger.Info("Successfully retrieved chats overview",
		zap.String("session_name", a.sessionName),
		zap.Int("count", len(chats)))

	return chats, nil
}

// GetChatMessages returns messages from a specific chat
func (a *ChatProviderAdapter) GetChatMessages(ctx context.Context, chatID string, opts channel.ChatMessagesOptions) ([]channel.ChatMessage, error) {
	a.logger.Info("Getting chat messages",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID),
		zap.Int("limit", opts.Limit),
		zap.Int("offset", opts.Offset))

	path := fmt.Sprintf("/api/%s/chats/%s/messages", a.sessionName, chatID)

	// Build query parameters
	params := []string{}
	if opts.DownloadMedia {
		params = append(params, "downloadMedia=true")
	}
	if opts.Limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", opts.Limit))
	}
	if opts.Offset > 0 {
		params = append(params, fmt.Sprintf("offset=%d", opts.Offset))
	}
	if opts.TimestampLte != nil {
		params = append(params, fmt.Sprintf("timestamp_lte=%d", *opts.TimestampLte))
	}
	if opts.TimestampGte != nil {
		params = append(params, fmt.Sprintf("timestamp_gte=%d", *opts.TimestampGte))
	}
	if opts.FromMe != nil {
		params = append(params, fmt.Sprintf("fromMe=%t", *opts.FromMe))
	}
	if opts.AckStatus != nil {
		params = append(params, fmt.Sprintf("ackStatus=%s", *opts.AckStatus))
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		a.logger.Error("Failed to get chat messages",
			zap.String("session_name", a.sessionName),
			zap.String("chat_id", chatID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}

	var wahaMessages []WAHAChatMessage
	if err := a.client.ParseResponse(resp, &wahaMessages); err != nil {
		return nil, fmt.Errorf("failed to parse messages response: %w", err)
	}

	// Convert WAHA messages to domain messages
	messages := make([]channel.ChatMessage, len(wahaMessages))
	for i, wm := range wahaMessages {
		messages[i] = mapWAHAChatMessageToDomain(wm)
	}

	a.logger.Info("Successfully retrieved chat messages",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID),
		zap.Int("count", len(messages)))

	return messages, nil
}

// DeleteChat deletes a chat
func (a *ChatProviderAdapter) DeleteChat(ctx context.Context, chatID string) error {
	a.logger.Info("Deleting chat",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID))

	path := fmt.Sprintf("/api/%s/chats/%s", a.sessionName, chatID)

	resp, err := a.client.Delete(ctx, path)
	if err != nil {
		a.logger.Error("Failed to delete chat",
			zap.String("session_name", a.sessionName),
			zap.String("chat_id", chatID),
			zap.Error(err))
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Chat deleted successfully",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID))

	return nil
}

// ArchiveChat archives a chat
func (a *ChatProviderAdapter) ArchiveChat(ctx context.Context, chatID string) error {
	a.logger.Info("Archiving chat",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID))

	path := fmt.Sprintf("/api/%s/chats/%s/archive", a.sessionName, chatID)

	resp, err := a.client.Post(ctx, path, nil)
	if err != nil {
		a.logger.Error("Failed to archive chat",
			zap.String("session_name", a.sessionName),
			zap.String("chat_id", chatID),
			zap.Error(err))
		return fmt.Errorf("failed to archive chat: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Chat archived successfully",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID))

	return nil
}

// UnarchiveChat unarchives a chat
func (a *ChatProviderAdapter) UnarchiveChat(ctx context.Context, chatID string) error {
	a.logger.Info("Unarchiving chat",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID))

	path := fmt.Sprintf("/api/%s/chats/%s/archive", a.sessionName, chatID)

	resp, err := a.client.Delete(ctx, path)
	if err != nil {
		a.logger.Error("Failed to unarchive chat",
			zap.String("session_name", a.sessionName),
			zap.String("chat_id", chatID),
			zap.Error(err))
		return fmt.Errorf("failed to unarchive chat: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Chat unarchived successfully",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID))

	return nil
}

// MarkChatAsUnread marks chat as unread
func (a *ChatProviderAdapter) MarkChatAsUnread(ctx context.Context, chatID string) error {
	a.logger.Info("Marking chat as unread",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID))

	path := fmt.Sprintf("/api/%s/chats/%s/unread", a.sessionName, chatID)

	resp, err := a.client.Post(ctx, path, nil)
	if err != nil {
		a.logger.Error("Failed to mark chat as unread",
			zap.String("session_name", a.sessionName),
			zap.String("chat_id", chatID),
			zap.Error(err))
		return fmt.Errorf("failed to mark chat as unread: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Chat marked as unread successfully",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID))

	return nil
}

// ReadChatMessages marks messages in chat as read
//
// Parameters:
// - messages: how many messages to read (latest first)
// - days: how many days to read (latest first, default 7)
func (a *ChatProviderAdapter) ReadChatMessages(ctx context.Context, chatID string, messages *int, days *int) ([]string, error) {
	a.logger.Info("Reading chat messages",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID))

	path := fmt.Sprintf("/api/%s/chats/%s/messages/read", a.sessionName, chatID)

	// Build payload
	payload := make(map[string]interface{})
	if messages != nil {
		payload["messages"] = *messages
	}
	if days != nil {
		payload["days"] = *days
	}

	resp, err := a.client.Post(ctx, path, payload)
	if err != nil {
		a.logger.Error("Failed to read chat messages",
			zap.String("session_name", a.sessionName),
			zap.String("chat_id", chatID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to read chat messages: %w", err)
	}

	var result struct {
		MessageIDs []string `json:"message_ids"`
	}
	if err := a.client.ParseResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse read messages response: %w", err)
	}

	a.logger.Info("Chat messages read successfully",
		zap.String("session_name", a.sessionName),
		zap.String("chat_id", chatID),
		zap.Int("count", len(result.MessageIDs)))

	return result.MessageIDs, nil
}

// WAHAChatOverview represents WAHA chat overview response
type WAHAChatOverview struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Picture     *string          `json:"picture,omitempty"`
	LastMessage *WAHAChatMessage `json:"lastMessage,omitempty"`
}

// WAHAChatMessage represents WAHA chat message
type WAHAChatMessage struct {
	ID        string                 `json:"id"`
	Timestamp int64                  `json:"timestamp"`
	From      string                 `json:"from"`
	FromMe    bool                   `json:"fromMe"`
	To        string                 `json:"to"`
	Body      string                 `json:"body"`
	HasMedia  bool                   `json:"hasMedia"`
	MediaURL  *string                `json:"mediaUrl,omitempty"`
	MimeType  *string                `json:"mimeType,omitempty"`
	Ack       int                    `json:"ack"`
	AckName   string                 `json:"ackName"`
	ReplyTo   *string                `json:"replyTo,omitempty"`
	Location  *WAHAMessageLocation   `json:"location,omitempty"`
	VCards    []string               `json:"vCards,omitempty"`
	RawData   map[string]interface{} `json:"_data,omitempty"`
}

// WAHAMessageLocation represents location in a message
type WAHAMessageLocation struct {
	Latitude    string  `json:"latitude"`
	Longitude   string  `json:"longitude"`
	Description *string `json:"description,omitempty"`
}

// mapWAHAChatOverviewToDomain converts WAHA chat to domain chat
func mapWAHAChatOverviewToDomain(wc WAHAChatOverview) channel.ChatOverview {
	overview := channel.ChatOverview{
		ID:      wc.ID,
		Name:    wc.Name,
		Picture: wc.Picture,
	}

	if wc.LastMessage != nil {
		msg := mapWAHAChatMessageToDomain(*wc.LastMessage)
		overview.LastMessage = &msg
	}

	return overview
}

// mapWAHAChatMessageToDomain converts WAHA message to domain message
func mapWAHAChatMessageToDomain(wm WAHAChatMessage) channel.ChatMessage {
	msg := channel.ChatMessage{
		ID:        wm.ID,
		Timestamp: time.Unix(wm.Timestamp/1000, (wm.Timestamp%1000)*1000000),
		From:      wm.From,
		FromMe:    wm.FromMe,
		To:        wm.To,
		Body:      wm.Body,
		HasMedia:  wm.HasMedia,
		MediaURL:  wm.MediaURL,
		MimeType:  wm.MimeType,
		Ack:       wm.Ack,
		AckName:   wm.AckName,
		ReplyTo:   wm.ReplyTo,
		VCards:    wm.VCards,
		RawData:   wm.RawData,
	}

	if wm.Location != nil {
		msg.Location = &channel.MessageLocation{
			Latitude:    wm.Location.Latitude,
			Longitude:   wm.Location.Longitude,
			Description: wm.Location.Description,
		}
	}

	return msg
}
