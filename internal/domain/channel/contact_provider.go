package channel

import (
	"context"
	"time"
)

// ContactProvider is an abstraction for contact operations
//
// This interface allows different implementations:
// - WAHAContactProvider (for both "waha" and "whatsapp_business" channels)
// - MetaWhatsAppContactProvider (for Meta Business API)
// - TelegramContactProvider (for Telegram)
//
// The Channel aggregate uses this interface without knowing the concrete implementation.
type ContactProvider interface {
	// GetAllContacts returns all contacts from the channel
	//
	// Parameters:
	// - sortBy: "id" or "name"
	// - sortOrder: "asc" or "desc"
	// - limit, offset: pagination
	GetAllContacts(ctx context.Context, sortBy, sortOrder string, limit, offset int) ([]Contact, error)

	// GetContact gets basic contact information
	//
	// Always returns result even if phone is not registered in WhatsApp.
	// Use CheckExists to verify if number is registered.
	GetContact(ctx context.Context, contactID string) (*Contact, error)

	// CheckExists checks if phone number is registered in WhatsApp
	CheckExists(ctx context.Context, phoneNumber string) (*ContactExistence, error)

	// GetAbout gets contact's "about" status text
	//
	// Returns nil if you don't have permission to read their status.
	GetAbout(ctx context.Context, contactID string) (*string, error)

	// GetProfilePicture gets contact's profile picture URL
	//
	// If privacy settings don't allow, returns nil.
	// Set refresh=true to bypass 24h cache (use carefully - rate limits!)
	GetProfilePicture(ctx context.Context, contactID string, refresh bool) (*string, error)

	// BlockContact blocks a contact
	BlockContact(ctx context.Context, contactID string) error

	// UnblockContact unblocks a contact
	UnblockContact(ctx context.Context, contactID string) error

	// UpdateContact creates or updates contact in phone address book
	//
	// May not work if multiple WhatsApp apps installed on same phone.
	UpdateContact(ctx context.Context, contactID string, firstName, lastName string) error
}

// Contact represents basic contact information
type Contact struct {
	ID          string  `json:"id"`           // e.g., "11111111111@c.us"
	Name        string  `json:"name"`         // Display name
	PhoneNumber string  `json:"phone_number"` // Raw phone number
	PushName    *string `json:"push_name,omitempty"`
	ShortName   *string `json:"short_name,omitempty"`
	IsMe        bool    `json:"is_me"`
	IsWAContact bool    `json:"is_wa_contact"` // Is registered in WhatsApp
	IsBlocked   bool    `json:"is_blocked"`
}

// ContactExistence represents contact existence check result
type ContactExistence struct {
	PhoneNumber   string  `json:"phone_number"`
	NumberExists  bool    `json:"number_exists"`
	ChatID        *string `json:"chat_id,omitempty"` // Undefined if number doesn't exist
}

// ProfileManager is an abstraction for profile management operations
//
// This interface allows managing the channel's own profile (not other contacts).
// Different implementations:
// - WAHAProfileManager (for WAHA-based channels)
// - MetaWhatsAppProfileManager (for Meta Business API)
type ProfileManager interface {
	// GetProfile gets current profile information
	GetProfile(ctx context.Context) (*Profile, error)

	// SetProfileName sets profile display name
	SetProfileName(ctx context.Context, name string) error

	// SetProfileStatus sets profile status (About)
	SetProfileStatus(ctx context.Context, status string) error

	// SetProfilePicture sets profile picture
	//
	// file can be:
	// - URL (https://...)
	// - Base64 data (data:image/jpeg;base64,...)
	SetProfilePicture(ctx context.Context, file ProfilePictureFile) error

	// DeleteProfilePicture deletes current profile picture
	DeleteProfilePicture(ctx context.Context) error
}

// Profile represents channel's own profile
type Profile struct {
	ID      string  `json:"id"`       // e.g., "11111111111@c.us"
	Name    string  `json:"name"`     // Display name
	Status  *string `json:"status"`   // About text
	Picture *string `json:"picture"`  // Profile picture URL
}

// ProfilePictureFile represents a file to be uploaded as profile picture
type ProfilePictureFile struct {
	Mimetype string `json:"mimetype"` // e.g., "image/jpeg"
	Filename string `json:"filename"` // e.g., "profile.jpg"
	URL      string `json:"url,omitempty"` // URL or data URL
}

// ChatProvider is an abstraction for chat operations
//
// This interface allows fetching chats and their messages.
type ChatProvider interface {
	// GetChatsOverview returns overview of all chats
	//
	// Includes: chat id, name, picture, last message
	// Sorted by last message timestamp
	GetChatsOverview(ctx context.Context, limit, offset int, chatIDs []string) ([]ChatOverview, error)

	// GetChatMessages returns messages from a specific chat
	//
	// Parameters:
	// - downloadMedia: whether to download media files
	// - limit, offset: pagination
	// - filters: optional filters (timestamp, fromMe, ack status)
	GetChatMessages(ctx context.Context, chatID string, opts ChatMessagesOptions) ([]ChatMessage, error)

	// DeleteChat deletes a chat
	DeleteChat(ctx context.Context, chatID string) error

	// ArchiveChat archives a chat
	ArchiveChat(ctx context.Context, chatID string) error

	// UnarchiveChat unarchives a chat
	UnarchiveChat(ctx context.Context, chatID string) error

	// MarkChatAsUnread marks chat as unread
	MarkChatAsUnread(ctx context.Context, chatID string) error

	// ReadChatMessages marks messages in chat as read
	//
	// Parameters:
	// - messages: how many messages to read (latest first)
	// - days: how many days to read (latest first, default 7)
	ReadChatMessages(ctx context.Context, chatID string, messages *int, days *int) ([]string, error)
}

// ChatOverview represents a chat in overview list
type ChatOverview struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Picture     *string      `json:"picture,omitempty"`
	LastMessage *ChatMessage `json:"last_message,omitempty"`
}

// ChatMessage represents a message in a chat
type ChatMessage struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	From        string                 `json:"from"`
	FromMe      bool                   `json:"from_me"`
	To          string                 `json:"to"`
	Body        string                 `json:"body"`
	HasMedia    bool                   `json:"has_media"`
	MediaURL    *string                `json:"media_url,omitempty"`
	MimeType    *string                `json:"mime_type,omitempty"`
	Ack         int                    `json:"ack"` // -1=ERROR, 0=PENDING, 1=SERVER, 2=DEVICE, 3=READ, 4=PLAYED
	AckName     string                 `json:"ack_name"`
	ReplyTo     *string                `json:"reply_to,omitempty"`
	Location    *MessageLocation       `json:"location,omitempty"`
	VCards      []string               `json:"vcards,omitempty"`
	RawData     map[string]interface{} `json:"_data,omitempty"`
}

// MessageLocation represents location data in a message
type MessageLocation struct {
	Latitude    string  `json:"latitude"`
	Longitude   string  `json:"longitude"`
	Description *string `json:"description,omitempty"`
}

// ChatMessagesOptions represents options for fetching chat messages
type ChatMessagesOptions struct {
	DownloadMedia      bool
	Limit              int
	Offset             int
	TimestampLte       *int64  // Filter messages before this timestamp (inclusive)
	TimestampGte       *int64  // Filter messages after this timestamp (inclusive)
	FromMe             *bool   // Filter by fromMe
	AckStatus          *string // Filter by acknowledgment status
}
