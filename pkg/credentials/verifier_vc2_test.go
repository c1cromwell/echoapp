package credentials

import (
	"context"
	"testing"
	"time"
)

func TestParseJSONLDCredential_VC2FlatSubject(t *testing.T) {
	st := NewInMemoryStorage()
	cfg := DefaultConfig()
	v := NewVerifier(cfg, NewCryptoUtils(), st, NewRevocationManager(st, cfg.RevocationConfig.CacheTTL))

	raw := `{
		"@context": ["https://www.w3.org/ns/credentials/v2"],
		"type": ["VerifiableCredential", "KYCLite"],
		"id": "urn:credential:abc",
		"issuer": "did:key:issuer",
		"validFrom": "2026-04-27T10:00:00Z",
		"validUntil": "2028-04-27T10:00:00Z",
		"credentialSubject": {
			"id": "did:key:subj",
			"tier": "basic"
		},
		"proof": {
			"type": "DataIntegrityProof",
			"cryptosuite": "ecdsa-2019",
			"created": "2026-04-27T10:00:00Z",
			"verificationMethod": "did:key:issuer#k1",
			"proofPurpose": "assertionMethod",
			"proofValue": "zQmECHO_DEV_PLACEHOLDER_SIGN_WITH_REAL_ECDSA_2019"
		}
	}`

	vc, err := v.parseJSONLDCredential(raw)
	if err != nil {
		t.Fatal(err)
	}
	if vc.CredentialSubject.ID != "did:key:subj" {
		t.Fatalf("subject id: %q", vc.CredentialSubject.ID)
	}
	if vc.CredentialSubject.Claims["tier"] != "basic" {
		t.Fatalf("claims: %+v", vc.CredentialSubject.Claims)
	}
	if vc.IssuanceDate.IsZero() {
		t.Fatal("expected issuance from validFrom")
	}
	if vc.ExpirationDate == nil || !vc.ExpirationDate.Equal(time.Date(2028, 4, 27, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("expiration: %v", vc.ExpirationDate)
	}

	cfg2 := DefaultConfig()
	cfg2.VerifierConfig.EnableRevocation = false
	st2 := NewInMemoryStorage()
	v2 := NewVerifier(cfg2, NewCryptoUtils(), st2, NewRevocationManager(st2, cfg2.RevocationConfig.CacheTTL))
	res, err := v2.VerifyCredential(context.Background(), &CredentialVerificationRequest{
		Credential: raw,
		Format:     JSONLDFormat,
		IssuerDID:  "did:key:issuer",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.SignatureValid {
		t.Fatalf("expected placeholder DataIntegrity proof to verify: %+v", res.Errors)
	}
}

func TestVerifySignature_DataIntegrityNonPlaceholderFails(t *testing.T) {
	st := NewInMemoryStorage()
	cfg := DefaultConfig()
	v := NewVerifier(cfg, NewCryptoUtils(), st, NewRevocationManager(st, cfg.RevocationConfig.CacheTTL))
	vc := &VerifiableCredential{
		Proof: Proof{
			Type:        "DataIntegrityProof",
			Cryptosuite: "ecdsa-2019",
			ProofValue:  "real-signature-would-go-here",
		},
	}
	ok, err := v.verifySignature(vc, "did:x")
	if err == nil || ok {
		t.Fatalf("expected error, ok=%v err=%v", ok, err)
	}
}
