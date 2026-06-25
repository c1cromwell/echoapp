package contacts

import (
	"context"
	"strings"
	"sync"
)

// VisibilityLevel controls who can see a profile field (WO-190).
type VisibilityLevel string

const (
	VisibilityEveryone VisibilityLevel = "everyone"
	VisibilityContacts VisibilityLevel = "contacts"
	VisibilityNobody   VisibilityLevel = "nobody"
)

// PrivacySettings is the global profile privacy model (WO-187 / WO-190).
type PrivacySettings struct {
	ShowLastSeen       VisibilityLevel `json:"showLastSeen"`
	ShowOnlineStatus   VisibilityLevel `json:"showOnlineStatus"`
	ShowProfilePicture VisibilityLevel `json:"showProfilePicture"`
	ShowStatusMessage  VisibilityLevel `json:"showStatusMessage"`
	AllowGroupInvites  VisibilityLevel `json:"allowGroupInvites"`
	AllowCalls         VisibilityLevel `json:"allowCalls"`
	ShowTrustScore     VisibilityLevel `json:"showTrustScore"`
}

// ContactPrivacyOverride per-contact notification / disappearing prefs (WO-39).
type ContactPrivacyOverride struct {
	NotificationsEnabled *bool `json:"notificationsEnabled,omitempty"`
	DisappearingEnabled  *bool `json:"disappearingEnabled,omitempty"`
}

func defaultPrivacySettings() PrivacySettings {
	return PrivacySettings{
		ShowLastSeen:       VisibilityEveryone,
		ShowOnlineStatus:   VisibilityContacts,
		ShowProfilePicture: VisibilityEveryone,
		ShowStatusMessage:  VisibilityContacts,
		AllowGroupInvites:  VisibilityContacts,
		AllowCalls:         VisibilityContacts,
		ShowTrustScore:     VisibilityContacts,
	}
}

type profileRecord struct {
	DisplayName   string
	Bio           string
	StatusMessage string
	AvatarURL     string
	Privacy       PrivacySettings
	Overrides     map[string]ContactPrivacyOverride
}

// UserProfile is the privacy-filtered profile view returned to clients.
type UserProfile struct {
	DID           string `json:"did"`
	DisplayName   string `json:"displayName,omitempty"`
	Username      string `json:"username,omitempty"`
	AvatarURL     string `json:"avatarURL,omitempty"`
	Bio           string `json:"bio,omitempty"`
	StatusMessage string `json:"statusMessage,omitempty"`
	TrustTier     int    `json:"trustTier"`
	IsVerified    bool   `json:"isVerified"`
	IsContact     bool   `json:"isContact"`
	IsBlocked     bool   `json:"isBlocked"`
}

func (s *Service) profileMu() *sync.RWMutex {
	if s.profileMuPtr == nil {
		s.profileMuPtr = &sync.RWMutex{}
	}
	return s.profileMuPtr
}

func (s *Service) ensureProfile(did string) *profileRecord {
	s.profileMu().Lock()
	defer s.profileMu().Unlock()
	if s.profiles == nil {
		s.profiles = make(map[string]*profileRecord)
	}
	rec, ok := s.profiles[did]
	if !ok {
		rec = &profileRecord{Privacy: defaultPrivacySettings(), Overrides: make(map[string]ContactPrivacyOverride)}
		s.profiles[did] = rec
	}
	return rec
}

// UpdateOwnProfile patches the caller's editable profile fields.
func (s *Service) UpdateOwnProfile(ctx context.Context, did string, displayName, bio, statusMessage, avatarURL *string) (*UserProfile, error) {
	rec := s.ensureProfile(did)
	s.profileMu().Lock()
	if displayName != nil {
		rec.DisplayName = strings.TrimSpace(*displayName)
	}
	if bio != nil {
		rec.Bio = strings.TrimSpace(*bio)
	}
	if statusMessage != nil {
		rec.StatusMessage = strings.TrimSpace(*statusMessage)
	}
	if avatarURL != nil {
		rec.AvatarURL = strings.TrimSpace(*avatarURL)
	}
	s.profileMu().Unlock()
	return s.GetProfile(ctx, did, did)
}

