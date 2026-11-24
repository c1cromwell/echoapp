package besu

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// ErrNoIdentityContract means the client was built without an EchoIdentity address.
var ErrNoIdentityContract = errors.New("besu: identity contract is not configured")

// ErrUsernameTaken means the username is registered to a different subject.
var ErrUsernameTaken = errors.New("besu: username is registered to a different DID")

func (c *Client) identityWrite(ctx context.Context, method string, args ...interface{}) (Receipt, error) {
	if c.cfg.IdentityContract == (common.Address{}) {
		return Receipt{}, ErrNoIdentityContract
	}
	return c.transact(ctx, c.cfg.IdentityContract, identityABI, method, args...)
}

func (c *Client) identityRead(ctx context.Context, method string, args ...interface{}) ([]interface{}, error) {
	if c.cfg.IdentityContract == (common.Address{}) {
		return nil, ErrNoIdentityContract
	}
	ec, err := c.conn(ctx)
	if err != nil {
		return nil, err
	}
	var out []interface{}
	if err := bind.NewBoundContract(c.cfg.IdentityContract, identityABI, ec, ec, ec).Call(&bind.CallOpts{Context: ctx}, &out, method, args...); err != nil {
		return nil, fmt.Errorf("besu: read %s: %w", method, err)
	}
	return out, nil
}

func requireText(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("besu: %s must not be empty", field)
	}
	return nil
}

// RegisterDeviceKey records that a DID has a device key. Both are stored as hashes.
func (c *Client) RegisterDeviceKey(ctx context.Context, did, publicKeyHex string) (Receipt, error) {
	if err := errors.Join(requireText("did", did), requireText("public key", publicKeyHex)); err != nil {
		return Receipt{}, err
	}
	return c.identityWrite(ctx, "registerDeviceKey", SubjectID(did), DeviceKeyID(publicKeyHex))
}

// DeviceKeyRegisteredAt returns the block a device key was registered in, or 0 if it is not.
func (c *Client) DeviceKeyRegisteredAt(ctx context.Context, did, publicKeyHex string) (uint64, error) {
	out, err := c.identityRead(ctx, "deviceKeyAt", SubjectID(did), DeviceKeyID(publicKeyHex))
	if err != nil {
		return 0, err
	}
	return out[0].(uint64), nil
}

// AnchorTrustTier records a trust-tier commitment, H(tier || nonce), for a DID. The commitment is
// already a hash; the DID is stored as one.
func (c *Client) AnchorTrustTier(ctx context.Context, did, commitment string) (Receipt, error) {
	if err := requireText("did", did); err != nil {
		return Receipt{}, err
	}
	h, err := parseHash32("commitment", commitment)
	if err != nil {
		return Receipt{}, err
	}
	return c.identityWrite(ctx, "anchorTrustTier", SubjectID(did), h)
}

// TrustTierAnchoredAt returns the block a commitment was anchored in, or 0 if it is not.
func (c *Client) TrustTierAnchoredAt(ctx context.Context, commitment string) (uint64, error) {
	h, err := parseHash32("commitment", commitment)
	if err != nil {
		return 0, err
	}
	out, err := c.identityRead(ctx, "trustTierAt", h)
	if err != nil {
		return 0, err
	}
	return out[0].(uint64), nil
}

// CurrentTrustTier returns a DID's latest commitment, or the zero hash if it has none.
func (c *Client) CurrentTrustTier(ctx context.Context, did string) (common.Hash, error) {
	out, err := c.identityRead(ctx, "trustTierOf", SubjectID(did))
	if err != nil {
		return common.Hash{}, err
	}
	return out[0].([32]byte), nil
}

// RegisterUsername binds a username to a DID. Registering a name the DID already owns succeeds
// and changes nothing; a name owned by another DID returns ErrUsernameTaken.
func (c *Client) RegisterUsername(ctx context.Context, username, did string) (Receipt, error) {
	if err := errors.Join(requireText("username", username), requireText("did", did)); err != nil {
		return Receipt{}, err
	}
	receipt, err := c.identityWrite(ctx, "registerUsername", UsernameID(username), SubjectID(did))
	if err != nil && strings.Contains(err.Error(), "UsernameTaken") {
		return Receipt{}, ErrUsernameTaken
	}
	return receipt, err
}

