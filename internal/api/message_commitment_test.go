package api

import (
	"testing"
)

func TestCommitmentFromWSMessage(t *testing.T) {
	msg := WSMessage{
		Type:    "text",
		Payload: []byte(`{"message_id":"m1","text":"hi"}`),
	}
	id, hash, ok := commitmentFromWSMessage(msg)
	if !ok || id != "m1" || len(hash) != 32 {
		t.Fatalf("unexpected: id=%q hash=%d ok=%v", id, len(hash), ok)
	}

	withHex := WSMessage{
		Type:    "text",
		Payload: []byte(`{"message_id":"m2","commitment_hash":"abababababababababababababababababababababababababababababababab"}`),
	}
	id2, _, ok2 := commitmentFromWSMessage(withHex)
	if !ok2 || id2 != "m2" {
		t.Fatalf("expected m2, got %q", id2)
	}
}
