# Ventros CRM - Makefile
# Container runtime (docker ou podman)
CONTAINER_RUNTIME ?= docker
COMPOSE = $(CONTAINER_RUNTIME) compose

.PHONY: help

##@ Ajuda
help: ## Mostra esta ajuda
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1mVentros CRM - Comandos Disponíveis\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ 🚀 Workflows Principais

.PHONY: infra api dev infra-stop infra-clean infra-reset infra-logs infra-ps dev-stop dev-clean dev-logs
.PHONY: container container-stop container-clean container-logs
.PHONY: k8s k8s-upgrade k8s-delete k8s-logs k8s-pods k8s-status

infra: ## [INFRA] Sobe APENAS infraestrutura (PostgreSQL, RabbitMQ, Redis, Temporal)
	@echo "📦 Subindo Infraestrutura"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🚀 Subindo serviços..."
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml up -d
	@echo ""
	@echo "⏳ Aguardando serviços (15s)..."
	@sleep 15
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Infraestrutura pronta!"
	@echo ""
	@echo "🌐 Serviços disponíveis:"
	@echo "   • PostgreSQL:  localhost:5432 (ventros/ventros123)"
	@echo "   • RabbitMQ:    localhost:5672 (UI: http://localhost:15672)"
	@echo "   • Redis:       localhost:6379"
	@echo "   • Temporal:    localhost:7233 (UI: http://localhost:8088)"
	@echo ""
	@echo "💡 Migrations e RLS são automáticos (na inicialização da API)"
	@echo ""
	@echo "🎯 Próximo passo:"
	@echo "   make api    # Roda a API (faz migrations + RLS automaticamente)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

api: ## [API] Roda APENAS a API (requer infra rodando)
	@echo "🎯 Rodando API"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📚 Gerando Swagger docs..."
	@swag fmt
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@echo ""
	@echo "🌐 Endpoints:"
	@echo "   • API:     http://localhost:8080"
	@echo "   • Swagger: http://localhost:8080/swagger/index.html"
	@echo "   • Health:  http://localhost:8080/health"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@go run cmd/api/main.go

dev: ## [DEV] Sobe infra + API (via compose.api.yaml)
	@echo "🚀 Modo Desenvolvimento Completo"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📦 Subindo infraestrutura..."
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml up -d
	@echo ""
	@echo "⏳ Aguardando serviços (15s)..."
	@sleep 15
	@echo ""
	@echo "✅ Infraestrutura pronta!"
	@echo ""
	@echo "🎯 Agora rode a API em outro terminal:"
	@echo "   make api"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

infra-stop: ## Para infraestrutura
	@echo "🛑 Parando infraestrutura..."
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml down

infra-clean: ## Para e remove volumes (DESTRUTIVO - apaga TODOS os dados)
	@echo "⚠️  ATENÇÃO: Isso vai APAGAR todos os dados!"
	@echo "   • API (se estiver rodando)"
	@echo "   • PostgreSQL (tabelas, dados)"
	@echo "   • RabbitMQ (filas, mensagens)"
	@echo "   • Redis (cache)"
	@echo "   • Temporal (workflows)"
	@echo ""
	@echo "Pressione Ctrl+C para cancelar, ou Enter para continuar..."
	@read confirm
	@echo ""
	@echo "🛑 Parando API (se estiver rodando)..."
	@-pkill -f "go run cmd/api/main.go" 2>/dev/null || true
	@-pkill -f "ventros-crm" 2>/dev/null || true
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@sleep 1
	@echo "🗑️  Removendo containers e volumes..."
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml down -v
	@echo "✅ Tudo limpo!"

infra-reset: infra-clean infra ## Para, limpa volumes e sobe infra novamente (FRESH START)

infra-logs: ## Mostra logs da infra
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml logs -f

infra-ps: ## Lista containers da infra
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml ps

# Aliases
dev-stop: infra-stop ## Alias para infra-stop
dev-clean: infra-clean ## Alias para infra-clean
dev-logs: infra-logs ## Alias para infra-logs

container: ## [CONTAINER] Sobe tudo containerizado (infra + API)
	@echo "🐳 Modo Containerizado"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🔨 1. Building imagem..."
	@$(CONTAINER_RUNTIME) build -f .deploy/container/Containerfile -t ventros-crm:latest .
	@echo ""
	@echo "🚀 2. Subindo full stack..."
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.yaml up -d
	@echo ""
	@echo "⏳ 3. Aguardando API (30s)..."
	@sleep 30
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Stack completo rodando!"
	@echo ""
	@echo "🌐 Endpoints:"
	@echo "   • API:         http://localhost:8080"
	@echo "   • Health:      http://localhost:8080/health"
	@echo "   • Swagger:     http://localhost:8080/swagger/index.html"
	@echo "   • RabbitMQ UI: http://localhost:15672 (guest/guest)"
	@echo "   • Temporal UI: http://localhost:8088"
	@echo ""
	@echo "🔍 Testar health:"
	@echo "   curl http://localhost:8080/health"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

container-stop: ## Para containers
	@echo "🛑 Parando containers..."
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.yaml down

container-clean: ## Para e remove volumes (DESTRUTIVO)
	@echo "⚠️  Removendo containers e volumes..."
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.yaml down -v
	@echo "✅ Limpo!"

container-logs: ## Mostra logs dos containers
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.yaml logs -f

k8s: ## [K8S] Deploy no Minikube com Helm
	@echo "☸️  Deploy Kubernetes com Helm"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🔍 1. Verificando Minikube..."
	@minikube status || (echo "❌ Minikube não está rodando. Execute: minikube start" && exit 1)
	@echo ""
	@echo "📦 2. Instalando Helm chart..."
	@helm install ventros-crm ./.deploy/helm/ventros-crm \
		-n ventros-crm \
		--create-namespace \
		-f .deploy/helm/ventros-crm/values-dev.yaml
	@echo ""
	@echo "⏳ 3. Aguardando pods (30s)..."
	@sleep 30
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Deploy concluído!"
	@echo ""
	@echo "🔍 Ver status:"
	@echo "   kubectl get pods -n ventros-crm"
	@echo ""
	@echo "🌐 Acessar API:"
	@echo "   kubectl port-forward -n ventros-crm svc/ventros-crm 8080:8080"
	@echo "   Depois: http://localhost:8080"
	@echo ""
	@echo "📋 Ver logs:"
	@echo "   kubectl logs -n ventros-crm -l app.kubernetes.io/name=ventros-crm -f"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

k8s-upgrade: ## Atualiza deploy no K8s
	@echo "🔄 Atualizando Helm release..."
	@helm upgrade ventros-crm ./.deploy/helm/ventros-crm \
		-n ventros-crm \
		-f .deploy/helm/ventros-crm/values-dev.yaml

k8s-delete: ## Remove do K8s
	@echo "🗑️  Removendo do Kubernetes..."
	@helm uninstall ventros-crm -n ventros-crm || true
	@echo "⏳ Aguardando namespace ser removido..."
	@kubectl delete namespace ventros-crm --force --grace-period=0 2>/dev/null || true
	@echo "✅ Removido!"

k8s-logs: ## Mostra logs do K8s
	@kubectl logs -n ventros-crm -l app.kubernetes.io/name=ventros-crm -f

k8s-pods: ## Lista pods do K8s
	@kubectl get pods -n ventros-crm

k8s-status: ## Status completo do K8s
	@kubectl get all -n ventros-crm

##@ 🛠️  Utilitários

.PHONY: run build test test-domain test-domain-coverage test-e2e test-waha test-all
.PHONY: setup-webhook-n8n e2e-webhook test-coverage lint swagger
.PHONY: migrate migrate-force db-seed db-clean clean deps

run: api ## Alias para 'make api'

build: ## Compila binário
	@echo "🔨 Compilando..."
	@go build -o ventros-crm cmd/api/main.go
	@echo "✅ Binário: ./ventros-crm"

test: ## Roda testes unitários
	@echo "🧪 Rodando testes unitários..."
	@go test -v -race ./internal/... ./infrastructure/...

test-domain: ## Roda testes de domínio com coverage
	@echo "🧪 Rodando testes de domínio..."
	@go test -v -race -coverprofile=coverage-domain.out ./internal/domain/...
	@go tool cover -func=coverage-domain.out
	@echo ""
	@echo "✅ Relatório detalhado: make test-domain-coverage"

test-domain-coverage: ## Abre relatório HTML de coverage dos testes de domínio
	@echo "📊 Gerando relatório de coverage..."
	@go test -v -race -coverprofile=coverage-domain.out ./internal/domain/...
	@go tool cover -html=coverage-domain.out -o coverage-domain.html
	@echo "✅ Relatório salvo em: coverage-domain.html"

test-e2e: ## Roda testes E2E (requer API rodando)
	@echo "🧪 Rodando testes E2E"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  Certifique-se que a API está rodando:"
	@echo "   Terminal 1: make infra"
	@echo "   Terminal 2: make api"
	@echo ""
	@echo "🔍 Testando conexão com API..."
	@curl -f -s http://localhost:8080/health > /dev/null || (echo "❌ API não está rodando!" && exit 1)
	@echo "✅ API respondendo!"
	@echo ""
	@echo "🚀 Executando testes E2E..."
	@echo ""
	@go test -v -timeout 5m ./tests/e2e/...
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Testes E2E concluídos!"

test-waha: ## Roda testes E2E do webhook WAHA (requer API rodando)
	@echo "🧪 Rodando testes E2E - WAHA Webhook"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  Certifique-se que a API está rodando:"
	@echo "   Terminal 1: make infra"
	@echo "   Terminal 2: make api"
	@echo ""
	@echo "🔍 Testando conexão com API..."
	@curl -f -s http://localhost:8080/health > /dev/null || (echo "❌ API não está rodando!" && exit 1)
	@echo "✅ API respondendo!"
	@echo ""
	@echo "🚀 Executando testes WAHA..."
	@echo ""
	@go test -v -timeout 10m -run TestWAHAWebhookTestSuite ./tests/e2e/
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Testes WAHA concluídos!"

test-all: test test-e2e ## Roda todos os testes (unit + E2E)

