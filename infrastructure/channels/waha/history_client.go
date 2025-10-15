package waha

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// HistoryClient provides methods for fetching chat history from WAHA
type HistoryClient struct {
	baseClient *WAHAClient
	logger     *zap.Logger
}

// NewHistoryClient creates a new history client
func NewHistoryClient(baseClient *WAHAClient, logger *zap.Logger) *HistoryClient {
	return &HistoryClient{
		baseClient: baseClient,
		logger:     logger,
	}
}

// Chat represents a chat in WAHA
type Chat struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Picture     string          `json:"picture,omitempty"`
	LastMessage *MessagePayload `json:"lastMessage,omitempty"`
}

// HistoryMessage represents a message from history
type HistoryMessage struct {
	ID        string `json:"id"`
	ChatID    string `json:"chatId"`
	From      string `json:"from"`
	FromMe    bool   `json:"fromMe"`
	To        string `json:"to"`
	Body      string `json:"body"`
	Type      string `json:"type"` // "chat", "image", "video", "audio", "document"
	MimeType  string `json:"mimeType,omitempty"`
	MediaURL  string `json:"mediaUrl,omitempty"`
	HasMedia  bool   `json:"hasMedia"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp
	Ack       int    `json:"ack"`       // 0=pending, 1=sent, 2=delivered, 3=read, 4=played
	AckName   string `json:"ackName,omitempty"`
	ReplyTo   string `json:"replyTo,omitempty"`
}

// FetchChatsRequest represents a request to fetch chats
type FetchChatsRequest struct {
	SessionID string
	Limit     int
	Offset    int
	SortBy    string // "messageTimestamp"
	SortOrder string // "desc" or "asc"
}

// FetchMessagesRequest represents a request to fetch messages
type FetchMessagesRequest struct {
	SessionID     string
	ChatID        string
	Limit         int
	Offset        int
	DownloadMedia bool
	FromTimestamp *time.Time // Filter messages after this timestamp
	ToTimestamp   *time.Time // Filter messages before this timestamp
	FilterFromMe  *bool      // Filter messages sent by me
	FilterAck     *string    // Filter by ack status
}

// FetchChatsResponse represents the response from fetching chats
type FetchChatsResponse struct {
	Chats   []Chat
	HasMore bool
	Total   int
}

// FetchMessagesResponse represents the response from fetching messages
type FetchMessagesResponse struct {
	Messages []HistoryMessage
	HasMore  bool
	Total    int
}

// FetchChats fetches chats from WAHA with pagination
func (c *HistoryClient) FetchChats(ctx context.Context, req FetchChatsRequest) (*FetchChatsResponse, error) {
	// Default values
	if req.Limit == 0 || req.Limit > 100 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "messageTimestamp"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Build URL
	endpoint := fmt.Sprintf("%s/api/%s/chats", c.baseClient.baseURL, req.SessionID)
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("limit", fmt.Sprintf("%d", req.Limit))
	q.Set("offset", fmt.Sprintf("%d", req.Offset))
	q.Set("sortBy", req.SortBy)
	q.Set("sortOrder", req.SortOrder)
	u.RawQuery = q.Encode()

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.baseClient.setAuthHeaders(httpReq)

	// Execute request
	resp, err := c.baseClient.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chats []Chat
	if err := json.NewDecoder(resp.Body).Decode(&chats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Fetched chats from WAHA",
		zap.String("session_id", req.SessionID),
		zap.Int("count", len(chats)),
		zap.Int("limit", req.Limit),
		zap.Int("offset", req.Offset))

	// Determine if there are more chats
	hasMore := len(chats) == req.Limit

	return &FetchChatsResponse{
		Chats:   chats,
		HasMore: hasMore,
		Total:   len(chats), // WAHA doesn't return total count
	}, nil
}

// FetchMessages fetches messages from a specific chat with pagination and filters
func (c *HistoryClient) FetchMessages(ctx context.Context, req FetchMessagesRequest) (*FetchMessagesResponse, error) {
	// Default values
	if req.Limit == 0 || req.Limit > 100 {
		req.Limit = 100
	}

	// Build URL
	endpoint := fmt.Sprintf("%s/api/%s/chats/%s/messages", c.baseClient.baseURL, req.SessionID, req.ChatID)
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("limit", fmt.Sprintf("%d", req.Limit))
	q.Set("offset", fmt.Sprintf("%d", req.Offset))
	q.Set("downloadMedia", fmt.Sprintf("%t", req.DownloadMedia))

	// Add timestamp filters
	if req.FromTimestamp != nil {
		q.Set("filter.timestamp.gte", fmt.Sprintf("%d", req.FromTimestamp.Unix()))
	}
	if req.ToTimestamp != nil {
		q.Set("filter.timestamp.lte", fmt.Sprintf("%d", req.ToTimestamp.Unix()))
	}

	// Add fromMe filter
	if req.FilterFromMe != nil {
		q.Set("filter.fromMe", fmt.Sprintf("%t", *req.FilterFromMe))
	}

	// Add ack filter
	if req.FilterAck != nil {
		q.Set("filter.ack", *req.FilterAck)
	}

	u.RawQuery = q.Encode()

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.baseClient.setAuthHeaders(httpReq)

	// Execute request
	resp, err := c.baseClient.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var messages []HistoryMessage
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Fetched messages from WAHA",
		zap.String("session_id", req.SessionID),
		zap.String("chat_id", req.ChatID),
		zap.Int("count", len(messages)),
		zap.Int("limit", req.Limit),
		zap.Int("offset", req.Offset))

	// Determine if there are more messages
	hasMore := len(messages) == req.Limit

	return &FetchMessagesResponse{
		Messages: messages,
		HasMore:  hasMore,
		Total:    len(messages), // WAHA doesn't return total count
	}, nil
}

// FetchAllMessages fetches all messages from a chat with automatic pagination
// maxMessages: 0 = ilimitado, >0 = limite máximo de mensagens
func (c *HistoryClient) FetchAllMessages(
	ctx context.Context,
	sessionID, chatID string,
	fromTimestamp *time.Time,
	maxMessages int,
) ([]HistoryMessage, error) {
	var allMessages []HistoryMessage
	offset := 0
	limit := 100

	c.logger.Info("Starting to fetch all messages",
		zap.String("session_id", sessionID),
		zap.String("chat_id", chatID),
		zap.Int("max_messages", maxMessages))

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Check if we reached the limit
		if maxMessages > 0 && len(allMessages) >= maxMessages {
			c.logger.Info("Reached max messages limit",
				zap.Int("max_messages", maxMessages),
				zap.Int("total_fetched", len(allMessages)))
			break
		}

		// Adjust limit if we're close to the max
		batchLimit := limit
		if maxMessages > 0 {
			remaining := maxMessages - len(allMessages)
			if remaining < limit {
				batchLimit = remaining
			}
		}

		// Fetch batch of messages
		resp, err := c.FetchMessages(ctx, FetchMessagesRequest{
			SessionID:     sessionID,
			ChatID:        chatID,
			Limit:         batchLimit,
			Offset:        offset,
			DownloadMedia: false, // Don't download media during history import
			FromTimestamp: fromTimestamp,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to fetch messages at offset %d: %w", offset, err)
		}

		allMessages = append(allMessages, resp.Messages...)

		c.logger.Debug("Fetched message batch",
			zap.Int("batch_size", len(resp.Messages)),
			zap.Int("total_so_far", len(allMessages)),
			zap.Int("offset", offset))

		// Check if there are more messages
		if !resp.HasMore {
			break
		}

		offset += limit

		// Rate limiting: sleep 500ms between requests to avoid overwhelming WAHA
		time.Sleep(500 * time.Millisecond)
	}

	c.logger.Info("Finished fetching all messages",
		zap.String("session_id", sessionID),
		zap.String("chat_id", chatID),
		zap.Int("total_messages", len(allMessages)),
		zap.Int("max_messages", maxMessages))

	return allMessages, nil
}

// FetchAllChats fetches all chats with automatic pagination
func (c *HistoryClient) FetchAllChats(ctx context.Context, sessionID string) ([]Chat, error) {
	var allChats []Chat
	offset := 0
	limit := 100

	c.logger.Info("Starting to fetch all chats", zap.String("session_id", sessionID))

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Fetch batch of chats
		resp, err := c.FetchChats(ctx, FetchChatsRequest{
			SessionID: sessionID,
			Limit:     limit,
			Offset:    offset,
			SortBy:    "messageTimestamp",
			SortOrder: "desc",
		})

		if err != nil {
			return nil, fmt.Errorf("failed to fetch chats at offset %d: %w", offset, err)
		}

		allChats = append(allChats, resp.Chats...)

		c.logger.Debug("Fetched chat batch",
			zap.Int("batch_size", len(resp.Chats)),
			zap.Int("total_so_far", len(allChats)),
			zap.Int("offset", offset))

		// Check if there are more chats
		if !resp.HasMore {
			break
		}

		offset += limit

		// Rate limiting: sleep 500ms between requests
		time.Sleep(500 * time.Millisecond)
	}

	c.logger.Info("Finished fetching all chats",
		zap.String("session_id", sessionID),
		zap.Int("total_chats", len(allChats)))

	return allChats, nil
}

// GetOldestAvailableDate busca a data da mensagem mais antiga disponível em um chat
// Isso é usado para otimizar importações: se o usuário pedir 180 dias mas o chat
// só tem mensagens dos últimos 30 dias, retornamos 30 dias atrás
// Retorna nil se o chat não tiver mensagens
func (c *HistoryClient) GetOldestAvailableDate(ctx context.Context, sessionID, chatID string) (*time.Time, error) {
	// Busca as últimas 100 mensagens ordenadas DESC (mais recentes primeiro)
	// A API WAHA sempre retorna mensagens ordenadas por timestamp DESC
	resp, err := c.FetchMessages(ctx, FetchMessagesRequest{
		SessionID:     sessionID,
		ChatID:        chatID,
		Limit:         100,
		Offset:        0,
		DownloadMedia: false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	if len(resp.Messages) == 0 {
		// Chat sem mensagens
		return nil, nil
	}

	// Se temos menos de 100 mensagens, a última mensagem é a mais antiga
	if len(resp.Messages) < 100 {
		oldestMsg := resp.Messages[len(resp.Messages)-1]
		oldestDate := time.Unix(oldestMsg.Timestamp, 0)

		c.logger.Info("Found oldest message date (all messages fit in one page)",
			zap.String("session_id", sessionID),
			zap.String("chat_id", chatID),
			zap.Time("oldest_date", oldestDate),
			zap.Int("total_messages", len(resp.Messages)))

		return &oldestDate, nil
	}

	// Se temos exatamente 100 mensagens, precisamos buscar mais páginas
	// até encontrar a última mensagem
	offset := 100
	var oldestMsg *HistoryMessage

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := c.FetchMessages(ctx, FetchMessagesRequest{
			SessionID:     sessionID,
			ChatID:        chatID,
			Limit:         100,
			Offset:        offset,
			DownloadMedia: false,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to fetch messages at offset %d: %w", offset, err)
		}

		if len(resp.Messages) == 0 {
			// Não há mais mensagens, usamos a última da página anterior
			break
		}

		oldestMsg = &resp.Messages[len(resp.Messages)-1]

		if !resp.HasMore {
			// Esta é a última página
			break
		}

		offset += 100
		time.Sleep(200 * time.Millisecond) // Rate limiting
	}

	if oldestMsg != nil {
		oldestDate := time.Unix(oldestMsg.Timestamp, 0)

		c.logger.Info("Found oldest message date (multiple pages)",
			zap.String("session_id", sessionID),
			zap.String("chat_id", chatID),
			zap.Time("oldest_date", oldestDate),
			zap.Int("total_pages", offset/100))

		return &oldestDate, nil
	}

	return nil, nil
}
