package messaging

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConfig contém a configuração do RabbitMQ.
type RabbitMQConfig struct {
	URL             string
	ReconnectDelay  time.Duration
	MaxReconnects   int
}

// RabbitMQConnection gerencia a conexão com RabbitMQ com auto-reconnect.
type RabbitMQConnection struct {
	config     RabbitMQConfig
	conn       *amqp.Connection
	channel    *amqp.Channel
	notifyClose chan *amqp.Error
	isConnected bool
}

// NewRabbitMQConnection cria uma nova conexão com RabbitMQ.
func NewRabbitMQConnection(config RabbitMQConfig) (*RabbitMQConnection, error) {
	r := &RabbitMQConnection{
		config: config,
	}
	
	if err := r.connect(); err != nil {
		return nil, err
	}
	
	// Inicia goroutine para auto-reconnect
	go r.handleReconnect()
	
	return r, nil
}

// connect estabelece conexão com RabbitMQ.
func (r *RabbitMQConnection) connect() error {
	conn, err := amqp.Dial(r.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}
	
	r.conn = conn
	r.channel = ch
	r.notifyClose = make(chan *amqp.Error)
	r.channel.NotifyClose(r.notifyClose)
	r.isConnected = true
	
	return nil
}

// handleReconnect monitora e reconecta automaticamente.
func (r *RabbitMQConnection) handleReconnect() {
	for {
		err := <-r.notifyClose
		if err != nil {
			r.isConnected = false
			fmt.Printf("Connection closed: %v. Reconnecting...\n", err)
			
			// Tenta reconectar
			for i := 0; i < r.config.MaxReconnects; i++ {
				time.Sleep(r.config.ReconnectDelay)
				
				if err := r.connect(); err == nil {
					fmt.Println("Reconnected successfully")
					break
				}
				
				fmt.Printf("Reconnect attempt %d failed\n", i+1)
			}
		}
	}
}

// Channel retorna o canal RabbitMQ.
func (r *RabbitMQConnection) Channel() *amqp.Channel {
	return r.channel
}

// Close fecha a conexão.
func (r *RabbitMQConnection) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// IsConnected retorna se está conectado.
func (r *RabbitMQConnection) IsConnected() bool {
	return r.isConnected
}

// DeclareQueue declara uma fila com configurações padrão.
// A declaração é idempotente - se a fila já existe com os mesmos argumentos, não faz nada.
// Se existir com argumentos diferentes, retorna erro (comportamento padrão do RabbitMQ).
func (r *RabbitMQConnection) DeclareQueue(name string) error {
	args := amqp.Table{
		"x-queue-type": "quorum", // Alta disponibilidade
	}
	
	// Declaração idempotente - RabbitMQ aceita se a fila já existe com mesmos args
	_, err := r.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,
	)
	
	if err != nil {
		// Se falhar por incompatibilidade de argumentos, loga mas não falha
		// Isso permite usar filas legadas que foram criadas sem quorum
		if isQueuePreconditionError(err) {
			fmt.Printf("⚠️  Queue %s exists with different args (legacy queue), using it anyway\n", name)
			
			// Reabre o canal (PRECONDITION_FAILED fecha o canal)
			if reopenErr := r.reopenChannel(); reopenErr != nil {
				return fmt.Errorf("failed to reopen channel: %w", reopenErr)
			}
			
			return nil // Aceita a fila legada
		}
		
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	
	return nil
}

// isQueuePreconditionError verifica se o erro é de incompatibilidade de argumentos de fila
func isQueuePreconditionError(err error) bool {
	if err == nil {
		return false
	}
	
	// Verifica se é um erro AMQP 406 (PRECONDITION_FAILED)
	if amqpErr, ok := err.(*amqp.Error); ok {
		return amqpErr.Code == 406
	}
	
	return false
}

// reopenChannel reabre o canal após um erro que o fechou
func (r *RabbitMQConnection) reopenChannel() error {
	if r.conn == nil || r.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}
	
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to reopen channel: %w", err)
	}
	
	r.channel = ch
	r.notifyClose = make(chan *amqp.Error)
	r.channel.NotifyClose(r.notifyClose)
	
	return nil
}

