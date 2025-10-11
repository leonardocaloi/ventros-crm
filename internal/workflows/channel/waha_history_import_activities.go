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

// WAHAHistoryImportActivities contém todas as activities necessárias para importação
type WAHAHistoryImportActivities struct {
	logger         *zap.Logger
	wahaClient     *waha.WAHAClient
	channelRepo    channel.Repository
	contactRepo    contact.Repository
	sessionRepo    session.Repository
	messageRepo    message.Repository
	sessionTimeout time.Duration
}

// NewWAHAHistoryImportActivities cria novas activities
func NewWAHAHistoryImportActivities(
	logger *zap.Logger,
	wahaClient *waha.WAHAClient,
	channelRepo channel.Repository,
	contactRepo contact.Repository,
	sessionRepo session.Repository,
	messageRepo message.Repository,
) *WAHAHistoryImportActivities {
	return &WAHAHistoryImportActivities{
		logger:         logger,
		wahaClient:     wahaClient,
		channelRepo:    channelRepo,
		contactRepo:    contactRepo,
		sessionRepo:    sessionRepo,
		messageRepo:    messageRepo,
		sessionTimeout: 30 * time.Minute, // Timeout padrão entre mensagens
	}
}

// FetchWAHAChatsActivity busca lista de chats do WAHA
func (a *WAHAHistoryImportActivities) FetchWAHAChatsActivity(ctx context.Context, input FetchChatsActivityInput) ([]ChatInfo, error) {
	a.logger.Info("Fetching chats from WAHA", zap.String("session_id", input.SessionID))

	// Buscar chats com paginação
	allChats := []waha.ChatOverview{}
	limit := 100
	offset := 0

	for {
		chats, err := a.wahaClient.GetChatsOverview(ctx, input.SessionID, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch chats: %w", err)
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

	// Converter para ChatInfo
	result := make([]ChatInfo, len(allChats))
	for i, chat := range allChats {
		result[i] = ChatInfo{
			ID:   chat.ID,
			Name: chat.Name,
		}
	}

	a.logger.Info("Chats fetched successfully", zap.Int("count", len(result)))
	return result, nil
}

// ImportChatHistoryActivity importa histórico de um chat específico
func (a *WAHAHistoryImportActivities) ImportChatHistoryActivity(ctx context.Context, input ImportChatHistoryActivityInput) (*ImportChatHistoryActivityResult, error) {
	a.logger.Info("Importing chat history",
		zap.String("channel_id", input.ChannelID),
		zap.String("chat_id", input.ChatID),
		zap.String("chat_name", input.ChatName))

	result := &ImportChatHistoryActivityResult{
		ChatID: input.ChatID,
	}

	// Buscar mensagens do chat
	limit := input.Limit
	if limit == 0 {
		limit = 1000 // Limite padrão
	}

	messages, err := a.wahaClient.GetChatMessages(ctx, input.SessionID, input.ChatID, limit, false)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	if len(messages) == 0 {
		a.logger.Debug("Chat has no messages", zap.String("chat_id", input.ChatID))
		return result, nil
	}

	// Ordenar mensagens por timestamp (mais antigas primeiro)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp < messages[j].Timestamp
	})

	// Extrair número de telefone do chat ID
	phoneNumber := extractPhoneNumber(input.ChatID)

	// Parse IDs
	channelID, err := uuid.Parse(input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel_id: %w", err)
	}

	projectID, err := uuid.Parse(input.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project_id: %w", err)
	}

	// Buscar ou criar contato
	cont, isNew, err := a.getOrCreateContact(ctx, projectID, input.TenantID, phoneNumber, input.ChatName)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create contact: %w", err)
	}

	if isNew {
		result.ContactsCreated = 1
	}

	// Agrupar mensagens em sessões
	sessions := a.groupMessagesIntoSessions(messages, cont.ID(), input.TenantID, channelID)

	// Criar sessões e mensagens
	for _, sess := range sessions {
		// Salvar sessão
		if err := a.sessionRepo.Save(ctx, sess.session); err != nil {
			return nil, fmt.Errorf("failed to save session: %w", err)
		}
		result.SessionsCreated++

		// Salvar mensagens da sessão
		for _, msg := range sess.messages {
			if err := a.messageRepo.Save(ctx, msg); err != nil {
				a.logger.Warn("Failed to save message",
					zap.String("message_id", msg.ID().String()),
					zap.Error(err))
				continue
			}
			result.MessagesImported++
		}
	}

	a.logger.Info("Chat history imported successfully",
		zap.String("chat_id", input.ChatID),
		zap.Int("messages_imported", result.MessagesImported),
		zap.Int("sessions_created", result.SessionsCreated))

	return result, nil
}

