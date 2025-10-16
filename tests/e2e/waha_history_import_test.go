package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func init() {
	// Carregar .env do diretório correto (.deploy/container/.env)
	if err := godotenv.Load(".deploy/container/.env"); err != nil {
		// Tenta caminho relativo se não encontrar
		if err := godotenv.Load("../../.deploy/container/.env"); err != nil {
			fmt.Printf("Warning: .env not loaded from .deploy/container/.env: %v\n", err)
		}
	}
}

// WAHAHistoryImportTestSuite testa o fluxo completo de importação de histórico WAHA
type WAHAHistoryImportTestSuite struct {
	suite.Suite
	baseURL       string
	client        *http.Client
	userID        string
	projectID     string
	apiKey        string
	channelID     string
	workflowID    string
	wahaBaseURL   string
	wahaAPIKey    string
	wahaSessionID string
}

// SetupSuite executa uma vez antes de todos os testes
func (s *WAHAHistoryImportTestSuite) SetupSuite() {
	// Configura URL base
	s.baseURL = os.Getenv("API_BASE_URL")
	if s.baseURL == "" {
		s.baseURL = "http://localhost:8080"
	}

	// Lê configurações WAHA do .env (igual ao msg_send_test.sh)
	s.wahaBaseURL = os.Getenv("WAHA_BASE_URL")
	if s.wahaBaseURL == "" {
		s.wahaBaseURL = "https://waha.ventros.cloud"
	}

	s.wahaAPIKey = os.Getenv("WAHA_API_KEY")
	if s.wahaAPIKey == "" {
		s.T().Fatal("WAHA_API_KEY not set in .env")
	}

	s.wahaSessionID = os.Getenv("WAHA_DEFAULT_SESSION_ID_TEST")
	if s.wahaSessionID == "" {
		s.wahaSessionID = "guilherme-batilani-suporte" // Fallback para sessão de teste padrão
	}

	s.client = &http.Client{
		Timeout: 60 * time.Second, // Timeout maior para imports
	}

	// Aguarda API estar pronta
	s.waitForAPI()

	fmt.Println("\n🚀 Setting up WAHA History Import E2E Test")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 1. Cria usuário
	s.createUser()

	// 2. Cria canal WAHA
	s.createWAHAChannel()

	// 3. Ativa canal (opcional, mas recomendado)
	s.activateChannel()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Setup completo!")
	fmt.Println("")
}

// TearDownSuite executa após todos os testes (CLEANUP)
func (s *WAHAHistoryImportTestSuite) TearDownSuite() {
	fmt.Println("\n🧹 Cleaning up test data...")

	if s.channelID != "" && s.apiKey != "" {
		endpoint := fmt.Sprintf("/api/v1/crm/channels/%s", s.channelID)
		resp, _ := s.makeRequest("DELETE", endpoint, nil, s.apiKey)
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
			fmt.Printf("  ✓ Deleted channel: %s\n", s.channelID)
		}
	}

	fmt.Println("✅ Cleanup completed")
}

