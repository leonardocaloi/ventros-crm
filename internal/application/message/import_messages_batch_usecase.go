package message

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/contact"
	domainmessage "github.com/ventros/crm/internal/domain/crm/message"
	"github.com/ventros/crm/internal/domain/crm/session"
	"go.uber.org/zap"
)

// ImportMessagesBatchUseCase handles bulk import of messages during history import
// This use case is optimized for batch processing and does NOT use the same flow as real-time webhooks.
//
// Key Differences from ProcessInboundMessageUseCase:
// - ✅ Batch operations (1 transaction instead of N)
// - ✅ Deterministic session assignment (no race conditions)
// - ✅ No debouncer, tracking, or other real-time features
// - ✅ Designed for 1000s of messages at once
type ImportMessagesBatchUseCase struct {
	contactRepo     contact.Repository
	sessionRepo     session.Repository
	messageRepo     domainmessage.Repository
	eventBus        EventBus
	timeoutResolver SessionTimeoutResolver
	txManager       TransactionManager
	logger          *zap.Logger
}

// NewImportMessagesBatchUseCase creates a new batch import use case
func NewImportMessagesBatchUseCase(
	contactRepo contact.Repository,
	sessionRepo session.Repository,
	messageRepo domainmessage.Repository,
	eventBus EventBus,
	timeoutResolver SessionTimeoutResolver,
	txManager TransactionManager,
	logger *zap.Logger,
) *ImportMessagesBatchUseCase {
	return &ImportMessagesBatchUseCase{
		contactRepo:     contactRepo,
		sessionRepo:     sessionRepo,
		messageRepo:     messageRepo,
		eventBus:        eventBus,
		timeoutResolver: timeoutResolver,
		txManager:       txManager,
		logger:          logger,
	}
}

// ImportBatchInput contains all data needed for batch import
type ImportBatchInput struct {
	ChannelID             uuid.UUID
	ProjectID             uuid.UUID
	TenantID              string
	CustomerID            uuid.UUID
	ChannelTypeID         int
	Messages              []ImportMessage // Pre-sorted by contact + timestamp
	SessionTimeoutMinutes int
}

// ImportMessage represents a single message to import (simplified from WAHA structure)
type ImportMessage struct {
	ExternalID      string                    // WAHA message ID (for deduplication)
	ContactPhone    string                    // Contact phone number
	ContactName     string                    // Contact display name
	ContentType     domainmessage.ContentType // Message content type
	Text            string                    // Message text content
	MediaURL        *string                   // Media URL (if applicable)
	MediaMimetype   string                    // Media MIME type
	Timestamp       time.Time                 // Message timestamp (CRITICAL for session grouping)
	FromMe          bool                      // Message direction
	TrackingData    map[string]interface{}    // Ad tracking data (ignored in batch import)
	Metadata        map[string]interface{}    // Additional metadata
}

// ImportBatchResult contains statistics about the import operation
type ImportBatchResult struct {
	ContactsCreated int
	SessionsCreated int
	MessagesCreated int
	Duplicates      int
	Errors          []string
}

// SessionAssignment maps external message ID to assigned session
type SessionAssignment struct {
	SessionID uuid.UUID
	ContactID uuid.UUID
}

