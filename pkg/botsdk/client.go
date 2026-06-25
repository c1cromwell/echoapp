package botsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config holds bot credentials and API base URL.
type Config struct {
	BaseURL  string
	BotDID   string
	APIToken string
	HTTP     *http.Client
}

// relayClient performs authenticated calls to Echo bot relay endpoints.
type relayClient struct {
	baseURL string
	botDID  string
	token   string
	http    *http.Client
}

func newRelayClient(cfg Config) *relayClient {
	client := cfg.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	base := strings.TrimRight(cfg.BaseURL, "/")
	return &relayClient{
		baseURL: base,
		botDID:  cfg.BotDID,
		token:   cfg.APIToken,
		http:    client,
	}
}

type relayMessageRequest struct {
	RecipientDID   string `json:"recipient_did"`
	ConversationID string `json:"conversation_id"`
	Text           string `json:"text,omitempty"`
	Ciphertext     []byte `json:"ciphertext,omitempty"`
}

type relayMessageResponse struct {
	Delivered bool `json:"delivered"`
}

func (c *relayClient) postJSON(ctx context.Context, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bot-DID", c.botDID)
	req.Header.Set("X-Bot-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == http.StatusTooManyRequests {
		return newBotError(CodeRateLimitExceeded, string(raw), 60)
	}
	if resp.StatusCode == http.StatusForbidden {
		return newBotError(CodePermissionDenied, string(raw), 0)
	}
	if resp.StatusCode >= 400 {
		return newBotError(CodeRelayFailed, fmt.Sprintf("http %d: %s", resp.StatusCode, string(raw)), 0)
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return err
		}
	}
	return nil
}

func (c *relayClient) sendMessage(ctx context.Context, recipientDID string, content MessageContent) (bool, error) {
	var resp relayMessageResponse
	err := c.postJSON(ctx, "/v3/bots/relay/message", relayMessageRequest{
		RecipientDID:   recipientDID,
		ConversationID: content.ConversationID,
		Text:           content.Text,
		Ciphertext:     content.Ciphertext,
	}, &resp)
	if err != nil {
		return false, err
	}
	return resp.Delivered, nil
}

func (c *relayClient) registerWebhook(ctx context.Context, url string) error {
	return c.postJSON(ctx, "/v3/bots/relay/webhook", map[string]string{"url": url}, nil)
}

func (c *relayClient) requestPayment(ctx context.Context, userDID, amount, reason string) (*PaymentRequest, error) {
	var resp struct {
		Payment PaymentRequest `json:"payment"`
	}
	err := c.postJSON(ctx, "/v3/bots/relay/payment", map[string]string{
		"user_did": userDID,
		"amount":   amount,
		"reason":   reason,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Payment, nil
}

func (c *relayClient) uploadFile(ctx context.Context, userDID string, ciphertext []byte, mimeType string) (string, error) {
	var resp struct {
		MediaID string `json:"media_id"`
	}
	err := c.postJSON(ctx, "/v3/bots/relay/upload", map[string]any{
		"user_did":   userDID,
		"ciphertext": ciphertext,
		"mime_type":  mimeType,
	}, &resp)
	if err != nil {
		return "", err
	}
	return resp.MediaID, nil
}

func (c *relayClient) readChainState(ctx context.Context, userDID string, query ChainQuery) (any, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/v3/bots/relay/chain?user_did=%s&module=%s&key=%s",
			c.baseURL, userDID, query.Module, query.Key),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Bot-DID", c.botDID)
	req.Header.Set("X-Bot-Token", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 400 {
		return nil, newBotError(CodeRelayFailed, string(raw), 0)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out["state"], nil
}
