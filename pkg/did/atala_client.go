package did

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AtalaClient provides an HTTP client for Atala PRISM API interactions
type AtalaClient struct {
	baseURL      string
	apiKey       string
	apiSecret    string
	httpClient   *http.Client
	timeout      time.Duration
	maxRetries   int
	retryBackoff time.Duration
	mu           sync.RWMutex
	lastUsedTime time.Time
}

// NewAtalaClient creates a new Atala PRISM client
func NewAtalaClient(config *AtalaPRISMConfig) *AtalaClient {
	// Create a custom transport with connection pooling
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		MaxIdleConns:        config.ConnectionPool,
		MaxIdleConnsPerHost: config.ConnectionPool,
		IdleConnTimeout:     90 * time.Second,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &AtalaClient{
		baseURL:      strings.TrimSuffix(config.Endpoint, "/"),
		apiKey:       config.APIKey,
		apiSecret:    config.APISecret,
		httpClient:   httpClient,
		timeout:      config.Timeout,
		maxRetries:   config.MaxRetries,
		retryBackoff: config.RetryBackoff,
	}
}

// CreateDID creates a new DID via Atala PRISM
func (c *AtalaClient) CreateDID(ctx context.Context, document *DIDDocument) (string, error) {
	payload := map[string]interface{}{
		"document": document,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", NewDIDError(ErrCodeAtalaPRISMError, "Failed to marshal DID document", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/dids", body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", NewDIDError(ErrCodeAtalaPRISMError, "Failed to parse Atala response", err)
	}

	did, ok := result["did"].(string)
	if !ok {
		return "", NewDIDError(ErrCodeAtalaPRISMError, "DID not found in response", nil)
	}

	return did, nil
}

// ResolveDID resolves a DID via Atala PRISM
func (c *AtalaClient) ResolveDID(ctx context.Context, did string) (*DIDDocument, error) {
	endpoint := fmt.Sprintf("/v1/dids/%s/resolve", did)
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, NewDIDError(ErrCodeAtalaPRISMError, "Failed to parse resolution response", err)
	}

	// Extract and deserialize the DID document
	docData, ok := result["document"]
	if !ok {
		return nil, NewDIDError(ErrCodeResolutionFailed, "Document not found in resolution response", nil)
	}

	docBytes, err := json.Marshal(docData)
	if err != nil {
		return nil, NewDIDError(ErrCodeAtalaPRISMError, "Failed to serialize document", err)
	}

	document := &DIDDocument{}
	if err := json.Unmarshal(docBytes, document); err != nil {
		return nil, NewDIDError(ErrCodeAtalaPRISMError, "Failed to parse DID document", err)
	}

	return document, nil
}

// UpdateDID updates a DID document via Atala PRISM
func (c *AtalaClient) UpdateDID(ctx context.Context, did string, document *DIDDocument) error {
	payload := map[string]interface{}{
		"document": document,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return NewDIDError(ErrCodeAtalaPRISMError, "Failed to marshal DID document", err)
	}

	endpoint := fmt.Sprintf("/v1/dids/%s", did)
	_, err = c.doRequest(ctx, "PUT", endpoint, body)
	return err
}

// AnchorDID anchors a DID to the Cardano blockchain via Atala PRISM
func (c *AtalaClient) AnchorDID(ctx context.Context, did string, document *DIDDocument) (string, error) {
	payload := map[string]interface{}{
		"did":      did,
		"document": document,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", NewDIDError(ErrCodeAtalaPRISMError, "Failed to marshal anchor request", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/dids/anchor", body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", NewDIDError(ErrCodeAtalaPRISMError, "Failed to parse anchor response", err)
	}

	txHash, ok := result["transactionHash"].(string)
	if !ok {
		return "", NewDIDError(ErrCodeAnchoringFailed, "Transaction hash not found in response", nil)
	}

	return txHash, nil
}

