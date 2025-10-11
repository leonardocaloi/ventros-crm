package channel

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/domain/crm/channel"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
	"github.com/caloi/ventros-crm/internal/domain/crm/message"
	"github.com/caloi/ventros-crm/internal/domain/crm/session"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WAHAHistoryImporter gerencia a importação de histórico de mensagens do WAHA
type WAHAHistoryImporter struct {
	logger         *zap.Logger
	wahaClient     *waha.WAHAClient
	channelRepo    channel.Repository
	contactRepo    contact.Repository
	sessionRepo    session.Repository
	messageRepo    message.Repository
	sessionTimeout time.Duration // Timeout para agrupar mensagens em sessões
}

// NewWAHAHistoryImporter cria um novo importador de histórico
func NewWAHAHistoryImporter(
	logger *zap.Logger,
	wahaClient *waha.WAHAClient,
	channelRepo channel.Repository,
	contactRepo contact.Repository,
	sessionRepo session.Repository,
	messageRepo message.Repository,
) *WAHAHistoryImporter {
	return &WAHAHistoryImporter{
		logger:         logger,
		wahaClient:     wahaClient,
		channelRepo:    channelRepo,
		contactRepo:    contactRepo,
		sessionRepo:    sessionRepo,
		messageRepo:    messageRepo,
		sessionTimeout: 30 * time.Minute, // Padrão: 30 minutos de inatividade fecha sessão
	}
}

// ImportHistoryRequest representa uma requisição de importação
type ImportHistoryRequest struct {
	ChannelID uuid.UUID
	Strategy  channel.WAHAImportStrategy
	Limit     int // Limite de mensagens por chat (0 = todas)
}

// ImportHistoryResult representa o resultado da importação
type ImportHistoryResult struct {
	ChatsProcessed   int
	MessagesImported int
	SessionsCreated  int
	ContactsCreated  int
	Errors           []string
	StartedAt        time.Time
	CompletedAt      time.Time
}

// ImportHistory importa histórico de mensagens do WAHA
func (h *WAHAHistoryImporter) ImportHistory(ctx context.Context, req ImportHistoryRequest) (*ImportHistoryResult, error) {
	result := &ImportHistoryResult{
		StartedAt: time.Now(),
		Errors:    []string{},
	}

	// Buscar canal
	ch, err := h.channelRepo.GetByID(req.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	if !ch.IsWAHA() {
		return nil, fmt.Errorf("channel is not WAHA type")
	}

	// Obter configuração WAHA
	wahaConfig, err := ch.GetWAHAConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get WAHA config: %w", err)
	}

	h.logger.Info("Starting WAHA history import",
		zap.String("channel_id", ch.ID.String()),
		zap.String("session_id", wahaConfig.SessionID),
		zap.String("strategy", string(req.Strategy)))

	// Buscar todos os chats
	chats, err := h.getAllChats(ctx, wahaConfig.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}

	h.logger.Info("Found chats to process", zap.Int("count", len(chats)))

	// Processar cada chat
	for _, chat := range chats {
		if err := h.processChat(ctx, ch, chat, req.Limit, result); err != nil {
			errMsg := fmt.Sprintf("Error processing chat %s: %v", chat.ID, err)
			h.logger.Error(errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}
		result.ChatsProcessed++
	}

	// Marcar importação como concluída
	ch.SetWAHAImportCompleted()
	if err := h.channelRepo.Update(ch); err != nil {
		h.logger.Warn("Failed to mark import as completed", zap.Error(err))
	}

	result.CompletedAt = time.Now()

	h.logger.Info("WAHA history import completed",
		zap.String("channel_id", ch.ID.String()),
		zap.Int("chats_processed", result.ChatsProcessed),
		zap.Int("messages_imported", result.MessagesImported),
		zap.Int("sessions_created", result.SessionsCreated),
		zap.Int("contacts_created", result.ContactsCreated),
		zap.Int("errors", len(result.Errors)),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)))

	return result, nil
}