setup-webhook-n8n: ## [SETUP] Configura webhook N8N com todos eventos de domínio (WEBHOOK_URL=https://dev.webhook.n8n.ventros.cloud/webhook/ventros/crm/events/all API_BASE_URL=http://localhost:8080)
	@echo "🧪 Configurando Webhook N8N para Eventos de Domínio"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@API_URL="$(API_BASE_URL)"; \
	if [ -z "$$API_URL" ]; then \
		API_URL="http://localhost:8080"; \
	fi; \
	WEBHOOK="$(WEBHOOK_URL)"; \
	if [ -z "$$WEBHOOK" ]; then \
		WEBHOOK="https://dev.webhook.n8n.ventros.cloud/webhook/ventros/crm/events/all"; \
	fi; \
	echo "🌐 API URL: $$API_URL"; \
	echo "🔗 Webhook URL: $$WEBHOOK"; \
	echo ""; \
	echo "🔍 Testando conexão com API..."; \
	curl -f -s $$API_URL/health > /dev/null || (echo "❌ API não está rodando em $$API_URL!" && exit 1); \
	echo "✅ API respondendo!"; \
	echo ""; \
	echo "1️⃣ Configurando ambiente de teste..."; \
	SETUP_RESPONSE=$$(curl -s -X POST "$$API_URL/api/v1/test/setup?webhook_url=$$WEBHOOK&api_base_url=$$API_URL"); \
	USER_ID=$$(echo $$SETUP_RESPONSE | jq -r '.data.user_id'); \
	PROJECT_ID=$$(echo $$SETUP_RESPONSE | jq -r '.data.project_id'); \
	PIPELINE_ID=$$(echo $$SETUP_RESPONSE | jq -r '.data.pipeline_id'); \
	CHANNEL_ID=$$(echo $$SETUP_RESPONSE | jq -r '.data.channel_id'); \
	CHANNEL_WEBHOOK_URL=$$(echo $$SETUP_RESPONSE | jq -r '.data.channel_webhook_url'); \
	WEBHOOK_ID=$$(echo $$SETUP_RESPONSE | jq -r '.data.webhook_id'); \
	API_KEY=$$(echo $$SETUP_RESPONSE | jq -r '.data.api_key'); \
	echo "✅ User: $$USER_ID"; \
	echo "✅ Project: $$PROJECT_ID"; \
	echo "✅ Pipeline: $$PIPELINE_ID"; \
	echo "✅ Channel: $$CHANNEL_ID"; \
	echo "✅ Webhook Subscription: $$WEBHOOK_ID"; \
	API_KEY_SHORT=$$(echo "$$API_KEY" | cut -c1-20); \
	echo "✅ API Key: $${API_KEY_SHORT}..."; \
	echo ""; \
	echo "2️⃣ Atualizando timeout da sessão para 1 minuto (teste)..."; \
	curl -s -X PUT "$$API_URL/api/v1/test/pipeline/$$PIPELINE_ID/timeout?minutes=1" \
		-H "Authorization: Bearer $$API_KEY" > /dev/null; \
	echo "✅ Timeout atualizado para 1 minuto!"; \
	echo ""; \
	echo "3️⃣ Atualizando webhook com todos os eventos de domínio..."; \
	UPDATE_RESPONSE=$$(curl -s -X PUT "$$API_URL/api/v1/webhook-subscriptions/$$WEBHOOK_ID" \
		-H "Authorization: Bearer $$API_KEY" \
		-H "Content-Type: application/json" \
		-d '{"name":"Webhook N8N - Todos Eventos","url":"'"$$WEBHOOK"'","events":["contact.created","contact.updated","contact.deleted","contact.merged","contact.enriched","session.started","session.ended","session.agent_assigned","session.resolved","session.escalated","session.summarized","session.abandoned","tracking.message.meta_ads","pipeline.created","pipeline.updated","pipeline.activated","pipeline.deactivated","status.created","status.updated","contact.status_changed","contact.entered_pipeline","contact.exited_pipeline"],"active":true,"retry_count":3,"timeout_seconds":30}'); \
	echo "✅ Webhook atualizado!"; \
	echo ""; \
	echo "4️⃣ Verificando eventos configurados..."; \
	WEBHOOK_INFO=$$(curl -s -X GET "$$API_URL/api/v1/webhook-subscriptions/$$WEBHOOK_ID" \
		-H "Authorization: Bearer $$API_KEY"); \
	echo "📋 Eventos ativos:"; \
	echo $$WEBHOOK_INFO | jq -r '.webhook.events[] | "   ✓ \(.)"'; \
	echo ""; \
	echo "5️⃣ Enviando TODAS as mensagens de teste para gerar eventos..."; \
	echo ""; \
	SESSION_ID=$$(echo $$CHANNEL_WEBHOOK_URL | sed 's/.*waha\///'); \
	echo "📝 Enviando mensagem de texto..."; \
	sed "s/\"session\": \"[^\"]*\"/\"session\": \"$$SESSION_ID\"/" events_waha/message_text.json | curl -s -X POST "$$CHANNEL_WEBHOOK_URL" -H "Content-Type: application/json" -d @- > /dev/null; \
	sleep 1; \
	echo "🖼️  Enviando mensagem de imagem..."; \
	sed "s/\"session\": \"[^\"]*\"/\"session\": \"$$SESSION_ID\"/" events_waha/message_image.json | curl -s -X POST "$$CHANNEL_WEBHOOK_URL" -H "Content-Type: application/json" -d @- > /dev/null; \
	sleep 1; \
	echo "🎤 Enviando mensagem de voz (PTT)..."; \
	sed "s/\"session\": \"[^\"]*\"/\"session\": \"$$SESSION_ID\"/" events_waha/message_recorded_audio.json | curl -s -X POST "$$CHANNEL_WEBHOOK_URL" -H "Content-Type: application/json" -d @- > /dev/null; \
	sleep 1; \
	echo "📍 Enviando mensagem de localização..."; \
	sed "s/\"session\": \"[^\"]*\"/\"session\": \"$$SESSION_ID\"/" events_waha/message_location.json | curl -s -X POST "$$CHANNEL_WEBHOOK_URL" -H "Content-Type: application/json" -d @- > /dev/null; \
	sleep 1; \
	echo "👤 Enviando mensagem de contato..."; \
	sed "s/\"session\": \"[^\"]*\"/\"session\": \"$$SESSION_ID\"/" events_waha/message_contact.json | curl -s -X POST "$$CHANNEL_WEBHOOK_URL" -H "Content-Type: application/json" -d @- > /dev/null; \
	sleep 1; \
	echo "📄 Enviando mensagem de documento..."; \
	sed "s/\"session\": \"[^\"]*\"/\"session\": \"$$SESSION_ID\"/" events_waha/message_document_pdf.json | curl -s -X POST "$$CHANNEL_WEBHOOK_URL" -H "Content-Type: application/json" -d @- > /dev/null; \
	sleep 1; \
	echo "🔊 Enviando mensagem de áudio..."; \
	sed "s/\"session\": \"[^\"]*\"/\"session\": \"$$SESSION_ID\"/" events_waha/message_audio.json | curl -s -X POST "$$CHANNEL_WEBHOOK_URL" -H "Content-Type: application/json" -d @- > /dev/null; \
	sleep 1; \
	echo "🖼️📝 Enviando mensagem de imagem com texto..."; \
	sed "s/\"session\": \"[^\"]*\"/\"session\": \"$$SESSION_ID\"/" events_waha/message_image_text.json | curl -s -X POST "$$CHANNEL_WEBHOOK_URL" -H "Content-Type: application/json" -d @- > /dev/null; \
	sleep 1; \
	echo "📢 Enviando mensagem de FB Ads (tracking)..."; \
	sed "s/\"session\": \"[^\"]*\"/\"session\": \"$$SESSION_ID\"/" events_waha/fb_ads_message.json | curl -s -X POST "$$CHANNEL_WEBHOOK_URL" -H "Content-Type: application/json" -d @- > /dev/null; \
	sleep 2; \
	echo "✅ Todas as 9 mensagens enviadas!"; \
	echo ""; \
	echo "5️⃣ Verificando canal..."; \
	CHANNEL_INFO=$$(curl -s -X GET "$$API_URL/api/v1/channels/$$CHANNEL_ID" \
		-H "Authorization: Bearer $$API_KEY"); \
	echo $$CHANNEL_INFO | jq '.channel | {id, name, type, webhook_url, webhook_active, messages_received}'; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ Webhook N8N configurado com sucesso!"; \
	echo ""; \
	echo "📤 Eventos que serão enviados para N8N:"; \
	echo "   • Contatos: created, updated, deleted, merged, enriched"; \
	echo "   • Sessões: started, ended, agent_assigned, resolved, escalated, summarized, abandoned"; \
	echo "   • Tracking: tracking.message.meta_ads (Meta Ads: FB/Instagram)"; \
	echo "   • Pipelines: created, updated, activated, deactivated"; \
	echo "   • Status: created, updated, contact.status_changed, contact.entered_pipeline, contact.exited_pipeline"; \
	echo ""; \
	echo "🔗 Webhook URL: $$WEBHOOK"; \
	echo "📋 Webhook ID: $$WEBHOOK_ID"; \
	echo "🔑 API Key: $$API_KEY"; \
	echo ""; \
	echo "💡 Para testar, envie mensagens para o canal ou use:"; \
	echo "   curl -X POST \"$$CHANNEL_WEBHOOK_URL\" \\"; \
	echo "     -H \"Content-Type: application/json\" \\"; \
	echo "     -d @events_waha/message_text.json"; \
	echo ""; \
	echo "🌐 Verifique os eventos em: $$WEBHOOK"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

