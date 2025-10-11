package chat

type ChatStatus string

const (
	ChatStatusActive   ChatStatus = "active"   // Active conversation
	ChatStatusArchived ChatStatus = "archived" // Archived (hidden but can be reopened)
	ChatStatusClosed   ChatStatus = "closed"   // Closed (permanent - historical only)
)

func (cs ChatStatus) IsValid() bool {
	switch cs {
	case ChatStatusActive, ChatStatusArchived, ChatStatusClosed:
		return true
	default:
		return false
	}
}

func (cs ChatStatus) String() string {
	return string(cs)
}

func ParseChatStatus(s string) (ChatStatus, error) {
	cs := ChatStatus(s)
	if !cs.IsValid() {
		return "", ErrInvalidChatStatus
	}
	return cs, nil
}
