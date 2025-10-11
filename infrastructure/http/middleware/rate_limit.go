package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiterConfig configuração do rate limiter
type RateLimiterConfig struct {
	// MaxRequests número máximo de requisições permitidas
	MaxRequests int
	// Window duração da janela de tempo
	Window time.Duration
	// KeyPrefix prefixo da chave no Redis
	KeyPrefix string
}

// RateLimiter implementa rate limiting baseado em Redis
type RateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewRateLimiter cria um novo rate limiter
func NewRateLimiter(redisClient *redis.Client, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		logger: logger,
	}
}

// RateLimitMiddleware cria um middleware de rate limiting por IP
func (rl *RateLimiter) RateLimitMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Se Redis não está disponível, permitir todas as requisições
		if rl.redis == nil {
			c.Next()
			return
		}

		// Obter IP do cliente
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s:%s", config.KeyPrefix, clientIP)

		// Verificar e incrementar contador
		allowed, remaining, resetAt, err := rl.checkAndIncrement(key, config.MaxRequests, config.Window)
		if err != nil {
			rl.logger.Error("Rate limit check failed",
				zap.Error(err),
				zap.String("client_ip", clientIP))
			// Em caso de erro, permitir a requisição (fail open)
			c.Next()
			return
		}

		// Adicionar headers informativos
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.MaxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if !allowed {
			retryAfter := time.Until(resetAt).Seconds()
			c.Header("Retry-After", strconv.Itoa(int(retryAfter)))

			rl.logger.Warn("Rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.String("key_prefix", config.KeyPrefix),
				zap.Int("max_requests", config.MaxRequests))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Maximum %d requests per %v. Try again in %d seconds.",
					config.MaxRequests, config.Window, int(retryAfter)),
				"retry_after": int(retryAfter),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUserMiddleware cria um middleware de rate limiting por usuário autenticado
func (rl *RateLimiter) RateLimitByUserMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Se Redis não está disponível, permitir todas as requisições
		if rl.redis == nil {
			c.Next()
			return
		}

		// Obter user_id do contexto (definido pelo auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// Se não autenticado, usar IP
			rl.RateLimitMiddleware(config)(c)
			return
		}

		key := fmt.Sprintf("%s:%v", config.KeyPrefix, userID)

		// Verificar e incrementar contador
		allowed, remaining, resetAt, err := rl.checkAndIncrement(key, config.MaxRequests, config.Window)
		if err != nil {
			rl.logger.Error("Rate limit check failed",
				zap.Error(err),
				zap.String("user_id", fmt.Sprint(userID)))
			// Em caso de erro, permitir a requisição (fail open)
			c.Next()
			return
		}

		// Adicionar headers informativos
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.MaxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if !allowed {
			retryAfter := time.Until(resetAt).Seconds()
			c.Header("Retry-After", strconv.Itoa(int(retryAfter)))

			rl.logger.Warn("Rate limit exceeded",
				zap.String("user_id", fmt.Sprint(userID)),
				zap.String("key_prefix", config.KeyPrefix),
				zap.Int("max_requests", config.MaxRequests))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Maximum %d requests per %v. Try again in %d seconds.",
					config.MaxRequests, config.Window, int(retryAfter)),
				"retry_after": int(retryAfter),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkAndIncrement verifica e incrementa o contador de rate limiting
// Retorna: (allowed bool, remaining int, resetAt time.Time, error)
func (rl *RateLimiter) checkAndIncrement(key string, maxRequests int, window time.Duration) (bool, int, time.Time, error) {
	ctx := context.Background()

	// Incrementar contador
	count, err := rl.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, time.Time{}, err
	}

	// Se é a primeira requisição, definir TTL
	if count == 1 {
		if err := rl.redis.Expire(ctx, key, window).Err(); err != nil {
			return false, 0, time.Time{}, err
		}
	}

	// Obter TTL para calcular reset time
	ttl, err := rl.redis.TTL(ctx, key).Result()
	if err != nil {
		return false, 0, time.Time{}, err
	}

	resetAt := time.Now().Add(ttl)
	remaining := maxRequests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	allowed := count <= int64(maxRequests)
	return allowed, remaining, resetAt, nil
}
