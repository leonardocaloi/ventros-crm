package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// AuthClaims representa os claims do JWT do Keycloak
type AuthClaims struct {
	jwt.RegisteredClaims
	Email             string                 `json:"email"`
	EmailVerified     bool                   `json:"email_verified"`
	PreferredUsername string                 `json:"preferred_username"`
	GivenName         string                 `json:"given_name"`
	FamilyName        string                 `json:"family_name"`
	Name              string                 `json:"name"`
	RealmAccess       map[string]interface{} `json:"realm_access"`
	ResourceAccess    map[string]interface{} `json:"resource_access"`
}

// UserContext representa o contexto do usuário autenticado
type UserContext struct {
	Subject           string   // Keycloak user ID (sub claim)
	Email             string
	Name              string
	PreferredUsername string
	Roles             []string // System-level roles (customer, agent)
	TenantID          string   // Optional: se user pertence a um tenant específico
}

// Context keys
type contextKey string

const (
	UserContextKey contextKey = "user_context"
)

// JWTConfig configurações do middleware JWT
type JWTConfig struct {
	KeycloakURL   string
	Realm         string
	ClientID      string
	RequiredRoles []string // Roles necessários (customer, agent)
	Logger        *logrus.Logger
}

// JWTAuthMiddleware cria o middleware de autenticação JWT
func JWTAuthMiddleware(config JWTConfig) gin.HandlerFunc {
	// Initialize JWKS (JSON Web Key Set) client
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", config.KeycloakURL, config.Realm)

	// Keyfunc will automatically refresh keys every 5 minutes
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval:   5 * time.Minute,
		RefreshRateLimit:  time.Minute,
		RefreshTimeout:    10 * time.Second,
		RefreshErrorHandler: func(err error) {
			config.Logger.Errorf("JWKS refresh error: %v", err)
		},
	})
	if err != nil {
		config.Logger.Fatalf("Failed to get JWKS: %v", err)
	}

	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respondUnauthorized(c, "missing authorization header")
			return
		}

		// Validate Bearer format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondUnauthorized(c, "invalid authorization format")
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT
		token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, jwks.Keyfunc)
		if err != nil {
			config.Logger.Warnf("JWT parse error: %v", err)
			respondUnauthorized(c, "invalid token")
			return
		}

		// Validate token
		if !token.Valid {
			respondUnauthorized(c, "invalid token")
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*AuthClaims)
		if !ok {
			respondUnauthorized(c, "invalid token claims")
			return
		}

		// Validate required claims
		if err := validateClaims(claims, config); err != nil {
			config.Logger.Warnf("Claims validation error: %v", err)
			respondUnauthorized(c, "invalid token claims")
			return
		}

		// Extract roles from realm_access
		roles := extractRoles(claims)

		// Validate required roles (if configured)
		if len(config.RequiredRoles) > 0 {
			if !hasAnyRole(roles, config.RequiredRoles) {
				respondForbidden(c, "insufficient permissions")
				return
			}
		}

		// Create user context
		userCtx := &UserContext{
			Subject:           claims.Subject,
			Email:             claims.Email,
			Name:              claims.Name,
			PreferredUsername: claims.PreferredUsername,
			Roles:             roles,
		}

		// Store in Gin context
		c.Set(string(UserContextKey), userCtx)

		// Store in request context (for use in application layer)
		ctx := context.WithValue(c.Request.Context(), UserContextKey, userCtx)
		c.Request = c.Request.WithContext(ctx)

		// Log authentication (without sensitive data)
		config.Logger.WithFields(logrus.Fields{
			"user_id": userCtx.Subject,
			"email":   userCtx.Email,
			"roles":   userCtx.Roles,
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		}).Debug("User authenticated")

		c.Next()
	}
}

// OptionalJWTAuthMiddleware middleware JWT opcional (permite requests sem auth)
func OptionalJWTAuthMiddleware(config JWTConfig) gin.HandlerFunc {
	requiredMiddleware := JWTAuthMiddleware(config)

	return func(c *gin.Context) {
		// If no Authorization header, continue without auth
		if c.GetHeader("Authorization") == "" {
			c.Next()
			return
		}

		// If header present, validate it
		requiredMiddleware(c)
	}
}

// validateClaims valida os claims obrigatórios do token
func validateClaims(claims *AuthClaims, config JWTConfig) error {
	// Validate expiration
	if claims.ExpiresAt != nil {
		if time.Now().After(claims.ExpiresAt.Time) {
			return errors.New("token expired")
		}
	}

	// Validate issuer
	expectedIssuer := fmt.Sprintf("%s/realms/%s", config.KeycloakURL, config.Realm)
	if claims.Issuer != expectedIssuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Issuer)
	}

	// Validate audience (if client_id is configured)
	if config.ClientID != "" {
		if claims.Audience != nil {
			found := false
			for _, aud := range claims.Audience {
				if aud == config.ClientID {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid audience")
			}
		}
	}

	// Validate subject
	if claims.Subject == "" {
		return errors.New("missing subject claim")
	}

	return nil
}

// extractRoles extrai os roles do realm_access
func extractRoles(claims *AuthClaims) []string {
	roles := []string{}

	if claims.RealmAccess != nil {
		if rolesInterface, ok := claims.RealmAccess["roles"]; ok {
			if rolesSlice, ok := rolesInterface.([]interface{}); ok {
				for _, role := range rolesSlice {
					if roleStr, ok := role.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}
	}

	return roles
}

// hasAnyRole verifica se o usuário tem algum dos roles necessários
func hasAnyRole(userRoles, requiredRoles []string) bool {
	roleMap := make(map[string]bool)
	for _, role := range userRoles {
		roleMap[role] = true
	}

	for _, required := range requiredRoles {
		if roleMap[required] {
			return true
		}
	}

	return false
}

// GetUserContext extrai o UserContext do Gin context
func GetUserContext(c *gin.Context) (*UserContext, error) {
	value, exists := c.Get(string(UserContextKey))
	if !exists {
		return nil, errors.New("user context not found")
	}

	userCtx, ok := value.(*UserContext)
	if !ok {
		return nil, errors.New("invalid user context type")
	}

	return userCtx, nil
}

// MustGetUserContext extrai o UserContext ou panic (usar após middleware)
func MustGetUserContext(c *gin.Context) *UserContext {
	userCtx, err := GetUserContext(c)
	if err != nil {
		// This should never happen after JWT middleware
		panic("user context not found after authentication")
	}
	return userCtx
}

// GetUserContextFromRequest extrai UserContext do context.Context
func GetUserContextFromRequest(ctx context.Context) (*UserContext, error) {
	value := ctx.Value(UserContextKey)
	if value == nil {
		return nil, errors.New("user context not found")
	}

	userCtx, ok := value.(*UserContext)
	if !ok {
		return nil, errors.New("invalid user context type")
	}

	return userCtx, nil
}

// respondUnauthorized responde com 401 Unauthorized
func respondUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":   "unauthorized",
		"message": message,
	})
}

// respondForbidden responde com 403 Forbidden
func respondForbidden(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error":   "forbidden",
		"message": message,
	})
}
