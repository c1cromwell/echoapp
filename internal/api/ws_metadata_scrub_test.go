package api

import (
	"encoding/json"
	"testing"
)

func TestScrubRelayMetadataForQueue_SealedText(t *testing.T) {
	raw := []byte(`{"type":"sealed_text","from":"did:key:alice","to":"did:key:bob","payload":{}}`)
	out := scrubRelayMetadataForQueue(raw)
	if string(out) == string(raw) {
		t.Fatal("expected scrubbed output")
	}
	var msg WSMessage
	if err := json.Unmarshal(out, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.From != "" {
		t.Fatalf("from should be stripped, got %q", msg.From)
	}
}

func TestScrubRelayMetadataForQueue_PlainTextUnchanged(t *testing.T) {
	raw := []byte(`{"type":"text","from":"did:key:alice","to":"did:key:bob","payload":{}}`)
	out := scrubRelayMetadataForQueue(raw)
	if string(out) != string(raw) {
		t.Fatal("plain text should be unchanged")
	}
}
