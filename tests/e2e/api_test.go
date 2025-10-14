package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// APITestSuite é a suite de testes E2E
type APITestSuite struct {
	suite.Suite
	baseURL   string
	client    *http.Client
	fixtures  *TestFixtures
	createdIDs map[string]string // Rastreia IDs criados para cleanup
}

// SetupSuite executa uma vez antes de todos os testes
func (s *APITestSuite) SetupSuite() {
	// Configura URL base (pode ser sobrescrita por env var)
	s.baseURL = os.Getenv("API_BASE_URL")
	if s.baseURL == "" {
		s.baseURL = "http://localhost:8080"
	}

	s.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	s.fixtures = GetDefaultFixtures()
	s.createdIDs = make(map[string]string)

	// Aguarda API estar pronta
	s.waitForAPI()
}

// TearDownSuite executa uma vez após todos os testes (CLEANUP)
func (s *APITestSuite) TearDownSuite() {
	fmt.Println("\n🧹 Cleaning up test data...")
	
	// Cleanup em ordem reversa (do mais dependente ao menos dependente)
	s.cleanupChannels()
	s.cleanupContacts()
	s.cleanupUsers()
	
	fmt.Println("✅ Cleanup completed")
}

// SetupTest executa antes de cada teste
func (s *APITestSuite) SetupTest() {
	// Pode adicionar setup específico por teste aqui
}

