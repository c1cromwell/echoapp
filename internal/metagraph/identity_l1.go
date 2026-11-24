// Identity records the gateway anchors on Echo's chain. These structs are the gateway's own
// description of a record; what reaches the chain is hashes only (internal/besu/ids.go):
//
//	DeviceKeyRegistrationUpdate — the DID and the key are stored as hashes; the device label,
//	  which is free text, is not sent to the chain at all.
//	TrustTierCommitmentUpdate — Commitment = H(tier || nonce), already a hash; the DID is hashed.
//	UsernameRegistrationUpdate — the username and the DID are stored as hashes. A username is
//	  guessable, so its hash does not hide it; ownership is meant to be checkable.
//
// See docs/data-classification.md for the full T0–T7 policy.
package metagraph

import (
	"crypto/sha256"
	"encoding/hex"
)

// DeviceKeyRegistrationUpdate is the JSON request body the Identity Service
// sends to Identity Metagraph L1 (Tessellation CurrencyL1-style POST …/transactions).
//
// Field names match Scala circe generic encoding for
// com.echo.shared_data.types.DeviceKeyRegistrationUpdate (camelCase).
// Euclid public port: 9500 (see metagraph/euclid.json).
//
// Discrimination: this shape is the only IdentityUpdate variant that includes
// publicKeyHex + deviceLabel + addedAt without bitVector, commitment, or schemaVersion.
type DeviceKeyRegistrationUpdate struct {
	SubjectDID   string `json:"subjectDID"`
	PublicKeyHex string `json:"publicKeyHex"`
	DeviceLabel  string `json:"deviceLabel"`
	AddedAt      int64  `json:"addedAt"`
}

// TrustTierCommitmentUpdate is the Identity L1 wire shape for H(tier || nonce) commitments
// (com.echo.shared_data.types.TrustTierCommitmentUpdate).
type TrustTierCommitmentUpdate struct {
	SubjectDID string `json:"subjectDID"`
	Commitment string `json:"commitment"`
	AnchoredAt int64  `json:"anchoredAt"`
}

// UsernameRegistrationUpdate is the Identity L1 wire shape for a public
// @username -> DID binding (com.echo.shared_data.types.UsernameRegistrationUpdate).
//
// T7 (public chain data): username and subjectDID are public; the metagraph is
// the authoritative, portable index for username ownership. The backend Postgres
// users table is a read-through cache. Discriminated from other IdentityUpdate
// variants by the presence of `username`.
type UsernameRegistrationUpdate struct {
	SubjectDID   string `json:"subjectDID"`
	Username     string `json:"username"`
	RegisteredAt int64  `json:"registeredAt"`
}

// TrustTierCommitmentHex returns SHA256(byte(tier) || nonce) as 64-char lowercase hex,
// matching internal/api trust tier handlers and Scala IdentityValidations expectations.
func TrustTierCommitmentHex(tier int, nonce string) string {
	preimage := make([]byte, 0, 1+len(nonce))
	preimage = append(preimage, byte(tier))
	preimage = append(preimage, []byte(nonce)...)
	sum := sha256.Sum256(preimage)
	return hex.EncodeToString(sum[:])
}
