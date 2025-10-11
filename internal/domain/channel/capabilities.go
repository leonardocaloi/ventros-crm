package channel

import (
	"github.com/caloi/ventros-crm/internal/domain/shared"
)

// ChannelCapabilities defines what features a channel type supports
type ChannelCapabilities interface {
	// SupportsLabels indica se o canal suporta labels/tags nativamente
	SupportsLabels() bool

	// SupportsBidirectionalLabelSync indica se labels podem ser sincronizadas bidireccionalmente
	SupportsBidirectionalLabelSync() bool

	// SupportsCustomFields indica se o canal suporta custom fields
	SupportsCustomFields() bool

	// GetSystemFields retorna os custom fields obrigatórios e fixos para este tipo de canal
	// Estes campos são criados automaticamente e não podem ser removidos
	GetSystemFields() []*ChannelCustomFieldDefinition

	// GetAvailableCustomFieldTypes retorna os tipos de custom fields suportados por este canal
	GetAvailableCustomFieldTypes() []shared.FieldType

	// SupportsGroups indica se o canal suporta mensagens de grupos
	SupportsGroups() bool

	// SupportsTracking indica se o canal suporta rastreamento de origem de mensagens
	SupportsTracking() bool

	// SupportsAI indica se o canal suporta processamento de IA
	SupportsAI() bool

	// SupportsMedia indica se o canal suporta envio/recebimento de mídia
	SupportsMedia() bool

	// SupportedMediaTypes retorna os tipos de mídia suportados
	SupportedMediaTypes() []string
}

// ChannelCustomFieldDefinition define um custom field de canal
type ChannelCustomFieldDefinition struct {
	Key         string            // Chave única do campo
	Type        shared.FieldType  // Tipo do campo (text, label, etc)
	Required    bool              // Se é obrigatório
	Fixed       bool              // Se não pode ser removido
	System      bool              // Se é um campo de sistema (auto-criado)
	Description string            // Descrição do campo
	DefaultValue interface{}      // Valor padrão (opcional)
	Options     []string          // Opções para select/multi_select (opcional)
}

// WAHAChannelCapabilities implementa capabilities para canais WAHA
type WAHAChannelCapabilities struct{}

func (w *WAHAChannelCapabilities) SupportsLabels() bool {
	return true
}

func (w *WAHAChannelCapabilities) SupportsBidirectionalLabelSync() bool {
	return true
}

func (w *WAHAChannelCapabilities) SupportsCustomFields() bool {
	return true
}

func (w *WAHAChannelCapabilities) GetSystemFields() []*ChannelCustomFieldDefinition {
	return []*ChannelCustomFieldDefinition{
		{
			Key:         "labels",
			Type:        shared.FieldTypeLabel,
			Required:    true,
			Fixed:       true,
			System:      true,
			Description: "WhatsApp labels synchronized from WAHA",
		},
	}
}

func (w *WAHAChannelCapabilities) GetAvailableCustomFieldTypes() []shared.FieldType {
	return []shared.FieldType{
		shared.FieldTypeText,
		shared.FieldTypeNumber,
		shared.FieldTypeBoolean,
		shared.FieldTypeDate,
		shared.FieldTypeLabel,
		shared.FieldTypeSelect,
		shared.FieldTypeMultiSelect,
	}
}

func (w *WAHAChannelCapabilities) SupportsGroups() bool {
	return true
}

func (w *WAHAChannelCapabilities) SupportsTracking() bool {
	return true
}

func (w *WAHAChannelCapabilities) SupportsAI() bool {
	return true
}

func (w *WAHAChannelCapabilities) SupportsMedia() bool {
	return true
}

func (w *WAHAChannelCapabilities) SupportedMediaTypes() []string {
	return []string{"image", "video", "audio", "document", "voice"}
}

// MessengerChannelCapabilities implementa capabilities para Messenger
type MessengerChannelCapabilities struct{}

func (m *MessengerChannelCapabilities) SupportsLabels() bool {
	return true // Via API customizada
}

func (m *MessengerChannelCapabilities) SupportsBidirectionalLabelSync() bool {
	return false // Apenas via API, não automático
}

func (m *MessengerChannelCapabilities) SupportsCustomFields() bool {
	return true
}

func (m *MessengerChannelCapabilities) GetSystemFields() []*ChannelCustomFieldDefinition {
	return []*ChannelCustomFieldDefinition{} // Nenhum campo obrigatório
}

func (m *MessengerChannelCapabilities) GetAvailableCustomFieldTypes() []shared.FieldType {
	return []shared.FieldType{
		shared.FieldTypeText,
		shared.FieldTypeLabel,
		shared.FieldTypeSelect,
		shared.FieldTypeMultiSelect,
	}
}

