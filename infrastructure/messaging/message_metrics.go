package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	messageports "github.com/ventros/crm/internal/application/message"
)

// MessageMetrics implementa coleta de métricas de mensagens
// Seguindo Single Responsibility Principle (SRP)
type MessageMetrics struct {
	stats map[string]*ChannelStats
	mutex sync.RWMutex
}

// ChannelStats representa estatísticas de um canal
type ChannelStats struct {
	ChannelType    string                             `json:"channel_type"`
	TotalSent      int64                              `json:"total_sent"`
	TotalFailed    int64                              `json:"total_failed"`
	TypeBreakdown  map[messageports.MessageType]int64 `json:"type_breakdown"`
	FailureReasons map[string]int64                   `json:"failure_reasons"`
	LatencySum     time.Duration                      `json:"-"`
	LatencyCount   int64                              `json:"-"`
	LastReset      time.Time                          `json:"last_reset"`
}

// NewMessageMetrics cria uma nova instância de métricas
func NewMessageMetrics() messageports.MessageMetrics {
	return &MessageMetrics{
		stats: make(map[string]*ChannelStats),
	}
}

// RecordMessageSent registra uma mensagem enviada
func (m *MessageMetrics) RecordMessageSent(channelType string, messageType messageports.MessageType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	stats := m.getOrCreateStats(channelType)
	stats.TotalSent++

	if stats.TypeBreakdown == nil {
		stats.TypeBreakdown = make(map[messageports.MessageType]int64)
	}
	stats.TypeBreakdown[messageType]++
}

// RecordMessageFailed registra uma mensagem falhada
func (m *MessageMetrics) RecordMessageFailed(channelType string, messageType messageports.MessageType, reason string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	stats := m.getOrCreateStats(channelType)
	stats.TotalFailed++

	if stats.TypeBreakdown == nil {
		stats.TypeBreakdown = make(map[messageports.MessageType]int64)
	}
	stats.TypeBreakdown[messageType]++

	if stats.FailureReasons == nil {
		stats.FailureReasons = make(map[string]int64)
	}
	stats.FailureReasons[reason]++
}

// RecordDeliveryTime registra tempo de entrega
func (m *MessageMetrics) RecordDeliveryTime(channelType string, duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	stats := m.getOrCreateStats(channelType)
	stats.LatencySum += duration
	stats.LatencyCount++
}

// GetMessageStats retorna estatísticas de um canal
func (m *MessageMetrics) GetMessageStats(ctx context.Context, channelType string) (*messageports.MessageStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats, exists := m.stats[channelType]
	if !exists {
		return &messageports.MessageStats{
			TotalSent:      0,
			TotalFailed:    0,
			SuccessRate:    0,
			AverageLatency: 0,
			TypeBreakdown:  make(map[messageports.MessageType]int64),
		}, nil
	}

	var averageLatency time.Duration
	if stats.LatencyCount > 0 {
		averageLatency = stats.LatencySum / time.Duration(stats.LatencyCount)
	}

	var successRate float64
	total := stats.TotalSent + stats.TotalFailed
	if total > 0 {
		successRate = float64(stats.TotalSent) / float64(total) * 100
	}

	return &messageports.MessageStats{
		TotalSent:      stats.TotalSent,
		TotalFailed:    stats.TotalFailed,
		SuccessRate:    successRate,
		AverageLatency: averageLatency,
		TypeBreakdown:  stats.TypeBreakdown,
	}, nil
}

// GetAllStats retorna estatísticas de todos os canais
func (m *MessageMetrics) GetAllStats(ctx context.Context) map[string]*messageports.MessageStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*messageports.MessageStats)

	for channelType := range m.stats {
		stats, _ := m.GetMessageStats(ctx, channelType)
		result[channelType] = stats
	}

	return result
}

// ResetStats reseta estatísticas de um canal
func (m *MessageMetrics) ResetStats(channelType string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if stats, exists := m.stats[channelType]; exists {
		stats.TotalSent = 0
		stats.TotalFailed = 0
		stats.TypeBreakdown = make(map[messageports.MessageType]int64)
		stats.FailureReasons = make(map[string]int64)
		stats.LatencySum = 0
		stats.LatencyCount = 0
		stats.LastReset = time.Now()
	}
}

// ResetAllStats reseta todas as estatísticas
func (m *MessageMetrics) ResetAllStats() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for channelType := range m.stats {
		stats := m.stats[channelType]
		stats.TotalSent = 0
		stats.TotalFailed = 0
		stats.TypeBreakdown = make(map[messageports.MessageType]int64)
		stats.FailureReasons = make(map[string]int64)
		stats.LatencySum = 0
		stats.LatencyCount = 0
		stats.LastReset = time.Now()
	}
}

// getOrCreateStats obtém ou cria estatísticas para um canal
func (m *MessageMetrics) getOrCreateStats(channelType string) *ChannelStats {
	stats, exists := m.stats[channelType]
	if !exists {
		stats = &ChannelStats{
			ChannelType:    channelType,
			TotalSent:      0,
			TotalFailed:    0,
			TypeBreakdown:  make(map[messageports.MessageType]int64),
			FailureReasons: make(map[string]int64),
			LatencySum:     0,
			LatencyCount:   0,
			LastReset:      time.Now(),
		}
		m.stats[channelType] = stats
	}
	return stats
}

