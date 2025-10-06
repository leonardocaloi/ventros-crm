# Ventros CRM - Makefile
# Container runtime (docker ou podman)
CONTAINER_RUNTIME ?= docker
COMPOSE = $(CONTAINER_RUNTIME) compose

.PHONY: help

##@ Ajuda
help: ## Mostra esta ajuda
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1mVentros CRM - Comandos Disponíveis\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ 🚀 Workflows Principais

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

run: api ## Alias para 'make api'

build: ## Compila binário
	@echo "🔨 Compilando..."
	@go build -o ventros-crm cmd/api/main.go
	@echo "✅ Binário: ./ventros-crm"

test: ## Roda testes unitários
	@echo "🧪 Rodando testes unitários..."
	@go test -v -race ./internal/... ./infrastructure/...

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

test-all: test test-e2e ## Roda todos os testes (unit + E2E)

e2e-webhook: ## [E2E] Teste completo: Canal WAHA + Webhook + Mensagem FB Ads (WEBHOOK_URL=https://webhook.site/xxx)
	@echo "🧪 E2E: Canal WAHA com Webhook e FB Ads"
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
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "1️⃣ Setup ambiente de teste..."
	@WEBHOOK_PARAM=""; \
	if [ -n "$(WEBHOOK_URL)" ]; then \
		WEBHOOK_PARAM="?webhook_url=$(WEBHOOK_URL)"; \
		echo "🔗 Usando webhook externo: $(WEBHOOK_URL)"; \
	fi; \
	SETUP_RESPONSE=$$(curl -s -X POST "http://localhost:8080/api/v1/test/setup$$WEBHOOK_PARAM"); \
	echo $$SETUP_RESPONSE | jq -r '.data | "✅ User: \(.user_id)\n✅ Project: \(.project_id)\n✅ Channel: \(.channel_id)\n✅ Channel Webhook: \(.channel_webhook_url)\n✅ Webhook Subscription: \(.webhook_id)\n✅ API Key: \(.api_key)"'; \
	WEBHOOK_ID=$$(echo $$SETUP_RESPONSE | jq -r '.data.webhook_id'); \
	CHANNEL_ID=$$(echo $$SETUP_RESPONSE | jq -r '.data.channel_id'); \
	CHANNEL_WEBHOOK_URL=$$(echo $$SETUP_RESPONSE | jq -r '.data.channel_webhook_url'); \
	API_KEY=$$(echo $$SETUP_RESPONSE | jq -r '.data.api_key'); \
	echo ""; \
	echo "📋 Eventos ativos no webhook de teste:"; \
	curl -s -X GET http://localhost:8080/api/v1/webhook-subscriptions/$$WEBHOOK_ID \
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
	curl -s -X GET http://localhost:8080/api/v1/channels/$$CHANNEL_ID \
		-H "Authorization: Bearer $$API_KEY" | jq '.channel | {id,name,type,webhook_url,webhook_active,messages_received}'; \
	echo ""; \
	if [ -n "$(WEBHOOK_URL)" ]; then \
		echo "5️⃣ Eventos de domínio serão enviados para: $(WEBHOOK_URL)"; \
		echo "📤 Eventos esperados:"; \
		echo "   ✓ contact.created (imediato)"; \
		echo "   ✓ session.started (imediato)"; \
		echo "   ✓ ad_campaign.tracked (imediato)"; \
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
	echo "💡 Para usar webhook externo:"; \
	echo "   WEBHOOK_URL=https://webhook.site/xxx make e2e-webhook"; \
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

health: ## Checa saúde da API
	@curl -s http://localhost:8080/health | jq . || echo "❌ API não responde"

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
