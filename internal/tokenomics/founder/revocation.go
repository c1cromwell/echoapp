// Package founder implements WO-225 founder departure revocation coordination.
package founder

import (
	"errors"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/tokenomics/genesis"
)

var (
	ErrRevocationNotFound       = errors.New("revocation request not found")
	ErrInsufficientSignatures   = errors.New("insufficient founder signatures for revocation")
	ErrInvalidRevocationTarget  = errors.New("target is not a genesis founder")
	ErrRevocationAmountExceeded = errors.New("revocation amount exceeds remaining locked balance")
)

// RevocationRequest tracks 3-of-5 multi-sig founder TokenLock revocation.
type RevocationRequest struct {
	ID               string    `json:"id"`
	TargetFounderDID string    `json:"targetFounderDID"`
	Amount           int64     `json:"amount"`
	Signatures       []string  `json:"signatures"`
	Required         int       `json:"required"`
	CreatedAt        time.Time `json:"createdAt"`
	ExpiresAt        time.Time `json:"expiresAt"`
	Status           string    `json:"status"`
	TxHash           string    `json:"txHash,omitempty"`
	RevokerDIDs      []string  `json:"revokerDIDs,omitempty"`
}

// Event is the on-chain auditable revocation record.
type Event struct {
	TargetFounderDID string   `json:"targetFounderDID"`
	RevokedAmount    int64    `json:"revokedAmount"`
	RevokerDIDs      []string `json:"revokerDIDs"`
	Timestamp        string   `json:"timestamp"`
	TxHash           string   `json:"txHash"`
	DestinationPool  string   `json:"destinationPool"`
}

// Coordinator manages founder revocation multi-sig flow.
type Coordinator struct {
	mu           sync.Mutex
	founderDIDs  []string
	lockedAmount map[string]int64
	requests     map[string]*RevocationRequest
	events       []Event
	submit       func(req RevocationRequest) (string, error)
}

// NewCoordinator creates a revocation coordinator.
func NewCoordinator() *Coordinator {
	dids := genesis.DefaultFounderDIDs()
	locked := make(map[string]int64, len(dids))
	amounts := []int64{genesis.FounderCEOAmount, genesis.FounderMemberAmount, genesis.FounderMemberAmount, genesis.FounderMemberAmount, genesis.FounderMemberAmount}
	for i, did := range dids {
		locked[did] = amounts[i]
	}
	return &Coordinator{
		founderDIDs:  dids,
		lockedAmount: locked,
		requests:     make(map[string]*RevocationRequest),
	}
}

// WithSubmitter wires Currency L1 AtomicAction submission.
func (c *Coordinator) WithSubmitter(fn func(req RevocationRequest) (string, error)) *Coordinator {
	c.submit = fn
	return c
}

// Initiate starts a revocation request (24h signature window).
func (c *Coordinator) Initiate(id, targetDID string, amount int64) (*RevocationRequest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.isFounder(targetDID) {
		return nil, ErrInvalidRevocationTarget
	}
	remaining := c.lockedAmount[targetDID]
	if amount <= 0 || amount > remaining {
		return nil, ErrRevocationAmountExceeded
	}
	now := time.Now().UTC()
	req := &RevocationRequest{
		ID:               id,
		TargetFounderDID: targetDID,
		Amount:           amount,
		Required:         3,
		CreatedAt:        now,
		ExpiresAt:        now.Add(24 * time.Hour),
		Status:           "collecting_signatures",
	}
	c.requests[id] = req
	return req, nil
}

// Sign records a founder signature on a revocation request.
func (c *Coordinator) Sign(id, founderDID string) (*RevocationRequest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	req, ok := c.requests[id]
	if !ok {
		return nil, ErrRevocationNotFound
	}
	if !c.isFounder(founderDID) || founderDID == req.TargetFounderDID {
		return nil, errors.New("invalid revoker DID")
	}
	for _, s := range req.Signatures {
		if s == founderDID {
			return req, nil
		}
	}
	req.Signatures = append(req.Signatures, founderDID)
	return req, nil
}

// Finalize submits when 3-of-5 signatures collected; credits Future Team pool.
func (c *Coordinator) Finalize(id string) (*Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	req, ok := c.requests[id]
	if !ok {
		return nil, ErrRevocationNotFound
	}
	if len(req.Signatures) < req.Required {
		return nil, ErrInsufficientSignatures
	}
	if req.Status == "executed" {
		return c.eventFor(req), nil
	}

	txHash := "revocation_local"
	if c.submit != nil {
		var err error
		txHash, err = c.submit(*req)
		if err != nil {
			return nil, err
		}
	}

	c.lockedAmount[req.TargetFounderDID] -= req.Amount
	req.Status = "executed"
	req.TxHash = txHash
	req.RevokerDIDs = append([]string(nil), req.Signatures...)

	ev := Event{
		TargetFounderDID: req.TargetFounderDID,
		RevokedAmount:    req.Amount,
		RevokerDIDs:      req.RevokerDIDs,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		TxHash:           txHash,
		DestinationPool:  genesis.PoolFutureTeam,
	}
	c.events = append(c.events, ev)
	return &ev, nil
}

// Get returns revocation request status.
func (c *Coordinator) Get(id string) (*RevocationRequest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	req, ok := c.requests[id]
	if !ok {
		return nil, ErrRevocationNotFound
	}
	cp := *req
	return &cp, nil
}

// RemainingLocked returns the target founder's remaining TokenLock balance.
func (c *Coordinator) RemainingLocked(targetDID string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lockedAmount[targetDID]
}

// Events returns executed revocation history for a founder.
func (c *Coordinator) Events(targetDID string) []Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Event, 0)
	for _, e := range c.events {
		if e.TargetFounderDID == targetDID {
			out = append(out, e)
		}
	}
	return out
}

func (c *Coordinator) isFounder(did string) bool {
	for _, f := range c.founderDIDs {
		if f == did {
			return true
		}
	}
	return false
}

func (c *Coordinator) eventFor(req *RevocationRequest) *Event {
	for i := range c.events {
		if c.events[i].TxHash == req.TxHash {
			ev := c.events[i]
			return &ev
		}
	}
	return nil
}
