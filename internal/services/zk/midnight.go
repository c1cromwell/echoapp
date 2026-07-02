package zk

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type midnightEnvelope struct {
	Commitment    string   `json:"commitment"`
	PublicSignals []string `json:"public_signals,omitempty"`
	Proof         string   `json:"proof"`
}

func verifyMidnightEnvelope(req VerifyRequest) (VerifyResult, error) {
	payload := strings.TrimSpace(strings.TrimPrefix(req.Proof, "midnight:"))
	env, err := parseMidnightEnvelope(payload)
	if err != nil {
		return VerifyResult{Verified: false, Mode: "midnight", Detail: err.Error()}, nil
	}
	expected := commitmentHash(req.SubjectDID, req.ClaimType, req.Nonce)
	if !strings.EqualFold(strings.TrimSpace(env.Commitment), expected) {
		return VerifyResult{
			Verified: false,
			Mode:     "midnight",
			Detail:   "commitment mismatch in midnight envelope",
		}, nil
	}
	if url := strings.TrimSpace(os.Getenv("MIDNIGHT_VERIFY_URL")); url != "" && env.Proof != "" {
		ok, detail := postMidnightVerify(url, env)
		if !ok {
			return VerifyResult{Verified: false, Mode: "midnight", Detail: detail}, nil
		}
		return VerifyResult{Verified: true, Mode: "midnight", Detail: detail}, nil
	}
	return VerifyResult{
		Verified: true,
		Mode:     "midnight",
		Detail:   "Midnight envelope commitment verified",
	}, nil
}

// BuildMidnightEnvelope encodes a commitment-bound midnight: proof string (tests/clients).
func BuildMidnightEnvelope(subjectDID, claimType, nonce, circuitProof string) string {
	env := midnightEnvelope{
		Commitment: commitmentHash(subjectDID, claimType, nonce),
		Proof:      circuitProof,
	}
	raw, _ := json.Marshal(env)
	return "midnight:" + base64.StdEncoding.EncodeToString(raw)
}

func parseMidnightEnvelope(payload string) (midnightEnvelope, error) {
	var env midnightEnvelope
	if err := json.Unmarshal([]byte(payload), &env); err == nil && env.Commitment != "" {
		return env, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return midnightEnvelope{}, fmt.Errorf("invalid midnight envelope")
	}
	if err := json.Unmarshal(decoded, &env); err != nil || env.Commitment == "" {
		return midnightEnvelope{}, fmt.Errorf("invalid midnight envelope json")
	}
	return env, nil
}

func postMidnightVerify(endpoint string, env midnightEnvelope) (bool, string) {
	body, err := json.Marshal(env)
	if err != nil {
		return false, "marshal failed"
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return false, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return false, fmt.Sprintf("verifier HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var out struct {
		Valid    bool   `json:"valid"`
		Verified bool   `json:"verified"`
		Detail   string `json:"detail"`
	}
	_ = json.Unmarshal(raw, &out)
	if out.Valid || out.Verified {
		if out.Detail != "" {
			return true, out.Detail
		}
		return true, "Midnight circuit verified"
	}
	if out.Detail != "" {
		return false, out.Detail
	}
	return false, "midnight verifier rejected proof"
}
