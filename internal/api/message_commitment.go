package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// wsTextWire is the client JSON inside WS `text` payloads (subset for commitment extraction).
type wsTextWire struct {
	MessageID      string          `json:"message_id"`
	CommitmentHash string          `json:"commitment_hash,omitempty"`
	Text           string          `json:"text,omitempty"`
	Encrypted      json.RawMessage `json:"encrypted,omitempty"`
}

// commitmentFromWSMessage extracts a message id + SHA-256 commitment from a relayed text blob.
// When the client omits commitment_hash, the relay hashes the opaque payload bytes (content-blind).
func commitmentFromWSMessage(msg WSMessage) (messageID string, hash []byte, ok bool) {
	if msg.Type != "text" && msg.Type != "sealed_text" {
		return "", nil, false
	}
	if len(msg.Payload) == 0 {
		return "", nil, false
	}
	var wire wsTextWire
	if err := json.Unmarshal(msg.Payload, &wire); err != nil {
		sum := sha256.Sum256(msg.Payload)
		return "", sum[:], true
	}
	if wire.MessageID == "" {
		sum := sha256.Sum256(msg.Payload)
		return "", sum[:], true
	}
	if wire.CommitmentHash != "" {
		decoded, err := hex.DecodeString(wire.CommitmentHash)
		if err != nil || len(decoded) != sha256.Size {
			sum := sha256.Sum256(msg.Payload)
			return wire.MessageID, sum[:], true
		}
		return wire.MessageID, decoded, true
	}
	sum := sha256.Sum256(msg.Payload)
	return wire.MessageID, sum[:], true
}
