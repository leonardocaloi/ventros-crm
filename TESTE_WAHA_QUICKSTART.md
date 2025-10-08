# 🚀 Teste WAHA - Guia Rápido

## ⚡ TL;DR

```bash
# Terminal 1: Infraestrutura
make infra

# Terminal 2: API
make api

# Terminal 3: Testes
make test-waha
```

## 📋 O que o teste faz?

1. ✅ **Cria usuário automaticamente**
   - Email: `test-waha-{timestamp}@example.com`
   - Gera API key
   - Cria projeto padrão

2. ✅ **Cria canal WAHA automaticamente**
   - Nome: `Test WAHA Channel {timestamp}`
   - Session ID: `test-session-{timestamp}`
   - Webhook URL: `http://localhost:8080/api/v1/webhooks/waha/{session}`

3. ✅ **Testa 8 tipos de mensagem**
   - TEXT, IMAGE, VOICE (PTT), LOCATION, CONTACT, DOCUMENT, AUDIO, IMAGE+TEXT

4. ✅ **Valida processamento**
   - Webhook retorna 200
   - Mensagem é enfileirada
   - Contact é criado
   - Session é criada
   - Estatísticas atualizadas

5. ✅ **Limpa tudo automaticamente**

## 🎯 Resultado Esperado

```
✅ API is ready
🚀 Setting up WAHA Webhook E2E Test
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1️⃣ User created: test-waha-1696598400@example.com
2️⃣ Channel created: Test WAHA Channel 1696598400
   • Webhook URL: http://localhost:8080/api/v1/webhooks/waha/test-session-1696598400
3️⃣ Channel activated
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Setup completo!

📝 Testing TEXT message...
✅ TEXT message processed

🖼️  Testing IMAGE message...
✅ IMAGE message processed

🎤 Testing VOICE (PTT) message...
✅ VOICE message processed

📍 Testing LOCATION message...
✅ LOCATION message processed

👤 Testing CONTACT message...
✅ CONTACT message processed

📄 Testing DOCUMENT message...
✅ DOCUMENT message processed

🔊 Testing AUDIO message...
✅ AUDIO message processed

🖼️📝 Testing IMAGE with TEXT message...
✅ IMAGE with TEXT message processed

PASS
ok      github.com/caloi/ventros-crm/tests/e2e  25.123s
```

## 🔍 O que está sendo testado?

### Fluxo Completo

```
Teste → Webhook URL Dinâmico → RabbitMQ → Consumer → Adapter → Service → Use Case → Repository → DB
```

### Validações

- ✅ **Webhook aceita evento** (200 OK)
- ✅ **Evento é enfileirado** (RabbitMQ)
- ✅ **Adapter detecta tipo correto** (text/image/voice/etc)
- ✅ **Contact é criado/atualizado**
- ✅ **Session é criada/estendida**
- ✅ **Message é salva no DB**
- ✅ **Estatísticas do canal atualizadas**

### Tipos de Mensagem

| Tipo | Arquivo | content_type | Campos Testados |
|------|---------|--------------|-----------------|
| Texto | `message_text.json` | `text` | `text` |
| Imagem | `message_image.json` | `image` | `media_url`, `media_mimetype` |
| Voz (PTT) | `message_recorded_audio.json` | `voice` | `media_url`, `media_mimetype` |
| Localização | `message_location.json` | `location` | `metadata.location` |
| Contato | `message_contact.json` | `contact` | `metadata.contact` |
| Documento | `message_document_pdf.json` | `document` | `media_url`, `metadata.filename` |
| Áudio | `message_audio.json` | `audio` | `media_url`, `media_mimetype` |
| Imagem+Texto | `message_image_text.json` | `image` | `media_url`, `text` (caption) |

## 🐛 Troubleshooting

### ❌ "API não está rodando"

```bash
# Verifique se a API está up
curl http://localhost:8080/health

# Se não estiver, rode:
make api
```

### ❌ "Failed to load event file"

```bash
# Rode do diretório raiz
cd /home/caloi/ventros-crm
make test-waha
```

### ❌ Teste falha em um tipo específico

