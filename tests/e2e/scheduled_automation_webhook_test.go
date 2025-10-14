package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	_ "github.com/lib/pq"
)

// ScheduledAutomationWebhookTestSuite testa automações agendadas COM verificação de webhooks
type ScheduledAutomationWebhookTestSuite struct {
	suite.Suite
	baseURL         string
	client          *http.Client
	userID          string
	projectID       string
	apiKey          string
	channelID       string
	pipelineID      string
	contactID       string
	ruleID          string
	subscriptionID  string
	db              *sql.DB
	webhookServer   *httptest.Server
	webhookReceived []map[string]interface{}
	webhookMutex    sync.Mutex
}

// SetupSuite executa uma vez antes de todos os testes
func (s *ScheduledAutomationWebhookTestSuite) SetupSuite() {
	// Configura URL base
	s.baseURL = os.Getenv("API_BASE_URL")
	if s.baseURL == "" {
		s.baseURL = "http://localhost:8080"
	}

	s.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Conecta ao banco de dados
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://ventros:ventros123@localhost:5432/ventros_crm?sslmode=disable"
	}

	var err error
	s.db, err = sql.Open("postgres", dbURL)
	assert.NoError(s.T(), err, "Failed to connect to database")

	err = s.db.Ping()
	assert.NoError(s.T(), err, "Failed to ping database")

	// Inicia webhook server de teste (recebe webhooks enviados pela API)
	s.startWebhookServer()

	// Aguarda API estar pronta
	s.waitForAPI()

	fmt.Println("\n🔔 Setting up Scheduled Automation + Webhook E2E Test")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 1. Cria usuário
	s.createUser()

	// 2. Cria canal WAHA
	s.createWAHAChannel()

	// 3. Busca ou cria pipeline
	s.getOrCreatePipeline()

	// 4. Cria contato de teste
	s.createContact()

	// 5. Cria webhook subscription para automation.executed
	s.createWebhookSubscription()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Setup completo!")
	fmt.Printf("📍 Webhook Server: %s\n", s.webhookServer.URL)
	fmt.Println("")
}

// TearDownSuite executa após todos os testes (CLEANUP)
func (s *ScheduledAutomationWebhookTestSuite) TearDownSuite() {
	fmt.Println("\n🧹 Cleaning up test data...")

	// Deleta webhook subscription
	if s.subscriptionID != "" {
		endpoint := fmt.Sprintf("/api/v1/webhooks/subscriptions/%s", s.subscriptionID)
		s.makeRequest("DELETE", endpoint, nil, s.apiKey)
		fmt.Printf("  ✓ Deleted webhook subscription: %s\n", s.subscriptionID)
	}

	// Deleta automation rule se existir
	if s.ruleID != "" {
		s.db.Exec("DELETE FROM automation_rules WHERE id = $1", s.ruleID)
		fmt.Printf("  ✓ Deleted automation rule: %s\n", s.ruleID)
	}

	// Deleta contato se existir
	if s.contactID != "" {
		s.db.Exec("DELETE FROM contacts WHERE id = $1", s.contactID)
		fmt.Printf("  ✓ Deleted contact: %s\n", s.contactID)
	}

	// Deleta canal se existir
	if s.channelID != "" && s.apiKey != "" {
		endpoint := fmt.Sprintf("/api/v1/crm/channels/%s", s.channelID)
		s.makeRequest("DELETE", endpoint, nil, s.apiKey)
		fmt.Printf("  ✓ Deleted channel: %s\n", s.channelID)
	}

	if s.db != nil {
		s.db.Close()
	}

	if s.webhookServer != nil {
		s.webhookServer.Close()
		fmt.Println("  ✓ Stopped webhook server")
	}

	fmt.Println("✅ Cleanup completed")
}

// startWebhookServer inicia servidor HTTP para receber webhooks
func (s *ScheduledAutomationWebhookTestSuite) startWebhookServer() {
	s.webhookReceived = make([]map[string]interface{}, 0)

	s.webhookServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err == nil {
			s.webhookMutex.Lock()
			s.webhookReceived = append(s.webhookReceived, payload)
			s.webhookMutex.Unlock()

			fmt.Printf("   📨 Webhook received: event=%s\n", payload["event"])
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))

	fmt.Printf("🔔 Webhook server started: %s\n", s.webhookServer.URL)
}

