package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// /v1/enrollment/wallet is unauthenticated, so what it stores must stay
// bounded no matter how many DIDs it is called with.
func TestEnrollmentWallet_ProvisionalEntriesAreCappedAndExpire(t *testing.T) {
	rt := &Router{}
	start := time.Now()

	for i := 0; i < enrollmentWalletMaxProvisional+500; i++ {
		did := fmt.Sprintf("did:key:zFlood%d", i)
		if addr := rt.provisionEnrollmentWallet(did, start); addr != deterministicDAGAddress(did) {
			t.Fatalf("address for %s = %q", did, addr)
		}
	}
	if n := len(rt.enrollmentWalletByDID); n != enrollmentWalletMaxProvisional {
		t.Fatalf("map holds %d entries, want the cap %d", n, enrollmentWalletMaxProvisional)
	}

	// An authenticated link is permanent: it is neither counted nor expired.
	rt.enrollmentWalletMu.Lock()
	rt.enrollmentWalletByDID["did:key:zLinked"] = "DAGlinked"
	rt.enrollmentWalletMu.Unlock()

	// Once the provisional entries are older than the TTL they are evicted
	// and new ones fit again.
	later := start.Add(enrollmentWalletProvisionalTTL + time.Minute)
	rt.provisionEnrollmentWallet("did:key:zFresh", later)
	rt.enrollmentWalletMu.Lock()
	defer rt.enrollmentWalletMu.Unlock()
	if n := len(rt.enrollmentWalletProvisional); n != 1 {
		t.Fatalf("after expiry %d provisional entries remain, want 1", n)
	}
	if len(rt.enrollmentWalletByDID) != 2 || rt.enrollmentWalletByDID["did:key:zLinked"] != "DAGlinked" || rt.enrollmentWalletByDID["did:key:zFresh"] == "" {
		t.Fatalf("after expiry map has %d entries", len(rt.enrollmentWalletByDID))
	}
}

func TestEnrollmentWallet_EndpointStillAnswers(t *testing.T) {
	rt := NewRouter([]string{"*"})
	rec := secPostJSON(t, rt, "/v1/enrollment/wallet", map[string]string{"did": "did:key:zNew"}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d %s", rec.Code, rec.Body.String())
	}
	again := secPostJSON(t, rt, "/v1/enrollment/wallet", map[string]string{"did": "did:key:zNew"}, "")
	if again.Code != http.StatusOK {
		t.Fatalf("repeat: %d", again.Code)
	}
	long := make([]byte, enrollmentWalletMaxDIDLen+1)
	for i := range long {
		long[i] = 'a'
	}
	if rec := secPostJSON(t, rt, "/v1/enrollment/wallet", map[string]string{"did": string(long)}, ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized did: want 400, got %d", rec.Code)
	}
}

// One wallet address belongs to one DID.
func TestLinkWallet_RejectsAddressLinkedToAnotherDID(t *testing.T) {
	rt := NewRouter([]string{"*"})
	alice := accessTokenFor(t, rt, "did:key:zAlice")
	bob := accessTokenFor(t, rt, "did:key:zBob")
	link := func(token, did, addr string) int {
		return secPostJSON(t, rt, "/v1/identity/link-wallet", map[string]string{"did": did, "wallet_address": addr}, token).Code
	}

	if code := link(alice, "did:key:zAlice", "DAGshared"); code != http.StatusOK {
		t.Fatalf("first link: %d", code)
	}
	if code := link(bob, "did:key:zBob", "DAGshared"); code != http.StatusConflict {
		t.Fatalf("second DID, same address: want 409, got %d", code)
	}
	rt.enrollmentWalletMu.Lock()
	_, bobLinked := rt.enrollmentWalletByDID["did:key:zBob"]
	rt.enrollmentWalletMu.Unlock()
	if bobLinked {
		t.Fatal("a refused link must not be stored")
	}
	// Re-linking your own address, and linking a different one, still work.
	if code := link(alice, "did:key:zAlice", "DAGshared"); code != http.StatusOK {
		t.Fatalf("idempotent re-link: %d", code)
	}
	if code := link(bob, "did:key:zBob", "DAGbob"); code != http.StatusOK {
		t.Fatalf("bob's own address: %d", code)
	}
	// The address the server derived for another DID is taken too.
	rt.provisionEnrollmentWallet("did:key:zCarol", time.Now())
	if code := link(bob, "did:key:zBob", deterministicDAGAddress("did:key:zCarol")); code != http.StatusConflict {
		t.Fatalf("another DID's derived address: want 409, got %d", code)
	}
}
