package recovery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Store persists recovery metadata (never share bytes).
type Store interface {
	UpsertPolicy(ctx context.Context, policy Policy) error
	GetPolicy(ctx context.Context, holderDID string) (*Policy, error)
	ReplaceShareholders(ctx context.Context, holderDID string, shareholders []Shareholder) error
	ListShareholders(ctx context.Context, holderDID string) ([]Shareholder, error)
	InsertSession(ctx context.Context, session Session) error
	GetSession(ctx context.Context, holderDID, sessionID string) (*Session, error)
	CompleteSession(ctx context.Context, holderDID, sessionID, commitment string, completedAt time.Time) error
}

// Service coordinates recovery metadata and client-side Shamir helpers.
type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Setup(ctx context.Context, holderDID string, req SetupRequest) (*Policy, []Shareholder, error) {
	if err := ValidatePolicy(req.Threshold, req.Total); err != nil {
		return nil, nil, err
	}
	if len(req.Shareholders) != req.Total {
		return nil, nil, fmt.Errorf("%w: expected %d shareholders, got %d", ErrInvalidShareholder, req.Total, len(req.Shareholders))
	}

	now := time.Now().UTC()
	seen := make(map[int]struct{}, req.Total)
	shareholders := make([]Shareholder, 0, len(req.Shareholders))
	for _, in := range req.Shareholders {
		if err := ValidateShareholderInput(in); err != nil {
			return nil, nil, err
		}
		if _, dup := seen[in.ShareIndex]; dup {
			return nil, nil, fmt.Errorf("%w: duplicate share_index %d", ErrInvalidShareholder, in.ShareIndex)
		}
		seen[in.ShareIndex] = struct{}{}
		status := in.Status
		if status == "" {
			status = StatusPending
		}
		shareholders = append(shareholders, Shareholder{
			ShareID:      uuid.NewString(),
			HolderDID:    holderDID,
			ShareIndex:   in.ShareIndex,
			GuardianDID:  strings.TrimSpace(in.GuardianDID),
			Role:         in.Role,
			Status:       status,
			GuardianVCID: strings.TrimSpace(in.GuardianVCID),
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	policy := Policy{
		HolderDID: holderDID,
		Threshold: req.Threshold,
		Total:     req.Total,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.UpsertPolicy(ctx, policy); err != nil {
		return nil, nil, err
	}
	if err := s.store.ReplaceShareholders(ctx, holderDID, shareholders); err != nil {
		return nil, nil, err
	}
	return &policy, shareholders, nil
}

func (s *Service) GetStatus(ctx context.Context, holderDID string) (*Policy, []Shareholder, error) {
	policy, err := s.store.GetPolicy(ctx, holderDID)
	if err != nil {
		return nil, nil, err
	}
	if policy == nil {
		return nil, nil, ErrNotConfigured
	}
	shareholders, err := s.store.ListShareholders(ctx, holderDID)
	if err != nil {
		return nil, nil, err
	}
	return policy, shareholders, nil
}

func (s *Service) Initiate(ctx context.Context, holderDID string) (*InitiateResponse, error) {
	policy, shareholders, err := s.GetStatus(ctx, holderDID)
	if err != nil {
		return nil, err
	}
	active := filterShareholders(shareholders, StatusActive)
	if len(active) < policy.Threshold {
		return nil, fmt.Errorf("%w: need at least %d active guardians", ErrNotConfigured, policy.Threshold)
	}

	now := time.Now().UTC()
	session := Session{
		SessionID:      uuid.NewString(),
		HolderDID:      holderDID,
		Status:         SessionInitiated,
		RequiredShares: policy.Threshold,
		ExpiresAt:      now.Add(30 * time.Minute),
		CreatedAt:      now,
	}
	if err := s.store.InsertSession(ctx, session); err != nil {
		return nil, err
	}
	return &InitiateResponse{Session: session, Shareholders: active}, nil
}

func (s *Service) Complete(ctx context.Context, holderDID string, req CompleteRequest) (*Session, error) {
	req.RootKeyCommitment = strings.ToLower(strings.TrimSpace(req.RootKeyCommitment))
	if len(req.RootKeyCommitment) != 64 {
		return nil, fmt.Errorf("root_key_commitment must be 64-char hex SHA-256")
	}
	if _, err := hex.DecodeString(req.RootKeyCommitment); err != nil {
		return nil, fmt.Errorf("invalid root_key_commitment hex: %w", err)
	}

	session, err := s.store.GetSession(ctx, holderDID, req.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	if session.Status != SessionInitiated {
		return nil, fmt.Errorf("session is not active")
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	now := time.Now().UTC()
	if err := s.store.CompleteSession(ctx, holderDID, req.SessionID, req.RootKeyCommitment, now); err != nil {
		return nil, err
	}
	session.Status = SessionCompleted
	session.RootKeyCommitment = req.RootKeyCommitment
	session.CompletedAt = &now
	return session, nil
}

// RootKeyCommitment returns hex(SHA-256(passportRootKey)) for completion attestation.
func RootKeyCommitment(passportRootKey []byte) string {
	sum := sha256.Sum256(passportRootKey)
	return hex.EncodeToString(sum[:])
}

func filterShareholders(all []Shareholder, wantStatus string) []Shareholder {
	var out []Shareholder
	for _, sh := range all {
		if sh.Status == wantStatus {
			out = append(out, sh)
		}
	}
	return out
}

// ErrIsNotConfigured reports whether err is ErrNotConfigured.
func ErrIsNotConfigured(err error) bool {
	return errors.Is(err, ErrNotConfigured)
}
