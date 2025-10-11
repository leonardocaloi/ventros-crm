package chat

import (
	"time"

	"github.com/google/uuid"
)

type Participant struct {
	ID       uuid.UUID       // Contact ID or Agent ID
	Type     ParticipantType // contact or agent
	JoinedAt time.Time       // When joined the chat
	LeftAt   *time.Time      // When left (for groups/channels)
	IsAdmin  bool            // Is admin/moderator (for groups)
}

type ParticipantType string

const (
	ParticipantTypeContact ParticipantType = "contact"
	ParticipantTypeAgent   ParticipantType = "agent"
)

func (pt ParticipantType) IsValid() bool {
	switch pt {
	case ParticipantTypeContact, ParticipantTypeAgent:
		return true
	default:
		return false
	}
}

func (pt ParticipantType) String() string {
	return string(pt)
}

func ParseParticipantType(s string) (ParticipantType, error) {
	pt := ParticipantType(s)
	if !pt.IsValid() {
		return "", ErrInvalidChatType // Reuse existing error
	}
	return pt, nil
}
