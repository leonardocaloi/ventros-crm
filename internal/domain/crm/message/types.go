package message

import "errors"

type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeImage    ContentType = "image"
	ContentTypeVideo    ContentType = "video"
	ContentTypeAudio    ContentType = "audio"
	ContentTypeVoice    ContentType = "voice"
	ContentTypeDocument ContentType = "document"
	ContentTypeLocation ContentType = "location"
	ContentTypeContact  ContentType = "contact"
	ContentTypeSticker  ContentType = "sticker"
	ContentTypeSystem   ContentType = "system"
)

func (ct ContentType) IsValid() bool {
	switch ct {
	case ContentTypeText, ContentTypeImage, ContentTypeVideo, ContentTypeAudio,
		ContentTypeVoice, ContentTypeDocument, ContentTypeLocation, ContentTypeContact,
		ContentTypeSticker, ContentTypeSystem:
		return true
	default:
		return false
	}
}

func (ct ContentType) String() string {
	return string(ct)
}

func (ct ContentType) IsText() bool {
	return ct == ContentTypeText
}

func (ct ContentType) IsMedia() bool {
	switch ct {
	case ContentTypeImage, ContentTypeVideo, ContentTypeAudio, ContentTypeVoice,
		ContentTypeDocument, ContentTypeSticker:
		return true
	default:
		return false
	}
}

func (ct ContentType) IsSystem() bool {
	return ct == ContentTypeSystem
}

func (ct ContentType) RequiresURL() bool {
	return ct.IsMedia()
}

type Status string

const (
	StatusQueued    Status = "queued"
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "read"
	StatusPlayed    Status = "played" // SOMENTE para mensagens de voz/áudio reproduzidas (ACK 4)
	StatusFailed    Status = "failed"
)

func (s Status) String() string {
	return string(s)
}

func ParseContentType(s string) (ContentType, error) {
	ct := ContentType(s)
	if !ct.IsValid() {
		return "", errors.New("invalid content type")
	}
	return ct, nil
}

func ParseStatus(s string) (Status, error) {
	status := Status(s)
	switch status {
	case StatusQueued, StatusSent, StatusDelivered, StatusRead, StatusPlayed, StatusFailed:
		return status, nil
	default:
		return "", errors.New("invalid status")
	}
}
