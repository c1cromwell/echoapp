package zk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// VerifyRequest is a ZK proof verification input (WO-236).
type VerifyRequest struct {
	SubjectDID string `json:"subject_did"`
	ClaimType  string `json:"claim_type"`
	Proof      string `json:"proof"`
	Nonce      string `json:"nonce"`
}

// VerifyResult returns verification outcome with graceful degradation.
type VerifyResult struct {
	Verified bool   `json:"verified"`
	Mode     string `json:"mode"` // "midnight" | "commitment_fallback"
	Detail   string `json:"detail,omitempty"`
}

// Verifier validates commitment-style proofs; Midnight path deferred (WO-235).
type Verifier struct{}

// NewVerifier creates a WO-236 verifier.
func NewVerifier() *Verifier { return &Verifier{} }

// Verify checks required fields and validates H(claim||nonce||subject) commitment proofs.
func (v *Verifier) Verify(req VerifyRequest) (VerifyResult, error) {
	if req.Proof == "" || req.Nonce == "" || req.SubjectDID == "" || req.ClaimType == "" {
		return VerifyResult{}, fmt.Errorf("subject_did, claim_type, proof and nonce required")
	}
	if strings.HasPrefix(req.Proof, "midnight:") {
		return verifyMidnightEnvelope(req)
	}
	expected := commitmentHash(req.SubjectDID, req.ClaimType, req.Nonce)
	proof := strings.TrimSpace(req.Proof)
	if !strings.EqualFold(proof, expected) {
		return VerifyResult{
			Verified: false,
			Mode:     "commitment_fallback",
			Detail:   "proof does not match tier commitment",
		}, nil
	}
	return VerifyResult{
		Verified: true,
		Mode:     "commitment_fallback",
		Detail:   "Commitment proof verified (Midnight deferred)",
	}, nil
}

func commitmentHash(subjectDID, claimType, nonce string) string {
	h := sha256.New()
	h.Write([]byte(subjectDID))
	h.Write([]byte{0})
	h.Write([]byte(claimType))
	h.Write([]byte{0})
	h.Write([]byte(nonce))
	return hex.EncodeToString(h.Sum(nil))
}
