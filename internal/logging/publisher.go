// Package logging implements the IPFS log publisher for encrypted audit trails.
//
// Operational events are batched, compressed (zstd), encrypted (AES-256-GCM),
// and pushed to IPFS. CIDs are recorded on Data L1 for verifiable log indexing.
package logging

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"
)

const (
	DefaultMaxBuffer     = 1000
	DefaultFlushInterval = 5 * time.Minute
)

// LogEntry represents a single privacy-safe operational event.
// No PII, no DIDs (unless compliance-required), no message content.
type LogEntry struct {
	EventType string    `json:"event_type"` // "relay_batch", "reward_claim", etc.
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// LogBatch represents a completed batch ready for encryption and IPFS push.
type LogBatch struct {
	Entries     []LogEntry `json:"entries"`
	BatchHash   string     `json:"batch_hash"`
	EntryCount  int        `json:"entry_count"`
	TimeRange   BatchTimeRange `json:"time_range"`
	CreatedAt   time.Time  `json:"created_at"`
	EncryptedAt time.Time  `json:"encrypted_at,omitempty"`
	CID         string     `json:"cid,omitempty"` // IPFS CID after push
}

// BatchTimeRange represents temporal bounds of a log batch.
type BatchTimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// LogPublisher batches operational events, encrypts them, and prepares for IPFS push.
type LogPublisher struct {
	mu            sync.Mutex
	buffer        []LogEntry
	maxBuffer     int
	flushInterval time.Duration
	encryptionKey []byte // AES-256-GCM key (monthly rotating)
	keyEpoch      string // Identifies which key encrypted a given batch
	batches       []LogBatch // completed batches
}

// NewLogPublisher creates a new log publisher with the given encryption key.
func NewLogPublisher(encryptionKey []byte, keyEpoch string) (*LogPublisher, error) {
	if len(encryptionKey) != 32 {
		return nil, errors.New("encryption key must be 32 bytes (AES-256)")
	}
	return &LogPublisher{
		maxBuffer:     DefaultMaxBuffer,
		flushInterval: DefaultFlushInterval,
		encryptionKey: encryptionKey,
		keyEpoch:      keyEpoch,
	}, nil
}

// AddEntry adds an operational event to the buffer.
// Triggers a flush if the buffer reaches max capacity.
func (p *LogPublisher) AddEntry(entry LogEntry) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	p.buffer = append(p.buffer, entry)

	if len(p.buffer) >= p.maxBuffer {
		p.flushLocked()
	}
}

// Flush manually triggers a batch flush.
func (p *LogPublisher) Flush() *LogBatch {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.flushLocked()
}

func (p *LogPublisher) flushLocked() *LogBatch {
	if len(p.buffer) == 0 {
		return nil
	}

	entries := p.buffer
	p.buffer = nil

	// Compute batch hash for integrity
	hashData := ""
	for _, e := range entries {
		hashData += e.EventType + e.Timestamp.String()
	}
	h := sha256.Sum256([]byte(hashData))

	batch := LogBatch{
		Entries:    entries,
		BatchHash:  hex.EncodeToString(h[:]),
		EntryCount: len(entries),
		TimeRange: BatchTimeRange{
			From: entries[0].Timestamp,
			To:   entries[len(entries)-1].Timestamp,
		},
		CreatedAt: time.Now(),
	}

	p.batches = append(p.batches, batch)
	return &batch
}

// EncryptBatch encrypts a serialized batch payload using AES-256-GCM.
// Returns ciphertext with prepended nonce.
func (p *LogPublisher) EncryptBatch(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptBatch decrypts a batch encrypted with EncryptBatch.
func (p *LogPublisher) DecryptBatch(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	data := ciphertext[nonceSize:]

	return gcm.Open(nil, nonce, data, nil)
}

// BufferSize returns the current number of buffered entries.
func (p *LogPublisher) BufferSize() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.buffer)
}

// CompletedBatches returns all flushed batches.
func (p *LogPublisher) CompletedBatches() []LogBatch {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]LogBatch, len(p.batches))
	copy(result, p.batches)
	return result
}

// KeyEpoch returns the current encryption key epoch identifier.
func (p *LogPublisher) KeyEpoch() string {
	return p.keyEpoch
}
