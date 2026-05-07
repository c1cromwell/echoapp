package api

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

func TestV3HandleAuthRegister_DerivesDidKey(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubUncompressed := elliptic.Marshal(elliptic.P256(), priv.PublicKey.X, priv.PublicKey.Y)
	pubHex := hex.EncodeToString(pubUncompressed)

	wantDID, err := didkey.Derive(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	db := database.NewMemoryDB()
	h := &V3Handlers{DB: db}

	body, _ := json.Marshal(map[string]string{
		"username":  "alice",
		"publicKey": pubHex,
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-req")
	rec := httptest.NewRecorder()

	h.handleAuthRegister(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	got, _ := resp["did"].(string)
	if got != wantDID {
		t.Fatalf("did %q, want %q", got, wantDID)
	}
	if !strings.HasPrefix(got, "did:key:") {
		t.Fatalf("expected did:key, got %q", got)
	}
}

func TestV3HandleAuthRegister_InvalidPublicKey(t *testing.T) {
	h := &V3Handlers{DB: database.NewMemoryDB()}
	body, _ := json.Marshal(map[string]string{
		"username":  "bob",
		"publicKey": "not-hex",
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-req-2")
	rec := httptest.NewRecorder()

	h.handleAuthRegister(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