// VerifyDIDDocument verifies a DID document signature via Atala PRISM
func (c *AtalaClient) VerifyDIDDocument(ctx context.Context, document *DIDDocument) (bool, error) {
	payload := map[string]interface{}{
		"document": document,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false, NewDIDError(ErrCodeAtalaPRISMError, "Failed to marshal DID document", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/dids/verify", body)
	if err != nil {
		return false, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return false, NewDIDError(ErrCodeAtalaPRISMError, "Failed to parse verification response", err)
	}

	valid, ok := result["valid"].(bool)
	if !ok {
		return false, NewDIDError(ErrCodeAtalaPRISMError, "Valid flag not found in response", nil)
	}

	return valid, nil
}

// GetAnchorStatus retrieves the anchor status via Atala PRISM
func (c *AtalaClient) GetAnchorStatus(ctx context.Context, txHash string) (string, error) {
	endpoint := fmt.Sprintf("/v1/transactions/%s/status", txHash)
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", NewDIDError(ErrCodeAtalaPRISMError, "Failed to parse status response", err)
	}

	status, ok := result["status"].(string)
	if !ok {
		return "", NewDIDError(ErrCodeAtalaPRISMError, "Status not found in response", nil)
	}

	return status, nil
}

// doRequest performs an HTTP request with retry logic
func (c *AtalaClient) doRequest(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	c.mu.Lock()
	c.lastUsedTime = time.Now()
	c.mu.Unlock()

	url := c.baseURL + path
	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, NewDIDError(ErrCodeTimeout, "Request context cancelled", ctx.Err())
		default:
		}

		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, NewDIDError(ErrCodeAtalaPRISMError, "Failed to create request", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		if c.apiKey != "" {
			req.Header.Set("X-API-Key", c.apiKey)
		}
		if c.apiSecret != "" {
			req.Header.Set("X-API-Secret", c.apiSecret)
		}

		// Add body if present
		if len(body) > 0 {
			req.Body = io.NopCloser(bytes.NewReader(body))
			req.ContentLength = int64(len(body))
		}

		// Execute request
		httpResp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			// Exponential backoff on error
			if attempt < c.maxRetries-1 {
				time.Sleep(c.retryBackoff * time.Duration(1<<uint(attempt)))
			}
			continue
		}

		// Read response body
		respBody, err := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()

		if err != nil {
			lastErr = err
			if attempt < c.maxRetries-1 {
				time.Sleep(c.retryBackoff * time.Duration(1<<uint(attempt)))
			}
			continue
		}

		// Check HTTP status code
		if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
			return respBody, nil
		}

		// Parse error response
		errResp := &AtalaResponse{}
		if err := json.Unmarshal(respBody, errResp); err != nil {
			lastErr = NewDIDError(ErrCodeAtalaPRISMError, fmt.Sprintf("HTTP %d: %s", httpResp.StatusCode, string(respBody)), nil)
		} else {
			lastErr = NewDIDError(ErrCodeAtalaPRISMError, errResp.Message, nil)
		}

		// Retry on 5xx errors or specific 4xx errors
		if httpResp.StatusCode >= 500 || httpResp.StatusCode == 429 {
			if attempt < c.maxRetries-1 {
				time.Sleep(c.retryBackoff * time.Duration(1<<uint(attempt)))
				continue
			}
		}

		return nil, lastErr
	}

	if lastErr == nil {
		lastErr = NewDIDError(ErrCodeAtalaPRISMError, "Max retries exceeded", nil)
	}
	return nil, lastErr
}

// Health checks the connectivity to Atala PRISM
func (c *AtalaClient) Health(ctx context.Context) (bool, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/health", nil)
	if err != nil {
		return false, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return false, NewDIDError(ErrCodeAtalaPRISMError, "Failed to parse health response", err)
	}

	status, ok := result["status"].(string)
	if !ok {
		return false, NewDIDError(ErrCodeAtalaPRISMError, "Status not found in health response", nil)
	}

	return strings.ToLower(status) == "healthy", nil
}

// Close closes the client connection pool
func (c *AtalaClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
