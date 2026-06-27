package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/wallet"
)

func TestWalletGetState_GenesisCredit(t *testing.T) {
	t.Setenv("ECHO_WALLET_GENESIS_AUTO", "1")

	store := wallet.NewMemStore()
	ledger := wallet.NewLedgerQuerier(store, nil)
	svc := wallet.NewWalletService(ledger, &walletTestRewards{})

	h := &WalletHandlers{Service: svc, Store: store}
	mux := http.NewServeMux()
	(&V3Handlers{Wallet: h}).RegisterV3Routes(mux)

	did := "did:key:zWalletTest"
	req := httptest.NewRequest(http.MethodGet, "/v3/wallet", nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, did))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var state wallet.WalletState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if state.DID != did {
		t.Fatalf("did = %q", state.DID)
	}
	if state.TotalBalance <= 0 {
		t.Fatalf("expected genesis credit, total=%d", state.TotalBalance)
	}
}

func TestWalletStakeAndUnstake(t *testing.T) {
	t.Setenv("ECHO_WALLET_GENESIS_AUTO", "1")
	t.Setenv("ECHO_WALLET_GENESIS_ECHO", "1000")

	store := wallet.NewMemStore()
	ledger := wallet.NewLedgerQuerier(store, nil)
	svc := wallet.NewWalletService(ledger, &walletTestRewards{})
	h := &WalletHandlers{Service: svc, Store: store}
	mux := http.NewServeMux()
	(&V3Handlers{Wallet: h}).RegisterV3Routes(mux)

	did := "did:key:zStakeTest"
	ctx := context.WithValue(context.Background(), ContextKeyUserID, did)

	// Seed balance via GET wallet (genesis).
	getReq := httptest.NewRequest(http.MethodGet, "/v3/wallet", nil)
	getReq = getReq.WithContext(ctx)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get wallet: %d %s", getRec.Code, getRec.Body.String())
	}

	stakeBody, _ := json.Marshal(map[string]interface{}{
		"amount": wallet.DatumPerECHO * 100,
		"tier":   "bronze",
	})
	stakeReq := httptest.NewRequest(http.MethodPost, "/v3/wallet/stake", bytes.NewReader(stakeBody))
	stakeReq = stakeReq.WithContext(ctx)
	stakeReq.Header.Set("Content-Type", "application/json")
	stakeRec := httptest.NewRecorder()
	mux.ServeHTTP(stakeRec, stakeReq)
	if stakeRec.Code != http.StatusOK {
		t.Fatalf("stake: %d %s", stakeRec.Code, stakeRec.Body.String())
	}

	var stakeResult wallet.StakeResult
	_ = json.Unmarshal(stakeRec.Body.Bytes(), &stakeResult)
	if stakeResult.TxHash == "" {
		t.Fatal("expected tx hash")
	}

	getRec2 := httptest.NewRecorder()
	mux.ServeHTTP(getRec2, getReq)
	var after wallet.WalletState
	_ = json.Unmarshal(getRec2.Body.Bytes(), &after)
	if len(after.Locks) != 1 {
		t.Fatalf("expected 1 lock, got %d", len(after.Locks))
	}

	unstakeBody, _ := json.Marshal(map[string]interface{}{
		"stakeId": after.Locks[0].ID,
		"amount":  wallet.DatumPerECHO * 100,
	})
	unstakeReq := httptest.NewRequest(http.MethodPost, "/v3/wallet/unstake", bytes.NewReader(unstakeBody))
	unstakeReq = unstakeReq.WithContext(ctx)
	unstakeReq.Header.Set("Content-Type", "application/json")
	unstakeRec := httptest.NewRecorder()
	mux.ServeHTTP(unstakeRec, unstakeReq)
	if unstakeRec.Code != http.StatusOK {
		t.Fatalf("unstake: %d %s", unstakeRec.Code, unstakeRec.Body.String())
	}
}

func TestWalletLink(t *testing.T) {
	store := wallet.NewMemStore()
	h := &WalletHandlers{Store: store}
	mux := http.NewServeMux()
	(&V3Handlers{Wallet: h}).RegisterV3Routes(mux)

	did := "did:key:zLink"
	body, _ := json.Marshal(map[string]string{"address": "DAGabc123"})
	req := httptest.NewRequest(http.MethodPost, "/v3/wallet/link", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, did))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("link: %d %s", rec.Code, rec.Body.String())
	}

	addr, err := store.GetDAGAddress(context.Background(), did)
	if err != nil {
		t.Fatal(err)
	}
	if addr != "DAGabc123" {
		t.Fatalf("address = %q", addr)
	}
}

func TestWalletUnavailableWithoutPG(t *testing.T) {
	_ = os.Unsetenv("ECHO_WALLET_GENESIS_AUTO")
	rt := &Router{V3: &V3Handlers{DB: database.NewMemoryDB()}}
	// Wallet nil — route not registered; use handler directly.
	h := &WalletHandlers{}
	req := httptest.NewRequest(http.MethodGet, "/v3/wallet", nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:x"))
	rec := httptest.NewRecorder()
	h.handleWalletRoot(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	_ = rt
}

type walletTestRewards struct{}

func (walletTestRewards) GetPending(context.Context, string) (int64, error) { return 0, nil }
func (walletTestRewards) GetPendingByType(context.Context, string, string) (int64, error) {
	return 0, nil
}
func (walletTestRewards) GetAutoScaleState(context.Context, string) (*wallet.AutoScaleState, error) {
	return &wallet.AutoScaleState{CurrentRate: 1}, nil
}
func (walletTestRewards) ClearPending(context.Context, string, []string) error { return nil }
