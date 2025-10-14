package channel

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/channels/waha"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"github.com/ventros/crm/internal/domain/crm/contact"
	"github.com/ventros/crm/internal/domain/crm/message"
	"github.com/ventros/crm/internal/domain/crm/session"
	"go.uber.org/zap"
)

// WAHAHistoryImportActivities contém todas as activities necessárias para importação
type WAHAHistoryImportActivities struct {
	logger      *zap.Logger
	wahaClient  *waha.WAHAClient
	channelRepo channel.Repository
	contactRepo contact.Repository
	sessionRepo session.Repository
	messageRepo message.Repository
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
		logger:      logger,
		wahaClient:  wahaClient,
		channelRepo: channelRepo,
		contactRepo: contactRepo,
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
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

	// Buscar mensagens do chat com filtro de data se especificado
	// limit == 0 significa SEM LIMITE (buscar todas as mensagens disponíveis)
	limit := input.Limit

	// ⚠️ IMPORTANTE: TimeRangeDays deve ser > 0 para ativar filtro de tempo
	// Se TimeRangeDays == 0, busca TODAS as mensagens disponíveis no WAHA
	var timestampGte int64
	if input.TimeRangeDays > 0 {
		cutoffTime := time.Now().AddDate(0, 0, -input.TimeRangeDays)
		timestampGte = cutoffTime.Unix()

		a.logger.Info("Fetching messages with time filter",
			zap.String("chat_id", input.ChatID),
			zap.String("chat_name", input.ChatName),
			zap.Int("time_range_days", input.TimeRangeDays),
			zap.Int64("timestamp_gte", timestampGte),
			zap.String("cutoff_date", cutoffTime.Format(time.RFC3339)))
	} else {
		a.logger.Info("Fetching ALL available messages (no time filter)",
			zap.String("chat_id", input.ChatID),
			zap.String("chat_name", input.ChatName),
			zap.Int("limit", limit))
	}

	// Buscar mensagens da API WAHA
	// timestampGte=0 significa sem filtro de tempo (buscar todas)
	messages, err := a.wahaClient.GetChatMessagesWithFilter(ctx, input.SessionID, input.ChatID, limit, false, timestampGte, 0)
	if err != nil {
		a.logger.Error("Failed to fetch messages from WAHA",
			zap.String("chat_id", input.ChatID),
			zap.String("session_id", input.SessionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	a.logger.Info("Messages fetched from WAHA successfully",
		zap.String("chat_id", input.ChatID),
		zap.Int("messages_count", len(messages)),
		zap.Int64("timestamp_filter", timestampGte))

	if len(messages) == 0 {
		a.logger.Debug("Chat has no messages in time range", zap.String("chat_id", input.ChatID))
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

	// Configurar timeout de sessão (default 30 minutos se não especificado)
	sessionTimeoutMinutes := input.SessionTimeoutMinutes
	if sessionTimeoutMinutes == 0 {
		sessionTimeoutMinutes = 30
	}
	sessionTimeout := time.Duration(sessionTimeoutMinutes) * time.Minute

	// Agrupar mensagens em sessões
	sessions := a.groupMessagesIntoSessions(messages, cont.ID(), projectID, input.TenantID, channelID, sessionTimeout)

	a.logger.Info("Messages grouped into sessions",
		zap.String("chat_id", input.ChatID),
		zap.String("chat_name", input.ChatName),
		zap.Int("sessions_count", len(sessions)),
		zap.Int("total_messages", len(messages)),
		zap.Int("session_timeout_minutes", sessionTimeoutMinutes))

	// Criar sessões e mensagens com controle de duplicação
	for _, sess := range sessions {
		// Salvar sessão
		if err := a.sessionRepo.Save(ctx, sess.session); err != nil {
			return nil, fmt.Errorf("failed to save session: %w", err)
		}
		result.SessionsCreated++

		// Salvar mensagens da sessão com controle de duplicação
		for _, msg := range sess.messages {
			// Verificar se mensagem já existe (por channel_message_id)
			// Se já existe, skip (evitar duplicação em re-imports)
			channelMsgID := msg.ChannelMessageID()
			if channelMsgID != nil && *channelMsgID != "" {
				existingMsg, err := a.messageRepo.FindByChannelMessageID(ctx, *channelMsgID)
				if err == nil && existingMsg != nil {
					a.logger.Debug("Message already imported, skipping",
						zap.String("channel_message_id", *channelMsgID),
						zap.String("chat_id", input.ChatID))
					continue // Skip duplicata
				}
			}

			if err := a.messageRepo.Save(ctx, msg); err != nil {
				channelMsgIDStr := ""
				if channelMsgID != nil {
					channelMsgIDStr = *channelMsgID
				}
				a.logger.Warn("Failed to save message",
					zap.String("message_id", msg.ID().String()),
					zap.String("channel_message_id", channelMsgIDStr),
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
	projectID uuid.UUID,
	tenantID string,
	channelID uuid.UUID,
	sessionTimeout time.Duration,
) []*sessionWithMessages {
	if len(wahaMessages) == 0 {
		return nil
	}

	sessions := []*sessionWithMessages{}
	var currentSession *sessionWithMessages

	for msgIndex, wahaMsg := range wahaMessages {
		msgTime := time.Unix(wahaMsg.Timestamp, 0)

		// Se não tem sessão atual ou passou do timeout, cria nova sessão
		if currentSession == nil || a.shouldCreateNewSession(currentSession, msgTime, sessionTimeout) {
			// Log da nova sessão sendo criada
			sessionNumber := len(sessions) + 1

			// Se temos sessão anterior, encerrar e logar estatísticas
			if currentSession != nil {
				firstMsgTime := currentSession.messages[0].Timestamp()
				lastMsgTime := currentSession.messages[len(currentSession.messages)-1].Timestamp()
				sessionDuration := lastMsgTime.Sub(firstMsgTime)

				a.logger.Info("Session completed",
					zap.Int("session_number", sessionNumber-1),
					zap.Int("messages_count", len(currentSession.messages)),
					zap.Time("start_time", firstMsgTime),
					zap.Time("end_time", lastMsgTime),
					zap.Duration("duration", sessionDuration),
					zap.Duration("gap_to_next", msgTime.Sub(lastMsgTime)))
			}

			// Use msgTime as session start time (critical for history import!)
			// This ensures sessions are ordered correctly by actual message time
			sess, err := session.NewSessionWithTimestamp(contactID, tenantID, nil, sessionTimeout, msgTime)
			if err != nil {
				a.logger.Error("Failed to create session", zap.Error(err))
				continue
			}

			a.logger.Info("New session created",
				zap.Int("session_number", sessionNumber),
				zap.Time("start_time", msgTime),
				zap.Duration("timeout", sessionTimeout),
				zap.Int("message_index", msgIndex))

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

		// Reconstruir mensagem com timestamp original do WAHA (não usar NewMessage!)
		// customerID será o próprio contactID para mensagens históricas
		customerID := contactID
		sessionIDPtr := currentSession.session.ID()

		// Preparar dados opcionais
		var textPtr *string
		if wahaMsg.Body != "" {
			textPtr = &wahaMsg.Body
		}

		var mediaURLPtr *string
		var mediaMimetypePtr *string
		if wahaMsg.MediaURL != "" {
			mediaURLPtr = &wahaMsg.MediaURL
			mediaMimetypePtr = &wahaMsg.MimeType
		}

		// ✅ CRÍTICO: Usar ReconstructMessage() para preservar timestamp original!
		// Isso garante que sessões sejam agrupadas corretamente por timeout
		msg := message.ReconstructMessage(
			uuid.New(),                  // id
			msgTime,                     // ✅ timestamp ORIGINAL do WAHA!
			customerID,                  // customerID
			projectID,                   // projectID
			nil,                         // channelTypeID
			wahaMsg.FromMe,              // fromMe
			channelID,                   // channelID
			contactID,                   // contactID
			&sessionIDPtr,               // sessionID
			contentType,                 // contentType
			textPtr,                     // text
			mediaURLPtr,                 // mediaURL
			mediaMimetypePtr,            // mediaMimetype
			&wahaMsg.ID,                 // channelMessageID
			nil,                         // replyToID
			message.StatusSent,          // status
			nil,                         // language
			nil,                         // agentID
			message.SourceHistoryImport, // source
			nil,                         // metadata
			nil,                         // deliveredAt
			nil,                         // readAt
			nil,                         // playedAt
			nil,                         // mentions
		)

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
func (a *WAHAHistoryImportActivities) shouldCreateNewSession(current *sessionWithMessages, msgTime time.Time, sessionTimeout time.Duration) bool {
	if len(current.messages) == 0 {
		return false
	}

	// Pegar timestamp da última mensagem
	lastMsg := current.messages[len(current.messages)-1]
	lastMsgTime := lastMsg.Timestamp()

	// Se passou do timeout, cria nova sessão
	gap := msgTime.Sub(lastMsgTime)
	return gap > sessionTimeout
}

// ProcessBufferedWebhooksActivity processa webhooks que foram buffered durante import
// SAGA Pattern: webhooks recebidos durante import são enfileirados e processados após
func (a *WAHAHistoryImportActivities) ProcessBufferedWebhooksActivity(ctx context.Context, input ProcessBufferedWebhooksActivityInput) (*ProcessBufferedWebhooksActivityResult, error) {
	a.logger.Info("Processing buffered webhooks after import",
		zap.String("channel_id", input.ChannelID))

	result := &ProcessBufferedWebhooksActivityResult{
		ChannelID:         input.ChannelID,
		WebhooksProcessed: 0,
		Errors:            []string{},
	}

	// Buscar channel
	channelID, err := uuid.Parse(input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel_id: %w", err)
	}

	ch, err := a.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// Mudar status do canal para "completed" (não está mais importando)
	ch.CompleteHistoryImport(channel.HistoryImportStats{
		Total:     input.TotalMessagesImported,
		Processed: input.TotalMessagesImported,
		Failed:    0,
		StartedAt: time.Now().Add(-time.Duration(input.ImportDurationSeconds) * time.Second),
	})

	if err := a.channelRepo.Update(ch); err != nil {
		return nil, fmt.Errorf("failed to update channel status: %w", err)
	}

	a.logger.Info("Channel status updated to completed, buffered webhooks will be processed by consumer",
		zap.String("channel_id", input.ChannelID),
		zap.Int("total_imported", input.TotalMessagesImported))

	// Nota: Os webhooks buffered serão processados automaticamente pelo consumer RabbitMQ
	// pois agora ch.IsHistoryImportInProgress() retorna false
	// O consumer já está configurado para processar a fila webhooks.buffered.{channel_id}

	return result, nil
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
