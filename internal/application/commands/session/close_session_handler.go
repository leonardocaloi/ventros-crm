package session

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/crm/session"
)

// CloseSessionHandler handler para o comando CloseSession
type CloseSessionHandler struct {
	repository session.Repository
	logger     *logrus.Logger
}

// NewCloseSessionHandler cria uma nova instância do handler
func NewCloseSessionHandler(repository session.Repository, logger *logrus.Logger) *CloseSessionHandler {
	return &CloseSessionHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de encerramento de sessão
func (h *CloseSessionHandler) Handle(ctx context.Context, cmd CloseSessionCommand) (*session.Session, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid CloseSession command")
		return nil, err
	}

	// Find session
	sess, err := h.repository.FindByID(ctx, cmd.SessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", cmd.SessionID).Error("Session not found")
		return nil, fmt.Errorf("%w: %v", ErrSessionNotFound, err)
	}

	if sess == nil {
		h.logger.WithField("session_id", cmd.SessionID).Warn("Session not found (nil returned)")
		return nil, ErrSessionNotFound
	}

	// Validate if session is already ended
	if sess.Status() == session.StatusEnded {
		h.logger.WithField("session_id", cmd.SessionID).Warn("Session is already ended")
		return nil, ErrSessionAlreadyEnded
	}

	// Close session based on reason
	switch cmd.Reason {
	case "resolved":
		if err := sess.Resolve(); err != nil {
			h.logger.WithError(err).Error("Failed to resolve session")
			return nil, fmt.Errorf("%w: %v", ErrSessionCloseFailed, err)
		}
	case "escalated":
		if err := sess.Escalate(); err != nil {
			h.logger.WithError(err).Error("Failed to escalate session")
			return nil, fmt.Errorf("%w: %v", ErrSessionCloseFailed, err)
		}
	case "transferred", "agent_closed":
		// End session with custom reason
		if err := sess.End(session.EndReason(cmd.Reason)); err != nil {
			h.logger.WithError(err).Error("Failed to end session")
			return nil, fmt.Errorf("%w: %v", ErrSessionCloseFailed, err)
		}
	default:
		// This should never happen due to Validate(), but keeping for safety
		return nil, ErrInvalidReason
	}

	// Save to repository
	if err := h.repository.Save(ctx, sess); err != nil {
		h.logger.WithError(err).Error("Failed to save closed session")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"session_id": sess.ID(),
		"reason":     cmd.Reason,
		"status":     sess.Status(),
	}).Info("Session closed successfully")

	return sess, nil
}
