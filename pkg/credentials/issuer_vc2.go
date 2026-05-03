package credentials

import (
	"encoding/json"
	"time"
)

// vc2CredentialDocMap builds the VC 2.0 JSON-LD object as nested maps. When includeProof
// is false, the proof object is omitted (DataIntegrity signing input).
func vc2CredentialDocMap(vc *VerifiableCredential, includeProof bool) (map[string]interface{}, error) {
	subj := map[string]interface{}{
		"id": vc.CredentialSubject.ID,
	}
	for k, v := range vc.CredentialSubject.Claims {
		subj[k] = v
	}

	doc := map[string]interface{}{
		"@context": []string{
			"https://www.w3.org/ns/credentials/v2",
			"https://w3id.org/security/multikey/v1",
		},
		"type":              vc.Type,
		"issuer":            vc.Issuer,
		"validFrom":         vc.IssuanceDate.UTC().Format(time.RFC3339),
		"credentialSubject": subj,
	}
	if vc.ID != "" {
		doc["id"] = vc.ID
	}
	if vc.ExpirationDate != nil {
		doc["validUntil"] = vc.ExpirationDate.UTC().Format(time.RFC3339)
	}
	if vc.CredentialStatus != nil {
		doc["credentialStatus"] = map[string]interface{}{
			"id":                   vc.CredentialStatus.ID,
			"type":                 vc.CredentialStatus.Type,
			"statusPurpose":        vc.CredentialStatus.StatusPurpose,
			"statusListIndex":      vc.CredentialStatus.StatusListIndex,
			"statusListCredential": vc.CredentialStatus.StatusListCredential,
		}
	}

	if includeProof && vc.Proof.Type != "" {
		proof := map[string]interface{}{
			"type":               vc.Proof.Type,
			"cryptosuite":        vc.Proof.Cryptosuite,
			"created":            vc.Proof.Created.UTC().Format(time.RFC3339Nano),
			"verificationMethod": vc.Proof.VerificationMethod,
			"proofPurpose":       vc.Proof.ProofPurpose,
			"proofValue":         vc.Proof.ProofValue,
		}
		if vc.Proof.ChallengeNonce != "" {
			proof["nonce"] = vc.Proof.ChallengeNonce
		}
		doc["proof"] = proof
	}

	return doc, nil
}

// marshalVC2JSONLD builds W3C VC 2.0 JSON-LD (WO-274) for wallet delivery.
func (i *Issuer) marshalVC2JSONLD(vc *VerifiableCredential) (string, error) {
	doc, err := vc2CredentialDocMap(vc, true)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(doc)
	if err != nil {
		return "", NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to encode VC 2.0 JSON-LD",
			err.Error(),
		)
	}
	return string(b), nil
}
