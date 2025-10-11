package chat

type ChatType string

const (
	ChatTypeIndividual ChatType = "individual" // 1-on-1 chat
	ChatTypeGroup      ChatType = "group"      // WhatsApp group, Telegram group
	ChatTypeChannel    ChatType = "channel"    // Telegram channel, WhatsApp Business broadcast
)

func (ct ChatType) IsValid() bool {
	switch ct {
	case ChatTypeIndividual, ChatTypeGroup, ChatTypeChannel:
		return true
	default:
		return false
	}
}

func (ct ChatType) String() string {
	return string(ct)
}

func ParseChatType(s string) (ChatType, error) {
	ct := ChatType(s)
	if !ct.IsValid() {
		return "", ErrInvalidChatType
	}
	return ct, nil
}
