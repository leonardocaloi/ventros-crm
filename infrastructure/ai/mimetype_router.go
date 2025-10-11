package ai

import (
	"strings"

	"github.com/caloi/ventros-crm/internal/domain/crm/channel"
)

// MimetypeRouter roteia mimetypes para tipos de conteúdo de IA
type MimetypeRouter struct{}

// NewMimetypeRouter cria um novo router de mimetype
func NewMimetypeRouter() *MimetypeRouter {
	return &MimetypeRouter{}
}

// RouteToContentType converte mimetype para AIContentType
func (r *MimetypeRouter) RouteToContentType(mimetype string, isPTT bool) channel.AIContentType {
	mimetype = strings.ToLower(strings.TrimSpace(mimetype))

	// Áudio de voz (PTT - Push-to-Talk) tem prioridade
	if isPTT {
		return channel.AIContentTypeVoice
	}

	// Áudio
	if strings.HasPrefix(mimetype, "audio/") {
		return channel.AIContentTypeAudio
	}

	// Imagem
	if strings.HasPrefix(mimetype, "image/") {
		return channel.AIContentTypeImage
	}

	// Vídeo
	if strings.HasPrefix(mimetype, "video/") {
		return channel.AIContentTypeVideo
	}

	// Documentos
	switch mimetype {
	case "application/pdf":
		return channel.AIContentTypeDocument
	case "application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return channel.AIContentTypeDocument
	}

	// Texto por padrão
	return channel.AIContentTypeText
}

// IsProcessableType verifica se o mimetype pode ser processado
func (r *MimetypeRouter) IsProcessableType(mimetype string) bool {
	mimetype = strings.ToLower(strings.TrimSpace(mimetype))

	processable := []string{
		// Áudio
		"audio/",
		// Imagem
		"image/",
		// Vídeo
		"video/",
		// Documentos
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats",
		"application/vnd.ms-",
	}

	for _, prefix := range processable {
		if strings.HasPrefix(mimetype, prefix) {
			return true
		}
	}

	return false
}

// GetEstimatedDuration estima duração de processamento em segundos
func (r *MimetypeRouter) GetEstimatedDuration(contentType channel.AIContentType, sizeBytes int64) int {
	switch contentType {
	case channel.AIContentTypeText:
		return 2 // 2 segundos
	case channel.AIContentTypeVoice:
		return 5 // 5 segundos (PTT geralmente são curtos)
	case channel.AIContentTypeAudio:
		// Áudio: ~1 segundo por minuto de áudio
		// Estimativa: 1MB ~= 1 minuto de áudio
		minutes := int(sizeBytes / (1024 * 1024))
		if minutes < 1 {
			minutes = 1
		}
		return minutes * 60
	case channel.AIContentTypeImage:
		return 10 // 10 segundos
	case channel.AIContentTypeVideo:
		// Vídeo: mais pesado, ~10 segundos por MB
		mb := int(sizeBytes / (1024 * 1024))
		if mb < 1 {
			mb = 1
		}
		return mb * 10
	case channel.AIContentTypeDocument:
		// PDF: ~5 segundos por MB
		mb := int(sizeBytes / (1024 * 1024))
		if mb < 1 {
			mb = 1
		}
		return mb * 5
	default:
		return 30 // Padrão: 30 segundos
	}
}
