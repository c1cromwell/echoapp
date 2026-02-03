package did

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handlers provides HTTP handlers for DID operations
type Handlers struct {
	service *Service
}

// NewHandlers creates a new handlers instance
func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// CreateDID handles DID creation requests
// POST /v1/dids
func (h *Handlers) CreateDID(c *gin.Context) {
	var req DIDCreationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.service.CreateDID(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ResolveDID handles DID resolution requests
// GET /v1/dids/:did
func (h *Handlers) ResolveDID(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID is required",
		})
		return
	}

	// Check if metadata is requested
	withMetadata := c.DefaultQuery("metadata", "false") == "true"

	if withMetadata {
		doc, metadata, err := h.service.ResolveDIDWithMetadata(c.Request.Context(), did)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, &DIDResolutionResponse{
			DIDDocument: doc,
			Metadata:    *metadata,
		})
		return
	}

	doc, err := h.service.ResolveDID(c.Request.Context(), did)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document": doc,
	})
}

// UpdateDID handles DID update requests
// PUT /v1/dids/:did
func (h *Handlers) UpdateDID(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID is required",
		})
		return
	}

	var document DIDDocument
	if err := c.ShouldBindJSON(&document); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.UpdateDID(c.Request.Context(), did, &document); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "DID updated successfully",
		"did":     did,
	})
}

// GetDIDMapping handles requests to retrieve DID mapping
// GET /v1/dids/:did/mapping
func (h *Handlers) GetDIDMapping(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID is required",
		})
		return
	}

	mapping, err := h.service.GetDIDMapping(c.Request.Context(), did)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapping)
}

// GetDIDByUserID handles requests to retrieve DID by user ID
// GET /v1/users/:userId/did
func (h *Handlers) GetDIDByUserID(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	mapping, err := h.service.GetDIDMappingByUserID(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapping)
}

// ListDevices handles requests to list devices for a DID
// GET /v1/dids/:did/devices
func (h *Handlers) ListDevices(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID is required",
		})
		return
	}

	devices, err := h.service.GetDevices(did)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
	})
}

// RegisterDevice handles device registration requests
// POST /v1/dids/:did/devices
func (h *Handlers) RegisterDevice(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID is required",
		})
		return
	}

	var device DeviceRegistration
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.RegisterDevice(did, &device); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Device registered successfully",
		"device":  device,
	})
}

// UnregisterDevice handles device unregistration requests
// DELETE /v1/dids/:did/devices/:deviceId
func (h *Handlers) UnregisterDevice(c *gin.Context) {
	did := c.Param("did")
	deviceID := c.Param("deviceId")

	if did == "" || deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID and device ID are required",
		})
		return
	}

	if err := h.service.UnregisterDevice(did, deviceID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device unregistered successfully",
	})
}

// InitiateDeviceRegistration handles device registration initiation
// POST /v1/dids/:did/devices/register/initiate
func (h *Handlers) InitiateDeviceRegistration(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID is required",
		})
		return
	}

	pending, err := h.service.InitiateDeviceRegistration(did)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device_id":  pending.DeviceID,
		"challenge":  pending.Challenge,
		"expires_at": pending.ExpiresAt,
	})
}

// GenerateQRCodeForDeviceRegistration generates a QR code for device registration
// POST /v1/dids/:did/devices/register/qrcode
func (h *Handlers) GenerateQRCodeForDeviceRegistration(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID is required",
		})
		return
	}

	qrData, qrCode, err := h.service.GenerateQRCodeForDeviceRegistration(did)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"qr_data":        qrData,
		"qr_code_base64": qrCode,
	})
}

