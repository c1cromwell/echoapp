package wallet

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

// RealFundsEnabled reports whether the wallet runs in real-funds custody mode.
//
// Default false: the interim deterministic-address flow (DAG = SHA256(did)) is
// server-derivable and therefore NOT user-held custody. It is acceptable only
// for TestFlight / no-real-funds use until the Constellation signing SDK ships
// and clients can prove ownership of a key the backend cannot reproduce.
func RealFundsEnabled() bool {
	return os.Getenv("ECHO_WALLET_REAL_FUNDS") == "1"
}

// CustodyMode labels the current custody posture for clients so the UI can show
// whether real funds are in play.
func CustodyMode() string {
	if RealFundsEnabled() {
		return "real_funds"
	}
	return "interim"
}

// ServerDerivableAddress reproduces the iOS interim derivation
// (WalletProvisioner.deterministicDAGAddress): "DAG" + first 36 hex chars of
// SHA256(did). An address equal to this is reproducible by the backend, so it
// cannot represent user-held custody and must be rejected in real-funds mode.
func ServerDerivableAddress(did string) string {
	sum := sha256.Sum256([]byte(did))
	return "DAG" + hex.EncodeToString(sum[:])[:36]
}

// ProofVerifier verifies a client-held proof-of-ownership of a DAG address and
// returns the validated secp256k1 public key (which the link handler binds to
// the DID). The client signs a server challenge with the wallet key; see
// DagProofVerifier. Until a verifier is wired, real-funds mode hard-blocks.
type ProofVerifier interface {
	VerifyOwnership(did, address, proof string) (publicKey string, err error)
}
