package onboarding

import (
	"encoding/json"
	"fmt"

	pkgcred "github.com/thechadcromwell/echoapp/pkg/credentials"
)

// verifyCredentialProof checks the credential proof. When the trusted issuer publishes
// VerificationPublicKeyBase64, verification delegates to pkg/credentials crypto (Wave A.3).
// Entries without a published key keep the legacy presence-only check for unit fixtures.
func verifyCredentialProof(crypto *pkgcred.CryptoUtils, issuer *TrustedIssuer, vc *VerifiableCredential) error {
	if vc.ProofValue == "" {
		return fmt.Errorf("credential has no valid signature")
	}
	if issuer == nil {
		return fmt.Errorf("issuer required")
	}
	if issuer.VerificationPublicKeyBase64 == "" {
		return nil
	}
	if crypto == nil {
		crypto = pkgcred.NewCryptoUtils()
	}
	payload, err := canonicalOnboardingCredentialBytes(vc)
	if err != nil {
		return err
	}
	ok, err := crypto.VerifySignature(issuer.VerificationPublicKeyBase64, payload, vc.ProofValue)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("credential signature verification failed")
	}
	return nil
}

func canonicalOnboardingCredentialBytes(vc *VerifiableCredential) ([]byte, error) {
	if vc == nil {
		return nil, fmt.Errorf("credential required")
	}
	canonical := struct {
		ID         string                 `json:"id"`
		Issuer     string                 `json:"issuer"`
		Issuance   string                 `json:"issuanceDate"`
		Expiration string                 `json:"expirationDate,omitempty"`
		Subject    map[string]interface{} `json:"credentialSubject"`
	}{
		ID:         vc.ID,
		Issuer:     vc.Issuer,
		Issuance:   vc.IssuanceDate,
		Expiration: vc.ExpirationDate,
		Subject:    vc.CredentialSubject,
	}
	return json.Marshal(canonical)
}
