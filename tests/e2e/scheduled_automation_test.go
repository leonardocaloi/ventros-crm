package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	_ "github.com/lib/pq"
)

// ScheduledAutomationTestSuite testa o fluxo completo de scheduled automations
type ScheduledAutomationTestSuite struct {
	suite.Suite
	baseURL    string
	client     *http.Client
	userID     string
	projectID  string
	apiKey     string
	channelID  string
	pipelineID string
	contactID  string
	ruleID     string
	db         *sql.DB
}

// SetupSuite executa uma vez antes de todos os testes
func (s *ScheduledAutomationTestSuite) SetupSuite() {
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

	// Aguarda API estar pronta
	s.waitForAPI()

	fmt.Println("\n🤖 Setting up Scheduled Automation E2E Test")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 1. Cria usuário
	s.createUser()

	// 2. Cria canal WAHA
	s.createWAHAChannel()

	// 3. Busca ou cria pipeline
	s.getOrCreatePipeline()

	// 4. Cria contato de teste
	s.createContact()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Setup completo!")
	fmt.Println("")
}

// TearDownSuite executa após todos os testes (CLEANUP)
func (s *ScheduledAutomationTestSuite) TearDownSuite() {
	fmt.Println("\n🧹 Cleaning up test data...")

	// Deleta automation rule se existir
	if s.ruleID != "" {
		_, err := s.db.Exec("DELETE FROM automation_rules WHERE id = $1", s.ruleID)
		if err == nil {
			fmt.Printf("  ✓ Deleted automation rule: %s\n", s.ruleID)
		}
	}

	// Deleta contato se existir
	if s.contactID != "" {
		_, err := s.db.Exec("DELETE FROM contacts WHERE id = $1", s.contactID)
		if err == nil {
			fmt.Printf("  ✓ Deleted contact: %s\n", s.contactID)
		}
	}

	// Deleta canal se existir
	if s.channelID != "" && s.apiKey != "" {
		endpoint := fmt.Sprintf("/api/v1/channels/%s", s.channelID)
		resp, _ := s.makeRequest("DELETE", endpoint, nil, s.apiKey)
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
			fmt.Printf("  ✓ Deleted channel: %s\n", s.channelID)
		}
	}

	if s.db != nil {
		s.db.Close()
	}

	fmt.Println("✅ Cleanup completed")
}