// GetTopFailureReasons retorna as principais razões de falha
func (m *MessageMetrics) GetTopFailureReasons(channelType string, limit int) []FailureReason {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats, exists := m.stats[channelType]
	if !exists || stats.FailureReasons == nil {
		return []FailureReason{}
	}

	// Converter para slice e ordenar
	var reasons []FailureReason
	for reason, count := range stats.FailureReasons {
		reasons = append(reasons, FailureReason{
			Reason: reason,
			Count:  count,
		})
	}

	// Ordenar por contagem (decrescente)
	for i := 0; i < len(reasons)-1; i++ {
		for j := i + 1; j < len(reasons); j++ {
			if reasons[j].Count > reasons[i].Count {
				reasons[i], reasons[j] = reasons[j], reasons[i]
			}
		}
	}

	// Aplicar limite
	if limit > 0 && len(reasons) > limit {
		reasons = reasons[:limit]
	}

	return reasons
}

// FailureReason representa uma razão de falha com contagem
type FailureReason struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}

// GetChannelPerformance retorna performance comparativa dos canais
func (m *MessageMetrics) GetChannelPerformance() []ChannelPerformance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var performance []ChannelPerformance

	for channelType, stats := range m.stats {
		total := stats.TotalSent + stats.TotalFailed
		var successRate float64
		var avgLatency time.Duration

		if total > 0 {
			successRate = float64(stats.TotalSent) / float64(total) * 100
		}

		if stats.LatencyCount > 0 {
			avgLatency = stats.LatencySum / time.Duration(stats.LatencyCount)
		}

		performance = append(performance, ChannelPerformance{
			ChannelType:     channelType,
			TotalMessages:   total,
			SuccessRate:     successRate,
			AverageLatency:  avgLatency,
			MessagesPerHour: m.calculateMessagesPerHour(stats),
		})
	}

	return performance
}

// ChannelPerformance representa performance de um canal
type ChannelPerformance struct {
	ChannelType     string        `json:"channel_type"`
	TotalMessages   int64         `json:"total_messages"`
	SuccessRate     float64       `json:"success_rate"`
	AverageLatency  time.Duration `json:"average_latency"`
	MessagesPerHour float64       `json:"messages_per_hour"`
}

// calculateMessagesPerHour calcula mensagens por hora
func (m *MessageMetrics) calculateMessagesPerHour(stats *ChannelStats) float64 {
	duration := time.Since(stats.LastReset)
	if duration.Hours() == 0 {
		return 0
	}

	total := stats.TotalSent + stats.TotalFailed
	return float64(total) / duration.Hours()
}

// ExportMetrics exporta métricas em formato estruturado
func (m *MessageMetrics) ExportMetrics() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	export := make(map[string]interface{})
	export["timestamp"] = time.Now()
	export["channels"] = make(map[string]interface{})

	for channelType, stats := range m.stats {
		channelData := map[string]interface{}{
			"total_sent":        stats.TotalSent,
			"total_failed":      stats.TotalFailed,
			"success_rate":      m.calculateSuccessRate(stats),
			"average_latency":   m.calculateAverageLatency(stats),
			"type_breakdown":    stats.TypeBreakdown,
			"failure_reasons":   stats.FailureReasons,
			"last_reset":        stats.LastReset,
			"messages_per_hour": m.calculateMessagesPerHour(stats),
		}
		export["channels"].(map[string]interface{})[channelType] = channelData
	}

	return export
}

// calculateSuccessRate calcula taxa de sucesso
func (m *MessageMetrics) calculateSuccessRate(stats *ChannelStats) float64 {
	total := stats.TotalSent + stats.TotalFailed
	if total == 0 {
		return 0
	}
	return float64(stats.TotalSent) / float64(total) * 100
}

// calculateAverageLatency calcula latência média
func (m *MessageMetrics) calculateAverageLatency(stats *ChannelStats) time.Duration {
	if stats.LatencyCount == 0 {
		return 0
	}
	return stats.LatencySum / time.Duration(stats.LatencyCount)
}

// StartPeriodicExport inicia exportação periódica de métricas
func (m *MessageMetrics) StartPeriodicExport(interval time.Duration, callback func(map[string]interface{})) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			metrics := m.ExportMetrics()
			callback(metrics)
		}
	}()
}

// GetHealthStatus retorna status de saúde baseado nas métricas
func (m *MessageMetrics) GetHealthStatus() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	health := map[string]interface{}{
		"status":   "healthy",
		"channels": make(map[string]string),
		"alerts":   []string{},
	}

	alerts := []string{}

	for channelType, stats := range m.stats {
		successRate := m.calculateSuccessRate(stats)
		avgLatency := m.calculateAverageLatency(stats)

		channelStatus := "healthy"

		// Verificar taxa de sucesso
		if successRate < 90 {
			channelStatus = "degraded"
			alerts = append(alerts, fmt.Sprintf("Channel %s has low success rate: %.2f%%", channelType, successRate))
		}

		// Verificar latência
		if avgLatency > 10*time.Second {
			channelStatus = "degraded"
			alerts = append(alerts, fmt.Sprintf("Channel %s has high latency: %v", channelType, avgLatency))
		}

		health["channels"].(map[string]string)[channelType] = channelStatus
	}

	health["alerts"] = alerts

	// Status geral
	if len(alerts) > 0 {
		health["status"] = "degraded"
	}

	return health
}