// CompleteDeviceRegistration completes the device registration flow
// POST /v1/devices/register/complete
func (h *Handlers) CompleteDeviceRegistration(c *gin.Context) {
	var req struct {
		DeviceID   string `json:"deviceId" binding:"required"`
		Challenge  string `json:"challenge" binding:"required"`
		PublicKey  string `json:"publicKey" binding:"required"`
		DeviceName string `json:"deviceName"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	device, err := h.service.CompleteDeviceRegistration(req.DeviceID, req.Challenge, req.PublicKey, req.DeviceName)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device registration completed",
		"device":  device,
	})
}

// VerifyDIDDocument handles DID document verification requests
// POST /v1/dids/verify
func (h *Handlers) VerifyDIDDocument(c *gin.Context) {
	var document DIDDocument
	if err := c.ShouldBindJSON(&document); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	valid, err := h.service.VerifyDIDDocument(c.Request.Context(), &document)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": valid,
	})
}

// GetGenerationProgress handles requests to get DID generation progress
// GET /v1/dids/:did/generation/progress
func (h *Handlers) GetGenerationProgress(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID is required",
		})
		return
	}

	progress := h.service.GetGenerationProgress(did)
	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No generation progress found for DID",
		})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// InvalidateCache handles cache invalidation requests
// POST /v1/cache/invalidate/:did
func (h *Handlers) InvalidateCache(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID is required",
		})
		return
	}

	if err := h.service.InvalidateCache(did); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache invalidated successfully",
	})
}

// ClearCache handles cache clearing requests
// POST /v1/cache/clear
func (h *Handlers) ClearCache(c *gin.Context) {
	if err := h.service.ClearCache(); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
	})
}

// GetCacheStats handles requests to get cache statistics
// GET /v1/cache/stats
func (h *Handlers) GetCacheStats(c *gin.Context) {
	stats := h.service.GetCacheStats()
	c.JSON(http.StatusOK, stats)
}

// Health handles health check requests
// GET /v1/health
func (h *Handlers) Health(c *gin.Context) {
	healthy, err := h.service.Health(c.Request.Context())

	response := &HealthCheckResponse{
		Timestamp:           time.Now(),
		AtalaPRISMConnected: true,
		CardanoConnected:    true,
		CacheStatus:         "healthy",
		DatabaseConnected:   true,
	}

	if !healthy || err != nil {
		response.Status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response.Status = "healthy"
	c.JSON(http.StatusOK, response)
}

// RegisterRoutes registers all DID routes
func (h *Handlers) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/v1")
	{
		// DID operations
		v1.POST("/dids", h.CreateDID)
		v1.GET("/dids/:did", h.ResolveDID)
		v1.PUT("/dids/:did", h.UpdateDID)
		v1.GET("/dids/:did/mapping", h.GetDIDMapping)

		// User to DID mapping
		v1.GET("/users/:userId/did", h.GetDIDByUserID)

		// Device management
		v1.GET("/dids/:did/devices", h.ListDevices)
		v1.POST("/dids/:did/devices", h.RegisterDevice)
		v1.DELETE("/dids/:did/devices/:deviceId", h.UnregisterDevice)

		// Device registration flow
		v1.POST("/dids/:did/devices/register/initiate", h.InitiateDeviceRegistration)
		v1.POST("/dids/:did/devices/register/qrcode", h.GenerateQRCodeForDeviceRegistration)
		v1.POST("/devices/register/complete", h.CompleteDeviceRegistration)

		// DID verification
		v1.POST("/dids/verify", h.VerifyDIDDocument)

		// Generation progress
		v1.GET("/dids/:did/generation/progress", h.GetGenerationProgress)

		// Cache management
		v1.POST("/cache/invalidate/:did", h.InvalidateCache)
		v1.POST("/cache/clear", h.ClearCache)
		v1.GET("/cache/stats", h.GetCacheStats)

		// Health check
		v1.GET("/health", h.Health)
	}
}

// handleError handles errors and returns appropriate HTTP responses
func handleError(c *gin.Context, err error) {
	var didErr *DIDError
	if _, ok := err.(*ValidationErrors); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if ok := asError(err, &didErr); ok {
		statusCode := getStatusCodeForError(didErr.Code)
		c.JSON(statusCode, gin.H{
			"error":   didErr.Code,
			"message": didErr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "Internal server error",
	})
}

// asError checks if err is a DIDError and assigns it to target
func asError(err error, target **DIDError) bool {
	if e, ok := err.(*DIDError); ok {
		*target = e
		return true
	}
	return false
}

// getStatusCodeForError returns the HTTP status code for a DID error code
func getStatusCodeForError(code string) int {
	switch code {
	case ErrCodeInvalidDID, ErrCodeInvalidRequest, ErrCodeInvalidPublicKey, ErrCodeInvalidDocument:
		return http.StatusBadRequest
	case ErrCodeDIDNotFound, ErrCodeDeviceNotFound:
		return http.StatusNotFound
	case ErrCodeDIDAlreadyExists, ErrCodeDeviceAlreadyExists:
		return http.StatusConflict
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeTimeout:
		return http.StatusGatewayTimeout
	case ErrCodeDatabaseError, ErrCodeAtalaPRISMError, ErrCodeBlockchainError, ErrCodeCacheError:
		return http.StatusServiceUnavailable
	case ErrCodeGenerationFailed, ErrCodeAnchoringFailed, ErrCodeResolutionFailed:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
