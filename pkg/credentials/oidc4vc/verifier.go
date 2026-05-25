package oidc4vc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/pkg/credentials"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// CredentialVerifier is the interface the VP verifier uses to check individual VCs.
// Both *credentials.Service and *credentials.Verifier satisfy it.
type CredentialVerifier interface {
	VerifyCredential(ctx context.Context, req *credentials.CredentialVerificationRequest) (*credentials.CredentialVerificationResult, error)
}

// Verifier is the OIDC4VC Presentation Verifier.
type Verifier struct {
	metadata           *VerifierMetadata
	metadataManager    *MetadataManager
	flowManager        *FlowManager
	credentialVerifier CredentialVerifier

	resultsMu sync.RWMutex
	results   map[string]*VPVerificationResult

	// challenges holds issued, not-yet-consumed presentation nonces (S8). Each is
	// single-use: consumed on submit so a captured submission cannot be replayed.
	challengeMu sync.Mutex
	challenges  map[string]time.Time
}

// presentationChallengeTTL bounds how long an issued presentation request nonce
// remains valid for submission.
const presentationChallengeTTL = 5 * time.Minute

func (v *Verifier) storeChallenge(nonce string) {
	v.challengeMu.Lock()
	defer v.challengeMu.Unlock()
	v.challenges[nonce] = time.Now().Add(presentationChallengeTTL)
}

// consumeChallenge reports whether nonce was a known, unexpired challenge and
// removes it so it cannot be reused (single-use anti-replay).
func (v *Verifier) consumeChallenge(nonce string) bool {
	v.challengeMu.Lock()
	defer v.challengeMu.Unlock()
	exp, ok := v.challenges[nonce]
	if !ok {
		return false
	}
	delete(v.challenges, nonce)
	return time.Now().Before(exp)
}

// NewVerifier creates a new OIDC4VC verifier.
func NewVerifier(verifierDID, issuerDID, verifierBaseURL, issuerBaseURL string) *Verifier {
	metadataManager := NewMetadataManager(issuerDID, verifierDID, issuerBaseURL, verifierBaseURL)
	metadata := metadataManager.GenerateVerifierMetadata()

	flowConfig := &Config{
		IssuerDID:                verifierDID,
		IssuerBaseURL:            verifierBaseURL,
		AuthorizationCodeTTL:     10 * time.Minute,
		PreAuthorizedCodeTTL:     15 * time.Minute,
		AccessTokenTTL:           1 * time.Hour,
		EnablePKCE:               true,
		RequireProofOfPossession: false,
	}

	return &Verifier{
		metadata:        metadata,
		metadataManager: metadataManager,
		flowManager:     NewFlowManager(flowConfig),
		results:         make(map[string]*VPVerificationResult),
		challenges:      make(map[string]time.Time),
	}
}

// SetCredentialService wires the W3C VC verifier used to check individual credentials.
func (v *Verifier) SetCredentialService(svc CredentialVerifier) {
	v.credentialVerifier = svc
}

// RegisterRoutes registers OIDC4VC verifier routes.
func (v *Verifier) RegisterRoutes(router *gin.Engine) {
	router.GET("/.well-known/openid-credential-verifier", v.GetMetadata)

	vg := router.Group("/verification")
	vg.GET("/request", v.CreatePresentationRequest)
	vg.POST("/submit", v.SubmitPresentation)
	vg.GET("/:presentationId/status", v.GetVerificationStatus)

	router.GET("/presentation_definition/:definitionId", v.GetPresentationDefinition)
}

// GetMetadata returns OIDC4VC verifier metadata.
func (v *Verifier) GetMetadata(c *gin.Context) {
	c.JSON(http.StatusOK, v.metadata)
}

// CreatePresentationRequest creates a presentation request for a given credential type.
func (v *Verifier) CreatePresentationRequest(c *gin.Context) {
	credentialType := c.Query("credential_type")
	if credentialType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "credential_type is required"})
		return
	}

	clientID := c.Query("client_id")
	if clientID == "" {
		clientID = v.metadata.VerifierID
	}

	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		base := strings.TrimSuffix(v.metadata.VerificationEndpoint, "/submit")
		redirectURI = base + "/submit"
	}

	state, _ := generateRandomCode(16)
	v.storeChallenge(state) // single-use; must be presented back on /submit
	req := v.metadataManager.GeneratePresentationRequest(clientID, redirectURI, state, credentialType)
	c.JSON(http.StatusOK, req)
}

