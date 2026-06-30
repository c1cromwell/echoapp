package wallet

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
)

// WalletProof is the client's proof-of-ownership, carried base64-encoded in the
// X-Wallet-Proof header. The client signs the server-issued challenge with the
// wallet's secp256k1 key; the backend verifies against the bound public key.
type WalletProof struct {
	PublicKey string `json:"publicKey"`
	Challenge string `json:"challenge"`
	Signature string `json:"signature"`
}

var (
	errProofMalformed   = errors.New("malformed wallet proof")
	errProofSignature   = errors.New("invalid wallet proof signature")
	errProofChallenge   = errors.New("unknown or expired challenge")
	errProofAddrBinding = errors.New("address does not match public key")
	errProofNotLinked   = errors.New("wallet not linked")
	errProofKeyMismatch = errors.New("public key does not match linked wallet")
)

// DagProofVerifier implements ProofVerifier using the dag4 address derivation
// and the secp256k1 signature stack, plus single-use server challenges. It is
// the real-funds custody gate the iOS client satisfies by signing locally.
type DagProofVerifier struct {
	store      Store
	challenges *ChallengeStore
}

// NewDagProofVerifier wires the verifier to the wallet store (for the bound
// public key) and the challenge store (for replay protection).
func NewDagProofVerifier(store Store, challenges *ChallengeStore) *DagProofVerifier {
	return &DagProofVerifier{store: store, challenges: challenges}
}

// VerifyOwnership checks the proof and returns the validated public key.
//
//   - address != "" (link): establishes the binding — the address MUST derive
//     from the proof's public key.
//   - address == "" (value-moving op): the proof's public key MUST match the
//     key already bound to the DID.
//
// In both cases the signature must verify over a fresh, single-use challenge.
func (v *DagProofVerifier) VerifyOwnership(did, address, proof string) (string, error) {
	p, err := parseWalletProof(proof)
	if err != nil {
		return "", err
	}
	if !VerifyDagMessageSignature(p.PublicKey, p.Challenge, p.Signature) {
		return "", errProofSignature
	}
	if !v.challenges.Consume(did, p.Challenge) {
		return "", errProofChallenge
	}
	if address != "" {
		if DagAddressFromPubKey(p.PublicKey) != address {
			return "", errProofAddrBinding
		}
		return p.PublicKey, nil
	}
	bound, err := v.store.GetDAGAccountPubKey(context.Background(), did)
	if err != nil || bound == "" {
		return "", errProofNotLinked
	}
	if bound != p.PublicKey {
		return "", errProofKeyMismatch
	}
	return p.PublicKey, nil
}

func parseWalletProof(proof string) (WalletProof, error) {
	var p WalletProof
	raw, err := base64.StdEncoding.DecodeString(proof)
	if err != nil {
		// Tolerate raw JSON in addition to base64 for ease of testing.
		raw = []byte(proof)
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, errProofMalformed
	}
	if p.PublicKey == "" || p.Challenge == "" || p.Signature == "" {
		return p, errProofMalformed
	}
	return p, nil
}

var _ ProofVerifier = (*DagProofVerifier)(nil)