// UpdatePrivacy replaces global privacy settings for a user.
func (s *Service) UpdatePrivacy(ctx context.Context, did string, settings PrivacySettings) (PrivacySettings, error) {
	rec := s.ensureProfile(did)
	s.profileMu().Lock()
	rec.Privacy = settings
	s.profileMu().Unlock()
	return settings, nil
}

// GetPrivacy returns stored privacy settings for the owner.
func (s *Service) GetPrivacy(ctx context.Context, did string) PrivacySettings {
	rec := s.ensureProfile(did)
	s.profileMu().RLock()
	defer s.profileMu().RUnlock()
	return rec.Privacy
}

// SetContactPrivacyOverride stores per-contact prefs for the owner.
func (s *Service) SetContactPrivacyOverride(ctx context.Context, ownerDID, peerDID string, override ContactPrivacyOverride) error {
	if ownerDID == "" || peerDID == "" || ownerDID == peerDID {
		return ErrSelfContact
	}
	rec := s.ensureProfile(ownerDID)
	s.profileMu().Lock()
	rec.Overrides[peerDID] = override
	s.profileMu().Unlock()
	return nil
}

// GetContactPrivacyOverride returns per-contact prefs for a peer.
func (s *Service) GetContactPrivacyOverride(ctx context.Context, ownerDID, peerDID string) (ContactPrivacyOverride, bool) {
	rec := s.ensureProfile(ownerDID)
	s.profileMu().RLock()
	defer s.profileMu().RUnlock()
	o, ok := rec.Overrides[peerDID]
	return o, ok
}

// GetProfile returns a privacy-filtered profile for viewer → target.
func (s *Service) GetProfile(ctx context.Context, viewerDID, targetDID string) (*UserProfile, error) {
	user, err := s.db.GetUserByDID(ctx, targetDID)
	if err != nil {
		return nil, err
	}
	rec := s.ensureProfile(targetDID)
	tier := user.TrustTier
	if ts, err := s.db.GetTrustScore(ctx, targetDID); err == nil {
		tier = ts.Tier
	}

	isContact := false
	if viewerDID != targetDID {
		contacts, _ := s.db.GetContacts(ctx, viewerDID)
		for _, c := range contacts {
			if c.ContactDID == targetDID && !c.Blocked {
				isContact = true
				break
			}
		}
	} else {
		isContact = true
	}

	blocked, _ := s.IsBlocked(ctx, viewerDID, targetDID)
	blockedMe, _ := s.IsBlocked(ctx, targetDID, viewerDID)

	out := &UserProfile{
		DID:        targetDID,
		Username:   user.Username,
		TrustTier:  tier,
		IsVerified: tier >= 4,
		IsContact:  isContact,
		IsBlocked:  blocked || blockedMe,
	}

	s.profileMu().RLock()
	priv := rec.Privacy
	displayName := rec.DisplayName
	bio := rec.Bio
	status := rec.StatusMessage
	avatar := rec.AvatarURL
	s.profileMu().RUnlock()

	if displayName == "" {
		displayName = user.Username
	}

	if viewerDID == targetDID {
		out.DisplayName = displayName
		out.Bio = bio
		out.StatusMessage = status
		out.AvatarURL = avatar
		return out, nil
	}

	if blocked || blockedMe {
		return out, nil
	}

	if visibleTo(priv.ShowProfilePicture, isContact) {
		out.AvatarURL = avatar
	}
	if visibleTo(priv.ShowStatusMessage, isContact) {
		out.StatusMessage = status
	}
	out.DisplayName = displayName
	if isContact || priv.ShowProfilePicture == VisibilityEveryone {
		out.Bio = bio
	}
	return out, nil
}

func visibleTo(level VisibilityLevel, isContact bool) bool {
	switch level {
	case VisibilityEveryone:
		return true
	case VisibilityContacts:
		return isContact
	default:
		return false
	}
}