// DeclareQueueWithDLQ declara fila com Dead Letter Queue.
// Suporta retry automático antes de enviar para DLQ.
func (r *RabbitMQConnection) DeclareQueueWithDLQ(name string, maxRetries int) error {
	// Cria DLQ primeiro
	dlqName := name + ".dlq"
	dlqArgs := amqp.Table{
		"x-queue-type": "quorum",
	}
	
	// Declaração idempotente da DLQ
	_, err := r.channel.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		dlqArgs,
	)
	
	if err != nil {
		// Se falhar por incompatibilidade, aceita a fila legada
		if isQueuePreconditionError(err) {
			fmt.Printf("⚠️  DLQ %s exists with different args (legacy), using it anyway\n", dlqName)
			
			if reopenErr := r.reopenChannel(); reopenErr != nil {
				return fmt.Errorf("failed to reopen channel: %w", reopenErr)
			}
			
			err = nil // Aceita a DLQ legada
		} else {
			return fmt.Errorf("failed to declare DLQ: %w", err)
		}
	}
	
	// Cria fila principal com DLQ configurada
	queueArgs := amqp.Table{
		"x-queue-type":             "quorum",
		"x-dead-letter-exchange":   "",       // Default exchange
		"x-dead-letter-routing-key": dlqName, // Roteia para DLQ
	}
	
	// Se maxRetries > 0, configura delay para retry
	if maxRetries > 0 {
		// Após N rejeições, vai para DLQ
		// RabbitMQ 3.13+ suporta x-delivery-limit
		queueArgs["x-delivery-limit"] = maxRetries
	}
	
	// Declaração idempotente da fila principal
	_, err = r.channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		queueArgs,
	)
	
	if err != nil {
		// Se falhar por incompatibilidade, aceita a fila legada
		if isQueuePreconditionError(err) {
			fmt.Printf("⚠️  Queue %s exists with different args (legacy), using it anyway\n", name)
			
			if reopenErr := r.reopenChannel(); reopenErr != nil {
				return fmt.Errorf("failed to reopen channel: %w", reopenErr)
			}
			
			return nil // Aceita a fila legada
		}
		
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	
	return nil
}

// Publish publica uma mensagem.
func (r *RabbitMQConnection) Publish(ctx context.Context, queue string, body []byte) error {
	return r.channel.PublishWithContext(
		ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
}

// Consumer interface para processadores de mensagens.
type Consumer interface {
	ProcessMessage(ctx context.Context, delivery amqp.Delivery) error
}

// StartConsumer inicia um consumer para uma fila.
func (r *RabbitMQConnection) StartConsumer(
	ctx context.Context,
	queueName string,
	consumerTag string,
	consumer Consumer,
	prefetchCount int,
) error {
	// Configura QoS (quantas mensagens processar simultaneamente)
	if err := r.channel.Qos(prefetchCount, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}
	
	// Inicia consumer
	msgs, err := r.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer tag
		false,       // auto-ack (manual para garantir processamento)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}
	
	// Processa mensagens
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Consumer stopped")
				return
							case msg, ok := <-msgs:
					if !ok {
						fmt.Println("Channel closed")
						return
					}
					
					// Processa mensagem
					if err := consumer.ProcessMessage(ctx, msg); err != nil {
						fmt.Printf("Error processing message: %v\n", err)
						
						// Verifica número de tentativas (x-death header)
						retryCount := getRetryCount(msg)
						fmt.Printf("Retry count: %d\n", retryCount)
						
						// Se ainda tem retries, requeue
						// Senão, Nack sem requeue → vai para DLQ
						if retryCount < 3 {
							// Nack com requeue (tenta novamente)
							msg.Nack(false, true)
						} else {
							// Após 3 tentativas, envia para DLQ
							fmt.Printf("Max retries reached, sending to DLQ\n")
							msg.Nack(false, false) // requeue=false → DLQ
						}
					} else {
						// Ack após sucesso
						msg.Ack(false)
					}
			}
		}
	}()
	
	return nil
}