func (m *MessengerChannelCapabilities) SupportsGroups() bool {
	return false
}

func (m *MessengerChannelCapabilities) SupportsTracking() bool {
	return true
}

func (m *MessengerChannelCapabilities) SupportsAI() bool {
	return true
}

func (m *MessengerChannelCapabilities) SupportsMedia() bool {
	return true
}

func (m *MessengerChannelCapabilities) SupportedMediaTypes() []string {
	return []string{"image", "video", "audio", "file"}
}

// InstagramChannelCapabilities implementa capabilities para Instagram
type InstagramChannelCapabilities struct{}

func (i *InstagramChannelCapabilities) SupportsLabels() bool {
	return true // Beta
}

func (i *InstagramChannelCapabilities) SupportsBidirectionalLabelSync() bool {
	return false
}

func (i *InstagramChannelCapabilities) SupportsCustomFields() bool {
	return true
}

func (i *InstagramChannelCapabilities) GetSystemFields() []*ChannelCustomFieldDefinition {
	return []*ChannelCustomFieldDefinition{}
}

func (i *InstagramChannelCapabilities) GetAvailableCustomFieldTypes() []shared.FieldType {
	return []shared.FieldType{
		shared.FieldTypeText,
		shared.FieldTypeLabel,
		shared.FieldTypeSelect,
	}
}

func (i *InstagramChannelCapabilities) SupportsGroups() bool {
	return false
}

func (i *InstagramChannelCapabilities) SupportsTracking() bool {
	return true
}

func (i *InstagramChannelCapabilities) SupportsAI() bool {
	return true
}

func (i *InstagramChannelCapabilities) SupportsMedia() bool {
	return true
}

func (i *InstagramChannelCapabilities) SupportedMediaTypes() []string {
	return []string{"image", "video"}
}

// WeChatChannelCapabilities implementa capabilities para WeChat
type WeChatChannelCapabilities struct{}

func (w *WeChatChannelCapabilities) SupportsLabels() bool {
	return true // Tags + descriptions
}

func (w *WeChatChannelCapabilities) SupportsBidirectionalLabelSync() bool {
	return false
}

func (w *WeChatChannelCapabilities) SupportsCustomFields() bool {
	return true
}

func (w *WeChatChannelCapabilities) GetSystemFields() []*ChannelCustomFieldDefinition {
	return []*ChannelCustomFieldDefinition{}
}

func (w *WeChatChannelCapabilities) GetAvailableCustomFieldTypes() []shared.FieldType {
	return []shared.FieldType{
		shared.FieldTypeText,
		shared.FieldTypeLabel,
		shared.FieldTypeSelect,
	}
}

func (w *WeChatChannelCapabilities) SupportsGroups() bool {
	return true
}

func (w *WeChatChannelCapabilities) SupportsTracking() bool {
	return true
}

func (w *WeChatChannelCapabilities) SupportsAI() bool {
	return true
}

func (w *WeChatChannelCapabilities) SupportsMedia() bool {
	return true
}

func (w *WeChatChannelCapabilities) SupportedMediaTypes() []string {
	return []string{"image", "video", "voice"}
}

// TwilioSMSChannelCapabilities implementa capabilities para Twilio SMS
type TwilioSMSChannelCapabilities struct{}

func (t *TwilioSMSChannelCapabilities) SupportsLabels() bool {
	return true // Tags + Custom Fields
}

func (t *TwilioSMSChannelCapabilities) SupportsBidirectionalLabelSync() bool {
	return false
}

func (t *TwilioSMSChannelCapabilities) SupportsCustomFields() bool {
	return true
}

func (t *TwilioSMSChannelCapabilities) GetSystemFields() []*ChannelCustomFieldDefinition {
	return []*ChannelCustomFieldDefinition{}
}

func (t *TwilioSMSChannelCapabilities) GetAvailableCustomFieldTypes() []shared.FieldType {
	return []shared.FieldType{
		shared.FieldTypeText,
		shared.FieldTypeNumber,
		shared.FieldTypeLabel,
		shared.FieldTypeSelect,
		shared.FieldTypeMultiSelect,
	}
}

func (t *TwilioSMSChannelCapabilities) SupportsGroups() bool {
	return false
}

func (t *TwilioSMSChannelCapabilities) SupportsTracking() bool {
	return true
}

func (t *TwilioSMSChannelCapabilities) SupportsAI() bool {
	return true
}

func (t *TwilioSMSChannelCapabilities) SupportsMedia() bool {
	return true
}

