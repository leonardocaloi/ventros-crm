package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	messageports "github.com/caloi/ventros-crm/internal/application/message"
	"github.com/google/uuid"
)

// MessageScheduler implementa agendamento de mensagens
// Seguindo Single Responsibility Principle (SRP)
type MessageScheduler struct {
	scheduledMessages map[uuid.UUID]*ScheduledMessage
	mutex             sync.RWMutex
	ticker            *time.Ticker
	stopChan          chan struct{}
	messageQueue      messageports.MessageQueue
	running           bool
}

// ScheduledMessage representa uma mensagem agendada
type ScheduledMessage struct {
	Message     *messageports.OutboundMessage
	ScheduledAt time.Time
	Timer       *time.Timer
	Cancelled   bool
}

// NewMessageScheduler cria uma nova instância do scheduler
func NewMessageScheduler(messageQueue messageports.MessageQueue) messageports.MessageScheduler {
	return &MessageScheduler{
		scheduledMessages: make(map[uuid.UUID]*ScheduledMessage),
		messageQueue:      messageQueue,
		stopChan:          make(chan struct{}),
	}
}

// ScheduleMessage agenda uma mensagem para envio futuro
func (s *MessageScheduler) ScheduleMessage(ctx context.Context, message *messageports.OutboundMessage) error {
	if message.ScheduledAt == nil {
		return fmt.Errorf("message must have a scheduled time")
	}

	if message.ScheduledAt.Before(time.Now()) {
		return fmt.Errorf("cannot schedule message in the past")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Calcular delay
	delay := time.Until(*message.ScheduledAt)

	// Criar timer para a mensagem
	timer := time.AfterFunc(delay, func() {
		s.processScheduledMessage(message.ID)
	})

	// Armazenar mensagem agendada
	s.scheduledMessages[message.ID] = &ScheduledMessage{
		Message:     message,
		ScheduledAt: *message.ScheduledAt,
		Timer:       timer,
		Cancelled:   false,
	}

	return nil
}

// GetScheduledMessages retorna mensagens agendadas antes de um tempo específico
func (s *MessageScheduler) GetScheduledMessages(ctx context.Context, before time.Time) ([]*messageports.OutboundMessage, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var messages []*messageports.OutboundMessage
	for _, scheduled := range s.scheduledMessages {
		if !scheduled.Cancelled && scheduled.ScheduledAt.Before(before) {
			messages = append(messages, scheduled.Message)
		}
	}

	return messages, nil
}

// CancelScheduledMessage cancela uma mensagem agendada
func (s *MessageScheduler) CancelScheduledMessage(ctx context.Context, messageID uuid.UUID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	scheduled, exists := s.scheduledMessages[messageID]
	if !exists {
		return fmt.Errorf("scheduled message not found")
	}

	if scheduled.Cancelled {
		return fmt.Errorf("message already cancelled")
	}

	// Parar o timer
	if scheduled.Timer != nil {
		scheduled.Timer.Stop()
	}

	// Marcar como cancelada
	scheduled.Cancelled = true

	// Remover da lista
	delete(s.scheduledMessages, messageID)

	return nil
}

// processScheduledMessage processa uma mensagem quando chega a hora
func (s *MessageScheduler) processScheduledMessage(messageID uuid.UUID) {
	s.mutex.Lock()
	scheduled, exists := s.scheduledMessages[messageID]
	if !exists || scheduled.Cancelled {
		s.mutex.Unlock()
		return
	}

	message := scheduled.Message
	delete(s.scheduledMessages, messageID)
	s.mutex.Unlock()

	// Adicionar à fila para processamento
	ctx := context.Background()
	if err := s.messageQueue.Enqueue(ctx, message); err != nil {
		// Log error - em produção, usar logger apropriado
		// Por enquanto, apenas ignorar o erro
	}
}

// Start inicia o scheduler
func (s *MessageScheduler) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return
	}

	s.running = true
	s.ticker = time.NewTicker(1 * time.Minute) // Verificar a cada minuto

	go s.cleanupLoop()
}