// MarkImportCompletedActivity marca a importação como concluída no canal
func (a *WAHAHistoryImportActivities) MarkImportCompletedActivity(ctx context.Context, input MarkImportCompletedActivityInput) error {
	a.logger.Info("Marking import as completed", zap.String("channel_id", input.ChannelID))

	channelID, err := uuid.Parse(input.ChannelID)
	if err != nil {
		return fmt.Errorf("invalid channel_id: %w", err)
	}

	ch, err := a.channelRepo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	ch.SetWAHAImportCompleted()

	if err := a.channelRepo.Update(ch); err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	a.logger.Info("Import marked as completed", zap.String("channel_id", input.ChannelID))
	return nil
}

// getOrCreateContact busca ou cria um contato
func (a *WAHAHistoryImportActivities) getOrCreateContact(
	ctx context.Context,
	projectID uuid.UUID,
	tenantID string,
	phoneNumber, name string,
) (*contact.Contact, bool, error) {
	// Tentar buscar por telefone
	cont, err := a.contactRepo.FindByPhone(ctx, projectID, phoneNumber)
	if err == nil && cont != nil {
		return cont, false, nil
	}

	// Criar novo contato
	if name == "" {
		name = phoneNumber // Usar telefone como nome se não tiver nome
	}

	newContact, err := contact.NewContact(projectID, tenantID, name)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create contact: %w", err)
	}

	// Definir telefone
	if err := newContact.SetPhone(phoneNumber); err != nil {
		a.logger.Warn("Invalid phone number", zap.String("phone", phoneNumber), zap.Error(err))
	}

	if err := a.contactRepo.Save(ctx, newContact); err != nil {
		return nil, false, fmt.Errorf("failed to save contact: %w", err)
	}

	a.logger.Info("Created new contact from history",
		zap.String("phone", phoneNumber),
		zap.String("name", name))

	return newContact, true, nil
}

// sessionWithMessages agrupa uma sessão com suas mensagens
type sessionWithMessages struct {
	session  *session.Session
	messages []*message.Message
}

// groupMessagesIntoSessions agrupa mensagens em sessões baseado em timeout
func (a *WAHAHistoryImportActivities) groupMessagesIntoSessions(
	wahaMessages []waha.MessagePayload,
	contactID uuid.UUID,
	tenantID string,
	channelID uuid.UUID,
) []*sessionWithMessages {
	if len(wahaMessages) == 0 {
		return nil
	}

	sessions := []*sessionWithMessages{}
	var currentSession *sessionWithMessages

	for _, wahaMsg := range wahaMessages {
		msgTime := time.Unix(wahaMsg.Timestamp, 0)

		// Se não tem sessão atual ou passou do timeout, cria nova sessão
		if currentSession == nil || a.shouldCreateNewSession(currentSession, msgTime) {
			sess, err := session.NewSession(contactID, tenantID, nil, a.sessionTimeout)
			if err != nil {
				a.logger.Error("Failed to create session", zap.Error(err))
				continue
			}

			currentSession = &sessionWithMessages{
				session:  sess,
				messages: []*message.Message{},
			}
			sessions = append(sessions, currentSession)
		}

		// Determinar tipo de conteúdo
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

		// Criar mensagem
		// TODO: Get actual projectID and customerID
		projectID := uuid.Nil
		customerID := uuid.Nil

		msg, err := message.NewMessage(
			contactID,
			projectID,
			customerID,
			contentType,
			wahaMsg.FromMe,
		)
		if err != nil {
			a.logger.Error("Failed to create message", zap.Error(err))
			continue
		}

		// Configurar campos adicionais
		msg.AssignToSession(currentSession.session.ID())
		// TODO: Add SetChannelID method to message if needed
		// msg.SetChannelID(channelID)
		if wahaMsg.Body != "" {
			msg.SetText(wahaMsg.Body)
		}
		msg.SetChannelMessageID(wahaMsg.ID)

		if wahaMsg.MediaURL != "" {
			msg.SetMediaContent(wahaMsg.MediaURL, wahaMsg.MimeType)
		}

		// Registrar mensagem na sessão
		currentSession.session.RecordMessage(!wahaMsg.FromMe, msgTime)
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
func (a *WAHAHistoryImportActivities) shouldCreateNewSession(current *sessionWithMessages, msgTime time.Time) bool {
	if len(current.messages) == 0 {
		return false
	}

	// Pegar timestamp da última mensagem
	lastMsg := current.messages[len(current.messages)-1]
	lastMsgTime := lastMsg.Timestamp()

	// Se passou do timeout, cria nova sessão
	return msgTime.Sub(lastMsgTime) > a.sessionTimeout
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
