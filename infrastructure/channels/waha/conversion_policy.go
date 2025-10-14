package waha

// ConversionPolicy define a política de conversão de mídia para WhatsApp
type ConversionPolicy interface {
	ShouldConvertVideo(mimetype string) bool
	ShouldConvertAudio(mimetype string) bool
}

// DefaultConversionPolicy é a política padrão: sempre converte
type DefaultConversionPolicy struct{}

func NewDefaultConversionPolicy() *DefaultConversionPolicy {
	return &DefaultConversionPolicy{}
}

// ShouldConvertVideo decide se deve converter vídeo
func (p *DefaultConversionPolicy) ShouldConvertVideo(mimetype string) bool {
	// Sempre converte vídeos para garantir compatibilidade
	return true
}

// ShouldConvertAudio decide se deve converter áudio
func (p *DefaultConversionPolicy) ShouldConvertAudio(mimetype string) bool {
	// Converte se não for opus (formato nativo do WhatsApp)
	return mimetype != "audio/ogg; codecs=opus"
}

// NoConversionPolicy nunca converte (para testes ou casos específicos)
type NoConversionPolicy struct{}

func NewNoConversionPolicy() *NoConversionPolicy {
	return &NoConversionPolicy{}
}

func (p *NoConversionPolicy) ShouldConvertVideo(mimetype string) bool {
	return false
}

func (p *NoConversionPolicy) ShouldConvertAudio(mimetype string) bool {
	return false
}

// SmartConversionPolicy converte apenas formatos não suportados
type SmartConversionPolicy struct{}

func NewSmartConversionPolicy() *SmartConversionPolicy {
	return &SmartConversionPolicy{}
}

// ShouldConvertVideo converte apenas formatos não compatíveis com WhatsApp
func (p *SmartConversionPolicy) ShouldConvertVideo(mimetype string) bool {
	supportedFormats := map[string]bool{
		"video/mp4":  true,
		"video/3gpp": true,
	}
	return !supportedFormats[mimetype]
}

// ShouldConvertAudio converte apenas se não for opus
func (p *SmartConversionPolicy) ShouldConvertAudio(mimetype string) bool {
	return mimetype != "audio/ogg; codecs=opus"
}

// ChannelTypeConversionPolicy define política baseada no tipo de canal
type ChannelTypeConversionPolicy struct {
	channelType string
}

// NewChannelTypeConversionPolicy cria política baseada no tipo de canal
func NewChannelTypeConversionPolicy(channelType string) *ChannelTypeConversionPolicy {
	return &ChannelTypeConversionPolicy{
		channelType: channelType,
	}
}

// ShouldConvertVideo decide conversão baseado no tipo de canal
func (p *ChannelTypeConversionPolicy) ShouldConvertVideo(mimetype string) bool {
	switch p.channelType {
	case "waha":
		// WAHA tem ffmpeg embutido, sempre converte para garantir compatibilidade
		return true
	case "whatsapp_business":
		// WhatsApp Business API só aceita MP4
		return mimetype != "video/mp4"
	case "telegram":
		// Telegram aceita vários formatos, não precisa converter
		return false
	default:
		// Padrão: sempre converte para garantir compatibilidade
		return true
	}
}

// ShouldConvertAudio decide conversão de áudio baseado no tipo de canal
func (p *ChannelTypeConversionPolicy) ShouldConvertAudio(mimetype string) bool {
	switch p.channelType {
	case "waha":
		// WAHA converte para opus se necessário
		return mimetype != "audio/ogg; codecs=opus"
	case "whatsapp_business":
		// WhatsApp Business aceita vários formatos de áudio
		supportedFormats := map[string]bool{
			"audio/ogg; codecs=opus": true,
			"audio/mpeg":             true, // MP3
			"audio/mp4":              true, // AAC/M4A
		}
		return !supportedFormats[mimetype]
	case "telegram":
		// Telegram aceita praticamente tudo
		return false
	default:
		// Padrão: converte se não for opus
		return mimetype != "audio/ogg; codecs=opus"
	}
}
