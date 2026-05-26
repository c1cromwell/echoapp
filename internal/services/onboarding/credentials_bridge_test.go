package onboarding

import (
	"testing"
	"time"

	pkgcred "github.com/thechadcromwell/echoapp/pkg/credentials"
)

func TestVerifyCredentialProof_ed25519(t *testing.T) {
	crypto := pkgcred.NewCryptoUtils()
	pub, priv, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	issuer := &TrustedIssuer{
		ID:                          "crypto_issuer",
		Name:                        "Test Issuer",
		Status:                      "active",
		TrustLevel:                  TrustLevelHigh,
		VerificationPublicKeyBase64: pub,
	}

	vc := &VerifiableCredential{
		ID:             "cred_123",
		Issuer:         issuer.ID,
		IssuanceDate:   time.Now().Add(-time.Hour).Format(time.RFC3339),
		ExpirationDate: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		CredentialSubject: map[string]interface{}{
			"name": "Alice",
		},
	}

	payload, err := canonicalOnboardingCredentialBytes(vc)
	if err != nil {
		t.Fatalf("canonicalOnboardingCredentialBytes: %v", err)
	}

	sig, err := crypto.SignMessage(priv, payload)
	if err != nil {
		t.Fatalf("SignMessage: %v", err)
	}
	vc.ProofValue = sig

	if err := verifyCredentialProof(crypto, issuer, vc); err != nil {
		t.Fatalf("expected valid proof, got %v", err)
	}

	vc.ProofValue = "invalid_signature"
	if err := verifyCredentialProof(crypto, issuer, vc); err == nil {
		t.Fatal("expected invalid signature to fail")
	}
}

func TestVerifyCredentialProof_legacyIssuerWithoutKey(t *testing.T) {
	issuer := &TrustedIssuer{ID: "legacy", Status: "active"}
	vc := &VerifiableCredential{
		Issuer:       issuer.ID,
		IssuanceDate: time.Now().Format(time.RFC3339),
		ProofValue:   "presence_only_sig",
	}
	if err := verifyCredentialProof(nil, issuer, vc); err != nil {
		t.Fatalf("legacy issuer without published key should pass presence check: %v", err)
	}

	vc.ProofValue = ""
	if err := verifyCredentialProof(nil, issuer, vc); err == nil {
		t.Fatal("expected empty proof to fail")
	}
}
