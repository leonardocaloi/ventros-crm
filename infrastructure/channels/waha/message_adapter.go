package waha

import (
	"fmt"
	"strings"

	"github.com/ventros/crm/internal/domain/crm/message"
)

// WAHAMessageEvent representa o evento completo que chega do WAHA via webhook/RabbitMQ.
// Esta é a estrutura externa que precisa ser isolada do domínio.
type WAHAMessageEvent struct {
	ID          string            `json:"id"`
	Timestamp   int64             `json:"timestamp"`
	Event       string            `json:"event"` // "message", "message.any", "message.ack", etc
	Session     string            `json:"session"`
	Metadata    map[string]string `json:"metadata"`
	Me          WAHAMe            `json:"me"`
	Payload     WAHAPayload       `json:"payload"`
	Environment WAHAEnvironment   `json:"environment"`
}

type WAHAEnvironment struct {
	Version string  `json:"version"`
	Engine  string  `json:"engine"`
	Tier    string  `json:"tier"`
	Browser *string `json:"browser"`
}

type WAHAMe struct {
	ID       string `json:"id"`
	PushName string `json:"pushName"`
	LID      string `json:"lid"`
	JID      string `json:"jid"`
}

type WAHAPayload struct {
	ID           string      `json:"id"`
	Timestamp    int64       `json:"timestamp"`
	From         string      `json:"from"`
	FromMe       bool        `json:"fromMe"`
	Source       string      `json:"source"` // "app", "web", etc
	To           string      `json:"to"`
	Participant  *string     `json:"participant"`  // Para grupos: quem ENVIOU a mensagem
	MentionedJid []string    `json:"mentionedJid"` // Quem foi MENCIONADO (@marcado)
	HasMedia     bool        `json:"hasMedia"`
	Media        *WAHAMedia  `json:"media"`
	Body         *string     `json:"body"`    // Texto da mensagem
	ReplyTo      interface{} `json:"replyTo"` // Pode ser string ou objeto
	Data         WAHAData    `json:"_data"`
}

type WAHAMedia struct {
	URL      string  `json:"url"`
	Mimetype string  `json:"mimetype"`
	Filename string  `json:"filename,omitempty"`
	S3       *WAHAS3 `json:"s3"`
}

type WAHAS3 struct {
	Bucket string `json:"Bucket"`
	Key    string `json:"Key"`
}

type WAHAData struct {
	Info    WAHAInfo    `json:"Info"`
	Message WAHAMessage `json:"Message"`
}

type WAHAInfo struct {
	Chat      string `json:"Chat"`
	Sender    string `json:"Sender"`
	IsFromMe  bool   `json:"IsFromMe"`
	IsGroup   bool   `json:"IsGroup"`
	SenderAlt string `json:"SenderAlt"` // LID alternativo
	ID        string `json:"ID"`
	Type      string `json:"Type"` // "text", "media"
	PushName  string `json:"PushName"`
	MediaType string `json:"MediaType"` // "image", "video", "audio", etc
}

type WAHAMessage struct {
	// Diferentes tipos de mensagens
	Conversation    *string              `json:"conversation"` // Texto simples
	ImageMessage    *WAHAMediaMessage    `json:"imageMessage"`
	VideoMessage    *WAHAMediaMessage    `json:"videoMessage"`
	AudioMessage    *WAHAMediaMessage    `json:"audioMessage"`
	DocumentMessage *WAHAMediaMessage    `json:"documentMessage"`
	StickerMessage  *WAHAMediaMessage    `json:"stickerMessage"`
	LocationMessage *WAHALocationMessage `json:"locationMessage"`
	ContactMessage  *WAHAContactMessage  `json:"contactMessage"`
	ExtendedTextMsg *WAHAExtendedText    `json:"extendedTextMessage"`
}

type WAHAMediaMessage struct {
	URL      string `json:"URL"`
	Mimetype string `json:"mimetype"`
	Caption  string `json:"caption,omitempty"`
	PTT      bool   `json:"PTT,omitempty"` // Push-to-Talk (voice message)
	FileName string `json:"fileName,omitempty"`
}

