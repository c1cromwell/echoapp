package passport

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("credential ref not found")
	ErrDuplicateHash = errors.New("credential already registered for holder")
)

// RefStore persists holder credential references.
type RefStore interface {
	ListCredentialRefs(ctx context.Context, holderDID string) ([]CredentialRef, error)
	GetCredentialRef(ctx context.Context, holderDID, refID string) (*CredentialRef, error)
	InsertCredentialRef(ctx context.Context, ref CredentialRef) error
}

// RevocationChecker resolves StatusList2021 slot state for a credential id/hash.
type RevocationChecker interface {
	IsRevoked(ctx context.Context, issuerDID string, statusListIndex int) (bool, error)
}

// Service aggregates holder credential refs with live revocation status.
type Service struct {
	store      RefStore
	revocation RevocationChecker
}

func NewService(store RefStore, revocation RevocationChecker) *Service {
	return &Service{store: store, revocation: revocation}
}

func (s *Service) List(ctx context.Context, holderDID string) ([]CredentialRef, error) {
	refs, err := s.store.ListCredentialRefs(ctx, holderDID)
	if err != nil {
		return nil, err
	}
	for i := range refs {
		s.applyRevocation(ctx, &refs[i])
	}
	return refs, nil
}

func (s *Service) Get(ctx context.Context, holderDID, refID string) (*CredentialRef, error) {
	ref, err := s.store.GetCredentialRef(ctx, holderDID, refID)
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, ErrNotFound
	}
	s.applyRevocation(ctx, ref)
	return ref, nil
}

func (s *Service) Register(ctx context.Context, holderDID string, req RegisterRefRequest) (*CredentialRef, error) {
	req.IssuerDID = strings.TrimSpace(req.IssuerDID)
	req.CredentialType = strings.TrimSpace(req.CredentialType)
	req.CredentialHash = strings.ToLower(strings.TrimSpace(req.CredentialHash))
	if req.IssuerDID == "" || req.CredentialType == "" || len(req.CredentialHash) != 64 {
		return nil, fmt.Errorf("issuer_did, credential_type, and 64-char credential_hash are required")
	}

	now := time.Now().UTC()
	ref := CredentialRef{
		RefID:           uuid.NewString(),
		HolderDID:       holderDID,
		IssuerDID:       req.IssuerDID,
		CredentialType:  req.CredentialType,
		CredentialHash:  req.CredentialHash,
		StatusListIndex: req.StatusListIndex,
		StatusListCred:  strings.TrimSpace(req.StatusListCred),
		RevocationStatus: "unknown",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.store.InsertCredentialRef(ctx, ref); err != nil {
		if errors.Is(err, ErrDuplicateHash) {
			return nil, ErrDuplicateHash
		}
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, ErrDuplicateHash
		}
		return nil, err
	}
	s.applyRevocation(ctx, &ref)
	return &ref, nil
}

func (s *Service) applyRevocation(ctx context.Context, ref *CredentialRef) {
	ref.RevocationStatus = "unknown"
	if ref.StatusListIndex == nil || s.revocation == nil {
		return
	}
	revoked, err := s.revocation.IsRevoked(ctx, ref.IssuerDID, *ref.StatusListIndex)
	if err != nil {
		return
	}
	if revoked {
		ref.RevocationStatus = "revoked"
	} else {
		ref.RevocationStatus = "active"
	}
}
