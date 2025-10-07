package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TestHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewTestHandler(db *gorm.DB, logger *zap.Logger) *TestHandler {
	return &TestHandler{
		db:     db,
		logger: logger,
	}
}

// cleanupTestData remove todas as entidades de teste antes de criar novas
func (h *TestHandler) cleanupTestData(tx *gorm.DB) error {
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	projectID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	// Set user context for RLS
	if err := tx.Exec("SELECT set_config('app.current_user_id', ?, false)", userID.String()).Error; err != nil {
		h.logger.Warn("Failed to set user context", zap.Error(err))
	}

	// Delete in reverse order of dependencies
	h.logger.Info("Cleaning up test data...",
		zap.String("user_id", userID.String()),
		zap.String("project_id", projectID.String()))

	// 1. Delete webhook subscriptions (sem RLS)
	result := tx.Unscoped().Where("name = ?", "Webhook Teste N8N").Delete(&entities.WebhookSubscriptionEntity{})
	if result.Error != nil {
		h.logger.Warn("Failed to delete webhook", zap.Error(result.Error))
	} else {
		h.logger.Info("Deleted webhooks", zap.Int64("count", result.RowsAffected))
	}

	// 2. Delete pipeline statuses (by project pipelines) - hard delete
	if err := tx.Exec("DELETE FROM pipeline_statuses WHERE pipeline_id IN (SELECT id FROM pipelines WHERE project_id = ?)", projectID).Error; err != nil {
		h.logger.Warn("Failed to delete pipeline statuses", zap.Error(err))
	}

	// 3. Delete pipelines - hard delete
	if err := tx.Unscoped().Where("project_id = ?", projectID).Delete(&entities.PipelineEntity{}).Error; err != nil {
		h.logger.Warn("Failed to delete pipelines", zap.Error(err))
	}

	// 4. Delete sessions and related data - hard delete
	// Sessions relate to contacts, not directly to projects
	if err := tx.Exec("DELETE FROM messages WHERE session_id IN (SELECT id FROM sessions WHERE contact_id IN (SELECT id FROM contacts WHERE project_id = ?))", projectID).Error; err != nil {
		h.logger.Warn("Failed to delete messages", zap.Error(err))
	}

	if err := tx.Exec("DELETE FROM sessions WHERE contact_id IN (SELECT id FROM contacts WHERE project_id = ?)", projectID).Error; err != nil {
		h.logger.Warn("Failed to delete sessions", zap.Error(err))
	}

	// 5. Delete contacts - hard delete
	if err := tx.Unscoped().Where("project_id = ?", projectID).Delete(&entities.ContactEntity{}).Error; err != nil {
		h.logger.Warn("Failed to delete contacts", zap.Error(err))
	}

	// 6. Delete project - hard delete
	if err := tx.Unscoped().Where("id = ?", projectID).Delete(&entities.ProjectEntity{}).Error; err != nil {
		h.logger.Warn("Failed to delete project", zap.Error(err))
	}

	// 7. Delete billing account - hard delete (must be before user delete)
	if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.BillingAccountEntity{}).Error; err != nil {
		h.logger.Warn("Failed to delete billing account", zap.Error(err))
	}

	// 8. Delete API keys - hard delete (must be before user delete due to FK)
	if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserAPIKeyEntity{}).Error; err != nil {
		h.logger.Warn("Failed to delete API keys", zap.Error(err))
	}

	// 9. Delete user - hard delete (CRITICAL: users have soft delete)
	result = tx.Unscoped().Where("id = ?", userID).Delete(&entities.UserEntity{})
	if result.Error != nil {
		h.logger.Error("Failed to delete user", zap.Error(result.Error))
		return result.Error
	}
	h.logger.Info("Deleted user", zap.Int64("count", result.RowsAffected))

	h.logger.Info("Test data cleanup completed successfully")
	return nil
}

