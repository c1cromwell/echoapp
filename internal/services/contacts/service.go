// Package contacts implements the privacy-preserving contact discovery service.
package contacts

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"

	"github.com/thechadcromwell/echoapp/internal/database"
)

var (
	ErrSelfContact     = errors.New("cannot add yourself as a contact")
	ErrTier1Limit      = errors.New("tier 1 accounts limited to 10 contacts")
	ErrAlreadyBlocked  = errors.New("contact is already blocked")
	ErrInvalidInvite   = errors.New("invalid or expired invite code")
	ErrRateLimited     = errors.New("PSI discovery rate limited")
	ErrOPRFUnavailable = errors.New("contact discovery is not configured")

	maxBlindedPerRequest = 1000 // cap blinded elements per discovery request
)

const (
	tier1ContactLimit = 10
	inviteReward      = 50_00000000 // 50 ECHO
	argon2Time        = 1
	argon2Memory      = 64 * 1024
	argon2Threads     = 4
	argon2KeyLen      = 32
)

// Service provides contact management operations.
type Service struct {
	db database.DB

	// oprf powers private (OPRF-PSI) contact discovery. The index
	// (hex(OPRF_k(phone)) -> DID) is persisted via the durable
	// database.ContactDiscoveryStore; raw phone numbers are never stored.
	// Nil oprf disables discovery.
	oprf *OPRFService
}

// NewService creates a contacts service.
func NewService(db database.DB) *Service {
	return &Service{db: db}
}

// SetOPRF wires the OPRF engine used for private contact discovery.
func (s *Service) SetOPRF(o *OPRFService) { s.oprf = o }

// DiscoveryKey computes the OPRF index key hex(OPRF_k(phone)) for a raw number.
// It is pseudorandom and safe to hold transiently (e.g. in an OTP session) until
// the number is confirmed; the raw number itself is never stored.
func (s *Service) DiscoveryKey(e164 string) (string, error) {
	if s.oprf == nil {
		return "", ErrOPRFUnavailable
	}
	return s.oprf.IndexKey(e164)
}

// CommitDiscoveryKey records a confirmed OPRF index key -> DID binding in the
// durable store when the user is discoverable per tier/opt-in policy (WO-220).
func (s *Service) CommitDiscoveryKey(ctx context.Context, key, did string) error {
	if key == "" {
		return nil
	}
	user, err := s.db.GetUserByDID(ctx, did)
	if err != nil {
		return nil
	}
	if !IsPhoneDiscoverable(user.TrustTier, user.PhoneDiscoveryOptIn) {
		return nil
	}
	return s.db.PutDiscoveryKey(ctx, key, did)
}

// RegisterPhoneForDiscovery computes and commits the discovery binding in one
// step (raw number used transiently, never persisted). Convenience for callers
// that already hold a confirmed number.
func (s *Service) RegisterPhoneForDiscovery(ctx context.Context, did, e164 string) error {
	key, err := s.DiscoveryKey(e164)
	if err != nil {
		return err
	}
	return s.CommitDiscoveryKey(ctx, key, did)
}

// OPRFEvaluate runs the oblivious server step over client-blinded phone numbers.
// The server learns nothing about the numbers — only blinded group elements.
func (s *Service) OPRFEvaluate(blindedB64 []string) ([]string, error) {
	if s.oprf == nil {
		return nil, ErrOPRFUnavailable
	}
	if len(blindedB64) > maxBlindedPerRequest {
		return nil, ErrRateLimited
	}
	return s.oprf.Evaluate(blindedB64)
}

