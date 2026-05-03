package credentials

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const statusListBitCount = 131072

// StatusList2021BatchUpdate matches Scala StatusList2021BatchUpdate (Identity L1 POST /transactions).
type statusList2021BatchWire struct {
	IssuerOrgDID string `json:"issuerOrgDID"`
	BitVector    string `json:"bitVector"`
	PublishedAt  int64  `json:"publishedAt"`
	Sequence     int64  `json:"sequence"`
}

// StatusListPublisher maintains an in-memory StatusList2021 bit vector. Revocations mark the
// vector dirty and enqueue an asynchronous L1 publish (coalesced + retried); a periodic ticker
// retries while dirty so failed pushes are not lost.
type StatusListPublisher struct {
	mu        sync.Mutex
	vec       []byte
	credToIdx map[string]int
	nextIdx   int
	seq       int64
	dirty     bool
	cfg       MetagraphConfig
	client    *http.Client
	stop      chan struct{}
	// publishTrigger coalesces: at most one pending wakeup (hybrid async publish).
	publishTrigger chan struct{}
	wg             sync.WaitGroup
	stopOnce       sync.Once

	statsMu              sync.RWMutex
	lastPublishSuccessAt time.Time
	lastPublishErr       string
}

// NewStatusListPublisher returns a publisher, or nil when anchoring/L1 is disabled.
func NewStatusListPublisher(cfg MetagraphConfig) *StatusListPublisher {
	if !cfg.EnableAnchor || cfg.IdentityL1URL == "" {
		return nil
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &StatusListPublisher{
		vec:            make([]byte, statusListBitCount/8),
		credToIdx:      make(map[string]int),
		client:         &http.Client{Timeout: timeout},
		cfg:            cfg,
		stop:           make(chan struct{}),
		publishTrigger: make(chan struct{}, 1),
	}
}

// Start begins a single worker: reacts to revoke-driven signals and a periodic dirty check.
func (p *StatusListPublisher) Start() {
	if p == nil {
		return
	}
	interval := p.cfg.StatusListPublishInterval
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-p.stop:
				return
			case <-p.publishTrigger:
				p.runPublishWithRetry()
			case <-t.C:
				if p.PublishPending() {
					p.runPublishWithRetry()
				}
			}
		}
	}()
}

func (p *StatusListPublisher) signalPublish() {
	select {
	case p.publishTrigger <- struct{}{}:
	default:
	}
}

// Stop shuts down the publisher goroutine.
func (p *StatusListPublisher) Stop() {
	if p == nil {
		return
	}
	p.stopOnce.Do(func() { close(p.stop) })
	p.wg.Wait()
}

// AllocateIndex assigns the next status list index for a stored credential ID (internal UUID).
func (p *StatusListPublisher) AllocateIndex(credentialStorageID string) int {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if idx, ok := p.credToIdx[credentialStorageID]; ok {
		return idx
	}
	if p.nextIdx >= statusListBitCount {
		p.nextIdx = statusListBitCount - 1
	}
	idx := p.nextIdx
	p.nextIdx++
	p.credToIdx[credentialStorageID] = idx
	return idx
}

// MarkRevoked sets the revocation bit and enqueues an asynchronous L1 publish. Returns false
// if this credential had no allocated status-list index.
func (p *StatusListPublisher) MarkRevoked(credentialStorageID string) bool {
	if p == nil {
		return false
	}
	p.mu.Lock()
	idx, ok := p.credToIdx[credentialStorageID]
	if !ok {
		p.mu.Unlock()
		return false
	}
	byteIdx := idx / 8
	bit := idx % 8
	if byteIdx < len(p.vec) {
		p.vec[byteIdx] |= 1 << bit
	}
	p.dirty = true
	p.mu.Unlock()
	p.signalPublish()
	return true
}

// PublishPending is true when the local bit vector has not yet been successfully pushed to L1
// after the last change (or a push is still catching up).
func (p *StatusListPublisher) PublishPending() bool {
	if p == nil {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.dirty
}

// Stats returns lightweight observability for health endpoints.
func (p *StatusListPublisher) Stats() map[string]interface{} {
	if p == nil {
		return nil
	}
	p.statsMu.RLock()
	lastErr := p.lastPublishErr
	lastOK := p.lastPublishSuccessAt
	p.statsMu.RUnlock()
	return map[string]interface{}{
		"publish_pending":           p.PublishPending(),
		"last_publish_success_unix": lastOK.Unix(),
		"last_publish_error":        lastErr,
	}
}

func (p *StatusListPublisher) runPublishWithRetry() {
	if p == nil || !p.PublishPending() {
		return
	}
	max := int(p.cfg.MaxRetries)
	if max <= 0 {
		max = 5
	}
	backoff := p.cfg.RetryBackoff
	if backoff <= 0 {
		backoff = time.Second
	}
	var lastErr error
	for attempt := 0; attempt < max; attempt++ {
		if attempt > 0 {
			d := backoff << uint(attempt-1)
			const maxSleep = 30 * time.Second
			if d > maxSleep {
				d = maxSleep
			}
			time.Sleep(d)
		}
		ctx, cancel := context.WithTimeout(context.Background(), p.cfg.Timeout)
		lastErr = p.Publish(ctx)
		cancel()
		if lastErr == nil {
			return
		}
		fmt.Printf("Warning: StatusList2021 publish attempt %d/%d: %v\n", attempt+1, max, lastErr)
	}
	if lastErr != nil {
		p.statsMu.Lock()
		p.lastPublishErr = lastErr.Error()
		p.statsMu.Unlock()
	}
}

// Publish posts one snapshot of the bit vector to Identity L1. On HTTP success, bumps sequence
// and clears dirty if no concurrent revocation modified the vector during the round trip.
func (p *StatusListPublisher) Publish(ctx context.Context) error {
	if p == nil {
		return nil
	}

	p.mu.Lock()
	if !p.dirty {
		// Nothing to push — avoid empty heartbeat writes
		p.mu.Unlock()
		return nil
	}
	vecCopy := append([]byte(nil), p.vec...)
	seq := p.seq
	issuer := p.cfg.IssuerDID
	p.mu.Unlock()

	body := statusList2021BatchWire{
		IssuerOrgDID: issuer,
		BitVector:    base64.StdEncoding.EncodeToString(vecCopy),
		PublishedAt:  time.Now().UnixMilli(),
		Sequence:     seq,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := p.cfg.IdentityL1URL + "/transactions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		p.statsMu.Lock()
		p.lastPublishErr = err.Error()
		p.statsMu.Unlock()
		return fmt.Errorf("identity L1 status list POST: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf("identity L1 status list: status %d: %s", resp.StatusCode, string(respBody))
		p.statsMu.Lock()
		p.lastPublishErr = err.Error()
		p.statsMu.Unlock()
		return err
	}

	needFollowUp := false
	p.mu.Lock()
	p.seq++
	if bytes.Equal(p.vec, vecCopy) {
		p.dirty = false
	} else {
		p.dirty = true
		needFollowUp = true
	}
	p.mu.Unlock()

	p.statsMu.Lock()
	p.lastPublishSuccessAt = time.Now()
	p.lastPublishErr = ""
	p.statsMu.Unlock()

	if needFollowUp {
		p.signalPublish()
	}
	return nil
}