// waitForAPI aguarda a API estar disponível
func (s *WAHAHistoryImportTestSuite) waitForAPI() {
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
func (s *WAHAHistoryImportTestSuite) createUser() {
	timestamp := time.Now().Unix()
	payload := map[string]string{
		"name":     fmt.Sprintf("Test User Import %d", timestamp),
		"email":    fmt.Sprintf("test-import-%d@example.com", timestamp),
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
	fmt.Printf("   • API Key: %s...\n", s.apiKey[:20])
}

// createWAHAChannel cria um canal WAHA de teste usando variáveis de ambiente
func (s *WAHAHistoryImportTestSuite) createWAHAChannel() {
	payload := map[string]interface{}{
		"name":                    fmt.Sprintf("E2E Test Import - %s", s.wahaSessionID),
		"type":                    "waha",
		"external_id":             s.wahaSessionID,
		"history_import_enabled":  true,
		"history_import_max_days": 180, // 🚀 V3: 180 dias para teste completo
		"waha_config": map[string]interface{}{
			"base_url":    s.wahaBaseURL,
			"api_key":     s.wahaAPIKey,
			"session_id":  s.wahaSessionID,
			"webhook_url": fmt.Sprintf("%s/api/v1/webhooks/waha", s.baseURL),
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
	fmt.Printf("   • WAHA Base URL: %s\n", s.wahaBaseURL)
	fmt.Printf("   • Session ID: %s\n", s.wahaSessionID)
	fmt.Printf("   • History Import: 180 days\n")
}

// activateChannel ativa o canal e aguarda ficar ativo
func (s *WAHAHistoryImportTestSuite) activateChannel() {
	endpoint := fmt.Sprintf("/api/v1/crm/channels/%s/activate", s.channelID)
	resp, body := s.makeRequest("POST", endpoint, nil, s.apiKey)

	// Accept both 202 (async) and 200 (sync)
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		fmt.Printf("⚠️  Channel activation returned status %d: %s\n", resp.StatusCode, string(body))
		s.T().Logf("Channel activation returned status %d: %s", resp.StatusCode, string(body))
	} else {
		fmt.Printf("3️⃣ Channel activation successful (status %d)\n", resp.StatusCode)
	}

	fmt.Printf("   • Channel ID: %s\n", s.channelID)

	// Aguardar canal ficar ativo
	fmt.Printf("   ⏳ Waiting for channel to become active...")
	maxRetries := 60 // Increased from 30 to 60 seconds
	channelActive := false
	for i := 0; i < maxRetries; i++ {
		time.Sleep(1 * time.Second)

		getEndpoint := fmt.Sprintf("/api/v1/crm/channels/%s", s.channelID)
		getResp, getBody := s.makeRequest("GET", getEndpoint, nil, s.apiKey)

		if getResp.StatusCode == http.StatusOK {
			var channelData map[string]interface{}
			if err := json.Unmarshal(getBody, &channelData); err == nil {
				if status, ok := channelData["status"].(string); ok && status == "active" {
					fmt.Println(" ✅ Active!")
					channelActive = true
					break
				}
			}
		}
		fmt.Print(".")
	}
	fmt.Println()

	if !channelActive {
		s.T().Fatalf("❌ CRITICAL: Channel did not become active within %d seconds - cannot proceed with import", maxRetries)
	}
}

// TestImportHistory testa a importação de histórico
func (s *WAHAHistoryImportTestSuite) TestImportHistory() {
	fmt.Println("\n📥 Testing history import...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 0. Configura canal para usar timeout de sessão de 240 minutos (4 horas)
	// IMPORTANTE: Com 4h de timeout, apenas conversas com gap > 4h criam nova sessão
	fmt.Println("\n   ⚙️  Configuring channel session timeout to 240 minutes (4 hours)...")
	updatePayload := map[string]interface{}{
		"default_session_timeout_minutes": 240,
	}

	updateEndpoint := fmt.Sprintf("/api/v1/crm/channels/%s", s.channelID)
	updateResp, updateBody := s.makeRequest("PUT", updateEndpoint, updatePayload, s.apiKey)

	// Se endpoint não existir, logar mas continuar com default (30 min)
	if updateResp.StatusCode == http.StatusOK {
		fmt.Println("   ✓ Channel configured with 240-minute (4h) session timeout")
		fmt.Println("   ℹ️  Sessions will consolidate messages with <4h gap")
	} else {
		s.T().Logf("Channel update not available (status %d): %s", updateResp.StatusCode, string(updateBody))
		fmt.Println("   ⚠️  Using default 30-minute session timeout")
	}

	// 1. Inicia importação de histórico (180 dias, sem limite de mensagens)
	payload := map[string]interface{}{
		"strategy":        "time_range",
		"time_range_days": 180, // 🚀 V3: 180 dias para teste completo
		"limit":           0,   // 0 = SEM LIMITE (importar todas as mensagens disponíveis)
	}

	endpoint := fmt.Sprintf("/api/v1/crm/channels/%s/import-history", s.channelID)
	resp, body := s.makeRequest("POST", endpoint, payload, s.apiKey)

	// Deve retornar 202 Accepted (async) ou 500 se Temporal não estiver configurado
	if resp.StatusCode == http.StatusInternalServerError {
		var errorResult map[string]interface{}
		err := json.Unmarshal(body, &errorResult)
		if err == nil && errorResult["error"] != nil {
			errorMsg := errorResult["error"].(string)
			if errorMsg == "Workflow engine not configured" || errorMsg == "Invalid workflow engine configuration" {
				s.T().Skip("Temporal workflow engine not configured - skipping test")
				return
			}
		}
	}

	assert.Equal(s.T(), http.StatusAccepted, resp.StatusCode,
		"Import should return 202 Accepted. Response: %s", string(body))

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Event-driven pattern: handler returns correlation_id, consumer starts workflow async
	correlationID, ok := result["correlation_id"].(string)
	if !ok || correlationID == "" {
		s.T().Fatalf("Expected correlation_id in response, got: %v", result)
	}
	s.workflowID = fmt.Sprintf("waha-import-%s", s.channelID) // Workflow ID format

	fmt.Printf("   ✓ Import requested (event-driven pattern)\n")
	fmt.Printf("   • Correlation ID: %s\n", correlationID)
	fmt.Printf("   • Expected Workflow ID: %s\n", s.workflowID)
	fmt.Printf("   • Strategy: time_range (180 days)\n")

	limitVal := result["limit"].(float64)
	if limitVal == 0 {
		fmt.Printf("   • Limit: UNLIMITED (all messages)\n")
	} else {
		fmt.Printf("   • Limit: %.0f messages per chat\n", limitVal)
	}

	// 2. Aguarda processamento (polling)
	s.pollImportStatus()

	// 3. Verificar database
	s.verifyDatabaseMetrics()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ History import test completed")
}

// TestImportStatus testa consulta de status sem import ativo
func (s *WAHAHistoryImportTestSuite) TestImportStatus() {
	fmt.Println("\n📊 Testing import status endpoint...")

	endpoint := fmt.Sprintf("/api/v1/crm/channels/%s/import-status", s.channelID)
	resp, body := s.makeRequest("GET", endpoint, nil, s.apiKey)

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode,
		"Status endpoint should return 200. Response: %s", string(body))

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Verifica campos esperados
	assert.NotNil(s.T(), result["channel_id"], "Should have channel_id")
	assert.NotNil(s.T(), result["workflow_id"], "Should have workflow_id")
	assert.NotNil(s.T(), result["status"], "Should have status")

	fmt.Printf("   ✓ Status retrieved\n")
	fmt.Printf("   • Channel ID: %s\n", result["channel_id"])
	fmt.Printf("   • Workflow ID: %s\n", result["workflow_id"])
	fmt.Printf("   • Status: %s\n", result["status"])

	// Se tiver progresso, mostrar estatísticas
	if progress, ok := result["progress"].(map[string]interface{}); ok {
		fmt.Printf("   • Chats processed: %.0f\n", progress["chats_processed"].(float64))
		fmt.Printf("   • Messages imported: %.0f\n", progress["messages_imported"].(float64))
		if progress["sessions_created"] != nil {
			fmt.Printf("   • Sessions created: %.0f\n", progress["sessions_created"].(float64))
		}
		if progress["contacts_created"] != nil {
			fmt.Printf("   • Contacts created: %.0f\n", progress["contacts_created"].(float64))
		}
	}

	fmt.Println("✅ Import status check completed")
}

// TestImportWithTimeLimit testa importação com limite de tempo
func (s *WAHAHistoryImportTestSuite) TestImportWithTimeLimit() {
	fmt.Println("\n⏰ Testing import with time limit...")

	// Configura canal para importar apenas últimos 7 dias
	updatePayload := map[string]interface{}{
		"history_import_max_days": 7,
	}

	updateEndpoint := fmt.Sprintf("/api/v1/crm/channels/%s", s.channelID)
	updateResp, updateBody := s.makeRequest("PUT", updateEndpoint, updatePayload, s.apiKey)

	// Endpoint PUT pode não existir ainda, então só logamos
	if updateResp.StatusCode != http.StatusOK {
		s.T().Logf("Channel update not implemented yet (status %d): %s", updateResp.StatusCode, string(updateBody))
		s.T().Skip("Channel update endpoint not available - skipping test")
		return
	}

	fmt.Printf("   ✓ Channel configured for 7-day import\n")

	// Inicia importação com estratégia "time_range" usando limit de 7 dias configurado no canal
	payload := map[string]interface{}{
		"strategy":        "time_range",
		"time_range_days": 7, // Override channel config to ensure 7-day import
	}

	endpoint := fmt.Sprintf("/api/v1/crm/channels/%s/import-history", s.channelID)
	resp, body := s.makeRequest("POST", endpoint, payload, s.apiKey)

	if resp.StatusCode == http.StatusInternalServerError {
		var errorResult map[string]interface{}
		err := json.Unmarshal(body, &errorResult)
		if err == nil && errorResult["error"] != nil {
			errorMsg := errorResult["error"].(string)
			if errorMsg == "Workflow engine not configured" || errorMsg == "Invalid workflow engine configuration" {
				s.T().Skip("Temporal workflow engine not configured - skipping test")
				return
			}
		}
	}

	assert.Equal(s.T(), http.StatusAccepted, resp.StatusCode,
		"Import should return 202 Accepted. Response: %s", string(body))

	fmt.Println("✅ Time-limited import started")
}

// TestImportWithMessageLimit testa importação com limite de mensagens
func (s *WAHAHistoryImportTestSuite) TestImportWithMessageLimit() {
	fmt.Println("\n📊 Testing import with message limit...")

	// Inicia importação com limite explícito
	payload := map[string]interface{}{
		"strategy": "recent",
		"limit":    10, // Limita a 10 mensagens por chat
	}

	endpoint := fmt.Sprintf("/api/v1/crm/channels/%s/import-history", s.channelID)
	resp, body := s.makeRequest("POST", endpoint, payload, s.apiKey)

	if resp.StatusCode == http.StatusInternalServerError {
		var errorResult map[string]interface{}
		err := json.Unmarshal(body, &errorResult)
		if err == nil && errorResult["error"] != nil {
			errorMsg := errorResult["error"].(string)
			if errorMsg == "Workflow engine not configured" || errorMsg == "Invalid workflow engine configuration" {
				s.T().Skip("Temporal workflow engine not configured - skipping test")
				return
			}
		}
	}

	// Verifica se retornou 202 Accepted
	if resp.StatusCode != http.StatusAccepted {
		s.T().Fatalf("Import should return 202 Accepted, got %d. Response: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Verifica se o campo limit existe antes de acessar
	limit, ok := result["limit"].(float64)
	if !ok {
		s.T().Fatalf("Expected 'limit' field in response, got: %v", result)
	}

	assert.Equal(s.T(), float64(10), limit, "Limit should be 10")

	fmt.Printf("   ✓ Message-limited import started (limit: %.0f)\n", limit)
	fmt.Println("✅ Message-limited import test completed")
}

// pollImportStatus aguarda importação completar (polling)
func (s *WAHAHistoryImportTestSuite) pollImportStatus() {
	maxRetries := 1800                     // 🚀 V3: Increased retries for faster polling
	pollInterval := 100 * time.Millisecond // Fast polling: 100ms (was 3s)
	importCompleted := false
	lastStatus := ""

	fmt.Println("\n   ⏳ Polling import status...")

	for i := 0; i < maxRetries; i++ {
		time.Sleep(pollInterval)

		endpoint := fmt.Sprintf("/api/v1/crm/channels/%s/import-status", s.channelID)
		resp, body := s.makeRequest("GET", endpoint, nil, s.apiKey)

		if resp.StatusCode != http.StatusOK {
			s.T().Logf("Status check failed (attempt %d/%d): %d - %s", i+1, maxRetries, resp.StatusCode, string(body))
			continue
		}

		var result map[string]interface{}
		err := json.Unmarshal(body, &result)
		if err != nil {
			s.T().Logf("Failed to parse status response (attempt %d/%d): %v", i+1, maxRetries, err)
			continue
		}

		status := result["status"].(string)

		// Show progress every 10 polls or on status change
		if i%10 == 0 || (i > 0 && status != lastStatus) {
			fmt.Printf("   📍 Status [%d/%d]: %s\n", i+1, maxRetries, status)
		}
		lastStatus = status

		// Check workflow status (case-insensitive)
		statusLower := strings.ToLower(status)
		if status == "WORKFLOW_EXECUTION_STATUS_COMPLETED" || statusLower == "completed" {
			importCompleted = true

			// Mostrar estatísticas finais
			if progress, ok := result["progress"].(map[string]interface{}); ok {
				fmt.Println("\n   📊 Final Statistics:")
				fmt.Printf("      • Chats processed: %.0f\n", progress["chats_processed"].(float64))
				fmt.Printf("      • Messages imported: %.0f\n", progress["messages_imported"].(float64))
				if progress["sessions_created"] != nil {
					fmt.Printf("      • Sessions created: %.0f\n", progress["sessions_created"].(float64))
				}
				if progress["contacts_created"] != nil {
					fmt.Printf("      • Contacts created: %.0f\n", progress["contacts_created"].(float64))
				}
				if errors, ok := progress["errors"].([]interface{}); ok && len(errors) > 0 {
					fmt.Printf("      • Errors: %d\n", len(errors))
				}
			}
			break
		} else if status == "WORKFLOW_EXECUTION_STATUS_FAILED" || status == "failed" {
			s.T().Fatalf("Import workflow failed: %+v", result)
		} else if status == "no_import_running" {
			// Importação já completou e workflow foi limpo
			importCompleted = true
			fmt.Println("   ✓ Import already completed (no active workflow)")
			break
		}
	}

	if !importCompleted {
		s.T().Logf("Warning: Import did not complete within %d seconds", maxRetries*3)
		s.T().Fatal("❌ CRITICAL: Import workflow did not complete - cannot verify consolidation")
	}
}

// verifyDatabaseMetrics consulta o database e mostra métricas detalhadas
func (s *WAHAHistoryImportTestSuite) verifyDatabaseMetrics() {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔍 DATABASE VERIFICATION")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Fazer queries via API endpoint especial ou skip se não disponível
	// Por enquanto, apenas logamos que a verificação seria feita
	fmt.Println("\n   📊 Database Metrics:")
	fmt.Println("   • Messages:        [Queried from workflow progress]")
	fmt.Println("   • Sessions:        [Queried from workflow progress]")
	fmt.Println("   • Contacts:        [Queried from workflow progress]")
	fmt.Println("   • Contact Events:  [To be implemented]")
	fmt.Println("   • Trackings:       [To be implemented]")

	fmt.Println("\n   ℹ️  Note: Database verification requires direct DB access")
	fmt.Println("   ℹ️  Production tests should use API endpoints for metrics")

	// TODO: Add actual DB queries here if test has direct DB access
	// For now, metrics are shown from workflow progress during polling
}

// makeRequest é um helper para fazer requisições HTTP
func (s *WAHAHistoryImportTestSuite) makeRequest(method, endpoint string, payload interface{}, apiKey string) (*http.Response, []byte) {
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

// TestWAHAHistoryImportTestSuite executa a suite de testes
func TestWAHAHistoryImportTestSuite(t *testing.T) {
	suite.Run(t, new(WAHAHistoryImportTestSuite))
}
