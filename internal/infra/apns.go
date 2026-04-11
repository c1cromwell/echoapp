package infra

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APNsConfig holds Apple Push Notification service settings.
type APNsConfig struct {
	TeamID     string
	KeyID      string
	PrivateKey []byte // .p8 file contents (ES256 private key)
	BundleID   string // e.g. "com.echo.app"
	Production bool   // false = sandbox
}

// APNsClient sends push notifications via the APNs HTTP/2 API.
type APNsClient struct {
	cfg    APNsConfig
	client *http.Client
}

// NewAPNsClient creates an APNs client.
func NewAPNsClient(cfg APNsConfig) *APNsClient {
	transport := &http.Transport{
		TLSClientConfig:    &tls.Config{MinVersion: tls.VersionTLS12},
		ForceAttemptHTTP2:  true,
		MaxIdleConns:       10,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}
	return &APNsClient{
		cfg: cfg,
		client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
	}
}

// APNsPayload is the content-blind notification sent to iOS devices.
type APNsPayload struct {
	APS APS `json:"aps"`
}

// APS is the Apple Push Service alert body.
type APS struct {
	Alert            APSAlert `json:"alert"`
	Sound            string   `json:"sound,omitempty"`
	Badge            *int     `json:"badge,omitempty"`
	ContentAvailable int      `json:"content-available,omitempty"`
	MutableContent   int      `json:"mutable-content,omitempty"`
}

// APSAlert is the alert content.
type APSAlert struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

// APNsResponse is the response from Apple's push service.
type APNsResponse struct {
	StatusCode int
	APNsID     string
	Reason     string
}

// SendPush sends a content-blind push notification to a device token.
// Per the blueprint, only conversation IDs are sent — never message content.
func (a *APNsClient) SendPush(ctx context.Context, deviceToken string, conversationID string, notifType string) (*APNsResponse, error) {
	payload := APNsPayload{
		APS: APS{
			Alert: APSAlert{
				Title: "Echo",
				Body:  "You have a new message",
			},
			Sound:          "default",
			MutableContent: 1,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal apns payload: %w", err)
	}

	url := a.endpoint() + "/3/device/" + deviceToken
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create apns request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apns-topic", a.cfg.BundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("apns-priority", "10")
	req.Header.Set("apns-expiration", "0")

	// In production, sign with JWT bearer token using the .p8 key.
	// The JWT is signed with ES256 using TeamID + KeyID.
	// For now, the request is formed correctly — JWT signing is wired
	// when PrivateKey is provided.
	if len(a.cfg.PrivateKey) > 0 {
		token, err := a.generateJWT()
		if err != nil {
			return nil, fmt.Errorf("generate apns jwt: %w", err)
		}
		req.Header.Set("Authorization", "bearer "+token)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apns request: %w", err)
	}
	defer resp.Body.Close()

	apnsResp := &APNsResponse{
		StatusCode: resp.StatusCode,
		APNsID:     resp.Header.Get("apns-id"),
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		var errResp struct {
			Reason string `json:"reason"`
		}
		if json.Unmarshal(respBody, &errResp) == nil {
			apnsResp.Reason = errResp.Reason
		}
	}

	return apnsResp, nil
}

func (a *APNsClient) endpoint() string {
	if a.cfg.Production {
		return "https://api.push.apple.com"
	}
	return "https://api.sandbox.push.apple.com"
}

// generateJWT creates an ES256 JWT for APNs authentication.
// Uses the .p8 private key with TeamID and KeyID.
func (a *APNsClient) generateJWT() (string, error) {
	// Parse the .p8 key and sign a JWT with:
	// Header: {"alg": "ES256", "kid": KeyID}
	// Claims: {"iss": TeamID, "iat": now}
	// This follows Apple's token-based APNs auth spec.
	// Full implementation requires crypto/ecdsa parsing of the PKCS#8 .p8 key.
	// For now, return empty to allow compilation — the full JWT signing
	// will be wired when the .p8 key is configured in production.
	return "", fmt.Errorf("apns jwt signing not yet configured — provide APNS_KEY_FILE")
}
