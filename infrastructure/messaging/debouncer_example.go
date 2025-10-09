package messaging

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// ExampleUsage mostra como usar o debouncer de diferentes formas

// Exemplo 1: Debouncer SEM processamento (apenas agrupa)
func ExampleDebouncerOnly(redisClient *redis.Client) *MessageDebouncerV2 {
	// Cria debouncer sem processor
	// Útil quando você só quer agrupar mensagens mas processar externamente
	debouncer := NewMessageDebouncerV2(redisClient, 0, nil)

	return debouncer
}

// Exemplo 2: Debouncer COM processamento simples
func ExampleDebouncerWithSimpleProcessor(redisClient *redis.Client) *DebouncerIntegration {
	// Cria processor básico (apenas concatena)
	processor := NewMessageBatchProcessor(
		SimpleConcatenator{}, // concatena texto
		NoopValidator{},      // não valida
		nil,                  // sem enrichment
		nil,                  // sem sender (você processa manualmente)
	)

	// Integração completa
	integration := NewDebouncerIntegration(redisClient, processor)

	return integration
}

// Exemplo 3: Debouncer COM processamento completo (AI-ready)
func ExampleDebouncerForAI(redisClient *redis.Client, aiSender MessageSender) *DebouncerIntegration {
	// Cria processor sofisticado para IA
	processor := NewMessageBatchProcessor(
		MediaAwareConcatenator{},         // detecta mídia
		MinMessageValidator{MinCount: 1}, // mínimo 1 mensagem
		nil,                              // enricher customizado (implemente se necessário)
		aiSender,                         // envia para IA
	)

	integration := NewDebouncerIntegration(redisClient, processor)

	return integration
}

// Exemplo 4: Custom Sender para OpenAI
type OpenAISender struct {
	// seus campos aqui
}

func (s *OpenAISender) Send(ctx context.Context, sessionKey string, content string, metadata interface{}) error {
	// 1. Monta prompt com contexto
	fmt.Printf("📤 [OpenAI] Sending to AI: session=%s\n", sessionKey)
	fmt.Printf("Content: %s\n", content)

	// 2. Chama API OpenAI (implementar)
	// response := openai.CreateChatCompletion(...)

	// 3. Processa resposta e envia ao contato
	// wahaClient.SendMessage(...)

	return nil
}

// Exemplo 5: Uso no WAHAMessageConsumer (substituindo lógica existente)
func ExampleIntegrationWithWAHA(ctx context.Context, integration *DebouncerIntegration) {
	// No seu WAHAMessageConsumer.ProcessMessage:

	// wahaEvent := ... (recebeu do RabbitMQ)

	// Ao invés de processar direto, usa debouncer:
	// err := integration.ProcessWAHAMessage(ctx, wahaEvent)

	// OU versão genérica:
	// err := integration.ProcessMessage(
	// 	ctx,
	// 	contactID,
	// 	"waha",
	// 	channelID,
	// 	messageID,
	// 	text,
	// 	messageType,
	// 	timestamp,
	// 	fromContact,
	// 	metadata,
	// )
}

// Exemplo 6: Uso direto sem integração (máximo controle)
func ExampleManualControl(ctx context.Context, debouncer *MessageDebouncerV2) {
	sessionKey := "contact123:waha:channel456"

	// 1. Cria mensagem
	msg := BufferedMessage{
		MessageID:   "msg_001",
		Text:        "Olá",
		Type:        "text",
		Timestamp:   1234567890000,
		FromContact: true,
		ContactID:   "contact123",
	}

	// 2. Push manual
	err := debouncer.Push(ctx, sessionKey, msg)
	if err != nil {
		panic(err)
	}

	// 3. Verifica tamanho do buffer
	size, _ := debouncer.GetBufferSize(ctx, sessionKey)
	fmt.Printf("Buffer size: %d\n", size)

	// 4. Pull manual (quando quiser processar)
	messages, err := debouncer.Pull(ctx, sessionKey)
	if err != nil {
		panic(err)
	}

	// 5. Processa como quiser
	for _, m := range messages {
		fmt.Printf("Mensagem: %s\n", m.Text)
	}

	// 6. Limpa buffer
	debouncer.ClearBuffer(ctx, sessionKey)
}

// Exemplo 7: Configuração completa no main.go
func ExampleMainSetup(redisClient *redis.Client) {
	// Opção A: Sem processamento automático (você controla tudo)
	debouncerOnly := ExampleDebouncerOnly(redisClient)
	_ = debouncerOnly

	// Opção B: Com processamento simples
	simpleIntegration := ExampleDebouncerWithSimpleProcessor(redisClient)
	_ = simpleIntegration

	// Opção C: Com AI (OpenAI/Anthropic)
	aiSender := &OpenAISender{}
	aiIntegration := ExampleDebouncerForAI(redisClient, aiSender)
	_ = aiIntegration

	// Use no seu WAHAMessageConsumer:
	// consumer := NewWAHAMessageConsumer(wahaMessageService, idempotencyChecker, aiIntegration)
}