type WAHAExtendedText struct {
	Text        string           `json:"text"`
	ContextInfo *WAHAContextInfo `json:"contextInfo"`
}

type WAHAContextInfo struct {
	// Tracking de conversão (ads, etc)
	ConversionSource                   string               `json:"conversionSource,omitempty"`
	ConversionData                     string               `json:"conversionData,omitempty"`
	EntryPointConversionSource         string               `json:"entryPointConversionSource,omitempty"`
	EntryPointConversionApp            string               `json:"entryPointConversionApp,omitempty"`
	EntryPointConversionExternalSource string               `json:"entryPointConversionExternalSource,omitempty"`
	EntryPointConversionExternalMedium string               `json:"entryPointConversionExternalMedium,omitempty"`
	ExternalAdReply                    *WAHAExternalAdReply `json:"externalAdReply,omitempty"`
}

type WAHAExternalAdReply struct {
	SourceType string `json:"sourceType"`
	SourceID   string `json:"sourceID"`
	SourceApp  string `json:"sourceApp"`
	SourceURL  string `json:"sourceURL"`
	CTWAClid   string `json:"ctwaClid"` // Click ID do Click-to-WhatsApp Ad
}

type WAHALocationMessage struct {
	DegreesLatitude  float64 `json:"degreesLatitude"`
	DegreesLongitude float64 `json:"degreesLongitude"`
	Name             string  `json:"name,omitempty"`
	Address          string  `json:"address,omitempty"`
}

type WAHAContactMessage struct {
	DisplayName string `json:"displayName"`
	VCard       string `json:"vcard"`
}

// MessageAdapter adapta eventos do WAHA para o modelo de domínio limpo.
type MessageAdapter struct{}

// NewMessageAdapter cria um novo adapter para WAHA.
func NewMessageAdapter() *MessageAdapter {
	return &MessageAdapter{}
}

// ToContentType converte o tipo do WAHA para ContentType do domínio.
// Isola a complexidade da estrutura externa.
func (a *MessageAdapter) ToContentType(event WAHAMessageEvent) (message.ContentType, error) {
	payload := event.Payload
	info := payload.Data.Info
	msg := payload.Data.Message

	// 1. Tenta determinar pelo campo Type do Info (mais confiável)
	if info.Type == "text" {
		return message.ContentTypeText, nil
	}

	// 2. Verifica estruturas específicas de mensagem
	if msg.Conversation != nil {
		return message.ContentTypeText, nil
	}
	if msg.ImageMessage != nil {
		return message.ContentTypeImage, nil
	}
	if msg.VideoMessage != nil {
		return message.ContentTypeVideo, nil
	}
	if msg.AudioMessage != nil {
		// Verificar se é PTT (Push-to-Talk)
		if a.isPTT(event) {
			return message.ContentTypeVoice, nil
		}
		return message.ContentTypeAudio, nil
	}
	if msg.DocumentMessage != nil {
		return message.ContentTypeDocument, nil
	}
	if msg.LocationMessage != nil {
		return message.ContentTypeLocation, nil
	}
	if msg.ContactMessage != nil {
		return message.ContentTypeContact, nil
	}
	if msg.ExtendedTextMsg != nil {
		return message.ContentTypeText, nil
	}

	// 3. Fallback: usa MediaType do Info se disponível
	if info.MediaType != "" {
		switch info.MediaType {
		case "image":
			return message.ContentTypeImage, nil
		case "video":
			return message.ContentTypeVideo, nil
		case "audio":
			return message.ContentTypeAudio, nil
		case "ptt": // Push-to-Talk (voice message)
			return message.ContentTypeVoice, nil
		case "document":
			return message.ContentTypeDocument, nil
		case "vcard", "contact":
			return message.ContentTypeContact, nil
		}
	}

	return "", fmt.Errorf("unsupported content type: unsupported media type: %s", info.MediaType)
}

