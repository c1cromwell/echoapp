// Package disclosure implements SD-JWT selective disclosure for Echo Passport (WO-295).
package disclosure

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const digestAlg = "sha-256"

var (
	ErrInvalidSDJWT       = errors.New("invalid SD-JWT")
	ErrDisclosureMismatch = errors.New("disclosure digest mismatch")
	ErrFieldNotAllowed    = errors.New("disclosed field not in allowed set")
)

// BuildFromSubjectClaims seals subject claims into an SD-JWT presentation shell.
// Returns the SD-JWT (JWT + ~disclosures) and a map of claim name → full disclosure string.
func BuildFromSubjectClaims(jwtCore string, subjectClaims map[string]interface{}) (string, map[string]string, error) {
	jwtCore = strings.TrimSuffix(strings.TrimSpace(jwtCore), "~")
	parts := strings.Split(jwtCore, ".")
	if len(parts) != 3 {
		return "", nil, ErrInvalidSDJWT
	}

	disclosures := make(map[string]string)
	digests := make([]string, 0, len(subjectClaims))
	for name, value := range subjectClaims {
		salt := make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			return "", nil, err
		}
		disc, digest, err := sealClaim(salt, name, value)
		if err != nil {
			return "", nil, err
		}
		disclosures[name] = disc
		digests = append(digests, digest)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("decode jwt payload: %w", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", nil, err
	}

	vcRaw, ok := payload["vc"]
	if !ok {
		return "", nil, errors.New("jwt missing vc claim")
	}
	vcMap, ok := vcRaw.(map[string]interface{})
	if !ok {
		return "", nil, errors.New("vc claim must be object")
	}
	subjectRaw, ok := vcMap["credentialSubject"]
	if !ok {
		return "", nil, errors.New("vc missing credentialSubject")
	}
	subject, ok := subjectRaw.(map[string]interface{})
	if !ok {
		return "", nil, errors.New("credentialSubject must be object")
	}
	for k := range subjectClaims {
		delete(subject, k)
	}
	subject["_sd"] = digests
	subject["_sd_alg"] = digestAlg
	vcMap["credentialSubject"] = subject
	payload["vc"] = vcMap

	newPayload, err := json.Marshal(payload)
	if err != nil {
		return "", nil, err
	}
	parts[1] = base64.RawURLEncoding.EncodeToString(newPayload)
	out := strings.Join(parts, ".")
	for _, disc := range disclosures {
		out += "~" + disc
	}
	out += "~"
	return out, disclosures, nil
}

// PresentSubset returns an SD-JWT with only the requested claim disclosures appended.
func PresentSubset(fullSDJWT string, discloseFields []string) (string, error) {
	fullSDJWT = strings.TrimSpace(fullSDJWT)
	idx := strings.Index(fullSDJWT, "~")
	if idx < 0 {
		return "", ErrInvalidSDJWT
	}
	jwtPart := fullSDJWT[:idx]
	tail := fullSDJWT[idx+1:]
	allDisc := strings.Split(strings.TrimSuffix(tail, "~"), "~")

	want := make(map[string]struct{}, len(discloseFields))
	for _, f := range discloseFields {
		want[f] = struct{}{}
	}

	var selected []string
	for _, disc := range allDisc {
		if disc == "" {
			continue
		}
		name, err := claimNameFromDisclosure(disc)
		if err != nil {
			return "", err
		}
		if _, ok := want[name]; ok {
			selected = append(selected, disc)
		}
	}
	if len(selected) != len(discloseFields) {
		return "", fmt.Errorf("missing disclosures for requested fields")
	}
	out := jwtPart
	for _, disc := range selected {
		out += "~" + disc
	}
	out += "~"
	return out, nil
}

// ValidatePresentation checks that every disclosed claim is in allowedFields.
func ValidatePresentation(sdJWT string, allowedFields []string) error {
	allowed := make(map[string]struct{}, len(allowedFields))
	for _, f := range allowedFields {
		allowed[f] = struct{}{}
	}
	idx := strings.Index(sdJWT, "~")
	if idx < 0 {
		return nil // bare JWT — nothing selectively disclosed
	}
	tail := sdJWT[idx+1:]
	for _, disc := range strings.Split(strings.TrimSuffix(tail, "~"), "~") {
		if disc == "" {
			continue
		}
		name, err := claimNameFromDisclosure(disc)
		if err != nil {
			return err
		}
		if _, ok := allowed[name]; !ok {
			return fmt.Errorf("%w: %s", ErrFieldNotAllowed, name)
		}
		if err := verifyDisclosureDigest(sdJWT[:idx], disc); err != nil {
			return err
		}
	}
	return nil
}

// DisclosedClaims merges disclosed claim values into a credentialSubject map.
func DisclosedClaims(sdJWT string) (map[string]interface{}, error) {
	idx := strings.Index(sdJWT, "~")
	if idx < 0 {
		return map[string]interface{}{}, nil
	}
	out := make(map[string]interface{})
	for _, disc := range strings.Split(strings.TrimSuffix(sdJWT[idx+1:], "~"), "~") {
		if disc == "" {
			continue
		}
		var arr []interface{}
		raw, err := base64.RawURLEncoding.DecodeString(disc)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, err
		}
		if len(arr) < 3 {
			return nil, ErrInvalidSDJWT
		}
		name, _ := arr[1].(string)
		out[name] = arr[2]
	}
	return out, nil
}

func sealClaim(salt []byte, name string, value interface{}) (disclosure, digest string, err error) {
	arr := []interface{}{base64.RawURLEncoding.EncodeToString(salt), name, value}
	raw, err := json.Marshal(arr)
	if err != nil {
		return "", "", err
	}
	disclosure = base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256(raw)
	digest = base64.RawURLEncoding.EncodeToString(sum[:])
	return disclosure, digest, nil
}

func claimNameFromDisclosure(disc string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(disc)
	if err != nil {
		return "", err
	}
	var arr []interface{}
	if err := json.Unmarshal(raw, &arr); err != nil {
		return "", err
	}
	if len(arr) < 2 {
		return "", ErrInvalidSDJWT
	}
	name, ok := arr[1].(string)
	if !ok || name == "" {
		return "", ErrInvalidSDJWT
	}
	return name, nil
}

func verifyDisclosureDigest(jwtPart, disc string) error {
	raw, err := base64.RawURLEncoding.DecodeString(disc)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	digest := base64.RawURLEncoding.EncodeToString(sum[:])

	payloadBytes, err := base64.RawURLEncoding.DecodeString(strings.Split(jwtPart, ".")[1])
	if err != nil {
		return err
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return err
	}
	vc := payload["vc"].(map[string]interface{})
	subject := vc["credentialSubject"].(map[string]interface{})
	sdRaw, ok := subject["_sd"].([]interface{})
	if !ok {
		return ErrDisclosureMismatch
	}
	for _, d := range sdRaw {
		if ds, ok := d.(string); ok && ds == digest {
			return nil
		}
	}
	return ErrDisclosureMismatch
}
