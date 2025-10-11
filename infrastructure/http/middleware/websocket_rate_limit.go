package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// WebSocketRateLimiter implementa rate limiting para conexões WebSocket
type WebSocketRateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger

	// Fallback: in-memory rate limiting se Redis falhar
	connections map[string]*connectionTracker
	mu          sync.RWMutex
}

type connectionTracker struct {
	count       int
	windowStart time.Time
}

// NewWebSocketRateLimiter cria novo rate limiter
func NewWebSocketRateLimiter(redis *redis.Client, logger *zap.Logger) *WebSocketRateLimiter {
	limiter := &WebSocketRateLimiter{
		redis:       redis,
		logger:      logger,
		connections: make(map[string]*connectionTracker),
	}

	// Cleanup goroutine para in-memory tracker
	go limiter.cleanupExpired()

	return limiter
}

// RateLimit middleware para WebSocket
func (rl *WebSocketRateLimiter) RateLimit(maxConnections int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Tentar rate limit via Redis primeiro
		if rl.redis != nil {
			allowed, err := rl.checkRedis(c.Request.Context(), clientIP, maxConnections, window)
			if err != nil {
				rl.logger.Warn("Redis rate limit check failed, falling back to in-memory",
					zap.Error(err),
					zap.String("client_ip", clientIP))
				// Fallback para in-memory
				if !rl.checkInMemory(clientIP, maxConnections, window) {
					rl.rateLimitExceeded(c, clientIP)
					return
				}
			} else if !allowed {
				rl.rateLimitExceeded(c, clientIP)
				return
			}
		} else {
			// Usar in-memory se Redis não disponível
			if !rl.checkInMemory(clientIP, maxConnections, window) {
				rl.rateLimitExceeded(c, clientIP)
				return
			}
		}

		c.Next()
	}
}

// checkRedis verifica rate limit usando Redis
func (rl *WebSocketRateLimiter) checkRedis(ctx context.Context, clientIP string, maxConnections int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ws:ratelimit:%s", clientIP)

	// Incrementar contador
	pipe := rl.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := incr.Val()

	rl.logger.Debug("WebSocket rate limit check (Redis)",
		zap.String("client_ip", clientIP),
		zap.Int64("current_count", count),
		zap.Int("max_connections", maxConnections))

	return count <= int64(maxConnections), nil
}

// checkInMemory verifica rate limit usando memória local
func (rl *WebSocketRateLimiter) checkInMemory(clientIP string, maxConnections int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	tracker, exists := rl.connections[clientIP]

	if !exists || now.Sub(tracker.windowStart) > window {
		// Nova janela
		rl.connections[clientIP] = &connectionTracker{
			count:       1,
			windowStart: now,
		}
		return true
	}

	// Mesma janela
	if tracker.count >= maxConnections {
		rl.logger.Warn("WebSocket rate limit exceeded (in-memory)",
			zap.String("client_ip", clientIP),
			zap.Int("current_count", tracker.count),
			zap.Int("max_connections", maxConnections))
		return false
	}

	tracker.count++
	return true
}

// rateLimitExceeded retorna erro HTTP quando rate limit excedido
func (rl *WebSocketRateLimiter) rateLimitExceeded(c *gin.Context, clientIP string) {
	rl.logger.Warn("WebSocket connection rejected - rate limit exceeded",
		zap.String("client_ip", clientIP),
		zap.String("user_agent", c.GetHeader("User-Agent")))

	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":       "Rate limit exceeded",
		"message":     "Too many WebSocket connection attempts. Please try again later.",
		"retry_after": "60s",
	})
	c.Abort()
}

// cleanupExpired remove trackers expirados da memória
func (rl *WebSocketRateLimiter) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, tracker := range rl.connections {
			if now.Sub(tracker.windowStart) > 5*time.Minute {
				delete(rl.connections, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// GetStats retorna estatísticas de rate limiting
func (rl *WebSocketRateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"in_memory_trackers": len(rl.connections),
		"redis_enabled":      rl.redis != nil,
	}
}