// isPTT verifica se o áudio é do tipo PTT (Push-to-Talk / gravação de voz).
func (a *MessageAdapter) isPTT(event WAHAMessageEvent) bool {
	// 1. Verifica pelo MediaType do Info
	if event.Payload.Data.Info.MediaType == "ptt" {
		return true
	}

	// 2. Verifica pelo campo PTT na estrutura AudioMessage
	msg := event.Payload.Data.Message
	if msg.AudioMessage != nil && msg.AudioMessage.PTT {
		return true
	}

	return false
}

// ExtractText extrai o texto da mensagem do evento WAHA.
func (a *MessageAdapter) ExtractText(event WAHAMessageEvent) string {
	payload := event.Payload
	msg := payload.Data.Message

	// 1. Texto direto do body
	if payload.Body != nil {
		return *payload.Body
	}

	// 2. Conversation simples
	if msg.Conversation != nil {
		return *msg.Conversation
	}

	// 3. Extended text message
	if msg.ExtendedTextMsg != nil {
		return msg.ExtendedTextMsg.Text
	}

	// 4. Caption de mídia
	if msg.ImageMessage != nil && msg.ImageMessage.Caption != "" {
		return msg.ImageMessage.Caption
	}
	if msg.VideoMessage != nil && msg.VideoMessage.Caption != "" {
		return msg.VideoMessage.Caption
	}
	if msg.DocumentMessage != nil && msg.DocumentMessage.Caption != "" {
		return msg.DocumentMessage.Caption
	}

	return ""
}

// ExtractMediaURL extrai a URL da mídia do evento WAHA.
func (a *MessageAdapter) ExtractMediaURL(event WAHAMessageEvent) *string {
	payload := event.Payload

	// 1. Se tem hasMedia e media.url, usa diretamente
	if payload.HasMedia && payload.Media != nil && payload.Media.URL != "" {
		return &payload.Media.URL
	}

	// 2. Tenta extrair da estrutura interna
	msg := payload.Data.Message

	if msg.ImageMessage != nil && msg.ImageMessage.URL != "" {
		return &msg.ImageMessage.URL
	}
	if msg.VideoMessage != nil && msg.VideoMessage.URL != "" {
		return &msg.VideoMessage.URL
	}
	if msg.AudioMessage != nil && msg.AudioMessage.URL != "" {
		return &msg.AudioMessage.URL
	}
	if msg.DocumentMessage != nil && msg.DocumentMessage.URL != "" {
		return &msg.DocumentMessage.URL
	}
	if msg.StickerMessage != nil && msg.StickerMessage.URL != "" {
		return &msg.StickerMessage.URL
	}

	return nil
}

// ExtractMimeType extrai o mimetype da mídia do evento WAHA.
func (a *MessageAdapter) ExtractMimeType(event WAHAMessageEvent) *string {
	payload := event.Payload

	// 1. Se tem hasMedia e media.mimetype, usa diretamente
	if payload.HasMedia && payload.Media != nil && payload.Media.Mimetype != "" {
		return &payload.Media.Mimetype
	}

	// 2. Tenta extrair da estrutura interna
	msg := payload.Data.Message

	if msg.ImageMessage != nil && msg.ImageMessage.Mimetype != "" {
		return &msg.ImageMessage.Mimetype
	}
	if msg.VideoMessage != nil && msg.VideoMessage.Mimetype != "" {
		return &msg.VideoMessage.Mimetype
	}
	if msg.AudioMessage != nil && msg.AudioMessage.Mimetype != "" {
		return &msg.AudioMessage.Mimetype
	}
	if msg.DocumentMessage != nil && msg.DocumentMessage.Mimetype != "" {
		return &msg.DocumentMessage.Mimetype
	}
	if msg.StickerMessage != nil && msg.StickerMessage.Mimetype != "" {
		return &msg.StickerMessage.Mimetype
	}

	return nil
}