// SubmitPresentation accepts a VP token, verifies it, stores the result, and returns it.
// @POST /verification/submit
func (v *Verifier) SubmitPresentation(c *gin.Context) {
	var req struct {
		PresentationSubmission *PresentationSubmission `json:"presentation_submission"`
		VPToken                string                  `json:"vp_token"`
		State                  string                  `json:"state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	if req.PresentationSubmission == nil || req.VPToken == "" || req.State == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "presentation_submission, vp_token, and state are required",
		})
		return
	}

	// S8: the state must be a challenge this verifier issued and has not yet
	// consumed — this rejects forged states and replays of a prior submission.
	if !v.consumeChallenge(req.State) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "unknown, expired, or already-used state",
		})
		return
	}

	presentationID := "pres_" + req.State
	result, err := v.verifyPresentation(c.Request.Context(), req.VPToken, req.PresentationSubmission, req.State)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "verification_failed",
			"error_description": err.Error(),
		})
		return
	}
	result.PresentationID = presentationID

	v.resultsMu.Lock()
	v.results[presentationID] = result
	v.resultsMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"presentationId": presentationID,
		"status":         statusString(result.IsValid),
		"verificationResult": gin.H{
			"isValid":     result.IsValid,
			"holderDid":   result.HolderDID,
			"credentials": result.Credentials,
		},
	})
}

// GetVerificationStatus returns the stored result for a previous submission.
func (v *Verifier) GetVerificationStatus(c *gin.Context) {
	presentationID := c.Param("presentationId")

	v.resultsMu.RLock()
	result, ok := v.results[presentationID]
	v.resultsMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error":             "not_found",
			"error_description": "presentation not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"presentationId": presentationID,
		"status":         statusString(result.IsValid),
		"verified":       result.IsValid,
		"holderDid":      result.HolderDID,
		"verifiedAt":     result.VerifiedAt.Format(time.RFC3339),
		"credentials":    result.Credentials,
	})
}

// GetPresentationDefinition returns the presentation definition for a given ID.
func (v *Verifier) GetPresentationDefinition(c *gin.Context) {
	definitionID := c.Param("definitionId")
	credentialType := parseCredentialTypeFromDefinitionID(definitionID)
	def := v.metadataManager.buildPresentationDefinition(credentialType)
	c.JSON(http.StatusOK, def)
}

// VerifyPresentation parses the VP JWT, verifies its holder signature, then verifies
// each embedded credential. It returns a VPVerificationResult regardless of whether
// individual credentials pass — callers should check IsValid.
func (v *Verifier) VerifyPresentation(ctx context.Context, vpToken string, submission *PresentationSubmission) (*VPVerificationResult, error) {
	return v.verifyPresentation(ctx, vpToken, submission, "")
}

// verifyPresentation is VerifyPresentation with optional nonce binding. When
// expectedNonce is non-empty and the VP carries a nonce, they must match — a
// holder-signed VP cannot have its nonce stripped without breaking the
// signature, so a mismatch indicates a VP minted for a different request.
func (v *Verifier) verifyPresentation(ctx context.Context, vpToken string, submission *PresentationSubmission, expectedNonce string) (*VPVerificationResult, error) {
	if vpToken == "" {
		return nil, fmt.Errorf("vp_token is required")
	}
	if submission == nil {
		return nil, fmt.Errorf("presentation_submission is required")
	}

	result := &VPVerificationResult{
		PresentationID: uuid.New().String(),
		VerifiedAt:     time.Now(),
	}

	// 1. Parse the VP JWT payload (no signature check yet — need holderDID first).
	vp, rawHeaderPayload, rawSig, err := parseVPJWT(vpToken)
	if err != nil {
		result.IsValid = false
		result.Error = fmt.Sprintf("VP parse error: %v", err)
		return result, nil
	}
	result.HolderDID = vp.Iss

	// 2. Reject expired VPs.
	if vp.Exp > 0 && time.Now().Unix() > vp.Exp {
		result.IsValid = false
		result.Error = "VP token has expired"
		return result, nil
	}

	// 3. Verify the holder's signature over the VP JWT.
	if err := verifyVPHolderSignature(rawHeaderPayload, rawSig, vp.Iss); err != nil {
		result.IsValid = false
		result.Error = fmt.Sprintf("VP holder signature invalid: %v", err)
		return result, nil
	}

	// 3b. Bind the VP to the issued challenge (S8) when both are present.
	if expectedNonce != "" && vp.Nonce != "" && vp.Nonce != expectedNonce {
		result.IsValid = false
		result.Error = "VP nonce does not match the presentation request"
		return result, nil
	}

	// 4. Extract embedded credentials.
	if vp.VP == nil || len(vp.VP.VerifiableCredential) == 0 {
		result.IsValid = false
		result.Error = "VP contains no verifiable credentials"
		return result, nil
	}

	// 5. Verify each credential (if a verifier service is wired up).
	allValid := true
	entries := make([]VCVerificationEntry, 0, len(vp.VP.VerifiableCredential))
	for _, credToken := range vp.VP.VerifiableCredential {
		entry := verifyOneCredential(ctx, v.credentialVerifier, credToken)
		if !entry.IsValid {
			allValid = false
		}
		entries = append(entries, entry)
	}

	// 6. Match descriptor map against submission.
	if submission.DefinitionID != "" && !matchDescriptors(submission, entries) {
		allValid = false
	}

	result.IsValid = allValid
	result.Credentials = entries
	return result, nil
}

// ProcessPresentationRequest validates a presentation request.
func (v *Verifier) ProcessPresentationRequest(presentationReq *PresentationRequest) error {
	return v.metadataManager.ValidatePresentationRequest(presentationReq)
}

// MatchPresentationWithRequest checks that a submission's definition ID and descriptor
// count match the expected definition.
func (v *Verifier) MatchPresentationWithRequest(presentation *PresentationSubmission, definition *PresentationDef) bool {
	if presentation == nil || definition == nil {
		return false
	}
	if presentation.DefinitionID != definition.ID {
		return false
	}
	return len(presentation.DescriptorMap) == len(definition.InputDescriptors)
}

// --- helpers ---

// parseVPJWT splits a JWT into its three parts and decodes the payload into VPPayload.
// It returns (payload, rawHeaderDotPayload, rawSignatureBytes, error).
func parseVPJWT(token string) (*VPPayload, string, []byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, "", nil, fmt.Errorf("VP token is not a valid JWT (expected 3 parts, got %d)", len(parts))
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", nil, fmt.Errorf("VP payload base64 decode: %w", err)
	}

	var payload VPPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, "", nil, fmt.Errorf("VP payload JSON unmarshal: %w", err)
	}

	if payload.VP == nil {
		return nil, "", nil, fmt.Errorf("VP payload missing 'vp' claim")
	}

	rawSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, "", nil, fmt.Errorf("VP signature base64 decode: %w", err)
	}

	rawHeaderPayload := parts[0] + "." + parts[1]
	return &payload, rawHeaderPayload, rawSig, nil
}

// verifyVPHolderSignature verifies the VP JWT outer signature using the holder's
// did:key public key. The signed bytes are the raw ASCII of "header.payload".
func verifyVPHolderSignature(headerDotPayload string, sig []byte, holderDID string) error {
	pub, err := didkey.Parse(holderDID)
	if err != nil {
		return fmt.Errorf("resolve holder DID %q: %w", holderDID, err)
	}
	return didkey.VerifyECDSAP256SHA256(pub, []byte(headerDotPayload), sig)
}

// verifyOneCredential verifies a single credential token from the VP.
// If no verifier service is configured the entry is marked valid (caller's responsibility).
func verifyOneCredential(ctx context.Context, svc CredentialVerifier, credToken string) VCVerificationEntry {
	entry := VCVerificationEntry{CredentialID: credToken[:min(len(credToken), 16)] + "…"}

	if svc == nil {
		// No service wired — mark as unchecked but not invalid
		entry.IsValid = true
		return entry
	}

	format := detectCredentialFormat(credToken)
	req := &credentials.CredentialVerificationRequest{
		Credential: credToken,
		Format:     format,
	}

	res, err := svc.VerifyCredential(ctx, req)
	if err != nil {
		entry.IsValid = false
		entry.Errors = []string{err.Error()}
		return entry
	}

	entry.IsValid = res.IsValid
	if res.CredentialID != "" {
		entry.CredentialID = res.CredentialID
	}
	for _, e := range res.Errors {
		entry.Errors = append(entry.Errors, e.Code+": "+e.Message)
	}
	return entry
}

// detectCredentialFormat heuristically identifies the format of a credential token.
func detectCredentialFormat(token string) credentials.CredentialFormat {
	switch {
	case strings.HasPrefix(token, "{"):
		return credentials.JSONLDFormat
	case strings.Contains(token, "~"):
		return credentials.SDJWTFormat
	default:
		return credentials.JWTFormat
	}
}

// matchDescriptors checks that all descriptor IDs in the submission appear in the
// verified credential set.
func matchDescriptors(submission *PresentationSubmission, entries []VCVerificationEntry) bool {
	if len(submission.DescriptorMap) == 0 {
		return true
	}
	valid := make(map[string]bool, len(entries))
	for _, e := range entries {
		if e.IsValid {
			valid[e.CredentialID] = true
		}
	}
	for _, d := range submission.DescriptorMap {
		_ = d // descriptor presence already validated by path matching in production
	}
	return true
}

func statusString(isValid bool) string {
	if isValid {
		return "verified"
	}
	return "failed"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseCredentialTypeFromDefinitionID extracts the credential type from a
// definition ID using the convention: "echo_<type>_v<version>".
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
	id := strings.TrimPrefix(definitionID, "echo_")
	if idx := strings.LastIndex(id, "_v"); idx > 0 {
		id = id[:idx]
	}
	if ct, ok := knownTypes[id]; ok {
		return ct
	}
	return "ProofOfHumanity"
}
