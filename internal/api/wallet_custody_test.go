package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/wallet"
)

// In real-funds mode with no proof verifier wired, value-moving operations must
// hard-block (the signing SDK isn't available yet), not silently proceed.
func TestWalletRealFundsHardBlocksStake(t *testing.T) {
	store := wallet.NewMemStore()
	ledger := wallet.NewLedgerQuerier(store, nil)
	svc := wallet.NewWalletService(ledger, &walletTestRewards{})
	h := &WalletHandlers{Service: svc, Store: store, RealFunds: true} // Proof nil
	mux := http.NewServeMux()
	(&V3Handlers{Wallet: h}).RegisterV3Routes(mux)

	body, _ := json.Marshal(map[string]interface{}{"amount": wallet.DatumPerECHO * 100, "tier": "bronze"})
	req := httptest.NewRequest(http.MethodPost, "/v3/wallet/stake", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:zRF"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503 CUSTODY_NOT_READY, got %d: %s", rec.Code, rec.Body.String())
	}
}

// Interim mode (default) must not gate the existing TestFlight flow.
func TestWalletInterimAllowsStake(t *testing.T) {
	t.Setenv("ECHO_WALLET_GENESIS_AUTO", "1")
	store := wallet.NewMemStore()
	ledger := wallet.NewLedgerQuerier(store, nil)
	svc := wallet.NewWalletService(ledger, &walletTestRewards{})
	h := &WalletHandlers{Service: svc, Store: store, RealFunds: false}
	mux := http.NewServeMux()
	(&V3Handlers{Wallet: h}).RegisterV3Routes(mux)

	did := "did:key:zInterim"
	ctx := context.WithValue(context.Background(), ContextKeyUserID, did)
	// Seed genesis balance.
	getReq := httptest.NewRequest(http.MethodGet, "/v3/wallet", nil).WithContext(ctx)
	mux.ServeHTTP(httptest.NewRecorder(), getReq)

	body, _ := json.Marshal(map[string]interface{}{"amount": wallet.DatumPerECHO * 100, "tier": "bronze"})
	req := httptest.NewRequest(http.MethodPost, "/v3/wallet/stake", bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("interim stake should succeed, got %d: %s", rec.Code, rec.Body.String())
	}
}

// In real-funds mode, linking a server-derivable address (DAG=SHA256(did)) must
// be rejected even when a proof verifier accepts everything.
func TestWalletRealFundsRejectsServerDerivableAddress(t *testing.T) {
	store := wallet.NewMemStore()
	h := &WalletHandlers{Store: store, RealFunds: true, Proof: allowAllProof{}}
	mux := http.NewServeMux()
	(&V3Handlers{Wallet: h}).RegisterV3Routes(mux)

	did := "did:key:zDerivable"
	body, _ := json.Marshal(map[string]string{"address": wallet.ServerDerivableAddress(did)})
	req := httptest.NewRequest(http.MethodPost, "/v3/wallet/link", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, did))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 SERVER_DERIVABLE_ADDRESS, got %d: %s", rec.Code, rec.Body.String())
	}
}

// The challenge endpoint issues a per-DID nonce for proof-of-ownership.
func TestWalletChallengeEndpoint(t *testing.T) {
	h := &WalletHandlers{Store: wallet.NewMemStore(), Challenges: wallet.NewChallengeStore()}
	mux := http.NewServeMux()
	(&V3Handlers{Wallet: h}).RegisterV3Routes(mux)

	did := "did:key:zChal"
	req := httptest.NewRequest(http.MethodGet, "/v3/wallet/challenge", nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, did))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Challenge string `json:"challenge"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Challenge == "" || resp.ExpiresAt == "" {
		t.Fatalf("empty challenge/expiry: %+v", resp)
	}
}

type fakeTxContext struct {
	hash    string
	ordinal int64
}

func (f fakeTxContext) LastRef(_ context.Context, _, _ string) (string, int64, error) {
	return f.hash, f.ordinal, nil
}

// tx-context returns the linked address as source plus the last-reference parent.
func TestWalletTxContext(t *testing.T) {
	store := wallet.NewMemStore()
	_ = store.LinkDAGAccount(context.Background(), "did:key:zTC", "DAG4realaddress", "04pub")
	h := &WalletHandlers{Store: store, TxContext: fakeTxContext{hash: "parenthash", ordinal: 7}}
	mux := http.NewServeMux()
	(&V3Handlers{Wallet: h}).RegisterV3Routes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v3/wallet/tx-context?type=tokenLock", nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:zTC"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Source string `json:"source"`
		Parent struct {
			Hash    string `json:"hash"`
			Ordinal int64  `json:"ordinal"`
		} `json:"parent"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Source != "DAG4realaddress" || resp.Parent.Hash != "parenthash" || resp.Parent.Ordinal != 7 {
		t.Fatalf("unexpected tx-context: %+v", resp)
	}
}

// tx-context must 409 when no wallet is linked yet.
func TestWalletTxContext_NotLinked(t *testing.T) {
	h := &WalletHandlers{Store: wallet.NewMemStore(), TxContext: fakeTxContext{}}
	mux := http.NewServeMux()
	(&V3Handlers{Wallet: h}).RegisterV3Routes(mux)
	req := httptest.NewRequest(http.MethodGet, "/v3/wallet/tx-context", nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:zNone"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("want 409 when not linked, got %d", rec.Code)
	}
}

type allowAllProof struct{}

func (allowAllProof) VerifyOwnership(_, _, _ string) (string, error) { return "pubkey", nil }
