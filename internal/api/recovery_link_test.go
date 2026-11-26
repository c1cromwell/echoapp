package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/wallet"
)

func linkWallet(rt *Router, did, address string) int {
	req := httptest.NewRequest(http.MethodPost, "/v1/identity/link-wallet",
		strings.NewReader(`{"did":"`+did+`","wallet_address":"`+address+`"}`))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, did))
	rec := httptest.NewRecorder()
	rt.handleLinkWallet(rec, req)
	return rec.Code
}

// Restoring an account on a new device looks the account up by its recovery address. That link
// has to survive a server restart; it used to live only in the process's memory.
func TestRecoveryLinkSurvivesARestart(t *testing.T) {
	store := wallet.NewMemStore() // stands in for the database, which outlives the process
	const alice, address = "did:key:zAlice", "DAG4realRecoveryAddressOfAlice0000000000"

	before := NewRouter(nil)
	before.RecoveryKeys = store
	if code := linkWallet(before, alice, address); code != http.StatusOK {
		t.Fatalf("link: status %d", code)
	}

	after := NewRouter(nil) // a new process: nothing in memory
	after.RecoveryKeys = store
	if got := after.lookupDIDByWallet(context.Background(), address); got != alice {
		t.Fatalf("after a restart the address resolves to %q, want %q", got, alice)
	}

	without := NewRouter(nil) // and with no durable store the link is, correctly, unknown
	if got := without.lookupDIDByWallet(context.Background(), address); got != "" {
		t.Fatalf("a router with no store resolved the address to %q", got)
	}
}

// One recovery address, one account: a second account cannot claim an address, before or after a
// restart, and so cannot make itself the target of someone else's restore.
func TestRecoveryAddressCannotBeClaimedTwice(t *testing.T) {
	store := wallet.NewMemStore()
	const alice, mallory, address = "did:key:zAlice", "did:key:zMallory", "DAG4realRecoveryAddressOfAlice0000000000"

	rt := NewRouter(nil)
	rt.RecoveryKeys = store
	if code := linkWallet(rt, alice, address); code != http.StatusOK {
		t.Fatalf("link: status %d", code)
	}
	if code := linkWallet(rt, mallory, address); code != http.StatusConflict {
		t.Fatalf("second account claiming the address: status %d, want 409", code)
	}

	restarted := NewRouter(nil)
	restarted.RecoveryKeys = store
	if code := linkWallet(restarted, mallory, address); code != http.StatusConflict {
		t.Fatalf("after a restart, second account claiming the address: status %d, want 409", code)
	}
	if got := restarted.lookupDIDByWallet(context.Background(), address); got != alice {
		t.Fatalf("the address now resolves to %q", got)
	}
	// Linking again for the rightful account is fine.
	if code := linkWallet(restarted, alice, address); code != http.StatusOK {
		t.Fatalf("re-link by the owner: status %d", code)
	}
}

// A caller may only link an address to its own account.
func TestRecoveryLinkIsForTheCallerOnly(t *testing.T) {
	rt := NewRouter(nil)
	rt.RecoveryKeys = wallet.NewMemStore()
	req := httptest.NewRequest(http.MethodPost, "/v1/identity/link-wallet",
		strings.NewReader(`{"did":"did:key:zAlice","wallet_address":"DAG4x"}`))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:zMallory"))
	rec := httptest.NewRecorder()
	rt.handleLinkWallet(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d, want 403", rec.Code)
	}
	if got := rt.lookupDIDByWallet(context.Background(), "DAG4x"); got != "" {
		t.Fatalf("a refused link was stored for %q", got)
	}
}
