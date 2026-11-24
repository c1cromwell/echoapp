package infra

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"
)

func TestGenerateAPNsJWT(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	client := NewAPNsClient(APNsConfig{
		TeamID:     "TEAMID12",
		KeyID:      "KEYID123",
		PrivateKey: pemBytes,
		BundleID:   "com.echo.app",
	})
	token, err := client.generateJWT()
	if err != nil {
		t.Fatalf("generateJWT: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("jwt parts = %d", len(parts))
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	var header map[string]string
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatal(err)
	}
	if header["alg"] != "ES256" || header["kid"] != "KEYID123" {
		t.Fatalf("header = %+v", header)
	}
}

func TestAPNsPayloadOmitsMessageBody(t *testing.T) {
	client := NewAPNsClient(APNsConfig{BundleID: "com.echo.app"})
	// What is actually sent for a message addressed by a conversation id that names both parties.
	payload := buildAPNsPayload("dm:did:key:zAlice:did:key:zBob", "message", false)
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if strings.Contains(s, "hello") || strings.Contains(s, "ciphertext") {
		t.Fatalf("payload leaked content: %s", s)
	}
	_ = client
}

// Apple keeps what it relays. A push must not tell it who is talking to whom.
func TestAPNsPayloadCarriesNoConversationOrParty(t *testing.T) {
	for _, silent := range []bool{false, true} {
		raw, err := json.Marshal(buildAPNsPayload("dm:did:key:zAlice:did:key:zBob", "message", silent))
		if err != nil {
			t.Fatal(err)
		}
		for _, leak := range []string{"zAlice", "zBob", "dm:", "conversation"} {
			if strings.Contains(string(raw), leak) {
				t.Fatalf("silent=%v: push payload contains %q: %s", silent, leak, raw)
			}
		}
	}
	// A missed call carries its own random id, and only that.
	raw, _ := json.Marshal(buildAPNsPayload("call-7f3a", "missed_call", false))
	if !strings.Contains(string(raw), `"callId":"call-7f3a"`) {
		t.Fatalf("missed call payload lost its call id: %s", raw)
	}
}
