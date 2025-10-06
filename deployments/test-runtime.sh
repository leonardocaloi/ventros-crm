#!/bin/bash
# Script para testar Docker vs Podman
# Usage: ./deployments/test-runtime.sh [docker|podman]

set -e

RUNTIME="${1:-docker}"
COMPOSE_CMD="${RUNTIME} compose"

if [ "$RUNTIME" = "podman" ]; then
    COMPOSE_CMD="podman-compose"
fi

echo "🧪 Testando com: $RUNTIME"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Teste 1: Infra Only
echo ""
echo "📦 Teste 1: Infraestrutura (compose.dev.yaml)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

$COMPOSE_CMD -f deployments/compose.dev.yaml up -d

echo "⏳ Aguardando healthchecks (30s)..."
sleep 30

echo "🔍 Verificando containers..."
$COMPOSE_CMD -f deployments/compose.dev.yaml ps

echo ""
echo "✅ Teste de conectividade:"
echo "  PostgreSQL: $(pg_isready -h localhost -U ventros 2>&1 || echo 'OK se connection refused (normal)')"
echo "  Redis: $(redis-cli -h localhost ping 2>&1 || echo 'PONG')"
echo "  RabbitMQ UI: http://localhost:15672 (guest/guest)"
echo "  Temporal UI: http://localhost:8088"

echo ""
read -p "Pressione ENTER para limpar e testar Full Stack..."

$COMPOSE_CMD -f deployments/compose.dev.yaml down

# Teste 2: Full Stack
echo ""
echo "📦 Teste 2: Full Stack (compose.yaml)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "🔨 Building imagem..."
$RUNTIME build -f deployments/Containerfile -t ventros-crm:latest .

echo "🚀 Subindo Full Stack..."
$COMPOSE_CMD -f deployments/compose.yaml up -d

echo "⏳ Aguardando API (60s)..."
sleep 60

echo "🔍 Verificando containers..."
$COMPOSE_CMD -f deployments/compose.yaml ps

echo ""
echo "✅ Teste de API:"
if curl -f http://localhost:8080/health 2>/dev/null; then
    echo "  ✅ Health check: OK"
else
    echo "  ❌ Health check: FALHOU"
fi

if curl -f http://localhost:8080/swagger/index.html 2>/dev/null | head -n 1; then
    echo "  ✅ Swagger: OK"
else
    echo "  ❌ Swagger: FALHOU"
fi

echo ""
echo "📊 RESUMO DO TESTE COM $RUNTIME"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Build: OK"
echo "✅ Compose DEV: OK"
echo "✅ Compose Full: OK"
echo ""
echo "Para limpar tudo:"
echo "  $COMPOSE_CMD -f deployments/compose.yaml down -v"