func (t *TwilioSMSChannelCapabilities) SupportedMediaTypes() []string {
	return []string{"image", "video"} // MMS
}

// WebFormChannelCapabilities implementa capabilities para Web Form/Webhook
type WebFormChannelCapabilities struct{}

func (w *WebFormChannelCapabilities) SupportsLabels() bool {
	return true // Custom implementation
}

func (w *WebFormChannelCapabilities) SupportsBidirectionalLabelSync() bool {
	return false // One-way only
}

func (w *WebFormChannelCapabilities) SupportsCustomFields() bool {
	return true
}

func (w *WebFormChannelCapabilities) GetSystemFields() []*ChannelCustomFieldDefinition {
	return []*ChannelCustomFieldDefinition{
		{
			Key:         "source_url",
			Type:        shared.FieldTypeURL,
			Required:    false,
			Fixed:       true,
			System:      true,
			Description: "URL where the form was submitted",
		},
		{
			Key:         "form_id",
			Type:        shared.FieldTypeText,
			Required:    false,
			Fixed:       true,
			System:      true,
			Description: "Form identifier",
		},
	}
}

func (w *WebFormChannelCapabilities) GetAvailableCustomFieldTypes() []shared.FieldType {
	return []shared.FieldType{
		shared.FieldTypeText,
		shared.FieldTypeNumber,
		shared.FieldTypeBoolean,
		shared.FieldTypeDate,
		shared.FieldTypeURL,
		shared.FieldTypeEmail,
		shared.FieldTypePhone,
		shared.FieldTypeLabel,
		shared.FieldTypeSelect,
		shared.FieldTypeMultiSelect,
		shared.FieldTypeJSON,
	}
}

func (w *WebFormChannelCapabilities) SupportsGroups() bool {
	return false
}

func (w *WebFormChannelCapabilities) SupportsTracking() bool {
	return true
}

func (w *WebFormChannelCapabilities) SupportsAI() bool {
	return true
}

func (w *WebFormChannelCapabilities) SupportsMedia() bool {
	return true
}

func (w *WebFormChannelCapabilities) SupportedMediaTypes() []string {
	return []string{"image", "video", "document", "file"}
}

// WhatsAppCloudChannelCapabilities implementa capabilities para WhatsApp Cloud API
type WhatsAppCloudChannelCapabilities struct{}

func (w *WhatsAppCloudChannelCapabilities) SupportsLabels() bool {
	return false // WhatsApp Cloud API não suporta labels nativamente
}

func (w *WhatsAppCloudChannelCapabilities) SupportsBidirectionalLabelSync() bool {
	return false
}

func (w *WhatsAppCloudChannelCapabilities) SupportsCustomFields() bool {
	return true
}

func (w *WhatsAppCloudChannelCapabilities) GetSystemFields() []*ChannelCustomFieldDefinition {
	return []*ChannelCustomFieldDefinition{}
}

func (w *WhatsAppCloudChannelCapabilities) GetAvailableCustomFieldTypes() []shared.FieldType {
	return []shared.FieldType{
		shared.FieldTypeText,
		shared.FieldTypeNumber,
		shared.FieldTypeSelect,
		shared.FieldTypeLabel, // Custom implementation
	}
}

func (w *WhatsAppCloudChannelCapabilities) SupportsGroups() bool {
	return false // Limited support
}

func (w *WhatsAppCloudChannelCapabilities) SupportsTracking() bool {
	return true
}

func (w *WhatsAppCloudChannelCapabilities) SupportsAI() bool {
	return true
}

func (w *WhatsAppCloudChannelCapabilities) SupportsMedia() bool {
	return true
}

func (w *WhatsAppCloudChannelCapabilities) SupportedMediaTypes() []string {
	return []string{"image", "video", "audio", "document"}
}

// GetCapabilitiesForChannelType retorna as capabilities para um tipo de canal
func GetCapabilitiesForChannelType(channelType ChannelType) ChannelCapabilities {
	switch channelType {
	case TypeWAHA, TypeWhatsAppBusiness:
		return &WAHAChannelCapabilities{}
	case TypeMessenger:
		return &MessengerChannelCapabilities{}
	case TypeInstagram:
		return &InstagramChannelCapabilities{}
	case TypeWhatsApp: // WhatsApp Cloud API
		return &WhatsAppCloudChannelCapabilities{}
	default:
		// Return minimal capabilities for unknown types
		return &WebFormChannelCapabilities{}
	}
}

// Adicionar método ao Channel para obter capabilities
func (c *Channel) GetCapabilities() ChannelCapabilities {
	return GetCapabilitiesForChannelType(c.Type)
}
