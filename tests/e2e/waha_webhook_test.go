package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// WAHAWebhookTestSuite testa o fluxo completo de webhook WAHA
type WAHAWebhookTestSuite struct {
	suite.Suite
	baseURL    string
	client     *http.Client
	userID     string
	projectID  string
	apiKey     string
	channelID  string
	webhookURL string
}

// SetupSuite executa uma vez antes de todos os testes
func (s *WAHAWebhookTestSuite) SetupSuite() {
	// Configura URL base
	s.baseURL = os.Getenv("API_BASE_URL")
	if s.baseURL == "" {
		s.baseURL = "http://localhost:8080"
	}

	s.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Aguarda API estar pronta
	s.waitForAPI()

	fmt.Println("\n🚀 Setting up WAHA Webhook E2E Test")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// 1. Cria usuário
	s.createUser()
	
	// 2. Cria canal WAHA
	s.createWAHAChannel()
	
	// 3. Ativa canal
	s.activateChannel()
	
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Setup completo!")
	fmt.Printf("📍 Webhook URL: %s\n", s.webhookURL)
	fmt.Println("")
}

// TearDownSuite executa após todos os testes (CLEANUP)
func (s *WAHAWebhookTestSuite) TearDownSuite() {
	fmt.Println("\n🧹 Cleaning up test data...")
	
	if s.channelID != "" && s.apiKey != "" {
		endpoint := fmt.Sprintf("/api/v1/channels/%s", s.channelID)
		resp, _ := s.makeRequest("DELETE", endpoint, nil, s.apiKey)
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
			fmt.Printf("  ✓ Deleted channel: %s\n", s.channelID)
		}
	}
	
	fmt.Println("✅ Cleanup completed")
}

// waitForAPI aguarda a API estar disponível
func (s *WAHAWebhookTestSuite) waitForAPI() {
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
func (s *WAHAWebhookTestSuite) createUser() {
	timestamp := time.Now().Unix()
	payload := map[string]string{
		"name":     fmt.Sprintf("Test User WAHA %d", timestamp),
		"email":    fmt.Sprintf("test-waha-%d@example.com", timestamp),
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

// createWAHAChannel cria um canal WAHA de teste
func (s *WAHAWebhookTestSuite) createWAHAChannel() {
	timestamp := time.Now().Unix()
	payload := map[string]interface{}{
		"name": fmt.Sprintf("Test WAHA Channel %d", timestamp),
		"type": "waha",
		"waha_config": map[string]interface{}{
			"session_id":  fmt.Sprintf("test-session-%d", timestamp),
			"base_url":    "https://waha.example.com",
			"api_key":     "test-waha-key",
			"webhook_url": "", // Será preenchido automaticamente
		},
	}
	
	endpoint := fmt.Sprintf("/api/v1/channels?project_id=%s", s.projectID)
	resp, body := s.makeRequest("POST", endpoint, payload, s.apiKey)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode, "Failed to create channel")
	
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	
	s.channelID = result["id"].(string)
	
	// Busca canal para pegar webhook_url
	endpoint = fmt.Sprintf("/api/v1/channels/%s", s.channelID)
	resp, body = s.makeRequest("GET", endpoint, nil, s.apiKey)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var channelData map[string]interface{}
	err = json.Unmarshal(body, &channelData)
	assert.NoError(s.T(), err)
	
	channel := channelData["channel"].(map[string]interface{})
	s.webhookURL = channel["webhook_url"].(string)
	
	fmt.Printf("2️⃣ Channel created: %s\n", result["name"])
	fmt.Printf("   • Channel ID: %s\n", s.channelID)
	fmt.Printf("   • Webhook URL: %s\n", s.webhookURL)
}

// activateChannel ativa o canal
func (s *WAHAWebhookTestSuite) activateChannel() {
	endpoint := fmt.Sprintf("/api/v1/channels/%s/activate", s.channelID)
	resp, _ := s.makeRequest("POST", endpoint, nil, s.apiKey)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode, "Failed to activate channel")
	
	fmt.Printf("3️⃣ Channel activated: %s\n", s.channelID)
}

// TestTextMessage testa mensagem de texto
func (s *WAHAWebhookTestSuite) TestTextMessage() {
	fmt.Println("\n📝 Testing TEXT message...")
	
	event := s.loadEventFile("message_text.json")
	s.sendWebhookEvent(event)
	
	// Aguarda processamento
	time.Sleep(2 * time.Second)
	
	// Verifica que canal foi atualizado (ordem alfabética: 7º teste)
	s.verifyChannelStats(7)
	
	fmt.Println("✅ TEXT message processed")
}

// TestImageMessage testa mensagem de imagem
func (s *WAHAWebhookTestSuite) TestImageMessage() {
	fmt.Println("\n🖼️  Testing IMAGE message...")
	
	event := s.loadEventFile("message_image.json")
	s.sendWebhookEvent(event)
	
	time.Sleep(2 * time.Second)
	// Ordem alfabética: 4º teste
	s.verifyChannelStats(4)
	
	fmt.Println("✅ IMAGE message processed")
}

// TestVoiceMessage testa mensagem de voz (PTT)
func (s *WAHAWebhookTestSuite) TestVoiceMessage() {
	fmt.Println("\n🎤 Testing VOICE (PTT) message...")
	
	event := s.loadEventFile("message_recorded_audio.json")
	s.sendWebhookEvent(event)
	
	time.Sleep(2 * time.Second)
	// Ordem alfabética: 8º teste
	s.verifyChannelStats(8)
	
	fmt.Println("✅ VOICE message processed")
}

// TestLocationMessage testa mensagem de localização
func (s *WAHAWebhookTestSuite) TestLocationMessage() {
	fmt.Println("\n📍 Testing LOCATION message...")
	
	event := s.loadEventFile("message_location.json")
	s.sendWebhookEvent(event)
	
	time.Sleep(2 * time.Second)
	// Ordem alfabética: 6º teste
	s.verifyChannelStats(6)
	
	fmt.Println("✅ LOCATION message processed")
}

// TestContactMessage testa mensagem de contato
func (s *WAHAWebhookTestSuite) TestContactMessage() {
	fmt.Println("\n👤 Testing CONTACT message...")
	
	event := s.loadEventFile("message_contact.json")
	s.sendWebhookEvent(event)
	
	time.Sleep(2 * time.Second)
	// Ordem alfabética: 2º teste
	s.verifyChannelStats(2)
	
	fmt.Println("✅ CONTACT message processed")
}

// TestDocumentMessage testa mensagem de documento
func (s *WAHAWebhookTestSuite) TestDocumentMessage() {
	fmt.Println("\n📄 Testing DOCUMENT message...")
	
	event := s.loadEventFile("message_document_pdf.json")
	s.sendWebhookEvent(event)
	
	time.Sleep(2 * time.Second)
	// Ordem alfabética: 3º teste
	s.verifyChannelStats(3)
	
	fmt.Println("✅ DOCUMENT message processed")
}

// TestAudioMessage testa mensagem de áudio
func (s *WAHAWebhookTestSuite) TestAudioMessage() {
	fmt.Println("\n🔊 Testing AUDIO message...")
	
	event := s.loadEventFile("message_audio.json")
	s.sendWebhookEvent(event)
	
	time.Sleep(2 * time.Second)
	// Ordem alfabética: 1º teste
	s.verifyChannelStats(1)
	
	fmt.Println("✅ AUDIO message processed")
}

// TestImageWithTextMessage testa imagem com legenda
func (s *WAHAWebhookTestSuite) TestImageWithTextMessage() {
	fmt.Println("\n🖼️📝 Testing IMAGE with TEXT message...")
	
	event := s.loadEventFile("message_image_text.json")
	s.sendWebhookEvent(event)
	
	time.Sleep(2 * time.Second)
	// Ordem alfabética: 5º teste
	s.verifyChannelStats(5)
	
	fmt.Println("✅ IMAGE with TEXT message processed")
}

// loadEventFile carrega um arquivo de evento JSON
func (s *WAHAWebhookTestSuite) loadEventFile(filename string) map[string]interface{} {
	// Tenta vários caminhos possíveis
	paths := []string{
		filepath.Join("../../events_waha", filename),
		filepath.Join("events_waha", filename),
		filepath.Join("../events_waha", filename),
	}
	
	var data []byte
	var err error
	
	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		s.T().Fatalf("Failed to load event file %s: %v", filename, err)
	}
	
	var event map[string]interface{}
	err = json.Unmarshal(data, &event)
	assert.NoError(s.T(), err, "Failed to parse event JSON")
	
	// Atualiza session para o canal de teste
	if payload, ok := event["payload"].(map[string]interface{}); ok {
		if data, ok := payload["_data"].(map[string]interface{}); ok {
			if info, ok := data["Info"].(map[string]interface{}); ok {
				// Extrai session_id do webhook_url
				// URL format: /api/v1/webhooks/waha/{session_id}
				sessionID := s.extractSessionFromURL(s.webhookURL)
				event["session"] = sessionID
				info["Chat"] = sessionID + "@s.whatsapp.net"
			}
		}
	}
	
	return event
}

