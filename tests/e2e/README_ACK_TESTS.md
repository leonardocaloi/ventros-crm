# Testes E2E de ACKs

Scripts para testar o fluxo completo de ACKs (acknowledgments) do WhatsApp localmente.

## 📋 Pré-requisitos

- API rodando localmente (`make api`)
- PostgreSQL rodando
- Variáveis de ambiente configuradas no `.env`

```bash
# Variáveis necessárias no .env
WAHA_DEFAULT_SESSION_ID_TEST=test-session
TEST_PHONE_NUMBER=5544970444747
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ventros_crm
DB_USER=postgres
DB_PASSWORD=postgres
```

## 🧪 Scripts Disponíveis

### 1. `test_ack_simulation.sh` - Teste Completo Automatizado

Testa o fluxo completo de ACKs de forma automatizada:
1. Cria usuário, projeto, canal e contato
2. Envia uma mensagem de vídeo
3. Simula webhooks de ACK progressivos (1 → 2 → 3 → 4)
4. Verifica se os status foram atualizados corretamente

**Uso:**

```bash
./tests/e2e/test_ack_simulation.sh
```

**Saída esperada:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ACK Webhook Simulation Test
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

→ Setting up test environment...
✓ User registered
✓ Channel created
✓ Channel activated
✓ Contact created

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Step 1: Send Media Message
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ Message sent
  Message ID: 8c0d012d-9fd5-4c0b-8d9e-09ce7a0c7fc7

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Step 2: Simulate ACK 1 (SERVER)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ Webhook accepted (queued for processing)
✓ Status correctly updated to 'sent'

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Step 3: Simulate ACK 2 (DEVICE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ Webhook accepted (queued for processing)
✓ Status correctly updated to 'delivered'
  ✓✓ Delivered at: 2025-01-13T10:15:30Z

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Step 4: Simulate ACK 3 (READ)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ Webhook accepted (queued for processing)
✓ Status correctly updated to 'read'
  ✓✓ (blue) Read at: 2025-01-13T10:15:35Z

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Step 5: Simulate ACK 4 (PLAYED)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ Webhook accepted (queued for processing)
✓ Status correctly updated to 'played'
  ▶️ Played at: 2025-01-13T10:15:40Z

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ✓ ACK Simulation Test Complete!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 2. `send_ack_webhook.sh` - Enviar ACK para Mensagem Específica

Envia um webhook de ACK específico para uma mensagem já existente.

**Uso:**

```bash
./tests/e2e/send_ack_webhook.sh <channel_message_id> <ack_number> [webhook_id]
```

**Argumentos:**
- `channel_message_id` - O WAMID da mensagem (ex: `wamid.HBgNNTU0NDk3MDQ0NDc0FQIAERgS...`)
- `ack_number` - Número do ACK (veja tabela abaixo)
- `webhook_id` - (Opcional) ID do webhook do canal. Se não fornecido, busca automaticamente do banco

**Números de ACK:**
```
-1 = ERROR   → failed
 0 = PENDING → queued
 1 = SERVER  → sent
 2 = DEVICE  → delivered (✓✓)
 3 = READ    → read (✓✓ azul)
 4 = PLAYED  → played (▶️ mídia visualizada)
```

**Exemplos:**

```bash
# Enviar ACK 2 (DELIVERED) para uma mensagem específica
./tests/e2e/send_ack_webhook.sh 'wamid.HBgNNTU0NDk3MDQ0NDc0FQIAERgS...' 2

# Enviar ACK 4 (PLAYED) para uma mensagem específica com webhook_id explícito
./tests/e2e/send_ack_webhook.sh 'wamid.HBgN...' 4 '550e8400-e29b-41d4-a716-446655440000'

# Enviar ACK -1 (ERROR) para simular falha
./tests/e2e/send_ack_webhook.sh 'wamid.HBgN...' -1
```

**Como obter o channel_message_id:**

Opção 1 - Via API:
```bash
curl -X GET "http://localhost:8080/api/v1/crm/messages/$MESSAGE_ID" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.channel_message_id'
```

Opção 2 - Via Banco:
```sql
SELECT id, channel_message_id, status
FROM messages
WHERE id = 'message-uuid'
OR contact_id = 'contact-uuid'
ORDER BY created_at DESC LIMIT 5;
```

## 🔍 Verificando os Resultados

### No Banco de Dados

```sql
-- Ver status de uma mensagem específica
SELECT
  id,
  channel_message_id,
  status,
  delivered_at,
  read_at,
  played_at,
  created_at
FROM messages
WHERE channel_message_id = 'wamid.HBgN...';

-- Ver últimas mensagens e seus status
SELECT
  id,
  status,
  content_type,
  delivered_at IS NOT NULL as delivered,
  read_at IS NOT NULL as read,
  played_at IS NOT NULL as played
FROM messages
ORDER BY created_at DESC
LIMIT 10;
```

### Nos Logs da API

```bash
# Ver ACKs sendo processados
docker logs ventros-api | grep "Message status updated from ACK"

# Ver apenas ACKs de PLAYED
docker logs ventros-api | grep "PLAYED"

# Ver últimas 20 linhas com ACKs
docker logs ventros-api | grep "ack" | tail -20
```

**Exemplo de log esperado:**

```
INFO: Message status updated from ACK
  message_id: "8c0d012d-9fd5-4c0b-8d9e-09ce7a0c7fc7"
  channel_message_id: "wamid.HBgNNTU0NDk3MDQ0NDc0FQIAERgS..."
  new_status: "played"
  ack: "PLAYED"
  ack_value: 4
  ack_name: "PLAYED"
```

## 📊 Fluxo de ACKs do WhatsApp

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ pending  │ ──→ │   sent   │ ──→ │delivered │ ──→ │   read   │ ──→ │  played  │
│ (queue)  │     │  (ACK 1) │     │ (ACK 2)  │     │ (ACK 3)  │     │ (ACK 4)  │
└──────────┘     └──────────┘     └──────────┘     └──────────┘     └──────────┘
                        ✓              ✓✓             ✓✓ (azul)         ▶️
```

**Notas:**
- ACK 4 (PLAYED) **SOMENTE** ocorre para mensagens de **voz/áudio** (voice messages)
- Mensagens de texto, imagem, vídeo e documento param no ACK 3 (READ)
- ACKs podem chegar fora de ordem (ex: receber ACK 3 antes do ACK 2)
- ACKs duplicados são tratados automaticamente (idempotência)

## 🐛 Debugging

### ACK não atualiza o status

1. Verifique se o webhook_id está correto:
```sql
SELECT id, name, webhook_id FROM channels WHERE type = 'waha';
```

2. Verifique se a mensagem tem channel_message_id:
```sql
SELECT id, channel_message_id, status FROM messages WHERE id = 'message-uuid';
```

3. Verifique os logs do RabbitMQ:
```bash
docker logs ventros-rabbitmq | grep "message.ack"
```

4. Verifique se o processador está rodando:
```bash
docker logs ventros-api | grep "WAHA Raw Event Processor"
```

### Mensagem não encontrada para ACK

Isso pode acontecer se:
- O `channel_message_id` está incorreto
- A mensagem ainda não foi processada pelo WAHA
- A mensagem foi deletada

**Solução:** Espere alguns segundos após enviar a mensagem antes de enviar ACKs.

## 📚 Referências

- [Guia de Implementação de ACKs](../../guides/ACK_IMPLEMENTATION_GUIDE.md)
- [WAHA Docs - Message ACK Events](https://waha.devlike.pro/docs/how-to/webhooks/#messageack)
- [WhatsApp Business API - Message Status](https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/components#message-object)

## 💡 Dicas

1. **Teste ACK 4 apenas com mensagens de voz/áudio:** Somente voice messages têm ACK 4 (PLAYED)
2. **Aguarde entre ACKs:** Na prática, ACKs chegam com alguns segundos de intervalo
3. **Use o script completo primeiro:** O `test_ack_simulation.sh` cria todo o ambiente necessário
4. **Use o script rápido para testes:** O `send_ack_webhook.sh` é útil para testar mensagens já existentes

## ✅ Checklist para Testar ACKs

- [ ] API rodando localmente
- [ ] Banco de dados configurado
- [ ] Variáveis de ambiente no `.env`
- [ ] Script de teste executável (`chmod +x`)
- [ ] Rodar teste completo: `./tests/e2e/test_ack_simulation.sh`
- [ ] Verificar logs: `docker logs ventros-api | grep "Message status updated from ACK"`
- [ ] Verificar no banco: `SELECT status, delivered_at, read_at, played_at FROM messages WHERE id = '...'`
- [ ] Testar mensagem individual com `send_ack_webhook.sh`
- [ ] Confirmar que ACK 4 (PLAYED) funciona para mensagens de voz/áudio
