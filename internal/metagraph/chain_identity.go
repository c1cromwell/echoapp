package metagraph

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// IdentitySubmitter writes identity records to Echo's chain and returns a reference to the write.
// Implementations store hashes only (internal/besu/ids.go): the DID, username and key in the
// update never reach the chain as text.
type IdentitySubmitter interface {
	SubmitIdentityL1(ctx context.Context, tx interface{}) (string, error)
}

// IdentitySubmitterFunc adapts a function to IdentitySubmitter, for tests and small adapters.
type IdentitySubmitterFunc func(ctx context.Context, tx interface{}) (string, error)

// SubmitIdentityL1 implements IdentitySubmitter.
func (f IdentitySubmitterFunc) SubmitIdentityL1(ctx context.Context, tx interface{}) (string, error) {
	return f(ctx, tx)
}

var errNoChain = errors.New("chain: not configured")

// HasIdentity reports whether identity records can be written (the identity contract is set).
func (a *BesuAnchor) HasIdentity() bool {
	return a != nil && a.Client != nil && a.Client.IdentityContract() != (common.Address{})
}

// SubmitIdentityL1 implements IdentitySubmitter. It returns once the record is final.
func (a *BesuAnchor) SubmitIdentityL1(ctx context.Context, tx interface{}) (string, error) {
	if a == nil || a.Client == nil {
		return "", errNoChain
	}
	switch u := tx.(type) {
	case DeviceKeyRegistrationUpdate:
		// The device label is free text chosen by the user; it stays in the gateway's database.
		return txRef(a.Client.RegisterDeviceKey(ctx, u.SubjectDID, u.PublicKeyHex))
	case TrustTierCommitmentUpdate:
		return txRef(a.Client.AnchorTrustTier(ctx, u.SubjectDID, u.Commitment))
	case UsernameRegistrationUpdate:
		return txRef(a.Client.RegisterUsername(ctx, u.Username, u.SubjectDID))
	default:
		return "", fmt.Errorf("%w (%T)", ErrUnsupportedUpdate, tx)
	}
}

// PublishCredentialIssuance implements credentials.ChainPublisher.
func (a *BesuAnchor) PublishCredentialIssuance(ctx context.Context, credentialID, subjectDID, issuerDID, kind string) (string, error) {
	if a == nil || a.Client == nil {
		return "", errNoChain
	}
	return txRef(a.Client.IssueCredential(ctx, credentialID, subjectDID, issuerDID, kind))
}

// PublishStatusList implements credentials.ChainPublisher. The publisher counts publications
// from zero; the chain requires a positive, strictly increasing sequence, so it is shifted by one.
func (a *BesuAnchor) PublishStatusList(ctx context.Context, issuerDID string, sequence int64, list []byte) error {
	if a == nil || a.Client == nil {
		return errNoChain
	}
	_, err := a.Client.PublishStatusList(ctx, issuerDID, sequence+1, list)
	return err
}

// attest anchors the SHA-256 of a record's JSON under the record's "type". Used for records that
// are too rich for a fixed on-chain shape and hold identifiers (a group's admin DID, a voter's
// DID): the chain gets one digest, and the record stays with the gateway. To check a record later,
// re-serialise it the same way and compare digests.
//
// A digest hides a record only as well as the record is unpredictable. A vote ("yes" by a known
// DID on a known proposal) can be confirmed by anyone who guesses it.
func (a *BesuAnchor) attest(ctx context.Context, kind string, record interface{}) (string, error) {
	if kind == "" {
		return "", fmt.Errorf("%w: record has no type", ErrUnsupportedUpdate)
	}
	raw, err := json.Marshal(record) // struct field order is fixed; map keys are sorted
	if err != nil {
		return "", err
	}
	return txRef(a.Client.Attest(ctx, kind, sha256.Sum256(raw)))
}

func txRef(receipt interface{ Ref() string }, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return receipt.Ref(), nil
}