// DiscoveryIndex returns discoverable {hex(OPRF_k(phone)) -> DID} entries only.
func (s *Service) DiscoveryIndex(ctx context.Context) (map[string]string, error) {
	raw, err := s.db.AllDiscoveryKeys(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make(map[string]string, len(raw))
	for key, did := range raw {
		user, err := s.db.GetUserByDID(ctx, did)
		if err != nil {
			continue
		}
		if IsPhoneDiscoverable(user.TrustTier, user.PhoneDiscoveryOptIn) {
			filtered[key] = did
		}
	}
	return filtered, nil
}

// SetPhoneDiscoveryOptIn updates the user's PSI discoverability preference.
// When disabled, existing index entries for the DID are removed.
func (s *Service) SetPhoneDiscoveryOptIn(ctx context.Context, did string, optIn bool) error {
	val := optIn
	if err := s.db.UpdatePhoneDiscoveryOptIn(ctx, did, &val); err != nil {
		return err
	}
	if !optIn {
		return s.db.DeleteDiscoveryKeysForDID(ctx, did)
	}
	return nil
}

// GetPhoneDiscoverySettings returns the effective discoverability state for a user.
func (s *Service) GetPhoneDiscoverySettings(ctx context.Context, did string) (map[string]interface{}, error) {
	user, err := s.db.GetUserByDID(ctx, did)
	if err != nil {
		return nil, err
	}
	discoverable := IsPhoneDiscoverable(user.TrustTier, user.PhoneDiscoveryOptIn)
	var explicit *bool
	if user.PhoneDiscoveryOptIn != nil {
		v := *user.PhoneDiscoveryOptIn
		explicit = &v
	}
	return map[string]interface{}{
		"did":                       did,
		"trust_tier":                user.TrustTier,
		"phone_discovery_opt_in":    explicit,
		"phone_discoverable":        discoverable,
		"tier_default_discoverable": user.TrustTier >= minTierForDefaultPhoneDiscovery,
	}, nil
}

// SearchByUsername searches for users by handle.
func (s *Service) SearchByUsername(ctx context.Context, callerDID, handle string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	user, err := s.db.GetUserByUsername(ctx, handle)
	if err == nil && user.DID != callerDID {
		results = append(results, map[string]interface{}{
			"did":      user.DID,
			"username": user.Username,
			"tier":     user.TrustTier,
		})
	}
	if results == nil {
		results = make([]map[string]interface{}, 0)
	}
	return results, nil
}

// AddContact adds a contact to the caller's contact list.
func (s *Service) AddContact(ctx context.Context, callerDID, contactDID, addedVia string) (*database.Contact, error) {
	if callerDID == contactDID {
		return nil, ErrSelfContact
	}

	// Check tier 1 limit
	count, err := s.db.GetContactCount(ctx, callerDID)
	if err != nil {
		return nil, err
	}

	caller, _ := s.db.GetUserByDID(ctx, callerDID)
	if caller != nil && caller.TrustTier <= 1 && count >= tier1ContactLimit {
		return nil, ErrTier1Limit
	}

	contact := &database.Contact{
		OwnerDID:   callerDID,
		ContactDID: contactDID,
		AddedVia:   addedVia,
	}

	// Get trust badge for the contact
	ts, err := s.db.GetTrustScore(ctx, contactDID)
	if err == nil {
		contact.TrustBadge = tierBadge(ts.Tier)
	}

	if err := s.db.AddContact(ctx, contact); err != nil {
		return nil, err
	}
	return contact, nil
}

// GetContacts returns all contacts for the caller.
func (s *Service) GetContacts(ctx context.Context, callerDID string) ([]*database.Contact, error) {
	contacts, err := s.db.GetContacts(ctx, callerDID)
	if err != nil {
		return nil, err
	}

	// Refresh trust badges
	for _, c := range contacts {
		ts, err := s.db.GetTrustScore(ctx, c.ContactDID)
		if err == nil {
			c.TrustBadge = tierBadge(ts.Tier)
		}
	}
	return contacts, nil
}

// BlockContact blocks a contact.
func (s *Service) BlockContact(ctx context.Context, callerDID, contactDID string) error {
	return s.db.SetBlocked(ctx, callerDID, contactDID, true)
}

// UnblockContact unblocks a contact.
func (s *Service) UnblockContact(ctx context.Context, callerDID, contactDID string) error {
	return s.db.SetBlocked(ctx, callerDID, contactDID, false)
}

// RemoveContact removes a contact.
func (s *Service) RemoveContact(ctx context.Context, callerDID, contactDID string) error {
	return s.db.RemoveContact(ctx, callerDID, contactDID)
}

// CreateInviteLink creates a new invite link for the caller.
func (s *Service) CreateInviteLink(ctx context.Context, callerDID string) (*database.InviteLink, error) {
	invite := &database.InviteLink{
		Code:       uuid.New().String()[:8],
		CreatorDID: callerDID,
	}
	if err := s.db.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}
	return invite, nil
}

// VerifyInvite checks if an invite code is valid.
func (s *Service) VerifyInvite(ctx context.Context, code string) (*database.InviteLink, error) {
	invite, err := s.db.GetInvite(ctx, code)
	if err != nil {
		return nil, ErrInvalidInvite
	}
	return invite, nil
}

// AcceptInvite accepts an invite and creates a bidirectional contact relationship.
func (s *Service) AcceptInvite(ctx context.Context, code, accepterDID string) (*database.InviteLink, error) {
	invite, err := s.db.GetInvite(ctx, code)
	if err != nil {
		return nil, ErrInvalidInvite
	}

	if invite.CreatorDID == accepterDID {
		return nil, ErrSelfContact
	}

	if err := s.db.AcceptInvite(ctx, code, accepterDID); err != nil {
		return nil, err
	}

	// Add bidirectional contacts
	s.db.AddContact(ctx, &database.Contact{
		OwnerDID:   invite.CreatorDID,
		ContactDID: accepterDID,
		AddedVia:   "invite",
	})
	s.db.AddContact(ctx, &database.Contact{
		OwnerDID:   accepterDID,
		ContactDID: invite.CreatorDID,
		AddedVia:   "invite",
	})

	invite.AcceptedBy = accepterDID
	invite.Accepted = true
	return invite, nil
}

// HashPhone hashes a phone number using Argon2id for PSI.
func HashPhone(phone string) []byte {
	normalized := normalizePhone(phone)
	salt := []byte("echo-psi-v1")
	return argon2.IDKey([]byte(normalized), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
}

func normalizePhone(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	return phone
}

func tierBadge(tier int) string {
	switch tier {
	case 5:
		return "verified_plus"
	case 4:
		return "verified"
	case 3:
		return "trusted"
	case 2:
		return "basic"
	default:
		return "new"
	}
}
