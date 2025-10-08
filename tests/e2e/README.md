# Testes E2E - Ventros CRM

## 📋 Visão Geral

Testes end-to-end que validam o fluxo completo da aplicação, desde criação de usuário até processamento de webhooks WAHA.

## 🚀 Como Rodar

### Pré-requisitos

1. **Infraestrutura rodando:**
   ```bash
   make infra
   ```

2. **API rodando (em outro terminal):**
   ```bash
   make api
   ```

### Rodar Todos os Testes E2E

```bash
make test-e2e
```

### Rodar Apenas Testes WAHA

```bash
make test-waha
```

## 🧪 Teste WAHA Webhook (`waha_webhook_test.go`)

### O que o teste faz

1. **Setup Automático:**
   - ✅ Cria usuário de teste
   - ✅ Gera API key automaticamente
   - ✅ Cria projeto padrão
   - ✅ Cria canal WAHA
   - ✅ Ativa canal
   - ✅ Gera webhook URL dinâmico

2. **Testa Todos os Tipos de Mensagem:**
   - ✅ TEXT - Mensagem de texto simples
   - ✅ IMAGE - Imagem
   - ✅ VOICE (PTT) - Áudio gravado (push-to-talk)
   - ✅ LOCATION - Localização com coordenadas
   - ✅ CONTACT - Contato (vCard)
   - ✅ DOCUMENT - Documento (PDF, HEIC, etc)
   - ✅ AUDIO - Áudio normal
   - ✅ IMAGE + TEXT - Imagem com legenda

3. **Validações:**
   - ✅ Webhook retorna 200 OK
   - ✅ Evento é enfileirado
   - ✅ Mensagem é processada
   - ✅ Estatísticas do canal são atualizadas
   - ✅ Contact é criado/atualizado
   - ✅ Session é criada/estendida

4. **Cleanup Automático:**
   - ✅ Remove canal criado
   - ✅ Limpa dados de teste

### Estrutura do Teste

```go
type WAHAWebhookTestSuite struct {
    baseURL    string  // http://localhost:8080
    client     *http.Client
    userID     string  // Criado automaticamente
    projectID  string  // Criado automaticamente
    apiKey     string  // Gerado automaticamente
    channelID  string  // Criado automaticamente
    webhookURL string  // Gerado automaticamente
}
```

### Webhook URL Dinâmico

O teste gera automaticamente uma URL de webhook única para cada execução:

```
http://localhost:8080/api/v1/webhooks/waha/{session_id}
```

Onde `{session_id}` é gerado dinamicamente baseado no canal criado.

### Exemplo de Saída

```
🚀 Setting up WAHA Webhook E2E Test
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1️⃣ User created: test-waha-1696598400@example.com
   • User ID: 550e8400-e29b-41d4-a716-446655440000
   • Project ID: 550e8400-e29b-41d4-a716-446655440001
   • API Key: sk_test_abc123...
2️⃣ Channel created: Test WAHA Channel 1696598400
   • Channel ID: 550e8400-e29b-41d4-a716-446655440002
   • Webhook URL: http://localhost:8080/api/v1/webhooks/waha/test-session-1696598400
3️⃣ Channel activated: 550e8400-e29b-41d4-a716-446655440002
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Setup completo!
📍 Webhook URL: http://localhost:8080/api/v1/webhooks/waha/test-session-1696598400

📝 Testing TEXT message...
   📊 Channel stats: 1 messages received
✅ TEXT message processed

🖼️  Testing IMAGE message...
   📊 Channel stats: 2 messages received
✅ IMAGE message processed

🎤 Testing VOICE (PTT) message...
   📊 Channel stats: 3 messages received
✅ VOICE message processed

📍 Testing LOCATION message...
   📊 Channel stats: 4 messages received
✅ LOCATION message processed

👤 Testing CONTACT message...
   📊 Channel stats: 5 messages received
✅ CONTACT message processed

📄 Testing DOCUMENT message...
   📊 Channel stats: 6 messages received
✅ DOCUMENT message processed

🔊 Testing AUDIO message...
   📊 Channel stats: 7 messages received
✅ AUDIO message processed

🖼️📝 Testing IMAGE with TEXT message...
   📊 Channel stats: 8 messages received
✅ IMAGE with TEXT message processed

🧹 Cleaning up test data...
  ✓ Deleted channel: 550e8400-e29b-41d4-a716-446655440002
✅ Cleanup completed
```

## 📁 Arquivos de Teste

### Eventos WAHA (`/events_waha/`)

Os testes usam eventos reais do WAHA:

