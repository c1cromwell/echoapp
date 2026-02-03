package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thechadcromwell/echoapp/pkg/cardano"
	"github.com/thechadcromwell/echoapp/pkg/utils"
)

// IssuerHandlers handles issuer management endpoints
type IssuerHandlers struct {
	client *cardano.Client
}

// NewIssuerHandlers creates new issuer handlers
func NewIssuerHandlers(client *cardano.Client) *IssuerHandlers {
	return &IssuerHandlers{client: client}
}

// RegisterIssuer registers a new issuer
// POST /api/v1/issuers
func (h *IssuerHandlers) RegisterIssuer(c *gin.Context) {
	var issuer cardano.IssuerRegistration
	if err := c.ShouldBindJSON(&issuer); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid issuer data", err.Error())
		return
	}

	txHash, err := h.client.RegisterIssuer(c.Request.Context(), &issuer)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to register issuer", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":         true,
		"issuer_did":      issuer.IssuerDID,
		"status":          issuer.Status,
		"authority_level": issuer.AuthorityLevel,
		"tx_hash":         txHash,
		"registered_at":   issuer.RegisteredAt,
		"request_id":      c.GetString("request_id"),
	})
}

// GetIssuer retrieves issuer information
// GET /api/v1/issuers/:issuerId
func (h *IssuerHandlers) GetIssuer(c *gin.Context) {
	issuerID := c.Param("issuerId")

	issuer, err := h.client.GetIssuer(c.Request.Context(), issuerID)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve issuer", err.Error())
		return
	}

	if issuer == nil {
		utils.GinErrorResponse(c, http.StatusNotFound, "Issuer not found", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"issuer":     issuer,
		"request_id": c.GetString("request_id"),
	})
}

// VerifyIssuer verifies an issuer
// POST /api/v1/issuers/:issuerId/verify
func (h *IssuerHandlers) VerifyIssuer(c *gin.Context) {
	issuerID := c.Param("issuerId")

	var req struct {
		VerificationURL string `json:"verification_url"`
		AuthorityLevel  string `json:"authority_level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	txHash, err := h.client.VerifyIssuer(c.Request.Context(), issuerID, req.VerificationURL, cardano.AuthorityLevel(req.AuthorityLevel))
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to verify issuer", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"issuer_did":  issuerID,
		"status":      "verified",
		"tx_hash":     txHash,
		"verified_at": time.Now(),
		"request_id":  c.GetString("request_id"),
	})
}

// SuspendIssuer suspends an issuer
// POST /api/v1/issuers/:issuerId/suspend
func (h *IssuerHandlers) SuspendIssuer(c *gin.Context) {
	issuerID := c.Param("issuerId")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	txHash, err := h.client.SuspendIssuer(c.Request.Context(), issuerID, req.Reason)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to suspend issuer", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"issuer_did": issuerID,
		"status":     "suspended",
		"reason":     req.Reason,
		"tx_hash":    txHash,
		"request_id": c.GetString("request_id"),
	})
}

// RevokeIssuer revokes an issuer
// POST /api/v1/issuers/:issuerId/revoke
func (h *IssuerHandlers) RevokeIssuer(c *gin.Context) {
	issuerID := c.Param("issuerId")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	txHash, err := h.client.RevokeIssuer(c.Request.Context(), issuerID, req.Reason)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to revoke issuer", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"issuer_did": issuerID,
		"status":     "revoked",
		"reason":     req.Reason,
		"tx_hash":    txHash,
		"request_id": c.GetString("request_id"),
	})
}

// GetIssuerAuditTrail retrieves audit trail for an issuer
// GET /api/v1/issuers/:issuerId/audit
func (h *IssuerHandlers) GetIssuerAuditTrail(c *gin.Context) {
	issuerID := c.Param("issuerId")

	trail, err := h.client.GetIssuerAuditTrail(c.Request.Context(), issuerID)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve audit trail", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"issuer_did":  issuerID,
		"audit_count": len(trail),
		"audit_trail": trail,
		"request_id":  c.GetString("request_id"),
	})
}

// GetIssuerCredentials retrieves all credentials issued by an issuer
// GET /api/v1/issuers/:issuerId/credentials
func (h *IssuerHandlers) GetIssuerCredentials(c *gin.Context) {
	issuerID := c.Param("issuerId")

	credentials, err := h.client.GetIssuerCredentials(c.Request.Context(), issuerID)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve credentials", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"issuer_did":       issuerID,
		"credential_count": len(credentials),
		"credentials":      credentials,
		"request_id":       c.GetString("request_id"),
	})
}

// UpdateIssuerMetadata updates issuer metadata
// PUT /api/v1/issuers/:issuerId/metadata
func (h *IssuerHandlers) UpdateIssuerMetadata(c *gin.Context) {
	issuerID := c.Param("issuerId")

	var metadata map[string]interface{}
	if err := c.ShouldBindJSON(&metadata); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid metadata", err.Error())
		return
	}

	txHash, err := h.client.UpdateIssuerMetadata(c.Request.Context(), issuerID, metadata)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to update metadata", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"issuer_did": issuerID,
		"metadata":   metadata,
		"tx_hash":    txHash,
		"request_id": c.GetString("request_id"),
	})
}
