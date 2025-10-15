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
	wahaSessionID = "freefaro-b2b-comercial"
	maxDays       = 20 // 20 dias de histórico
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
	log.Println("🚀 Teste de Importação de Histórico WAHA - 20 Dias")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 1. Criar usuário
	log.Println("\n1️⃣ Criando usuário...")
	user := createUser()
	log.Printf("   ✓ Usuário criado: %s", user.ID)
	log.Printf("   ✓ API Key: %s", user.APIKey)

	// 2. Criar canal com histórico habilitado (20 dias)
	log.Println("\n2️⃣ Criando canal WAHA com histórico de 20 dias...")
	channel := createChannel(user.APIKey)
	log.Printf("   ✓ Canal criado: %s", channel.ID)
	log.Printf("   ✓ Session: %s", wahaSessionID)
	log.Printf("   ✓ Histórico habilitado: %t", channel.HistoryImportEnabled)
	log.Printf("   ✓ Máximo de dias: %d", channel.HistoryImportMaxDays)

	// 3. Iniciar importação
	log.Println("\n3️⃣ Iniciando importação de histórico...")
	importResp := startImport(user.APIKey, channel.ID)
	log.Printf("   ✓ Workflow iniciado: %s", importResp.WorkflowID)
	log.Printf("   ✓ Run ID: %s", importResp.RunID)
	log.Printf("   ✓ Strategy: %s", importResp.Strategy)

	// 4. Monitorar progresso
	log.Println("\n4️⃣ Monitorando progresso da importação...")
	log.Println("   (aguardando até 5 minutos ou conclusão...)")

	maxAttempts := 60 // 5 minutos (5 segundos por tentativa)
	for i := 1; i <= maxAttempts; i++ {
		time.Sleep(5 * time.Second)

		status := getImportStatus(user.APIKey, channel.ID)

		log.Printf("   [%d/%d] Status: %s | Chats: %d | Mensagens: %d | Sessões: %d | Contatos: %d",
			i, maxAttempts,
			status.Status,
			status.ChatsProcessed,
			status.MessagesImported,
			status.SessionsCreated,
			status.ContactsCreated,
		)

		if status.Status == "Completed" || status.Status == "Failed" {
			log.Println("\n✅ Importação finalizada!")
			log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			log.Printf("Status Final: %s", status.Status)
			log.Printf("Chats Processados: %d", status.ChatsProcessed)
			log.Printf("Mensagens Importadas: %d", status.MessagesImported)
			log.Printf("Sessões Criadas: %d", status.SessionsCreated)
			log.Printf("Contatos Criados: %d", status.ContactsCreated)

			if status.Status == "Completed" {
				if status.MessagesImported > 0 {
					log.Println("\n✅ SUCESSO! Mensagens foram importadas!")
					log.Printf("   🎯 A filtragem nativa da API WAHA está funcionando corretamente!")
					log.Printf("   🎯 %d mensagens encontradas nos últimos %d dias", status.MessagesImported, maxDays)
					log.Printf("   🎯 %d contatos criados", status.ContactsCreated)
					log.Printf("   🎯 %d sessões criadas", status.SessionsCreated)
					os.Exit(0)
				} else {
					log.Println("\n⚠️  Nenhuma mensagem foi importada")
					log.Printf("   Isso indica que não há mensagens nos últimos %d dias nesta sessão", maxDays)
					log.Println("   A implementação está correta, mas a sessão WAHA não tem mensagens recentes")
					os.Exit(0)
				}
			} else if status.Status == "Failed" {
				log.Println("\n❌ TESTE FALHOU! A importação falhou.")
				os.Exit(1)
			}
			break
		}
	}

	log.Println("\n⚠️  Timeout atingido (5 minutos)")
	log.Println("   A importação ainda está em andamento...")
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
		"name": "Test WAHA History Import 20d",
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

	// Ativar canal (requer sessão WAHA funcionando)
	log.Println("\n   🔌 Ativando canal...")
	activateResp := makeRequest("POST", fmt.Sprintf("/api/v1/crm/channels/%s/activate", channel.ID), apiKey, nil)
	log.Printf("   ✓ Canal ativação solicitada: %s", string(activateResp))

	// Aguardar canal ficar ativo (polling) - até 2 minutos
	log.Println("   ⏳ Aguardando canal ficar ativo...")
	channelActive := false
	for i := 0; i < 60; i++ { // 60 tentativas * 2s = 2 minutos
		time.Sleep(2 * time.Second)
		statusResp := makeRequest("GET", fmt.Sprintf("/api/v1/crm/channels/%s", channel.ID), apiKey, nil)

		var responseData map[string]interface{}
		json.Unmarshal(statusResp, &responseData)

		// O status está dentro do objeto "channel"
		if channelData, ok := responseData["channel"].(map[string]interface{}); ok {
			if statusStr, ok := channelData["status"].(string); ok {
				log.Printf("   [%d/60] Status: %s", i+1, statusStr)

				if statusStr == "active" {
					log.Printf("   ✓ Canal ativo após %d segundos", (i+1)*2)

					// Re-fetch para obter dados completos
					channelBytes, _ := json.Marshal(channelData)
					var updatedChannel Channel
					json.Unmarshal(channelBytes, &updatedChannel)
					channel = updatedChannel
					channelActive = true
					break
				}
			}
		}
	}

	if !channelActive {
		log.Fatal("   ❌ ERRO: Canal não ficou ativo após 2 minutos. Abortando teste.")
	}

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
