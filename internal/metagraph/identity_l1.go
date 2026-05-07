package metagraph

import (
	"context"
	"fmt"
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

// SubmitIdentityL1 posts an Identity Metagraph L1 transaction (e.g. device key
// registration). The payload must encode as one of the Identity L1 update types;
// most callers use [DeviceKeyRegistrationUpdate].
func (c *MetagraphClient) SubmitIdentityL1(ctx context.Context, tx interface{}) (string, error) {
	base := c.config.IdentityL1URL
	if base == "" {
		return "", fmt.Errorf("metagraph: IdentityL1URL is not configured")
	}
	return c.submitTransaction(ctx, base+"/transactions", tx)
}
