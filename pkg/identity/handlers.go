package identity

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers manages HTTP handlers for identity operations
type Handlers struct {
	service *Service
}

// NewHandlers creates new identity handlers
func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// RegisterRoutes registers all identity routes
func (h *Handlers) RegisterRoutes(router *gin.Engine) {
	router.POST("/identities", h.createIdentity)
	router.GET("/identities/:id", h.getIdentity)
	router.GET("/identities", h.listIdentities)
	router.GET("/trust-level/:userID", h.getTrustLevel)
	router.PUT("/trust-level/:userID", h.updateTrustLevel)
	router.POST("/credentials", h.storeCredential)
	router.GET("/credentials/:id", h.getCredential)
	router.GET("/credentials/user/:userID", h.getUserCredentials)
}

// createIdentity creates a new identity
func (h *Handlers) createIdentity(c *gin.Context) {
	var req struct {
		Method string `json:"method" binding:"required"`
		Name   string `json:"name"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":     "did:example:123",
		"method": req.Method,
		"name":   req.Name,
	})
}

// getIdentity retrieves an identity
func (h *Handlers) getIdentity(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id":     id,
		"method": "example",
	})
}

// listIdentities lists all identities
func (h *Handlers) listIdentities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"identities": []interface{}{},
	})
}

// getTrustLevel retrieves the trust level for a user
func (h *Handlers) getTrustLevel(c *gin.Context) {
	userID := c.Param("userID")

	level, err := h.service.GetTrustLevel(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID": userID,
		"level":  level,
	})
}

// updateTrustLevel updates a user's trust level
func (h *Handlers) updateTrustLevel(c *gin.Context) {
	userID := c.Param("userID")

	var req struct {
		NewLevel           string `json:"new_level" binding:"required"`
		VerificationMethod string `json:"verification_method"`
		VerifierID         string `json:"verifier_id"`
		Reason             string `json:"reason"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.service.UpdateTrustLevel(c.Request.Context(), userID, req.NewLevel, req.VerificationMethod, req.VerifierID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":   userID,
		"newLevel": req.NewLevel,
		"status":   "updated",
	})
}

// storeCredential stores a credential
func (h *Handlers) storeCredential(c *gin.Context) {
	var req struct {
		UserID       string                 `json:"user_id" binding:"required"`
		CredentialID string                 `json:"credential_id" binding:"required"`
		Data         map[string]interface{} `json:"data"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.service.StoreCredential(c.Request.Context(), req.UserID, req.CredentialID, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"credentialID": req.CredentialID,
		"userID":       req.UserID,
		"status":       "stored",
	})
}

// getCredential retrieves a credential
func (h *Handlers) getCredential(c *gin.Context) {
	credentialID := c.Param("id")

	cred, err := h.service.GetCredential(c.Request.Context(), credentialID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if cred == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credential": cred,
	})
}

// getUserCredentials retrieves all credentials for a user
func (h *Handlers) getUserCredentials(c *gin.Context) {
	userID := c.Param("userID")

	creds, err := h.service.GetUserCredentials(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":      userID,
		"credentials": creds,
	})
}
