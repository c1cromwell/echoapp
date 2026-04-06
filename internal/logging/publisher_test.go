package logging

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"
)

func testKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)
	return key
}

func TestNewLogPublisher_Valid(t *testing.T) {
	key := testKey()
	lp, err := NewLogPublisher(key, "2026-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lp.KeyEpoch() != "2026-03" {
		t.Errorf("expected epoch 2026-03, got %s", lp.KeyEpoch())
	}
}

func TestNewLogPublisher_InvalidKeyLength(t *testing.T) {
	_, err := NewLogPublisher([]byte("short"), "2026-03")
	if err == nil {
		t.Fatal("expected error for invalid key length")
	}
}

func TestLogPublisher_AddEntry(t *testing.T) {
	lp, _ := NewLogPublisher(testKey(), "2026-03")

	lp.AddEntry(LogEntry{
		EventType: "relay_batch",
		Count:     42,
	})

	if lp.BufferSize() != 1 {
		t.Errorf("expected buffer size 1, got %d", lp.BufferSize())
	}
}

func TestLogPublisher_AddEntry_AutoTimestamp(t *testing.T) {
	lp, _ := NewLogPublisher(testKey(), "2026-03")

	before := time.Now()
	lp.AddEntry(LogEntry{EventType: "test"})
	after := time.Now()

	batch := lp.Flush()
	ts := batch.Entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Error("auto-timestamp should be roughly now")
	}
}

func TestLogPublisher_Flush(t *testing.T) {
	lp, _ := NewLogPublisher(testKey(), "2026-03")

	lp.AddEntry(LogEntry{EventType: "relay_batch", Count: 10})
	lp.AddEntry(LogEntry{EventType: "reward_claim", Count: 5})
	lp.AddEntry(LogEntry{EventType: "circuit_change", Count: 1})

	batch := lp.Flush()
	if batch == nil {
		t.Fatal("batch should not be nil")
	}
	if batch.EntryCount != 3 {
		t.Errorf("expected 3 entries, got %d", batch.EntryCount)
	}
	if batch.BatchHash == "" {
		t.Error("batch hash should not be empty")
	}
	if batch.TimeRange.From.IsZero() {
		t.Error("time range should be set")
	}
	if lp.BufferSize() != 0 {
		t.Error("buffer should be empty after flush")
	}
}

func TestLogPublisher_FlushEmpty(t *testing.T) {
	lp, _ := NewLogPublisher(testKey(), "2026-03")
	batch := lp.Flush()
	if batch != nil {
		t.Error("flushing empty publisher should return nil")
	}
}

func TestLogPublisher_AutoFlushAtMaxBuffer(t *testing.T) {
	lp, _ := NewLogPublisher(testKey(), "2026-03")

	for i := 0; i < DefaultMaxBuffer; i++ {
		lp.AddEntry(LogEntry{EventType: "test", Count: i})
	}

	if lp.BufferSize() != 0 {
		t.Errorf("expected 0 buffered after auto-flush, got %d", lp.BufferSize())
	}

	completed := lp.CompletedBatches()
	if len(completed) != 1 {
		t.Errorf("expected 1 completed batch, got %d", len(completed))
	}
}

func TestLogPublisher_EncryptDecrypt(t *testing.T) {
	key := testKey()
	lp, _ := NewLogPublisher(key, "2026-03")

	plaintext := []byte(`{"entries":[{"event_type":"relay_batch","count":42}]}`)

	ciphertext, err := lp.EncryptBatch(plaintext)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := lp.DecryptBatch(ciphertext)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("decrypted should match original plaintext")
	}
}

func TestLogPublisher_EncryptDecrypt_DifferentPayloads(t *testing.T) {
	key := testKey()
	lp, _ := NewLogPublisher(key, "2026-03")

	ct1, _ := lp.EncryptBatch([]byte("payload A"))
	ct2, _ := lp.EncryptBatch([]byte("payload B"))

	if bytes.Equal(ct1, ct2) {
		t.Error("different payloads should produce different ciphertext")
	}

	// Same payload should produce different ciphertext (random nonce)
	ct3, _ := lp.EncryptBatch([]byte("payload A"))
	if bytes.Equal(ct1, ct3) {
		t.Error("same payload should produce different ciphertext due to random nonce")
	}
}

func TestLogPublisher_DecryptWithWrongKey(t *testing.T) {
	key1 := testKey()
	key2 := testKey()

	lp1, _ := NewLogPublisher(key1, "2026-03")
	lp2, _ := NewLogPublisher(key2, "2026-03")

	ciphertext, _ := lp1.EncryptBatch([]byte("secret data"))
	_, err := lp2.DecryptBatch(ciphertext)
	if err == nil {
		t.Error("decrypting with wrong key should fail")
	}
}

func TestLogPublisher_DecryptTooShort(t *testing.T) {
	lp, _ := NewLogPublisher(testKey(), "2026-03")
	_, err := lp.DecryptBatch([]byte("short"))
	if err == nil {
		t.Error("decrypting too-short data should fail")
	}
}

func TestLogPublisher_CompletedBatches(t *testing.T) {
	lp, _ := NewLogPublisher(testKey(), "2026-03")

	lp.AddEntry(LogEntry{EventType: "a"})
	lp.Flush()
	lp.AddEntry(LogEntry{EventType: "b"})
	lp.Flush()

	batches := lp.CompletedBatches()
	if len(batches) != 2 {
		t.Errorf("expected 2 completed batches, got %d", len(batches))
	}
}

func TestLogPublisher_EntryMetadata(t *testing.T) {
	lp, _ := NewLogPublisher(testKey(), "2026-03")

	lp.AddEntry(LogEntry{
		EventType: "circuit_change",
		Count:     1,
		Metadata: map[string]string{
			"circuit": "data_l1",
			"state":   "open",
		},
	})

	batch := lp.Flush()
	if batch.Entries[0].Metadata["circuit"] != "data_l1" {
		t.Error("metadata should be preserved")
	}
}
