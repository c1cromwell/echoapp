package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thechadcromwell/echoapp/pkg/cardano"
	"github.com/thechadcromwell/echoapp/pkg/utils"
)

// SchemaHandlers handles credential schema endpoints
type SchemaHandlers struct {
	client *cardano.Client
}

// NewSchemaHandlers creates new schema handlers
func NewSchemaHandlers(client *cardano.Client) *SchemaHandlers {
	return &SchemaHandlers{client: client}
}

// CreateSchema creates a new credential schema
// POST /api/v1/schemas
func (h *SchemaHandlers) CreateSchema(c *gin.Context) {
	var schema cardano.CredentialSchema
	if err := c.ShouldBindJSON(&schema); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid schema", err.Error())
		return
	}

	schema.SchemaID = fmt.Sprintf("schema_%d", time.Now().UnixNano())
	schema.CreatedAt = time.Now()
	schema.UpdatedAt = time.Now()
	schema.Status = "active"

	txHash, err := h.client.StoreSchema(c.Request.Context(), &schema)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to store schema", err.Error())
		return
	}

	response := gin.H{
		"success":    true,
		"schema_id":  schema.SchemaID,
		"version":    schema.Version,
		"tx_hash":    txHash,
		"timestamp":  time.Now(),
		"request_id": c.GetString("request_id"),
	}

	c.JSON(http.StatusCreated, response)
}

// GetSchema retrieves a schema
// GET /api/v1/schemas/:schemaId
func (h *SchemaHandlers) GetSchema(c *gin.Context) {
	schemaID := c.Param("schemaId")
	version := 1
	if v := c.Query("version"); v != "" {
		fmt.Sscanf(v, "%d", &version)
	}

	schema, err := h.client.GetSchema(c.Request.Context(), schemaID, version)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve schema", err.Error())
		return
	}

	if schema == nil {
		utils.GinErrorResponse(c, http.StatusNotFound, "Schema not found", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"schema":     schema,
		"request_id": c.GetString("request_id"),
	})
}

// UpdateSchema updates a schema (creates new version)
// PUT /api/v1/schemas/:schemaId
func (h *SchemaHandlers) UpdateSchema(c *gin.Context) {
	schemaID := c.Param("schemaId")

	var schema cardano.CredentialSchema
	if err := c.ShouldBindJSON(&schema); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid schema update", err.Error())
		return
	}

	schema.SchemaID = schemaID
	schema.UpdatedAt = time.Now()

	txHash, err := h.client.UpdateSchema(c.Request.Context(), &schema)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to update schema", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"schema_id":  schemaID,
		"version":    schema.Version,
		"tx_hash":    txHash,
		"timestamp":  time.Now(),
		"request_id": c.GetString("request_id"),
	})
}

// DeprecateSchema marks a schema as deprecated
// POST /api/v1/schemas/:schemaId/deprecate
func (h *SchemaHandlers) DeprecateSchema(c *gin.Context) {
	schemaID := c.Param("schemaId")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	txHash, err := h.client.DeprecateSchema(c.Request.Context(), schemaID, req.Reason)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to deprecate schema", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"schema_id":  schemaID,
		"action":     "deprecated",
		"tx_hash":    txHash,
		"timestamp":  time.Now(),
		"request_id": c.GetString("request_id"),
	})
}

// GetSchemaVersionHistory retrieves all versions of a schema
// GET /api/v1/schemas/:schemaId/history
func (h *SchemaHandlers) GetSchemaVersionHistory(c *gin.Context) {
	schemaID := c.Param("schemaId")

	history, err := h.client.GetSchemaVersionHistory(c.Request.Context(), schemaID)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve schema history", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"schema_id":     schemaID,
		"version_count": len(history),
		"versions":      history,
		"request_id":    c.GetString("request_id"),
	})
}

// QuerySchemas queries schemas with filters
// GET /api/v1/schemas
func (h *SchemaHandlers) QuerySchemas(c *gin.Context) {
	query := &cardano.SchemaQuery{
		SchemaID:  c.Query("schema_id"),
		CreatedBy: c.Query("created_by"),
		Status:    c.Query("status"),
		Limit:     100,
		Offset:    0,
	}

	result, err := h.client.QuerySchemas(c.Request.Context(), query)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to query schemas", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"total":      result.Total,
		"count":      result.Count,
		"schemas":    result.Schemas,
		"query_time": result.QueryTime.String(),
		"request_id": c.GetString("request_id"),
	})
}

// ValidateCredentialAgainstSchema validates a credential
// POST /api/v1/schemas/:schemaId/validate
func (h *SchemaHandlers) ValidateCredentialAgainstSchema(c *gin.Context) {
	schemaID := c.Param("schemaId")

	var req struct {
		Credential map[string]interface{} `json:"credential"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	schema, err := h.client.GetSchema(c.Request.Context(), schemaID, 1)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve schema", err.Error())
		return
	}

	credential := &cardano.Credential{
		Data: req.Credential,
	}

	valid, errors, err := h.client.ValidateCredentialAgainstSchema(c.Request.Context(), credential, schema)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Validation error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"valid":      valid,
		"errors":     errors,
		"request_id": c.GetString("request_id"),
	})
}
