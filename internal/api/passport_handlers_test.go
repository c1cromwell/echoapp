package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/pkg/passport"
	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

const testHolderDID = "did:key:z6MkTestHolderForPassportApiTests"

func testPassportRouter(t *testing.T) *Router {
	t.Helper()
	rt := NewRouter(nil)
	store := &passportMemStore{}
	rt.Passport = passport.NewService(store, nil)
	rt.TokenValidator = func(string) bool { return true }
	rt.UserIDExtractor = func(string) string { return testHolderDID }
	return rt
}

func testPassportSyncRouter(t *testing.T) *Router {
	t.Helper()
	rt := testPassportRouter(t)
	rt.PassportSync = passport.NewSyncService(passport.NewMemSyncStore(), encblob.NewStubStorage())
	return rt
}

func TestPassportCredentialsListRequiresAuth(t *testing.T) {
	rt := NewRouter(nil)
	rt.Passport = passport.NewService(&passportMemStore{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/passport/credentials", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPassportCredentialsRegisterAndList(t *testing.T) {
	rt := testPassportRouter(t)

	body, _ := json.Marshal(map[string]interface{}{
		"issuer_did":      "did:key:zIssuer",
		"credential_type": "ProofOfHumanity",
		"credential_hash": "aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/passport/credentials", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register expected 201, got %d: %s", w.Code, w.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/passport/credentials", nil)
	req2.Header.Set("Authorization", "Bearer test-token")
	w2 := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestPassportSyncPushAndPull(t *testing.T) {
	rt := testPassportSyncRouter(t)
	body, _ := json.Marshal(map[string]string{
		"ciphertext_base64": passport.EncodeBase64([]byte("encrypted-wallet")),
	})
	post := httptest.NewRequest(http.MethodPost, "/v1/passport/sync", bytes.NewReader(body))
	post.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, post)
	if w.Code != http.StatusOK {
		t.Fatalf("push expected 200, got %d: %s", w.Code, w.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/v1/passport/sync", nil)
	get.Header.Set("Authorization", "Bearer test-token")
	w2 := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w2, get)
	if w2.Code != http.StatusOK {
		t.Fatalf("pull expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
}

// passportMemStore is a minimal in-memory RefStore for handler tests.
type passportMemStore struct {
	refs []passport.CredentialRef
}

func (m *passportMemStore) ListCredentialRefs(_ context.Context, holderDID string) ([]passport.CredentialRef, error) {
	var out []passport.CredentialRef
	for _, r := range m.refs {
		if r.HolderDID == holderDID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *passportMemStore) GetCredentialRef(_ context.Context, holderDID, refID string) (*passport.CredentialRef, error) {
	for _, r := range m.refs {
		if r.HolderDID == holderDID && r.RefID == refID {
			cp := r
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *passportMemStore) InsertCredentialRef(_ context.Context, ref passport.CredentialRef) error {
	m.refs = append(m.refs, ref)
	return nil
}