- `message_text.json` - Mensagem de texto
- `message_image.json` - Imagem
- `message_image_text.json` - Imagem com legenda
- `message_audio.json` - Áudio normal
- `message_recorded_audio.json` - Áudio PTT (voz)
- `message_location.json` - Localização
- `message_contact.json` - Contato (vCard)
- `message_document_pdf.json` - Documento PDF
- `message_document_image_heic.json` - Documento HEIC

### Fixtures (`fixtures.go`)

Dados de teste reutilizáveis:

```go
type TestFixtures struct {
    Users    []UserFixture
    Channels []ChannelFixture
    Contacts []ContactFixture
}
```

## 🔧 Configuração

### Variáveis de Ambiente

```bash
# URL base da API (padrão: http://localhost:8080)
export API_BASE_URL=http://localhost:8080

# Timeout dos testes (padrão: 10m)
export TEST_TIMEOUT=10m
```

### Rodar Contra API Remota

```bash
API_BASE_URL=https://api.ventros.com make test-waha
```

## 🐛 Debug

### Ver logs detalhados

```bash
go test -v -timeout 10m -run TestWAHAWebhookTestSuite ./tests/e2e/
```

### Rodar teste específico

```bash
go test -v -run TestWAHAWebhookTestSuite/TestTextMessage ./tests/e2e/
```

### Ver requisições HTTP

Adicione logging no teste:

```go
resp, body := s.makeRequest("POST", endpoint, payload, apiKey)
fmt.Printf("Request: %s %s\n", "POST", endpoint)
fmt.Printf("Response: %d - %s\n", resp.StatusCode, string(body))
```

## 📊 Cobertura

### O que é testado

- ✅ Criação de usuário via `/api/v1/auth/register`
- ✅ Criação de canal via `/api/v1/channels`
- ✅ Ativação de canal via `/api/v1/channels/:id/activate`
- ✅ Webhook WAHA via `/api/v1/webhooks/waha/{session}`
- ✅ Processamento de 8 tipos de mensagem
- ✅ Estatísticas do canal
- ✅ Criação automática de Contact
- ✅ Criação automática de Session
- ✅ Enfileiramento via RabbitMQ
- ✅ Processamento assíncrono

### O que NÃO é testado

- ❌ Envio de mensagens (outbound)
- ❌ Temporal workflows (session timeout)
- ❌ Webhooks de eventos de domínio
- ❌ Integração com WAHA real

## 🚨 Troubleshooting

### Erro: "API não está rodando"

```bash
# Terminal 1
make infra

# Terminal 2
make api

# Terminal 3
make test-waha
```

### Erro: "Failed to load event file"

Certifique-se que está rodando do diretório raiz:

```bash
cd /home/caloi/ventros-crm
make test-waha
```

### Erro: "Channel not found"

O canal pode ter sido deletado. Rode novamente:

```bash
make test-waha
```

### Timeout nos testes

Aumente o timeout:

```bash
go test -v -timeout 20m -run TestWAHAWebhookTestSuite ./tests/e2e/
```

## 📝 Adicionar Novos Testes

### 1. Adicionar novo tipo de mensagem

```go
func (s *WAHAWebhookTestSuite) TestVideoMessage() {
    fmt.Println("\n🎥 Testing VIDEO message...")
    
    event := s.loadEventFile("message_video.json")
    s.sendWebhookEvent(event)
    
    time.Sleep(2 * time.Second)
    s.verifyChannelStats(9) // Incrementa contador
    
    fmt.Println("✅ VIDEO message processed")
}
```

### 2. Adicionar novo arquivo de evento

Coloque o JSON em `/events_waha/message_video.json`

### 3. Rodar teste

```bash
make test-waha
```

## 🎯 Casos de Uso

### CI/CD Pipeline

```yaml
# .github/workflows/test.yml
- name: Run E2E Tests
  run: |
    make infra
    make api &
    sleep 15
    make test-waha
```

### Teste Local Rápido

```bash
# Setup (uma vez)
make infra
make api

# Rodar testes (quantas vezes quiser)
make test-waha
```

### Validar Correções

Após corrigir bugs (PTT, Location, Contact):

```bash
make test-waha
```

Todos os testes devem passar ✅

## 📚 Referências

- **Documentação WAHA:** https://waha.devlike.pro/
- **Eventos WAHA:** `/events_waha/*.json`
- **Análise do Fluxo:** `/ANALISE_FLUXO_MENSAGENS.md`
- **Correções:** `/CORRECOES_IMPLEMENTADAS.md`
- **Mapeamento:** `/MAPEAMENTO_CAMPOS_MENSAGENS.md`