e2e-webhook: ## [E2E] Teste completo: Canal WAHA + Webhook + Mensagem FB Ads (WEBHOOK_URL=https://webhook.site/xxx API_BASE_URL=http://localhost:8080)
	@echo "🧪 E2E: Canal WAHA com Webhook e FB Ads"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@API_URL="$(API_BASE_URL)"; \
	if [ -z "$$API_URL" ]; then \
		API_URL="http://localhost:8080"; \
	fi; \
	echo "🌐 Base URL: $$API_URL"; \
	echo ""; \
	echo "⚠️  Certifique-se que a API está rodando em $$API_URL"; \
	echo "   Terminal 1: make infra"; \
	echo "   Terminal 2: make api"; \
	echo ""; \
	echo "🔍 Testando conexão com API..."; \
	curl -f -s $$API_URL/health > /dev/null || (echo "❌ API não está rodando em $$API_URL!" && exit 1); \
	echo "✅ API respondendo!"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "1️⃣ Setup ambiente de teste..."; \
	QUERY_PARAMS="?api_base_url=$$API_URL"; \
	if [ -n "$(WEBHOOK_URL)" ]; then \
		QUERY_PARAMS="$$QUERY_PARAMS&webhook_url=$(WEBHOOK_URL)"; \
		echo "🔗 Usando webhook externo: $(WEBHOOK_URL)"; \
	fi; \
	SETUP_RESPONSE=$$(curl -s -X POST "$$API_URL/api/v1/test/setup$$QUERY_PARAMS"); \
	echo $$SETUP_RESPONSE | jq -r '.data | "✅ User: \(.user_id)\n✅ Project: \(.project_id)\n✅ Channel: \(.channel_id)\n✅ Channel Webhook: \(.channel_webhook_url)\n✅ Webhook Subscription: \(.webhook_id)\n✅ API Key: \(.api_key)"'; \
	WEBHOOK_ID=$$(echo $$SETUP_RESPONSE | jq -r '.data.webhook_id'); \
	CHANNEL_ID=$$(echo $$SETUP_RESPONSE | jq -r '.data.channel_id'); \
	CHANNEL_WEBHOOK_URL=$$(echo $$SETUP_RESPONSE | jq -r '.data.channel_webhook_url'); \
	API_KEY=$$(echo $$SETUP_RESPONSE | jq -r '.data.api_key'); \
	echo ""; \
	echo "📋 Eventos ativos no webhook de teste:"; \
	curl -s -X GET $$API_URL/api/v1/webhook-subscriptions/$$WEBHOOK_ID \
		-H "Authorization: Bearer $$API_KEY" | jq -r '.webhook.events[] | "   ✓ \(.)"'; \
	echo ""; \
	echo "2️⃣ Webhook do canal já configurado automaticamente!"; \
	echo "📍 URL: $$CHANNEL_WEBHOOK_URL"; \
	echo ""; \
	echo "3️⃣ Simulando mensagem do FB Ads no webhook..."; \
	WEBHOOK_RESPONSE=$$(curl -s -X POST "$$CHANNEL_WEBHOOK_URL" \
		-H "Content-Type: application/json" \
		-d '{"id":"evt_e2e_fb_ads","timestamp":1696598400000,"event":"message","session":"test-session-waha","payload":{"id":"msg_fb_001","from":"5511999999999@c.us","fromMe":false,"body":"Olá! Tenho interesse na imersão e queria mais informações, por favor.","_data":{"Info":{"PushName":"Cliente FB Ads"},"Message":{"extendedTextMessage":{"contextInfo":{"conversionSource":"FB_Ads","entryPointConversionSource":"ctwa_ad","entryPointConversionApp":"instagram","ctwaClid":"test_click_id_123"}}}}}}'); \
	echo $$WEBHOOK_RESPONSE | jq -r '"✅ Webhook processado: \(.status)"'; \
	echo ""; \
	echo "4️⃣ Verificando canal atualizado..."; \
	curl -s -X GET $$API_URL/api/v1/channels/$$CHANNEL_ID \
		-H "Authorization: Bearer $$API_KEY" | jq '.channel | {id,name,type,webhook_url,webhook_active,messages_received}'; \
	echo ""; \
	if [ -n "$(WEBHOOK_URL)" ]; then \
		echo "5️⃣ Eventos de domínio serão enviados para: $(WEBHOOK_URL)"; \
		echo "📤 Eventos esperados:"; \
		echo "   ✓ contact.created (imediato)"; \
		echo "   ✓ session.started (imediato)"; \
		echo "   ✓ tracking.message.meta_ads (imediato)"; \
		echo "   ✓ session.ended (após 1 minuto de inatividade)"; \
		echo ""; \
		echo "💡 Verifique em: $(WEBHOOK_URL)"; \
		echo "⏰ O evento session.ended chegará em ~1 minuto"; \
		echo ""; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ Teste E2E completo!"; \
	echo ""; \
	echo "🔍 O que aconteceu:"; \
	echo "   • Canal WAHA criado com webhook URL automático"; \
	echo "   • Mensagem FB Ads enviada para o webhook"; \
	echo "   • Contact criado com tracking do FB Ads"; \
	echo "   • Session iniciada"; \
	echo "   • Message salva"; \
	echo "   • Eventos de domínio disparados para webhook externo"; \
	echo ""; \
	echo "💡 Exemplos de uso:"; \
	echo "   # Webhook externo:"; \
	echo "   WEBHOOK_URL=https://webhook.site/xxx make e2e-webhook"; \
	echo ""; \
	echo "   # API na nuvem:"; \
	echo "   API_BASE_URL=https://sua-api.com make e2e-webhook"; \
	echo ""; \
	echo "   # Ambos:"; \
	echo "   API_BASE_URL=https://sua-api.com WEBHOOK_URL=https://webhook.site/xxx make e2e-webhook"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test-coverage: ## Testes com coverage
	@echo "🧪 Rodando testes com coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage: coverage.html"

lint: ## Roda linters
	@echo "🔍 Rodando golangci-lint..."
	@golangci-lint run

swagger: ## Gera documentação Swagger
	@echo "📚 Gerando Swagger..."
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@echo "✅ Docs: http://localhost:8080/swagger/index.html"

migrate: ## Roda migrations GORM (manual)
	@echo "🔄 Rodando migrations manualmente..."
	@echo "⚠️  Normalmente não é necessário - a API faz AutoMigrate"
	@go run cmd/migrate-gorm/main.go
	@echo "✅ Migrations concluídas!"

migrate-force: infra-clean infra api ## Força fresh start (limpa DB + sobe + migrations automáticas)

db-seed: ## Popula banco com dados de teste
	@echo "🌱 Seeding database..."
	@echo "⚠️  Arquivo seed.sql não encontrado - usar scripts/run-seeds.sh se disponível"
	@echo "✅ Seed completo!"

db-clean: ## Limpa database (DESTRUTIVO)
	@echo "⚠️  Isso vai limpar TODOS os dados!"
	@echo "Pressione Ctrl+C para cancelar, ou Enter para continuar..."
	@read confirm
	@PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" 2>/dev/null
	@$(MAKE) migrate
	@echo "✅ Database limpo!"

clean: ## Remove arquivos gerados
	@echo "🧹 Limpando..."
	@rm -rf docs/swagger/
	@rm -f ventros-crm coverage.out coverage.html
	@rm -f cmd/api/api
	@echo "✅ Limpo!"

deps: ## Atualiza dependências Go
	@echo "📦 Atualizando dependências..."
	@go mod download
	@go mod tidy

##@ 📊 Debug e Health

.PHONY: health setup-waha-channel setup-waha-complete

health: ## Checa saúde da API
	@curl -s http://localhost:8080/health | jq . || echo "❌ API não responde"

setup-waha-channel: ## Cria e ativa canal WAHA completo (requer API rodando)
	@echo "🚀 Setup Canal WAHA Completo"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@API_URL="http://localhost:8080"; \
	WAHA_URL="https://waha.ventros.cloud"; \
	WAHA_TOKEN="4bffec302d5f4312b8b73700da3ff3cb"; \
	SESSION_ID="mylomaic-rotate"; \
	echo "🔍 Verificando API..."; \
	curl -f -s $$API_URL/health > /dev/null || (echo "❌ API não está rodando! Execute: make api" && exit 1); \
	echo "✅ API respondendo!"; \
	echo ""; \
	echo "🔍 Verificando WAHA..."; \
	WAHA_STATUS=$$(curl -s "$$WAHA_URL/api/sessions/$$SESSION_ID" -H "X-Api-Key: $$WAHA_TOKEN" | jq -r '.status'); \
	if [ "$$WAHA_STATUS" != "WORKING" ]; then \
		echo "❌ Sessão WAHA não está WORKING (status: $$WAHA_STATUS)"; \
		exit 1; \
	fi; \
	echo "✅ WAHA sessão WORKING!"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "1️⃣ Criando usuário..."; \
	REGISTER_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"waha@ventros.com","password":"waha123","name":"WAHA User"}'); \
	echo "✅ Usuário criado"; \
	echo ""; \
	echo "2️⃣ Fazendo login..."; \
	LOGIN_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"waha@ventros.com","password":"waha123"}'); \
	USER_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.user_id'); \
	PROJECT_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.default_project_id'); \
	API_KEY=$$(echo $$LOGIN_RESPONSE | jq -r '.api_key // .token'); \
	if [ "$$USER_ID" = "null" ] || [ -z "$$USER_ID" ]; then \
		echo "❌ Erro ao obter user_id"; \
		echo $$LOGIN_RESPONSE | jq .; \
		exit 1; \
	fi; \
	if [ "$$PROJECT_ID" = "null" ] || [ -z "$$PROJECT_ID" ]; then \
		echo "❌ Erro ao obter project_id"; \
		echo $$LOGIN_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Login OK"; \
	echo "   User ID: $$USER_ID"; \
	echo "   Project ID: $$PROJECT_ID"; \
	echo ""; \
	echo "3️⃣ Criando canal WAHA com timeout de 1 minuto..."; \
	CHANNEL_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/channels \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"WhatsApp Ventros\",\"type\":\"waha\",\"waha_config\":{\"base_url\":\"$$WAHA_URL\",\"api_key\":\"$$WAHA_TOKEN\",\"session_id\":\"$$SESSION_ID\",\"import_strategy\":\"all\"}}"); \
	CHANNEL_ID=$$(echo $$CHANNEL_RESPONSE | jq -r '.id'); \
	WEBHOOK_URL=$$(echo $$CHANNEL_RESPONSE | jq -r '.channel.webhook_url'); \
	if [ "$$CHANNEL_ID" = "null" ] || [ -z "$$CHANNEL_ID" ]; then \
		echo "❌ Erro ao criar canal"; \
		echo $$CHANNEL_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Canal criado: $$CHANNEL_ID"; \
	echo "📍 Webhook URL: $$WEBHOOK_URL"; \
	echo ""; \
	echo "4️⃣ Ativando canal (health check)..."; \
	ACTIVATE_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/channels/$$CHANNEL_ID/activate-waha \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	if echo "$$ACTIVATE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then \
		echo "❌ Erro ao ativar canal"; \
		echo $$ACTIVATE_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Canal ativado!"; \
	echo ""; \
	echo "5️⃣ Verificando canal final..."; \
	CHANNEL_INFO=$$(curl -s $$API_URL/api/v1/channels/$$CHANNEL_ID \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ Canal WAHA configurado com sucesso!"; \
	echo ""; \
	echo "📋 Informações do Canal:"; \
	echo $$CHANNEL_INFO | jq '.channel | {id, name, type, status, external_id, webhook_url, webhook_active}'; \
	echo ""; \
	echo "🔑 Credenciais:"; \
	echo "   Email: waha@ventros.com"; \
	echo "   Senha: waha123"; \
	echo "   User ID: $$USER_ID"; \
	echo "   API Key: $$API_KEY"; \
	echo ""; \
	echo "📍 Webhook URL (para receber eventos):"; \
	echo "   $$WEBHOOK_URL"; \
	echo ""; \
	echo "💡 Próximos passos:"; \
	echo "   • Canal está ATIVO e pronto para receber mensagens"; \
	echo "   • Webhook configurado automaticamente"; \
	echo "   • Para importar histórico:"; \
	echo "     curl -X POST $$API_URL/api/v1/channels/$$CHANNEL_ID/import-history \\"; \
	echo "       -H \"X-Dev-User-ID: $$USER_ID\" \\"; \
	echo "       -H \"X-Dev-Project-ID: $$PROJECT_ID\" \\"; \
	echo "       -H \"Content-Type: application/json\" \\"; \
	echo "       -d '{\"limit\": 100}'"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

