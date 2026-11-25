package ledger

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

// The gateway's identity path against a live chain: the three record types it writes itself, the
// two the credentials package writes, and a group record anchored as a digest.
func TestLiveIdentityRecordsThroughTheGatewayPath(t *testing.T) {
	anchor, err := BesuAnchorFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if anchor == nil || !anchor.HasIdentity() {
		t.Skip("live chain with BESU_IDENTITY_CONTRACT not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	suffix := make([]byte, 10)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatal(err)
	}
	tag := hex.EncodeToString(suffix)
	did, username, key := "did:key:zLive"+tag, "live_"+tag, "04"+tag+tag
	commitment := TrustTierCommitmentHex(2, tag)
	client := anchor.Client

	for name, tx := range map[string]interface{}{
		"device key": DeviceKeyRegistrationUpdate{SubjectDID: did, PublicKeyHex: key, DeviceLabel: "Alice's iPhone", AddedAt: time.Now().UnixMilli()},
		"trust tier": TrustTierCommitmentUpdate{SubjectDID: did, Commitment: commitment, AnchoredAt: time.Now().UnixMilli()},
		"username":   UsernameRegistrationUpdate{SubjectDID: did, Username: username, RegisteredAt: time.Now().UnixMilli()},
	} {
		if ref, err := anchor.SubmitIdentityL1(ctx, tx); err != nil || ref == "" {
			t.Fatalf("%s: ref=%q err=%v", name, ref, err)
		}
	}
	if at, err := client.DeviceKeyRegisteredAt(ctx, did, key); err != nil || at == 0 {
		t.Fatalf("device key: block %d, %v", at, err)
	}
	if at, err := client.TrustTierAnchoredAt(ctx, commitment); err != nil || at == 0 {
		t.Fatalf("trust tier: block %d, %v", at, err)
	}
	if registered, mine, err := client.UsernameOwnedBy(ctx, username, did); err != nil || !registered || !mine {
		t.Fatalf("username: registered=%v mine=%v err=%v", registered, mine, err)
	}
	// A second DID asking for the same username is refused by the chain.
	if _, err := anchor.SubmitIdentityL1(ctx, UsernameRegistrationUpdate{SubjectDID: did + "x", Username: username}); !errors.Is(err, besu.ErrUsernameTaken) {
		t.Fatalf("taken username: %v", err)
	}

	// Credentials package path.
	credential, issuer := "urn:uuid:"+tag, "did:key:zIssuer"+tag
	if ref, err := anchor.PublishCredentialIssuance(ctx, credential, did, issuer, "TrustTierCredential"); err != nil || ref == "" {
		t.Fatalf("credential: ref=%q err=%v", ref, err)
	}
	if at, err := client.CredentialIssuedAt(ctx, credential); err != nil || at == 0 {
		t.Fatalf("credential: block %d, %v", at, err)
	}
	// The publisher counts publications from zero; the first two must both be accepted, in order.
	for seq, list := range [][]byte{{0x01}, {0x03}} {
		if err := anchor.PublishStatusList(ctx, issuer, int64(seq), list); err != nil {
			t.Fatalf("status list publication %d: %v", seq, err)
		}
	}
	current, err := client.CurrentStatusList(ctx, issuer)
	if err != nil || current.Sequence != 2 || current.ListHash != besu.ListHash([]byte{0x03}) {
		t.Fatalf("status list: %+v, %v", current, err)
	}
	// Publishing an old publication again must be refused.
	if err := anchor.PublishStatusList(ctx, issuer, 0, []byte{0x01}); err == nil {
		t.Fatal("a stale revocation list replaced a newer one")
	}

	// A group record holds an admin DID; only its digest is anchored.
	group := GroupMetadataUpdate{Type: "group_metadata", GroupID: "g-" + tag, AdminDID: did, MemberCountHash: commitment, CreatedAt: 1}
	if ref, err := anchor.SubmitDataL1(ctx, group); err != nil || ref == "" {
		t.Fatalf("group metadata: ref=%q err=%v", ref, err)
	}
	raw, _ := json.Marshal(group)
	if at, err := client.AttestedAt(ctx, "group_metadata", sha256.Sum256(raw)); err != nil || at == 0 {
		t.Fatalf("group digest: block %d, %v", at, err)
	}
	// A governance vote arrives as a map; the same vote always gives the same digest.
	vote := map[string]interface{}{"type": "governance_vote", "proposalId": "p-" + tag, "voterDid": did, "value": "yes"}
	if _, err := anchor.SubmitDataL1(ctx, vote); err != nil {
		t.Fatalf("vote: %v", err)
	}
	if _, err := anchor.SubmitDataL1(ctx, map[string]interface{}{"value": "yes", "voterDid": did, "proposalId": "p-" + tag, "type": "governance_vote"}); err == nil {
		t.Fatal("the same vote was anchored twice (key order must not change the digest)")
	}
}
