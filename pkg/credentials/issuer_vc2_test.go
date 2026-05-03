package credentials

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestMarshalVC2JSONLD(t *testing.T) {
	exp := time.Date(2027, 4, 27, 0, 0, 0, 0, time.UTC)
	i := &Issuer{
		config: &Config{
			CredentialConfig: CredentialConfig{UseW3CVC2: true},
		},
	}
	vc := &VerifiableCredential{
		Type:         []string{"VerifiableCredential", "ProofOfHumanity"},
		ID:           "urn:credential:x",
		Issuer:       "did:key:z6MkrTEST",
		IssuanceDate: time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		ExpirationDate: &exp,
		CredentialSubject: CredentialSubject{
			ID: "did:key:z2DAuser",
			Claims: map[string]interface{}{
				"humanityProof": true,
			},
		},
		CredentialStatus: &CredentialStatus{
			ID:                   "https://identity-metagraph.echo.app/status/0#42",
			Type:                 "StatusList2021Entry",
			StatusPurpose:        "revocation",
			StatusListIndex:      "42",
			StatusListCredential: "https://identity-metagraph.echo.app/status/0",
		},
		Proof: Proof{
			Type:                 "DataIntegrityProof",
			Cryptosuite:          "ecdsa-2019",
			Created:              time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC),
			VerificationMethod:   "did:key:z6MkrTEST#key1",
			ProofPurpose:         "assertionMethod",
			ChallengeNonce:       "n1",
			ProofValue:           "zQmECHO_DEV_PLACEHOLDER_SIGN_WITH_REAL_ECDSA_2019",
		},
	}

	ld, err := i.marshalVC2JSONLD(vc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ld, `"validFrom"`) || !strings.Contains(ld, `"validUntil"`) {
		t.Fatalf("expected validFrom/validUntil in JSON-LD: %s", ld)
	}
	if !strings.Contains(ld, `"https://www.w3.org/ns/credentials/v2"`) {
		t.Fatalf("expected v2 context: %s", ld)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(ld), &m); err != nil {
		t.Fatal(err)
	}
	proof, _ := m["proof"].(map[string]interface{})
	if proof["nonce"] != "n1" {
		t.Fatalf("expected nonce in proof, got %v", proof)
	}
}
