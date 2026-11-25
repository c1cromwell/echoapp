package ledger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/besu"
)

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func testLeaves(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = sha256Hex([]byte(fmt.Sprintf("leaf-%d", i)))
	}
	return out
}

func TestVerifyMerklePath_AllLeavesAllSizes(t *testing.T) {
	for n := 1; n <= 17; n++ {
		leaves := testLeaves(n)
		tree := BuildMerkleTree(leaves)
		for i := range leaves {
			if err := VerifyMerklePath(leaves[i], tree.ProofSiblings(i), i, n, tree.Root); err != nil {
				t.Fatalf("n=%d leaf=%d: %v", n, i, err)
			}
		}
	}
}

func TestVerifyMerklePath_Failures(t *testing.T) {
	leaves := testLeaves(5)
	tree := BuildMerkleTree(leaves)
	sib := tree.ProofSiblings(2)
	other := sha256Hex([]byte("other"))
	tampered := append([]string(nil), sib...)
	tampered[0] = other

	cases := map[string]error{
		"wrong commitment": VerifyMerklePath(other, sib, 2, 5, tree.Root),
		"wrong root":       VerifyMerklePath(leaves[2], sib, 2, 5, other),
		"tampered sibling": VerifyMerklePath(leaves[2], tampered, 2, 5, tree.Root),
		"wrong index":      VerifyMerklePath(leaves[2], sib, 3, 5, tree.Root),
		"wrong leaf count": VerifyMerklePath(leaves[2], sib, 2, 9, tree.Root),
		"short path":       VerifyMerklePath(leaves[2], sib[:len(sib)-1], 2, 5, tree.Root),
		"long path":        VerifyMerklePath(leaves[2], append(append([]string(nil), sib...), other), 2, 5, tree.Root),
		"index range":      VerifyMerklePath(leaves[2], sib, 5, 5, tree.Root),
		"zero leaf count":  VerifyMerklePath(leaves[2], sib, 2, 0, tree.Root),
	}
	for name, err := range cases {
		if err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestParseChainNetwork(t *testing.T) {
	for _, in := range []string{"", "besu", " BESU "} {
		got, err := ParseChainNetwork(in)
		if err != nil || got != ChainNetworkBesu {
			t.Errorf("ParseChainNetwork(%q) = %q, %v", in, got, err)
		}
	}
	// Settings left over from the chains Echo no longer uses must stop startup.
	for _, in := range []string{"sovereign", "constellation", "mainnet", "ethereum"} {
		if _, err := ParseChainNetwork(in); err == nil {
			t.Errorf("ParseChainNetwork(%q) accepted", in)
		}
	}
}

// fakeChain is a DataSubmitter and RootProver that records what it was asked.
type fakeChain struct {
	submitErr error
	proveErr  error
	submitted []interface{}
	proved    []string
}

func (f *fakeChain) SubmitDataL1(_ context.Context, tx interface{}) (string, error) {
	if f.submitErr != nil {
		return "", f.submitErr
	}
	f.submitted = append(f.submitted, tx)
	return "0xtx", nil
}

func (f *fakeChain) ProveRoot(_ context.Context, root string) (ChainAnchor, error) {
	f.proved = append(f.proved, root)
	if f.proveErr != nil {
		return ChainAnchor{}, f.proveErr
	}
	return ChainAnchor{BlockNumber: 42, Contract: "0xc0", Proof: json.RawMessage(`{"root":"x"}`)}, nil
}

func TestAnchorRoot_RecordsFinalityAndProof(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryProofStore()
	chain := &fakeChain{}
	root := sha256Hex([]byte("root"))

	ref, err := AnchorRoot(ctx, chain, store, ChainNetworkBesu, root, 3)
	if err != nil || ref != "0xtx" {
		t.Fatalf("AnchorRoot = %q, %v", ref, err)
	}
	rec, ok, _ := store.GetRoot(ctx, root)
	if !ok || !rec.Finalized || rec.FinalizedAt == nil || rec.LeafCount != 3 || rec.Network != ChainNetworkBesu {
		t.Fatalf("record: %+v", rec)
	}
	if rec.Chain == nil || rec.Chain.BlockNumber != 42 || rec.Chain.TxHash != "0xtx" || len(rec.Chain.Proof) == 0 {
		t.Fatalf("chain evidence: %+v", rec.Chain)
	}
	if incomplete, _ := store.ListIncompleteRoots(ctx, 0); len(incomplete) != 0 {
		t.Fatalf("a fully recorded root is listed as incomplete: %+v", incomplete)
	}
}

func TestAnchorRoot_SubmitFailureRecordsNothing(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryProofStore()
	root := sha256Hex([]byte("root"))
	if _, err := AnchorRoot(ctx, &fakeChain{submitErr: errors.New("chain down")}, store, ChainNetworkBesu, root, 1); err == nil {
		t.Fatal("submit failure not returned")
	}
	if _, ok, _ := store.GetRoot(ctx, root); ok {
		t.Fatal("a root that never reached the chain was recorded")
	}
}

// A proof that cannot be taken right after anchoring must not fail the anchor (the root is on the
// chain) and must be completed later.
func TestAnchorRoot_ProofFailureIsBackfilled(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryProofStore()
	chain := &fakeChain{proveErr: errors.New("node restarting")}
	root := sha256Hex([]byte("root"))

	if _, err := AnchorRoot(ctx, chain, store, ChainNetworkBesu, root, 1); err != nil {
		t.Fatalf("anchor must succeed when only the proof step fails: %v", err)
	}
	rec, _, _ := store.GetRoot(ctx, root)
	if !rec.Finalized || rec.Chain != nil {
		t.Fatalf("record after failed proof: %+v", rec)
	}
	if incomplete, _ := store.ListIncompleteRoots(ctx, 0); len(incomplete) != 1 {
		t.Fatalf("root without a proof must be listed as incomplete, got %d", len(incomplete))
	}

	chain.proveErr = nil
	backfillOnce(ctx, chain, store)
	rec, _, _ = store.GetRoot(ctx, root)
	if rec.Chain == nil || len(rec.Chain.Proof) == 0 {
		t.Fatalf("backfill did not record the proof: %+v", rec)
	}
	if incomplete, _ := store.ListIncompleteRoots(ctx, 0); len(incomplete) != 0 {
		t.Fatal("root still incomplete after backfill")
	}
}

func TestBackfill_RootNotOnChainStaysIncomplete(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryProofStore()
	root := sha256Hex([]byte("never anchored"))
	_ = store.RecordRootSubmitted(ctx, root, 1, ChainNetworkBesu, time.Now())
	backfillOnce(ctx, &fakeChain{proveErr: besu.ErrNotAnchored}, store)
	rec, _, _ := store.GetRoot(ctx, root)
	if rec.Finalized || rec.Chain != nil {
		t.Fatalf("a root the chain does not hold was marked final: %+v", rec)
	}
}

func TestAnchorRoot_WithoutStoreOrProver(t *testing.T) {
	ctx := context.Background()
	if _, err := AnchorRoot(ctx, nil, nil, ChainNetworkBesu, "r", 1); err == nil {
		t.Fatal("nil submitter accepted")
	}
	if ref, err := AnchorRoot(ctx, &fakeChain{}, nil, ChainNetworkBesu, sha256Hex([]byte("r")), 1); err != nil || ref == "" {
		t.Fatalf("anchoring with no store: %q, %v", ref, err)
	}
}

func TestBesuAnchor_RejectsUpdatesWithNoOnChainForm(t *testing.T) {
	a := NewBesuAnchor(nil)
	if _, err := a.SubmitDataL1(context.Background(), DataL1MerkleRootUpdate{}); err == nil {
		t.Fatal("unconfigured anchor accepted a write")
	}
	client, err := besu.New(besu.Config{RPCURLs: []string{"http://127.0.0.1:9"}, AnchorContract: [20]byte{1}})
	if err != nil {
		t.Fatal(err)
	}
	// A type with no on-chain form is refused, never written in some default way.
	type unknown struct{ SubjectDID string }
	if _, err := NewBesuAnchor(client).SubmitDataL1(context.Background(), unknown{"did:key:z123"}); !errors.Is(err, ErrUnsupportedUpdate) {
		t.Fatalf("err = %v, want ErrUnsupportedUpdate", err)
	}
	if _, err := NewBesuAnchor(client).SubmitIdentityL1(context.Background(), unknown{"did:key:z123"}); !errors.Is(err, ErrUnsupportedUpdate) {
		t.Fatalf("identity: err = %v, want ErrUnsupportedUpdate", err)
	}
	// Group metadata carries an admin DID: it goes on-chain only as a digest, through the identity
	// contract. With no identity contract it is refused, not sent anywhere else.
	_, err = NewBesuAnchor(client).SubmitDataL1(context.Background(), GroupMetadataUpdate{Type: "group_metadata", AdminDID: "did:key:z123"})
	if !errors.Is(err, besu.ErrNoIdentityContract) {
		t.Fatalf("err = %v, want ErrNoIdentityContract", err)
	}
	// A record with no type cannot be attested.
	if _, err := NewBesuAnchor(client).SubmitDataL1(context.Background(), map[string]interface{}{"voterDid": "did:key:z1"}); !errors.Is(err, ErrUnsupportedUpdate) {
		t.Fatalf("untyped record: err = %v, want ErrUnsupportedUpdate", err)
	}
}

func TestBesuAnchorFromEnv(t *testing.T) {
	t.Setenv("BESU_RPC_URLS", "")
	if a, err := BesuAnchorFromEnv(); a != nil || err != nil {
		t.Fatalf("unset: %v, %v (want nil, nil: the chain is optional)", a, err)
	}
	t.Setenv("BESU_RPC_URLS", "http://127.0.0.1:8545")
	t.Setenv("BESU_ANCHOR_CONTRACT", "not-an-address")
	if _, err := BesuAnchorFromEnv(); err == nil {
		t.Fatal("malformed contract address accepted")
	}
	t.Setenv("BESU_ANCHOR_CONTRACT", "0x0000000000000000000000000000000000000001")
	t.Setenv("IDENTITY_SERVICE_KEY_PEM", "")
	if _, err := BesuAnchorFromEnv(); err == nil {
		t.Fatal("a chain with no signing key must be a startup error, not silent")
	}
	keyFile := filepath.Join(t.TempDir(), "gateway.key")
	if err := os.WriteFile(keyFile, []byte(strings.Repeat("11", 32)), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("IDENTITY_SERVICE_KEY_PEM", keyFile)
	a, err := BesuAnchorFromEnv()
	if err != nil || a == nil {
		t.Fatalf("complete configuration: %v, %v", a, err)
	}
}

func TestBuildProofBundle(t *testing.T) {
	ctx := context.Background()
	store := NewRedisProofStore(nil, nil) // fallback store must satisfy AnchorStore
	leaves := testLeaves(3)
	tree := BuildMerkleTree(leaves)
	if err := store.RecordRootSubmitted(ctx, tree.Root, 3, ChainNetworkBesu, time.Now()); err != nil {
		t.Fatal(err)
	}
	_ = store.Put(ctx, MessageAnchorProof{MessageID: "m2", Commitment: leaves[2], Siblings: tree.ProofSiblings(2), MerkleLeafIndex: 2, MerkleRoot: tree.Root})
	_ = store.Put(ctx, MessageAnchorProof{MessageID: "other", Commitment: "x", MerkleRoot: "some-other-root"})

	b, err := BuildProofBundle(ctx, store, "fallback", tree.Root, "m2")
	if err != nil {
		t.Fatal(err)
	}
	if b.Network != ChainNetworkBesu || b.Finalized || b.LeafCount != 3 || b.Message == nil {
		t.Fatalf("bundle: %+v", b)
	}
	if err := VerifyMerklePath(b.Message.Commitment, b.Message.Siblings, b.Message.LeafIndex, b.LeafCount, b.MerkleRoot); err != nil {
		t.Fatalf("bundle path does not verify: %v", err)
	}
	if _, err := BuildProofBundle(ctx, store, "", "missing", ""); !errors.Is(err, ErrRootNotFound) {
		t.Fatalf("missing root: %v", err)
	}
	for _, id := range []string{"other", "nope"} {
		if _, err := BuildProofBundle(ctx, store, "", tree.Root, id); !errors.Is(err, ErrMessageNotInRoot) {
			t.Fatalf("message %q: %v", id, err)
		}
	}
}
