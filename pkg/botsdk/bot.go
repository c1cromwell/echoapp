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

// OnMessage registers an inbound handler (local stub — webhook transport is out of scope for M3).
func (b *EchoBot) OnMessage(handler MessageHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handler = handler
	return nil
}

// RequestPayment triggers a user authorization prompt (stub).
func (b *EchoBot) RequestPayment(_ context.Context, userDID, amount, reason string) (*PaymentRequest, error) {
	return nil, newBotError(CodeInvalidRequest, "payment relay not implemented in this SDK slice", 0)
}

// UploadFile stores an opaque blob (stub).
func (b *EchoBot) UploadFile(_ context.Context, _ []byte, _ string) (string, error) {
	return "", newBotError(CodeInvalidRequest, "file relay not implemented in this SDK slice", 0)
}

// ReadChainState queries public metagraph state (stub).
func (b *EchoBot) ReadChainState(_ context.Context, _ ChainQuery) (any, error) {
	return nil, newBotError(CodeInvalidRequest, "chain read not implemented in this SDK slice", 0)
}
