package credentials

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// buildJWT builds a minimal compact JWT with the given payload for testing.
func buildJWT(t *testing.T, payload map[string]interface{}) string {
	t.Helper()
	header := map[string]interface{}{"alg": "EdDSA", "typ": "JWT"}
	hb, _ := json.Marshal(header)
	pb, _ := json.Marshal(payload)
	h := base64.RawURLEncoding.EncodeToString(hb)
	p := base64.RawURLEncoding.EncodeToString(pb)
	// Signature not validated in FromJWT — use a placeholder.
	return h + "." + p + ".fakesig"
}

func TestFromJWT_ValidVCClaim(t *testing.T) {
	fh := NewFormatHandler(NewCryptoUtils())
	now := time.Now().UTC().Truncate(time.Second)
	vc := &VerifiableCredential{
		ID:     "https://example.com/creds/1",
		Issuer: "did:key:zTestIssuer",
		CredentialSubject: CredentialSubject{
			ID: "did:key:zTestSubject",
		},
		IssuanceDate: now,
	}
	payload := map[string]interface{}{
		"vc":  vc,
		"iss": "did:key:zTestIssuer",
		"sub": "did:key:zTestSubject",
		"jti": "https://example.com/creds/1",
	}
	jwtStr := buildJWT(t, payload)

	got, err := fh.FromJWT(jwtStr)
	if err != nil {
		t.Fatalf("FromJWT: %v", err)
	}
	if got.ID != "https://example.com/creds/1" {
		t.Errorf("ID: want %q, got %q", "https://example.com/creds/1", got.ID)
	}
	if got.Issuer != "did:key:zTestIssuer" {
		t.Errorf("Issuer: want %q, got %q", "did:key:zTestIssuer", got.Issuer)
	}
}

func TestFromJWT_BackfillsIssuerFromClaim(t *testing.T) {
	fh := NewFormatHandler(NewCryptoUtils())
	// VC itself has no issuer; the outer `iss` JWT claim should be backfilled.
	vc := map[string]interface{}{
		"@context":          []string{"https://www.w3.org/2018/credentials/v1"},
		"type":              []string{"VerifiableCredential"},
		"credentialSubject": map[string]string{"id": "did:key:zSub"},
	}
	payload := map[string]interface{}{
		"vc":  vc,
		"iss": "did:key:zIssuer",
		"jti": "https://example.com/creds/2",
	}
	jwtStr := buildJWT(t, payload)

	got, err := fh.FromJWT(jwtStr)
	if err != nil {
		t.Fatalf("FromJWT: %v", err)
	}
	if got.Issuer != "did:key:zIssuer" {
		t.Errorf("expected issuer backfill, got %q", got.Issuer)
	}
}

func TestFromJWT_MalformedJWT(t *testing.T) {
	fh := NewFormatHandler(NewCryptoUtils())
	if _, err := fh.FromJWT("notajwt"); err == nil {
		t.Error("expected error for non-JWT string")
	}
	if _, err := fh.FromJWT("a.b"); err == nil {
		t.Error("expected error for two-part JWT")
	}
}

func TestFromJWT_MissingVCClaim(t *testing.T) {
	fh := NewFormatHandler(NewCryptoUtils())
	payload := map[string]interface{}{"iss": "did:key:z1", "sub": "did:key:z2"}
	jwtStr := buildJWT(t, payload)
	if _, err := fh.FromJWT(jwtStr); err == nil {
		t.Error("expected error when 'vc' claim is absent")
	}
}

func TestFromJWT_InvalidBase64Payload(t *testing.T) {
	fh := NewFormatHandler(NewCryptoUtils())
	if _, err := fh.FromJWT("header.!!!.sig"); err == nil {
		t.Error("expected error for invalid base64url payload")
	}
}

func TestFromSDJWT_StripsDisclosures(t *testing.T) {
	fh := NewFormatHandler(NewCryptoUtils())
	vc := &VerifiableCredential{
		ID:     "https://example.com/creds/sd",
		Issuer: "did:key:zSDIssuer",
		CredentialSubject: CredentialSubject{
			ID: "did:key:zSDSubject",
		},
		IssuanceDate: time.Now().UTC(),
	}
	payload := map[string]interface{}{"vc": vc, "iss": "did:key:zSDIssuer"}
	jwtStr := buildJWT(t, payload)
	// Append SD-JWT disclosures (separated by ~)
	sdJWT := jwtStr + "~disclosure1~disclosure2~"

	got, err := fh.FromSDJWT(sdJWT)
	if err != nil {
		t.Fatalf("FromSDJWT: %v", err)
	}
	if got.ID != "https://example.com/creds/sd" {
		t.Errorf("ID mismatch: %s", got.ID)
	}
}

func TestFromJWT_RoundTripBuildParse(t *testing.T) {
	fh := NewFormatHandler(NewCryptoUtils())
	now := time.Now().UTC().Truncate(time.Second)
	vc := &VerifiableCredential{
		Context:      []string{"https://www.w3.org/2018/credentials/v1"},
		ID:           "https://example.com/creds/rt1",
		Type:         []string{"VerifiableCredential"},
		Issuer:       "did:key:zRoundTrip",
		IssuanceDate: now,
		CredentialSubject: CredentialSubject{
			ID: "did:key:zSubRoundTrip",
		},
	}
	payload := map[string]interface{}{
		"vc":  vc,
		"iss": vc.Issuer,
		"sub": vc.CredentialSubject.ID,
		"jti": vc.ID,
	}
	jwtStr := buildJWT(t, payload)

	if parts := strings.Split(jwtStr, "."); len(parts) != 3 {
		t.Errorf("JWT should have 3 parts, got %d", len(parts))
	}

	got, err := fh.FromJWT(jwtStr)
	if err != nil {
		t.Fatalf("FromJWT: %v", err)
	}
	if got.ID != vc.ID {
		t.Errorf("ID round-trip failed: want %q, got %q", vc.ID, got.ID)
	}
}
