package metagraph

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/besu"
)

// Runs only against a live chain (`make dev`, then `make chain-test`): anchors a real batch
// through the same code path the gateway uses and checks the served bundle verifies offline.
// With BESU_UPDATE_FIXTURE=1 it also records the bundle used by cmd/verifyproof's tests.
func TestLiveAnchorBatchAndServeVerifiableBundle(t *testing.T) {
	validatorList := os.Getenv("BESU_VALIDATORS")
	anchor, err := BesuAnchorFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if anchor == nil || validatorList == "" {
		t.Skip("live chain not configured (BESU_RPC_URLS, BESU_ANCHOR_CONTRACT, BESU_VALIDATORS, IDENTITY_SERVICE_KEY_PEM)")
	}
	validators, err := besu.ParseAddresses(validatorList)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Three message commitments, unique per run so the root has never been anchored.
	leaves := make([]string, 3)
	for i := range leaves {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			t.Fatal(err)
		}
		leaves[i] = hex.EncodeToString(b)
	}
	tree := BuildMerkleTree(leaves)
	store := NewMemoryProofStore()

	if _, err := AnchorRoot(ctx, anchor, store, ChainNetworkBesu, tree.Root, len(leaves)); err != nil {
		t.Fatal(err)
	}
	if err := store.Put(ctx, MessageAnchorProof{MessageID: "msg-1", Commitment: leaves[1], Siblings: tree.ProofSiblings(1), MerkleLeafIndex: 1, MerkleRoot: tree.Root}); err != nil {
		t.Fatal(err)
	}
	bundle, err := BuildProofBundle(ctx, store, "", tree.Root, "msg-1")
	if err != nil {
		t.Fatal(err)
	}
	if !bundle.Finalized || bundle.Chain == nil || len(bundle.Chain.Proof) == 0 {
		t.Fatalf("bundle is missing its chain evidence: %+v", bundle)
	}

	// What an outside verifier does, with no access to the chain.
	if err := VerifyMerklePath(bundle.Message.Commitment, bundle.Message.Siblings, bundle.Message.LeafIndex, bundle.LeafCount, bundle.MerkleRoot); err != nil {
		t.Fatal(err)
	}
	var proof besu.Proof
	if err := json.Unmarshal(bundle.Chain.Proof, &proof); err != nil {
		t.Fatal(err)
	}
	v, err := besu.Verify(&proof, validators)
	if err != nil {
		t.Fatalf("served bundle does not verify offline: %v", err)
	}
	if int(v.LeafCount) != len(leaves) || int64(v.AnchoredBlock) != bundle.Chain.BlockNumber {
		t.Fatalf("verified %+v does not match bundle %+v", v, bundle.Chain)
	}

	if os.Getenv("BESU_UPDATE_FIXTURE") == "1" {
		body, _ := json.MarshalIndent(struct {
			Validators string       `json:"validators"`
			Bundle     *ProofBundle `json:"bundle"`
		}{validatorList, bundle}, "", "  ")
		if err := os.WriteFile("../../cmd/verifyproof/testdata/bundle.json", append(body, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
