package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/ledger"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// TestV3AuthRegister_AnchorsUsername verifies D1: registering a user anchors the
// public @username -> DID binding on the Identity Metagraph (Postgres is just a
// cache), and the on-chain payload carries the username and the derived DID.
func TestV3AuthRegister_AnchorsUsername(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubHex := hex.EncodeToString(elliptic.Marshal(elliptic.P256(), priv.PublicKey.X, priv.PublicKey.Y))
	wantDID, err := didkey.Derive(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	var anchored ledger.UsernameRegistrationUpdate
	h := &V3Handlers{
		DB: database.NewMemoryDB(),
		IdentityL1: ledger.IdentitySubmitterFunc(func(_ context.Context, tx interface{}) (string, error) {
			anchored = tx.(ledger.UsernameRegistrationUpdate)
			return "0xtx", nil
		}),
	}

	body, _ := json.Marshal(map[string]string{"username": "alice", "publicKey": pubHex})
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.handleAuthRegister(rec, asCallerOfKey(t, req))

	if rec.Code != http.StatusCreated {
		t.Fatalf("register want 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// anchorUsername runs synchronously before the response, so anchored is set.
	if anchored.Username != "alice" {
		t.Fatalf("anchored username = %q, want alice", anchored.Username)
	}
	if anchored.SubjectDID != wantDID {
		t.Fatalf("anchored subjectDID = %q, want %q", anchored.SubjectDID, wantDID)
	}
	if anchored.RegisteredAt <= 0 {
		t.Fatalf("anchored registeredAt should be positive, got %d", anchored.RegisteredAt)
	}
}

// TestV3AuthRegister_NoMetagraph confirms registration still succeeds when no
// Identity L1 client is configured (anchoring is best-effort).
func TestV3AuthRegister_NoMetagraph(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubHex := hex.EncodeToString(elliptic.Marshal(elliptic.P256(), priv.PublicKey.X, priv.PublicKey.Y))

	h := &V3Handlers{DB: database.NewMemoryDB()} // no IdentityL1
	body, _ := json.Marshal(map[string]string{"username": "bob", "publicKey": pubHex})
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.handleAuthRegister(rec, asCallerOfKey(t, req))

	if rec.Code != http.StatusCreated {
		t.Fatalf("register want 201 without metagraph, got %d", rec.Code)
	}
}
