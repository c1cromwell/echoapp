package evidence

import (
	"context"
	"fmt"
	"time"
)

// EvidenceSubmitter abstracts the Digital Evidence API HTTP client.
type EvidenceSubmitter interface {
	SubmitFingerprint(ctx context.Context, req *FingerprintRequest) (*FingerprintResponse, error)
	VerifyFingerprint(ctx context.Context, eventID string) (*VerificationResult, error)
}

// VerificationResult is the on-chain verification status for a fingerprint.
type VerificationResult struct {
	EventID     string `json:"eventId"`
	Status      string `json:"status"` // "verified", "pending", "failed"
	ExplorerURL string `json:"explorerUrl"`
	VerifiedAt  string `json:"verifiedAt,omitempty"`
}

// EvidenceService orchestrates Digital Evidence fingerprint operations.
type EvidenceService struct {
	client EvidenceSubmitter
	config *ClientConfig
}

// NewEvidenceService creates a new evidence orchestration service.
func NewEvidenceService(client EvidenceSubmitter, config *ClientConfig) *EvidenceService {
	return &EvidenceService{
		client: client,
		config: config,
	}
}

// MediaFingerprintReq is the input for fingerprinting media content.
type MediaFingerprintReq struct {
	ContentHash string `json:"contentHash"`
	MessageID   string `json:"messageId"`
	SenderDID   string `json:"senderDid"`
}

// AuditBatchReq is the input for fingerprinting an IPFS audit log batch.
type AuditBatchReq struct {
	BatchHash  string    `json:"batchHash"`
	IPFSCID    string    `json:"ipfsCid"`
	EntryCount int       `json:"entryCount"`
	TimeFrom   time.Time `json:"timeFrom"`
	TimeTo     time.Time `json:"timeTo"`
}

// FingerprintMedia creates a SHA-256 fingerprint for media content (VIP+ users).
func (s *EvidenceService) FingerprintMedia(ctx context.Context, req MediaFingerprintReq) (*FingerprintResponse, error) {
	fpReq, err := MediaFingerprint(s.config, req.ContentHash, req.SenderDID)
	if err != nil {
		return nil, err
	}
	fpReq.Metadata.BatchID = req.MessageID

	return s.client.SubmitFingerprint(ctx, fpReq)
}

// FingerprintAuditBatch creates a fingerprint for an IPFS audit log batch (Org tier).
func (s *EvidenceService) FingerprintAuditBatch(ctx context.Context, req AuditBatchReq) (*FingerprintResponse, error) {
	fpReq, err := AuditBatchFingerprint(s.config, req.IPFSCID, req.BatchHash, "echo_platform")
	if err != nil {
		return nil, err
	}
	fpReq.Metadata.Description = fmt.Sprintf(
		"Audit batch: %d entries from %s to %s",
		req.EntryCount,
		req.TimeFrom.Format(time.RFC3339),
		req.TimeTo.Format(time.RFC3339),
	)

	return s.client.SubmitFingerprint(ctx, fpReq)
}

// VerifyFingerprintStatus checks on-chain verification status for an event ID.
func (s *EvidenceService) VerifyFingerprintStatus(ctx context.Context, eventID string) (*VerificationResult, error) {
	return s.client.VerifyFingerprint(ctx, eventID)
}