// getAllChats busca todos os chats com paginação
func (h *WAHAHistoryImporter) getAllChats(ctx context.Context, sessionID string) ([]waha.ChatOverview, error) {
	allChats := []waha.ChatOverview{}
	limit := 100
	offset := 0

	for {
		chats, err := h.wahaClient.GetChatsOverview(ctx, sessionID, limit, offset)
		if err != nil {
			return nil, err
		}

		if len(chats) == 0 {
			break
		}

		allChats = append(allChats, chats...)
		offset += len(chats)

		// Se retornou menos que o limite, acabou
		if len(chats) < limit {
			break
		}
	}

	return allChats, nil
}

// processChat processa um chat específico
func (h *WAHAHistoryImporter) processChat(
	ctx context.Context,
	ch *channel.Channel,
	chat waha.ChatOverview,
	limit int,
	result *ImportHistoryResult,
) error {
	wahaConfig, _ := ch.GetWAHAConfig()

	// Buscar mensagens do chat
	if limit == 0 {
		limit = 1000 // Limite padrão para evitar sobrecarga
	}

	messages, err := h.wahaClient.GetChatMessages(ctx, wahaConfig.SessionID, chat.ID, limit, false)
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	if len(messages) == 0 {
		return nil // Chat sem mensagens
	}

	h.logger.Debug("Processing chat",
		zap.String("chat_id", chat.ID),
		zap.String("chat_name", chat.Name),
		zap.Int("messages", len(messages)))

	// Ordenar mensagens por timestamp (mais antigas primeiro)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp < messages[j].Timestamp
	})

	// Extrair número de telefone do chat ID (formato: 5511999999999@c.us)
	phoneNumber := extractPhoneNumber(chat.ID)

	// Buscar ou criar contato
	cont, err := h.getOrCreateContact(ctx, ch.ProjectID, ch.TenantID, phoneNumber, chat.Name)
	if err != nil {
		return fmt.Errorf("failed to get or create contact: %w", err)
	}

	if cont == nil {
		return fmt.Errorf("contact is nil after get or create")
	}

	// Agrupar mensagens em sessões
	sessions := h.groupMessagesIntoSessions(messages, cont.ID(), ch.TenantID)

	// Criar sessões e mensagens
	for _, sess := range sessions {
		// Criar sessão
		if err := h.sessionRepo.Save(ctx, sess.session); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}
		result.SessionsCreated++

		// Criar mensagens da sessão
		for _, msg := range sess.messages {
			if err := h.messageRepo.Save(ctx, msg); err != nil {
				h.logger.Warn("Failed to create message",
					zap.Error(err),
					zap.String("message_id", msg.ID().String()))
				continue
			}
			result.MessagesImported++
		}
	}

	return nil
}

// sessionWithMessages agrupa uma sessão com suas mensagens
type sessionWithMessages struct {
	session  *session.Session
	messages []*message.Message
}

