// Package botsdk is the Echo bot developer SDK (WO-11).
package botsdk

import "time"

// Permission mirrors internal/services/bots.Permission.
type Permission string

const (
	PermSendMessage    Permission = "send_message"
	PermReadMessages   Permission = "read_messages"
	PermRequestPayment Permission = "request_payment"
	PermUploadFile     Permission = "upload_file"
	PermReadChainState Permission = "read_chain_state"
)

// MessageContent is opaque to the relay — bots should send client-encrypted payloads in production.
type MessageContent struct {
	ConversationID string `json:"conversation_id"`
	Text           string `json:"text,omitempty"`
	Ciphertext     []byte `json:"ciphertext,omitempty"`
}

// PaymentRequest is returned when a bot requests an in-chat payment.
type PaymentRequest struct {
	ID        string    `json:"id"`
	UserDID   string    `json:"user_did"`
	Amount    string    `json:"amount"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status,omitempty"`
}

// ChainQuery is a read-only metagraph query stub.
type ChainQuery struct {
	Module string `json:"module"`
	Key    string `json:"key"`
}