// waitForAPI aguarda a API estar disponível
func (s *ScheduledAutomationTestSuite) waitForAPI() {
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
func (s *ScheduledAutomationTestSuite) createUser() {
	timestamp := time.Now().Unix()
	payload := map[string]string{
		"name":     fmt.Sprintf("Test User Automation %d", timestamp),
		"email":    fmt.Sprintf("test-automation-%d@example.com", timestamp),
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
func (s *ScheduledAutomationTestSuite) createWAHAChannel() {
	timestamp := time.Now().Unix()
	payload := map[string]interface{}{
		"name": fmt.Sprintf("Test Automation Channel %d", timestamp),
		"type": "waha",
		"waha_config": map[string]interface{}{
			"session_id":  fmt.Sprintf("test-automation-%d", timestamp),
			"base_url":    "https://waha.example.com",
			"api_key":     "test-waha-key",
			"webhook_url": "",
		},
	}

	endpoint := fmt.Sprintf("/api/v1/channels?project_id=%s", s.projectID)
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
func (s *ScheduledAutomationTestSuite) getOrCreatePipeline() {
	// Tenta buscar pipeline existente
	row := s.db.QueryRow("SELECT id FROM pipelines WHERE tenant_id = $1 LIMIT 1", s.projectID)
	err := row.Scan(&s.pipelineID)

	if err == nil {
		fmt.Printf("3️⃣ Using existing pipeline: %s\n", s.pipelineID)
		return
	}

	// Se não encontrou, cria um novo pipeline
	pipelineID := uuid.New().String()
	_, err = s.db.Exec(`
		INSERT INTO pipelines (id, tenant_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, pipelineID, s.projectID, "Test Pipeline", "Pipeline for E2E tests")

	assert.NoError(s.T(), err, "Failed to create pipeline")
	s.pipelineID = pipelineID

	fmt.Printf("3️⃣ Pipeline created: %s\n", s.pipelineID)
}

// createContact cria um contato de teste
func (s *ScheduledAutomationTestSuite) createContact() {
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

// TestScheduledAutomationDailyExecution testa execução de automation agendada diária
func (s *ScheduledAutomationTestSuite) TestScheduledAutomationDailyExecution() {
	fmt.Println("\n📅 Testing DAILY scheduled automation...")

	// 1. Cria automation rule agendada para executar AGORA
	ruleID := s.createScheduledAutomation("daily", map[string]interface{}{
		"type":   "daily",
		"hour":   time.Now().Hour(),
		"minute": time.Now().Minute(),
	})
	s.ruleID = ruleID

	fmt.Printf("   • Rule ID: %s\n", ruleID)
	fmt.Println("   • Waiting 70 seconds for worker to process (polls every 1 minute)...")

	// 2. Aguarda worker processar (1 minuto de polling + margem)
	time.Sleep(70 * time.Second)

	// 3. Verifica que rule foi executada (last_executed atualizado)
	var lastExecuted *time.Time
	var nextExecution *time.Time
	err := s.db.QueryRow(`
		SELECT last_executed, next_execution
		FROM automation_rules
		WHERE id = $1
	`, ruleID).Scan(&lastExecuted, &nextExecution)

	assert.NoError(s.T(), err, "Failed to query rule")
	assert.NotNil(s.T(), lastExecuted, "Rule should have been executed")
	assert.NotNil(s.T(), nextExecution, "Rule should have next_execution set")

	fmt.Printf("   ✅ Rule executed at: %s\n", lastExecuted.Format(time.RFC3339))
	fmt.Printf("   ✅ Next execution: %s\n", nextExecution.Format(time.RFC3339))

	// 4. Verifica que next_execution é aproximadamente 24 horas no futuro
	expectedNext := time.Now().Add(23 * time.Hour) // Margem de 1 hora
	assert.True(s.T(), nextExecution.After(expectedNext),
		"Next execution should be ~24 hours in the future")

	fmt.Println("✅ DAILY scheduled automation processed successfully")
}

// TestScheduledAutomationWeeklyExecution testa execução semanal
func (s *ScheduledAutomationTestSuite) TestScheduledAutomationWeeklyExecution() {
	fmt.Println("\n📆 Testing WEEKLY scheduled automation...")

	// Cria automation rule semanal para o mesmo dia da semana de hoje
	today := time.Now().Weekday()
	weekdayMap := map[time.Weekday]string{
		time.Sunday:    "sunday",
		time.Monday:    "monday",
		time.Tuesday:   "tuesday",
		time.Wednesday: "wednesday",
		time.Thursday:  "thursday",
		time.Friday:    "friday",
		time.Saturday:  "saturday",
	}

	ruleID := s.createScheduledAutomation("weekly", map[string]interface{}{
		"type":    "weekly",
		"weekday": weekdayMap[today],
		"hour":    time.Now().Hour(),
		"minute":  time.Now().Minute(),
	})

	// Cleanup diferente para este teste
	defer func() {
		s.db.Exec("DELETE FROM automation_rules WHERE id = $1", ruleID)
	}()

	fmt.Printf("   • Rule ID: %s (weekday: %s)\n", ruleID, weekdayMap[today])
	fmt.Println("   • Waiting 70 seconds for worker...")

	time.Sleep(70 * time.Second)

	var lastExecuted *time.Time
	err := s.db.QueryRow(`
		SELECT last_executed FROM automation_rules WHERE id = $1
	`, ruleID).Scan(&lastExecuted)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), lastExecuted, "Weekly rule should have been executed")

	fmt.Printf("   ✅ Weekly rule executed at: %s\n", lastExecuted.Format(time.RFC3339))
	fmt.Println("✅ WEEKLY scheduled automation processed successfully")
}

// TestScheduledAutomationOnceExecution testa execução única
func (s *ScheduledAutomationTestSuite) TestScheduledAutomationOnceExecution() {
	fmt.Println("\n⏰ Testing ONCE scheduled automation...")

	// Cria automation rule para executar uma única vez AGORA
	executeAt := time.Now().Add(2 * time.Second)
	ruleID := s.createScheduledAutomation("once", map[string]interface{}{
		"type":       "once",
		"execute_at": executeAt.Format(time.RFC3339),
	})

	defer func() {
		s.db.Exec("DELETE FROM automation_rules WHERE id = $1", ruleID)
	}()

	fmt.Printf("   • Rule ID: %s\n", ruleID)
	fmt.Printf("   • Scheduled for: %s\n", executeAt.Format(time.RFC3339))
	fmt.Println("   • Waiting 70 seconds for worker...")

	time.Sleep(70 * time.Second)

	var lastExecuted *time.Time
	var nextExecution *time.Time
	err := s.db.QueryRow(`
		SELECT last_executed, next_execution
		FROM automation_rules
		WHERE id = $1
	`, ruleID).Scan(&lastExecuted, &nextExecution)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), lastExecuted, "Once rule should have been executed")
	assert.Nil(s.T(), nextExecution, "Once rule should NOT have next_execution (runs only once)")

	fmt.Printf("   ✅ Once rule executed at: %s\n", lastExecuted.Format(time.RFC3339))
	fmt.Println("   ✅ Next execution is NULL (as expected for ONCE type)")
	fmt.Println("✅ ONCE scheduled automation processed successfully")
}

// createScheduledAutomation cria uma automation rule agendada
func (s *ScheduledAutomationTestSuite) createScheduledAutomation(name string, schedule map[string]interface{}) string {
	ruleID := uuid.New().String()

	// Cria schedule JSON
	scheduleJSON, err := json.Marshal(schedule)
	assert.NoError(s.T(), err)

	// Cria conditions JSON (vazio para MVP)
	conditions := []map[string]interface{}{}
	conditionsJSON, _ := json.Marshal(conditions)

	// Cria actions JSON (send_message action)
	actions := []map[string]interface{}{
		{
			"type": "send_message",
			"params": map[string]interface{}{
				"template_id": "welcome",
				"message":     "Scheduled automation test message",
			},
		},
	}
	actionsJSON, _ := json.Marshal(actions)

	// Define next_execution para AGORA (para executar imediatamente)
	nextExecution := time.Now().Add(-1 * time.Minute) // 1 minuto no passado

	_, err = s.db.Exec(`
		INSERT INTO automation_rules (
			id, tenant_id, pipeline_id, name, description,
			automation_type, trigger, conditions, actions, priority, enabled,
			schedule, next_execution, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
	`, ruleID, s.projectID, s.pipelineID,
	   fmt.Sprintf("Test %s Automation", name),
	   "E2E test automation rule",
	   "scheduled", // automation_type
	   "scheduled", // trigger
	   conditionsJSON, actionsJSON,
	   1,    // priority
	   true, // enabled
	   scheduleJSON, nextExecution)

	assert.NoError(s.T(), err, "Failed to create automation rule")

	return ruleID
}

// makeRequest é um helper para fazer requisições HTTP
func (s *ScheduledAutomationTestSuite) makeRequest(method, endpoint string, payload interface{}, apiKey string) (*http.Response, []byte) {
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

// TestScheduledAutomationTestSuite executa a suite de testes
func TestScheduledAutomationTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduledAutomationTestSuite))
}
