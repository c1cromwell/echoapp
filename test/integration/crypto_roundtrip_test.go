package integration

// TestCryptoRoundTrip_X25519ChaCha20 exercises the full WO-13 key-exchange flow:
//  1. Fetch the backend's static X25519 public key (GET /v1/crypto/server-key).
//  2. Encrypt a payload locally using X25519ChaChaSvc.EncryptWithKeyAgreement.
//  3. Decrypt with the server's private key (accessible via globalServerKeys in tests).
//  4. Confirm plaintext is preserved.
//
// This is the integration-level proof that the iOS → backend encryption
// chain defined in WO-13 works end-to-end.

import (
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/api"
	"github.com/thechadcromwell/echoapp/internal/crypto"
	"github.com/thechadcromwell/echoapp/internal/testutil"
)

// TestCryptoRoundTrip_X25519ChaCha20 confirms the full iOS→backend encryption
// round-trip using the server's published X25519 key.
func TestCryptoRoundTrip_X25519ChaCha20(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	// --- Step 1: fetch server public key ---
	resp, err := http.Get(ts.BaseURL + "/v1/crypto/server-key")
	if err != nil {
		t.Fatalf("GET /v1/crypto/server-key: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	var keyResp api.ServerKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&keyResp); err != nil {
		t.Fatalf("decode server key response: %v", err)
	}
	if keyResp.Algorithm != "X25519-ChaCha20Poly1305" {
		t.Fatalf("unexpected algorithm: %q", keyResp.Algorithm)
	}
	serverPubBytes, err := base64.StdEncoding.DecodeString(keyResp.PublicKey)
	if err != nil {
		t.Fatalf("decode server public key: %v", err)
	}

	// --- Step 2: encrypt with the server's public key (iOS side) ---
	svc := crypto.NewX25519ChaChaSvc()
	plaintext := []byte("WO-13 X25519+ChaCha20Poly1305 round-trip test payload")
	msg, err := svc.EncryptWithKeyAgreement(plaintext, serverPubBytes)
	if err != nil {
		t.Fatalf("EncryptWithKeyAgreement: %v", err)
	}
	if msg.Algorithm != crypto.X25519ChaChaAlgorithm {
		t.Errorf("algorithm: want %q, got %q", crypto.X25519ChaChaAlgorithm, msg.Algorithm)
	}

	// --- Step 3: decrypt with server's static private key ---
	// globalServerKeys is the process-level singleton used by handleServerKey.
	serverPriv := api.GlobalServerKeys().PrivateKey()
	if serverPriv == nil {
		t.Fatal("server private key is nil — key was not initialised")
	}
	got, err := svc.DecryptWithKeyAgreement(msg, serverPriv)
	if err != nil {
		t.Fatalf("DecryptWithKeyAgreement: %v", err)
	}

	// --- Step 4: confirm plaintext matches ---
	if string(got) != string(plaintext) {
		t.Errorf("plaintext mismatch:\n  want: %q\n  got:  %q", plaintext, got)
	}
}

// TestCryptoRoundTrip_TamperedCiphertext confirms that a tampered message
// is rejected during decryption.
func TestCryptoRoundTrip_TamperedCiphertext(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	resp, err := http.Get(ts.BaseURL + "/v1/crypto/server-key")
	if err != nil {
		t.Fatalf("GET /v1/crypto/server-key: %v", err)
	}
	defer resp.Body.Close()

	var keyResp api.ServerKeyResponse
	json.NewDecoder(resp.Body).Decode(&keyResp)
	serverPubBytes, _ := base64.StdEncoding.DecodeString(keyResp.PublicKey)

	svc := crypto.NewX25519ChaChaSvc()
	msg, _ := svc.EncryptWithKeyAgreement([]byte("tamper test"), serverPubBytes)

	// Flip a byte in the ciphertext.
	ct, _ := base64.StdEncoding.DecodeString(msg.Ciphertext)
	ct[0] ^= 0xFF
	msg.Ciphertext = base64.StdEncoding.EncodeToString(ct)

	serverPriv := api.GlobalServerKeys().PrivateKey()
	if _, err := svc.DecryptWithKeyAgreement(msg, serverPriv); err == nil {
		t.Error("tampered ciphertext should fail decryption")
	}
}

// TestCryptoRoundTrip_DifferentKeyRejected verifies that a message encrypted
// for a different X25519 key cannot be decrypted with the server key.
func TestCryptoRoundTrip_DifferentKeyRejected(t *testing.T) {
	svc := crypto.NewX25519ChaChaSvc()
	_, wrongPub, _ := svc.GenerateKeyPair()

	msg, _ := svc.EncryptWithKeyAgreement([]byte("wrong key test"), wrongPub)

	serverPriv := api.GlobalServerKeys().PrivateKey()
	if _, err := svc.DecryptWithKeyAgreement(msg, serverPriv); err == nil {
		t.Error("message encrypted for wrong key should fail decryption with server key")
	}
}

// GlobalServerKeys exposes the process-level singleton for test use.
// It is the same singleton that handleServerKey serves — no keys are generated twice.
func init() {
	// Warm up the singleton so TestCryptoRoundTrip_DifferentKeyRejected has a
	// non-nil private key even without calling the HTTP endpoint.
	_, _ = api.GlobalServerKeys().PublicKeyBase64()
}

// Ensure ecdh import is used by the file (avoids "imported and not used" errors
// from the *ecdh.PrivateKey return type in GlobalServerKeys).
var _ *ecdh.PrivateKey = nil