// SetupTestEnvironment configura ambiente de teste completo
// @Summary Setup test environment
// @Description Limpa e cria project, pipeline, channel types e webhook para testes
// @Tags test
// @Produce json
// @Param webhook_url query string false "URL do webhook externo (opcional)"
// @Param api_base_url query string false "Base URL da API (opcional, default: http://localhost:8080)"
// @Success 200 {object} map[string]interface{}
// @Router /test/setup [post]
func (h *TestHandler) SetupTestEnvironment(c *gin.Context) {
	// Obter webhook URL do query param (opcional)
	webhookURL := c.Query("webhook_url")
	if webhookURL == "" {
		webhookURL = "https://dev.webhook.n8n.ventros.cloud/webhook/ventros-crm-test"
	}

	// Obter base URL da API do query param (opcional)
	apiBaseURL := c.Query("api_base_url")
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:8080"
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 0. CLEANUP - Remove dados de teste antigos
	if err := h.cleanupTestData(tx); err != nil {
		tx.Rollback()
		h.logger.Error("Failed to cleanup test data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup test data"})
		return
	}

	// 1. Criar User novo
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	user := entities.UserEntity{
		ID:        userID,
		Name:      "Usuario Teste",
		Email:     "teste@ventros.com",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 2. Criar Billing Account
	billingAccountID := uuid.New()
	billingAccount := entities.BillingAccountEntity{
		ID:            billingAccountID,
		UserID:        userID,
		Name:          "Conta Teste",
		PaymentStatus: "active",
		BillingEmail:  "teste@ventros.com",
		Suspended:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := tx.Create(&billingAccount).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create billing account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create billing account: %v", err)})
		return
	}

	h.logger.Info("Billing account created", zap.String("id", billingAccountID.String()))

	// 3. Criar Project novo
	projectID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	// Usar o ID do billing account criado
	project := entities.ProjectEntity{
		ID:               projectID,
		UserID:           userID,
		BillingAccountID: billingAccount.ID, // Usar o ID do objeto criado
		TenantID:         "test-tenant",
		Name:             "Projeto Teste WAHA",
		Description:      "Projeto para testes de integração WAHA",
		Active:           true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	h.logger.Info("Creating project",
		zap.String("project_id", projectID.String()),
		zap.String("billing_account_id", project.BillingAccountID.String()))

	if err := tx.Create(&project).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create project: %v", err)})
		return
	}

	// 3. Criar Pipeline padrão
	pipelineID := uuid.New()
	pipeline := entities.PipelineEntity{
		ID:          pipelineID,
		ProjectID:   projectID,
		TenantID:    "default",
		Name:        "Pipeline Padrão",
		Description: "Pipeline padrão para novos contatos",
		Color:       "#3B82F6",
		Position:    1,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := tx.Create(&pipeline).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create pipeline", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pipeline"})
		return
	}

	// 4. Criar Status do Pipeline
	statusID := uuid.New()
	status := entities.PipelineStatusEntity{
		ID:          statusID,
		PipelineID:  pipelineID,
		Name:        "Novo Lead",
		Description: "Contato recém chegado",
		Color:       "#10B981",
		StatusType:  "open",
		Position:    1,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := tx.Create(&status).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create pipeline status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pipeline status"})
		return
	}

	// 5. Criar Channel WAHA de teste
	channelID := uuid.New()
	externalID := "test-session-waha"
	now := time.Now()

	// Gerar URL do webhook interno automaticamente usando a base URL configurada
	channelWebhookURL := fmt.Sprintf("%s/api/v1/webhooks/waha?session=%s", apiBaseURL, externalID)

	channel := entities.ChannelEntity{
		ID:                  channelID,
		UserID:              userID,
		ProjectID:           projectID,
		TenantID:            "test-tenant",
		Name:                "Canal WAHA Teste",
		Type:                "waha",
		Status:              "active",
		ExternalID:          externalID,
		WebhookURL:          channelWebhookURL,
		WebhookConfiguredAt: &now,
		WebhookActive:       true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := tx.Create(&channel).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}

	// 6. Criar Webhook Subscription de teste
	webhookID := uuid.New()
	webhook := entities.WebhookSubscriptionEntity{
		ID:             webhookID,
		UserID:         userID,
		ProjectID:      projectID,
		TenantID:       "test-tenant",
		Name:           "Webhook Teste N8N",
		URL:            webhookURL,
		Events:         pq.StringArray{"contact.created", "session.started", "session.ended", "ad_campaign.tracked"},
		Active:         true,
		RetryCount:     3,
		TimeoutSeconds: 30,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := tx.Create(&webhook).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook"})
		return
	}

	// 7. Criar API Key para autenticação nos testes
	apiKey := "test-api-key-" + userID.String()
	apiKeyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(apiKey)))
	apiKeyEntity := entities.UserAPIKeyEntity{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "Test API Key",
		KeyHash:   apiKeyHash,
		Active:    true,
		ExpiresAt: nil, // Sem expiração
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Create(&apiKeyEntity).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	tx.Commit()

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_id":             userID,
			"project_id":          projectID,
			"pipeline_id":         pipelineID,
			"status_id":           statusID,
			"channel_id":          channelID,
			"channel_webhook_url": channelWebhookURL,
			"webhook_id":          webhookID,
			"api_key":             apiKey,
		},
		"message": "Test environment cleaned and setup completed successfully",
	}

	h.logger.Info("Test environment setup completed",
		zap.String("user_id", userID.String()),
		zap.String("project_id", projectID.String()),
		zap.String("pipeline_id", pipelineID.String()),
		zap.String("status_id", statusID.String()),
		zap.String("webhook_id", webhookID.String()),
	)

	c.JSON(http.StatusOK, response)
}

