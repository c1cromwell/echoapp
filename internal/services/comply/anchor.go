package comply

import (
	"context"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// DataL1Submitter anchors T6 compliance commitments on Echo Data L1.
type DataL1Submitter interface {
	SubmitComplianceAnchor(ctx context.Context, commitment string, epoch int64) (txRef string, err error)
	Enabled() bool
}

// AnchorHealth reports metagraph submission health for WO-252 dashboard.
type AnchorHealth interface {
	Status() string
	LastSuccessAt() *time.Time
}

// MetagraphAnchor submits TrustCommitmentUpdate payloads via the metagraph client.
type MetagraphAnchor struct {
	client  *metagraph.MetagraphClient
	mu      sync.Mutex
	lastOK  *time.Time
	lastErr error
}

func NewMetagraphAnchor(client *metagraph.MetagraphClient) *MetagraphAnchor {
	return &MetagraphAnchor{client: client}
}

func (a *MetagraphAnchor) Enabled() bool {
	return a != nil && a.client != nil
}

func (a *MetagraphAnchor) SubmitComplianceAnchor(ctx context.Context, commitment string, epoch int64) (string, error) {
	if !a.Enabled() {
		return "", nil
	}
	txID, err := a.client.SubmitDataL1(ctx, metagraph.DataL1TrustCommitmentUpdate{
		Commitment: commitment,
		Epoch:      epoch,
	})
	a.mu.Lock()
	defer a.mu.Unlock()
	if err != nil {
		a.lastErr = err
		return "", err
	}
	now := time.Now().UTC()
	a.lastOK = &now
	a.lastErr = nil
	if txID != "" {
		return txID, nil
	}
	return commitment, nil
}

func (a *MetagraphAnchor) Status() string {
	if !a.Enabled() {
		return "healthy"
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.lastOK == nil {
		if a.lastErr != nil {
			return "down"
		}
		return "degraded"
	}
	if time.Since(*a.lastOK) > 15*time.Minute {
		return "degraded"
	}
	return "healthy"
}

func (a *MetagraphAnchor) LastSuccessAt() *time.Time {
	if !a.Enabled() {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.lastOK == nil {
		return nil
	}
	t := *a.lastOK
	return &t
}

// complianceAnchor computes a T6 commitment and optionally submits it to Data L1.
func (s *Service) complianceAnchor(ctx context.Context, orgDID, eventKind, refID, detail string) string {
	commitment := policyAnchorRef(orgDID, database.RetentionPolicyType(eventKind), refID, detail)
	if s.dataL1 == nil || !s.dataL1.Enabled() {
		return commitment
	}
	txRef, err := s.dataL1.SubmitComplianceAnchor(ctx, commitment, time.Now().Unix())
	if err != nil || txRef == "" {
		return commitment
	}
	return txRef
}

func (s *Service) anchorHealthStatus() string {
	if s.dataL1 == nil {
		return "healthy"
	}
	if h, ok := s.dataL1.(AnchorHealth); ok {
		return h.Status()
	}
	if s.dataL1.Enabled() {
		return "degraded"
	}
	return "healthy"
}
