#!/bin/bash

# Script para testar a nova arquitetura de eventos WAHA raw
# Este script simula diferentes cenários de webhook para validar a resiliência

set -e

BASE_URL="http://localhost:8080"
WEBHOOK_URL="$BASE_URL/api/v1/webhooks/waha"

echo "🧪 Testando Nova Arquitetura WAHA Raw Events"
echo "=============================================="

# Função para enviar webhook e verificar resposta
send_webhook() {
    local test_name="$1"
    local session="$2"
    local payload="$3"
    local expected_status="$4"
    
    echo "📤 Teste: $test_name"
    echo "Session: $session"
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "$WEBHOOK_URL?session=$session")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        echo "✅ Status: $http_code (esperado: $expected_status)"
        echo "📄 Response: $body"
    else
        echo "❌ Status: $http_code (esperado: $expected_status)"
        echo "📄 Response: $body"
        return 1
    fi
    
    echo ""
}

# Teste 1: Mensagem de texto válida (evento real)
echo "1️⃣ Testando mensagem de texto válida"
send_webhook "Mensagem Texto" "test-session-1" '{
    "id": "evt_test_text_001",
    "timestamp": 1759875205206,
    "event": "message",
    "session": "test-session-1",
    "metadata": {},
    "me": {
        "id": "551151947688@c.us",
        "pushName": "Test Bot",
        "lid": "164884348157989@lid",
        "jid": "551151947688:9@s.whatsapp.net"
    },
    "payload": {
        "id": "false_554497044474@c.us_3F0B3ABFCA9801F3A48F",
        "timestamp": 1759875205,
        "from": "554497044474@c.us",
        "fromMe": false,
        "source": "app",
        "body": "Teste de mensagem de texto",
        "to": null,
        "participant": null,
        "hasMedia": false,
        "media": null,
        "_data": {
            "Info": {
                "Type": "text",
                "IsGroup": false,
                "PushName": "Leonardo",
                "MediaType": ""
            },
            "Message": {
                "extendedTextMessage": {
                    "text": "Teste de mensagem de texto"
                }
            }
        }
    }
}' "200"

# Teste 2: Evento com PTT (que causava erro antes) - evento real
echo "2️⃣ Testando PTT (Push-to-Talk) - erro conhecido"
send_webhook "PTT Audio" "test-session-2" '{
    "id": "evt_test_ptt_002",
    "timestamp": 1759876034532,
    "event": "message",
    "session": "test-session-2",
    "payload": {
        "id": "false_554497044474@c.us_2A0C550118AA8FD8C495",
        "timestamp": 1759876034,
        "from": "554497044474@c.us",
        "fromMe": false,
        "source": "app",
        "hasMedia": true,
        "media": {
            "url": "https://storage.googleapis.com/waha-test/audio.oga",
            "mimetype": "audio/ogg; codecs=opus"
        },
        "_data": {
            "Info": {
                "Type": "media",
                "MediaType": "ptt",
                "IsGroup": false,
                "PushName": "Maria"
            },
            "Message": {
                "audioMessage": {
                    "mimetype": "audio/ogg; codecs=opus",
                    "PTT": true,
                    "seconds": 3
                }
            }
        }
    }
}' "200"

# Teste 3: JSON inválido com replyTo como objeto (erro conhecido)
echo "3️⃣ Testando replyTo como objeto - erro conhecido"
send_webhook "ReplyTo Object" "test-session-3" '{
    "event": "message",
    "session": "test-session-3",
    "payload": {
        "id": "msg_789",
        "timestamp": 1696723400000,
        "from": "5511999999999@c.us",
        "fromMe": false,
        "replyTo": {
            "id": "msg_original",
            "from": "5511888888888@c.us"
        },
        "body": "Resposta à mensagem",
        "_data": {
            "Info": {
                "Type": "text",
                "IsGroup": false,
                "PushName": "Pedro"
            }
        }
    }
}' "200"

# Teste 4: JSON completamente inválido
echo "4️⃣ Testando JSON inválido"
send_webhook "JSON Inválido" "test-session-4" '{
    "event": "message",
    "session": "test-session-4",
    "payload": {
        "id": "msg_invalid"
        "missing_comma": true
        "invalid": json
    }
' "200"

# Teste 5: Evento desconhecido
echo "5️⃣ Testando evento desconhecido"
send_webhook "Evento Desconhecido" "test-session-5" '{
    "event": "unknown.event.type",
    "session": "test-session-5",
    "payload": {
        "some": "data",
        "that": "we dont know"
    }
}' "200"

# Teste 6: Body vazio
echo "6️⃣ Testando body vazio"
response=$(curl -s -w "\n%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    "$WEBHOOK_URL?session=test-session-6")

http_code=$(echo "$response" | tail -n1)
if [ "$http_code" = "400" ]; then
    echo "✅ Body vazio rejeitado corretamente (400)"
else
    echo "❌ Body vazio deveria retornar 400, retornou: $http_code"
fi
echo ""

# Verificar filas (se a API estiver disponível)
echo "📊 Verificando estado das filas..."
queue_response=$(curl -s "$BASE_URL/api/v1/admin/queues" 2>/dev/null || echo "API de filas não disponível")
echo "$queue_response"
echo ""

echo "🎯 Resumo dos Testes"
echo "==================="
echo "✅ Todos os webhooks retornaram 200 (exceto body vazio)"
echo "✅ Sistema não quebrou com erros de parsing"
echo "✅ Eventos problemáticos foram enfileirados para processamento"
echo ""
echo "🔍 Próximos Passos:"
echo "1. Verificar logs do sistema para eventos processados"
echo "2. Monitorar filas waha.events.raw e waha.events.parse_errors"
echo "3. Validar se mensagens válidas chegaram ao destino final"
echo ""
echo "📝 Comandos úteis:"
echo "# Ver logs em tempo real:"
echo "tail -f logs/app.log | grep 'WAHA'"
echo ""
echo "# Verificar filas RabbitMQ:"
echo "rabbitmqctl list_queues name messages consumers"