// CleanupTestEnvironment remove todas as entidades de teste
// @Summary Cleanup test environment
// @Description Remove todos os dados de teste criados
// @Tags test
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /test/cleanup [post]
func (h *TestHandler) CleanupTestEnvironment(c *gin.Context) {
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := h.cleanupTestData(tx); err != nil {
		tx.Rollback()
		h.logger.Error("Failed to cleanup test data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup test data"})
		return
	}

	tx.Commit()

	h.logger.Info("Test environment cleanup completed")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Test environment cleaned successfully",
	})
}

// TestWAHAMessage envia mensagem WAHA de teste com tracking
// @Summary Test WAHA message with tracking
// @Description Envia mensagem WAHA de teste com dados de tracking do Facebook/Instagram
// @Tags test
// @Produce json
// @Param type query string false "Message type: fb_ads, text, image (default: fb_ads)"
// @Success 200 {object} map[string]interface{}
// @Router /test/waha-message [post]
func (h *TestHandler) TestWAHAMessage(c *gin.Context) {
	messageType := c.Query("type")
	if messageType == "" {
		messageType = "fb_ads"
	}

	var curlCommand string
	var description string

	switch messageType {
	case "fb_ads":
		curlCommand = h.getFBAdsMessageCurl()
		description = "Mensagem com tracking do Facebook/Instagram Ads"
	case "text":
		curlCommand = h.getTextMessageCurl()
		description = "Mensagem de texto simples"
	case "image":
		curlCommand = h.getImageMessageCurl()
		description = "Mensagem com imagem"
	default:
		curlCommand = h.getFBAdsMessageCurl()
		description = "Mensagem com tracking do Facebook/Instagram Ads (padrão)"
	}

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"message_type":    messageType,
			"description":     description,
			"webhook_url":     "/api/v1/webhooks/waha",
			"expected_events": []string{"contact.created", "session.started", "ad_campaign.tracked"},
		},
		"message":         "Use the curl command below to test WAHA webhook",
		"curl_command":    curlCommand,
		"available_types": []string{"fb_ads", "text", "image"},
	}

	c.JSON(http.StatusOK, response)
}

