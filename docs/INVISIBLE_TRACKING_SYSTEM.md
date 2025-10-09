# Sistema de Tracking Invisível com Codificação Ternária

## 📋 Visão Geral

Sistema de rastreamento de mensagens WhatsApp utilizando códigos invisíveis baseados em codificação ternária (base 3). O sistema insere 7 caracteres invisíveis após o primeiro caractere de uma mensagem para rastrear a origem e permitir análise de conversão.

## 🎯 Objetivo

Permitir rastreamento preciso de campanhas de marketing sem alterar visualmente a mensagem enviada ao usuário, mantendo a experiência natural enquanto captura métricas de conversão.

## 🔧 Arquitetura

### Componentes Principais

1. **TernaryEncoder** (`internal/domain/tracking/ternary_encoder.go`)
   - Codificação/decodificação de IDs em base 3 → caracteres invisíveis
   - Recuperação de caracteres corrompidos pelo WhatsApp
   - Análise detalhada de códigos

2. **Encode/Decode Use Cases** (`internal/application/tracking/encode_decode_tracking_usecase.go`)
   - Lógica de negócio para codificar tracking IDs
   - Decodificação automática com análise de confiança
   - Geração de links WhatsApp com código embutido

3. **HTTP Handler** (`infrastructure/http/handlers/tracking_handler.go`)
   - Endpoints REST para encode/decode
   - Integração com sistema de autenticação

4. **Auto-detection** (`internal/application/message/process_inbound_message.go`)
   - Detecção automática de códigos invisíveis em mensagens recebidas
   - Associação automática de tracking ao contato e sessão

## 📊 Fluxo de Funcionamento

### 1. Criação de Tracking

```
Cliente cria tracking → Recebe ID numérico → Usa para codificar mensagens
```

### 2. Codificação de Mensagem

```
ID decimal (ex: 123)
  ↓
Conversão para ternário (ex: 0011120)
  ↓
Codificação em caracteres invisíveis:
  0 → U+200B (Zero Width Space)
  1 → U+2060 (Word Joiner)
  2 → U+FEFF (Zero Width No-Break Space)
  ↓
Inserção após 1º caractere:
"Olá" → "O[invisível×7]lá"
  ↓
Link WhatsApp gerado
```

### 3. Envio e Detecção

```
Mensagem enviada via WhatsApp
  ↓
Usuário responde (código pode ser corrompido)
  ↓
Sistema detecta código invisível
  ↓
Decodifica com recuperação de corrupção
  ↓
Associa tracking ao contato/sessão
  ↓
Métricas atualizadas
```

## 🔢 Sistema Ternário

### Por que Base 3?

- **Compacto**: 7 dígitos ternários = 2187 IDs únicos (0-2186)
- **Invisível**: 3 caracteres Unicode diferentes
- **Recuperável**: Heurísticas para corrupção do WhatsApp

### Limites

- **ID Máximo**: 2186 (2222222 em ternário)
- **Caracteres**: Exatamente 7 invisíveis
- **Posição**: Sempre após o 1º caractere

### Caracteres Invisíveis

| Base 3 | Unicode | Code Point | Nome |
|--------|---------|------------|------|
| 0 | \u200B | U+200B | Zero Width Space (ZWSP) |
| 1 | \u2060 | U+2060 | Word Joiner (WJ) |
| 2 | \uFEFF | U+FEFF | Zero Width No-Break Space (ZWNBSP) |

## 🛡️ Recuperação de Corrupção

O WhatsApp pode corromper caracteres invisíveis durante transmissão. O sistema implementa heurísticas de recuperação:

### Mapeamento de Corrupção

```go
// ZWSP (8203) → espaço normal (32) ou NBSP (160)
if charCode == 32 || charCode == 160 { return 0 }

// Word Joiner (8288) → outros joiners
if charCode == 8204 || charCode == 8205 { return 1 }

// ZWNBSP (65279) → BOM marks
if charCode == 65279 || charCode == 8206 || charCode == 8207 { return 2 }

// Hangul Filler (12644) → geralmente de ZWNBSP
if charCode == 12644 { return 2 }

// Braille/En Quad → geralmente de WJ
if charCode == 10240 || charCode == 8192 { return 1 }

// Heurística por código alto
if charCode > 10000 { return 2 }
if charCode > 8000 { return 1 }
return 0 // default
```

## 📡 API Endpoints

### POST /api/v1/trackings/encode

Codifica tracking ID em mensagem com código invisível.

**Request:**
```json
{
  "tracking_id": 123,
  "message": "Como posso te ajudar?",
  "phone": "5511999999999"
}
```

**Response:**
```json
{
  "success": true,
  "tracking_id": 123,
  "original_message": "Como posso te ajudar?",
  "ternary_encoded": "0011120",
  "decimal_value": 123,
  "phone": "5511999999999",
  "invisible_code": "[7 caracteres invisíveis]",
  "message_with_code": "C[invisível]omo posso te ajudar?",
  "whatsapp_link": "https://wa.me/5511999999999?text=...",
  "debug": {
    "input_original": 123,
    "ternary_value": "0011120",
    "decimal_equivalent": 123,
    "encoded_length": 7,
    "char_codes": [8203, 8203, 8203, 8288, 8288, 8203, 65279],
    "char_mapping": [...]
  }
}
```

### POST /api/v1/trackings/decode

Decodifica mensagem para extrair tracking ID.

**Request:**
```json
{
  "message": "C[invisível]omo posso te ajudar?"
}
```

