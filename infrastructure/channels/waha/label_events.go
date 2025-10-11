package waha

// WAHALabelPayload represents label data from WAHA webhooks
type WAHALabelPayload struct {
	ID       string `json:"id"`       // Label ID
	Name     string `json:"name"`     // Label name
	Color    int    `json:"color"`    // Color as number (0-19)
	ColorHex string `json:"colorHex"` // Color as hex string (#RRGGBB)
}

// WAHALabelChatPayload represents label<->chat association from WAHA
type WAHALabelChatPayload struct {
	LabelID string `json:"labelId"` // Label ID
	ChatID  string `json:"chatId"`  // Chat ID (WhatsApp chat ID like 5511999999999@c.us or groupId@g.us)
}

// WAHALabelEvent represents a label.upsert or label.deleted event
type WAHALabelEvent struct {
	ID        string           `json:"id"`
	Timestamp int64            `json:"timestamp"`
	Event     string           `json:"event"` // "label.upsert" or "label.deleted"
	Session   string           `json:"session"`
	Payload   WAHALabelPayload `json:"payload"`
}

// WAHALabelChatEvent represents a label.chat.added or label.chat.deleted event
type WAHALabelChatEvent struct {
	ID        string               `json:"id"`
	Timestamp int64                `json:"timestamp"`
	Event     string               `json:"event"` // "label.chat.added" or "label.chat.deleted"
	Session   string               `json:"session"`
	Payload   WAHALabelChatPayload `json:"payload"`
}
