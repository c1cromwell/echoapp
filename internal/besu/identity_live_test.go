package besu

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func liveIdentity(t *testing.T) (*Client, context.Context) {
	t.Helper()
	c, _ := liveClient(t)
	if c.IdentityContract() == (common.Address{}) {
		t.Skip("BESU_IDENTITY_CONTRACT not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	return c, ctx
}

// unique makes identifiers that no earlier run has used.
func unique(t *testing.T, prefix string) string {
	t.Helper()
	return prefix + randomHash(t)[:20]
}

func TestLiveDeviceKey(t *testing.T) {
	c, ctx := liveIdentity(t)
	did, key := unique(t, "did:key:z"), "04"+randomHash(t)+randomHash(t)
	if at, err := c.DeviceKeyRegisteredAt(ctx, did, key); err != nil || at != 0 {
		t.Fatalf("before: %d, %v", at, err)
	}
	receipt, err := c.RegisterDeviceKey(ctx, did, key)
	if err != nil {
		t.Fatal(err)
	}
	// Hex case must not matter: the same key is the same key.
	if at, err := c.DeviceKeyRegisteredAt(ctx, did, strings.ToUpper(key)); err != nil || at != receipt.BlockNumber {
		t.Fatalf("after: block %d, %v; want %d", at, err, receipt.BlockNumber)
	}
	if _, err := c.RegisterDeviceKey(ctx, did, key); err == nil || !strings.Contains(err.Error(), "AlreadyRecorded") {
		t.Fatalf("registering the same key twice: %v", err)
	}
	// The same key under another DID is a different record.
	if at, _ := c.DeviceKeyRegisteredAt(ctx, unique(t, "did:key:z"), key); at != 0 {
		t.Fatal("key appears registered to a DID it was not registered for")
	}
}

func TestLiveTrustTier(t *testing.T) {
	c, ctx := liveIdentity(t)
	did, first, second := unique(t, "did:key:z"), randomHash(t), randomHash(t)
	if _, err := c.AnchorTrustTier(ctx, did, first); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AnchorTrustTier(ctx, did, second); err != nil {
		t.Fatal(err)
	}
	current, err := c.CurrentTrustTier(ctx, did)
	if err != nil || current != common.HexToHash(second) {
		t.Fatalf("current = %s, %v; want the second commitment", current.Hex(), err)
	}
	for _, commitment := range []string{first, second} {
		if at, err := c.TrustTierAnchoredAt(ctx, commitment); err != nil || at == 0 {
			t.Fatalf("commitment %s: block %d, %v", commitment[:8], at, err)
		}
	}
	// A commitment is H(tier || nonce); replaying one must be refused.
	if _, err := c.AnchorTrustTier(ctx, did, first); err == nil || !strings.Contains(err.Error(), "AlreadyRecorded") {
		t.Fatalf("replayed commitment: %v", err)
	}
	if _, err := c.AnchorTrustTier(ctx, did, "not-a-hash"); err == nil {
		t.Fatal("malformed commitment accepted")
	}
}

func TestLiveUsername(t *testing.T) {
	c, ctx := liveIdentity(t)
	name := unique(t, "User_")
	alice, bob := unique(t, "did:key:zA"), unique(t, "did:key:zB")

	if registered, _, err := c.UsernameOwnedBy(ctx, name, alice); err != nil || registered {
		t.Fatalf("before: registered=%v, %v", registered, err)
	}
	if _, err := c.RegisterUsername(ctx, name, alice); err != nil {
		t.Fatal(err)
	}
	// Registering one's own name again is allowed and changes nothing.
	if _, err := c.RegisterUsername(ctx, strings.ToUpper(name), alice); err != nil {
		t.Fatalf("re-registering one's own name (different case): %v", err)
	}
	// Someone else cannot take it, in any letter case.
	for _, attempt := range []string{name, strings.ToLower(name), strings.ToUpper(name)} {
		if _, err := c.RegisterUsername(ctx, attempt, bob); !errors.Is(err, ErrUsernameTaken) {
			t.Fatalf("bob registering %q: %v, want ErrUsernameTaken", attempt, err)
		}
	}
	if registered, mine, err := c.UsernameOwnedBy(ctx, name, alice); err != nil || !registered || !mine {
		t.Fatalf("alice: registered=%v mine=%v err=%v", registered, mine, err)
	}
	if registered, mine, err := c.UsernameOwnedBy(ctx, name, bob); err != nil || !registered || mine {
		t.Fatalf("bob: registered=%v mine=%v err=%v", registered, mine, err)
	}
}

func TestLiveCredentialAndStatusList(t *testing.T) {
	c, ctx := liveIdentity(t)
	credential, subject, issuer := unique(t, "urn:uuid:"), unique(t, "did:key:zS"), unique(t, "did:key:zI")

	receipt, err := c.IssueCredential(ctx, credential, subject, issuer, "TrustTierCredential")
	if err != nil {
		t.Fatal(err)
	}
	if at, err := c.CredentialIssuedAt(ctx, credential); err != nil || at != receipt.BlockNumber {
		t.Fatalf("issued at %d, %v", at, err)
	}
	if _, err := c.IssueCredential(ctx, credential, subject, issuer, "TrustTierCredential"); err == nil {
		t.Fatal("the same credential was issued twice")
	}

	list := []byte{0x00, 0x01, 0x02}
	if _, err := c.PublishStatusList(ctx, issuer, 5, list); err != nil {
		t.Fatal(err)
	}
	// An older or equal sequence must never replace a newer list.
	for _, seq := range []int64{5, 4, 1} {
		if _, err := c.PublishStatusList(ctx, issuer, seq, []byte{0xff}); err == nil || !strings.Contains(err.Error(), "StaleSequence") {
			t.Fatalf("sequence %d after 5: %v", seq, err)
		}
	}
	newer := []byte{0x00, 0x01, 0x03}
	if _, err := c.PublishStatusList(ctx, issuer, 6, newer); err != nil {
		t.Fatal(err)
	}
	current, err := c.CurrentStatusList(ctx, issuer)
	if err != nil || current.Sequence != 6 || current.ListHash != ListHash(newer) {
		t.Fatalf("current list: %+v, %v", current, err)
	}
}

func TestLiveAttestation(t *testing.T) {
	c, ctx := liveIdentity(t)
	digest := common.HexToHash(randomHash(t))
	if _, err := c.Attest(ctx, "group_metadata", digest); err != nil {
		t.Fatal(err)
	}
	if at, err := c.AttestedAt(ctx, "group_metadata", digest); err != nil || at == 0 {
		t.Fatalf("attested at %d, %v", at, err)
	}
	// The same digest under another kind is a separate record.
	if at, _ := c.AttestedAt(ctx, "governance_vote", digest); at != 0 {
		t.Fatal("digest appears under a kind it was not attested under")
	}
	if _, err := c.Attest(ctx, "group_metadata", digest); err == nil {
		t.Fatal("the same digest was attested twice")
	}
}

func TestLiveIdentityUnlistedSignerIsRejected(t *testing.T) {
	c, ctx := liveIdentity(t)
	stranger, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	other, err := New(Config{RPCURLs: c.cfg.RPCURLs, AnchorContract: c.cfg.AnchorContract, IdentityContract: c.cfg.IdentityContract, Signer: stranger})
	if err != nil {
		t.Fatal(err)
	}
	name, did := unique(t, "u"), unique(t, "did:key:z")
	if _, err := other.RegisterUsername(ctx, name, did); err == nil || !strings.Contains(err.Error(), "NotAuthorisedSigner") {
		t.Fatalf("unlisted signer registering a username: %v", err)
	}
	if registered, _, _ := c.UsernameOwnedBy(ctx, name, did); registered {
		t.Fatal("the rejected username is on the chain")
	}
}
