package trustnet

import (
	"sync"
	"time"
)

// CircleTier represents the relationship tier
type CircleTier string

const (
	CircleInner       CircleTier = "inner"
	CircleTrusted     CircleTier = "trusted"
	CircleAcquaintance CircleTier = "acquaintance"
	// CirclePublic is the implicit tier for non-contacts (strangers)
	CirclePublic CircleTier = "public"
)

// CircleLimits defines max contacts per tier
var CircleLimits = map[CircleTier]int{
	CircleInner:       15,
	CircleTrusted:     50,
	CircleAcquaintance: 200,
	CirclePublic:      -1, // unlimited (implicit, non-contacts)
}

// CirclePermissions defines what each tier can access per spec
type CirclePermissions struct {
	// Persona visibility
	SeeAllPersonas      bool
	SeeSelectedPersonas bool
	SeeDefaultPersona   bool

	// Status
	SeeOnlineStatus bool
	SeeReadReceipts bool

	// Communication
	DirectCall  bool
	CallRequest bool
	AddToGroups bool // "Yes", "Ask" handled by AskToAddToGroups
	AskToAddToGroups bool

	// Profile visibility
	FullProfile     bool
	StandardProfile bool
	MinimalProfile  bool

	// Messages
	FilteredMessages   bool
	UnfilteredMessages bool
}

// DefaultPermissions returns the permissions for each tier per spec
var DefaultPermissions = map[CircleTier]CirclePermissions{
	CircleInner: {
		SeeAllPersonas: true, SeeSelectedPersonas: true, SeeDefaultPersona: true,
		SeeOnlineStatus: true, SeeReadReceipts: true,
		DirectCall: true, CallRequest: true,
		AddToGroups: true,
		FullProfile: true, StandardProfile: true, MinimalProfile: true,
		UnfilteredMessages: true,
	},
	CircleTrusted: {
		SeeSelectedPersonas: true, SeeDefaultPersona: true,
		SeeOnlineStatus: true, SeeReadReceipts: true,
		CallRequest: true,
		AskToAddToGroups: true,
		StandardProfile: true, MinimalProfile: true,
		UnfilteredMessages: true,
	},
	CircleAcquaintance: {
		SeeDefaultPersona: true,
		MinimalProfile:    true,
		FilteredMessages:  true,
	},
	CirclePublic: {
		MinimalProfile:   true,
		FilteredMessages: true,
	},
}

// AutoPromotionRules defines criteria for automatic circle promotion
type AutoPromotionRules struct {
	MessagesExchanged int
	CallCount         int
	DaysKnown         int
}

// DefaultAutoPromotion returns auto-promotion rules per tier
var DefaultAutoPromotion = map[CircleTier]AutoPromotionRules{
	CircleTrusted: {MessagesExchanged: 50, CallCount: 5, DaysKnown: 30},
	CircleInner:   {MessagesExchanged: 200, CallCount: 20, DaysKnown: 90},
}

// CircleContact represents a contact within a circle
type CircleContact struct {
	UserDID           string
	ContactDID        string
	Tier              CircleTier
	AddedAt           time.Time
	PromotedAt        *time.Time
	MessagesExchanged int
	CallCount         int
	LastInteraction   *time.Time
}

// CircleService manages trust circle relationships
type CircleService struct {
	mu       sync.RWMutex
	contacts map[string]map[string]*CircleContact // userDID -> contactDID -> contact
}

// NewCircleService creates a new circle service
func NewCircleService() *CircleService {
	return &CircleService{
		contacts: make(map[string]map[string]*CircleContact),
	}
}