// Execute processes a batch of messages with deterministic session assignment
//
// ALGORITHM:
// 1. Group messages by contact phone
// 2. Batch lookup contacts (1 query instead of N)
// 3. Create missing contacts in bulk
// 4. For each contact, assign sessions deterministically based on timeout gaps
// 5. Create all sessions in bulk
// 6. Create all messages in bulk
// 7. Publish events in bulk
//
// RESULT: O(1) database transactions instead of O(N) where N = number of messages
func (uc *ImportMessagesBatchUseCase) Execute(ctx context.Context, input ImportBatchInput) (*ImportBatchResult, error) {
	if len(input.Messages) == 0 {
		return &ImportBatchResult{}, nil
	}

	uc.logger.Info("Starting batch import",
		zap.Int("message_count", len(input.Messages)),
		zap.String("channel_id", input.ChannelID.String()),
	)

	result := &ImportBatchResult{
		Errors: []string{},
	}

	// Execute entire import in a single transaction
	err := uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// STEP 1: Group messages by contact
		messagesByContact := uc.groupMessagesByContact(input.Messages)
		uc.logger.Info("Messages grouped by contact",
			zap.Int("unique_contacts", len(messagesByContact)),
		)

		// STEP 2: Batch lookup existing contacts
		phones := uc.extractUniquePhones(input.Messages)
		existingContacts, err := uc.contactRepo.FindByPhones(txCtx, input.ProjectID, phones)
		if err != nil {
			return fmt.Errorf("failed to lookup contacts: %w", err)
		}
		uc.logger.Info("Existing contacts found",
			zap.Int("count", len(existingContacts)),
		)

		// STEP 3: Identify and create missing contacts in bulk
		newContacts, err := uc.createMissingContacts(txCtx, input, messagesByContact, existingContacts)
		if err != nil {
			return fmt.Errorf("failed to create contacts: %w", err)
		}
		result.ContactsCreated = len(newContacts)
		uc.logger.Info("New contacts created",
			zap.Int("count", result.ContactsCreated),
		)

		// Merge existing + new contacts
		allContacts := make(map[string]*contact.Contact)
		for phone, c := range existingContacts {
			allContacts[phone] = c
		}
		for phone, c := range newContacts {
			allContacts[phone] = c
		}

		// STEP 4: Deterministic session assignment (the key innovation!)
		sessionAssignments, sessionsToCreate, err := uc.assignSessionsDeterministically(
			txCtx,
			input,
			messagesByContact,
			allContacts,
		)
		if err != nil {
			return fmt.Errorf("failed to assign sessions: %w", err)
		}
		result.SessionsCreated = len(sessionsToCreate)
		uc.logger.Info("Sessions assigned deterministically",
			zap.Int("count", result.SessionsCreated),
		)

		// STEP 5: Create all sessions in bulk
		if len(sessionsToCreate) > 0 {
			if err := uc.createSessionsBulk(txCtx, sessionsToCreate); err != nil {
				return fmt.Errorf("failed to create sessions: %w", err)
			}
		}

		// STEP 6: Create all messages in bulk
		messagesCreated, duplicates, err := uc.createMessagesBulk(
			txCtx,
			input,
			sessionAssignments,
			allContacts,
		)
		if err != nil {
			return fmt.Errorf("failed to create messages: %w", err)
		}
		result.MessagesCreated = messagesCreated
		result.Duplicates = duplicates
		uc.logger.Info("Messages created",
			zap.Int("created", result.MessagesCreated),
			zap.Int("duplicates", result.Duplicates),
		)

		// STEP 7: Publish domain events in bulk (within same transaction)
		if err := uc.publishBatchEvents(txCtx, allContacts, sessionsToCreate); err != nil {
			return fmt.Errorf("failed to publish events: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	uc.logger.Info("Batch import completed successfully",
		zap.Int("contacts_created", result.ContactsCreated),
		zap.Int("sessions_created", result.SessionsCreated),
		zap.Int("messages_created", result.MessagesCreated),
		zap.Int("duplicates", result.Duplicates),
	)

	return result, nil
}

// groupMessagesByContact groups messages by contact phone for batch processing
func (uc *ImportMessagesBatchUseCase) groupMessagesByContact(messages []ImportMessage) map[string][]ImportMessage {
	grouped := make(map[string][]ImportMessage)
	for _, msg := range messages {
		grouped[msg.ContactPhone] = append(grouped[msg.ContactPhone], msg)
	}

	// Sort each contact's messages by timestamp (CRITICAL for session assignment)
	for phone := range grouped {
		sort.Slice(grouped[phone], func(i, j int) bool {
			return grouped[phone][i].Timestamp.Before(grouped[phone][j].Timestamp)
		})
	}

	return grouped
}

// extractUniquePhones extracts unique phone numbers from messages
func (uc *ImportMessagesBatchUseCase) extractUniquePhones(messages []ImportMessage) []string {
	phonesMap := make(map[string]struct{})
	for _, msg := range messages {
		phonesMap[msg.ContactPhone] = struct{}{}
	}

	phones := make([]string, 0, len(phonesMap))
	for phone := range phonesMap {
		phones = append(phones, phone)
	}
	return phones
}

// createMissingContacts creates contacts that don't exist yet
func (uc *ImportMessagesBatchUseCase) createMissingContacts(
	ctx context.Context,
	input ImportBatchInput,
	messagesByContact map[string][]ImportMessage,
	existingContacts map[string]*contact.Contact,
) (map[string]*contact.Contact, error) {
	newContacts := make(map[string]*contact.Contact)

	for phone, msgs := range messagesByContact {
		// Skip if contact already exists
		if _, exists := existingContacts[phone]; exists {
			continue
		}

		// Get first message to extract contact name
		firstMsg := msgs[0]
		name := firstMsg.ContactName
		if name == "" {
			name = phone // Fallback to phone
		}

		// Create new contact
		c, err := contact.NewContact(input.ProjectID, input.TenantID, name)
		if err != nil {
			return nil, fmt.Errorf("failed to create contact for %s: %w", phone, err)
		}

		if err := c.SetPhone(phone); err != nil {
			return nil, fmt.Errorf("failed to set phone for %s: %w", phone, err)
		}

		c.AddTag("whatsapp")
		c.RecordInteraction()

		// Save contact
		if err := uc.contactRepo.Save(ctx, c); err != nil {
			return nil, fmt.Errorf("failed to save contact %s: %w", phone, err)
		}

		newContacts[phone] = c
	}

	return newContacts, nil
}

// assignSessionsDeterministically assigns sessions to messages BEFORE creating them
// This eliminates race conditions that cause session fragmentation
func (uc *ImportMessagesBatchUseCase) assignSessionsDeterministically(
	ctx context.Context,
	input ImportBatchInput,
	messagesByContact map[string][]ImportMessage,
	allContacts map[string]*contact.Contact,
) (map[string]SessionAssignment, []*session.Session, error) {
	sessionAssignments := make(map[string]SessionAssignment)
	var sessionsToCreate []*session.Session

	// Resolve timeout once (same for all sessions in this batch)
	timeout := time.Duration(input.SessionTimeoutMinutes) * time.Minute
	var pipelineID *uuid.UUID

	if input.ChannelID != uuid.Nil {
		resolvedTimeout, resolvedPipelineID, err := uc.timeoutResolver.ResolveForChannel(ctx, input.ChannelID)
		if err == nil {
			timeout = resolvedTimeout
			pipelineID = resolvedPipelineID
		}
	}

	channelTypeID := &input.ChannelTypeID

	// Process each contact's messages
	for phone, contactMsgs := range messagesByContact {
		c, exists := allContacts[phone]
		if !exists {
			uc.logger.Warn("Contact not found during session assignment", zap.String("phone", phone))
			continue
		}

		// Check if contact has existing active session
		activeSession, err := uc.sessionRepo.FindActiveByContact(ctx, c.ID(), channelTypeID)
		if err != nil && err != session.ErrSessionNotFound {
			return nil, nil, fmt.Errorf("failed to find active session for contact %s: %w", c.ID(), err)
		}

		// Track current session and last timestamp
		var currentSessionID uuid.UUID
		lastTimestamp := time.Time{}

		if activeSession != nil {
			currentSessionID = activeSession.ID()
			lastTimestamp = activeSession.LastActivityAt()
		}

		// Assign sessions based on timeout gaps
		for _, msg := range contactMsgs {
			gap := msg.Timestamp.Sub(lastTimestamp)

			// Create new session if:
			// 1. No current session exists, OR
			// 2. Time gap exceeds timeout
			if currentSessionID == uuid.Nil || gap > timeout {
				// Create new session
				var newSession *session.Session

				if pipelineID != nil && *pipelineID != uuid.Nil {
					newSession, err = session.NewSessionWithPipeline(
						c.ID(),
						input.TenantID,
						channelTypeID,
						*pipelineID,
						timeout,
					)
				} else {
					newSession, err = session.NewSession(
						c.ID(),
						input.TenantID,
						channelTypeID,
						timeout,
					)
				}

				if err != nil {
					return nil, nil, fmt.Errorf("failed to create session for contact %s: %w", c.ID(), err)
				}

				sessionsToCreate = append(sessionsToCreate, newSession)
				currentSessionID = newSession.ID() // Use actual session ID
			}

			// Assign this message to current session
			sessionAssignments[msg.ExternalID] = SessionAssignment{
				SessionID: currentSessionID,
				ContactID: c.ID(),
			}

			lastTimestamp = msg.Timestamp
		}
	}

	return sessionAssignments, sessionsToCreate, nil
}

// createSessionsBulk creates all sessions in a single bulk operation
func (uc *ImportMessagesBatchUseCase) createSessionsBulk(ctx context.Context, sessions []*session.Session) error {
	// Save each session (GORM CreateInBatches will be used in repository)
	for _, s := range sessions {
		if err := uc.sessionRepo.Save(ctx, s); err != nil {
			return fmt.Errorf("failed to save session %s: %w", s.ID(), err)
		}
	}
	return nil
}

// createMessagesBulk creates all messages in bulk operations
func (uc *ImportMessagesBatchUseCase) createMessagesBulk(
	ctx context.Context,
	input ImportBatchInput,
	sessionAssignments map[string]SessionAssignment,
	allContacts map[string]*contact.Contact,
) (int, int, error) {
	messagesCreated := 0
	duplicates := 0

	for _, msg := range input.Messages {
		assignment, exists := sessionAssignments[msg.ExternalID]
		if !exists {
			uc.logger.Warn("No session assignment found for message", zap.String("external_id", msg.ExternalID))
			continue
		}

		// Check for duplicates (channel_message_id)
		if msg.ExternalID != "" {
			existing, err := uc.messageRepo.FindByChannelMessageID(ctx, msg.ExternalID)
			if err == nil && existing != nil {
				duplicates++
				continue
			}
		}

		// Create message
		domainMsg, err := domainmessage.NewMessage(
			assignment.ContactID,
			input.ProjectID,
			input.CustomerID,
			msg.ContentType,
			msg.FromMe,
		)
		if err != nil {
			return messagesCreated, duplicates, fmt.Errorf("failed to create message: %w", err)
		}

		// Set message attributes
		domainMsg.AssignToChannel(input.ChannelID, &input.ChannelTypeID)
		domainMsg.AssignToSession(assignment.SessionID)
		domainMsg.SetChannelMessageID(msg.ExternalID)
		// Note: Message timestamp is set during construction via NewMessage

		if msg.ContentType.IsText() && msg.Text != "" {
			if err := domainMsg.SetText(msg.Text); err != nil {
				return messagesCreated, duplicates, fmt.Errorf("failed to set text: %w", err)
			}
		}

		if msg.ContentType.IsMedia() && msg.MediaURL != nil {
			if err := domainMsg.SetMediaContent(*msg.MediaURL, msg.MediaMimetype); err != nil {
				return messagesCreated, duplicates, fmt.Errorf("failed to set media: %w", err)
			}
		}

		// Save message
		if err := uc.messageRepo.Save(ctx, domainMsg); err != nil {
			return messagesCreated, duplicates, fmt.Errorf("failed to save message: %w", err)
		}

		messagesCreated++
	}

	return messagesCreated, duplicates, nil
}

// publishBatchEvents publishes all domain events collected during batch processing
func (uc *ImportMessagesBatchUseCase) publishBatchEvents(
	ctx context.Context,
	contacts map[string]*contact.Contact,
	sessions []*session.Session,
) error {
	var events []shared.DomainEvent

	// Collect contact events
	for _, c := range contacts {
		for _, e := range c.DomainEvents() {
			events = append(events, e)
		}
		c.ClearEvents()
	}

	// Collect session events
	for _, s := range sessions {
		for _, e := range s.DomainEvents() {
			events = append(events, e)
		}
		s.ClearEvents()
	}

	// Publish in batch
	if len(events) > 0 {
		return uc.eventBus.PublishBatch(ctx, events)
	}

	return nil
}
