package passport_test

import (
	"context"
	"sync"
	"testing"

	"github.com/thechadcromwell/echoapp/pkg/passport"
)

type memStore struct {
	mu   sync.Mutex
	refs map[string]passport.CredentialRef
}

func (m *memStore) ListCredentialRefs(_ context.Context, holderDID string) ([]passport.CredentialRef, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []passport.CredentialRef
	for _, r := range m.refs {
		if r.HolderDID == holderDID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *memStore) GetCredentialRef(_ context.Context, holderDID, refID string) (*passport.CredentialRef, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.refs[refID]
	if !ok || r.HolderDID != holderDID {
		return nil, nil
	}
	cp := r
	return &cp, nil
}

func (m *memStore) InsertCredentialRef(_ context.Context, ref passport.CredentialRef) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.refs == nil {
		m.refs = make(map[string]passport.CredentialRef)
	}
	for _, r := range m.refs {
		if r.HolderDID == ref.HolderDID && r.CredentialHash == ref.CredentialHash {
			return passport.ErrDuplicateHash
		}
	}
	m.refs[ref.RefID] = ref
	return nil
}

type stubRevocation struct {
	revoked map[int]bool
}

func (s stubRevocation) IsRevoked(_ context.Context, _ string, idx int) (bool, error) {
	return s.revoked[idx], nil
}

func TestRegisterAndListCredentialRef(t *testing.T) {
	store := &memStore{}
	svc := passport.NewService(store, stubRevocation{revoked: map[int]bool{}})
	holder := "did:key:zHolder"

	ref, err := svc.Register(context.Background(), holder, passport.RegisterRefRequest{
		IssuerDID:      "did:key:zIssuer",
		CredentialType: "ProofOfHumanity",
		CredentialHash: "aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if ref.RefID == "" {
		t.Fatal("expected ref_id")
	}

	list, err := svc.List(context.Background(), holder)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(list))
	}
}

func TestRevocationStatusApplied(t *testing.T) {
	store := &memStore{}
	idx := 42
	svc := passport.NewService(store, stubRevocation{revoked: map[int]bool{42: true}})
	holder := "did:key:zHolder"

	ref, err := svc.Register(context.Background(), holder, passport.RegisterRefRequest{
		IssuerDID:       "did:key:zIssuer",
		CredentialType:  "KYCLite",
		CredentialHash:  "1122334411223344112233441122334411223344112233441122334411223344",
		StatusListIndex: &idx,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if ref.RevocationStatus != "revoked" {
		t.Fatalf("expected revoked, got %q", ref.RevocationStatus)
	}
}
