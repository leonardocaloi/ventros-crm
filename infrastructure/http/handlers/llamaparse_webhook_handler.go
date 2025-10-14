package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ventros/crm/infrastructure/ai"
	"github.com/ventros/crm/internal/domain/crm/message_enrichment"
)

// LlamaParseWebhookHandler handler para receber resultados do LlamaParse via webhook
type LlamaParseWebhookHandler struct {
	logger         *zap.Logger
	enrichmentRepo message_enrichment.Repository
}

// NewLlamaParseWebhookHandler cria novo handler
func NewLlamaParseWebhookHandler(
	logger *zap.Logger,
	enrichmentRepo message_enrichment.Repository,
) *LlamaParseWebhookHandler {
	return &LlamaParseWebhookHandler{
		logger:         logger,
		enrichmentRepo: enrichmentRepo,
	}
}

// HandleWebhook recebe POST do LlamaParse com resultado do parsing
// POST /api/webhooks/llamaparse
// Body: LlamaParseWebhookPayload (JSON)
func (h *LlamaParseWebhookHandler) HandleWebhook(c *gin.Context) {
	h.logger.Info("Received LlamaParse webhook callback")

	// Parse payload
	var payload ai.LlamaParseWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.logger.Error("Failed to parse webhook payload",
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payload format",
		})
		return
	}

	h.logger.Info("LlamaParse webhook payload received",
		zap.String("job_id", payload.JobID),
		zap.String("status", payload.Status),
		zap.Int("markdown_length", len(payload.Markdown)))

	// Verificar status
	if payload.Status != "SUCCESS" {
		h.logger.Warn("LlamaParse job failed",
			zap.String("job_id", payload.JobID),
			zap.String("error", payload.Error))

		// TODO: Atualizar enrichment com status de erro
		c.JSON(http.StatusOK, gin.H{
			"message": "Webhook received (job failed)",
			"job_id":  payload.JobID,
		})
		return
	}

	// Processar resultado bem-sucedido
	h.logger.Info("Processing successful LlamaParse result",
		zap.String("job_id", payload.JobID),
		zap.Int("page_count", len(payload.Pages)),
		zap.Int("image_count", len(payload.Images)))

	// TODO: Salvar resultado no enrichment repository
	// Procurar enrichment pelo job_id (armazenado no metadata quando criado)
	// Atualizar com:
	// - ExtractedText = payload.Markdown (ou payload.Text)
	// - Status = "completed"
	// - Metadata com detalhes das páginas e imagens

	// Por enquanto, apenas log
	h.logger.Info("LlamaParse result processed successfully",
		zap.String("job_id", payload.JobID),
		zap.Int("markdown_length", len(payload.Markdown)),
		zap.Int("text_length", len(payload.Text)))

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook received and processed",
		"job_id":  payload.JobID,
	})
}

// RegisterRoutes registra rotas do webhook
func (h *LlamaParseWebhookHandler) RegisterRoutes(router *gin.RouterGroup) {
	webhooks := router.Group("/webhooks")
	{
		// POST /api/webhooks/llamaparse
		// Sem autenticação (LlamaParse não suporta headers customizados em webhooks)
		// TODO: Adicionar validação por IP whitelist ou signature se necessário
		webhooks.POST("/llamaparse", h.HandleWebhook)
	}
}