// ExtractContactPhone extrai o número de telefone do contato (limpo).
func (a *MessageAdapter) ExtractContactPhone(event WAHAMessageEvent) string {
	from := event.Payload.From

	// Remove sufixos do WhatsApp (@c.us, @s.whatsapp.net, @lid)
	phone := from
	suffixes := []string{"@c.us", "@s.whatsapp.net", "@lid"}

	for _, suffix := range suffixes {
		if len(phone) > len(suffix) && phone[len(phone)-len(suffix):] == suffix {
			phone = phone[:len(phone)-len(suffix)]
			break
		}
	}

	return phone
}

// ExtractTrackingData extrai dados de rastreamento (ads, conversões).
func (a *MessageAdapter) ExtractTrackingData(event WAHAMessageEvent) map[string]string {
	tracking := make(map[string]string)

	msg := event.Payload.Data.Message
	if msg.ExtendedTextMsg != nil && msg.ExtendedTextMsg.ContextInfo != nil {
		ctx := msg.ExtendedTextMsg.ContextInfo

		if ctx.EntryPointConversionSource != "" {
			tracking["conversion_source"] = ctx.EntryPointConversionSource
		}
		if ctx.EntryPointConversionApp != "" {
			tracking["conversion_app"] = ctx.EntryPointConversionApp
		}
		if ctx.EntryPointConversionExternalSource != "" {
			tracking["external_source"] = ctx.EntryPointConversionExternalSource
		}
		if ctx.EntryPointConversionExternalMedium != "" {
			tracking["external_medium"] = ctx.EntryPointConversionExternalMedium
		}
		if ctx.ConversionData != "" {
			tracking["conversion_data"] = ctx.ConversionData
		}

		if ctx.ExternalAdReply != nil {
			tracking["ad_source_type"] = ctx.ExternalAdReply.SourceType
			tracking["ad_source_id"] = ctx.ExternalAdReply.SourceID
			tracking["ad_source_app"] = ctx.ExternalAdReply.SourceApp
			tracking["ad_source_url"] = ctx.ExternalAdReply.SourceURL
			tracking["ctwa_clid"] = ctx.ExternalAdReply.CTWAClid
		}
	}

	return tracking
}

// IsFromAd verifica se a mensagem veio de um anúncio.
func (a *MessageAdapter) IsFromAd(event WAHAMessageEvent) bool {
	msg := event.Payload.Data.Message
	if msg.ExtendedTextMsg != nil && msg.ExtendedTextMsg.ContextInfo != nil {
		ctx := msg.ExtendedTextMsg.ContextInfo
		return ctx.EntryPointConversionSource != "" || ctx.ExternalAdReply != nil
	}
	return false
}

// ExtractLocationData extrai dados de localização da mensagem.
func (a *MessageAdapter) ExtractLocationData(event WAHAMessageEvent) map[string]interface{} {
	msg := event.Payload.Data.Message
	if msg.LocationMessage != nil {
		data := make(map[string]interface{})
		data["latitude"] = msg.LocationMessage.DegreesLatitude
		data["longitude"] = msg.LocationMessage.DegreesLongitude
		if msg.LocationMessage.Name != "" {
			data["name"] = msg.LocationMessage.Name
		}
		if msg.LocationMessage.Address != "" {
			data["address"] = msg.LocationMessage.Address
		}
		return data
	}
	return nil
}

// ExtractContactData extrai dados de contato (vCard) da mensagem.
func (a *MessageAdapter) ExtractContactData(event WAHAMessageEvent) map[string]interface{} {
	msg := event.Payload.Data.Message
	if msg.ContactMessage != nil {
		data := make(map[string]interface{})
		data["display_name"] = msg.ContactMessage.DisplayName
		data["vcard"] = msg.ContactMessage.VCard
		return data
	}
	return nil
}

// ExtractFileName extrai o nome do arquivo de documentos.
func (a *MessageAdapter) ExtractFileName(event WAHAMessageEvent) string {
	payload := event.Payload

	// 1. Tenta do campo media.filename
	if payload.HasMedia && payload.Media != nil && payload.Media.Filename != "" {
		return payload.Media.Filename
	}

	// 2. Tenta da estrutura interna
	msg := payload.Data.Message
	if msg.DocumentMessage != nil && msg.DocumentMessage.FileName != "" {
		return msg.DocumentMessage.FileName
	}

	return ""
}