// AddContact adds a contact to a specific circle tier
func (s *CircleService) AddContact(userDID, contactDID string, tier CircleTier) (*CircleContact, error) {
	if userDID == contactDID {
		return nil, ErrCannotAddSelf
	}

	if !isValidTier(tier) {
		return nil, ErrInvalidCircleTier
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.contacts[userDID] == nil {
		s.contacts[userDID] = make(map[string]*CircleContact)
	}

	// Check if already exists
	if existing, ok := s.contacts[userDID][contactDID]; ok {
		if existing.Tier == tier {
			return nil, ErrContactAlreadyInCircle
		}
		// Move to new tier (handled as promotion/demotion)
		existing.Tier = tier
		now := time.Now()
		existing.PromotedAt = &now
		return existing, nil
	}

	// Check circle capacity
	count := s.countTierLocked(userDID, tier)
	limit := CircleLimits[tier]
	if limit >= 0 && count >= limit {
		return nil, ErrCircleFull
	}

	contact := &CircleContact{
		UserDID:    userDID,
		ContactDID: contactDID,
		Tier:       tier,
		AddedAt:    time.Now(),
	}

	s.contacts[userDID][contactDID] = contact
	return contact, nil
}

// PromoteContact moves a contact to a higher trust tier
func (s *CircleService) PromoteContact(userDID, contactDID string, newTier CircleTier) (*CircleContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	contact, ok := s.contacts[userDID][contactDID]
	if !ok {
		return nil, ErrContactNotFound
	}

	if !isValidTier(newTier) {
		return nil, ErrInvalidCircleTier
	}

	// Check capacity of target tier
	count := s.countTierLocked(userDID, newTier)
	limit := CircleLimits[newTier]
	if limit >= 0 && count >= limit {
		return nil, ErrCircleFull
	}

	contact.Tier = newTier
	now := time.Now()
	contact.PromotedAt = &now
	return contact, nil
}

// DemoteContact moves a contact to a lower trust tier
func (s *CircleService) DemoteContact(userDID, contactDID string, newTier CircleTier) (*CircleContact, error) {
	return s.PromoteContact(userDID, contactDID, newTier)
}

// RemoveContact removes a contact from all circles
func (s *CircleService) RemoveContact(userDID, contactDID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.contacts[userDID] == nil {
		return ErrContactNotFound
	}
	if _, ok := s.contacts[userDID][contactDID]; !ok {
		return ErrContactNotFound
	}

	delete(s.contacts[userDID], contactDID)
	return nil
}

// GetContact retrieves a specific contact relationship
func (s *CircleService) GetContact(userDID, contactDID string) (*CircleContact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.contacts[userDID] == nil {
		return nil, ErrContactNotFound
	}
	contact, ok := s.contacts[userDID][contactDID]
	if !ok {
		return nil, ErrContactNotFound
	}
	return contact, nil
}

// GetCircle returns all contacts in a specific tier
func (s *CircleService) GetCircle(userDID string, tier CircleTier) []*CircleContact {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var contacts []*CircleContact
	for _, c := range s.contacts[userDID] {
		if c.Tier == tier {
			contacts = append(contacts, c)
		}
	}
	return contacts
}

// GetAllContacts returns all contacts across all tiers
func (s *CircleService) GetAllContacts(userDID string) []*CircleContact {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var contacts []*CircleContact
	for _, c := range s.contacts[userDID] {
		contacts = append(contacts, c)
	}
	return contacts
}

// GetPermissions returns the permissions for a contact
func (s *CircleService) GetPermissions(userDID, contactDID string) CirclePermissions {
	s.mu.RLock()
	defer s.mu.RUnlock()

	contact, ok := s.contacts[userDID][contactDID]
	if !ok {
		return DefaultPermissions[CirclePublic]
	}
	return DefaultPermissions[contact.Tier]
}

// HasContact checks if a user has a specific contact
func (s *CircleService) HasContact(userDID, contactDID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.contacts[userDID] == nil {
		return false
	}
	_, ok := s.contacts[userDID][contactDID]
	return ok
}

// GetMutualContacts returns contacts that both users share
func (s *CircleService) GetMutualContacts(userDID, otherDID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userContacts := s.contacts[userDID]
	otherContacts := s.contacts[otherDID]
	if userContacts == nil || otherContacts == nil {
		return nil
	}

	var mutual []string
	for did := range userContacts {
		if _, ok := otherContacts[did]; ok {
			mutual = append(mutual, did)
		}
	}
	return mutual
}

// CountMutualContacts returns the number of mutual contacts
func (s *CircleService) CountMutualContacts(userDID, otherDID string) int {
	return len(s.GetMutualContacts(userDID, otherDID))
}

// RecordInteraction updates interaction stats for auto-promotion evaluation
func (s *CircleService) RecordInteraction(userDID, contactDID string, messages int, calls int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	contact, ok := s.contacts[userDID][contactDID]
	if !ok {
		return
	}

	contact.MessagesExchanged += messages
	contact.CallCount += calls
	now := time.Now()
	contact.LastInteraction = &now
}

// CheckAutoPromotion evaluates if a contact qualifies for auto-promotion
func (s *CircleService) CheckAutoPromotion(userDID, contactDID string) (*CircleTier, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	contact, ok := s.contacts[userDID][contactDID]
	if !ok {
		return nil, ErrContactNotFound
	}

	daysKnown := int(time.Since(contact.AddedAt).Hours() / 24)

	// Check promotion to inner from trusted
	if contact.Tier == CircleTrusted {
		rules := DefaultAutoPromotion[CircleInner]
		if contact.MessagesExchanged >= rules.MessagesExchanged &&
			contact.CallCount >= rules.CallCount &&
			daysKnown >= rules.DaysKnown {
			tier := CircleInner
			return &tier, nil
		}
	}

	// Check promotion to trusted from acquaintance
	if contact.Tier == CircleAcquaintance {
		rules := DefaultAutoPromotion[CircleTrusted]
		if contact.MessagesExchanged >= rules.MessagesExchanged &&
			contact.CallCount >= rules.CallCount &&
			daysKnown >= rules.DaysKnown {
			tier := CircleTrusted
			return &tier, nil
		}
	}

	return nil, nil
}

// CountCircle returns the number of contacts in a tier
func (s *CircleService) CountCircle(userDID string, tier CircleTier) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.countTierLocked(userDID, tier)
}

func (s *CircleService) countTierLocked(userDID string, tier CircleTier) int {
	count := 0
	for _, c := range s.contacts[userDID] {
		if c.Tier == tier {
			count++
		}
	}
	return count
}

func isValidTier(tier CircleTier) bool {
	switch tier {
	case CircleInner, CircleTrusted, CircleAcquaintance, CirclePublic:
		return true
	}
	return false
}
