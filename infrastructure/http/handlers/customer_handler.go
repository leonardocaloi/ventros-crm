package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CustomerHandler struct {
	logger *zap.Logger
}

func NewCustomerHandler(logger *zap.Logger) *CustomerHandler {
	return &CustomerHandler{
		logger: logger,
	}
}

// CreateCustomerRequest representa o payload para criar um cliente
type CreateCustomerRequest struct {
	Name         string                 `json:"name" binding:"required" example:"Empresa ABC"`
	Email        string                 `json:"email" example:"contato@empresa.com"`
	Phone        string                 `json:"phone" example:"+5511999999999"`
	Document     string                 `json:"document" example:"12.345.678/0001-90"`
	DocumentType string                 `json:"document_type" example:"cnpj"`
	Address      string                 `json:"address" example:"Rua das Flores, 123"`
	City         string                 `json:"city" example:"São Paulo"`
	State        string                 `json:"state" example:"SP"`
	Country      string                 `json:"country" example:"Brasil"`
	PostalCode   string                 `json:"postal_code" example:"01234-567"`
	Website      string                 `json:"website" example:"https://empresa.com"`
	Industry     string                 `json:"industry" example:"Tecnologia"`
	Size         string                 `json:"size" example:"medium"`
	Revenue      float64                `json:"revenue" example:"1000000.00"`
	TenantID     string                 `json:"tenant_id" binding:"required" example:"tenant_123"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateCustomerRequest representa o payload para atualizar um cliente
type UpdateCustomerRequest struct {
	Name         *string                `json:"name,omitempty"`
	Email        *string                `json:"email,omitempty"`
	Phone        *string                `json:"phone,omitempty"`
	Document     *string                `json:"document,omitempty"`
	DocumentType *string                `json:"document_type,omitempty"`
	Address      *string                `json:"address,omitempty"`
	City         *string                `json:"city,omitempty"`
	State        *string                `json:"state,omitempty"`
	Country      *string                `json:"country,omitempty"`
	PostalCode   *string                `json:"postal_code,omitempty"`
	Website      *string                `json:"website,omitempty"`
	Industry     *string                `json:"industry,omitempty"`
	Size         *string                `json:"size,omitempty"`
	Active       *bool                  `json:"active,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ListCustomers lists all customers with optional filters
// DEPRECATED: Customer endpoint not used in this CRM context
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	// TODO: Implement proper customer listing with filters
	c.JSON(http.StatusOK, gin.H{
		"message":    "Customer listing not yet implemented",
		"note":       "Use GET /api/v1/customers/{id} to get specific customer",
		"deprecated": true,
	})
}

// CreateCustomer creates a new customer
// DEPRECATED: Customer endpoint not used in this CRM context
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse customer request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implement customer creation
	c.JSON(http.StatusCreated, gin.H{
		"message":       "Customer creation not yet implemented",
		"name":          req.Name,
		"email":         req.Email,
		"document":      req.Document,
		"document_type": req.DocumentType,
		"industry":      req.Industry,
		"tenant_id":     req.TenantID,
	})
}

// GetCustomer gets a customer by ID
// DEPRECATED: Customer endpoint not used in this CRM context
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	idStr := c.Param("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid customer ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID format"})
		return
	}

	// TODO: Implement customer retrieval
	c.JSON(http.StatusOK, gin.H{
		"message":     "Customer retrieval not yet implemented",
		"customer_id": customerID,
	})
}

// UpdateCustomer updates a customer
// DEPRECATED: Customer endpoint not used in this CRM context
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID format"})
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implement customer update
	c.JSON(http.StatusOK, gin.H{
		"message":     "Customer update not yet implemented",
		"customer_id": customerID,
	})
}

// DeleteCustomer deletes a customer
// DEPRECATED: Customer endpoint not used in this CRM context
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID format"})
		return
	}

	// TODO: Implement customer deletion
	c.JSON(http.StatusOK, gin.H{
		"message":     "Customer deletion not yet implemented",
		"customer_id": customerID,
	})
}
