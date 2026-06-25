package contacts

import (
	"context"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// IsBlocked reports whether ownerDID has blocked contactDID.
func (s *Service) IsBlocked(ctx context.Context, ownerDID, contactDID string) (bool, error) {
	return s.db.IsContactBlocked(ctx, ownerDID, contactDID)
}

// IsEitherBlocked is true if either party has blocked the other (WO-190 relay/search).
func (s *Service) IsEitherBlocked(ctx context.Context, a, b string) (bool, error) {
	if a == "" || b == "" || a == b {
		return false, nil
	}
	left, err := s.IsBlocked(ctx, a, b)
	if err != nil {
		return false, err
	}
	if left {
		return true, nil
	}
	return s.IsBlocked(ctx, b, a)
}

// GetBlockedContacts returns contacts with blocked status for the owner.
func (s *Service) GetBlockedContacts(ctx context.Context, ownerDID string) ([]*database.Contact, error) {
	all, err := s.db.GetContacts(ctx, ownerDID)
	if err != nil {
		return nil, err
	}
	var blocked []*database.Contact
	for _, c := range all {
		if c.Blocked {
			blocked = append(blocked, c)
		}
	}
	if blocked == nil {
		blocked = []*database.Contact{}
	}
	return blocked, nil
}

// BlockContact blocks a contact, creating a block row if needed (WO-190).
func (s *Service) BlockContact(ctx context.Context, callerDID, contactDID string) error {
	if callerDID == contactDID {
		return ErrSelfContact
	}
	return s.db.SetBlocked(ctx, callerDID, contactDID, true)
}

// UnblockContact restores an active contact relationship.
func (s *Service) UnblockContact(ctx context.Context, callerDID, contactDID string) error {
	if callerDID == contactDID {
		return ErrSelfContact
	}
	return s.db.SetBlocked(ctx, callerDID, contactDID, false)
}
