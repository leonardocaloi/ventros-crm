package whatsapp

import (
	"errors"

	"github.com/caloi/ventros-crm/internal/domain/crm/message"
)

// WhatsAppMessagePayload representa a estrutura da API do WhatsApp/Meta.
// Essa estrutura espelha o formato real da API externa.
type WhatsAppMessagePayload struct {
	Type     string                `json:"type"` // "text" ou "interactive" ou outros
	Text     *WhatsAppText         `json:"text,omitempty"`
	Image    *WhatsAppMedia        `json:"image,omitempty"`
	Video    *WhatsAppMedia        `json:"video,omitempty"`
	Audio    *WhatsAppMedia        `json:"audio,omitempty"`
	Document *WhatsAppMedia        `json:"document,omitempty"`
	Sticker  *WhatsAppMedia        `json:"sticker,omitempty"`
	Location *WhatsAppLocation     `json:"location,omitempty"`
	Contacts []WhatsAppContactCard `json:"contacts,omitempty"`
}

type WhatsAppText struct {
	Body string `json:"body"`
}

type WhatsAppMedia struct {
	ID       string `json:"id,omitempty"`
	Link     string `json:"link,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type WhatsAppLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type WhatsAppContactCard struct {
	Name  WhatsAppContactName    `json:"name"`
	Phone []WhatsAppContactPhone `json:"phones,omitempty"`
}

type WhatsAppContactName struct {
	FormattedName string `json:"formatted_name"`
	FirstName     string `json:"first_name,omitempty"`
}

type WhatsAppContactPhone struct {
	Phone string `json:"phone"`
	Type  string `json:"type,omitempty"`
}

// MessageAdapter adapta mensagens da API do WhatsApp para o modelo de domínio.
type MessageAdapter struct{}

// NewMessageAdapter cria um novo adapter.
func NewMessageAdapter() *MessageAdapter {
	return &MessageAdapter{}
}

// ToContentType converte o tipo do WhatsApp para ContentType do domínio.
// Aqui é onde isolamos a complexidade da API externa.
func (a *MessageAdapter) ToContentType(payload WhatsAppMessagePayload) (message.ContentType, error) {
	switch payload.Type {
	case "text":
		return message.ContentTypeText, nil
	case "image":
		return message.ContentTypeImage, nil
	case "video":
		return message.ContentTypeVideo, nil
	case "audio", "voice":
		return message.ContentTypeAudio, nil
	case "document":
		return message.ContentTypeDocument, nil
	case "sticker":
		return message.ContentTypeSticker, nil
	case "location":
		return message.ContentTypeLocation, nil
	case "contacts":
		return message.ContentTypeContact, nil
	default:
		return "", errors.New("unsupported WhatsApp message type: " + payload.Type)
	}
}

// ExtractText extrai o texto da mensagem do payload do WhatsApp.
func (a *MessageAdapter) ExtractText(payload WhatsAppMessagePayload) string {
	if payload.Text != nil {
		return payload.Text.Body
	}

	// Caption de mídia também é considerado texto
	if payload.Image != nil && payload.Image.Caption != "" {
		return payload.Image.Caption
	}
	if payload.Video != nil && payload.Video.Caption != "" {
		return payload.Video.Caption
	}
	if payload.Document != nil && payload.Document.Caption != "" {
		return payload.Document.Caption
	}

	return ""
}

// ExtractMediaURL extrai a URL da mídia do payload do WhatsApp.
func (a *MessageAdapter) ExtractMediaURL(payload WhatsAppMessagePayload) *string {
	var url string

	switch payload.Type {
	case "image":
		if payload.Image != nil {
			url = payload.Image.Link
		}
	case "video":
		if payload.Video != nil {
			url = payload.Video.Link
		}
	case "audio", "voice":
		if payload.Audio != nil {
			url = payload.Audio.Link
		}
	case "document":
		if payload.Document != nil {
			url = payload.Document.Link
		}
	case "sticker":
		if payload.Sticker != nil {
			url = payload.Sticker.Link
		}
	}

	if url == "" {
		return nil
	}
	return &url
}

// ExtractMimeType extrai o mime type da mídia do payload do WhatsApp.
func (a *MessageAdapter) ExtractMimeType(payload WhatsAppMessagePayload) *string {
	var mimeType string

	switch payload.Type {
	case "image":
		if payload.Image != nil {
			mimeType = payload.Image.MimeType
		}
	case "video":
		if payload.Video != nil {
			mimeType = payload.Video.MimeType
		}
	case "audio", "voice":
		if payload.Audio != nil {
			mimeType = payload.Audio.MimeType
		}
	case "document":
		if payload.Document != nil {
			mimeType = payload.Document.MimeType
		}
	case "sticker":
		if payload.Sticker != nil {
			mimeType = payload.Sticker.MimeType
		}
	}

	if mimeType == "" {
		return nil
	}
	return &mimeType
}

// Exemplo de uso:
//
// func HandleWhatsAppWebhook(payload WhatsAppMessagePayload) error {
//     adapter := NewMessageAdapter()
//
//     // 1. Converte tipo do WhatsApp → Domínio
//     contentType, err := adapter.ToContentType(payload)
//     if err != nil {
//         return err
//     }
//
//     // 2. Cria mensagem no domínio (modelo limpo)
//     msg, err := message.NewMessage(contactID, projectID, customerID, contentType, false)
//     if err != nil {
//         return err
//     }
//
//     // 3. Preenche conteúdo baseado no tipo
//     if contentType.IsText() {
//         text := adapter.ExtractText(payload)
//         msg.SetText(text)
//     } else if contentType.IsMedia() {
//         url := adapter.ExtractMediaURL(payload)
//         mimeType := adapter.ExtractMimeType(payload)
//         if url != nil && mimeType != nil {
//             msg.SetMediaContent(*url, *mimeType)
//         }
//     }
//
//     // 4. Salva no repositório
//     return messageRepo.Save(ctx, msg)
// }
