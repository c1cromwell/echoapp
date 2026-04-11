package oidc4vc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Verifier represents OIDC4VC Presentation Verifier
type Verifier struct {
	metadata        *VerifierMetadata
	metadataManager *MetadataManager
	flowManager     *FlowManager
	verifyService   interface{} // Will be credential verifier service
}

// NewVerifier creates new OIDC4VC verifier
func NewVerifier(verifierDID, issuerDID, verifierBaseURL, issuerBaseURL string) *Verifier {
	metadataManager := NewMetadataManager(issuerDID, verifierDID, issuerBaseURL, verifierBaseURL)
	metadata := metadataManager.GenerateVerifierMetadata()

	flowConfig := &Config{
		IssuerDID:            verifierDID,
		IssuerBaseURL:        verifierBaseURL,
		AuthorizationCodeTTL: 10 * 60 * 1000,
		PreAuthorizedCodeTTL: 15 * 60 * 1000,
		AccessTokenTTL:       3600 * 1000,
		EnablePKCE:           true,
	}

	flowManager := NewFlowManager(flowConfig)

	return &Verifier{
		metadata:        metadata,
		metadataManager: metadataManager,
		flowManager:     flowManager,
	}
}

// RegisterRoutes registers OIDC4VC verifier routes
func (v *Verifier) RegisterRoutes(router *gin.Engine) {
	// Metadata endpoints
	router.GET("/.well-known/openid-credential-verifier", v.GetMetadata)

	// Verification endpoints
	verifyGroup := router.Group("/verification")
	verifyGroup.GET("/request", v.CreatePresentationRequest)
	verifyGroup.POST("/submit", v.SubmitPresentation)
	verifyGroup.GET("/:presentationId/status", v.GetVerificationStatus)

	// Presentation definition endpoint
	router.GET("/presentation_definition/:definitionId", v.GetPresentationDefinition)
}

// GetMetadata returns OIDC4VC verifier metadata
// @GET /.well-known/openid-credential-verifier
// @Produce json
func (v *Verifier) GetMetadata(c *gin.Context) {
	c.JSON(http.StatusOK, v.metadata)
}

// CreatePresentationRequest creates a presentation request
// @GET /verification/request?credential_type=ProofOfHumanity
// @Produce json
func (v *Verifier) CreatePresentationRequest(c *gin.Context) {
	credentialType := c.Query("credential_type")
	if credentialType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "credential_type is required",
		})
		return
	}

	clientID := c.Query("client_id")
	if clientID == "" {
		clientID = v.metadata.VerifierID
	}

	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = fmt.Sprintf("%s/verification/submit", v.metadata.VerificationEndpoint[:len(v.metadata.VerificationEndpoint)-11])
	}

	// Generate state
	state, _ := generateRandomCode(16)

	// Create presentation request
	presentationReq := v.metadataManager.GeneratePresentationRequest(
		clientID,
		redirectURI,
		state,
		credentialType,
	)

	c.JSON(http.StatusOK, presentationReq)
}

