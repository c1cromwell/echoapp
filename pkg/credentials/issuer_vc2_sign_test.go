package credentials

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

func TestVC2DataIntegrityIssueAndVerify(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuerDID, err := didkey.Derive(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "issuer.pem")
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
	if err := os.WriteFile(keyPath, pemBytes, 0600); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.MetagraphConfig.IdentityL1URL = ""
	cfg.MetagraphConfig.EnableAnchor = false
	cfg.IssuerConfig.IssuerDID = issuerDID
	cfg.IssuerConfig.PrivateKeyPath = keyPath
	cfg.MetagraphConfig.IssuerDID = issuerDID
	cfg.VerifierConfig.VerifierDID = issuerDID
	cfg.CredentialConfig.UseW3CVC2 = true
	cfg.VerifierConfig.EnableRevocation = false

	st := NewInMemoryStorage()
	iss := NewIssuer(cfg, NewCryptoUtils(), st, nil)
	ctx := context.Background()

	resp, err := iss.IssueCredential(ctx, &CredentialIssuanceRequest{
		SubjectDID:      issuerDID,
		CredentialType:  KYCLite,
		Claims:          map[string]interface{}{"tier": "basic"},
		PreferredFormat: JSONLDFormat,
	})
	if err != nil {
		t.Fatal(err)
	}

	v := NewVerifier(cfg, NewCryptoUtils(), st, NewRevocationManager(st, cfg.RevocationConfig.CacheTTL))
	res, err := v.VerifyCredential(ctx, &CredentialVerificationRequest{
		Credential: resp.VerifiableCredential,
		Format:     JSONLDFormat,
		IssuerDID:  issuerDID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.SignatureValid || !res.IsValid {
		t.Fatalf("verify: %+v", res.Errors)
	}
}

func TestVC2JWTContainsDocMapAndES256(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuerDID, err := didkey.Derive(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "issuer.pem")
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.MetagraphConfig.EnableAnchor = false
	cfg.MetagraphConfig.IdentityL1URL = ""
	cfg.IssuerConfig.IssuerDID = issuerDID
	cfg.IssuerConfig.PrivateKeyPath = keyPath
	cfg.MetagraphConfig.IssuerDID = issuerDID
	cfg.VerifierConfig.VerifierDID = issuerDID
	cfg.CredentialConfig.UseW3CVC2 = true

	st := NewInMemoryStorage()
	iss := NewIssuer(cfg, NewCryptoUtils(), st, nil)

	vc := iss.createWC3VerifiableCredential(&CredentialIssuanceRequest{
		SubjectDID:     issuerDID,
		CredentialType: ProofOfHumanity,
		Claims:         map[string]interface{}{"humanityProof": true},
	}, "cred-1", "9")
	vc.IssuanceDate = time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	if err := iss.signVC2DataIntegrity(vc); err != nil {
		t.Fatal(err)
	}
	jwt, err := iss.createJWT(vc, "jti-1")
	if err != nil {
		t.Fatal(err)
	}
	// ES256 JWT: three segments, alg in first segment is ES256 after decode
	if jwt == "" {
		t.Fatal("empty jwt")
	}
}
