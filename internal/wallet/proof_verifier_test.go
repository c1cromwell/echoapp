package wallet

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

// seedChallenge injects a known challenge so the real dag4 vector signature
// (over a fixed message) can drive the full verifier path.
func seedChallenge(cs *ChallengeStore, did, challenge string, expires time.Time) {
	cs.mu.Lock()
	cs.entries[did] = challengeEntry{challenge: challenge, expires: expires}
	cs.mu.Unlock()
}

func proofFor(v walletVector) string {
	b, _ := json.Marshal(WalletProof{PublicKey: v.PublicKey, Challenge: v.Message, Signature: v.Signature})
	return base64.StdEncoding.EncodeToString(b)
}

func TestDagProofVerifier_LinkBindsRealSignature(t *testing.T) {
	v := loadWalletVector(t)
	cs := NewChallengeStore()
	store := NewMemStore()
	verifier := NewDagProofVerifier(store, cs)
	did := "did:key:zProof"
	seedChallenge(cs, did, v.Message, time.Now().Add(time.Minute))

	// Link mode: address derives from the proof's public key.
	pub, err := verifier.VerifyOwnership(did, v.Address, proofFor(v))
	if err != nil {
		t.Fatalf("link verify failed: %v", err)
	}
	if pub != v.PublicKey {
		t.Fatalf("returned pubkey %q != %q", pub, v.PublicKey)
	}

	// Replay: the challenge was consumed, so a second attempt must fail.
	seedChallenge(cs, did, v.Message, time.Now().Add(time.Minute))
	cs.Consume(did, v.Message) // burn it
	if _, err := verifier.VerifyOwnership(did, v.Address, proofFor(v)); err != errProofChallenge {
		t.Fatalf("expected errProofChallenge on replay, got %v", err)
	}
}

func TestDagProofVerifier_RejectsWrongAddress(t *testing.T) {
	v := loadWalletVector(t)
	cs := NewChallengeStore()
	verifier := NewDagProofVerifier(NewMemStore(), cs)
	did := "did:key:zProof"
	seedChallenge(cs, did, v.Message, time.Now().Add(time.Minute))

	if _, err := verifier.VerifyOwnership(did, "DAG0wrongaddress0000000000000000000000000", proofFor(v)); err != errProofAddrBinding {
		t.Fatalf("expected errProofAddrBinding, got %v", err)
	}
}

func TestDagProofVerifier_ValueOpRequiresBoundKey(t *testing.T) {
	v := loadWalletVector(t)
	cs := NewChallengeStore()
	store := NewMemStore()
	verifier := NewDagProofVerifier(store, cs)
	did := "did:key:zProof"

	// Not linked yet -> value op (address="") must fail.
	seedChallenge(cs, did, v.Message, time.Now().Add(time.Minute))
	if _, err := verifier.VerifyOwnership(did, "", proofFor(v)); err != errProofNotLinked {
		t.Fatalf("expected errProofNotLinked, got %v", err)
	}

	// Bind the key, then the value op verifies.
	_ = store.LinkDAGAccount(context.Background(), did, v.Address, v.PublicKey)
	seedChallenge(cs, did, v.Message, time.Now().Add(time.Minute))
	if _, err := verifier.VerifyOwnership(did, "", proofFor(v)); err != nil {
		t.Fatalf("value op with bound key failed: %v", err)
	}

	// A different bound key must be rejected.
	_ = store.LinkDAGAccount(context.Background(), did, v.Address, "04deadbeef")
	seedChallenge(cs, did, v.Message, time.Now().Add(time.Minute))
	if _, err := verifier.VerifyOwnership(did, "", proofFor(v)); err != errProofKeyMismatch {
		t.Fatalf("expected errProofKeyMismatch, got %v", err)
	}
}

func TestDagProofVerifier_RejectsTamperedSignature(t *testing.T) {
	v := loadWalletVector(t)
	cs := NewChallengeStore()
	verifier := NewDagProofVerifier(NewMemStore(), cs)
	did := "did:key:zProof"
	seedChallenge(cs, did, v.Message, time.Now().Add(time.Minute))

	bad := v
	// Flip a hex nibble in the signature.
	sig := []byte(bad.Signature)
	if sig[len(sig)-1] == 'a' {
		sig[len(sig)-1] = 'b'
	} else {
		sig[len(sig)-1] = 'a'
	}
	bad.Signature = string(sig)
	if _, err := verifier.VerifyOwnership(did, v.Address, proofFor(bad)); err != errProofSignature {
		t.Fatalf("expected errProofSignature, got %v", err)
	}
}

func TestChallengeStore_IssueConsumeExpire(t *testing.T) {
	cs := NewChallengeStore()
	did := "did:key:zC"
	ch, _, err := cs.Issue(did)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	// Wrong challenge rejected.
	if cs.Consume(did, "nope") {
		t.Fatal("consumed wrong challenge")
	}
	// Correct challenge consumed once.
	if !cs.Consume(did, ch) {
		t.Fatal("failed to consume valid challenge")
	}
	if cs.Consume(did, ch) {
		t.Fatal("challenge consumed twice (replay)")
	}
	// Expiry.
	ch2, _, _ := cs.Issue(did)
	cs.now = func() time.Time { return time.Now().Add(10 * time.Minute) }
	if cs.Consume(did, ch2) {
		t.Fatal("expired challenge accepted")
	}
}
