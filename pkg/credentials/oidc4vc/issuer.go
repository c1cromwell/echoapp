package oidc4vc

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Issuer represents OIDC4VC Credential Issuer
type Issuer struct {
	metadata          *IssuerMetadata
	metadataManager   *MetadataManager
	flowManager       *FlowManager
	credentialService interface{} // Will be credential service
}

// NewIssuer creates new OIDC4VC issuer
func NewIssuer(issuerDID, verifierDID, issuerBaseURL, verifierBaseURL string) *Issuer {
	metadataManager := NewMetadataManager(issuerDID, verifierDID, issuerBaseURL, verifierBaseURL)
	metadata := metadataManager.GenerateIssuerMetadata()

	flowConfig := &Config{
		IssuerDID:            issuerDID,
		IssuerBaseURL:        issuerBaseURL,
		AuthorizationCodeTTL: 10 * 60 * 1000, // 10 minutes in milliseconds
		PreAuthorizedCodeTTL: 15 * 60 * 1000, // 15 minutes
		AccessTokenTTL:       3600 * 1000,    // 1 hour
		EnablePKCE:           true,
	}

	flowManager := NewFlowManager(flowConfig)

	return &Issuer{
		metadata:        metadata,
		metadataManager: metadataManager,
		flowManager:     flowManager,
	}
}

// RegisterRoutes registers OIDC4VC issuer routes
func (i *Issuer) RegisterRoutes(router *gin.Engine) {
	// Metadata endpoints
	router.GET("/.well-known/openid-credential-issuer", i.GetMetadata)
	router.GET("/.well-known/oauth-authorization-server", i.GetOAuthMetadata)

	// Authorization endpoints
	authGroup := router.Group("/oauth")
	authGroup.GET("/authorization", i.AuthorizationEndpoint)
	authGroup.POST("/token", i.TokenEndpoint)

	// Credential endpoints
	credGroup := router.Group("/credential")
	credGroup.POST("", i.CredentialEndpoint)
	credGroup.POST("/deferred", i.DeferredCredentialEndpoint)

	// Notification endpoint
	router.POST("/notification", i.NotificationEndpoint)
}

// GetMetadata returns OIDC4VC issuer metadata
// @GET /.well-known/openid-credential-issuer
// @Produce json
func (i *Issuer) GetMetadata(c *gin.Context) {
	c.JSON(http.StatusOK, i.metadata)
}

// GetOAuthMetadata returns OAuth authorization server metadata
// @GET /.well-known/oauth-authorization-server
// @Produce json
func (i *Issuer) GetOAuthMetadata(c *gin.Context) {
	metadata := map[string]interface{}{
		"issuer":                   i.metadata.CredentialIssuer,
		"authorization_endpoint":   fmt.Sprintf("%s/oauth/authorization", i.metadata.TokenEndpoint[:len(i.metadata.TokenEndpoint)-6]),
		"token_endpoint":           i.metadata.TokenEndpoint,
		"credential_endpoint":      i.metadata.CredentialEndpoint,
		"response_types_supported": []string{"code"},
		"grant_types_supported": []string{
			"authorization_code",
			"urn:ietf:params:oauth:grant-type:pre-authorized_code",
		},
		"code_challenge_methods_supported": []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{
			"none",
			"client_secret_basic",
		},
	}

	c.JSON(http.StatusOK, metadata)
}

// AuthorizationEndpoint handles authorization requests
// @GET /oauth/authorization
// @Produce json
func (i *Issuer) AuthorizationEndpoint(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")

	// Validate request
	if clientID == "" || redirectURI == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing required parameters",
		})
		return
	}

	if responseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unsupported_response_type",
		})
		return
	}

	// Create authorization code
	code, err := i.flowManager.CreateAuthorizationCode(clientID, redirectURI, scope, state, codeChallenge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server_error",
		})
		return
	}

	// Redirect to client with code
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
	c.Redirect(http.StatusFound, redirectURL)
}

// TokenEndpoint handles token requests
// @POST /oauth/token
// @Accept json
// @Produce json
func (i *Issuer) TokenEndpoint(c *gin.Context) {
	var req TokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_request",
		})
		return
	}

	var tokenResp *TokenResponse
	var err error

	switch req.GrantType {
	case "authorization_code":
		tokenResp, err = i.flowManager.ExchangeAuthorizationCode(
			req.Code,
			req.ClientID,
			req.CodeVerifier,
			req.RedirectURI,
		)

	case "urn:ietf:params:oauth:grant-type:pre-authorized_code":
		tokenResp, err = i.flowManager.ExchangePreAuthorizedCode(
			req.PreAuthorizedCode,
			req.TxCode,
		)

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unsupported_grant_type",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// CredentialEndpoint handles credential requests
// @POST /credential
// @Accept json
// @Produce json
func (i *Issuer) CredentialEndpoint(c *gin.Context) {
	// Get authorization token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_request",
			"error_description": "Authorization header required",
		})
		return
	}

	// Extract token (Bearer token)
	var token string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid_request",
		})
		return
	}

	// Validate token
	_, err := i.flowManager.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid_token",
		})
		return
	}

	var req CredentialRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_credential_request",
		})
		return
	}

	// Validate credential request
	if err := i.metadataManager.ValidateCredentialRequest(&req, i.metadata.CredentialConfigurationsSupported); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_credential_request",
			"error_description": err.Error(),
		})
		return
	}

	// In production, call credential service to issue credential
	// For now, return mock response
	credential := "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJ2YyI6eyJAY29udGV4dCI6WyJodHRwczovL3d3dy53My5vcmcvMjAxOC9jcmVkZW50aWFscy92MSJdLCJ0eXBlIjpbIlZlcmlmaWFibGVDcmVkZW50aWFsIl0sImNyZWRlbnRpYWxTdWJqZWN0Ijp7fX0sImlzcyI6ImRpZDpwcmlzbTpjb3JkYXJvOmNvbnRyb2xsZXIifQ.signature"

	// Generate new c_nonce
	nonce, _ := i.flowManager.GenerateCNonce(token)

	resp := &CredentialResponse{
		Format:     req.Format,
		Credential: credential,
		CNonc:      nonce,
	}

	c.JSON(http.StatusOK, resp)
}

// DeferredCredentialEndpoint handles deferred credential requests
// @POST /credential/deferred
// @Accept json
// @Produce json
func (i *Issuer) DeferredCredentialEndpoint(c *gin.Context) {
	var req DeferredCredentialRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_request",
		})
		return
	}

	// Validate acceptance token
	if req.AcceptanceToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_request",
		})
		return
	}

	// In production, check if credential issuance is complete
	// For now, return pending
	resp := &DeferredCredentialResponse{
		TransactionID: "tx_" + req.AcceptanceToken,
		IssuanceDate:  "2024-01-15T10:30:00Z",
	}

	c.JSON(http.StatusOK, resp)
}

// NotificationEndpoint handles notifications
// @POST /notification
// @Accept json
func (i *Issuer) NotificationEndpoint(c *gin.Context) {
	var req NotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_request",
		})
		return
	}

	// Process notification
	// In production, handle credential_issued, credential_deleted events

	c.JSON(http.StatusOK, gin.H{
		"status": "received",
	})
}
