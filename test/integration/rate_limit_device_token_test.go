package integration

// Wave 10 integration tests:
//   - Rate limiting: 429 + Retry-After header after limit exhaustion
//   - Device token flow: POST /identity/devices/token → POST /identity/devices (token path)
//   - Server key exchange: GET /v1/crypto/server-key returns valid X25519 key

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/infra"
	"github.com/thechadcromwell/echoapp/internal/testutil"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// Note: mustGenerateKeyPair / mustDeriveDID / mustRegisterDID are still used by
// the device token tests which need a real registered DID (token-based add-device
// requires the primary DID to be in the registry).

// --- Rate Limiting ---

// TestRateLimit_429AfterExhaustion verifies that once the per-DID limit is
// exhausted the server returns 429 with a Retry-After header.
func TestRateLimit_429AfterExhaustion(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	// Wire a tight rate limiter (2 requests/min) so the test finishes quickly.
	ts.Router.RateLimiter = infra.NewRateLimiter(map[string]infra.RateLimitConfig{
		"api_request": {MaxRequests: 2, Window: 60_000_000_000}, // 1 minute
	})

	// JWT auth — no DID registration needed; JWT subject IS the DID for rate limiting.
	const did = "did:key:zTestRateLimitExhaustion"

	// Use a protected path (requires auth → DID injected into context → rate limiter fires).
	// /v3/messages returns 503 when V3 handlers aren't wired, but the rate limiter
	// still applies before the handler is reached.
	const protectedPath = "/v3/messages"

	// Exhaust the 2-request limit.
	for i := 0; i < 2; i++ {
		code, _ := authenticatedGET(t, ts, did, nil, protectedPath)
		if code == http.StatusTooManyRequests {
			t.Fatalf("request %d was rate-limited too early", i)
		}
	}

	// Third request must be rejected with 429.
	code, resp := authenticatedGET(t, ts, did, nil, protectedPath)
	if code != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d", code)
	}
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		t.Error("missing Retry-After header on 429 response")
	}
	secs, err := strconv.Atoi(retryAfter)
	if err != nil || secs <= 0 {
		t.Errorf("Retry-After must be a positive integer, got %q", retryAfter)
	}
}

// TestRateLimit_DifferentDIDsIndependent confirms rate limiting is per-DID.
// JWT auth is used so no DID registration is required — the JWT subject IS the DID.
func TestRateLimit_DifferentDIDsIndependent(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	ts.Router.RateLimiter = infra.NewRateLimiter(map[string]infra.RateLimitConfig{
		"api_request": {MaxRequests: 1, Window: 60_000_000_000},
	})

	const did1 = "did:key:zTestUser1RateLimitIsolation"
	const did2 = "did:key:zTestUser2RateLimitIsolation"

	// Exhaust DID1's 1-request quota.
	authenticatedGET(t, ts, did1, nil, "/v3/messages")

	// DID2 has a fresh quota — must not be affected by DID1's exhaustion.
	code2, _ := authenticatedGET(t, ts, did2, nil, "/v3/messages")
	if code2 == http.StatusTooManyRequests {
		t.Error("DID2 should not be affected by DID1 rate limit exhaustion")
	}
}

// --- Device token flow ---

// TestDeviceTokenFlow_EndToEnd exercises the full QR-code add-device path:
//  1. Primary device registers.
//  2. Primary device obtains a 5-min registration token (POST /identity/devices/token).
//  3. Secondary device registers using the token (POST /identity/devices).
//  4. GET /identity/devices/{did} lists both devices.
func TestDeviceTokenFlow_EndToEnd(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	if ts.Router.Redis == nil {
		t.Skip("Redis not available — device token flow requires Redis")
	}

	// Register primary device.
	_, primaryPub := mustGenerateKeyPair(t)
	primaryDID := mustDeriveDID(t, primaryPub)
	mustRegisterDID(t, ts, primaryDID, hex.EncodeToString(primaryPub))

	// Obtain registration token (requires auth — use JWT for test simplicity).
	token := mustGetDeviceToken(t, ts, primaryDID)
	if len(token) != 64 { // 32 bytes hex
		t.Errorf("expected 64-char hex token, got len=%d", len(token))
	}

	// Register secondary device using the token.
	_, secondaryPub := mustGenerateKeyPair(t)
	secondaryDID := mustDeriveDID(t, secondaryPub)
	addBody := map[string]string{
		"token":            token,
		"new_public_key_hex": hex.EncodeToString(secondaryPub),
		"device_label":     "secondary-phone",
	}
	raw, _ := json.Marshal(addBody)
	resp, err := http.Post(ts.BaseURL+"/identity/devices",
		"application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("POST /identity/devices failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("want 201, got %d", resp.StatusCode)
	}

	// Verify both devices are listed.
	listResp, err := http.Get(ts.BaseURL + "/identity/devices/" + primaryDID)
	if err != nil {
		t.Fatalf("GET /identity/devices/%s failed: %v", primaryDID, err)
	}
	defer listResp.Body.Close()
	if listResp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", listResp.StatusCode)
	}

	var devices []map[string]string
	if err := json.NewDecoder(listResp.Body).Decode(&devices); err != nil {
		t.Fatalf("decode devices: %v", err)
	}
	if len(devices) < 2 {
		t.Errorf("expected at least 2 devices, got %d", len(devices))
	}

	// Confirm secondary DID appears in the list.
	found := false
	for _, d := range devices {
		if d["did"] == secondaryDID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("secondary DID %q not found in device list", secondaryDID)
	}
}

