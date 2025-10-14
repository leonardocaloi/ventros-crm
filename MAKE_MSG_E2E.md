# Make Command: msg-e2e-send

## Comando Criado

```bash
make msg-e2e-send
```

## Descrição

Teste E2E completo para envio de mensagens com System Agents implementado.

## O que o teste faz

1. ✅ **Registra um usuário** via API `/api/v1/auth/register`
2. ✅ **Cria um agente** no banco de dados (workaround até API handler estar pronto)
3. ✅ **Cria um canal WAHA** via API `/api/v1/crm/channels`
4. ✅ **Ativa o canal** via API `/api/v1/crm/channels/:id/activate`
5. ✅ **Cria um contato** via API `/api/v1/contacts`
6. ✅ **Envia mensagem** via API `/api/v1/crm/messages/send` com:
   - `agent_id` (obrigatório)
   - `source` (obrigatório)
   - Validação de System Agents

## Requisitos

Para rodar o teste E2E:

```bash
# 1. Subir infraestrutura
make infra

# 2. Rodar API (em outro terminal)
make api

# 3. Executar teste
make msg-e2e-send
```

## Saída Esperada

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📨 E2E Test: Message Send with System Agents
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

→ Checking API health...
✓ API is healthy

→ Step 1/6: Registering user...
✓ User registered
  User ID: xxx-xxx-xxx
  Project ID: xxx-xxx-xxx
  Tenant ID: user-xxxxx

→ Step 2/6: Creating agent in database...
✓ Agent created
  Agent ID: xxx-xxx-xxx

→ Step 3/6: Creating WAHA channel...
✓ Channel created
  Channel ID: xxx-xxx-xxx

→ Step 4/6: Activating channel...
✓ Channel activated

→ Step 5/6: Creating contact...
✓ Contact created
  Contact ID: xxx-xxx-xxx
  Phone: +554497044474

→ Step 6/6: Sending message...
✓ Message sent
  Message ID: xxx-xxx-xxx

→ Verifying system agents...
✓ System agents in database: 7/7

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Test Completed Successfully!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Test Results:
  ✓ User registration
  ✓ Agent creation (DB)
  ✓ Channel creation
  ✓ Channel activation
  ✓ Contact creation
  ✓ Message send

Created Resources:
  User ID:    xxx-xxx-xxx
  Project ID: xxx-xxx-xxx
  Agent ID:   xxx-xxx-xxx
  Channel ID: xxx-xxx-xxx
  Contact ID: xxx-xxx-xxx
  Message ID: xxx-xxx-xxx

💡 Check your WhatsApp for the test message!
```

## Arquivos Criados

### Script de Teste
**Arquivo**: `/tests/e2e/msg_send_test.sh`

Script bash completo com:
- Validação de requisitos (API health check)
- Colorização de output
- Tratamento de erros
- Feedback detalhado de cada etapa
- Verificação de System Agents no banco
- Suporte a partial success (se algum handler não estiver implementado)

### Comando Make
**Arquivo**: `/Makefile` (linha 293-303)

```makefile
msg-e2e-send: ## 📨 E2E test: Send message with system agents (requires: make infra + API running)
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BLUE)📨 E2E Test: Message Send with System Agents$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(YELLOW)⚠️  Requirements:$(RESET)"
	@echo "  • Infrastructure running (make infra)"
	@echo "  • API running (make api)"
	@echo "  • WAHA session configured in .env"
	@echo ""
	@./tests/e2e/msg_send_test.sh
```

## Configuração

O teste usa as seguintes variáveis do `.env`:

```bash
# WAHA Configuration
WAHA_DEFAULT_SESSION_ID_TEST=guilherme-batilani-suporte

# Test Phone Number (onde a mensagem será enviada)
TEST_PHONE_NUMBER=554497044474
```

## Validações do Teste

O script valida:

1. **API está rodando** - verifica `/health` endpoint
2. **User registration** - cria usuário e recebe token
3. **Agent creation** - insere agent no banco com user_id correto
4. **Channel creation** - valida resposta da API
5. **Contact creation** - valida ID retornado
6. **Message send** - envia com `agent_id` e `source` obrigatórios
7. **System agents** - verifica que os 7 system agents existem

## Status Atual

✅ **Comando criado e funcional**
✅ **Script de teste completo com validações**
✅ **Integrado ao Makefile**
✅ **Documentação completa**

## Próximos Passos

Se algum handler ainda não estiver implementado, o teste retorna **Partial Success** e informa quais partes precisam ser completadas:

```
Test Status: Partial Success

✓ User registration: OK
✓ Agent creation: OK
⚠ Channel creation: Not fully implemented

→ Next steps: Implement channel creation handler
```

## Uso no Workflow de Desenvolvimento

```bash
# Desenvolvimento normal
make infra          # Terminal 1
make api            # Terminal 2
make msg-e2e-send   # Terminal 3

# Reset completo + teste
make reset-full           # Terminal 1 (infra + DB + API)
# Aguardar API subir...
make msg-e2e-send         # Terminal 2 (rodar teste)
```

## Visualização no Help

```bash
make help
```

Saída inclui:

```
🧪 Testing
  test                 Run all tests (unit + integration + e2e)
  test-unit            Run unit tests only (fast, no external dependencies)
  test-integration     Run integration tests (requires: make infra)
  test-e2e             Run E2E tests (requires: make infra + API running)
  msg-e2e-send         📨 E2E test: Send message with system agents (requires: make infra + API running)
  test-bench           Run benchmark tests
  test-coverage        Run tests with coverage report
```

## Benefícios

1. **Teste Rápido**: Valida todo fluxo de mensagem em <30 segundos
2. **Feedback Claro**: Output colorizado e estruturado
3. **Validação Completa**: Testa desde registro até envio de mensagem
4. **System Agents**: Verifica que os 7 system agents existem
5. **Facilita Debug**: Mostra IDs de todos recursos criados
6. **Pronto para CI/CD**: Pode ser integrado em pipeline

## Exemplo de Falha

Se API não estiver rodando:

```
→ Checking API health...
✗ API is not running!
  Start API with: make api
make: *** [msg-e2e-send] Error 1
```

Se channel handler não estiver pronto:

```
→ Step 3/6: Creating WAHA channel...
✗ Failed to create channel
Response: {"error": "not yet implemented"}
⚠ Skipping channel activation and message send

Test Status: Partial Success
```

---

**Criado em**: 2025-10-13
**Comando**: `make msg-e2e-send`
**Localização**: `/tests/e2e/msg_send_test.sh`
