package oidc4vc

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/thechadcromwell/echoapp/pkg/credentials"
)

// Issuer represents OIDC4VC Credential Issuer
type Issuer struct {
	metadata          *IssuerMetadata
	metadataManager   *MetadataManager
	flowManager       *FlowManager
	credentialService *credentials.Service
}

// NewIssuer creates new OIDC4VC issuer
func NewIssuer(issuerDID, verifierDID, issuerBaseURL, verifierBaseURL string) *Issuer {
	metadataManager := NewMetadataManager(issuerDID, verifierDID, issuerBaseURL, verifierBaseURL)
	metadata := metadataManager.GenerateIssuerMetadata()

	flowConfig := &Config{
		IssuerDID:                issuerDID,
		IssuerBaseURL:            issuerBaseURL,
		AuthorizationCodeTTL:     10 * time.Minute,
		PreAuthorizedCodeTTL:     15 * time.Minute,
		AccessTokenTTL:           1 * time.Hour,
		EnablePKCE:               true,
		RequireProofOfPossession: false,
	}

	flowManager := NewFlowManager(flowConfig)

	return &Issuer{
		metadata:        metadata,
		metadataManager: metadataManager,
		flowManager:     flowManager,
	}
}

// SetCredentialService wires W3C VC issuance (OpenID4VCI credential endpoint).
func (i *Issuer) SetCredentialService(svc *credentials.Service) {
	i.credentialService = svc
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
	base := strings.TrimSuffix(i.metadata.CredentialIssuer, "/")
	metadata := map[string]interface{}{
		"issuer":                   base,
		"authorization_endpoint":   base + "/oauth/authorization",
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
	at, err := i.flowManager.ValidateAccessToken(token)
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

	nonce, err := i.flowManager.GenerateCNonce(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return
	}

	if i.credentialService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":             "server_error",
			"error_description": "credential issuance not configured",
		})
		return
	}

	subjectDID := strings.TrimSpace(at.SubjectDID)
	if req.CredentialSubject != nil {
		if id, ok := req.CredentialSubject["id"].(string); ok && strings.TrimSpace(id) != "" {
			if subjectDID != "" && strings.TrimSpace(id) != subjectDID {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":             "invalid_credential_request",
					"error_description": "credential_subject.id does not match issuance session",
				})
				return
			}
			if subjectDID == "" {
				subjectDID = strings.TrimSpace(id)
			}
		}
	}
	if subjectDID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_credential_request",
			"error_description": "holder DID missing (bind pre-authorized code with subject or send credentialSubject.id)",
		})
		return
	}

	credType := credentials.CredentialType(req.CredentialType[0])
	claims := map[string]interface{}{}
	for k, v := range req.Claims {
		claims[k] = v
	}
	for k, v := range req.CredentialSubject {
		if k == "id" {
			continue
		}
		claims[k] = v
	}

	issueReq := &credentials.CredentialIssuanceRequest{
		SubjectDID:      subjectDID,
		CredentialType:  credType,
		Claims:          claims,
		PreferredFormat: oidcFormatToCredentialFormat(req.Format),
	}

	issued, err := i.credentialService.IssueCredential(c.Request.Context(), issueReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "issuance_failed",
			"error_description": err.Error(),
		})
		return
	}

	resp := &CredentialResponse{
		Format:     req.Format,
		Credential: issued.VerifiableCredential,
		CNonc:      nonce,
	}

	c.JSON(http.StatusOK, resp)
}

func oidcFormatToCredentialFormat(f string) credentials.CredentialFormat {
	switch strings.ToLower(strings.TrimSpace(f)) {
	case "jwt_vc_json", "jwt":
		return credentials.JWTFormat
	case "sd-jwt", "sdjwt":
		return credentials.SDJWTFormat
	default:
		return credentials.JSONLDFormat
	}
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