setup-waha-complete: ## Setup COMPLETO: Projeto + Pipeline + Canal WAHA (tudo do zero)
	@echo "🚀 Setup COMPLETO - Projeto + Pipeline + Canal WAHA"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@API_URL="http://localhost:8080"; \
	WAHA_URL="https://waha.ventros.cloud"; \
	WAHA_TOKEN="4bffec302d5f4312b8b73700da3ff3cb"; \
	SESSION_ID="mylomaic-rotate"; \
	echo "🔍 Verificando API..."; \
	curl -f -s $$API_URL/health > /dev/null || (echo "❌ API não está rodando! Execute: make api" && exit 1); \
	echo "✅ API respondendo!"; \
	echo ""; \
	echo "🔍 Verificando WAHA..."; \
	WAHA_STATUS=$$(curl -s "$$WAHA_URL/api/sessions/$$SESSION_ID" -H "X-Api-Key: $$WAHA_TOKEN" | jq -r '.status'); \
	if [ "$$WAHA_STATUS" != "WORKING" ]; then \
		echo "❌ Sessão WAHA não está WORKING (status: $$WAHA_STATUS)"; \
		exit 1; \
	fi; \
	echo "✅ WAHA sessão WORKING!"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "1️⃣ Criando usuário..."; \
	REGISTER_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"waha@ventros.com","password":"waha123","name":"WAHA User"}'); \
	echo "✅ Usuário criado"; \
	echo ""; \
	echo "2️⃣ Fazendo login..."; \
	LOGIN_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"waha@ventros.com","password":"waha123"}'); \
	USER_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.user_id'); \
	PROJECT_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.default_project_id'); \
	if [ "$$USER_ID" = "null" ] || [ -z "$$USER_ID" ]; then \
		echo "❌ Erro ao obter user_id"; \
		echo $$LOGIN_RESPONSE | jq .; \
		exit 1; \
	fi; \
	if [ "$$PROJECT_ID" = "null" ] || [ -z "$$PROJECT_ID" ]; then \
		echo "❌ Erro ao obter project_id"; \
		echo $$LOGIN_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Login OK"; \
	echo "   User ID: $$USER_ID"; \
	echo "   Project ID: $$PROJECT_ID (projeto default)"; \
	echo ""; \
	echo "3️⃣ Criando canal WAHA com timeout de 1 minuto..."; \
	CHANNEL_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/channels \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"WhatsApp Ventros\",\"type\":\"waha\",\"waha_config\":{\"base_url\":\"$$WAHA_URL\",\"api_key\":\"$$WAHA_TOKEN\",\"session_id\":\"$$SESSION_ID\",\"import_strategy\":\"all\"}}"); \
	CHANNEL_ID=$$(echo $$CHANNEL_RESPONSE | jq -r '.id'); \
	WEBHOOK_URL=$$(echo $$CHANNEL_RESPONSE | jq -r '.channel.webhook_url'); \
	if [ "$$CHANNEL_ID" = "null" ] || [ -z "$$CHANNEL_ID" ]; then \
		echo "❌ Erro ao criar canal"; \
		echo $$CHANNEL_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Canal criado: $$CHANNEL_ID"; \
	echo "📍 Webhook URL: $$WEBHOOK_URL"; \
	echo ""; \
	echo "4️⃣ Ativando canal (health check)..."; \
	ACTIVATE_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/channels/$$CHANNEL_ID/activate-waha \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	if echo "$$ACTIVATE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then \
		echo "❌ Erro ao ativar canal"; \
		echo $$ACTIVATE_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Canal ativado!"; \
	echo ""; \
	echo "5️⃣ Verificando configuração final..."; \
	CHANNEL_INFO=$$(curl -s $$API_URL/api/v1/channels/$$CHANNEL_ID \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ Setup COMPLETO finalizado!"; \
	echo ""; \
	echo "📋 Resumo da Configuração:"; \
	echo ""; \
	echo "👤 Usuário:"; \
	echo "   Email: waha@ventros.com"; \
	echo "   Senha: waha123"; \
	echo "   ID: $$USER_ID"; \
	echo ""; \
	echo "📁 Projeto:"; \
	echo "   Nome: Projeto Principal (default)"; \
	echo "   ID: $$PROJECT_ID"; \
	echo ""; \
	echo "📱 Canal WAHA:"; \
	echo "   ID: $$CHANNEL_ID"; \
	echo "   Status: ACTIVE"; \
	echo "   Session: $$SESSION_ID"; \
	echo "   Webhook: $$WEBHOOK_URL"; \
	echo ""; \
	echo "💡 Próximos passos:"; \
	echo "   • Canal está ATIVO e pronto para receber mensagens"; \
	echo "   • Para importar histórico:"; \
	echo "     curl -X POST $$API_URL/api/v1/channels/$$CHANNEL_ID/import-history \\"; \
	echo "       -H \"X-Dev-User-ID: $$USER_ID\" \\"; \
	echo "       -H \"X-Dev-Project-ID: $$PROJECT_ID\" \\"; \
	echo "       -H \"Content-Type: application/json\" \\"; \
	echo "       -d '{\"limit\": 100}'"; \
	echo ""; \
	echo "🧪 Testar enviando mensagem para o WhatsApp conectado!"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test-all-message-types: ## 🧪 Testa TODOS os tipos de mensagens WAHA (text, image, audio, document, location, contact, etc)
	@echo "🧪 TESTE COMPLETO - TODOS OS TIPOS DE MENSAGENS WAHA"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@API_URL="http://localhost:8080"; \
	EVENTS_DIR="events_waha"; \
	SESSION_ID="mylomaic-rotate"; \
	echo "🔍 Verificando API..."; \
	curl -f -s $$API_URL/health > /dev/null || (echo "❌ API não está rodando! Execute: make api" && exit 1); \
	echo "✅ API respondendo!"; \
	echo ""; \
	echo "🔍 Verificando pasta de eventos..."; \
	if [ ! -d "$$EVENTS_DIR" ]; then \
		echo "❌ Pasta $$EVENTS_DIR não encontrada!"; \
		exit 1; \
	fi; \
	echo "✅ Pasta encontrada: $$EVENTS_DIR"; \
	echo ""; \
	echo "📡 Webhook URL: $$API_URL/api/v1/webhooks/waha/$$SESSION_ID"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	TOTAL=0; \
	SUCCESS=0; \
	FAILED=0; \
	for json_file in $$EVENTS_DIR/*.json; do \
		if [ -f "$$json_file" ]; then \
			TOTAL=$$((TOTAL + 1)); \
			filename=$$(basename "$$json_file"); \
			echo "📨 Testando: $$filename"; \
			RESPONSE=$$(curl -s -w "\n%{http_code}" -X POST $$API_URL/api/v1/webhooks/waha/$$SESSION_ID \
				-H "Content-Type: application/json" \
				-d @$$json_file); \
			HTTP_CODE=$$(echo "$$RESPONSE" | tail -n1); \
			BODY=$$(echo "$$RESPONSE" | head -n-1); \
			if [ "$$HTTP_CODE" = "200" ] || [ "$$HTTP_CODE" = "201" ]; then \
				echo "   ✅ Status: $$HTTP_CODE - OK"; \
				SUCCESS=$$((SUCCESS + 1)); \
			else \
				echo "   ❌ Status: $$HTTP_CODE - ERRO"; \
				echo "   Response: $$BODY" | head -c 200; \
				FAILED=$$((FAILED + 1)); \
			fi; \
			echo ""; \
		fi; \
	done; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "📊 RESULTADO FINAL:"; \
	echo "   Total de testes: $$TOTAL"; \
	echo "   ✅ Sucesso: $$SUCCESS"; \
	echo "   ❌ Falhas: $$FAILED"; \
	echo ""; \
	if [ $$FAILED -gt 0 ]; then \
		echo "⚠️  Alguns testes falharam! Verifique os logs acima."; \
		echo ""; \
		echo "💡 Tipos de mensagens testados:"; \
		ls -1 $$EVENTS_DIR/*.json | xargs -n1 basename | sed 's/^/   • /'; \
		exit 1; \
	else \
		echo "✅ TODOS OS TESTES PASSARAM!"; \
		echo ""; \
		echo "💡 Tipos de mensagens testados:"; \
		ls -1 $$EVENTS_DIR/*.json | xargs -n1 basename | sed 's/^/   • /'; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

setup-all-complete: ## 🚀🚀🚀 MEGA SETUP: User + Project + Pipeline + Channel + Webhooks + TEST ALL MESSAGES + Verify DB (TUDO!)
	@echo "🚀🚀🚀 MEGA SETUP COMPLETO - DO INÍCIO AO FIM!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@API_URL="http://localhost:8080"; \
	WAHA_URL="https://waha.ventros.cloud"; \
	WAHA_TOKEN="4bffec302d5f4312b8b73700da3ff3cb"; \
	SESSION_ID="mylomaic-rotate"; \
	N8N_WEBHOOK_1="https://dev.webhook.n8n.ventros.cloud/webhook/6e0918af-876a-4126-b7c2-e1d7d715639e"; \
	N8N_WEBHOOK_2="https://dev.webhook.n8n.ventros.cloud/webhook/ventros/crm/events/all"; \
	TRACKING_WEBHOOK="https://tracking.ventros.cloud/api/events"; \
	EVENTS_DIR="events_waha"; \
	echo "🔍 Verificando API..."; \
	curl -f -s $$API_URL/health > /dev/null || (echo "❌ API não está rodando! Execute: make api" && exit 1); \
	echo "✅ API respondendo!"; \
	echo ""; \
	echo "🔍 Verificando WAHA..."; \
	WAHA_STATUS=$$(curl -s "$$WAHA_URL/api/sessions/$$SESSION_ID" -H "X-Api-Key: $$WAHA_TOKEN" | jq -r '.status'); \
	if [ "$$WAHA_STATUS" != "WORKING" ]; then \
		echo "❌ Sessão WAHA não está WORKING (status: $$WAHA_STATUS)"; \
		exit 1; \
	fi; \
	echo "✅ WAHA sessão WORKING!"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "1️⃣ Criando usuário..."; \
	REGISTER_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"waha@ventros.com","password":"waha123","name":"WAHA User"}'); \
	echo "✅ Usuário criado"; \
	echo ""; \
	echo "2️⃣ Fazendo login..."; \
	LOGIN_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"waha@ventros.com","password":"waha123"}'); \
	USER_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.user_id'); \
	PROJECT_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.default_project_id'); \
	if [ "$$USER_ID" = "null" ] || [ -z "$$USER_ID" ]; then \
		echo "❌ Erro ao obter user_id"; \
		exit 1; \
	fi; \
	if [ "$$PROJECT_ID" = "null" ] || [ -z "$$PROJECT_ID" ]; then \
		echo "❌ Erro ao obter project_id"; \
		exit 1; \
	fi; \
	echo "✅ Login OK"; \
	echo "   User ID: $$USER_ID"; \
	echo "   Project ID: $$PROJECT_ID"; \
	echo ""; \
	echo "3️⃣ Verificando pipelines ativos..."; \
	PIPELINES_RESPONSE=$$(curl -s "$$API_URL/api/v1/pipelines?project_id=$$PROJECT_ID&active=true" \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	PIPELINE_COUNT=$$(echo $$PIPELINES_RESPONSE | jq 'length'); \
	if [ "$$PIPELINE_COUNT" = "0" ] || [ "$$PIPELINE_COUNT" = "null" ]; then \
		echo "❌ ERRO: Projeto não tem pipeline ativo!"; \
		exit 1; \
	fi; \
	DEFAULT_PIPELINE_ID=$$(echo $$PIPELINES_RESPONSE | jq -r '.[0].id'); \
	echo "✅ Pipeline ativo: $$DEFAULT_PIPELINE_ID"; \
	echo ""; \
	echo "⚙️  Configurando timeout do projeto para 1 minuto (testes rápidos)..."; \
	PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "UPDATE projects SET session_timeout_minutes = 1 WHERE id = '$$PROJECT_ID';" > /dev/null; \
	echo "✅ Timeout do projeto atualizado para 1 minuto (teste rápido de session.ended)"; \
	echo ""; \
	echo "4️⃣ Criando canal WAHA com timeout de 1 minuto..."; \
	CHANNEL_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/channels \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"WhatsApp Ventros\",\"type\":\"waha\",\"waha_config\":{\"base_url\":\"$$WAHA_URL\",\"api_key\":\"$$WAHA_TOKEN\",\"session_id\":\"$$SESSION_ID\",\"import_strategy\":\"all\"}}"); \
	CHANNEL_ID=$$(echo $$CHANNEL_RESPONSE | jq -r '.id'); \
	if [ "$$CHANNEL_ID" = "null" ] || [ -z "$$CHANNEL_ID" ]; then \
		echo "❌ Erro ao criar canal"; \
		echo $$CHANNEL_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Canal criado: $$CHANNEL_ID"; \
	echo ""; \
	echo "5️⃣ Ativando canal WAHA..."; \
	ACTIVATE_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/channels/$$CHANNEL_ID/activate-waha \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	if echo "$$ACTIVATE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then \
		echo "❌ Erro ao ativar canal"; \
		echo $$ACTIVATE_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Canal ativado!"; \
	echo ""; \
	echo "6️⃣ Verificando canal no banco de dados..."; \
	DB_CHANNEL=$$(psql postgresql://ventros:ventros123@localhost:5432/ventros_crm -t -c "SELECT external_id FROM channels WHERE id='$$CHANNEL_ID' AND deleted_at IS NULL;" 2>/dev/null | xargs); \
	if [ "$$DB_CHANNEL" = "$$SESSION_ID" ]; then \
		echo "✅ Canal no DB: external_id = $$DB_CHANNEL"; \
	else \
		echo "❌ ERRO: Canal no DB tem external_id diferente!"; \
		echo "   Esperado: $$SESSION_ID"; \
		echo "   Encontrado: $$DB_CHANNEL"; \
		exit 1; \
	fi; \
	echo ""; \
	echo "7️⃣ Deletando webhooks antigos..."; \
	OLD_WEBHOOKS=$$(curl -s "$$API_URL/api/v1/webhook-subscriptions" \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" | jq -r '.[] | .id'); \
	for webhook_id in $$OLD_WEBHOOKS; do \
		curl -s -X DELETE "$$API_URL/api/v1/webhook-subscriptions/$$webhook_id" \
			-H "X-Dev-User-ID: $$USER_ID" \
			-H "X-Dev-Project-ID: $$PROJECT_ID" > /dev/null; \
		echo "   🗑️  Deletado: $$webhook_id"; \
	done; \
	echo "✅ Webhooks antigos removidos"; \
	echo ""; \
	echo "8️⃣ Criando webhooks n8n..."; \
	WEBHOOK1_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"n8n Specific Events\",\"url\":\"$$N8N_WEBHOOK_1\",\"events\":[\"contact.created\",\"contact.updated\",\"session.created\",\"session.closed\"],\"active\":true}"); \
	WEBHOOK1_ID=$$(echo $$WEBHOOK1_RESPONSE | jq -r '.id'); \
	echo "✅ Webhook 1: $$WEBHOOK1_ID (eventos específicos sem message.*)"; \
	WEBHOOK2_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"n8n All Domain Events\",\"url\":\"$$N8N_WEBHOOK_2\",\"events\":[\"contact.*\",\"session.*\",\"pipeline.*\",\"tracking.*\",\"note.*\",\"agent.*\",\"channel.*\"],\"active\":true}"); \
	WEBHOOK2_ID=$$(echo $$WEBHOOK2_RESPONSE | jq -r '.id'); \
	echo "✅ Webhook 2: $$WEBHOOK2_ID (wildcards: contact.*, session.*, pipeline.*, tracking.*, note.*, agent.*, channel.*)"; \
	TRACKING_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"Event Tracking Analytics\",\"url\":\"$$TRACKING_WEBHOOK\",\"events\":[\"contact.*\",\"session.*\",\"pipeline.*\",\"tracking.*\",\"note.*\",\"agent.*\",\"channel.*\"],\"active\":true}"); \
	TRACKING_ID=$$(echo $$TRACKING_RESPONSE | jq -r '.id'); \
	echo "✅ Tracking: $$TRACKING_ID (wildcards: contact.*, session.*, pipeline.*, tracking.*, note.*, agent.*, channel.*)"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "9️⃣ TESTANDO TODOS OS TIPOS DE MENSAGENS..."; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	if [ ! -d "$$EVENTS_DIR" ]; then \
		echo "❌ Pasta $$EVENTS_DIR não encontrada!"; \
		exit 1; \
	fi; \
	TOTAL=0; \
	SUCCESS=0; \
	FAILED=0; \
	for json_file in $$EVENTS_DIR/*.json; do \
		if [ -f "$$json_file" ]; then \
			TOTAL=$$((TOTAL + 1)); \
			filename=$$(basename "$$json_file"); \
			echo "📨 $$filename"; \
			RESPONSE=$$(curl -s -w "\n%{http_code}" -X POST $$API_URL/api/v1/webhooks/waha/$$SESSION_ID \
				-H "Content-Type: application/json" \
				-d @$$json_file); \
			HTTP_CODE=$$(echo "$$RESPONSE" | tail -n1); \
			if [ "$$HTTP_CODE" = "200" ] || [ "$$HTTP_CODE" = "201" ]; then \
				echo "   ✅ OK"; \
				SUCCESS=$$((SUCCESS + 1)); \
			else \
				echo "   ❌ ERRO ($$HTTP_CODE)"; \
				FAILED=$$((FAILED + 1)); \
			fi; \
		fi; \
	done; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "🔟 Aguardando processamento (3s)..."; \
	sleep 3; \
	echo ""; \
	echo "1️⃣1️⃣ Verificando dados no banco..."; \
	CONTACTS_COUNT=$$(psql postgresql://ventros:ventros123@localhost:5432/ventros_crm -t -c "SELECT COUNT(*) FROM contacts WHERE project_id='$$PROJECT_ID' AND deleted_at IS NULL;" 2>/dev/null | xargs); \
	SESSIONS_COUNT=$$(psql postgresql://ventros:ventros123@localhost:5432/ventros_crm -t -c "SELECT COUNT(*) FROM sessions WHERE deleted_at IS NULL;" 2>/dev/null | xargs); \
	MESSAGES_COUNT=$$(psql postgresql://ventros:ventros123@localhost:5432/ventros_crm -t -c "SELECT COUNT(*) FROM messages WHERE deleted_at IS NULL;" 2>/dev/null | xargs); \
	EVENTS_COUNT=$$(psql postgresql://ventros:ventros123@localhost:5432/ventros_crm -t -c "SELECT COUNT(*) FROM domain_event_logs WHERE deleted_at IS NULL;" 2>/dev/null | xargs); \
	echo "   📊 Contatos criados: $$CONTACTS_COUNT"; \
	echo "   📊 Sessões criadas: $$SESSIONS_COUNT"; \
	echo "   📊 Mensagens criadas: $$MESSAGES_COUNT"; \
	echo "   📊 Eventos de domínio: $$EVENTS_COUNT"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ MEGA SETUP FINALIZADO COM SUCESSO!"; \
	echo ""; \
	echo "📋 Resumo:"; \
	echo "   👤 User: $$USER_ID"; \
	echo "   📁 Project: $$PROJECT_ID"; \
	echo "   📱 Channel: $$CHANNEL_ID (external_id: $$SESSION_ID)"; \
	echo "   🔗 Webhooks: 3 ativos"; \
	echo "   📨 Mensagens testadas: $$TOTAL (✅ $$SUCCESS | ❌ $$FAILED)"; \
	echo "   💾 Dados no DB: $$CONTACTS_COUNT contacts, $$SESSIONS_COUNT sessions, $$MESSAGES_COUNT messages"; \
	echo "   📊 Eventos publicados: $$EVENTS_COUNT"; \
	echo ""; \
	if [ $$FAILED -gt 0 ]; then \
		echo "⚠️  Alguns testes falharam!"; \
		exit 1; \
	fi; \
	if [ "$$CONTACTS_COUNT" = "0" ] || [ "$$MESSAGES_COUNT" = "0" ]; then \
		echo "⚠️  ATENÇÃO: Nenhum dado foi criado no banco!"; \
		echo "   Verifique os logs da API para ver os erros."; \
		exit 1; \
	fi; \
	echo "🎉 TUDO FUNCIONANDO PERFEITAMENTE!"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

setup-all: ## 🚀 SETUP COMPLETO: Usuário + Projeto + Pipeline + Canal + Webhooks + Teste (TUDO EM UM!)
	@echo "🚀 SETUP COMPLETO - TUDO EM UM COMANDO!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@API_URL="http://localhost:8080"; \
	WAHA_URL="https://waha.ventros.cloud"; \
	WAHA_TOKEN="4bffec302d5f4312b8b73700da3ff3cb"; \
	SESSION_ID="mylomaic-rotate"; \
	N8N_WEBHOOK_1="https://dev.webhook.n8n.ventros.cloud/webhook/6e0918af-876a-4126-b7c2-e1d7d715639e"; \
	N8N_WEBHOOK_2="https://dev.webhook.n8n.ventros.cloud/webhook/ventros/crm/events/all"; \
	TRACKING_WEBHOOK="https://tracking.ventros.cloud/api/events"; \
	echo "🔍 Verificando API..."; \
	curl -f -s $$API_URL/health > /dev/null || (echo "❌ API não está rodando! Execute: make api" && exit 1); \
	echo "✅ API respondendo!"; \
	echo ""; \
	echo "🔍 Verificando WAHA..."; \
	WAHA_STATUS=$$(curl -s "$$WAHA_URL/api/sessions/$$SESSION_ID" -H "X-Api-Key: $$WAHA_TOKEN" | jq -r '.status'); \
	if [ "$$WAHA_STATUS" != "WORKING" ]; then \
		echo "❌ Sessão WAHA não está WORKING (status: $$WAHA_STATUS)"; \
		exit 1; \
	fi; \
	echo "✅ WAHA sessão WORKING!"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "1️⃣ Criando usuário..."; \
	REGISTER_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"waha@ventros.com","password":"waha123","name":"WAHA User"}'); \
	echo "✅ Usuário criado"; \
	echo ""; \
	echo "2️⃣ Fazendo login..."; \
	LOGIN_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"waha@ventros.com","password":"waha123"}'); \
	USER_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.user_id'); \
	PROJECT_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.default_project_id'); \
	if [ "$$USER_ID" = "null" ] || [ -z "$$USER_ID" ]; then \
		echo "❌ Erro ao obter user_id"; \
		exit 1; \
	fi; \
	if [ "$$PROJECT_ID" = "null" ] || [ -z "$$PROJECT_ID" ]; then \
		echo "❌ Erro ao obter project_id"; \
		exit 1; \
	fi; \
	echo "✅ Login OK"; \
	echo "   User ID: $$USER_ID"; \
	echo "   Project ID: $$PROJECT_ID"; \
	echo ""; \
	echo "3️⃣ Verificando pipelines ativos do projeto..."; \
	PIPELINES_RESPONSE=$$(curl -s "$$API_URL/api/v1/pipelines?project_id=$$PROJECT_ID&active=true" \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	PIPELINE_COUNT=$$(echo $$PIPELINES_RESPONSE | jq 'length'); \
	if [ "$$PIPELINE_COUNT" = "0" ] || [ "$$PIPELINE_COUNT" = "null" ]; then \
		echo "❌ ERRO: Projeto não tem pipeline ativo!"; \
		echo "   Sessions PRECISAM de pipeline ativo para funcionar"; \
		exit 1; \
	fi; \
	DEFAULT_PIPELINE_ID=$$(echo $$PIPELINES_RESPONSE | jq -r '.[0].id'); \
	echo "✅ Pipeline ativo encontrado: $$DEFAULT_PIPELINE_ID"; \
	echo ""; \
	echo "4️⃣ Configurando timeout do projeto para 1min..."; \
	PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "UPDATE projects SET session_timeout_minutes = 1 WHERE id = '$$PROJECT_ID';" > /dev/null; \
	echo "✅ Timeout do projeto configurado para 1 minuto"; \
	echo ""; \
	echo "5️⃣ Criando pipeline de teste..."; \
	PIPELINE_RESPONSE=$$(curl -s -X POST "$$API_URL/api/v1/pipelines?project_id=$$PROJECT_ID" \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d '{"name":"Pipeline Teste","description":"Pipeline para testes","color":"#FF6B6B","position":1}'); \
	PIPELINE_ID=$$(echo $$PIPELINE_RESPONSE | jq -r '.pipeline.id'); \
	if [ "$$PIPELINE_ID" = "null" ] || [ -z "$$PIPELINE_ID" ]; then \
		echo "⚠️  Erro ao criar pipeline (não crítico, usando default)"; \
	else \
		echo "✅ Pipeline criado: $$PIPELINE_ID"; \
	fi; \
	echo ""; \
	echo "6️⃣ Criando canal WAHA com timeout de 1 minuto..."; \
	CHANNEL_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/channels \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"WhatsApp Ventros\",\"type\":\"waha\",\"waha_config\":{\"base_url\":\"$$WAHA_URL\",\"api_key\":\"$$WAHA_TOKEN\",\"session_id\":\"$$SESSION_ID\",\"import_strategy\":\"all\"}}"); \
	CHANNEL_ID=$$(echo $$CHANNEL_RESPONSE | jq -r '.id'); \
	WEBHOOK_URL=$$(echo $$CHANNEL_RESPONSE | jq -r '.channel.webhook_url'); \
	if [ "$$CHANNEL_ID" = "null" ] || [ -z "$$CHANNEL_ID" ]; then \
		echo "❌ Erro ao criar canal"; \
		echo $$CHANNEL_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Canal criado: $$CHANNEL_ID"; \
	echo ""; \
	echo "7️⃣ Ativando canal (health check WAHA)..."; \
	ACTIVATE_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/channels/$$CHANNEL_ID/activate-waha \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	if echo "$$ACTIVATE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then \
		echo "❌ Erro ao ativar canal"; \
		echo $$ACTIVATE_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Canal ativado!"; \
	echo ""; \
	echo "8️⃣ Criando webhook n8n #1 (eventos de domínio)..."; \
	WEBHOOK1_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"n8n Domain Events 1\",\"url\":\"$$N8N_WEBHOOK_1\",\"events\":[\"contact.created\",\"contact.updated\",\"session.created\",\"session.closed\",\"message.received\",\"message.sent\"],\"active\":true}"); \
	WEBHOOK1_ID=$$(echo $$WEBHOOK1_RESPONSE | jq -r '.id'); \
	if [ "$$WEBHOOK1_ID" != "null" ] && [ -n "$$WEBHOOK1_ID" ]; then \
		echo "✅ Webhook 1 criado: $$WEBHOOK1_ID"; \
	else \
		echo "⚠️  Webhook 1 já existe"; \
	fi; \
	echo ""; \
	echo "9️⃣ Criando webhook n8n #2 (eventos de domínio)..."; \
	WEBHOOK2_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"n8n Domain Events 2\",\"url\":\"$$N8N_WEBHOOK_2\",\"events\":[\"contact.created\",\"contact.updated\",\"session.created\",\"session.closed\",\"message.received\",\"message.sent\"],\"active\":true}"); \
	WEBHOOK2_ID=$$(echo $$WEBHOOK2_RESPONSE | jq -r '.id'); \
	if [ "$$WEBHOOK2_ID" != "null" ] && [ -n "$$WEBHOOK2_ID" ]; then \
		echo "✅ Webhook 2 criado: $$WEBHOOK2_ID"; \
	else \
		echo "⚠️  Webhook 2 já existe"; \
	fi; \
	echo ""; \
	echo "🔟 Criando webhook de tracking (analytics)..."; \
	TRACKING_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"Event Tracking Analytics\",\"url\":\"$$TRACKING_WEBHOOK\",\"events\":[\"contact.created\",\"session.created\",\"session.closed\",\"message.received\",\"message.sent\",\"pipeline.status.changed\",\"tracking.ad_conversion\"],\"active\":true}"); \
	TRACKING_ID=$$(echo $$TRACKING_RESPONSE | jq -r '.id'); \
	if [ "$$TRACKING_ID" != "null" ] && [ -n "$$TRACKING_ID" ]; then \
		echo "✅ Tracking webhook criado: $$TRACKING_ID"; \
	else \
		echo "⚠️  Tracking webhook já existe"; \
	fi; \
	echo ""; \
	echo "1️⃣1️⃣ Enviando mensagem de teste WAHA..."; \
	TEST_PHONE="5511999999999"; \
	TEST_MESSAGE="🎉 Sistema configurado com sucesso!\n\n✅ Canal WAHA ativo\n✅ Pipeline com timeout 30min\n✅ 3 webhooks configurados\n\nTudo pronto para uso!"; \
	TEST_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/test/send-waha-message \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"channel_id\":\"$$CHANNEL_ID\",\"phone\":\"$$TEST_PHONE\",\"message\":\"$$TEST_MESSAGE\"}"); \
	TEST_SUCCESS=$$(echo $$TEST_RESPONSE | jq -r '.success // false'); \
	if [ "$$TEST_SUCCESS" = "true" ]; then \
		echo "✅ Mensagem de teste enviada para $$TEST_PHONE!"; \
	else \
		echo "⚠️  Erro ao enviar mensagem de teste (não crítico)"; \
	fi; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ SETUP COMPLETO FINALIZADO!"; \
	echo ""; \
	echo "📋 Resumo da Configuração:"; \
	echo ""; \
	echo "👤 Usuário:"; \
	echo "   Email: waha@ventros.com"; \
	echo "   Senha: waha123"; \
	echo "   ID: $$USER_ID"; \
	echo ""; \
	echo "📁 Projeto:"; \
	echo "   ID: $$PROJECT_ID"; \
	echo "   Pipelines ativos: 2"; \
	echo "     • Pipeline Default (timeout: $${DEFAULT_TIMEOUT}min) - position 0"; \
	echo "     • Pipeline Teste (timeout: 1min) - position 1"; \
	echo ""; \
	echo "⚠️  IMPORTANTE - Relação Pipeline ↔ Session:"; \
	echo "   • Session pode ser criada COM ou SEM Pipeline"; \
	echo "   • Session usa timeout configurado NO PROJETO (default: 30min)"; \
	echo "   • Timeout é configurável nas settings do CRM"; \
	echo ""; \
	echo "📱 Canal WAHA:"; \
	echo "   ID: $$CHANNEL_ID"; \
	echo "   Status: ACTIVE"; \
	echo "   Session: $$SESSION_ID"; \
	echo "   Webhook: $$WEBHOOK_URL"; \
	echo ""; \
	echo "🔗 Webhooks (3 ativos):"; \
	echo "   • n8n Domain Events 1"; \
	echo "   • n8n Domain Events 2"; \
	echo "   • Event Tracking Analytics"; \
	echo ""; \
	echo "📊 Eventos monitorados:"; \
	echo "   • contact.created"; \
	echo "   • session.created / session.closed"; \
	echo "   • message.received / message.sent"; \
	echo "   • pipeline.status.changed"; \
	echo "   • tracking.ad_conversion"; \
	echo ""; \
	echo "💡 Sistema pronto para uso!"; \
	echo "   Envie mensagens no WhatsApp e veja os eventos nos webhooks n8n"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "🧪 Rodando testes de todos os tipos de mensagens..."; \
	echo ""; \
	$(MAKE) test-all-message-types || echo "⚠️  Alguns testes falharam, mas o setup está completo"

setup-webhooks-complete: ## Configura webhooks n8n + tracking de eventos (tudo em um!)
	@echo "🔗 Setup COMPLETO - Webhooks n8n + Event Tracking"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@API_URL="http://localhost:8080"; \
	N8N_WEBHOOK_1="https://dev.webhook.n8n.ventros.cloud/webhook/6e0918af-876a-4126-b7c2-e1d7d715639e"; \
	N8N_WEBHOOK_2="https://dev.webhook.n8n.ventros.cloud/webhook/ventros/crm/events/all"; \
	TRACKING_WEBHOOK="https://tracking.ventros.cloud/api/events"; \
	echo "🔍 Verificando API..."; \
	curl -f -s $$API_URL/health > /dev/null || (echo "❌ API não está rodando! Execute: make api" && exit 1); \
	echo "✅ API respondendo!"; \
	echo ""; \
	echo "1️⃣ Fazendo login..."; \
	LOGIN_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"waha@ventros.com","password":"waha123"}'); \
	USER_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.user_id'); \
	PROJECT_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.default_project_id'); \
	if [ "$$USER_ID" = "null" ] || [ -z "$$USER_ID" ]; then \
		echo "❌ Usuário não encontrado. Execute: make setup-waha-complete"; \
		exit 1; \
	fi; \
	echo "✅ Login OK"; \
	echo "   User ID: $$USER_ID"; \
	echo "   Project ID: $$PROJECT_ID"; \
	echo ""; \
	echo "2️⃣ Criando webhook n8n #1 (eventos de domínio)..."; \
	WEBHOOK1_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"n8n Domain Events 1\",\"url\":\"$$N8N_WEBHOOK_1\",\"events\":[\"contact.created\",\"contact.updated\",\"session.created\",\"session.closed\",\"message.received\",\"message.sent\"],\"active\":true}"); \
	WEBHOOK1_ID=$$(echo $$WEBHOOK1_RESPONSE | jq -r '.id'); \
	if [ "$$WEBHOOK1_ID" = "null" ] || [ -z "$$WEBHOOK1_ID" ]; then \
		echo "⚠️  Webhook 1 já existe ou erro"; \
		WEBHOOK1_ID=$$(echo $$WEBHOOK1_RESPONSE | jq -r '.id // "existing"'); \
	else \
		echo "✅ Webhook 1 criado: $$WEBHOOK1_ID"; \
	fi; \
	echo ""; \
	echo "3️⃣ Criando webhook n8n #2 (eventos de domínio)..."; \
	WEBHOOK2_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"n8n Domain Events 2\",\"url\":\"$$N8N_WEBHOOK_2\",\"events\":[\"contact.created\",\"contact.updated\",\"session.created\",\"session.closed\",\"message.received\",\"message.sent\"],\"active\":true}"); \
	WEBHOOK2_ID=$$(echo $$WEBHOOK2_RESPONSE | jq -r '.id'); \
	if [ "$$WEBHOOK2_ID" = "null" ] || [ -z "$$WEBHOOK2_ID" ]; then \
		echo "⚠️  Webhook 2 já existe ou erro"; \
		WEBHOOK2_ID=$$(echo $$WEBHOOK2_RESPONSE | jq -r '.id // "existing"'); \
	else \
		echo "✅ Webhook 2 criado: $$WEBHOOK2_ID"; \
	fi; \
	echo ""; \
	echo "4️⃣ Criando webhook de tracking (analytics)..."; \
	TRACKING_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d "{\"name\":\"Event Tracking Analytics\",\"url\":\"$$TRACKING_WEBHOOK\",\"events\":[\"contact.created\",\"session.created\",\"session.closed\",\"message.received\",\"message.sent\",\"pipeline.status.changed\",\"tracking.ad_conversion\"],\"active\":true}"); \
	TRACKING_ID=$$(echo $$TRACKING_RESPONSE | jq -r '.id'); \
	if [ "$$TRACKING_ID" = "null" ] || [ -z "$$TRACKING_ID" ]; then \
		echo "⚠️  Tracking webhook já existe ou erro"; \
		TRACKING_ID=$$(echo $$TRACKING_RESPONSE | jq -r '.id // "existing"'); \
	else \
		echo "✅ Tracking webhook criado: $$TRACKING_ID"; \
	fi; \
	echo ""; \
	echo "5️⃣ Listando todos os webhooks ativos..."; \
	WEBHOOKS_LIST=$$(curl -s $$API_URL/api/v1/webhook-subscriptions \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	echo "📋 Webhooks configurados:"; \
	WEBHOOK_COUNT=$$(echo $$WEBHOOKS_LIST | jq -r 'if type == "array" then length else 0 end'); \
	if [ "$$WEBHOOK_COUNT" -gt 0 ]; then \
		echo $$WEBHOOKS_LIST | jq -r '.[] | "   ✓ \(.name) - \(.url)"'; \
		echo ""; \
		echo "   Total: $$WEBHOOK_COUNT webhooks ativos"; \
	else \
		echo "   (nenhum webhook configurado)"; \
	fi; \
	echo ""; \
	echo "6️⃣ Buscando canal WAHA ativo..."; \
	CHANNELS_RESPONSE=$$(curl -s $$API_URL/api/v1/channels \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID"); \
	CHANNEL_ID=$$(echo $$CHANNELS_RESPONSE | jq -r '.[0].id // empty'); \
	if [ -z "$$CHANNEL_ID" ]; then \
		echo "⚠️  Nenhum canal encontrado - pule o teste"; \
	else \
		echo "✅ Canal encontrado: $$CHANNEL_ID"; \
		echo ""; \
		echo "7️⃣ Enviando mensagem de teste WAHA..."; \
		TEST_PHONE="5511999999999"; \
		TEST_MESSAGE="🎉 Teste de webhooks! Sistema configurado com:\n• 2 webhooks n8n\n• 1 webhook tracking\n• Total: $$WEBHOOK_COUNT webhooks ativos"; \
		TEST_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/test/send-waha-message \
			-H "X-Dev-User-ID: $$USER_ID" \
			-H "X-Dev-Project-ID: $$PROJECT_ID" \
			-H "Content-Type: application/json" \
			-d "{\"channel_id\":\"$$CHANNEL_ID\",\"phone\":\"$$TEST_PHONE\",\"message\":\"$$TEST_MESSAGE\"}"); \
		TEST_SUCCESS=$$(echo $$TEST_RESPONSE | jq -r '.success // false'); \
		if [ "$$TEST_SUCCESS" = "true" ]; then \
			echo "✅ Mensagem de teste enviada!"; \
			echo "   📱 Para: $$TEST_PHONE"; \
			echo "   📨 Verifique os webhooks n8n para ver os eventos!"; \
		else \
			echo "⚠️  Erro ao enviar mensagem de teste"; \
			echo $$TEST_RESPONSE | jq .; \
		fi; \
	fi; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ Setup COMPLETO finalizado!"; \
	echo ""; \
	echo "📝 Webhooks configurados:"; \
	echo "   🔹 n8n Domain Events 1 → $$N8N_WEBHOOK_1"; \
	echo "   🔹 n8n Domain Events 2 → $$N8N_WEBHOOK_2"; \
	echo "   🔹 Event Tracking → $$TRACKING_WEBHOOK"; \
	echo ""; \
	echo "📊 Eventos monitorados:"; \
	echo "   • contact.created"; \
	echo "   • session.created / session.closed"; \
	echo "   • message.received / message.sent"; \
	echo "   • pipeline.status.changed"; \
	echo "   • tracking.ad_conversion"; \
	echo ""; \
	echo "🧪 Teste automático executado!"; \
	echo "   Verifique os logs da API e os webhooks n8n"; \
	echo "   para confirmar que os eventos foram enviados"; \
	echo ""; \
	echo "💡 Para testar manualmente, envie mensagens para o WhatsApp!"; \
	echo "   Todos os eventos serão capturados e enviados"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

setup-project-pipeline: ## Cria projeto e pipeline do zero com timeout de 1min (requer API rodando)
	@echo "🏗️  Setup Projeto + Pipeline"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@API_URL="http://localhost:8080"; \
	echo "🔍 Verificando API..."; \
	curl -f -s $$API_URL/health > /dev/null || (echo "❌ API não está rodando! Execute: make api" && exit 1); \
	echo "✅ API respondendo!"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "1️⃣ Criando usuário..."; \
	REGISTER_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"pipeline@ventros.com","password":"pipeline123","name":"Pipeline User"}'); \
	echo "✅ Usuário criado"; \
	echo ""; \
	echo "2️⃣ Fazendo login..."; \
	LOGIN_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"pipeline@ventros.com","password":"pipeline123"}'); \
	USER_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.user_id'); \
	CUSTOMER_ID=$$(echo $$LOGIN_RESPONSE | jq -r '.user_id'); \
	if [ "$$USER_ID" = "null" ] || [ -z "$$USER_ID" ]; then \
		echo "❌ Erro ao obter user_id"; \
		echo $$LOGIN_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Login OK - User ID: $$USER_ID"; \
	echo ""; \
	echo "3️⃣ Criando projeto..."; \
	PROJECT_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/projects \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "Content-Type: application/json" \
		-d '{"name":"Projeto WAHA","description":"Projeto para testes WAHA"}'); \
	PROJECT_ID=$$(echo $$PROJECT_RESPONSE | jq -r '.project.id'); \
	if [ "$$PROJECT_ID" = "null" ] || [ -z "$$PROJECT_ID" ]; then \
		echo "❌ Erro ao criar projeto"; \
		echo $$PROJECT_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Projeto criado: $$PROJECT_ID"; \
	echo ""; \
	echo "4️⃣ Configurando timeout do projeto para 1 minuto..."; \
	PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "UPDATE projects SET session_timeout_minutes = 1 WHERE id = '$$PROJECT_ID';" > /dev/null; \
	echo "✅ Timeout do projeto configurado para 1 minuto"; \
	echo ""; \
	echo "5️⃣ Criando pipeline..."; \
	PIPELINE_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/pipelines \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d '{"name":"Pipeline WAHA","description":"Pipeline para atendimentos WAHA"}'); \
	PIPELINE_ID=$$(echo $$PIPELINE_RESPONSE | jq -r '.pipeline.id'); \
	if [ "$$PIPELINE_ID" = "null" ] || [ -z "$$PIPELINE_ID" ]; then \
		echo "❌ Erro ao criar pipeline"; \
		echo $$PIPELINE_RESPONSE | jq .; \
		exit 1; \
	fi; \
	echo "✅ Pipeline criado: $$PIPELINE_ID"; \
	echo ""; \
	echo "5️⃣ Criando canal WAHA..."; \
	CHANNEL_RESPONSE=$$(curl -s -X POST $$API_URL/api/v1/channels \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d '{"name":"Novo","type":"open","color":"#3B82F6","order":1}' | jq -r '.status.id'); \
	STATUS_ATENDIMENTO=$$(curl -s -X POST $$API_URL/api/v1/pipelines/$$PIPELINE_ID/statuses \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d '{"name":"Em Atendimento","type":"open","color":"#F59E0B","order":2}' | jq -r '.status.id'); \
	STATUS_RESOLVIDO=$$(curl -s -X POST $$API_URL/api/v1/pipelines/$$PIPELINE_ID/statuses \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" \
		-H "Content-Type: application/json" \
		-d '{"name":"Resolvido","type":"won","color":"#10B981","order":3}' | jq -r '.status.id'); \
	echo "✅ Status criados:"; \
	echo "   • Novo ($$STATUS_NEW)"; \
	echo "   • Em Atendimento ($$STATUS_ATENDIMENTO)"; \
	echo "   • Resolvido ($$STATUS_RESOLVIDO)"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ Projeto e Pipeline configurados!"; \
	echo ""; \
	echo "📋 Informações:"; \
	echo "   User ID: $$USER_ID"; \
	echo "   Project ID: $$PROJECT_ID"; \
	echo "   Pipeline ID: $$PIPELINE_ID"; \
	echo "   Project Session Timeout: 1 minuto"; \
	echo ""; \
	echo "🔑 Credenciais:"; \
	echo "   Email: pipeline@ventros.com"; \
	echo "   Senha: pipeline123"; \
	echo ""; \
	echo "💡 Use esses IDs para criar o canal WAHA:"; \
	echo "   export USER_ID=\"$$USER_ID\""; \
	echo "   export PROJECT_ID=\"$$PROJECT_ID\""; \
	echo "   export PIPELINE_ID=\"$$PIPELINE_ID\""; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test-waha-health: ## Testa conexão e saúde da sessão WAHA
	@echo "🧪 Testando WAHA - waha.ventros.cloud"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@WAHA_URL="https://waha.ventros.cloud"; \
	WAHA_TOKEN="4bffec302d5f4312b8b73700da3ff3cb"; \
	SESSION_ID="mylomaic-rotate"; \
	echo "🌐 WAHA URL: $$WAHA_URL"; \
	echo "🔑 Token: $$WAHA_TOKEN"; \
	echo "📱 Session ID: $$SESSION_ID"; \
	echo ""; \
	echo "1️⃣ Testando conexão com WAHA..."; \
	curl -f -s $$WAHA_URL/health > /dev/null 2>&1 && echo "✅ WAHA respondendo!" || echo "❌ WAHA não responde"; \
	echo ""; \
	echo "2️⃣ Verificando sessão $$SESSION_ID..."; \
	SESSION_RESPONSE=$$(curl -s "$$WAHA_URL/api/sessions/$$SESSION_ID" \
		-H "X-Api-Key: $$WAHA_TOKEN"); \
	if echo "$$SESSION_RESPONSE" | jq -e '.name' > /dev/null 2>&1; then \
		STATUS=$$(echo "$$SESSION_RESPONSE" | jq -r '.status'); \
		echo "📊 Status da sessão: $$STATUS"; \
		echo ""; \
		if [ "$$STATUS" = "WORKING" ]; then \
			echo "✅ Sessão está WORKING - pronta para usar!"; \
			echo ""; \
			echo "📋 Detalhes:"; \
			echo "$$SESSION_RESPONSE" | jq '{name, status, config: {webhooks: .config.webhooks}}'; \
		elif [ "$$STATUS" = "SCAN_QR_CODE" ]; then \
			echo "⚠️  Sessão aguardando QR Code"; \
			echo ""; \
			echo "📱 Obter QR Code:"; \
			echo "   curl \"$$WAHA_URL/api/sessions/$$SESSION_ID/auth/qr\" \\"; \
			echo "     -H \"X-Api-Key: $$WAHA_TOKEN\""; \
		else \
			echo "⚠️  Sessão com status: $$STATUS"; \
		fi; \
	else \
		echo "❌ Sessão não encontrada ou erro na API"; \
		echo ""; \
		echo "Resposta:"; \
		echo "$$SESSION_RESPONSE" | jq . 2>/dev/null || echo "$$SESSION_RESPONSE"; \
	fi; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if echo "$$SESSION_RESPONSE" | jq -e '.status' > /dev/null 2>&1 && [ "$$(echo "$$SESSION_RESPONSE" | jq -r '.status')" = "WORKING" ]; then \
		echo "✅ WAHA está pronto para integração!"; \
		echo ""; \
		echo "🚀 Próximos passos:"; \
		echo "   1. Rode a API: make api"; \
		echo "   2. Crie usuário e obtenha token"; \
		echo "   3. Crie canal WAHA com esses dados:"; \
		echo "      {"; \
		echo "        \"base_url\": \"$$WAHA_URL\","; \
		echo "        \"api_key\": \"$$WAHA_TOKEN\","; \
		echo "        \"session_id\": \"$$SESSION_ID\","; \
		echo "        \"import_strategy\": \"all\""; \
		echo "      }"; \
	else \
		echo "⚠️  Sessão não está WORKING"; \
		echo ""; \
		echo "💡 Verifique o status e conecte o WhatsApp se necessário"; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

logs-infra: ## Logs da infraestrutura
	@$(MAKE) dev-logs

logs-api: ## Logs da API (se containerizada)
	@$(MAKE) container-logs

ps: ## Lista containers rodando
	@$(CONTAINER_RUNTIME) ps --filter "name=ventros"

##@ 🔧 Setup Inicial

setup: ## Setup inicial completo (primeira vez)
	@echo "🎬 Setup Inicial do Ventros CRM"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "1. Verificando dependências..."
	@command -v go >/dev/null 2>&1 || (echo "❌ Go não instalado" && exit 1)
	@command -v $(CONTAINER_RUNTIME) >/dev/null 2>&1 || (echo "❌ $(CONTAINER_RUNTIME) não instalado" && exit 1)
	@command -v swag >/dev/null 2>&1 || (echo "⚠️  Swagger não instalado. Instalando..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@echo "✅ Dependências OK"
	@echo ""
	@echo "2. Criando .env..."
	@if [ ! -f .env ]; then cp .env.example .env && echo "✅ .env criado"; else echo "✅ .env já existe"; fi
	@echo ""
	@echo "3. Baixando dependências Go..."
	@go mod download
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Setup completo!"
	@echo ""
	@echo "🚀 Próximos passos:"
	@echo "   make dev       # Sobe infra + API (modo desenvolvimento)"
	@echo "   make container # Sobe tudo containerizado"
	@echo "   make k8s       # Deploy no Minikube"
	@echo ""
	@echo "📚 Ajuda completa: make help"

##@ 🔄 Atalhos Rápidos

restart-infra: infra-stop infra ## Reinicia infraestrutura (mantém dados)

restart-dev: infra-stop dev ## Reinicia modo dev completo

restart-container: container-stop container ## Reinicia container

restart-k8s: k8s-delete k8s ## Reinicia K8s

fresh-start: infra-reset ## Alias para infra-reset (limpa tudo e começa do zero)

.DEFAULT_GOAL := help

delete-all-webhooks: ## 🗑️ Deleta todos os webhooks cadastrados
	@echo "🗑️ Deletando todos os webhooks..."
	@API_URL="http://localhost:8080"; \
	USER_ID="eb9edbc8-0712-44b7-ba14-2ec10680ff95"; \
	PROJECT_ID="fb72d2bb-80a9-4834-ba5d-2a8d8402e120"; \
	WEBHOOKS=$$(curl -s "$$API_URL/api/v1/webhook-subscriptions" \
		-H "X-Dev-User-ID: $$USER_ID" \
		-H "X-Dev-Project-ID: $$PROJECT_ID" | jq -r '.[] | .id'); \
	COUNT=0; \
	for webhook_id in $$WEBHOOKS; do \
		curl -s -X DELETE "$$API_URL/api/v1/webhook-subscriptions/$$webhook_id" \
			-H "X-Dev-User-ID: $$USER_ID" \
			-H "X-Dev-Project-ID: $$PROJECT_ID" > /dev/null; \
		echo "   ✅ Deletado: $$webhook_id"; \
		COUNT=$$((COUNT + 1)); \
	done; \
	echo ""; \
	echo "✅ Total deletado: $$COUNT webhooks"
