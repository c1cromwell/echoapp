package api

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// asCallerOfKey signs the request in as the identity whose public key is in its body, which is
// what the auth middleware does for a real caller. A body with no usable key is left anonymous.
func asCallerOfKey(t *testing.T, req *http.Request) *http.Request {
	t.Helper()
	raw, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewReader(raw))
	var body struct {
		PublicKey string `json:"publicKey"`
	}
	_ = json.Unmarshal(raw, &body)
	did, err := didkey.DeriveFromPublicKeyHex(body.PublicKey)
	if err != nil {
		return asCaller(req, "did:key:zSomeoneSignedIn")
	}
	return asCaller(req, did)
}

func newIdentity(t *testing.T) (did, pubHex string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubHex = hex.EncodeToString(elliptic.Marshal(elliptic.P256(), priv.X, priv.Y))
	did, err = didkey.Derive(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return did, pubHex
}

func registerUsername(h *V3Handlers, caller, username, pubHex string) (int, map[string]interface{}) {
	body, _ := json.Marshal(map[string]string{"username": username, "publicKey": pubHex})
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/register", bytes.NewReader(body))
	if caller != "" {
		req = asCaller(req, caller)
	}
	rec := httptest.NewRecorder()
	h.handleAuthRegister(rec, req)
	var out map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestUsernameRegistration(t *testing.T) {
	h := &V3Handlers{DB: database.NewMemoryDB()}
	alice, alicePub := newIdentity(t)
	bob, bobPub := newIdentity(t)

	if code, out := registerUsername(h, alice, "Alice_1", alicePub); code != http.StatusCreated || out["did"] != alice || out["username"] != "Alice_1" {
		t.Fatalf("register: %d %v", code, out)
	}
	// A retry after a dropped reply must succeed, not fail as "taken".
	if code, _ := registerUsername(h, alice, "Alice_1", alicePub); code != http.StatusOK {
		t.Errorf("repeating one's own registration: %d, want 200", code)
	}
	if code, _ := registerUsername(h, alice, "@alice_1", alicePub); code != http.StatusOK {
		t.Errorf("repeating with different case and an @: %d, want 200", code)
	}
	// One identity, one username.
	if code, out := registerUsername(h, alice, "second_name", alicePub); code != http.StatusConflict || out["code"] != "ALREADY_REGISTERED" {
		t.Errorf("a second username for the same account: %d %v", code, out["code"])
	}
	// Taken is taken, whatever the capitals.
	for _, name := range []string{"Alice_1", "alice_1", "ALICE_1"} {
		if code, out := registerUsername(h, bob, name, bobPub); code != http.StatusConflict || out["code"] != "USERNAME_TAKEN" {
			t.Errorf("another account taking %q: %d %v", name, code, out["code"])
		}
	}
	if code, _ := registerUsername(h, bob, "bob", bobPub); code != http.StatusCreated {
		t.Errorf("bob: %d", code)
	}
	for name, u := range map[string]string{"Alice_1": alice, "alice_1": alice, "BOB": bob} {
		got, err := h.DB.GetUserByUsername(t.Context(), name)
		if err != nil || got.DID != u {
			t.Errorf("lookup %q: %v %v", name, got, err)
		}
	}
}

// A public key is not a secret. Without tying it to the signed-in caller, anyone could attach a
// username of their choosing to someone else's identity.
func TestUsernameRegistrationIsForTheCallersOwnIdentity(t *testing.T) {
	h := &V3Handlers{DB: database.NewMemoryDB()}
	victim, victimPub := newIdentity(t)
	mallory, _ := newIdentity(t)

	if code, out := registerUsername(h, mallory, "embarrassing_name", victimPub); code != http.StatusForbidden || out["code"] != "DID_MISMATCH" {
		t.Fatalf("registering a name for someone else's key: %d %v", code, out["code"])
	}
	if code, _ := registerUsername(h, "", "anyone", victimPub); code != http.StatusUnauthorized {
		t.Fatalf("registering with no sign-in: %d, want 401", code)
	}
	if u, err := h.DB.GetUserByDID(t.Context(), victim); err == nil && u != nil {
		t.Fatalf("a refused registration created a record: %+v", u)
	}
}

func TestUsernameRegistrationRejectsBadNames(t *testing.T) {
	h := &V3Handlers{DB: database.NewMemoryDB()}
	did, pub := newIdentity(t)
	for _, name := range []string{"ab", "has space", "semi;colon", "ünïcode", "this_name_is_far_too_long_to_be_allowed_x"} {
		if code, out := registerUsername(h, did, name, pub); code != http.StatusBadRequest || out["code"] != "INVALID_USERNAME" {
			t.Errorf("%q: %d %v", name, code, out["code"])
		}
	}
}
