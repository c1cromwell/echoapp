package metagraph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MetagraphConfig holds endpoint URLs for L0, Currency L1, and Data L1.
type MetagraphConfig struct {
	L0URL         string
	CurrencyL1URL string
	DataL1URL     string
	Timeout       time.Duration
}

// MetagraphClient is an HTTP client for the Constellation metagraph APIs.
type MetagraphClient struct {
	config MetagraphConfig
	http   *http.Client
}

// NewMetagraphClient creates a new metagraph gateway client.
func NewMetagraphClient(config MetagraphConfig) *MetagraphClient {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &MetagraphClient{
		config: config,
		http:   &http.Client{Timeout: timeout},
	}
}

// SnapshotEvent represents a finalized metagraph snapshot.
type SnapshotEvent struct {
	Hash         string    `json:"hash"`
	Ordinal      int64     `json:"ordinal"`
	Timestamp    time.Time `json:"timestamp"`
	TxCount      int       `json:"txCount"`
	DataL1Blocks int       `json:"dataL1Blocks"`
}

// CurrencyL1Transaction is a unified submission type for Currency L1 transactions.
type CurrencyL1Transaction struct {
	Type            string `json:"type"`
	TokenLock       *TokenLock       `json:"tokenLock,omitempty"`
	StakeDelegation *StakeDelegation `json:"stakeDelegation,omitempty"`
	WithdrawLock    *WithdrawLock    `json:"withdrawLock,omitempty"`
	FeeTransaction  *FeeTransaction  `json:"feeTransaction,omitempty"`
	AllowSpend      *AllowSpend      `json:"allowSpend,omitempty"`
	AtomicAction    *AtomicAction    `json:"atomicAction,omitempty"`
}

// SubmitCurrencyL1 submits a Currency L1 transaction and returns the tx hash.
func (c *MetagraphClient) SubmitCurrencyL1(ctx context.Context, tx CurrencyL1Transaction) (string, error) {
	return c.submitTransaction(ctx, c.config.CurrencyL1URL+"/transactions", tx)
}

// SubmitDataL1 submits a Data L1 transaction (MerkleCommitment, TrustCommitment).
func (c *MetagraphClient) SubmitDataL1(ctx context.Context, tx interface{}) (string, error) {
	return c.submitTransaction(ctx, c.config.DataL1URL+"/transactions", tx)
}

// QueryValidators returns active L1 validators from the metagraph.
func (c *MetagraphClient) QueryValidators(ctx context.Context) ([]ValidatorSnapshot, error) {
	url := c.config.L0URL + "/validators"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query validators: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query validators: status %d", resp.StatusCode)
	}

	var validators []ValidatorSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&validators); err != nil {
		return nil, fmt.Errorf("decode validators: %w", err)
	}

	return validators, nil
}

// ValidatorSnapshot is a validator's state from the metagraph snapshot.
type ValidatorSnapshot struct {
	ID                string  `json:"id"`
	Address           string  `json:"address"`
	UptimePercent     float64 `json:"uptimePercent"`
	CommissionPercent float64 `json:"commissionPercent"`
	TotalDelegated    int64   `json:"totalDelegated"`
	DelegatorCount    int     `json:"delegatorCount"`
	Layer             string  `json:"layer"`
}

func (c *MetagraphClient) submitTransaction(ctx context.Context, url string, payload interface{}) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal tx: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(
		jsonReader(body),
	))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("submit tx: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("submit tx failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		TxHash string `json:"txHash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode tx response: %w", err)
	}

	return result.TxHash, nil
}

type byteReaderCloser struct {
	data []byte
	pos  int
}

func (r *byteReaderCloser) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *byteReaderCloser) Close() error { return nil }

func jsonReader(data []byte) io.Reader {
	return &byteReaderCloser{data: data}
}