// IsGroupMessage verifica se a mensagem é de um grupo do WhatsApp.
// Grupos no WhatsApp têm ID terminando com "@g.us"
func (a *MessageAdapter) IsGroupMessage(event WAHAMessageEvent) bool {
	from := event.Payload.From
	// Grupos terminam com @g.us
	return len(from) > 5 && from[len(from)-5:] == "@g.us"
}

// ExtractParticipant extrai o ID do participante que enviou a mensagem em um grupo.
// Em grupos, o campo "from" é o ID do grupo, e "participant" é quem enviou.
// Retorna o participant limpo (sem sufixos do WhatsApp).
func (a *MessageAdapter) ExtractParticipant(event WAHAMessageEvent) string {
	participant := event.Payload.Participant
	if participant == nil || *participant == "" {
		return ""
	}

	// Remove sufixos do WhatsApp (@c.us, @s.whatsapp.net, @lid)
	phone := *participant
	suffixes := []string{"@c.us", "@s.whatsapp.net", "@lid"}

	for _, suffix := range suffixes {
		if len(phone) > len(suffix) && phone[len(phone)-len(suffix):] == suffix {
			phone = phone[:len(phone)-len(suffix)]
			break
		}
	}

	return phone
}

// ExtractMentions extrai os IDs dos usuários mencionados (@marcados) na mensagem.
// Retorna array de IDs no formato WAHA (ex: "5511999998888@c.us")
func (a *MessageAdapter) ExtractMentions(event WAHAMessageEvent) []string {
	// Retorna diretamente o array de MentionedJid do payload
	return event.Payload.MentionedJid
}

// ExtractGroupID extrai o ID externo do grupo (formato: "123456789@g.us")
func (a *MessageAdapter) ExtractGroupID(event WAHAMessageEvent) string {
	if !a.IsGroupMessage(event) {
		return ""
	}
	return event.Payload.From
}

// InferContentTypeFromPayload infere o tipo de conteúdo baseado em MimeType e HasMedia
// ✅ SOLID: MessageAdapter é a ÚNICA fonte de verdade para conversão de tipos
// Usado por: Import History (MessagePayload) - estrutura simples da API WAHA
// Similar a: ToContentType() - mas para estrutura diferente (webhook vs API)
func (a *MessageAdapter) InferContentTypeFromPayload(payload MessagePayload) message.ContentType {
	// 1. Se não tem mídia, é texto
	if !payload.HasMedia || payload.MimeType == "" {
		return message.ContentTypeText
	}

	// 2. Reaproveitar lógica de inferência por MimeType
	return a.inferFromMimeType(payload.MimeType)
}

// inferFromMimeType infere o ContentType baseado no MimeType
// ✅ DRY: Lógica compartilhada entre ToContentType (webhook) e InferContentTypeFromPayload (API)
func (a *MessageAdapter) inferFromMimeType(mimeType string) message.ContentType {
	// Images
	if strings.HasPrefix(mimeType, "image/") {
		if strings.Contains(mimeType, "webp") {
			return message.ContentTypeSticker // WhatsApp stickers são webp
		}
		return message.ContentTypeImage
	}

	// Videos
	if strings.HasPrefix(mimeType, "video/") {
		return message.ContentTypeVideo
	}

	// Audios
	if strings.HasPrefix(mimeType, "audio/") {
		// PTT (Push-to-Talk) detectado por outras fontes (Info.MediaType)
		// Aqui tratamos como audio genérico
		return message.ContentTypeAudio
	}

	// Documents
	if strings.HasPrefix(mimeType, "application/") {
		return message.ContentTypeDocument
	}

	// VCard (contatos)
	if strings.Contains(mimeType, "vcard") || strings.Contains(mimeType, "contact") {
		return message.ContentTypeContact
	}

	// Fallback: documento para qualquer outro tipo com mídia
	return message.ContentTypeDocument
}