**Response:**
```json
{
  "success": true,
  "decoded_ternary": "0011120",
  "decoded_decimal": 123,
  "confidence": "high",
  "analysis": {
    "first_char": "C",
    "extracted_chars": "[7 chars]",
    "char_codes": [8203, 8203, 8203, 8288, 8288, 8203, 65279],
    "char_analysis": [
      "PRESERVED: SAFE_0 (U+200B)",
      "CORRUPTED: U+0020 → Recovered as 0",
      ...
    ],
    "decoded_ternary": "0011120",
    "decoded_decimal": 123,
    "remaining_message": "omo posso te ajudar?"
  },
  "clean_message": "Como posso te ajudar?",
  "original_message": "C[invisível]omo posso te ajudar?"
}
```

### POST /api/v1/trackings

Cria novo tracking (padrão existente).

**Request:**
```json
{
  "contact_id": "uuid",
  "project_id": "uuid",
  "source": "meta_ads",
  "platform": "instagram",
  "campaign": "black_friday",
  "utm_source": "instagram",
  "utm_medium": "paid-social",
  "utm_campaign": "bf2025"
}
```

## 🔄 Detecção Automática

O sistema detecta automaticamente códigos invisíveis em mensagens recebidas:

1. **Mensagem chega** → `ProcessInboundMessageUseCase.Execute()`
2. **Verifica texto** → `ternaryEncoder.HasInvisibleCode()`
3. **Detectou código** → `ternaryEncoder.DecodeMessage()`
4. **Valida tracking** → Verifica se ID existe no banco
5. **Associa** → `UPDATE trackings SET contact_id, session_id WHERE id = ?`
6. **Log sucesso** → Tracking associado ao contato/sessão

### Exemplo de Log

```
🔍 Invisible tracking code detected: tracking_id=123, contact=uuid, clean_message=Como posso...
✅ Tracking 123 associated with contact uuid and session uuid
```

## 📈 Casos de Uso

### 1. Campanha Facebook Ads → WhatsApp

```
1. Criar tracking para campanha
2. Codificar mensagem padrão com tracking ID
3. Usar link no botão "Enviar mensagem" do anúncio
4. Usuário clica → abre WhatsApp com mensagem codificada
5. Usuário envia mensagem
6. Sistema detecta código → associa à campanha
7. Métricas: taxa de conversão, ROI, etc
```

### 2. Influenciadores / Disparos em Massa

```
1. Criar tracking por influenciador/lista
2. Gerar links personalizados
3. Distribuir links
4. Detectar origem de cada contato
5. Análise de performance por fonte
```

### 3. QR Codes Offline

```
1. Criar tracking para material impresso
2. Gerar QR code com link codificado
3. Cliente escaneia → abre WhatsApp
4. Sistema detecta origem offline
5. Atribuição cross-channel
```

## ⚙️ Configuração

### Variáveis de Ambiente

Nenhuma configuração adicional necessária. O sistema usa:
- Base de dados existente (tabela `trackings`)
- Autenticação existente (Bearer token)
- RLS existente (tenant isolation)

### Dependências

```go
import (
  "github.com/caloi/ventros-crm/internal/domain/tracking"
  "github.com/caloi/ventros-crm/internal/application/tracking"
)

// Inicializar encoder
encoder := tracking.NewTernaryEncoder()

// Codificar
encodedMsg, err := encoder.EncodeMessage("Olá", 123)

// Decodificar
trackingID, cleanMsg, err := encoder.DecodeMessage(encodedMsg)
```

## 🧪 Testes

### Teste de Codificação

```bash
curl -X POST http://localhost:8080/api/v1/trackings/encode \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tracking_id": 100,
    "message": "Olá! Como posso ajudar?",
    "phone": "5511999999999"
  }'
```

### Teste de Decodificação

```bash
curl -X POST http://localhost:8080/api/v1/trackings/decode \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "O[códigos invisíveis]lá! Como posso ajudar?"
  }'
```

## 🐛 Debug

### Verificar Códigos Invisíveis

```go
// Analisar mensagem detalhadamente
analysis := encoder.GetAnalysis(message)
fmt.Printf("%+v\n", analysis)
```

### Logs do Sistema

```
🔍 = Detecção de código
✅ = Associação bem-sucedida
⚠️  = Aviso (tracking não encontrado, etc)
```

## 📊 Métricas e Analytics

Com o sistema de tracking invisível, você pode:

- ✅ Rastrear origem exata de cada contato
- ✅ Calcular ROI por campanha/plataforma
- ✅ A/B testing de mensagens
- ✅ Atribuição cross-channel
- ✅ Performance de influenciadores
- ✅ Eficácia de disparos em massa
- ✅ Conversão de QR codes offline

## 🔒 Segurança e Privacidade

- ✅ Códigos não são visíveis ao usuário
- ✅ Não afetam experiência de uso
- ✅ Isolamento por tenant (RLS)
- ✅ Autenticação obrigatória
- ✅ Associação somente se tracking existe
- ✅ Não expõe dados sensíveis

## 📝 Notas Técnicas

### Limitações

1. **ID máximo**: 2186 (base 10) = 2222222 (base 3)
2. **Comprimento fixo**: Sempre 7 dígitos ternários
3. **Posição fixa**: Sempre após 1º caractere
4. **Recuperação parcial**: Nem sempre 100% de recuperação em caso de corrupção severa

### Extensões Futuras

- [ ] Suporte a mais caracteres (aumentar range)
- [ ] Codificação de metadata adicional
- [ ] Compressão de IDs maiores
- [ ] Múltiplos códigos em uma mensagem
- [ ] Validação de integridade (checksum)

## 🎓 Referências

- [Unicode Zero-Width Characters](https://en.wikipedia.org/wiki/Zero-width_space)
- [WhatsApp Business API Best Practices](https://developers.facebook.com/docs/whatsapp/business-management-api)
- [Base-3 (Ternary) Number System](https://en.wikipedia.org/wiki/Ternary_numeral_system)
