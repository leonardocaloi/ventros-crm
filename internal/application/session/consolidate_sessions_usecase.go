package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/crm/message"
	"github.com/ventros/crm/internal/domain/crm/session"
	"go.uber.org/zap"
)

// ConsolidateSessionsUseCase implements session consolidation following Clean Architecture
// Business Rule: Sessions from the same contact with gaps <= timeout should be consolidated
// This is critical for history imports where parallel processing creates fragmented sessions
type ConsolidateSessionsUseCase struct {
	sessionRepo session.Repository
	messageRepo message.Repository
	logger      *zap.Logger
}

// NewConsolidateSessionsUseCase creates a new instance
func NewConsolidateSessionsUseCase(
	sessionRepo session.Repository,
	messageRepo message.Repository,
	logger *zap.Logger,
) *ConsolidateSessionsUseCase {
	return &ConsolidateSessionsUseCase{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		logger:      logger,
	}
}

// ConsolidateInput represents the consolidation request
type ConsolidateInput struct {
	ChannelID             uuid.UUID // Channel to consolidate sessions for
	SessionTimeoutMinutes int       // Timeout duration to determine session boundaries
	BatchSize             int       // Number of sessions to process per batch (default: 5000)
}

// ConsolidateResult represents the consolidation outcome
type ConsolidateResult struct {
	ChannelID       uuid.UUID `json:"channel_id"`
	SessionsBefore  int64     `json:"sessions_before"`
	SessionsAfter   int64     `json:"sessions_after"`
	SessionsDeleted int64     `json:"sessions_deleted"`
	MessagesUpdated int64     `json:"messages_updated"`
	DurationSeconds float64   `json:"duration_seconds"`
}