// waitForAPI aguarda a API estar disponível
func (s *ScheduledAutomationWebhookTestSuite) waitForAPI() {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		resp, err := s.client.Get(s.baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Println("✅ API is ready")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	s.T().Fatal("API não ficou disponível após 30 segundos")
}

// createUser cria um usuário de teste
func (s *ScheduledAutomationWebhookTestSuite) createUser() {
	timestamp := time.Now().Unix()
	payload := map[string]string{
		"name":     fmt.Sprintf("Test User Webhook %d", timestamp),
		"email":    fmt.Sprintf("test-webhook-%d@example.com", timestamp),
		"password": "Test@123456",
		"role":     "admin",
	}

	resp, body := s.makeRequest("POST", "/api/v1/auth/register", payload, "")
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode, "Failed to create user")

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	s.userID = result["user_id"].(string)
	s.apiKey = result["api_key"].(string)
	s.projectID = result["default_project_id"].(string)

	fmt.Printf("1️⃣ User created: %s\n", result["email"])
	fmt.Printf("   • User ID: %s\n", s.userID)
	fmt.Printf("   • Project ID: %s\n", s.projectID)
}

// createWAHAChannel cria um canal WAHA de teste
func (s *ScheduledAutomationWebhookTestSuite) createWAHAChannel() {
	timestamp := time.Now().Unix()
	payload := map[string]interface{}{
		"name": fmt.Sprintf("Test Webhook Channel %d", timestamp),
		"type": "waha",
		"waha_config": map[string]interface{}{
			"session_id":  fmt.Sprintf("test-webhook-%d", timestamp),
			"base_url":    "https://waha.example.com",
			"api_key":     "test-waha-key",
			"webhook_url": "",
		},
	}

	endpoint := fmt.Sprintf("/api/v1/crm/channels?project_id=%s", s.projectID)
	resp, body := s.makeRequest("POST", endpoint, payload, s.apiKey)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode, "Failed to create channel")

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	s.channelID = result["id"].(string)

	fmt.Printf("2️⃣ Channel created: %s\n", result["name"])
	fmt.Printf("   • Channel ID: %s\n", s.channelID)
}

