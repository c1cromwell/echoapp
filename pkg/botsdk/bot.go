package botsdk

import (
	"context"
	"fmt"
	"sync"
)

// MessageHandler receives inbound bot messages (SDK webhook stub).
type MessageHandler func(ctx context.Context, fromDID string, content MessageContent) error

// EchoBot is the developer-facing bot runtime (WO-11).
type EchoBot struct {
	BotDID      string
	Permissions []Permission
	client      *relayClient

	mu      sync.Mutex
	handler MessageHandler
}

// New builds an EchoBot from configuration.
func New(cfg Config) (*EchoBot, error) {
	if cfg.BotDID == "" || cfg.APIToken == "" {
		return nil, fmt.Errorf("botsdk: BotDID and APIToken are required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8000"
	}
	return &EchoBot{
		BotDID: cfg.BotDID,
		client: newRelayClient(cfg),
	}, nil
}

// SendMessage delivers a message to recipientDID via the Echo relay.
func (b *EchoBot) SendMessage(ctx context.Context, recipientDID string, content MessageContent) error {
	if recipientDID == "" || content.ConversationID == "" {
		return newBotError(CodeInvalidRequest, "recipient_did and conversation_id required", 0)
	}
	if content.Text == "" && len(content.Ciphertext) == 0 {
		return newBotError(CodeInvalidRequest, "text or ciphertext required", 0)
	}
	_, err := b.client.sendMessage(ctx, recipientDID, content)
	return err
}

// OnMessage registers an inbound handler (local; pair with RegisterWebhook).
func (b *EchoBot) OnMessage(handler MessageHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handler = handler
	return nil
}

// RegisterWebhook registers the bot's inbound message webhook URL with Echo.
func (b *EchoBot) RegisterWebhook(ctx context.Context, url string) error {
	if url == "" {
		return newBotError(CodeInvalidRequest, "url required", 0)
	}
	return b.client.registerWebhook(ctx, url)
}

// RequestPayment triggers a user authorization prompt via the relay.
func (b *EchoBot) RequestPayment(ctx context.Context, userDID, amount, reason string) (*PaymentRequest, error) {
	if userDID == "" || amount == "" {
		return nil, newBotError(CodeInvalidRequest, "user_did and amount required", 0)
	}
	return b.client.requestPayment(ctx, userDID, amount, reason)
}

// UploadFile stores an opaque ciphertext blob via the media relay.
func (b *EchoBot) UploadFile(ctx context.Context, userDID string, data []byte, mimeType string) (string, error) {
	if userDID == "" || len(data) == 0 {
		return "", newBotError(CodeInvalidRequest, "user_did and data required", 0)
	}
	return b.client.uploadFile(ctx, userDID, data, mimeType)
}

// ReadChainState queries public metagraph state (stub when L1 unavailable).
func (b *EchoBot) ReadChainState(ctx context.Context, userDID string, query ChainQuery) (any, error) {
	if userDID == "" {
		return nil, newBotError(CodeInvalidRequest, "user_did required", 0)
	}
	return b.client.readChainState(ctx, userDID, query)
}