// getRetryCount retorna o número de tentativas de uma mensagem.
func getRetryCount(msg amqp.Delivery) int {
	// x-death é uma array de tables com informações sobre rejeições
	xDeath, ok := msg.Headers["x-death"]
	if !ok {
		return 0
	}
	
	// Parse x-death (é uma []interface{} de amqp.Table)
	deaths, ok := xDeath.([]interface{})
	if !ok || len(deaths) == 0 {
		return 0
	}
	
	// Pega o primeiro elemento (mais recente)
	firstDeath, ok := deaths[0].(amqp.Table)
	if !ok {
		return 0
	}
	
	// Conta quantas vezes foi rejeitada
	count, ok := firstDeath["count"]
	if !ok {
		return 0
	}
	
	// Converte para int
	switch v := count.(type) {
	case int64:
		return int(v)
	case int32:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

// QueueInfo representa informações sobre uma fila
type QueueInfo struct {
	Name      string `json:"name" example:"waha.events.message"`
	Messages  int    `json:"messages" example:"42"`
	Consumers int    `json:"consumers" example:"2"`
	IsDLQ     bool   `json:"is_dlq" example:"false"`
}

// ListQueues lista todas as filas declaradas dinamicamente
func (r *RabbitMQConnection) ListQueues() ([]QueueInfo, error) {
	if !r.isConnected {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	// Usa a API de management do RabbitMQ para listar filas existentes
	// Como não temos acesso direto à API de management, vamos usar uma abordagem
	// baseada nas filas que realmente declaramos no sistema
	
	var queues []QueueInfo
	
	
	// Tenta descobrir filas existentes baseado nos padrões conhecidos
	potentialQueues := []string{
		// WAHA Events - Mensagens
		"waha.events.message",
		"waha.events.message.dlq",
		"waha.events.message.any",
		"waha.events.message.any.dlq",
		"waha.events.message.ack",
		"waha.events.message.ack.dlq",
		"waha.events.message.reaction",
		"waha.events.message.reaction.dlq",
		"waha.events.message.edited",
		"waha.events.message.edited.dlq",
		
		// WAHA Events - Chamadas
		"waha.events.call.received",
		"waha.events.call.received.dlq",
		"waha.events.call.accepted",
		"waha.events.call.accepted.dlq",
		"waha.events.call.rejected",
		"waha.events.call.rejected.dlq",
		
		// WAHA Events - Falhas
		"waha.events.event.response.failed",
		"waha.events.event.response.failed.dlq",
		
		// WAHA Events - Labels
		"waha.events.label.upsert",
		"waha.events.label.upsert.dlq",
		"waha.events.label.deleted",
		"waha.events.label.deleted.dlq",
		"waha.events.label.chat.added",
		"waha.events.label.chat.added.dlq",
		"waha.events.label.chat.deleted",
		"waha.events.label.chat.deleted.dlq",
		
		// WAHA Events - Grupos v2
		"waha.events.group.v2.join",
		"waha.events.group.v2.join.dlq",
		"waha.events.group.v2.leave",
		"waha.events.group.v2.leave.dlq",
		"waha.events.group.v2.update",
		"waha.events.group.v2.update.dlq",
		"waha.events.group.v2.participants",
		"waha.events.group.v2.participants.dlq",
		
		// Domain Events
		"domain.events.contact.created",
		"domain.events.contact.created.dlq",
		"domain.events.session.started", 
		"domain.events.session.started.dlq",
	}

	for _, queueName := range potentialQueues {
		// QueueInspect retorna informações sobre a fila se ela existir
		queue, err := r.channel.QueueInspect(queueName)
		if err != nil {
			// Fila não existe, pula silenciosamente
			continue
		}

		queues = append(queues, QueueInfo{
			Name:      queue.Name,
			Messages:  queue.Messages,
			Consumers: queue.Consumers,
			IsDLQ:     len(queue.Name) > 4 && queue.Name[len(queue.Name)-4:] == ".dlq",
		})
	}

	return queues, nil
}

// SetupWAHAQueues declara todas as filas necessárias para eventos WAHA
func (r *RabbitMQConnection) SetupWAHAQueues() error {
	// WAHA event queues - todas as filas de eventos que chegam da WAHA
	wahaQueues := []string{
		// Fila de entrada (raw events) - NOVA ARQUITETURA
		"waha.events.raw",
		
		// Filas de saída (eventos processados) - NOVA ARQUITETURA
		"waha.events.message.parsed",
		"waha.events.call.parsed",
		"waha.events.presence.parsed",
		"waha.events.group.parsed",
		"waha.events.label.parsed",
		"waha.events.unknown.parsed",
		"waha.events.parse_errors",
		
		// Filas legadas (manter compatibilidade)
		"waha.events.message",
		"waha.events.message.any",
		"waha.events.message.ack",
		"waha.events.message.reaction",
		"waha.events.message.edited",
		
		// Chamadas
		"waha.events.call.received",
		"waha.events.call.accepted", 
		"waha.events.call.rejected",
		
		// Eventos de resposta com falha
		"waha.events.event.response.failed",
		
		// Labels/Tags
		"waha.events.label.upsert",
		"waha.events.label.deleted",
		"waha.events.label.chat.added",
		"waha.events.label.chat.deleted",
		
		// Grupos v2
		"waha.events.group.v2.join",
		"waha.events.group.v2.leave", 
		"waha.events.group.v2.update",
		"waha.events.group.v2.participants",
		
		// Eventos de sessão/conexão
		"waha.events.session.status",
		
		// Presença (online/offline)
		"waha.events.presence",
		
		// Eventos desconhecidos (fallback)
		"waha.events.unknown",
	}
	
	for _, queue := range wahaQueues {
		if err := r.DeclareQueueWithDLQ(queue, 3); err != nil {
			return fmt.Errorf("failed to declare %s queue: %w", queue, err)
		}
	}
	
	return nil
}

// SetupAllQueues configura todas as filas do sistema (WAHA + Domain Events)
func (r *RabbitMQConnection) SetupAllQueues() error {
	// Setup WAHA queues
	if err := r.SetupWAHAQueues(); err != nil {
		return fmt.Errorf("failed to setup WAHA queues: %w", err)
	}
	
	// Setup domain event queues
	eventBus := NewDomainEventBus(r, nil, nil) // nil webhook notifier and event log repo for setup only
	if err := eventBus.SetupEventQueues(); err != nil {
		return fmt.Errorf("failed to setup domain event queues: %w", err)
	}
	
	return nil
}
