package message

import "errors"

// ContentType representa o tipo de conteúdo da mensagem (modelo de domínio limpo).
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

// IsValid verifica se o tipo de conteúdo é válido.
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

// IsText verifica se é mensagem de texto.
func (ct ContentType) IsText() bool {
	return ct == ContentTypeText
}

// IsMedia verifica se é mensagem de mídia (qualquer tipo de arquivo).
func (ct ContentType) IsMedia() bool {
	switch ct {
	case ContentTypeImage, ContentTypeVideo, ContentTypeAudio, ContentTypeVoice,
		ContentTypeDocument, ContentTypeSticker:
		return true
	default:
		return false
	}
}

// IsSystem verifica se é mensagem do sistema.
func (ct ContentType) IsSystem() bool {
	return ct == ContentTypeSystem
}

// RequiresURL verifica se o tipo de conteúdo requer uma URL de arquivo.
func (ct ContentType) RequiresURL() bool {
	return ct.IsMedia()
}

// Status representa o status de entrega da mensagem.
type Status string

const (
	StatusQueued    Status = "queued"
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "read"
	StatusFailed    Status = "failed"
)

func (s Status) String() string {
	return string(s)
}

// ParseContentType converte string para ContentType.
func ParseContentType(s string) (ContentType, error) {
	ct := ContentType(s)
	if !ct.IsValid() {
		return "", errors.New("invalid content type")
	}
	return ct, nil
}

// ParseStatus converte string para Status.
func ParseStatus(s string) (Status, error) {
	status := Status(s)
	switch status {
	case StatusQueued, StatusSent, StatusDelivered, StatusRead, StatusFailed:
		return status, nil
	default:
		return "", errors.New("invalid status")
	}
}