// getOrCreatePipeline busca ou cria um pipeline
func (s *ScheduledAutomationWebhookTestSuite) getOrCreatePipeline() {
	row := s.db.QueryRow("SELECT id FROM pipelines WHERE tenant_id = $1 LIMIT 1", s.projectID)
	err := row.Scan(&s.pipelineID)

	if err == nil {
		fmt.Printf("3️⃣ Using existing pipeline: %s\n", s.pipelineID)
		return
	}

	pipelineID := uuid.New().String()
	_, err = s.db.Exec(`
		INSERT INTO pipelines (id, tenant_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, pipelineID, s.projectID, "Test Pipeline", "Pipeline for webhook E2E tests")

	assert.NoError(s.T(), err, "Failed to create pipeline")
	s.pipelineID = pipelineID

	fmt.Printf("3️⃣ Pipeline created: %s\n", s.pipelineID)
}

// createContact cria um contato de teste
func (s *ScheduledAutomationWebhookTestSuite) createContact() {
	contactID := uuid.New().String()
	timestamp := time.Now().Unix()

	_, err := s.db.Exec(`
		INSERT INTO contacts (id, tenant_id, channel_id, pipeline_id, name, phone, whatsapp_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`, contactID, s.projectID, s.channelID, s.pipelineID,
		fmt.Sprintf("Test Contact %d", timestamp),
		fmt.Sprintf("55449704447%d", timestamp%1000),
		fmt.Sprintf("55449704447%d@c.us", timestamp%1000))

	assert.NoError(s.T(), err, "Failed to create contact")
	s.contactID = contactID

	fmt.Printf("4️⃣ Contact created: %s\n", s.contactID)
}

// createWebhookSubscription cria uma subscription para o evento automation.executed
func (s *ScheduledAutomationWebhookTestSuite) createWebhookSubscription() {
	payload := map[string]interface{}{
		"url":    s.webhookServer.URL,
		"events": []string{"automation.executed", "automation.failed"},
		"active": true,
	}

	endpoint := "/api/v1/webhooks/subscriptions"
	resp, body := s.makeRequest("POST", endpoint, payload, s.apiKey)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode, "Failed to create webhook subscription")

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	s.subscriptionID = result["id"].(string)

	fmt.Printf("5️⃣ Webhook subscription created: %s\n", s.subscriptionID)
	fmt.Printf("   • Events: automation.executed, automation.failed\n")
	fmt.Printf("   • URL: %s\n", s.webhookServer.URL)
}

// TestScheduledAutomationWithWebhook testa automation + webhook notification
func (s *ScheduledAutomationWebhookTestSuite) TestScheduledAutomationWithWebhook() {
	fmt.Println("\n🔔 Testing scheduled automation WITH webhook notification...")

	// 1. Limpa webhooks recebidos
	s.webhookMutex.Lock()
	s.webhookReceived = make([]map[string]interface{}, 0)
	s.webhookMutex.Unlock()

	// 2. Cria automation rule agendada para executar AGORA
	ruleID := s.createScheduledAutomation("webhook-test", map[string]interface{}{
		"type":   "once",
		"execute_at": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
	})
	s.ruleID = ruleID

	fmt.Printf("   • Rule ID: %s\n", ruleID)
	fmt.Println("   • Waiting 75 seconds for worker to process...")

	// 3. Aguarda worker processar (1 minuto de polling + margem + webhook delivery)
	time.Sleep(75 * time.Second)

	// 4. Verifica que rule foi executada
	var lastExecuted *time.Time
	err := s.db.QueryRow(`
		SELECT last_executed FROM automation_rules WHERE id = $1
	`, ruleID).Scan(&lastExecuted)

	assert.NoError(s.T(), err, "Failed to query rule")
	assert.NotNil(s.T(), lastExecuted, "Rule should have been executed")

	fmt.Printf("   ✅ Rule executed at: %s\n", lastExecuted.Format(time.RFC3339))

	// 5. Verifica que webhook foi enviado
	s.webhookMutex.Lock()
	webhooksReceived := len(s.webhookReceived)
	s.webhookMutex.Unlock()

	assert.GreaterOrEqual(s.T(), webhooksReceived, 1, "Should have received at least 1 webhook")

	if webhooksReceived > 0 {
		s.webhookMutex.Lock()
		firstWebhook := s.webhookReceived[0]
		s.webhookMutex.Unlock()

		fmt.Printf("   ✅ Webhook received: %d total\n", webhooksReceived)
		fmt.Printf("   ✅ Event type: %s\n", firstWebhook["event"])

		// Verifica conteúdo do webhook
		assert.Contains(s.T(), []string{"automation.executed", "automation.failed"},
			firstWebhook["event"], "Webhook should be automation event")

		if payload, ok := firstWebhook["payload"].(map[string]interface{}); ok {
			fmt.Printf("   ✅ Payload contains: rule_id=%s\n", payload["rule_id"])
		}
	}

	fmt.Println("✅ Scheduled automation WITH webhook processed successfully")
}

// createScheduledAutomation cria uma automation rule agendada
func (s *ScheduledAutomationWebhookTestSuite) createScheduledAutomation(name string, schedule map[string]interface{}) string {
	ruleID := uuid.New().String()

	scheduleJSON, err := json.Marshal(schedule)
	assert.NoError(s.T(), err)

	conditions := []map[string]interface{}{}
	conditionsJSON, _ := json.Marshal(conditions)

	actions := []map[string]interface{}{
		{
			"type": "send_message",
			"params": map[string]interface{}{
				"template_id": "welcome",
				"message":     "Scheduled automation test message with webhook",
			},
		},
	}
	actionsJSON, _ := json.Marshal(actions)

	nextExecution := time.Now().Add(-1 * time.Minute)

	_, err = s.db.Exec(`
		INSERT INTO automation_rules (
			id, tenant_id, pipeline_id, name, description,
			automation_type, trigger, conditions, actions, priority, enabled,
			schedule, next_execution, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
	`, ruleID, s.projectID, s.pipelineID,
		fmt.Sprintf("Test %s Automation", name),
		"E2E test automation with webhook",
		"scheduled", "scheduled",
		conditionsJSON, actionsJSON,
		1, true,
		scheduleJSON, nextExecution)

	assert.NoError(s.T(), err, "Failed to create automation rule")

	return ruleID
}

// makeRequest é um helper para fazer requisições HTTP
func (s *ScheduledAutomationWebhookTestSuite) makeRequest(method, endpoint string, payload interface{}, apiKey string) (*http.Response, []byte) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		assert.NoError(s.T(), err)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, s.baseURL+endpoint, body)
	assert.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)
	resp.Body.Close()

	return resp, respBody
}

// TestScheduledAutomationWebhookTestSuite executa a suite de testes
func TestScheduledAutomationWebhookTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduledAutomationWebhookTestSuite))
}