// waitForAPI aguarda a API estar disponível
func (s *APITestSuite) waitForAPI() {
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

// Test1_CreateUser testa criação de usuário (cria projeto e pipeline automático)
func (s *APITestSuite) Test1_CreateUser() {
	userFixture := s.fixtures.Users[0]
	
	payload := map[string]string{
		"name":     userFixture.Name,
		"email":    userFixture.Email,
		"password": userFixture.Password,
		"role":     userFixture.Role,
	}
	
	resp, body := s.makeRequest("POST", "/api/v1/auth/register", payload, "")
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	
	// Valida que recebeu todos os dados esperados
	assert.NotEmpty(s.T(), result["user_id"])
	assert.NotEmpty(s.T(), result["api_key"])
	assert.NotEmpty(s.T(), result["default_project_id"])
	assert.NotEmpty(s.T(), result["default_pipeline_id"])
	
	// Salva IDs para uso posterior e cleanup
	s.createdIDs["user_id"] = result["user_id"].(string)
	s.createdIDs["api_key"] = result["api_key"].(string)
	s.createdIDs["project_id"] = result["default_project_id"].(string)
	s.createdIDs["pipeline_id"] = result["default_pipeline_id"].(string)
	
	fmt.Printf("✅ User created: %s (Project: %s, Pipeline: %s)\n", 
		result["email"], 
		result["default_project_id"], 
		result["default_pipeline_id"])
}

// Test2_CreateChannel testa criação de canal
func (s *APITestSuite) Test2_CreateChannel() {
	channelFixture := s.fixtures.Channels[0]
	apiKey := s.createdIDs["api_key"]
	projectID := s.createdIDs["project_id"]
	
	assert.NotEmpty(s.T(), apiKey, "API key must be set from Test1")
	assert.NotEmpty(s.T(), projectID, "Project ID must be set from Test1")
	
	payload := map[string]interface{}{
		"name":        channelFixture.Name,
		"type":        channelFixture.Type,
		"waha_config": channelFixture.WAHAConfig,
	}
	
	endpoint := fmt.Sprintf("/api/v1/crm/channels?project_id=%s", projectID)
	resp, body := s.makeRequest("POST", endpoint, payload, apiKey)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	
	assert.NotEmpty(s.T(), result["id"])
	s.createdIDs["channel_id"] = result["id"].(string)
	
	fmt.Printf("✅ Channel created: %s (ID: %s)\n", result["name"], result["id"])
}

// Test3_ActivateChannel testa ativação de canal (Event-Driven Architecture)
func (s *APITestSuite) Test3_ActivateChannel() {
	apiKey := s.createdIDs["api_key"]
	channelID := s.createdIDs["channel_id"]

	assert.NotEmpty(s.T(), apiKey, "API key must be set")
	assert.NotEmpty(s.T(), channelID, "Channel ID must be set from Test2")

	// Step 1: Request activation (should return 202 Accepted immediately)
	endpoint := fmt.Sprintf("/api/v1/crm/channels/%s/activate", channelID)
	resp, body := s.makeRequest("POST", endpoint, nil, apiKey)
	assert.Equal(s.T(), http.StatusAccepted, resp.StatusCode, "Should return 202 Accepted for async activation")

	var activationResponse map[string]interface{}
	err := json.Unmarshal(body, &activationResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "activating", activationResponse["status"], "Status should be 'activating'")

	fmt.Printf("✅ Channel activation requested (async): %s\n", channelID)
	fmt.Println("   Waiting for async processing...")

	// Step 2: Poll channel status until it becomes 'active' (max 10 seconds)
	maxRetries := 20
	pollInterval := 500 * time.Millisecond
	channelActive := false

	for i := 0; i < maxRetries; i++ {
		time.Sleep(pollInterval)

		getEndpoint := fmt.Sprintf("/api/v1/crm/channels/%s", channelID)
		getResp, getBody := s.makeRequest("GET", getEndpoint, nil, apiKey)
		assert.Equal(s.T(), http.StatusOK, getResp.StatusCode)

		var channel map[string]interface{}
		err := json.Unmarshal(getBody, &channel)
		assert.NoError(s.T(), err)

		status := channel["status"].(string)
		if status == "active" {
			channelActive = true
			fmt.Printf("✅ Channel activated successfully via Event-Driven Architecture (took %dms)\n",
				int((i+1)*int(pollInterval.Milliseconds())))
			break
		} else if status == "inactive" {
			// Activation failed
			lastError := ""
			if channel["last_error"] != nil {
				lastError = channel["last_error"].(string)
			}
			s.T().Fatalf("Channel activation failed: %s", lastError)
		}
		// Status still "activating", continue polling
	}

	assert.True(s.T(), channelActive, "Channel should be activated within 10 seconds")
}

// Test4_CreateContact testa criação de contato
func (s *APITestSuite) Test4_CreateContact() {
	contactFixture := s.fixtures.Contacts[0]
	apiKey := s.createdIDs["api_key"]
	projectID := s.createdIDs["project_id"]
	
	assert.NotEmpty(s.T(), apiKey, "API key must be set")
	assert.NotEmpty(s.T(), projectID, "Project ID must be set")
	
	payload := map[string]string{
		"name":  contactFixture.Name,
		"phone": contactFixture.Phone,
		"email": contactFixture.Email,
	}
	
	endpoint := fmt.Sprintf("/api/v1/contacts?project_id=%s", projectID)
	resp, body := s.makeRequest("POST", endpoint, payload, apiKey)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	
	assert.NotEmpty(s.T(), result["id"])
	s.createdIDs["contact_id"] = result["id"].(string)
	
	fmt.Printf("✅ Contact created: %s (ID: %s)\n", result["name"], result["id"])
}

// Test5_ListContacts testa listagem de contatos
func (s *APITestSuite) Test5_ListContacts() {
	apiKey := s.createdIDs["api_key"]
	projectID := s.createdIDs["project_id"]
	
	endpoint := fmt.Sprintf("/api/v1/contacts?project_id=%s", projectID)
	resp, body := s.makeRequest("GET", endpoint, nil, apiKey)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	
	contacts, ok := result["contacts"].([]interface{})
	assert.True(s.T(), ok)
	assert.GreaterOrEqual(s.T(), len(contacts), 1, "Deve ter pelo menos 1 contato")
	
	fmt.Printf("✅ Listed %d contacts\n", len(contacts))
}

// makeRequest é um helper para fazer requisições HTTP
func (s *APITestSuite) makeRequest(method, endpoint string, payload interface{}, apiKey string) (*http.Response, []byte) {
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

// Cleanup helpers
func (s *APITestSuite) cleanupChannels() {
	channelID := s.createdIDs["channel_id"]
	if channelID == "" {
		return
	}
	
	apiKey := s.createdIDs["api_key"]
	endpoint := fmt.Sprintf("/api/v1/crm/channels/%s", channelID)
	resp, _ := s.makeRequest("DELETE", endpoint, nil, apiKey)
	
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		fmt.Printf("  ✓ Deleted channel: %s\n", channelID)
	}
}

func (s *APITestSuite) cleanupContacts() {
	contactID := s.createdIDs["contact_id"]
	if contactID == "" {
		return
	}
	
	apiKey := s.createdIDs["api_key"]
	endpoint := fmt.Sprintf("/api/v1/contacts/%s", contactID)
	resp, _ := s.makeRequest("DELETE", endpoint, nil, apiKey)
	
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		fmt.Printf("  ✓ Deleted contact: %s\n", contactID)
	}
}

func (s *APITestSuite) cleanupUsers() {
	// Nota: Implementar endpoint DELETE /api/v1/users/:id se necessário
	// Por enquanto, apenas log
	userID := s.createdIDs["user_id"]
	if userID != "" {
		fmt.Printf("  ⚠ User cleanup not implemented yet: %s\n", userID)
		fmt.Println("  💡 Run: make test-cleanup-db para limpar banco de teste")
	}
}

// TestAPITestSuite executa a suite de testes
func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
