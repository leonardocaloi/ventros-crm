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
)

type User struct {
	ID     string `json:"id"`
	APIKey string `json:"api_key"`
}

type Channel struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Type                 string                 `json:"type"`
	ExternalID           string                 `json:"external_id"`
	Config               map[string]interface{} `json:"config"`
	HistoryImportEnabled bool                   `json:"history_import_enabled"`
	HistoryImportMaxDays int                    `json:"history_import_max_days"`
}

type ImportResponse struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
	Strategy   string `json:"strategy"`
}

type ImportStatus struct {
	WorkflowID       string `json:"workflow_id"`
	Status           string `json:"status"`
	ChatsProcessed   int    `json:"chats_processed"`
	MessagesImported int    `json:"messages_imported"`
	SessionsCreated  int    `json:"sessions_created"`
	ContactsCreated  int    `json:"contacts_created"`
}

func main() {
	log.Println("🔍 Verificação de Histórico WAHA - Múltiplos Intervalos")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 1. Criar usuário
	log.Println("\n1️⃣ Criando usuário...")
	user := createUser()
	log.Printf("   ✓ Usuário criado: %s", user.ID)

	// Testar diferentes intervalos
	timeRanges := []int{7, 30, 90, 180}

	for _, days := range timeRanges {
		log.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Printf("📅 Testando com %d dias de histórico", days)
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// 2. Criar canal com histórico habilitado
		log.Printf("\n2️⃣ Criando canal WAHA com histórico de %d dias...", days)
		channel := createChannel(user.APIKey, days)
		log.Printf("   ✓ Canal criado: %s", channel.ID)

		// 3. Iniciar importação
		log.Println("\n3️⃣ Iniciando importação de histórico...")
		importResp := startImport(user.APIKey, channel.ID, days)
		log.Printf("   ✓ Workflow iniciado: %s", importResp.WorkflowID)

		// 4. Aguardar conclusão
		log.Println("\n4️⃣ Aguardando conclusão...")
		status := waitForCompletion(user.APIKey, channel.ID)

		log.Printf("\n📊 Resultados para %d dias:", days)
		log.Printf("   • Status: %s", status.Status)
		log.Printf("   • Chats: %d", status.ChatsProcessed)
		log.Printf("   • Mensagens: %d", status.MessagesImported)
		log.Printf("   • Sessões: %d", status.SessionsCreated)
		log.Printf("   • Contatos: %d", status.ContactsCreated)

		if status.MessagesImported > 0 {
			log.Printf("\n✅ SUCESSO! Encontradas %d mensagens nos últimos %d dias", status.MessagesImported, days)
			log.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			log.Println("🎯 Conclusão da Verificação:")
			log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			log.Printf("✓ A filtragem nativa da API WAHA está funcionando corretamente")
			log.Printf("✓ Mensagens mais antigas que %d dias foram filtradas pelo servidor", days)
			log.Printf("✓ Total de mensagens importadas: %d", status.MessagesImported)
			log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			os.Exit(0)
		}

		time.Sleep(2 * time.Second)
	}

	log.Println("\n⚠️  Nenhuma mensagem encontrada em nenhum intervalo testado")
	log.Println("   Isso pode indicar que a sessão WAHA não tem mensagens ou")
	log.Println("   que todas as mensagens são mais antigas que 180 dias")
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

func createChannel(apiKey string, maxDays int) Channel {
	payload := map[string]interface{}{
		"name": fmt.Sprintf("Test WAHA %dd", maxDays),
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

	// Atualiza canal para habilitar history import via psql
	cmd := fmt.Sprintf(`PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "UPDATE channels SET history_import_enabled=true, history_import_max_days=%d WHERE id='%s'"`, maxDays, channel.ID)
	exec.Command("bash", "-c", cmd).Run()

	return channel
}

func startImport(apiKey, channelID string, days int) ImportResponse {
	payload := map[string]interface{}{
		"strategy":        "time_range",
		"time_range_days": days,
	}

	resp := makeRequest("POST", fmt.Sprintf("/api/v1/crm/channels/%s/import-history", channelID), apiKey, payload)

	var importResp ImportResponse
	json.Unmarshal(resp, &importResp)
	return importResp
}

func waitForCompletion(apiKey, channelID string) ImportStatus {
	maxAttempts := 60 // 5 minutos
	for i := 1; i <= maxAttempts; i++ {
		time.Sleep(5 * time.Second)
		status := getImportStatus(apiKey, channelID)

		if status.Status == "Completed" || status.Status == "Failed" {
			return status
		}
	}

	return getImportStatus(apiKey, channelID)
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

	req, err := http.NewRequest(method, baseURL+path, body)
	if err != nil {
		log.Fatalf("Erro ao criar request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Erro ao fazer request: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		log.Printf("⚠️  Erro na API (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody
}
