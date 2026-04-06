package trustnet

import (
	"sync"
	"time"
)

// VerificationBadge represents a verified credential
type VerificationBadge struct {
	Type       string // "identity", "did", "phone", "email", "social"
	Label      string // "Identity Verified", "DID Verified", "Phone Verified"
	Detail     string // masked detail e.g. "+1 ******* 4567"
	VerifiedAt time.Time
}

// ContactProfileView is the unified view of a contact's profile
type ContactProfileView struct {
	DID            string
	DisplayName    string
	Username       string
	TrustScore     int
	CircleTier     CircleTier
	Verifications  []VerificationBadge
	VisiblePersonas []*Persona
	MutualCount    int
	SharedGroups   []string
	IsMuted        bool
	IsBlocked      bool
	ContactSince   *time.Time
}

// MuteBlockState tracks mute/block state per contact
type MuteBlockState struct {
	UserDID    string
	ContactDID string
	Muted      bool
	MutedAt    *time.Time
	Blocked    bool
	BlockedAt  *time.Time
	BlockReason string
}

// ContactProfileService provides unified contact profile views
type ContactProfileService struct {
	mu         sync.RWMutex
	circles    *CircleService
	personas   *PersonaService
	discovery  *DiscoveryService
	muteBlock  map[string]*MuteBlockState // "userDID:contactDID" -> state
	verifications map[string][]VerificationBadge // userDID -> badges
	sharedGroups  map[string]map[string][]string // "userDID:contactDID" -> group names
}

// NewContactProfileService creates a new contact profile service
func NewContactProfileService(circles *CircleService, personas *PersonaService, discovery *DiscoveryService) *ContactProfileService {
	return &ContactProfileService{
		circles:       circles,
		personas:      personas,
		discovery:     discovery,
		muteBlock:     make(map[string]*MuteBlockState),
		verifications: make(map[string][]VerificationBadge),
		sharedGroups:  make(map[string]map[string][]string),
	}
}

// GetContactProfile builds a unified contact profile view
func (s *ContactProfileService) GetContactProfile(viewerDID, contactDID string) (*ContactProfileView, error) {
	profile, err := s.discovery.GetProfile(contactDID)
	if err != nil {
		return nil, ErrContactNotFound
	}

	view := &ContactProfileView{
		DID:         contactDID,
		DisplayName: profile.DisplayName,
		Username:    profile.Username,
		TrustScore:  profile.TrustScore,
		CircleTier:  CirclePublic,
	}

	// Get circle tier
	contact, err := s.circles.GetContact(viewerDID, contactDID)
	if err == nil {
		view.CircleTier = contact.Tier
		view.ContactSince = &contact.AddedAt
	}

	// Get visible personas
	if s.personas != nil {
		view.VisiblePersonas = s.personas.GetVisiblePersonas(contactDID, viewerDID)
	}

	// Get mutual contacts
	view.MutualCount = s.circles.CountMutualContacts(viewerDID, contactDID)

	// Get verifications
	s.mu.RLock()
	view.Verifications = s.verifications[contactDID]

	// Get mute/block state
	key := visKey(viewerDID, contactDID)
	if state, ok := s.muteBlock[key]; ok {
		view.IsMuted = state.Muted
		view.IsBlocked = state.Blocked
	}

	// Get shared groups
	if groups, ok := s.sharedGroups[key]; ok {
		for _, g := range groups {
			view.SharedGroups = append(view.SharedGroups, g...)
		}
	}
	s.mu.RUnlock()

	return view, nil
}

// AddVerification adds a verification badge to a user
func (s *ContactProfileService) AddVerification(userDID string, badge VerificationBadge) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.verifications[userDID] = append(s.verifications[userDID], badge)
}

// GetVerifications returns all verification badges for a user
func (s *ContactProfileService) GetVerifications(userDID string) []VerificationBadge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	badges := s.verifications[userDID]
	result := make([]VerificationBadge, len(badges))
	copy(result, badges)
	return result
}

// MuteContact mutes notifications for a contact
func (s *ContactProfileService) MuteContact(userDID, contactDID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := visKey(userDID, contactDID)
	state := s.getOrCreateStateLocked(key, userDID, contactDID)
	state.Muted = true
	now := time.Now()
	state.MutedAt = &now
}

// UnmuteContact unmutes a contact
func (s *ContactProfileService) UnmuteContact(userDID, contactDID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := visKey(userDID, contactDID)
	if state, ok := s.muteBlock[key]; ok {
		state.Muted = false
		state.MutedAt = nil
	}
}

// BlockContact blocks a contact
func (s *ContactProfileService) BlockContact(userDID, contactDID, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := visKey(userDID, contactDID)
	state := s.getOrCreateStateLocked(key, userDID, contactDID)
	state.Blocked = true
	now := time.Now()
	state.BlockedAt = &now
	state.BlockReason = reason
}

// UnblockContact unblocks a contact
func (s *ContactProfileService) UnblockContact(userDID, contactDID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := visKey(userDID, contactDID)
	if state, ok := s.muteBlock[key]; ok {
		state.Blocked = false
		state.BlockedAt = nil
		state.BlockReason = ""
	}
}

// IsBlocked checks if a contact is blocked
func (s *ContactProfileService) IsBlocked(userDID, contactDID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := visKey(userDID, contactDID)
	if state, ok := s.muteBlock[key]; ok {
		return state.Blocked
	}
	return false
}

// IsMuted checks if a contact is muted
func (s *ContactProfileService) IsMuted(userDID, contactDID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := visKey(userDID, contactDID)
	if state, ok := s.muteBlock[key]; ok {
		return state.Muted
	}
	return false
}

// SetSharedGroups sets the shared groups between two users
func (s *ContactProfileService) SetSharedGroups(userDID, contactDID string, groups []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := visKey(userDID, contactDID)
	if s.sharedGroups[key] == nil {
		s.sharedGroups[key] = make(map[string][]string)
	}
	s.sharedGroups[key]["groups"] = groups
}

func (s *ContactProfileService) getOrCreateStateLocked(key, userDID, contactDID string) *MuteBlockState {
	if state, ok := s.muteBlock[key]; ok {
		return state
	}
	state := &MuteBlockState{
		UserDID:    userDID,
		ContactDID: contactDID,
	}
	s.muteBlock[key] = state
	return state
}