// Stop para o scheduler
func (s *MessageScheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)

	if s.ticker != nil {
		s.ticker.Stop()
	}

	// Cancelar todos os timers
	for _, scheduled := range s.scheduledMessages {
		if scheduled.Timer != nil {
			scheduled.Timer.Stop()
		}
	}

	s.scheduledMessages = make(map[uuid.UUID]*ScheduledMessage)
}

// cleanupLoop limpa mensagens expiradas periodicamente
func (s *MessageScheduler) cleanupLoop() {
	for {
		select {
		case <-s.ticker.C:
			s.cleanup()
		case <-s.stopChan:
			return
		}
	}
}

// cleanup remove mensagens canceladas ou expiradas
func (s *MessageScheduler) cleanup() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for messageID, scheduled := range s.scheduledMessages {
		// Remover mensagens canceladas ou que já passaram do tempo (com margem)
		if scheduled.Cancelled || scheduled.ScheduledAt.Add(5*time.Minute).Before(now) {
			if scheduled.Timer != nil {
				scheduled.Timer.Stop()
			}
			delete(s.scheduledMessages, messageID)
		}
	}
}

// GetSchedulerStats retorna estatísticas do scheduler
func (s *MessageScheduler) GetSchedulerStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var totalScheduled, cancelled, upcoming int
	now := time.Now()

	for _, scheduled := range s.scheduledMessages {
		totalScheduled++
		if scheduled.Cancelled {
			cancelled++
		} else if scheduled.ScheduledAt.After(now) {
			upcoming++
		}
	}

	return map[string]interface{}{
		"total_scheduled": totalScheduled,
		"cancelled":       cancelled,
		"upcoming":        upcoming,
		"running":         s.running,
	}
}

// RescheduleMessage reagenda uma mensagem
func (s *MessageScheduler) RescheduleMessage(ctx context.Context, messageID uuid.UUID, newTime time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	scheduled, exists := s.scheduledMessages[messageID]
	if !exists {
		return fmt.Errorf("scheduled message not found")
	}

	if scheduled.Cancelled {
		return fmt.Errorf("cannot reschedule cancelled message")
	}

	if newTime.Before(time.Now()) {
		return fmt.Errorf("cannot reschedule message to the past")
	}

	// Parar timer atual
	if scheduled.Timer != nil {
		scheduled.Timer.Stop()
	}

	// Criar novo timer
	delay := time.Until(newTime)
	scheduled.Timer = time.AfterFunc(delay, func() {
		s.processScheduledMessage(messageID)
	})

	// Atualizar tempo agendado
	scheduled.ScheduledAt = newTime
	scheduled.Message.ScheduledAt = &newTime

	return nil
}

// GetScheduledMessage retorna uma mensagem agendada específica
func (s *MessageScheduler) GetScheduledMessage(messageID uuid.UUID) (*ScheduledMessage, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	scheduled, exists := s.scheduledMessages[messageID]
	return scheduled, exists
}

// ListScheduledMessages retorna todas as mensagens agendadas
func (s *MessageScheduler) ListScheduledMessages() []*ScheduledMessage {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var messages []*ScheduledMessage
	for _, scheduled := range s.scheduledMessages {
		if !scheduled.Cancelled {
			messages = append(messages, scheduled)
		}
	}

	return messages
}

// BatchScheduleMessages agenda múltiplas mensagens
func (s *MessageScheduler) BatchScheduleMessages(ctx context.Context, messages []*messageports.OutboundMessage) error {
	for _, message := range messages {
		if err := s.ScheduleMessage(ctx, message); err != nil {
			return fmt.Errorf("failed to schedule message %s: %w", message.ID, err)
		}
	}

	return nil
}

// BatchCancelMessages cancela múltiplas mensagens
func (s *MessageScheduler) BatchCancelMessages(ctx context.Context, messageIDs []uuid.UUID) error {
	for _, messageID := range messageIDs {
		if err := s.CancelScheduledMessage(ctx, messageID); err != nil {
			// Continue cancelando outras mensagens mesmo se uma falhar
			continue
		}
	}

	return nil
}
