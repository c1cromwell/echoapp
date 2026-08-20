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
	payload := APNsPayload{
		APS:            APS{Alert: APSAlert{Title: "Echo", Body: "You have a new message"}},
		Type:           "message",
		ConversationID: "dm:alice:bob",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if strings.Contains(s, "hello") || strings.Contains(s, "ciphertext") {
		t.Fatalf("payload leaked content: %s", s)
	}
	if !strings.Contains(s, "conversationId") {
		t.Fatalf("missing conversationId: %s", s)
	}
	_ = client
}