// SendWAHAMessage envia mensagem WAHA automaticamente para o webhook
// @Summary Send WAHA message automatically
// @Description Envia mensagem WAHA automaticamente para o webhook interno
// @Tags test
// @Produce json
// @Param type query string false "Message type: fb_ads, text, image (default: fb_ads)"
// @Success 200 {object} map[string]interface{}
// @Router /test/send-waha-message [post]
func (h *TestHandler) SendWAHAMessage(c *gin.Context) {
	messageType := c.Query("type")
	if messageType == "" {
		messageType = "fb_ads"
	}

	var payload map[string]interface{}
	var description string

	switch messageType {
	case "fb_ads":
		payload = h.getFBAdsMessagePayload()
		description = "Mensagem com tracking do Facebook/Instagram Ads"
	case "text":
		payload = h.getTextMessagePayload()
		description = "Mensagem de texto simples"
	case "image":
		payload = h.getImageMessagePayload()
		description = "Mensagem com imagem"
	default:
		payload = h.getFBAdsMessagePayload()
		description = "Mensagem com tracking do Facebook/Instagram Ads (padrão)"
	}

	// Converter payload para JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("Failed to marshal payload", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal payload"})
		return
	}

	// Fazer request interno para o webhook WAHA
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/webhooks/waha", bytes.NewBuffer(jsonData))
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to send request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to webhook"})
		return
	}
	defer resp.Body.Close()

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"message_type":    messageType,
			"description":     description,
			"webhook_status":  resp.StatusCode,
			"webhook_url":     "/api/v1/webhooks/waha",
			"expected_events": []string{"contact.created", "session.started", "ad_campaign.tracked"},
		},
		"message": fmt.Sprintf("WAHA message sent successfully! Webhook returned status: %d", resp.StatusCode),
	}

	if resp.StatusCode != 200 {
		response["success"] = false
		response["message"] = fmt.Sprintf("Webhook returned error status: %d", resp.StatusCode)
	}

	c.JSON(http.StatusOK, response)
}

// getFBAdsMessageCurl retorna curl para mensagem com tracking do Facebook/Instagram
func (h *TestHandler) getFBAdsMessageCurl() string {
	return `curl -X POST http://localhost:8080/api/v1/webhooks/waha \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt_01k6jymfk5s69jmpjq30n7pgd2",
    "timestamp": 1759425216101,
    "event": "message",
    "session": "test_session_e2e",
    "payload": {
      "id": "false_554498211518@c.us_2AF14CD157D1CF76DC78",
      "from": "554498211518@c.us",
      "fromMe": false,
      "body": "Olá! Tenho interesse na imersão e queria mais informações, por favor.",
      "_data": {
        "Info": {
          "PushName": "Nardin"
        },
        "Message": {
          "extendedTextMessage": {
            "contextInfo": {
              "conversionSource": "FB_Ads",
              "entryPointConversionSource": "ctwa_ad",
              "entryPointConversionApp": "instagram",
              "ctwaClid": "Afcg5DA4aj8pfp0faj5HIKtKi2vOUGt4agVezOhRfZA0MLUab3uUE98YGiB3KDFGsjR7Ohm5nIHEHugiymilThQM5gEEr_XAu5sHLZX5GGcnc9v9y3ImoTXetXaKP640XR222lch"
            }
          }
        }
      }
    }
  }'`
}

// getTextMessageCurl retorna curl para mensagem de texto simples
func (h *TestHandler) getTextMessageCurl() string {
	return `curl -X POST http://localhost:8080/api/v1/webhooks/waha \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt_01k5vejyxtvtrn6vkwh10317pc",
    "timestamp": 1758636637114,
    "event": "message",
    "session": "test_session_e2e",
    "payload": {
      "id": "3A3CE2C5C341306326CF",
      "from": "554498699850@c.us",
      "fromMe": false,
      "body": "No momento não vou poder",
      "_data": {
        "Info": {
          "PushName": "Ana Alves"
        },
        "Message": {
          "conversation": "No momento não vou poder"
        }
      }
    }
  }'`
}

