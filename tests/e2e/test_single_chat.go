package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	baseURL       = "http://localhost:8080"
	wahaBaseURL   = "https://waha.ventros.cloud"
	wahaToken     = "4bffec302d5f4312b8b73700da3ff3cb"
	wahaSessionID = "guilherme-batilani-suporte"
	testChatID    = "554891147884@c.us" // Chat que sabemos que tem mensagens
	maxDays       = 180
)

type User struct {
	ID     string `json:"id"`
	APIKey string `json:"api_key"`
}

type Channel struct {
	ID string `json:"id"`
}

type ImportResponse struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
}

type ImportStatus struct {
	Status           string `json:"status"`
	MessagesImported int    `json:"messages_imported"`
}

func main() {
	log.Println("🔍 Teste de Chat Único")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 1. Criar usuário
	log.Println("\n1️⃣ Criando usuário...")
	user := createUser()
	log.Printf("   ✓ User ID: %s", user.ID)

	// 2. Criar canal
	log.Println("\n2️⃣ Criando canal...")
	channel := createChannel(user.APIKey)
	log.Printf("   ✓ Channel ID: %s", channel.ID)

	// 3. Iniciar importação
	log.Println("\n3️⃣ Iniciando importação...")
	importResp := startImport(user.APIKey, channel.ID)
	log.Printf("   ✓ Workflow ID: %s", importResp.WorkflowID)

	// 4. Aguardar
	log.Println("\n4️⃣ Aguardando (30s)...")
	time.Sleep(30 * time.Second)

	status := getImportStatus(user.APIKey, channel.ID)
	log.Printf("\n✅ Resultado:")
	log.Printf("   Status: %s", status.Status)
	log.Printf("   Mensagens: %d", status.MessagesImported)

	if status.MessagesImported > 0 {
		log.Println("\n🎉 SUCESSO! Mensagens foram importadas!")
		os.Exit(0)
	} else {
		log.Println("\n⚠️  Nenhuma mensagem importada - verifique os logs")
		os.Exit(1)
	}
}

func createUser() User {
	timestamp := time.Now().Unix()
	payload := map[string]string{
		"name":     fmt.Sprintf("Test User %d", timestamp),
		"email":    fmt.Sprintf("test-%d@ventros.com", timestamp),
		"password": "test123456",
	}
	resp := makeRequest("POST", "/api/v1/auth/register", "", payload)
	var user User
	json.Unmarshal(resp, &user)
	return user
}

func createChannel(apiKey string) Channel {
	payload := map[string]interface{}{
		"name": "Test Single Chat",
		"type": "waha",
		"waha_config": map[string]string{
			"base_url":   wahaBaseURL,
			"token":      wahaToken,
			"session_id": wahaSessionID,
		},
	}
	resp := makeRequest("POST", "/api/v1/crm/channels", apiKey, payload)
	var channel Channel
	json.Unmarshal(resp, &channel)

	// Habilitar history import
	cmd := fmt.Sprintf(`PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "UPDATE channels SET history_import_enabled=true, history_import_max_days=%d WHERE id='%s'"`, maxDays, channel.ID)
	exec.Command("bash", "-c", cmd).Run()

	return channel
}

func startImport(apiKey, channelID string) ImportResponse {
	payload := map[string]interface{}{
		"strategy":        "time_range",
		"time_range_days": maxDays,
	}
	resp := makeRequest("POST", fmt.Sprintf("/api/v1/crm/channels/%s/import-history", channelID), apiKey, payload)
	var importResp ImportResponse
	json.Unmarshal(resp, &importResp)
	return importResp
}

func getImportStatus(apiKey, channelID string) ImportStatus {
	resp := makeRequest("GET", fmt.Sprintf("/api/v1/crm/channels/%s/import-status", channelID), apiKey, nil)
	var status ImportStatus
	json.Unmarshal(resp, &status)
	return status
}

func makeRequest(method, path, apiKey string, payload interface{}) []byte {
	var body io.Reader
	if payload != nil {
		jsonData, _ := json.Marshal(payload)
		body = bytes.NewBuffer(jsonData)
	}

	req, _ := http.NewRequest(method, baseURL+path, body)
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return respBody
}