// UsernameOwnedBy reports whether the username is registered, and if so whether to this DID.
func (c *Client) UsernameOwnedBy(ctx context.Context, username, did string) (registered, ownedByDID bool, err error) {
	out, err := c.identityRead(ctx, "usernameOwner", UsernameID(username))
	if err != nil {
		return false, false, err
	}
	owner := common.Hash(out[0].([32]byte))
	if owner == (common.Hash{}) {
		return false, false, nil
	}
	return true, owner == SubjectID(did), nil
}

// IssueCredential records that a credential of a kind was issued to a subject by an issuer.
func (c *Client) IssueCredential(ctx context.Context, credentialID, subjectDID, issuerDID, kind string) (Receipt, error) {
	if err := errors.Join(requireText("credential id", credentialID), requireText("subject did", subjectDID),
		requireText("issuer did", issuerDID), requireText("credential kind", kind)); err != nil {
		return Receipt{}, err
	}
	return c.identityWrite(ctx, "issueCredential", CredentialID(credentialID), SubjectID(subjectDID), SubjectID(issuerDID), CredentialKindID(kind))
}

// CredentialIssuedAt returns the block a credential was issued in, or 0 if there is no record.
func (c *Client) CredentialIssuedAt(ctx context.Context, credentialID string) (uint64, error) {
	out, err := c.identityRead(ctx, "credentials", CredentialID(credentialID))
	if err != nil {
		return 0, err
	}
	return out[3].(uint64), nil
}

// PublishStatusList publishes the hash of an issuer's revocation list. sequence must be greater
// than the issuer's last published sequence.
func (c *Client) PublishStatusList(ctx context.Context, issuerDID string, sequence int64, list []byte) (Receipt, error) {
	if err := requireText("issuer did", issuerDID); err != nil {
		return Receipt{}, err
	}
	if sequence <= 0 {
		return Receipt{}, errors.New("besu: status list sequence must be positive")
	}
	if len(list) == 0 {
		return Receipt{}, errors.New("besu: status list is empty")
	}
	return c.identityWrite(ctx, "publishStatusList", SubjectID(issuerDID), uint64(sequence), ListHash(list))
}

// StatusList is an issuer's current revocation list as the chain records it.
type StatusList struct {
	Sequence    uint64
	ListHash    common.Hash
	BlockNumber uint64
}

// CurrentStatusList returns an issuer's current list; Sequence 0 means none was published.
func (c *Client) CurrentStatusList(ctx context.Context, issuerDID string) (StatusList, error) {
	out, err := c.identityRead(ctx, "statusLists", SubjectID(issuerDID))
	if err != nil {
		return StatusList{}, err
	}
	return StatusList{Sequence: out[0].(uint64), ListHash: out[1].([32]byte), BlockNumber: out[2].(uint64)}, nil
}

// Attest anchors the digest of a record of the given kind (for example a group's metadata). The
// caller hashes the record; only the digest goes on-chain.
func (c *Client) Attest(ctx context.Context, kind string, digest common.Hash) (Receipt, error) {
	if err := requireText("kind", kind); err != nil {
		return Receipt{}, err
	}
	if digest == (common.Hash{}) {
		return Receipt{}, errors.New("besu: digest must not be zero")
	}
	return c.identityWrite(ctx, "attest", AttestationKindID(kind), digest)
}

// AttestedAt returns the block a digest was attested in, or 0 if it was not.
func (c *Client) AttestedAt(ctx context.Context, kind string, digest common.Hash) (uint64, error) {
	out, err := c.identityRead(ctx, "attestedAt", AttestationKindID(kind), digest)
	if err != nil {
		return 0, err
	}
	return out[0].(uint64), nil
}