// getImageMessageCurl retorna curl para mensagem com imagem
func (h *TestHandler) getImageMessageCurl() string {
	return `curl -X POST http://localhost:8080/api/v1/webhooks/waha \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt_01k68v8fnape65p4h8jk1jgh16",
    "timestamp": 1759086132906,
    "event": "message",
    "session": "test_session_e2e",
    "payload": {
      "id": "false_status@broadcast_AC94411C453CBE7D20F5EC3EB01DF066_554499223925@c.us",
      "from": "554499223925@c.us",
      "fromMe": false,
      "hasMedia": true,
      "media": {
        "url": "https://storage.googleapis.com/waha-ventros/ask-dermato-imersao/AC94411C453CBE7D20F5EC3EB01DF066.jpeg",
        "mimetype": "image/jpeg"
      },
      "_data": {
        "Info": {
          "PushName": "Aline Fatobene"
        },
        "Message": {
          "imageMessage": {
            "mimetype": "image/jpeg",
            "height": 1280,
            "width": 720
          }
        }
      }
    }
  }'`
}

// Métodos para retornar payloads como map (para envio automático)
func (h *TestHandler) getFBAdsMessagePayload() map[string]interface{} {
	return map[string]interface{}{
		"id":        "evt_01k6jymfk5s69jmpjq30n7pgd2",
		"timestamp": 1759425216101,
		"event":     "message",
		"session":   "test_session_e2e",
		"payload": map[string]interface{}{
			"id":     "false_554498211518@c.us_2AF14CD157D1CF76DC78",
			"from":   "554498211518@c.us",
			"fromMe": false,
			"body":   "Olá! Tenho interesse na imersão e queria mais informações, por favor.",
			"_data": map[string]interface{}{
				"Info": map[string]interface{}{
					"PushName": "Nardin",
				},
				"Message": map[string]interface{}{
					"extendedTextMessage": map[string]interface{}{
						"contextInfo": map[string]interface{}{
							"conversionSource":           "FB_Ads",
							"entryPointConversionSource": "ctwa_ad",
							"entryPointConversionApp":    "instagram",
							"ctwaClid":                   "Afcg5DA4aj8pfp0faj5HIKtKi2vOUGt4agVezOhRfZA0MLUab3uUE98YGiB3KDFGsjR7Ohm5nIHEHugiymilThQM5gEEr_XAu5sHLZX5GGcnc9v9y3ImoTXetXaKP640XR222lch",
						},
					},
				},
			},
		},
	}
}

func (h *TestHandler) getTextMessagePayload() map[string]interface{} {
	return map[string]interface{}{
		"id":        "evt_01k5vejyxtvtrn6vkwh10317pc",
		"timestamp": 1758636637114,
		"event":     "message",
		"session":   "test_session_e2e",
		"payload": map[string]interface{}{
			"id":     "3A3CE2C5C341306326CF",
			"from":   "554498699850@c.us",
			"fromMe": false,
			"body":   "No momento não vou poder",
			"_data": map[string]interface{}{
				"Info": map[string]interface{}{
					"PushName": "Ana Alves",
				},
				"Message": map[string]interface{}{
					"conversation": "No momento não vou poder",
				},
			},
		},
	}
}

func (h *TestHandler) getImageMessagePayload() map[string]interface{} {
	return map[string]interface{}{
		"id":        "evt_01k68v8fnape65p4h8jk1jgh16",
		"timestamp": 1759086132906,
		"event":     "message",
		"session":   "test_session_e2e",
		"payload": map[string]interface{}{
			"id":       "false_status@broadcast_AC94411C453CBE7D20F5EC3EB01DF066_554499223925@c.us",
			"from":     "554499223925@c.us",
			"fromMe":   false,
			"hasMedia": true,
			"media": map[string]interface{}{
				"url":      "https://storage.googleapis.com/waha-ventros/ask-dermato-imersao/AC94411C453CBE7D20F5EC3EB01DF066.jpeg",
				"mimetype": "image/jpeg",
			},
			"_data": map[string]interface{}{
				"Info": map[string]interface{}{
					"PushName": "Aline Fatobene",
				},
				"Message": map[string]interface{}{
					"imageMessage": map[string]interface{}{
						"mimetype": "image/jpeg",
						"height":   1280,
						"width":    720,
					},
				},
			},
		},
	}
}

