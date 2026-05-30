package disclosure

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func buildTestJWT(t *testing.T, subjectClaims map[string]interface{}) string {
	t.Helper()
	vc := map[string]interface{}{
		"@context": []string{"https://www.w3.org/2018/credentials/v1"},
		"type":     []string{"VerifiableCredential", "ProofOfHumanity"},
		"credentialSubject": subjectClaims,
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"vc":  vc,
		"iss": "did:key:zIssuer",
		"sub": "did:key:zSubject",
	})
	header, _ := json.Marshal(map[string]string{"alg": "EdDSA", "typ": "JWT"})
	h := base64.RawURLEncoding.EncodeToString(header)
	p := base64.RawURLEncoding.EncodeToString(payload)
	return h + "." + p + ".sig"
}

func TestBuildPresentValidateRoundTrip(t *testing.T) {
	jwt := buildTestJWT(t, map[string]interface{}{
		"id":    "did:key:zSubject",
		"tier":  4,
		"email": "sealed@example.com",
	})
	full, _, err := BuildFromSubjectClaims(jwt, map[string]interface{}{
		"tier":  4,
		"email": "sealed@example.com",
	})
	if err != nil {
		t.Fatalf("BuildFromSubjectClaims: %v", err)
	}
	if !strings.Contains(full, "~") {
		t.Fatal("expected disclosures")
	}

	presented, err := PresentSubset(full, []string{"tier"})
	if err != nil {
		t.Fatalf("PresentSubset: %v", err)
	}
	if err := ValidatePresentation(presented, []string{"tier"}); err != nil {
		t.Fatalf("ValidatePresentation: %v", err)
	}
	if err := ValidatePresentation(presented, []string{"email"}); err == nil {
		t.Fatal("expected rejection when extra field allowed but not disclosed")
	}

	claims, err := DisclosedClaims(presented)
	if err != nil {
		t.Fatalf("DisclosedClaims: %v", err)
	}
	if claims["tier"] != float64(4) {
		t.Fatalf("tier claim = %v", claims["tier"])
	}
	if _, ok := claims["email"]; ok {
		t.Fatal("email should not be disclosed")
	}
}

func TestValidatePresentationRejectsExtraDisclosure(t *testing.T) {
	jwt := buildTestJWT(t, map[string]interface{}{"id": "did:key:zSubject", "tier": 4})
	full, _, err := BuildFromSubjectClaims(jwt, map[string]interface{}{"tier": 4})
	if err != nil {
		t.Fatal(err)
	}
	presented, err := PresentSubset(full, []string{"tier"})
	if err != nil {
		t.Fatal(err)
	}
	err = ValidatePresentation(presented, []string{})
	if err == nil {
		t.Fatal("expected ErrFieldNotAllowed")
	}
}
