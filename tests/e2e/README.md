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

### Rodar Apenas Testes de Scheduled Automation

```bash
go test -v -timeout 5m -run TestScheduledAutomationTestSuite ./tests/e2e/
```

## 📁 Estrutura de Arquivos

```
tests/e2e/
├── README.md                          # Esta documentação
├── testdata/                          # Dados de teste (fixtures)
│   └── events_waha/                   # Eventos WAHA reais (WhatsApp)
│       ├── message_text.json
│       ├── message_image.json
│       └── ... (14 tipos de mensagens)
├── testdata_helpers.go                # Helpers para carregar fixtures
├── api_test.go                        # Setup geral e helpers HTTP
├── waha_webhook_test.go               # Testes de webhook WAHA
├── scheduled_automation_test.go       # Testes de automações agendadas ⭐ NOVO
└── message_send_test.go               # Testes de envio de mensagens
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

## 🤖 Teste Scheduled Automation (`scheduled_automation_test.go`) ⭐ NOVO

### O que o teste faz

Testa o sistema completo de automações agendadas (Scheduled Automation Worker):

1. **Setup Automático:**
   - ✅ Cria usuário de teste
   - ✅ Cria canal WAHA
   - ✅ Cria pipeline
   - ✅ Cria contato de teste
   - ✅ Conecta ao banco de dados para verificações

2. **Testa 3 Tipos de Agendamento:**
   - ✅ **DAILY** - Automação que executa diariamente em horário específico
   - ✅ **WEEKLY** - Automação que executa semanalmente em dia/horário específico
   - ✅ **ONCE** - Automação que executa apenas uma vez

3. **Validações Completas:**
   - ✅ Verifica que worker processa a automation (last_executed atualizado)
   - ✅ Verifica que next_execution é calculado corretamente
   - ✅ Para DAILY: next_execution é ~24h no futuro
   - ✅ Para WEEKLY: next_execution é ~7 dias no futuro
   - ✅ Para ONCE: next_execution é NULL (executa apenas uma vez)

4. **Cleanup Automático:**
   - ✅ Remove automation rules criadas
   - ✅ Remove contatos criados
   - ✅ Remove canais criados

### Estrutura do Teste

```go
type ScheduledAutomationTestSuite struct {
    suite.Suite
    baseURL    string   // http://localhost:8080
    client     *http.Client
    db         *sql.DB  // Conexão direta ao PostgreSQL
    userID     string
    projectID  string
    apiKey     string
    channelID  string
    pipelineID string
    contactID  string
    ruleID     string   // ID da automation rule criada
}
```

### Tempo de Execução

⏱️ **~3-4 minutos** total (70 segundos por teste x 3 testes)

**Por que tão longo?**
- Worker faz polling a cada **1 minuto**
- Cada teste aguarda 70 segundos para garantir que o worker processou
- Sem esse delay, o teste falharia pois a automation não seria executada

### Exemplo de Saída

```
🤖 Setting up Scheduled Automation E2E Test
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1️⃣ User created: test-automation-1696598400@example.com
   • User ID: 550e8400-e29b-41d4-a716-446655440000
   • Project ID: 550e8400-e29b-41d4-a716-446655440001
2️⃣ Channel created: Test Automation Channel 1696598400
   • Channel ID: 550e8400-e29b-41d4-a716-446655440002
3️⃣ Using existing pipeline: 550e8400-e29b-41d4-a716-446655440003
4️⃣ Contact created: 550e8400-e29b-41d4-a716-446655440004
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Setup completo!

📅 Testing DAILY scheduled automation...
   • Rule ID: 550e8400-e29b-41d4-a716-446655440005
   • Waiting 70 seconds for worker to process (polls every 1 minute)...
   ✅ Rule executed at: 2025-10-11T20:35:00Z
   ✅ Next execution: 2025-10-12T20:35:00Z
✅ DAILY scheduled automation processed successfully

📆 Testing WEEKLY scheduled automation...
   • Rule ID: 550e8400-e29b-41d4-a716-446655440006 (weekday: wednesday)
   • Waiting 70 seconds for worker...
   ✅ Weekly rule executed at: 2025-10-11T20:36:10Z
✅ WEEKLY scheduled automation processed successfully

⏰ Testing ONCE scheduled automation...
   • Rule ID: 550e8400-e29b-41d4-a716-446655440007
   • Scheduled for: 2025-10-11T20:37:02Z
   • Waiting 70 seconds for worker...
   ✅ Once rule executed at: 2025-10-11T20:37:10Z
   ✅ Next execution is NULL (as expected for ONCE type)
✅ ONCE scheduled automation processed successfully

🧹 Cleaning up test data...
  ✓ Deleted automation rule: 550e8400-e29b-41d4-a716-446655440005
  ✓ Deleted contact: 550e8400-e29b-41d4-a716-446655440004
  ✓ Deleted channel: 550e8400-e29b-41d4-a716-446655440002