// SubmitPresentation submits presentation for verification
// @POST /verification/submit
// @Accept json
// @Produce json
func (v *Verifier) SubmitPresentation(c *gin.Context) {
	var req struct {
		PresentationSubmission *PresentationSubmission `json:"presentation_submission"`
		VPToken                string                  `json:"vp_token"`
		IDToken                string                  `json:"id_token,omitempty"`
		State                  string                  `json:"state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_request",
		})
		return
	}

	// Validate presentation submission
	if req.PresentationSubmission == nil || req.VPToken == "" || req.State == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "presentation_submission, vp_token, and state are required",
		})
		return
	}

	// In production, verify VP token and validate presentation
	// For now, return success
	presentationID := "pres_" + req.State

	resp := map[string]interface{}{
		"presentationId": presentationID,
		"status":         "verified",
		"verificationResult": map[string]interface{}{
			"isValid": true,
			"credentials": []map[string]interface{}{
				{
					"credentialId": "cred_123",
					"verified":     true,
				},
			},
		},
	}

	c.JSON(http.StatusOK, resp)
}

// GetVerificationStatus gets verification status
// @GET /verification/:presentationId/status
// @Produce json
func (v *Verifier) GetVerificationStatus(c *gin.Context) {
	presentationID := c.Param("presentationId")

	// In production, lookup actual verification result
	status := map[string]interface{}{
		"presentationId": presentationID,
		"status":         "completed",
		"timestamp":      "2024-01-15T10:30:00Z",
		"verified":       true,
		"claims": map[string]interface{}{
			"credentialType": "ProofOfHumanity",
			"issuer":         "did:prism:cardano:issuer",
			"subject":        "did:prism:cardano:subject",
		},
	}

	c.JSON(http.StatusOK, status)
}

// GetPresentationDefinition gets presentation definition
// @GET /presentation_definition/:definitionId
// @Produce json
func (v *Verifier) GetPresentationDefinition(c *gin.Context) {
	definitionID := c.Param("definitionId")

	// Create presentation definition based on ID
	// Parse credential type from definitionID using convention: "echo_<type>_v<version>"
	var credentialType string
	if definitionID != "" {
		credentialType = parseCredentialTypeFromDefinitionID(definitionID)
	} else {
		credentialType = "ProofOfHumanity"
	}

	def := v.metadataManager.buildPresentationDefinition(credentialType)

	c.JSON(http.StatusOK, def)
}

// ProcessPresentationRequest processes a presentation request
func (v *Verifier) ProcessPresentationRequest(presentationReq *PresentationRequest) error {
	// Validate presentation request
	if err := v.metadataManager.ValidatePresentationRequest(presentationReq); err != nil {
		return err
	}

	// Store presentation request for later matching with submission
	// In production, store in database with TTL

	return nil
}

// VerifyPresentation verifies a presentation
func (v *Verifier) VerifyPresentation(vpToken string, presentationSubmission *PresentationSubmission) (bool, error) {
	// In production:
	// 1. Parse VP token (JWT or JSON-LD)
	// 2. Verify signature
	// 3. Validate credential types match presentation definition
	// 4. Verify each credential in the presentation
	// 5. Return verification result

	if vpToken == "" {
		return false, fmt.Errorf("vp_token is required")
	}

	if presentationSubmission == nil {
		return false, fmt.Errorf("presentation_submission is required")
	}

	// For now, return success
	return true, nil
}

// GetPresentationStatus gets presentation verification status
func (v *Verifier) GetPresentationStatus(presentationID string) map[string]interface{} {
	return map[string]interface{}{
		"presentationId": presentationID,
		"status":         "verified",
		"verified":       true,
	}
}

// MatchPresentationWithRequest matches presentation against request definition
func (v *Verifier) MatchPresentationWithRequest(presentation *PresentationSubmission, definition *PresentationDef) bool {
	// Validate that submitted credentials match the requested credential types
	if presentation == nil || definition == nil {
		return false
	}

	if presentation.DefinitionID != definition.ID {
		return false
	}

	// Check descriptor mappings
	if len(presentation.DescriptorMap) != len(definition.InputDescriptors) {
		return false
	}

	return true
}

// ExtractClaimsFromPresentation extracts verified claims from presentation
func (v *Verifier) ExtractClaimsFromPresentation(vpToken string) map[string]interface{} {
	// In production, parse VP token and extract claims
	return map[string]interface{}{
		"credentialType": "ProofOfHumanity",
		"issuer":         "did:prism:cardano:issuer",
		"subject":        "did:prism:cardano:subject",
	}
}

// RespondToPresentation responds to presentation request with verification result
func (v *Verifier) RespondToPresentation(presentationID string, isValid bool, details map[string]interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"presentationId": presentationID,
		"verified":       isValid,
		"timestamp":      "2024-01-15T10:30:00Z",
	}

	if details != nil {
		response["details"] = details
	}

	return response
}

// parseCredentialTypeFromDefinitionID extracts the credential type from a
// definition ID using the convention: "echo_<type>_v<version>".
// Examples:
//
//	"echo_proof_of_humanity_v1" → "ProofOfHumanity"
//	"echo_kyc_lite_v1"          → "KYCLite"
//	"echo_org_verified_v1"      → "OrgVerified"
//	unknown/empty               → "ProofOfHumanity" (default)
func parseCredentialTypeFromDefinitionID(definitionID string) string {
	knownTypes := map[string]string{
		"proof_of_humanity": "ProofOfHumanity",
		"kyc_lite":          "KYCLite",
		"kyc_full":          "KYCFull",
		"org_verified":      "OrgVerified",
		"apple_digital_id":  "AppleDigitalID",
		"phone_verified":    "PhoneVerified",
		"email_verified":    "EmailVerified",
	}

	// Strip "echo_" prefix and "_v<N>" suffix
	id := strings.TrimPrefix(definitionID, "echo_")
	if idx := strings.LastIndex(id, "_v"); idx > 0 {
		id = id[:idx]
	}

	if ct, ok := knownTypes[id]; ok {
		return ct
	}
	return "ProofOfHumanity"
}
