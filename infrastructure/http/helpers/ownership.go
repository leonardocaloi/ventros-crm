package helpers

import (
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OwnershipHelper ajuda a verificar se recursos pertencem ao usuário autenticado
type OwnershipHelper struct{}

func NewOwnershipHelper() *OwnershipHelper {
	return &OwnershipHelper{}
}

// RequireAuth verifica se o usuário está autenticado
func (o *OwnershipHelper) RequireAuth(c *gin.Context) (*middleware.AuthContext, bool) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil, false
	}
	return authCtx, true
}

// ParseUUID helper para parsear UUID com erro padronizado
func (o *OwnershipHelper) ParseUUID(c *gin.Context, value, fieldName string) (uuid.UUID, bool) {
	if value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fieldName + " is required"})
		return uuid.Nil, false
	}

	parsedUUID, err := uuid.Parse(value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + fieldName + " format"})
		return uuid.Nil, false
	}

	return parsedUUID, true
}

// CheckProjectOwnership verifica se um projeto pertence ao usuário
// TODO: Implementar verificação real no banco
func (o *OwnershipHelper) CheckProjectOwnership(c *gin.Context, projectID uuid.UUID, userID uuid.UUID) bool {
	// TODO: Implementar consulta real ao banco
	// Por enquanto, sempre permite (para desenvolvimento)
	return true
}

// CheckContactOwnership verifica se um contato pertence ao usuário (via projeto)
// TODO: Implementar verificação real no banco
func (o *OwnershipHelper) CheckContactOwnership(c *gin.Context, contactID uuid.UUID, userID uuid.UUID) bool {
	// TODO: Implementar consulta real ao banco
	// Por enquanto, sempre permite (para desenvolvimento)
	return true
}

// CheckPipelineOwnership verifica se um pipeline pertence ao usuário (via projeto)
// TODO: Implementar verificação real no banco
func (o *OwnershipHelper) CheckPipelineOwnership(c *gin.Context, pipelineID uuid.UUID, userID uuid.UUID) bool {
	// TODO: Implementar consulta real ao banco
	// Por enquanto, sempre permite (para desenvolvimento)
	return true
}

// DenyAccess retorna erro de acesso negado
func (o *OwnershipHelper) DenyAccess(c *gin.Context, resourceType string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": resourceType + " not found or access denied",
		"hint":  "You can only access your own resources",
	})
}

// GetUserProjects retorna os projetos do usuário (mock para desenvolvimento)
func (o *OwnershipHelper) GetUserProjects(userID uuid.UUID) []map[string]interface{} {
	// TODO: Implementar consulta real ao banco
	// Mock para desenvolvimento
	return []map[string]interface{}{
		{
			"id":          uuid.New(),
			"name":        "Projeto Demo",
			"description": "Projeto de demonstração",
			"user_id":     userID,
			"active":      true,
		},
	}
}
