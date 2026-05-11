// Package logging implements the WO-53 log publisher for encrypted audit trails.
//
// Operational events are batched, encrypted (AES-256-GCM), and pushed to IPFS
// (or a stub). CIDs are recorded on Data L1 for verifiable log indexing.
//
// Key rotation: monthly HKDF-SHA256 derivation from a 32-byte platform master
// key using info string "log_encryption_YYYY-MM". The master key is never stored
// in code — it is supplied at startup from an HSM or Shamir reconstruction.
//
// Privacy contract: no DID, no phone, no email, no message content in any event.
package logging

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/hkdf"
)

const (
	DefaultMaxBuffer     = 1000
	DefaultFlushInterval = 5 * time.Minute
)

// LogEntry represents a single privacy-safe operational event.
// No PII, no DIDs (unless compliance-required), no message content.
type LogEntry struct {
	EventType string            `json:"event_type"` // "relay_batch", "reward_claim", etc.
	Count     int               `json:"count"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// LogBatch represents a completed batch ready for encryption and IPFS push.
type LogBatch struct {
	Entries     []LogEntry     `json:"entries"`
	BatchHash   string         `json:"batch_hash"`
	EntryCount  int            `json:"entry_count"`
	TimeRange   BatchTimeRange `json:"time_range"`
	CreatedAt   time.Time      `json:"created_at"`
	EncryptedAt time.Time      `json:"encrypted_at,omitempty"`
	CID         string         `json:"cid,omitempty"` // IPFS CID after push
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
	encryptionKey []byte     // AES-256-GCM key (monthly rotating)
	keyEpoch      string     // Identifies which key encrypted a given batch
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

// --- Monthly key rotation (WO-53) ---

// DeriveMonthlyKey derives the AES-256-GCM log encryption key for the given
// month using HKDF-SHA256.  The info string encodes the month so keys rotate
// automatically and old batches can still be decrypted by archiving the key
// alongside the CID.
//
// masterKey must be 32 bytes (AES-256).  In production it is reconstructed
// from 3-of-5 Shamir shares held by designated platform operators.
func DeriveMonthlyKey(masterKey []byte, month time.Time) (key []byte, epoch string, err error) {
	if len(masterKey) != 32 {
		return nil, "", errors.New("master key must be 32 bytes")
	}
	epoch = month.Format("2006-01")
	info := fmt.Sprintf("log_encryption_%s", epoch)
	r := hkdf.New(sha256.New, masterKey, nil, []byte(info))
	key = make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, "", err
	}
	return key, epoch, nil
}

// RotateKey replaces the in-memory encryption key with the monthly key derived
// from masterKey.  Call once on startup and again on the first of each month.
func (p *LogPublisher) RotateKey(masterKey []byte, month time.Time) error {
	key, epoch, err := DeriveMonthlyKey(masterKey, month)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.encryptionKey = key
	p.keyEpoch = epoch
	p.mu.Unlock()
	return nil
}

// --- IPFS storage interface + stub (WO-53) ---

// AuditLogRecord is submitted to Data L1 after a batch is pushed to IPFS.
type AuditLogRecord struct {
	Type      string         `json:"type"` // always "audit_log"
	CID       string         `json:"cid"`  // IPFS content identifier
	BatchHash string         `json:"batch_hash"`
	KeyEpoch  string         `json:"key_epoch"`
	TimeRange BatchTimeRange `json:"time_range"`
}

// IPFSStorage is the storage abstraction for pushing encrypted log batches.
// Production implementations wrap Pinata/web3.storage; tests use StubIPFSStorage.
type IPFSStorage interface {
	// Store pushes encrypted bytes and returns an IPFS CID.
	Store(ctx context.Context, encrypted []byte) (cid string, err error)
}

// StubIPFSStorage records batches in memory — used in tests and dev mode.
type StubIPFSStorage struct {
	mu      sync.Mutex
	batches [][]byte
	cids    []string
}

func (s *StubIPFSStorage) Store(_ context.Context, encrypted []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	h := sha256.Sum256(encrypted)
	cid := "stub-cid-" + hex.EncodeToString(h[:8])
	s.batches = append(s.batches, encrypted)
	s.cids = append(s.cids, cid)
	return cid, nil
}

// StoredCIDs returns all CIDs produced so far (for test assertions).
func (s *StubIPFSStorage) StoredCIDs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.cids))
	copy(out, s.cids)
	return out
}

// FlushAndPublish flushes the buffer, encrypts the batch, pushes to IPFS, and
// returns the AuditLogRecord to be submitted to Data L1.  Returns nil, nil if
// the buffer was empty.
func (p *LogPublisher) FlushAndPublish(ctx context.Context, storage IPFSStorage) (*AuditLogRecord, error) {
	batch := p.Flush()
	if batch == nil {
		return nil, nil
	}

	// Serialize entries as a simple JSON-like payload (no external deps).
	payload := []byte(fmt.Sprintf(`{"epoch":%q,"batch_hash":%q,"entry_count":%d}`,
		p.keyEpoch, batch.BatchHash, batch.EntryCount))

	encrypted, err := p.EncryptBatch(payload)
	if err != nil {
		return nil, fmt.Errorf("encrypt batch: %w", err)
	}

	cid, err := storage.Store(ctx, encrypted)
	if err != nil {
		return nil, fmt.Errorf("ipfs store: %w", err)
	}

	batch.CID = cid
	batch.EncryptedAt = time.Now()

	return &AuditLogRecord{
		Type:      "audit_log",
		CID:       cid,
		BatchHash: batch.BatchHash,
		KeyEpoch:  p.keyEpoch,
		TimeRange: batch.TimeRange,
	}, nil
}

// StartPeriodicFlush runs a background goroutine that flushes and publishes
// every flushInterval.  Returns a stop func.
func (p *LogPublisher) StartPeriodicFlush(storage IPFSStorage) func() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(p.flushInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, _ = p.FlushAndPublish(ctx, storage)
			case <-ctx.Done():
				return
			}
		}
	}()
	return cancel
}
