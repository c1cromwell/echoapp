package evidence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultVerifyBase = "https://digitalevidence.constellationnetwork.io/verify/"

// HTTPClient submits fingerprints to Constellation Digital Evidence REST API.
type HTTPClient struct {
	config *ClientConfig
	http   *http.Client
}

// NewHTTPClient creates a Digital Evidence API client.
func NewHTTPClient(config *ClientConfig) (*HTTPClient, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &HTTPClient{
		config: config,
		http:   &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// SubmitFingerprint posts a fingerprint request and returns the DE event metadata.
func (c *HTTPClient) SubmitFingerprint(ctx context.Context, req *FingerprintRequest) (*FingerprintResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(c.config.BaseURL, "/") + "/evidence"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("X-Organization-Id", c.config.OrganizationID)
	httpReq.Header.Set("X-Tenant-Id", c.config.TenantID)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("digital evidence submit: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("digital evidence submit: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("decode digital evidence response: %w", err)
	}

	out := &FingerprintResponse{Status: "verified"}
	for _, key := range []string{"event_id", "eventId"} {
		if v, ok := raw[key]; ok {
			_ = json.Unmarshal(v, &out.EventID)
			break
		}
	}
	for _, key := range []string{"explorer_url", "explorerUrl", "verification_url", "verificationUrl"} {
		if v, ok := raw[key]; ok {
			_ = json.Unmarshal(v, &out.ExplorerURL)
			break
		}
	}
	if out.ExplorerURL == "" && out.EventID != "" {
		out.ExplorerURL = defaultVerifyBase + out.EventID
	}
	for _, key := range []string{"anchored_at", "anchoredAt", "timestamp"} {
		if v, ok := raw[key]; ok {
			_ = json.Unmarshal(v, &out.AnchoredAt)
			break
		}
	}
	if out.EventID == "" {
		return nil, fmt.Errorf("digital evidence response missing event_id")
	}
	return out, nil
}

// VerifyFingerprint fetches verification status for a DE event (read-only).
func (c *HTTPClient) VerifyFingerprint(ctx context.Context, eventID string) (*VerificationResult, error) {
	url := strings.TrimRight(c.config.BaseURL, "/") + "/evidence/" + eventID
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return &VerificationResult{EventID: eventID, Status: "pending"}, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("digital evidence verify: status %d", resp.StatusCode)
	}

	var result VerificationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &VerificationResult{
			EventID:     eventID,
			Status:      "verified",
			ExplorerURL: defaultVerifyBase + eventID,
		}, nil
	}
	if result.EventID == "" {
		result.EventID = eventID
	}
	if result.ExplorerURL == "" {
		result.ExplorerURL = defaultVerifyBase + eventID
	}
	return &result, nil
}

// LoadClientConfigFromEnv reads DIGITAL_EVIDENCE_* environment variables.
func LoadClientConfigFromEnv() (*ClientConfig, error) {
	cfg := &ClientConfig{
		APIKey:         strings.TrimSpace(osGetenv("DIGITAL_EVIDENCE_API_KEY")),
		OrganizationID: strings.TrimSpace(osGetenv("DIGITAL_EVIDENCE_ORG_ID")),
		TenantID:       strings.TrimSpace(osGetenv("DIGITAL_EVIDENCE_TENANT_ID")),
		BaseURL:        strings.TrimSpace(osGetenv("DIGITAL_EVIDENCE_BASE_URL")),
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://evidence.constellationnetwork.io"
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("DIGITAL_EVIDENCE_API_KEY not set")
	}
	return cfg, cfg.Validate()
}

// osGetenv is overridden in tests.
var osGetenv = func(key string) string {
	return os.Getenv(key)
}