```bash
# Rode apenas esse teste
go test -v -run TestWAHAWebhookTestSuite/TestVoiceMessage ./tests/e2e/

# Veja logs da API
# (no terminal onde rodou make api)
```

### ❌ Timeout

```bash
# Aumente o timeout
go test -v -timeout 20m -run TestWAHAWebhookTestSuite ./tests/e2e/
```

## 📊 Verificar Resultados no Banco

```sql
-- Ver canal criado
SELECT id, name, type, webhook_url, messages_received, status
FROM channels
WHERE name LIKE 'Test WAHA Channel%'
ORDER BY created_at DESC
LIMIT 1;

-- Ver mensagens processadas
SELECT 
  id, 
  content_type, 
  text, 
  media_url, 
  metadata,
  created_at
FROM messages
WHERE channel_id = 'ID_DO_CANAL_ACIMA'
ORDER BY created_at DESC;

-- Ver contato criado
SELECT id, name, phone, created_at
FROM contacts
WHERE phone = '554497044474'
ORDER BY created_at DESC
LIMIT 1;

-- Ver sessão criada
SELECT id, contact_id, status, messages_count, created_at
FROM sessions
WHERE contact_id = 'ID_DO_CONTATO_ACIMA'
ORDER BY created_at DESC
LIMIT 1;
```

## 🎯 Casos de Uso

### 1. Validar correções de bugs

Após corrigir PTT, Location, Contact:

```bash
make test-waha
```

Todos os testes devem passar ✅

### 2. Testar novo tipo de mensagem

1. Adicione JSON em `/events_waha/message_novo_tipo.json`
2. Adicione teste em `waha_webhook_test.go`:

```go
func (s *WAHAWebhookTestSuite) TestNovoTipo() {
    fmt.Println("\n🆕 Testing NOVO TIPO message...")
    event := s.loadEventFile("message_novo_tipo.json")
    s.sendWebhookEvent(event)
    time.Sleep(2 * time.Second)
    s.verifyChannelStats(9)
    fmt.Println("✅ NOVO TIPO message processed")
}
```

3. Rode:

```bash
make test-waha
```

### 3. CI/CD

```yaml
# .github/workflows/test.yml
name: E2E Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Start Infrastructure
        run: make infra
      
      - name: Start API
        run: |
          make api &
          sleep 15
      
      - name: Run WAHA Tests
        run: make test-waha
```

### 4. Teste manual do webhook

Use a URL gerada pelo teste:

```bash
# 1. Rode o teste para criar canal
make test-waha

# 2. Pegue a webhook URL da saída:
# http://localhost:8080/api/v1/webhooks/waha/test-session-1696598400

# 3. Envie evento manualmente
curl -X POST http://localhost:8080/api/v1/webhooks/waha/test-session-1696598400 \
  -H "Content-Type: application/json" \
  -d @events_waha/message_text.json
```

## 📚 Documentação Completa

- **README Detalhado:** `/tests/e2e/README.md`
- **Análise do Fluxo:** `/ANALISE_FLUXO_MENSAGENS.md`
- **Correções Implementadas:** `/CORRECOES_IMPLEMENTADAS.md`
- **Mapeamento de Campos:** `/MAPEAMENTO_CAMPOS_MENSAGENS.md`

## ✅ Checklist

Antes de rodar:
- [ ] Infraestrutura rodando (`make infra`)
- [ ] API rodando (`make api`)
- [ ] Diretório correto (`/home/caloi/ventros-crm`)

Durante:
- [ ] Todos os 8 testes passam
- [ ] Webhook URL é gerado
- [ ] Estatísticas são atualizadas

Depois:
- [ ] Cleanup automático executado
- [ ] Nenhum erro nos logs da API
- [ ] Banco de dados tem os registros

## 🎉 Pronto!

Agora você tem um teste E2E completo que:
- ✅ Cria todo o ambiente automaticamente
- ✅ Testa todos os tipos de mensagem
- ✅ Valida o fluxo completo
- ✅ Limpa tudo no final
- ✅ Pode ser usado em CI/CD
- ✅ Webhook URL dinâmico para cada execução
