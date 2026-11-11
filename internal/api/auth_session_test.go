package api

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

type testDevice struct {
	priv *ecdsa.PrivateKey
	did  string
}

// enrollTestDevice registers a fresh P-256 device key the way a new install does.
func enrollTestDevice(t *testing.T, rt *Router) testDevice {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubHex, err := didkey.PublicKeyHexUncompressed(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	did, err := didkey.DeriveFromPublicKeyHex(pubHex)
	if err != nil {
		t.Fatal(err)
	}
	rec := secPostJSON(t, rt, "/identity/register", map[string]string{"did": did, "public_key_hex": pubHex}, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("identity/register: %d %s", rec.Code, rec.Body.String())
	}
	return testDevice{priv: priv, did: did}
}

// signed builds a request authenticated with X-Sender-DID / X-Signature / X-Timestamp.
func (d testDevice) signed(t *testing.T, method, path string, body []byte) *http.Request {
	t.Helper()
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig, err := didkey.SignECDSAP256SHA256ASN1(d.priv, canonicalSigningString(method, path, ts, body))
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerSenderDID, d.did)
	req.Header.Set(headerSignature, base64.StdEncoding.EncodeToString(sig))
	req.Header.Set(headerTimestamp, ts)
	return req
}

// A newly enrolled device can turn its registered key into a session, and that
// session's access token is what the WebSocket upgrade accepts.
func TestAuthSession_EnrolledDeviceGetsUsableTokens(t *testing.T) {
	rt := NewRouter([]string{"*"})
	dev := enrollTestDevice(t, rt)

	body, _ := json.Marshal(map[string]string{"device_id": "install-1"})
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, dev.signed(t, http.MethodPost, "/v1/auth/session", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("session: %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		DID          string `json:"did"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.DID != dev.did || out.AccessToken == "" || out.RefreshToken == "" || out.ExpiresIn <= 0 {
		t.Fatalf("unexpected session response: %+v", out)
	}
	if got := rt.UserIDExtractor(out.AccessToken); got != dev.did {
		t.Fatalf("access token subject = %q, want the device's DID", got)
	}

	// The refresh token rotates into a new pair without another signature.
	refreshRec := secPostJSON(t, rt, "/v3/auth/refresh", map[string]string{
		"refresh_token": out.RefreshToken,
		"device_id":     "install-1",
	}, "")
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh: %d %s", refreshRec.Code, refreshRec.Body.String())
	}
}

func TestAuthSession_RequiresPasskeySignature(t *testing.T) {
	rt := NewRouter([]string{"*"})
	dev := enrollTestDevice(t, rt)
	body := map[string]string{"device_id": "install-1"}

	// No credentials at all.
	if rec := secPostJSON(t, rt, "/v1/auth/session", body, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: want 401, got %d %s", rec.Code, rec.Body.String())
	}
	// A valid access token must not be able to mint further sessions.
	bearer := accessTokenFor(t, rt, dev.did)
	if rec := secPostJSON(t, rt, "/v1/auth/session", body, bearer); rec.Code != http.StatusUnauthorized {
		t.Fatalf("bearer: want 401, got %d %s", rec.Code, rec.Body.String())
	}
	// A signature from a key that is not registered for the DID.
	stranger, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	raw, _ := json.Marshal(body)
	forged := testDevice{priv: stranger, did: dev.did}.signed(t, http.MethodPost, "/v1/auth/session", raw)
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, forged)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("forged signature: want 401, got %d %s", rec.Code, rec.Body.String())
	}
}