// extractSessionFromURL extrai session_id da webhook URL
func (s *WAHAWebhookTestSuite) extractSessionFromURL(url string) string {
	// URL format: http://localhost:8080/api/v1/webhooks/waha/{session_id}
	// Extrai o último segmento do path
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "test-session"
}

// sendWebhookEvent envia evento para o webhook
func (s *WAHAWebhookTestSuite) sendWebhookEvent(event map[string]interface{}) {
	jsonData, err := json.Marshal(event)
	assert.NoError(s.T(), err)
	
	req, err := http.NewRequest("POST", s.webhookURL, bytes.NewBuffer(jsonData))
	assert.NoError(s.T(), err)
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.client.Do(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode, 
		"Webhook should return 200. Response: %s", string(body))
	
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	
	assert.Equal(s.T(), "queued", result["status"], "Event should be queued")
}

// verifyChannelStats verifica estatísticas do canal
func (s *WAHAWebhookTestSuite) verifyChannelStats(expectedMessages int) {
	endpoint := fmt.Sprintf("/api/v1/channels/%s", s.channelID)
	resp, body := s.makeRequest("GET", endpoint, nil, s.apiKey)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	
	channel := result["channel"].(map[string]interface{})
	messagesReceived := int(channel["messages_received"].(float64))
	
	assert.GreaterOrEqual(s.T(), messagesReceived, expectedMessages, 
		"Channel should have at least %d messages", expectedMessages)
	
	fmt.Printf("   📊 Channel stats: %d messages received\n", messagesReceived)
}

// makeRequest é um helper para fazer requisições HTTP
func (s *WAHAWebhookTestSuite) makeRequest(method, endpoint string, payload interface{}, apiKey string) (*http.Response, []byte) {
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

// TestWAHAWebhookTestSuite executa a suite de testes
func TestWAHAWebhookTestSuite(t *testing.T) {
	suite.Run(t, new(WAHAWebhookTestSuite))
}
