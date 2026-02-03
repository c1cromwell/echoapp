package credentials

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers provides HTTP handlers for credentials
type Handlers struct {
	service *Service
}

// NewHandlers creates new credential handlers
func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// RegisterRoutes registers credential routes
func (h *Handlers) RegisterRoutes(router *gin.Engine) {
	credGroup := router.Group("/api/v1/credentials")

	// Credential operations
	credGroup.POST("", h.IssueCredential)
	credGroup.GET("/:credentialId", h.GetCredential)
	credGroup.POST("/verify", h.VerifyCredential)
	credGroup.POST("/:credentialId/revoke", h.RevokeCredential)
	credGroup.GET("/:credentialId/status", h.GetCredentialStatus)
	credGroup.GET("/subject/:subjectDid", h.ListCredentialsBySubject)
	credGroup.POST("/:credentialId/convert", h.ConvertFormat)

	// Issuance progress
	credGroup.GET("/:credentialId/progress", h.GetIssuanceProgress)

	// Trust score
	credGroup.GET("/:credentialId/trust-score", h.GetTrustScore)

	// Batch operations
	credGroup.POST("/batch/verify", h.BatchVerifyCredentials)

	// Revocation operations
	revoGroup := router.Group("/api/v1/revocation")
	revoGroup.GET("/status/:credentialId", h.GetRevocationStatus)
	revoGroup.POST("/batch-check", h.BatchCheckRevocation)
	revoGroup.GET("/cache-stats", h.GetRevocationCacheStats)

	// Service health
	router.GET("/api/v1/health", h.HealthCheck)
	router.GET("/api/v1/component-status", h.GetComponentStatus)
}

