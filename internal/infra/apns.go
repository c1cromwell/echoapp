package infra

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
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

	jwtMu     sync.Mutex
	jwt       string
	jwtExpiry time.Time
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
// Custom keys sit beside `aps` so the client can reconnect WS; never ciphertext.
type APNsPayload struct {
	APS              APS    `json:"aps"`
	Type             string `json:"type,omitempty"`
	ConversationID   string `json:"conversationId,omitempty"`
	ContentAvailable *int   `json:"-"`
}

// APS is the Apple Push Service alert body.
type APS struct {
	Alert            APSAlert `json:"alert,omitempty"`
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
// Only conversation IDs and a generic wake title are sent — never message content.
func (a *APNsClient) SendPush(ctx context.Context, deviceToken string, conversationID string, notifType string, silent bool) (*APNsResponse, error) {
	var payload APNsPayload
	if silent {
		payload = APNsPayload{
			APS: APS{
				ContentAvailable: 1,
			},
			Type:           notifType,
			ConversationID: conversationID,
		}
	} else {
		payload = APNsPayload{
			APS: APS{
				Alert: APSAlert{
					Title: "Echo",
					Body:  "You have a new message",
				},
				Sound:            "default",
				MutableContent:   1,
				ContentAvailable: 1,
			},
			Type:           notifType,
			ConversationID: conversationID,
		}
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
	if silent {
		req.Header.Set("apns-push-type", "background")
		req.Header.Set("apns-priority", "5")
	} else {
		req.Header.Set("apns-push-type", "alert")
		req.Header.Set("apns-priority", "10")
	}
	req.Header.Set("apns-expiration", fmt.Sprintf("%d", time.Now().Add(24*time.Hour).Unix()))

	if len(a.cfg.PrivateKey) > 0 {
		token, err := a.bearerJWT()
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
		return apnsResp, fmt.Errorf("apns status %d: %s", resp.StatusCode, apnsResp.Reason)
	}

	return apnsResp, nil
}

func (a *APNsClient) endpoint() string {
	if a.cfg.Production {
		return "https://api.push.apple.com"
	}
	return "https://api.sandbox.push.apple.com"
}

func (a *APNsClient) bearerJWT() (string, error) {
	a.jwtMu.Lock()
	defer a.jwtMu.Unlock()
	if a.jwt != "" && time.Now().Before(a.jwtExpiry) {
		return a.jwt, nil
	}
	token, err := a.generateJWT()
	if err != nil {
		return "", err
	}
	a.jwt = token
	a.jwtExpiry = time.Now().Add(50 * time.Minute)
	return token, nil
}

// generateJWT creates an ES256 JWT for APNs authentication from the .p8 key.
func (a *APNsClient) generateJWT() (string, error) {
	key, err := parseAPNsP8(a.cfg.PrivateKey)
	if err != nil {
		return "", err
	}
	headerJSON, _ := json.Marshal(map[string]string{"alg": "ES256", "kid": a.cfg.KeyID})
	now := time.Now().Unix()
	claimsJSON, _ := json.Marshal(map[string]interface{}{"iss": a.cfg.TeamID, "iat": now})
	signingInput := b64URL(headerJSON) + "." + b64URL(claimsJSON)
	sum := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, key, sum[:])
	if err != nil {
		return "", fmt.Errorf("sign apns jwt: %w", err)
	}
	sig := encodeES256Sig(r, s)
	return signingInput + "." + b64URL(sig), nil
}

func parseAPNsP8(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("apns p8: no pem block")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("apns p8: %w", err)
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok || key.Curve != elliptic.P256() {
		return nil, fmt.Errorf("apns p8: expected P-256 ECDSA key")
	}
	return key, nil
}

func encodeES256Sig(r, s *big.Int) []byte {
	out := make([]byte, 64)
	r.FillBytes(out[:32])
	s.FillBytes(out[32:])
	return out
}

func b64URL(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// APNsPushSender adapts APNsClient to notification.PushSender.
type APNsPushSender struct {
	Client *APNsClient
}

func (s APNsPushSender) SendPush(ctx context.Context, deviceToken, conversationID, notifType string, silent bool) error {
	if s.Client == nil {
		return fmt.Errorf("apns client not configured")
	}
	_, err := s.Client.SendPush(ctx, deviceToken, conversationID, notifType, silent)
	return err
}

// NewAPNsPushSenderFromEnv builds a PushSender when APNS_KEY_FILE is set.
// Missing config returns nil so local/dev still queues without talking to Apple.
func NewAPNsPushSenderFromEnv() *APNsPushSender {
	keyFile := strings.TrimSpace(os.Getenv("APNS_KEY_FILE"))
	if keyFile == "" {
		keyFile = strings.TrimSpace(os.Getenv("APNS_KEY_PATH"))
	}
	keyID := strings.TrimSpace(os.Getenv("APNS_KEY_ID"))
	teamID := strings.TrimSpace(os.Getenv("APNS_TEAM_ID"))
	if keyFile == "" || keyID == "" || teamID == "" {
		return nil
	}
	pemBytes, err := os.ReadFile(keyFile)
	if err != nil || len(pemBytes) == 0 {
		return nil
	}
	bundle := strings.TrimSpace(os.Getenv("APNS_BUNDLE_ID"))
	if bundle == "" {
		bundle = "com.echo.app"
	}
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APNS_ENVIRONMENT")))
	production := strings.EqualFold(os.Getenv("APNS_PRODUCTION"), "true") || env == "production"
	client := NewAPNsClient(APNsConfig{
		TeamID:     teamID,
		KeyID:      keyID,
		PrivateKey: pemBytes,
		BundleID:   bundle,
		Production: production,
	})
	return &APNsPushSender{Client: client}
}