// TestWAHAConnection testa a conexão com a WAHA
// @Summary Test WAHA connection
// @Description Testa a conexão com a API WAHA usando token e base URL
// @Tags test
// @Accept json
// @Produce json
// @Param request body TestWAHARequest true "WAHA connection data"
// @Success 200 {object} map[string]interface{} "Connection test result"
// @Router /api/v1/test/waha-connection [post]
func (h *TestHandler) TestWAHAConnection(c *gin.Context) {
	var req TestWAHARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	h.logger.Info("Testing WAHA connection",
		zap.String("base_url", req.BaseURL),
		zap.String("token", "***masked***"))

	// Teste básico de conectividade
	client := &http.Client{Timeout: 10 * time.Second}

	// Testa endpoint de health/status da WAHA
	healthURL := fmt.Sprintf("%s/api/sessions", req.BaseURL)
	httpReq, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create request",
			"details": err.Error(),
		})
		return
	}

	// Adiciona token se fornecido
	if req.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		h.logger.Error("WAHA connection failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":    "Failed to connect to WAHA",
			"details":  err.Error(),
			"base_url": req.BaseURL,
		})
		return
	}
	defer resp.Body.Close()

	h.logger.Info("WAHA connection successful",
		zap.Int("status_code", resp.StatusCode),
		zap.String("status", resp.Status))

	c.JSON(http.StatusOK, gin.H{
		"message":     "WAHA connection successful",
		"base_url":    req.BaseURL,
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

// TestWAHAQRCode simula recebimento de QR code da WAHA
// @Summary Test WAHA QR code
// @Description Simula recebimento de um QR code da WAHA para teste
// @Tags test
// @Accept json
// @Produce json
// @Param request body TestQRCodeRequest true "QR code test data"
// @Success 200 {object} map[string]interface{} "QR code test result"
// @Router /api/v1/test/waha-qr [post]
func (h *TestHandler) TestWAHAQRCode(c *gin.Context) {
	var req TestQRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Simula um QR code da WAHA (formato típico)
	mockQRCode := fmt.Sprintf("2@%s,null,null,%s",
		generateMockQRData(),
		time.Now().Format("20060102150405"))

	h.logger.Info("Simulating WAHA QR code",
		zap.String("session_id", req.SessionID),
		zap.String("channel_name", req.ChannelName))

	// Log do QR code no console
	separator := strings.Repeat("=", 80)
	fmt.Printf("\n%s\n", separator)
	fmt.Printf("📱 [TESTE WAHA QR CODE] Canal: %s | Session: %s\n", req.ChannelName, req.SessionID)
	fmt.Printf("🕒 Gerado em: %s\n", time.Now().Format("15:04:05"))
	fmt.Printf("⏰ Expira em: %s\n", time.Now().Add(45*time.Second).Format("15:04:05"))
	fmt.Printf("📋 QR Code (simulado):\n%s\n", mockQRCode)
	fmt.Printf("🔗 Para testar: Use um leitor de QR code ou WhatsApp Web\n")
	fmt.Printf("%s\n\n", separator)

	c.JSON(http.StatusOK, gin.H{
		"message":      "QR code generated successfully",
		"session_id":   req.SessionID,
		"channel_name": req.ChannelName,
		"qr_code":      mockQRCode,
		"generated_at": time.Now().Unix(),
		"expires_at":   time.Now().Add(45 * time.Second).Unix(),
		"status":       "SCAN_QR_CODE",
		"instructions": "Use WhatsApp mobile app to scan this QR code",
	})
}

// Estruturas para requests de teste
type TestWAHARequest struct {
	BaseURL string `json:"base_url" binding:"required" example:"http://localhost:3000"`
	Token   string `json:"token" example:"your-waha-token"`
}

type TestQRCodeRequest struct {
	SessionID   string `json:"session_id" binding:"required" example:"default"`
	ChannelName string `json:"channel_name" binding:"required" example:"WhatsApp Teste"`
}

// generateMockQRData gera dados simulados para QR code
func generateMockQRData() string {
	return fmt.Sprintf("1@%s@%s@%s",
		generateRandomString(25),
		generateRandomString(88),
		generateRandomString(43))
}

// generateRandomString gera string aleatória para simulação
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/="
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