// IssueCredential issues a new credential
// @POST /api/v1/credentials
// @Accept json
// @Produce json
func (h *Handlers) IssueCredential(c *gin.Context) {
	var req CredentialIssuanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid request format",
			"detail": err.Error(),
		})
		return
	}

	// Issue credential
	resp, err := h.service.IssueCredential(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetCredential retrieves a credential
// @GET /api/v1/credentials/:credentialId
// @Produce json
func (h *Handlers) GetCredential(c *gin.Context) {
	credentialID := c.Param("credentialId")

	vc, err := h.service.GetCredential(c.Request.Context(), credentialID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, vc)
}

// VerifyCredential verifies a credential
// @POST /api/v1/credentials/verify
// @Accept json
// @Produce json
func (h *Handlers) VerifyCredential(c *gin.Context) {
	var req CredentialVerificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	result, err := h.service.VerifyCredential(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// RevokeCredential revokes a credential
// @POST /api/v1/credentials/:credentialId/revoke
// @Accept json
func (h *Handlers) RevokeCredential(c *gin.Context) {
	credentialID := c.Param("credentialId")

	var req struct {
		IssuerDID  string `json:"issuerDid"`
		SubjectDID string `json:"subjectDid"`
		Reason     string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	err := h.service.RevokeCredential(c.Request.Context(), credentialID, req.IssuerDID, req.SubjectDID, req.Reason)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "revoked",
		"credentialId": credentialID,
	})
}

// GetCredentialStatus gets credential status
// @GET /api/v1/credentials/:credentialId/status
// @Produce json
func (h *Handlers) GetCredentialStatus(c *gin.Context) {
	credentialID := c.Param("credentialId")

	revStatus, err := h.service.CheckRevocationStatus(c.Request.Context(), credentialID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, revStatus)
}

// ListCredentialsBySubject lists credentials for a subject
// @GET /api/v1/credentials/subject/:subjectDid
// @Produce json
func (h *Handlers) ListCredentialsBySubject(c *gin.Context) {
	subjectDID := c.Param("subjectDid")

	credentials, err := h.service.ListCredentials(c.Request.Context(), subjectDID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subject":     subjectDID,
		"credentials": credentials,
		"count":       len(credentials),
	})
}

// ConvertFormat converts credential format
// @POST /api/v1/credentials/:credentialId/convert
// @Accept json
// @Produce json
func (h *Handlers) ConvertFormat(c *gin.Context) {
	credentialID := c.Param("credentialId")

	var req struct {
		Format     string `json:"format"`
		PrivateKey string `json:"privateKey"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	vc, err := h.service.GetCredential(c.Request.Context(), credentialID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	converted, err := h.service.ConvertCredentialFormat(vc, CredentialFormat(req.Format), req.PrivateKey)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credentialId": credentialID,
		"format":       req.Format,
		"credential":   converted,
	})
}

// GetIssuanceProgress gets issuance progress
// @GET /api/v1/credentials/:credentialId/progress
// @Produce json
func (h *Handlers) GetIssuanceProgress(c *gin.Context) {
	credentialID := c.Param("credentialId")

	progress := h.service.GetIssuanceProgress(credentialID)
	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "progress not found",
		})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetTrustScore gets trust score
// @GET /api/v1/credentials/:credentialId/trust-score
// @Produce json
func (h *Handlers) GetTrustScore(c *gin.Context) {
	credentialID := c.Param("credentialId")

	score, exists := h.service.GetTrustScore(credentialID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "trust score not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credentialId": credentialID,
		"trustScore":   score,
	})
}

// BatchVerifyCredentials verifies multiple credentials
// @POST /api/v1/credentials/batch/verify
// @Accept json
// @Produce json
func (h *Handlers) BatchVerifyCredentials(c *gin.Context) {
	var requests []CredentialVerificationRequest

	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	results, err := h.service.verifier.BatchVerify(c.Request.Context(), requests)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// GetRevocationStatus gets revocation status
// @GET /api/v1/revocation/status/:credentialId
// @Produce json
func (h *Handlers) GetRevocationStatus(c *gin.Context) {
	credentialID := c.Param("credentialId")

	status, err := h.service.CheckRevocationStatus(c.Request.Context(), credentialID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}

// BatchCheckRevocation checks revocation for multiple credentials
// @POST /api/v1/revocation/batch-check
// @Accept json
// @Produce json
func (h *Handlers) BatchCheckRevocation(c *gin.Context) {
	var req struct {
		CredentialIDs []string `json:"credentialIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	results, err := h.service.revocationManager.BatchCheckRevocation(c.Request.Context(), req.CredentialIDs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetRevocationCacheStats gets revocation cache stats
// @GET /api/v1/revocation/cache-stats
// @Produce json
func (h *Handlers) GetRevocationCacheStats(c *gin.Context) {
	stats := h.service.GetRevocationCacheStats()
	c.JSON(http.StatusOK, stats)
}

// HealthCheck checks service health
// @GET /api/v1/health
// @Produce json
func (h *Handlers) HealthCheck(c *gin.Context) {
	err := h.service.GetStorageHealth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// GetComponentStatus gets component status
// @GET /api/v1/component-status
// @Produce json
func (h *Handlers) GetComponentStatus(c *gin.Context) {
	status := h.service.GetComponentStatus(c.Request.Context())
	c.JSON(http.StatusOK, status)
}

// handleError handles errors and returns appropriate HTTP status
func (h *Handlers) handleError(c *gin.Context, err error) {
	if credErr, ok := err.(*CredentialError); ok {
		statusCode := h.credentialErrorToStatusCode(credErr.Code)
		c.JSON(statusCode, gin.H{
			"error":   string(credErr.Code),
			"message": credErr.Message,
			"details": credErr.Details,
		})
		return
	}

	if validErrs, ok := err.(ValidationErrors); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation error",
			"details": validErrs,
		})
		return
	}

	// Generic error
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "internal server error",
		"message": err.Error(),
	})
}

// credentialErrorToStatusCode maps credential error codes to HTTP status codes
func (h *Handlers) credentialErrorToStatusCode(code CredentialErrorCode) int {
	switch code {
	case ErrCodeInvalidCredential, ErrCodeInvalidFormat, ErrCodeInvalidRequest:
		return http.StatusBadRequest
	case ErrCodeCredentialNotFound:
		return http.StatusNotFound
	case ErrCodeRevokedCredential:
		return http.StatusForbidden
	case ErrCodeExpiredCredential:
		return http.StatusForbidden
	case ErrCodeTimeoutError:
		return http.StatusGatewayTimeout
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeStorageFailed, ErrCodeBlockchainError:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
