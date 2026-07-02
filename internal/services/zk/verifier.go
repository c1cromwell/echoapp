package zk

import "fmt"

// VerifyRequest is a ZK proof verification input (WO-236 stub).
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

// Verifier accepts proofs until Midnight SDK integration ships (WO-235).
type Verifier struct{}

// NewVerifier creates a WO-236 stub verifier.
func NewVerifier() *Verifier { return &Verifier{} }

// Verify validates required fields and returns commitment-fallback success.
func (v *Verifier) Verify(req VerifyRequest) (VerifyResult, error) {
	if req.Proof == "" || req.Nonce == "" {
		return VerifyResult{}, fmt.Errorf("proof and nonce required")
	}
	return VerifyResult{
		Verified: true,
		Mode:     "commitment_fallback",
		Detail:   "Midnight verification deferred; proof payload accepted",
	}, nil
}
