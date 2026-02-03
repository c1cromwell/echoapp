package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thechadcromwell/echoapp/pkg/cardano"
	"github.com/thechadcromwell/echoapp/pkg/utils"
)

// CredentialHandlers handles credential-related endpoints
type CredentialHandlers struct {
	client *cardano.Client
}

// NewCredentialHandlers creates new credential handlers
func NewCredentialHandlers(client *cardano.Client) *CredentialHandlers {
	return &CredentialHandlers{client: client}
}

// GetCredentialAuditTrail retrieves audit trail for a credential
// GET /api/v1/credentials/:credentialId/audit
func (h *CredentialHandlers) GetCredentialAuditTrail(c *gin.Context) {
	credentialID := c.Param("credentialId")

	trail, err := h.client.GetCredentialAuditTrail(c.Request.Context(), credentialID)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve audit trail", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"credential_id": credentialID,
		"event_count":   trail.EventCount,
		"audit_trail":   trail.Events,
		"last_event":    trail.LastEvent,
		"request_id":    c.GetString("request_id"),
	})
}

// RevokeCredential revokes a credential
// POST /api/v1/credentials/:credentialId/revoke
func (h *CredentialHandlers) RevokeCredential(c *gin.Context) {
	credentialID := c.Param("credentialId")

	var req struct {
		Reason string `json:"reason"`
		Actor  string `json:"actor"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	txHash, err := h.client.RevokeCredentialWithReason(c.Request.Context(), credentialID, req.Reason, req.Actor)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to revoke credential", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"credential_id": credentialID,
		"status":        "revoked",
		"reason":        req.Reason,
		"tx_hash":       txHash,
		"timestamp":     time.Now(),
		"request_id":    c.GetString("request_id"),
	})
}

// SuspendCredential suspends a credential
// POST /api/v1/credentials/:credentialId/suspend
func (h *CredentialHandlers) SuspendCredential(c *gin.Context) {
	credentialID := c.Param("credentialId")

	var req struct {
		Reason string `json:"reason"`
		Actor  string `json:"actor"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	txHash, err := h.client.SuspendCredential(c.Request.Context(), credentialID, req.Reason, req.Actor)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to suspend credential", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"credential_id": credentialID,
		"status":        "suspended",
		"reason":        req.Reason,
		"tx_hash":       txHash,
		"request_id":    c.GetString("request_id"),
	})
}

// RestoreCredential restores a suspended credential
// POST /api/v1/credentials/:credentialId/restore
func (h *CredentialHandlers) RestoreCredential(c *gin.Context) {
	credentialID := c.Param("credentialId")

	var req struct {
		Actor string `json:"actor"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	txHash, err := h.client.RestoreCredential(c.Request.Context(), credentialID, req.Actor)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to restore credential", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"credential_id": credentialID,
		"status":        "restored",
		"tx_hash":       txHash,
		"request_id":    c.GetString("request_id"),
	})
}

// AuthorizeAppAccess authorizes an app to access a credential
// POST /api/v1/credentials/:credentialId/authorize-app
func (h *CredentialHandlers) AuthorizeAppAccess(c *gin.Context) {
	credentialID := c.Param("credentialId")

	var req struct {
		AppID       string   `json:"app_id"`
		AppName     string   `json:"app_name"`
		Permissions []string `json:"permissions"`
		ExpiresIn   int      `json:"expires_in,omitempty"` // seconds
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var expiresIn *time.Duration
	if req.ExpiresIn > 0 {
		duration := time.Duration(req.ExpiresIn) * time.Second
		expiresIn = &duration
	}

	txHash, err := h.client.AuthorizeAppAccess(c.Request.Context(), credentialID, req.AppID, req.AppName, req.Permissions, expiresIn)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to authorize app access", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"credential_id": credentialID,
		"app_id":        req.AppID,
		"authorized":    true,
		"permissions":   req.Permissions,
		"tx_hash":       txHash,
		"authorized_at": time.Now(),
		"request_id":    c.GetString("request_id"),
	})
}

// RevokeAppAccess revokes an app's access to a credential
// POST /api/v1/credentials/:credentialId/revoke-app-access
func (h *CredentialHandlers) RevokeAppAccess(c *gin.Context) {
	credentialID := c.Param("credentialId")

	var req struct {
		AppID string `json:"app_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	txHash, err := h.client.RevokeAppAccess(c.Request.Context(), credentialID, req.AppID)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to revoke app access", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"credential_id": credentialID,
		"app_id":        req.AppID,
		"revoked":       true,
		"tx_hash":       txHash,
		"request_id":    c.GetString("request_id"),
	})
}

// CheckAppAccess checks if an app has access to a credential
// GET /api/v1/credentials/:credentialId/check-app-access
func (h *CredentialHandlers) CheckAppAccess(c *gin.Context) {
	credentialID := c.Param("credentialId")
	appID := c.Query("app_id")

	if appID == "" {
		utils.GinErrorResponse(c, http.StatusBadRequest, "app_id required", "")
		return
	}

	allowed, err := h.client.CheckAppAccess(c.Request.Context(), credentialID, appID)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Failed to check app access", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"credential_id": credentialID,
		"app_id":        appID,
		"allowed":       allowed,
		"request_id":    c.GetString("request_id"),
	})
}

// VerifyCredentialIntegrity verifies credential hasn't been tampered
// POST /api/v1/credentials/:credentialId/verify
func (h *CredentialHandlers) VerifyCredentialIntegrity(c *gin.Context) {
	credentialID := c.Param("credentialId")

	var req struct {
		ContentHash string `json:"content_hash"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	valid, err := h.client.VerifyCredentialIntegrity(c.Request.Context(), credentialID, req.ContentHash)
	if err != nil {
		utils.GinErrorResponse(c, http.StatusInternalServerError, "Verification failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"credential_id": credentialID,
		"valid":         valid,
		"request_id":    c.GetString("request_id"),
	})
}