// Execute performs session consolidation using pure domain logic
// Algorithm:
//  1. Load sessions in batches (ordered by contact_id, started_at)
//  2. Group sessions by contact_id
//  3. For each contact, identify sessions that should consolidate based on timeout gaps
//  4. Update messages to point to consolidated session (keep earliest session per group)
//  5. Delete orphaned sessions
func (uc *ConsolidateSessionsUseCase) Execute(ctx context.Context, input ConsolidateInput) (*ConsolidateResult, error) {
	startTime := time.Now()

	// Validate input
	if input.ChannelID == uuid.Nil {
		return nil, fmt.Errorf("channel_id is required")
	}
	if input.SessionTimeoutMinutes <= 0 {
		input.SessionTimeoutMinutes = 30 // Default: 30 minutes
	}
	if input.BatchSize <= 0 {
		input.BatchSize = 5000 // Default: 5000 sessions per batch
	}

	timeout := time.Duration(input.SessionTimeoutMinutes) * time.Minute

	uc.logger.Info("🔄 Starting session consolidation (Go pure implementation)",
		zap.String("channel_id", input.ChannelID.String()),
		zap.Int("timeout_minutes", input.SessionTimeoutMinutes),
		zap.Int("batch_size", input.BatchSize))

	// Count sessions before consolidation
	sessionsBefore, err := uc.sessionRepo.CountByChannel(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to count sessions before consolidation: %w", err)
	}

	uc.logger.Info("📊 Sessions before consolidation",
		zap.Int64("count", sessionsBefore))

	// 🔥 FIX: Process by CONTACT instead of by arbitrary session batches
	// Problem: BatchSize of 5000 sessions splits contacts across batches
	// Solution: Load ALL sessions for a limited number of contacts at a time
	totalMessagesUpdated := int64(0)
	orphanedSessionIDs := []uuid.UUID{}

	// Get unique contact IDs that have sessions in this channel
	contactIDs, err := uc.sessionRepo.GetContactIDsByChannel(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact IDs: %w", err)
	}

	uc.logger.Info("📊 Unique contacts with sessions",
		zap.Int("contact_count", len(contactIDs)))

	// Process contacts in batches (e.g., 100 contacts at a time)
	contactBatchSize := 100
	if input.BatchSize > 0 && input.BatchSize < 1000 {
		// If user specified small batch, use even smaller contact batches
		contactBatchSize = 50
	}

	for i := 0; i < len(contactIDs); i += contactBatchSize {
		end := i + contactBatchSize
		if end > len(contactIDs) {
			end = len(contactIDs)
		}
		contactBatch := contactIDs[i:end]

		uc.logger.Debug("Processing contact batch",
			zap.Int("batch_number", i/contactBatchSize+1),
			zap.Int("contacts_in_batch", len(contactBatch)))

		// Load ALL sessions for these contacts (ordered by contact_id, started_at)
		sessions, err := uc.sessionRepo.FindByChannelAndContacts(ctx, input.ChannelID, contactBatch)
		if err != nil {
			return nil, fmt.Errorf("failed to load sessions for contact batch: %w", err)
		}

		uc.logger.Debug("Sessions loaded for contact batch",
			zap.Int("session_count", len(sessions)))

		// Group sessions by contact_id
		sessionsByContact := make(map[uuid.UUID][]*session.Session)
		for _, sess := range sessions {
			contactID := sess.ContactID()
			sessionsByContact[contactID] = append(sessionsByContact[contactID], sess)
		}

		// Consolidate sessions for each contact
		for contactID, contactSessions := range sessionsByContact {
			if len(contactSessions) <= 1 {
				continue // Nothing to consolidate for this contact
			}

			// Identify session groups that should be consolidated
			consolidationGroups := uc.identifyConsolidationGroups(contactSessions, timeout)

			// Update messages for each consolidation group
			for _, group := range consolidationGroups {
				if len(group) <= 1 {
					continue // No consolidation needed
				}

				// Keep the earliest session as the consolidated session
				consolidatedSession := group[0] // Already sorted by started_at

				// Move messages from later sessions to consolidated session
				for i := 1; i < len(group); i++ {
					orphanSession := group[i]

					// Update all messages from orphan session to consolidated session
					updatedCount, err := uc.messageRepo.UpdateSessionIDForSession(
						ctx,
						orphanSession.ID(),
						consolidatedSession.ID(),
					)
					if err != nil {
						uc.logger.Error("Failed to update messages for session consolidation",
							zap.String("contact_id", contactID.String()),
							zap.String("orphan_session_id", orphanSession.ID().String()),
							zap.String("consolidated_session_id", consolidatedSession.ID().String()),
							zap.Error(err))
						continue
					}

					totalMessagesUpdated += updatedCount

					// Mark session for deletion
					orphanedSessionIDs = append(orphanedSessionIDs, orphanSession.ID())

					uc.logger.Debug("Consolidated session",
						zap.String("contact_id", contactID.String()),
						zap.String("orphan_session_id", orphanSession.ID().String()),
						zap.String("consolidated_session_id", consolidatedSession.ID().String()),
						zap.Int64("messages_moved", updatedCount))
				}
			}
		}
	}

	// Delete orphaned sessions
	sessionsDeleted := int64(0)
	if len(orphanedSessionIDs) > 0 {
		uc.logger.Info("🗑️ Deleting orphaned sessions",
			zap.Int("count", len(orphanedSessionIDs)))

		// Delete in batches to avoid huge IN clauses
		deleteBatchSize := 1000
		for i := 0; i < len(orphanedSessionIDs); i += deleteBatchSize {
			end := i + deleteBatchSize
			if end > len(orphanedSessionIDs) {
				end = len(orphanedSessionIDs)
			}

			batch := orphanedSessionIDs[i:end]
			if err := uc.sessionRepo.DeleteBatch(ctx, batch); err != nil {
				uc.logger.Error("Failed to delete orphaned sessions batch",
					zap.Int("batch_start", i),
					zap.Int("batch_size", len(batch)),
					zap.Error(err))
				continue
			}

			sessionsDeleted += int64(len(batch))
		}
	}

	// Count sessions after consolidation
	sessionsAfter, err := uc.sessionRepo.CountByChannel(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to count sessions after consolidation: %w", err)
	}

	duration := time.Since(startTime)

	result := &ConsolidateResult{
		ChannelID:       input.ChannelID,
		SessionsBefore:  sessionsBefore,
		SessionsAfter:   sessionsAfter,
		SessionsDeleted: sessionsDeleted,
		MessagesUpdated: totalMessagesUpdated,
		DurationSeconds: duration.Seconds(),
	}

	uc.logger.Info("✅ Session consolidation completed",
		zap.Int64("sessions_before", sessionsBefore),
		zap.Int64("sessions_after", sessionsAfter),
		zap.Int64("sessions_deleted", sessionsDeleted),
		zap.Int64("messages_updated", totalMessagesUpdated),
		zap.Float64("reduction_pct", float64(sessionsBefore-sessionsAfter)/float64(sessionsBefore)*100),
		zap.Float64("duration_seconds", duration.Seconds()))

	return result, nil
}

// identifyConsolidationGroups identifies which sessions should be consolidated
// Returns groups of sessions where each group should be merged into one session
// Groups are ordered by started_at (earliest first)
func (uc *ConsolidateSessionsUseCase) identifyConsolidationGroups(sessions []*session.Session, timeout time.Duration) [][]*session.Session {
	if len(sessions) <= 1 {
		return nil
	}

	// Sort sessions by started_at (already sorted by query, but ensure it)
	// sessions are already ordered by started_at from FindByChannelPaginated

	groups := [][]*session.Session{}
	currentGroup := []*session.Session{sessions[0]}

	for i := 1; i < len(sessions); i++ {
		prev := currentGroup[len(currentGroup)-1]
		curr := sessions[i]

		// ✅ Use domain logic to determine if should consolidate
		if curr.ShouldConsolidateWith(prev, timeout) {
			// Add to current group (sessions should consolidate)
			currentGroup = append(currentGroup, curr)
		} else {
			// Start new group (gap is too large)
			if len(currentGroup) > 1 {
				groups = append(groups, currentGroup)
			}
			currentGroup = []*session.Session{curr}
		}
	}

	// Add last group if it has consolidatable sessions
	if len(currentGroup) > 1 {
		groups = append(groups, currentGroup)
	}

	return groups
}
