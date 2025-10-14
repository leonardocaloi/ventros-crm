package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/storage"
	"go.uber.org/zap"
)

// MediaHandler gerencia upload e download de arquivos de mídia
type MediaHandler struct {
	storage storage.Storage
	logger  *zap.Logger
}

// NewMediaHandler cria um novo handler de mídia
func NewMediaHandler(storage storage.Storage, logger *zap.Logger) *MediaHandler {
	return &MediaHandler{
		storage: storage,
		logger:  logger,
	}
}

// UploadMedia godoc
//
//	@Summary		Upload media file
//	@Description	Faz upload de um arquivo de mídia (imagem, vídeo, áudio, documento) para uso em mensagens
//	@Tags			media
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file		formData	file	true	"Media file to upload (max 64MB)"
//	@Param			project_id	query		string	true	"Project ID"
//	@Success		200			{object}	UploadResponse
//	@Failure		400			{object}	map[string]interface{}	"Invalid request"
//	@Failure		413			{object}	map[string]interface{}	"File too large"
//	@Failure		500			{object}	map[string]interface{}	"Upload failed"
//	@Router			/api/v1/media/upload [post]
//	@Security		BearerAuth
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	// 1. Obter project_id do contexto (set pelo middleware de auth)
	projectID := c.GetString("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "project_id required",
		})
		return
	}

	// 2. Obter arquivo do form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Warn("No file in upload request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})
		return
	}
	defer file.Close()

	// 3. Validar tamanho (64MB max)
	maxSize := int64(64 * 1024 * 1024)
	if header.Size > maxSize {
		h.logger.Warn("File too large",
			zap.Int64("size", header.Size),
			zap.Int64("max_size", maxSize))
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":    "File too large",
			"max_size": "64MB",
			"size":     header.Size,
		})
		return
	}

	// 4. Detectar content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Fallback: detectar por extensão
		ext := filepath.Ext(header.Filename)
		contentType = getContentTypeByExtension(ext)
	}

	// 5. Validar tipo de arquivo
	if !isAllowedContentType(contentType) {
		h.logger.Warn("Invalid content type",
			zap.String("content_type", contentType),
			zap.String("filename", header.Filename))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        "Invalid file type",
			"content_type": contentType,
			"allowed_types": []string{
				"image/*", "video/*", "audio/*",
				"application/pdf", "text/plain",
				"application/vnd.ms-*", "application/vnd.openxmlformats-*",
			},
		})
		return
	}

	// 6. Gerar path único
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Organizar por tipo de mídia
	mediaType := getMediaType(contentType)
	path := fmt.Sprintf("media/%s/%s/%s", projectID, mediaType, filename)

	// 7. Upload para storage
	url, err := h.storage.Upload(c.Request.Context(), file, path, storage.UploadOptions{
		ContentType: contentType,
		Public:      true,
		MaxSize:     maxSize,
		Metadata: map[string]string{
			"original_filename": header.Filename,
			"project_id":        projectID,
			"uploaded_by":       c.GetString("user_id"),
		},
	})

	if err != nil {
		h.logger.Error("Failed to upload file",
			zap.Error(err),
			zap.String("filename", header.Filename),
			zap.String("project_id", projectID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Upload failed",
		})
		return
	}

	// 8. Resposta de sucesso
	h.logger.Info("File uploaded successfully",
		zap.String("url", url),
		zap.String("path", path),
		zap.Int64("size", header.Size),
		zap.String("content_type", contentType),
		zap.String("project_id", projectID))

	c.JSON(http.StatusOK, UploadResponse{
		URL:         url,
		Filename:    header.Filename,
		Size:        header.Size,
		Path:        path,
		ContentType: contentType,
	})
}

// UploadResponse é a resposta do endpoint de upload
type UploadResponse struct {
	URL         string `json:"url" example:"https://storage.googleapis.com/bucket/media/images/uuid.jpg"`
	Filename    string `json:"filename" example:"photo.jpg"`
	Size        int64  `json:"size" example:"1048576"`
	Path        string `json:"path" example:"media/project-id/images/uuid.jpg"`
	ContentType string `json:"content_type" example:"image/jpeg"`
}

// getContentTypeByExtension retorna o content-type baseado na extensão
func getContentTypeByExtension(ext string) string {
	types := map[string]string{
		// Images
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",

		// Videos
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mov":  "video/quicktime",
		".avi":  "video/x-msvideo",

		// Audio
		".mp3":  "audio/mpeg",
		".m4a":  "audio/mp4",
		".ogg":  "audio/ogg",
		".wav":  "audio/wav",
		".opus": "audio/opus",

		// Documents
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}

	if ct, ok := types[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// isAllowedContentType verifica se o content-type é permitido
func isAllowedContentType(contentType string) bool {
	allowed := []string{
		// Images
		"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml",

		// Videos
		"video/mp4", "video/webm", "video/quicktime", "video/x-msvideo", "video/3gpp",

		// Audio
		"audio/mpeg", "audio/mp4", "audio/ogg", "audio/wav", "audio/opus",
		"audio/ogg; codecs=opus",

		// Documents
		"application/pdf",
		"text/plain",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}

	for _, allowed := range allowed {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// getMediaType retorna o tipo de mídia baseado no content-type
func getMediaType(contentType string) string {
	switch {
	case len(contentType) >= 6 && contentType[:6] == "image/":
		return "images"
	case len(contentType) >= 6 && contentType[:6] == "video/":
		return "videos"
	case len(contentType) >= 6 && contentType[:6] == "audio/":
		return "audios"
	default:
		return "documents"
	}
}
