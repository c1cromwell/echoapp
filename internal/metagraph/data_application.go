package metagraph

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// signedDataUpdate is the Tessellation 4.x POST /data body (Signed[DataUpdate]).
type signedDataUpdate struct {
	Value  json.RawMessage `json:"value"`
	Proofs []dataProof     `json:"proofs"`
}

type dataProof struct {
	ID        string `json:"id"`
	Signature string `json:"signature"`
}

// SubmitSignedData posts a signed data-application update to {baseURL}/data.
func (c *MetagraphClient) SubmitSignedData(ctx context.Context, baseURL string, update interface{}, signer IdentitySigningConfig) (string, error) {
	if baseURL == "" {
		return "", fmt.Errorf("metagraph: base URL is not configured")
	}
	priv, err := signer.secp256k1PrivateKey()
	if err != nil {
		return "", err
	}

	canonical, hashHex, err := tessellationDataHash(update)
	if err != nil {
		return "", fmt.Errorf("tessellation data hash: %w", err)
	}

	sigDER, err := signTessellationDataHash(hashHex, priv)
	if err != nil {
		return "", fmt.Errorf("sign update: %w", err)
	}

	body, err := json.Marshal(signedDataUpdate{
		Value: json.RawMessage(canonical),
		Proofs: []dataProof{{
			ID:        strings.ToLower(signer.PublicHex),
			Signature: strings.ToLower(hex.EncodeToString(sigDER)),
		}},
	})
	if err != nil {
		return "", fmt.Errorf("marshal signed update: %w", err)
	}

	url := strings.TrimRight(baseURL, "/") + "/data"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("submit data: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read data response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("submit data failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	if len(bytes.TrimSpace(respBody)) == 0 {
		return "accepted", nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return "accepted", nil
	}
	for _, key := range []string{"txHash", "tx_id", "hash", "id"} {
		if v, ok := raw[key]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err == nil && s != "" {
				return s, nil
			}
		}
	}
	return "accepted", nil
}
