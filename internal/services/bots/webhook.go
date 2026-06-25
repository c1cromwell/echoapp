package bots

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// WebhookRegistry stores per-bot inbound webhook URLs (WO-11 dev slice).
type WebhookRegistry struct {
	mu   sync.RWMutex
	urls map[string]string // botDID -> URL
}

// NewWebhookRegistry creates an empty registry.
func NewWebhookRegistry() *WebhookRegistry {
	return &WebhookRegistry{urls: make(map[string]string)}
}

// Set stores the webhook URL for a bot.
func (r *WebhookRegistry) Set(botDID, url string) {
	if r == nil || botDID == "" || url == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.urls == nil {
		r.urls = make(map[string]string)
	}
	r.urls[botDID] = url
}

// Get returns the registered webhook URL.
func (r *WebhookRegistry) Get(botDID string) (string, bool) {
	if r == nil || botDID == "" {
		return "", false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.urls[botDID]
	return url, ok
}

// InboundPayload is POSTed to a bot developer webhook.
type InboundPayload struct {
	FromDID        string `json:"from_did"`
	ConversationID string `json:"conversation_id"`
	Text           string `json:"text,omitempty"`
	Ciphertext     []byte `json:"ciphertext,omitempty"`
	ReceivedAt     string `json:"received_at"`
}

// DispatchInbound POSTs to the bot webhook (best-effort, short timeout).
func DispatchInbound(registry *WebhookRegistry, botDID string, payload InboundPayload) bool {
	if registry == nil {
		return false
	}
	url, ok := registry.Get(botDID)
	if !ok {
		return false
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bot-DID", botDID)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// IsCatalogBot reports whether did is a known marketplace bot.
func IsCatalogBot(did string) bool {
	_, ok := LookupManifest(did)
	return ok
}