// groupMessagesIntoSessions agrupa mensagens em sessões baseado em timeout
func (h *WAHAHistoryImporter) groupMessagesIntoSessions(
	wahaMessages []waha.MessagePayload,
	contactID uuid.UUID,
	tenantID string,
) []*sessionWithMessages {
	if len(wahaMessages) == 0 {
		return nil
	}

	sessions := []*sessionWithMessages{}
	var currentSession *sessionWithMessages

	for _, wahaMsg := range wahaMessages {
		msgTime := time.Unix(wahaMsg.Timestamp, 0)

		// Se não tem sessão atual ou passou do timeout, cria nova sessão
		if currentSession == nil || h.shouldCreateNewSession(currentSession, msgTime) {
			// Criar nova sessão
			sess, err := session.NewSession(contactID, tenantID, nil, h.sessionTimeout)
			if err != nil {
				h.logger.Error("Failed to create session", zap.Error(err))
				continue
			}

			currentSession = &sessionWithMessages{
				session:  sess,
				messages: []*message.Message{},
			}
			sessions = append(sessions, currentSession)
		}

		// Criar mensagem
		fromMe := wahaMsg.FromMe

		contentType := message.ContentTypeText
		switch wahaMsg.Type {
		case "image":
			contentType = message.ContentTypeImage
		case "video":
			contentType = message.ContentTypeVideo
		case "audio", "ptt":
			contentType = message.ContentTypeAudio
		case "document":
			contentType = message.ContentTypeDocument
		case "location":
			contentType = message.ContentTypeLocation
		}

		// Precisamos do projectID e customerID - vamos buscar do contato
		// Por enquanto, vamos usar valores temporários que serão atualizados
		// TODO: Melhorar isso buscando os IDs corretos
		projectID := uuid.Nil  // Será preenchido pela camada de aplicação
		customerID := uuid.Nil // Será preenchido pela camada de aplicação

		msg, err := message.NewMessage(
			contactID,
			projectID,
			customerID,
			contentType,
			fromMe,
		)
		if err != nil {
			h.logger.Error("Failed to create message", zap.Error(err))
			continue
		}

		// Configurar campos adicionais via métodos públicos
		msg.AssignToSession(currentSession.session.ID())
		if wahaMsg.Body != "" {
			msg.SetText(wahaMsg.Body)
		}
		msg.SetChannelMessageID(wahaMsg.ID)

		if wahaMsg.MediaURL != "" {
			msg.SetMediaContent(wahaMsg.MediaURL, wahaMsg.MimeType)
		}

		// Registrar mensagem na sessão
		currentSession.session.RecordMessage(!wahaMsg.FromMe, time.Unix(wahaMsg.Timestamp, 0))
		currentSession.messages = append(currentSession.messages, msg)
	}

	// Encerrar todas as sessões importadas
	for _, sess := range sessions {
		if sess.session.IsActive() {
			sess.session.End(session.ReasonManualClose)
		}
	}

	return sessions
}

// shouldCreateNewSession verifica se deve criar nova sessão baseado no timeout
func (h *WAHAHistoryImporter) shouldCreateNewSession(current *sessionWithMessages, msgTime time.Time) bool {
	if len(current.messages) == 0 {
		return false
	}

	// Pegar timestamp da última mensagem
	lastMsg := current.messages[len(current.messages)-1]
	lastMsgTime := lastMsg.Timestamp()

	// Se passou do timeout, cria nova sessão
	return msgTime.Sub(lastMsgTime) > h.sessionTimeout
}

// getOrCreateContact busca ou cria um contato
func (h *WAHAHistoryImporter) getOrCreateContact(
	ctx context.Context,
	projectID uuid.UUID,
	tenantID string,
	phoneNumber, name string,
) (*contact.Contact, error) {
	// Tentar buscar por telefone
	cont, err := h.contactRepo.FindByPhone(ctx, projectID, phoneNumber)
	if err == nil && cont != nil {
		return cont, nil
	}

	// Criar novo contato
	if name == "" {
		name = phoneNumber // Usar telefone como nome se não tiver nome
	}

	newContact, err := contact.NewContact(projectID, tenantID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	// Definir telefone
	if err := newContact.SetPhone(phoneNumber); err != nil {
		h.logger.Warn("Invalid phone number", zap.String("phone", phoneNumber), zap.Error(err))
	}

	if err := h.contactRepo.Save(ctx, newContact); err != nil {
		return nil, fmt.Errorf("failed to save contact: %w", err)
	}

	h.logger.Info("Created new contact from history",
		zap.String("phone", phoneNumber),
		zap.String("name", name))

	return newContact, nil
}

// extractPhoneNumber extrai número de telefone do chat ID
// Formato: 5511999999999@c.us -> 5511999999999
func extractPhoneNumber(chatID string) string {
	// Remove sufixo @c.us, @g.us, etc
	for _, suffix := range []string{"@c.us", "@g.us", "@s.whatsapp.net"} {
		if len(chatID) > len(suffix) && chatID[len(chatID)-len(suffix):] == suffix {
			return chatID[:len(chatID)-len(suffix)]
		}
	}
	return chatID
}