// TestDeviceTokenFlow_ExpiredTokenRejected confirms a replayed / wrong token
// returns 401.
func TestDeviceTokenFlow_InvalidTokenRejected(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	if ts.Router.Redis == nil {
		t.Skip("Redis not available — device token flow requires Redis")
	}

	_, pub := mustGenerateKeyPair(t)
	addBody := map[string]string{
		"token":              "0000000000000000000000000000000000000000000000000000000000000000",
		"new_public_key_hex": hex.EncodeToString(pub),
	}
	raw, _ := json.Marshal(addBody)
	resp, err := http.Post(ts.BaseURL+"/identity/devices",
		"application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("POST /identity/devices failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("invalid token: want 401, got %d", resp.StatusCode)
	}
}

// --- Server key exchange ---

// TestServerKey_ValidX25519Key verifies GET /v1/crypto/server-key returns a
// 32-byte X25519 public key and correct algorithm string.
func TestServerKey_ValidX25519Key(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	resp, err := http.Get(ts.BaseURL + "/v1/crypto/server-key")
	if err != nil {
		t.Fatalf("GET /v1/crypto/server-key: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	var body struct {
		PublicKey string `json:"public_key"`
		Algorithm string `json:"algorithm"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Algorithm != "X25519-ChaCha20Poly1305" {
		t.Errorf("algorithm: want X25519-ChaCha20Poly1305, got %q", body.Algorithm)
	}
	if body.PublicKey == "" {
		t.Error("public_key must not be empty")
	}
}

// --- helpers ---

// testVector holds the fixed P-256 key pair used in golden-vector tests.
// Source: internal/api/passkey_auth_test.go test vectors.
var testVector = struct {
	priv string
	pub  string
}{
	priv: "eaff1084e1322774ce79ad393aa9d925d73473a9de5fe6a50d969672ac66be4f",
	pub:  "047bfc587ef5617b74f66c8c26765adfc6ac311be92ec5546f146b28a026e96edd56e82054fa9de5114d4e59938f29dbe2f63bffa40a321f0f472f28ea30f8d1f4",
}

func mustGenerateKeyPair(t *testing.T) (privBytes []byte, pubBytes []byte) {
	t.Helper()
	priv, _ := hex.DecodeString(testVector.priv)
	pub, _ := hex.DecodeString(testVector.pub)
	return priv, pub
}

func mustDeriveDID(t *testing.T, pubUncompressed []byte) string {
	t.Helper()
	did, err := didkey.DeriveFromPublicKeyHex(hex.EncodeToString(pubUncompressed))
	if err != nil {
		t.Fatalf("DeriveFromPublicKeyHex: %v", err)
	}
	return did
}

func mustRegisterDID(t *testing.T, ts *testutil.TestServer, did, pubKeyHex string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"did":            did,
		"public_key_hex": pubKeyHex,
	})
	resp, err := http.Post(ts.BaseURL+"/identity/register",
		"application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("register DID: %v", err)
	}
	resp.Body.Close()
}

func mustGetDeviceToken(t *testing.T, ts *testutil.TestServer, did string) string {
	t.Helper()
	// POST /identity/devices/token requires auth; use the test JWT helper.
	token := ts.IssueTestToken(did)
	req, _ := http.NewRequest(http.MethodPost, ts.BaseURL+"/identity/devices/token", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /identity/devices/token: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("device token: want 200, got %d", resp.StatusCode)
	}
	var body struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode device token: %v", err)
	}
	return body.Token
}

// authenticatedGET issues a GET authenticated via JWT (simplest for rate-limit tests).
func authenticatedGET(t *testing.T, ts *testutil.TestServer, did string, _ []byte, path string) (int, *http.Response) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, ts.BaseURL+path, nil)
	token := ts.IssueTestToken(did)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp.StatusCode, resp
}
