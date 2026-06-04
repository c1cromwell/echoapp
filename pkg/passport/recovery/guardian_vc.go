package recovery

import (
	"encoding/json"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/credentials"
)

// GuardianCredential builds the EchoGuardianCredential VC payload (ADR 0004 sketch).
func GuardianCredential(holderDID, guardianDID string, shareIndex int, acceptedAt time.Time) *credentials.VerifiableCredential {
	return &credentials.VerifiableCredential{
		Context:      []string{"https://www.w3.org/ns/credentials/v2"},
		Type:         []string{"VerifiableCredential", "EchoGuardianCredential"},
		Issuer:       "did:key:zEchoIdentityServicePlaceholder",
		IssuanceDate: acceptedAt.UTC(),
		CredentialSubject: credentials.CredentialSubject{
			ID: guardianDID,
			Claims: map[string]interface{}{
				"guardianFor": holderDID,
				"shareIndex":  shareIndex,
				"acceptedAt":  acceptedAt.UTC().Format(time.RFC3339),
			},
		},
	}
}

// GuardianCredentialJSON returns the VC as indented JSON-LD for wallet storage.
func GuardianCredentialJSON(holderDID, guardianDID string, shareIndex int, acceptedAt time.Time) (string, error) {
	vc := GuardianCredential(holderDID, guardianDID, shareIndex, acceptedAt)
	raw, err := json.MarshalIndent(vc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
