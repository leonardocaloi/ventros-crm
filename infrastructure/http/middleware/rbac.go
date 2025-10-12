package middleware

import (
	"context"
	"net/http"

	"github.com/caloi/ventros-crm/internal/domain/crm/project_member"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ProjectMemberRepository interface para buscar project members
type ProjectMemberRepository interface {
	FindByProjectAndAgent(ctx context.Context, projectID uuid.UUID, agentID string) (*project_member.ProjectMember, error)
}

// RBACMiddleware middleware de verificação de permissões RBAC
type RBACMiddleware struct {
	repo   ProjectMemberRepository
	logger *logrus.Logger
}

// NewRBACMiddleware cria nova instância do middleware RBAC
func NewRBACMiddleware(repo ProjectMemberRepository, logger *logrus.Logger) *RBACMiddleware {
	return &RBACMiddleware{
		repo:   repo,
		logger: logger,
	}
}

// RequireProjectMember verifica se o usuário é membro do projeto
func (m *RBACMiddleware) RequireProjectMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user context
		userCtx, err := GetUserContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden", "message": "authentication required",
			})
			return
		}

		// Extract project_id
		projectIDStr := c.Param("project_id")
		if projectIDStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "bad_request", "message": "project_id required",
			})
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "bad_request", "message": "invalid project_id",
			})
			return
		}

		// Check membership
		member, err := m.repo.FindByProjectAndAgent(c.Request.Context(), projectID, userCtx.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden", "message": "access denied",
			})
			return
		}

		c.Set("project_member", member)
		c.Set("project_id", projectID)
		c.Next()
	}
}

// RequirePermission verifica permissão específica
func (m *RBACMiddleware) RequirePermission(permission project_member.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		memberInterface, exists := c.Get("project_member")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden", "message": "project membership required",
			})
			return
		}

		member := memberInterface.(*project_member.ProjectMember)
		if !member.HasPermission(permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden", "message": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}
