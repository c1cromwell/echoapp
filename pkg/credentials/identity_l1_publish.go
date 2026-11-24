package credentials

import (
	"context"
	"fmt"
	"time"
)

// ChainPublisher anchors credential records on Echo's chain. Implementations store hashes only:
// no DID, credential id or type name reaches the chain as text.
type ChainPublisher interface {
	// PublishCredentialIssuance records that a credential of a kind was issued to a subject by
	// an issuer. It returns a reference to the on-chain write.
	PublishCredentialIssuance(ctx context.Context, credentialID, subjectDID, issuerDID, kind string) (string, error)
	// PublishStatusList publishes an issuer's revocation list. sequence is the publisher's
	// zero-based count of publications; a later list always has a higher sequence.
	PublishStatusList(ctx context.Context, issuerDID string, sequence int64, list []byte) error
}

// anchoringEnabled reports whether issuance records and revocation lists go on-chain.
func (c MetagraphConfig) anchoringEnabled() bool {
	return c.EnableAnchor && c.Chain != nil
}

// publishVCIssuanceMetadata anchors a credential's issuance record (WO-274). The issue time is
// the block's time; the caller's clock is not written.
func (i *Issuer) publishVCIssuanceMetadata(ctx context.Context, credentialID string, req *CredentialIssuanceRequest, _ time.Time) (string, error) {
	if !i.config.MetagraphConfig.anchoringEnabled() {
		return "", nil
	}
	timeout := i.config.MetagraphConfig.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ref, err := i.config.MetagraphConfig.Chain.PublishCredentialIssuance(ctx,
		fmt.Sprintf("urn:uuid:%s", credentialID), req.SubjectDID, i.effectiveIssuerDID(), string(req.CredentialType))
	if err != nil {
		return "", fmt.Errorf("anchor credential issuance: %w", err)
	}
	return ref, nil
}
