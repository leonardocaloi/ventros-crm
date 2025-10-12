package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitConfig configurações de rate limiting
type RateLimitConfig struct {
	// Requests per period (e.g., "100-M" = 100 requests per minute)
	Rate string

	// Custom key extractor (default: IP address)
	KeyExtractor func(*gin.Context) string
}

// RateLimitMiddleware cria middleware de rate limiting
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	// Parse rate string (format: "requests-period")
	// Examples: "100-M" (100/min), "10-S" (10/sec), "1000-H" (1000/hour)
	rate, err := limiter.NewRateFromFormatted(config.Rate)
	if err != nil {
		panic(fmt.Sprintf("invalid rate format: %s", config.Rate))
	}

	// Create in-memory store
	// Production: Use Redis store for distributed systems
	store := memory.NewStore()

	// Create limiter instance
	instance := limiter.New(store, rate)

	// Key extractor: default is IP-based
	var keyExtractor mgin.KeyGetter
	if config.KeyExtractor != nil {
		keyExtractor = func(c *gin.Context) string {
			return config.KeyExtractor(c)
		}
	} else {
		keyExtractor = mgin.DefaultKeyGetter
	}

	// Create middleware
	middleware := mgin.NewMiddleware(instance, mgin.WithKeyGetter(keyExtractor))

	return func(c *gin.Context) {
		middleware(c)

		// If rate limit exceeded, mgin already set the status
		// We just need to ensure the response format is consistent
		if c.Writer.Status() == http.StatusTooManyRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}

// GlobalRateLimitMiddleware rate limit global (IP-based)
// Recommended: 100 requests per minute per IP
func GlobalRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		Rate: "100-M", // 100 requests per minute
	})
}

// AuthRateLimitMiddleware rate limit para endpoints de autenticação
// Mais restritivo para prevenir brute force
// Recommended: 10 requests per minute per IP
func AuthRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		Rate: "10-M", // 10 requests per minute
	})
}

// UserBasedRateLimitMiddleware rate limit por usuário autenticado
// Use após JWT middleware
func UserBasedRateLimitMiddleware(rate string) gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		Rate: rate,
		KeyExtractor: func(c *gin.Context) string {
			// Try to get user context
			userCtx, err := GetUserContext(c)
			if err != nil {
				// Fallback to IP if user not authenticated
				return c.ClientIP()
			}
			return fmt.Sprintf("user:%s", userCtx.Subject)
		},
	})
}
