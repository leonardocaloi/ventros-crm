package ai

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// AIDebouncer previne processamento duplicado de mensagens
type AIDebouncer struct {
	logger  *zap.Logger
	cache   map[string]time.Time
	mu      sync.RWMutex
	janitor *time.Ticker
}

// NewAIDebouncer cria um novo debouncer de IA
func NewAIDebouncer(logger *zap.Logger) *AIDebouncer {
	d := &AIDebouncer{
		logger:  logger,
		cache:   make(map[string]time.Time),
		janitor: time.NewTicker(1 * time.Minute), // Limpeza a cada 1 minuto
	}

	// Iniciar goroutine de limpeza
	go d.cleanupLoop()

	return d
}

// ShouldProcess verifica se deve processar baseado no debounce
func (d *AIDebouncer) ShouldProcess(key string, debounceWindow time.Duration) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	lastProcessed, exists := d.cache[key]

	if !exists {
		// Primeira vez processando esta mensagem
		d.cache[key] = now
		return true
	}

	// Verificar se o tempo de debounce já passou
	if now.Sub(lastProcessed) > debounceWindow {
		// Tempo de debounce expirou, pode processar novamente
		d.cache[key] = now
		return true
	}

	// Ainda dentro da janela de debounce
	d.logger.Debug("Message debounced",
		zap.String("key", key),
		zap.Duration("time_since_last", now.Sub(lastProcessed)),
		zap.Duration("debounce_window", debounceWindow))

	return false
}

// cleanupLoop remove entradas antigas do cache
func (d *AIDebouncer) cleanupLoop() {
	for range d.janitor.C {
		d.cleanup()
	}
}

// cleanup remove entradas com mais de 5 minutos
func (d *AIDebouncer) cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-5 * time.Minute)
	cleaned := 0

	for key, timestamp := range d.cache {
		if timestamp.Before(cutoff) {
			delete(d.cache, key)
			cleaned++
		}
	}

	if cleaned > 0 {
		d.logger.Debug("Cleaned up debouncer cache",
			zap.Int("entries_removed", cleaned),
			zap.Int("entries_remaining", len(d.cache)))
	}
}

// Stop para a goroutine de limpeza
func (d *AIDebouncer) Stop() {
	d.janitor.Stop()
}

// Clear limpa todo o cache (útil para testes)
func (d *AIDebouncer) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[string]time.Time)
}