✅ Cleanup completed
```

### Como Funciona Internamente

1. **Teste cria automation rule** com `next_execution` no passado (para executar imediatamente)
2. **Worker acorda** (polling de 1 minuto)
3. **Worker busca rules** onde `next_execution <= NOW() AND enabled = true`
4. **Worker executa** via `AutomationEngine.ProcessScheduledTrigger()`
5. **Logging Executor** loga a ação (MVP mode - não executa ações reais)
6. **Worker atualiza** `last_executed` e calcula novo `next_execution`
7. **Teste verifica** que timestamps foram atualizados corretamente

### Schedule JSON Examples

**DAILY (diária):**
```json
{
  "type": "daily",
  "hour": 14,
  "minute": 30
}
```

**WEEKLY (semanal):**
```json
{
  "type": "weekly",
  "weekday": "monday",
  "hour": 9,
  "minute": 0
}
```

**ONCE (única):**
```json
{
  "type": "once",
  "execute_at": "2025-10-11T20:00:00Z"
}
```

### Verificar Logs do Worker

Durante o teste, você pode ver logs do worker na API:

```bash
# Em outro terminal
docker logs -f ventros-crm-api | grep -i "scheduled automation"

# Ou se rodando localmente
tail -f /tmp/api.log | grep -i "scheduled automation"
```

Você verá:
```
INFO  ✅ Scheduled Automation Worker started (polling every 1 minute, MVP mode)
INFO  starting scheduled rules worker
INFO  found 3 scheduled rules ready to execute
INFO  📋 Scheduled automation action (MVP - logging only)
      action_type=send_message rule_id=... tenant_id=...
```

## 🔔 Teste Scheduled Automation + Webhook (`scheduled_automation_webhook_test.go`) ⭐ NOVO

### O que o teste faz

Testa o fluxo **completo** de automação agendada COM notificação via webhook:

1. **Inicia servidor HTTP de teste** para receber webhooks
2. **Cria webhook subscription** para eventos `automation.executed` e `automation.failed`
3. **Cria automation rule** agendada
4. **Aguarda worker** processar (70 segundos)
5. **Verifica execução** no banco de dados
6. **Verifica webhook recebido** no servidor de teste

### Como Funciona

```go
// 1. Inicia servidor HTTP que captura webhooks
s.webhookServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    var payload map[string]interface{}
    json.NewDecoder(r.Body).Decode(&payload)

    s.webhookReceived = append(s.webhookReceived, payload)

    fmt.Printf("📨 Webhook received: event=%s\n", payload["event"])
    w.WriteHeader(http.StatusOK)
}))

// 2. Cria subscription apontando para o servidor de teste
POST /api/v1/webhooks/subscriptions
{
    "url": "http://127.0.0.1:xxxxx",  // URL do servidor de teste
    "events": ["automation.executed", "automation.failed"],
    "active": true
}

// 3. Worker executa automation

// 4. API envia webhook para servidor de teste

// 5. Teste verifica que webhook foi recebido
assert.GreaterOrEqual(t, len(s.webhookReceived), 1)
```

### Exemplo de Saída

```
🔔 Setting up Scheduled Automation + Webhook E2E Test
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔔 Webhook server started: http://127.0.0.1:54321
✅ API is ready
1️⃣ User created: test-webhook-1696598400@example.com
   • User ID: uuid...
   • Project ID: uuid...
2️⃣ Channel created: Test Webhook Channel 1696598400
   • Channel ID: uuid...
3️⃣ Using existing pipeline: uuid...
4️⃣ Contact created: uuid...
5️⃣ Webhook subscription created: uuid...
   • Events: automation.executed, automation.failed
   • URL: http://127.0.0.1:54321
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Setup completo!
📍 Webhook Server: http://127.0.0.1:54321

🔔 Testing scheduled automation WITH webhook notification...
   • Rule ID: uuid...
   • Waiting 75 seconds for worker to process...
   📨 Webhook received: event=automation.executed
   ✅ Rule executed at: 2025-10-11T20:45:00Z
   ✅ Webhook received: 1 total
   ✅ Event type: automation.executed
   ✅ Payload contains: rule_id=uuid...
✅ Scheduled automation WITH webhook processed successfully

🧹 Cleaning up test data...
  ✓ Deleted webhook subscription: uuid...
  ✓ Deleted automation rule: uuid...
  ✓ Deleted contact: uuid...
  ✓ Deleted channel: uuid...
  ✓ Stopped webhook server
✅ Cleanup completed
```

### Webhook Payload Exemplo

```json
{
  "event": "automation.executed",
  "timestamp": "2025-10-11T20:45:00Z",
  "payload": {
    "rule_id": "550e8400-e29b-41d4-a716-446655440005",
    "rule_name": "Test webhook-test Automation",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440001",
    "pipeline_id": "550e8400-e29b-41d4-a716-446655440003",
    "trigger": "scheduled",
    "actions_executed": 1,
    "execution_time_ms": 42
  }
}
```

### Rodar o Teste

```bash
go test -v -timeout 5m -run TestScheduledAutomationWebhookTestSuite ./tests/e2e/
```

## 📁 Arquivos de Teste

### Eventos WAHA (`testdata/events_waha/`)

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
- ✅ Criação de canal via `/api/v1/crm/channels`
- ✅ Ativação de canal via `/api/v1/crm/channels/:id/activate`
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

Certifique-se que está rodando do diretório raiz do projeto:

```bash
cd $(git rev-parse --show-toplevel)
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
