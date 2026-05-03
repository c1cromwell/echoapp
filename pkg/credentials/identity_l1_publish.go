package credentials

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// vcIssuanceL1Wire matches Scala com.echo.shared_data.types.VCIssuanceUpdate JSON
// (POST {IDENTITY_L1_URL}/transactions).
type vcIssuanceL1Wire struct {
	CredentialID   string `json:"credentialId"`
	SubjectDID     string `json:"subjectDID"`
	IssuerDID      string `json:"issuerDID"`
	CredentialType string `json:"credentialType"`
	IssuedAt       int64  `json:"issuedAt"`
	SchemaVersion  string `json:"schemaVersion"`
}

// publishVCIssuanceMetadata posts VC issuance metadata to Identity Metagraph L1 (WO-274).
func (i *Issuer) publishVCIssuanceMetadata(ctx context.Context, credentialID string, req *CredentialIssuanceRequest, issuedAt time.Time) (string, error) {
	if i.config.MetagraphConfig.IdentityL1URL == "" {
		return "", nil
	}
	if !i.config.MetagraphConfig.EnableAnchor {
		return "", nil
	}

	body := vcIssuanceL1Wire{
		CredentialID:   fmt.Sprintf("urn:uuid:%s", credentialID),
		SubjectDID:     req.SubjectDID,
		IssuerDID:      i.effectiveIssuerDID(),
		CredentialType: string(req.CredentialType),
		IssuedAt:       issuedAt.UnixMilli(),
		SchemaVersion:  "v1.0.0",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	timeout := i.config.MetagraphConfig.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	c := &http.Client{Timeout: timeout}
	url := i.config.MetagraphConfig.IdentityL1URL + "/transactions"
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	r.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(r)
	if err != nil {
		return "", fmt.Errorf("identity L1 POST: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("identity L1: status %d: %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		TxHash string `json:"txHash"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("identity L1 decode: %w", err)
	}
	return out.TxHash, nil
}
