package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ProfileData represents the full user profile
type ProfileData struct {
	DisplayName   string   `json:"display_name"`
	Username      string   `json:"username"`
	Bio           string   `json:"bio"`
	Status        string   `json:"status"`
	AvatarURL     string   `json:"avatar_url,omitempty"`
	Website       string   `json:"website,omitempty"`
	Links         []string `json:"links"`
	TrustScore    int      `json:"trust_score"`
	TrustLevel    string   `json:"trust_level"`
	IsVerified    bool     `json:"is_verified"`
	MessagesSent  int      `json:"messages_sent"`
	ContactsCount int      `json:"contacts_count"`
	EchoRewards   float64  `json:"echo_rewards"`
}

// Persona represents a user persona
type Persona struct {
	ID                   string                       `json:"id"`
	Type                 string                       `json:"type"`
	Name                 string                       `json:"name"`
	DisplayName          string                       `json:"display_name"`
	Username             string                       `json:"username,omitempty"`
	Bio                  string                       `json:"bio,omitempty"`
	AvatarURL            string                       `json:"avatar_url,omitempty"`
	UseMainAvatar        bool                         `json:"use_main_avatar"`
	Visibility           string                       `json:"visibility"`
	DefaultVisibility    string                       `json:"default_visibility"`
	Discoverability      bool                         `json:"discoverability"`
	SelectedContactIDs   []string                     `json:"selected_contact_ids"`
	AccessGrants         []AccessGrant                `json:"access_grants"`
	IsDefault            bool                         `json:"is_default"`
	CreatedAt            string                       `json:"created_at"`
	UpdatedAt            string                       `json:"updated_at"`
	LastActiveAt         string                       `json:"last_active_at,omitempty"`
	MessageCount         int                          `json:"message_count"`
	Status               string                       `json:"status,omitempty"`
	PrivacySettings      PersonaPrivacySettingsData   `json:"privacy_settings"`
	NotificationSettings PersonaNotifSettingsData     `json:"notification_settings"`
	FeatureSettings      PersonaFeatureSettingsData   `json:"feature_settings"`
	Badges               []PersonaBadgeData           `json:"badges"`
	DeletionState        *PersonaDeletionStateData    `json:"deletion_state,omitempty"`
}

// AccessGrant represents a per-contact permission grant
type AccessGrant struct {
	ID                  string  `json:"id"`
	ContactID           string  `json:"contact_id"`
	PersonaID           string  `json:"persona_id"`
	GrantedAt           string  `json:"granted_at"`
	GrantedBy           string  `json:"granted_by"`
	CanView             bool    `json:"can_view"`
	CanMessage          bool    `json:"can_message"`
	CanCall             bool    `json:"can_call"`
	CanSeeOtherPersonas bool    `json:"can_see_other_personas"`
	ExpiresAt           *string `json:"expires_at,omitempty"`
	Revocable           bool    `json:"revocable"`
}

// PersonaPrivacySettingsData represents per-persona privacy
type PersonaPrivacySettingsData struct {
	LastSeenVisibility       string `json:"last_seen_visibility"`
	OnlineStatusVisibility   string `json:"online_status_visibility"`
	ProfilePictureVisibility string `json:"profile_picture_visibility"`
	BioVisibility            string `json:"bio_visibility"`
	StatusMessageVisibility  string `json:"status_message_visibility"`
	WhoCanMessage            string `json:"who_can_message"`
	WhoCanCall               string `json:"who_can_call"`
	WhoCanAddToGroups        string `json:"who_can_add_to_groups"`
	RequireApproval          bool   `json:"require_approval_for_contact"`
	SendReadReceipts         bool   `json:"send_read_receipts"`
	SendTypingIndicators     bool   `json:"send_typing_indicators"`
	Searchable               bool   `json:"searchable"`
	ShowInSuggestions        bool   `json:"show_in_suggestions"`
	AllowContactSharing      bool   `json:"allow_contact_sharing"`
	AllowLinkingDiscovery    bool   `json:"allow_linking_discovery"`
	ShowSharedTrustScore     bool   `json:"show_shared_trust_score"`
	AllowCrossPersonaForward bool   `json:"allow_cross_persona_forward"`
}

// PersonaNotifSettingsData represents per-persona notifications
type PersonaNotifSettingsData struct {
	Enabled                  bool   `json:"enabled"`
	QuietHoursEnabled        bool   `json:"quiet_hours_enabled"`
	QuietHoursStart          string `json:"quiet_hours_start"`
	QuietHoursEnd            string `json:"quiet_hours_end"`
	QuietHoursTimezone       string `json:"quiet_hours_timezone"`
	QuietHoursAllowExceptions bool  `json:"quiet_hours_allow_exceptions"`
	MessagesMode             string `json:"messages_mode"`
	CallsMode                string `json:"calls_mode"`
	GroupActivityMode        string `json:"group_activity_mode"`
	ContactRequests          bool   `json:"contact_requests"`
	SoundEnabled             bool   `json:"sound_enabled"`
	SoundID                  string `json:"sound_id"`
	VibrationEnabled         bool   `json:"vibration_enabled"`
	ShowContent              bool   `json:"show_content"`
	ShowSender               bool   `json:"show_sender"`
	ShowPersonaName          bool   `json:"show_persona_name"`
}

// PersonaFeatureSettingsData represents per-persona feature toggles
type PersonaFeatureSettingsData struct {
	VoiceCalls               bool `json:"voice_calls"`
	VideoCalls               bool `json:"video_calls"`
	ScreenSharing            bool `json:"screen_sharing"`
	VoiceMessages            bool `json:"voice_messages"`
	FileSharing              bool `json:"file_sharing"`
	LocationSharing          bool `json:"location_sharing"`
	DisappearingMessages     bool `json:"disappearing_messages"`
	ScheduledMessages        bool `json:"scheduled_messages"`
	SilentMessages           bool `json:"silent_messages"`
	MaxGroupSize             int  `json:"max_group_size"`
	MaxFileSizeMB            int  `json:"max_file_size_mb"`
	VoiceMessageDurationSecs int  `json:"voice_message_duration_seconds"`
}

// PersonaBadgeData represents a badge on a persona
type PersonaBadgeData struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	IssuedAt   string `json:"issued_at"`
	Issuer     string `json:"issuer"`
	Verifiable bool   `json:"verifiable"`
	Proof      string `json:"proof,omitempty"`
}

// PersonaDeletionStateData represents soft-delete state
type PersonaDeletionStateData struct {
	DeletedAt            string  `json:"deleted_at"`
	RecoveryExpiresAt    *string `json:"recovery_expires_at,omitempty"`
	ArchiveConversations bool    `json:"archive_conversations"`
	NotifyContacts       bool    `json:"notify_contacts"`
	IsRecoverable        bool    `json:"is_recoverable"`
}

// PersonaConversationData represents an isolated persona conversation
type PersonaConversationData struct {
	ID               string  `json:"id"`
	PersonaID        string  `json:"persona_id"`
	ContactID        string  `json:"contact_id"`
	ContactPersonaID *string `json:"contact_persona_id,omitempty"`
	LastMessageAt    *string `json:"last_message_at,omitempty"`
	UnreadCount      int     `json:"unread_count"`
	MessageCount     int     `json:"message_count"`
}

// CreatePersonaRequest represents a persona creation request
type CreatePersonaRequest struct {
	Type               string   `json:"type"`
	Name               string   `json:"name"`
	DisplayName        string   `json:"display_name"`
	Username           string   `json:"username,omitempty"`
	Bio                string   `json:"bio,omitempty"`
	UseMainAvatar      bool     `json:"use_main_avatar"`
	Visibility         string   `json:"visibility"`
	SelectedContactIDs []string `json:"selected_contact_ids"`
}

// DeletePersonaRequest represents persona deletion options
type DeletePersonaRequest struct {
	ArchiveConversations bool `json:"archive_conversations"`
	NotifyContacts       bool `json:"notify_contacts"`
	RecoveryPeriodDays   int  `json:"recovery_period_days"`
	ExportBeforeDelete   bool `json:"export_before_delete"`
}

// GrantAccessRequest represents an access grant request
type GrantAccessRequest struct {
	ContactID           string  `json:"contact_id"`
	CanView             bool    `json:"can_view"`
	CanMessage          bool    `json:"can_message"`
	CanCall             bool    `json:"can_call"`
	CanSeeOtherPersonas bool    `json:"can_see_other_personas"`
	ExpiresAt           *string `json:"expires_at,omitempty"`
}

// ValidateSwitchRequest represents a persona switch validation request
type ValidateSwitchRequest struct {
	FromPersonaID *string `json:"from_persona_id,omitempty"`
	ToPersonaID   string  `json:"to_persona_id"`
	ContactID     string  `json:"contact_id"`
}

// AddBadgeRequest represents a badge addition request
type AddBadgeRequest struct {
	Type       string `json:"type"`
	Issuer     string `json:"issuer"`
	Verifiable bool   `json:"verifiable"`
	Proof      string `json:"proof,omitempty"`
}

// UpdatePersonaRequest represents a persona update request
type UpdatePersonaRequest struct {
	Type               *string  `json:"type,omitempty"`
	Name               *string  `json:"name,omitempty"`
	DisplayName        *string  `json:"display_name,omitempty"`
	Bio                *string  `json:"bio,omitempty"`
	UseMainAvatar      *bool    `json:"use_main_avatar,omitempty"`
	Visibility         *string  `json:"visibility,omitempty"`
	SelectedContactIDs []string `json:"selected_contact_ids,omitempty"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Username    *string `json:"username,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	Status      *string `json:"status,omitempty"`
	Website     *string `json:"website,omitempty"`
}

// NotificationSettingsData represents notification preferences
type NotificationSettingsData struct {
	MessageNotifications bool   `json:"message_notifications"`
	ShowPreviews         string `json:"show_previews"`
	MessageSound         string `json:"message_sound"`
	GroupNotifications   bool   `json:"group_notifications"`
	MentionsOnly         bool   `json:"mentions_only"`
	CallNotifications    bool   `json:"call_notifications"`
	RingSound            string `json:"ring_sound"`
	ContactRequests      bool   `json:"contact_requests"`
	TrustScoreChanges    bool   `json:"trust_score_changes"`
	RewardNotifications  bool   `json:"reward_notifications"`
	QuietHoursEnabled    bool   `json:"quiet_hours_enabled"`
	QuietHoursFrom       string `json:"quiet_hours_from"`
	QuietHoursTo         string `json:"quiet_hours_to"`
	AllowInnerCircle     bool   `json:"allow_inner_circle_calls"`
}

// PrivacySettingsData represents privacy preferences
type PrivacySettingsData struct {
	FindByUsername          string `json:"find_by_username"`
	ShowOnlineStatus        string `json:"show_online_status"`
	ShowTrustScore          string `json:"show_trust_score"`
	WhoCanMessage           string `json:"who_can_message"`
	ReadReceipts            bool   `json:"read_receipts"`
	TypingIndicators        bool   `json:"typing_indicators"`
	WhoCanCall              string `json:"who_can_call"`
	AnchorMessagesByDefault bool   `json:"anchor_messages_by_default"`
	ShowVerificationBadges  bool   `json:"show_verification_badges"`
	ScreenLockTimeout       string `json:"screen_lock_timeout"`
	HideMessagePreviews     bool   `json:"hide_message_previews"`
	ScreenshotNotifications bool   `json:"screenshot_notifications"`
}

// AppearanceSettingsData represents appearance preferences
type AppearanceSettingsData struct {
	Theme          string `json:"theme"`
	AccentColor    string `json:"accent_color"`
	ChatWallpaper  string `json:"chat_wallpaper"`
	MessageCorners string `json:"message_corners"`
	FontSize       string `json:"font_size"`
	AppIcon        string `json:"app_icon"`
}

// StorageInfoData represents storage and data info
type StorageInfoData struct {
	TotalUsedBytes      int64   `json:"total_used_bytes"`
	TotalCapacityBytes  int64   `json:"total_capacity_bytes"`
	PhotosVideosBytes   int64   `json:"photos_videos_bytes"`
	DocumentsBytes      int64   `json:"documents_bytes"`
	VoiceMessagesBytes  int64   `json:"voice_messages_bytes"`
	OtherBytes          int64   `json:"other_bytes"`
	CacheBytes          int64   `json:"cache_bytes"`
	AutoDownload        string  `json:"auto_download"`
	MediaQuality        string  `json:"media_quality"`
	KeepMedia           string  `json:"keep_media"`
	UseLessData         bool    `json:"use_less_data_for_calls"`
	AutoBackup          string  `json:"auto_backup"`
	IncludeMedia        bool    `json:"include_media_in_backup"`
	LastBackupDate      *string `json:"last_backup_date"`
}

// AccountInfoData represents account information
type AccountInfoData struct {
	Phone                     string `json:"phone"`
	Email                     string `json:"email,omitempty"`
	DID                       string `json:"did,omitempty"`
	PasskeyCount              int    `json:"passkey_count"`
	TwoFactorEnabled          bool   `json:"two_factor_enabled"`
	ActiveSessionCount        int    `json:"active_session_count"`
	RecoveryPhraseSetUp       bool   `json:"recovery_phrase_set_up"`
	TrustedRecoveryContacts   int    `json:"trusted_recovery_contact_count"`
}

// ProfileHandler manages profile and persona operations
type ProfileHandler struct {
	mu            sync.RWMutex
	personas      map[string][]Persona
	profiles      map[string]*ProfileData
	conversations map[string][]PersonaConversationData // key: personaID
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{
		personas:      make(map[string][]Persona),
		profiles:      make(map[string]*ProfileData),
		conversations: make(map[string][]PersonaConversationData),
	}
}

// Trust-level persona limits
var trustLevelLimits = map[string]int{
	"unverified": 2,
	"newcomer":   3,
	"member":     5,
	"basic":      5,
	"trusted":    7,
	"verified":   10,
	"elite":      10,
}

func (ph *ProfileHandler) getMaxPersonas(userID string) int {
	ph.mu.RLock()
	profile := ph.profiles[userID]
	ph.mu.RUnlock()

	if profile != nil {
		level := strings.ToLower(profile.TrustLevel)
		if max, ok := trustLevelLimits[level]; ok {
			return max
		}
	}
	return 2 // default for unverified
}

var validPersonaTypes = map[string]bool{
	"professional": true,
	"personal":     true,
	"family":       true,
	"gaming":       true,
	"dating":       true,
	"creative":     true,
	"anonymous":    true,
	"custom":       true,
}

var validVisibilities = map[string]bool{
	"all":      true,
	"selected": true,
	"hidden":   true,
}

// GetFullProfile handles GET /v2/users/profile/full
func (ph *ProfileHandler) GetFullProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.RLock()
	profile, exists := ph.profiles[userID]
	ph.mu.RUnlock()

	if !exists {
		profile = &ProfileData{
			DisplayName:   "Alex Echo",
			Username:      "alexecho",
			Bio:           "Product designer & crypto enthusiast. Building the future of trusted comms.",
			Status:        "Shipping new features",
			TrustScore:    72,
			TrustLevel:    "Trusted",
			IsVerified:    true,
			MessagesSent:  247,
			ContactsCount: 42,
			EchoRewards:   142.5,
			Links:         []string{},
		}
	}

	writeJSON(w, http.StatusOK, profile)
}

// UpdateProfile handles PUT /v2/users/profile
func (ph *ProfileHandler) UpdateFullProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PUT is allowed", "")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	profile, exists := ph.profiles[userID]
	if !exists {
		profile = &ProfileData{
			TrustScore:    72,
			TrustLevel:    "Trusted",
			IsVerified:    true,
			MessagesSent:  247,
			ContactsCount: 42,
			EchoRewards:   142.5,
			Links:         []string{},
		}
		ph.profiles[userID] = profile
	}

	if req.DisplayName != nil {
		profile.DisplayName = *req.DisplayName
	}
	if req.Username != nil {
		if len(*req.Username) < 3 || len(*req.Username) > 30 {
			ph.mu.Unlock()
			writeV2Error(w, http.StatusBadRequest, "INVALID_USERNAME", "Username must be 3-30 characters", "")
			return
		}
		profile.Username = *req.Username
	}
	if req.Bio != nil {
		if len(*req.Bio) > 500 {
			ph.mu.Unlock()
			writeV2Error(w, http.StatusBadRequest, "BIO_TOO_LONG", "Bio must be 500 characters or fewer", "")
			return
		}
		profile.Bio = *req.Bio
	}
	if req.Status != nil {
		profile.Status = *req.Status
	}
	if req.Website != nil {
		profile.Website = *req.Website
	}
	ph.mu.Unlock()

	writeJSON(w, http.StatusOK, profile)
}

// CheckUsernameAvailability handles GET /v2/users/check-username
func (ph *ProfileHandler) CheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_USERNAME", "Username query parameter is required", "")
		return
	}

	// Check against existing profiles
	ph.mu.RLock()
	available := true
	for _, profile := range ph.profiles {
		if strings.EqualFold(profile.Username, username) {
			available = false
			break
		}
	}
	ph.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"username":  username,
		"available": available,
	})
}

// ListPersonas handles GET /v2/users/personas
func (ph *ProfileHandler) ListPersonas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.RLock()
	personas := ph.personas[userID]
	ph.mu.RUnlock()

	if personas == nil {
		personas = []Persona{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  personas,
		"count": len(personas),
	})
}

// CreatePersona handles POST /v2/users/personas
func (ph *ProfileHandler) CreatePersona(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", "")
		return
	}

	var req CreatePersonaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", "")
		return
	}

	// Validate
	if req.Name == "" || req.DisplayName == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_FIELDS", "Name and display_name are required", "")
		return
	}

	if !validPersonaTypes[req.Type] {
		writeV2Error(w, http.StatusBadRequest, "INVALID_TYPE", "Invalid persona type", "")
		return
	}

	if !validVisibilities[req.Visibility] {
		writeV2Error(w, http.StatusBadRequest, "INVALID_VISIBILITY", "Invalid visibility setting", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	// Get max personas before acquiring write lock to avoid deadlock
	maxAllowed := ph.getMaxPersonas(userID)

	ph.mu.Lock()
	defer ph.mu.Unlock()

	if len(ph.personas[userID]) >= maxAllowed {
		writeV2Error(w, http.StatusConflict, "MAX_PERSONAS", fmt.Sprintf("Maximum of %d personas allowed for your trust level", maxAllowed), "")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	isDefault := len(ph.personas[userID]) == 0

	persona := Persona{
		ID:                fmt.Sprintf("persona-%d", time.Now().UnixNano()),
		Type:              req.Type,
		Name:              req.Name,
		DisplayName:       req.DisplayName,
		Username:          req.Username,
		Bio:               req.Bio,
		UseMainAvatar:     req.UseMainAvatar,
		Visibility:        req.Visibility,
		DefaultVisibility: "contacts",
		Discoverability:   true,
		SelectedContactIDs: req.SelectedContactIDs,
		AccessGrants:      []AccessGrant{},
		IsDefault:         isDefault,
		CreatedAt:         now,
		UpdatedAt:         now,
		MessageCount:      0,
		PrivacySettings:   defaultPersonaPrivacySettings(),
		NotificationSettings: defaultPersonaNotifSettings(),
		FeatureSettings:   defaultPersonaFeatureSettings(),
		Badges:            []PersonaBadgeData{},
	}

	if persona.SelectedContactIDs == nil {
		persona.SelectedContactIDs = []string{}
	}

	// Check username uniqueness if provided
	if req.Username != "" {
		for _, userPersonas := range ph.personas {
			for _, p := range userPersonas {
				if strings.EqualFold(p.Username, req.Username) {
					writeV2Error(w, http.StatusConflict, "USERNAME_TAKEN", "This username is already in use", "")
					return
				}
			}
		}
	}

	ph.personas[userID] = append(ph.personas[userID], persona)

	writeJSON(w, http.StatusCreated, persona)
}

// UpdatePersona handles PUT /v2/users/personas/{id}
func (ph *ProfileHandler) UpdatePersona(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PUT is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	if personaID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Persona ID is required", "")
		return
	}

	var req UpdatePersonaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]
	idx := -1
	for i, p := range personas {
		if p.ID == personaID {
			idx = i
			break
		}
	}

	if idx == -1 {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	if req.Type != nil {
		if !validPersonaTypes[*req.Type] {
			writeV2Error(w, http.StatusBadRequest, "INVALID_TYPE", "Invalid persona type", "")
			return
		}
		personas[idx].Type = *req.Type
	}
	if req.Name != nil {
		personas[idx].Name = *req.Name
	}
	if req.DisplayName != nil {
		personas[idx].DisplayName = *req.DisplayName
	}
	if req.Bio != nil {
		personas[idx].Bio = *req.Bio
	}
	if req.UseMainAvatar != nil {
		personas[idx].UseMainAvatar = *req.UseMainAvatar
	}
	if req.Visibility != nil {
		if !validVisibilities[*req.Visibility] {
			writeV2Error(w, http.StatusBadRequest, "INVALID_VISIBILITY", "Invalid visibility setting", "")
			return
		}
		personas[idx].Visibility = *req.Visibility
	}
	if req.SelectedContactIDs != nil {
		personas[idx].SelectedContactIDs = req.SelectedContactIDs
	}
	personas[idx].UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	ph.personas[userID] = personas
	writeJSON(w, http.StatusOK, personas[idx])
}

// DeletePersona handles DELETE /v2/users/personas/{id}
func (ph *ProfileHandler) DeletePersona(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only DELETE is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	if personaID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Persona ID is required", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]
	idx := -1
	for i, p := range personas {
		if p.ID == personaID {
			idx = i
			break
		}
	}

	if idx == -1 {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	wasDefault := personas[idx].IsDefault
	personas = append(personas[:idx], personas[idx+1:]...)

	// If we deleted the default, make the first one default
	if wasDefault && len(personas) > 0 {
		personas[0].IsDefault = true
	}

	ph.personas[userID] = personas

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"deleted":    true,
		"persona_id": personaID,
	})
}

// SetDefaultPersona handles POST /v2/users/personas/{id}/default
func (ph *ProfileHandler) SetDefaultPersona(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	if personaID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Persona ID is required", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]
	found := false
	for i := range personas {
		if personas[i].ID == personaID {
			personas[i].IsDefault = true
			found = true
		} else {
			personas[i].IsDefault = false
		}
	}

	if !found {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	ph.personas[userID] = personas

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"default_persona_id": personaID,
	})
}

// GetNotificationSettings handles GET /v2/users/settings/notifications
func (ph *ProfileHandler) GetNotificationSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	settings := NotificationSettingsData{
		MessageNotifications: true,
		ShowPreviews:         "always",
		MessageSound:         "Echo Default",
		GroupNotifications:   true,
		MentionsOnly:         false,
		CallNotifications:    true,
		RingSound:            "Reflection",
		ContactRequests:      true,
		TrustScoreChanges:    true,
		RewardNotifications:  true,
		QuietHoursEnabled:    false,
		QuietHoursFrom:       "22:00",
		QuietHoursTo:         "07:00",
		AllowInnerCircle:     true,
	}

	writeJSON(w, http.StatusOK, settings)
}

// GetPrivacySettings handles GET /v2/users/settings/privacy
func (ph *ProfileHandler) GetPrivacySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	settings := PrivacySettingsData{
		FindByUsername:          "everyone",
		ShowOnlineStatus:        "contacts",
		ShowTrustScore:          "everyone",
		WhoCanMessage:           "contacts",
		ReadReceipts:            true,
		TypingIndicators:        true,
		WhoCanCall:              "trusted",
		AnchorMessagesByDefault: false,
		ShowVerificationBadges:  true,
		ScreenLockTimeout:       "immediately",
		HideMessagePreviews:     true,
		ScreenshotNotifications: true,
	}

	writeJSON(w, http.StatusOK, settings)
}

// GetAppearanceSettings handles GET /v2/users/settings/appearance
func (ph *ProfileHandler) GetAppearanceSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	settings := AppearanceSettingsData{
		Theme:          "light",
		AccentColor:    "indigo",
		ChatWallpaper:  "default",
		MessageCorners: "rounded",
		FontSize:       "medium",
		AppIcon:        "default",
	}

	writeJSON(w, http.StatusOK, settings)
}

// GetStorageInfo handles GET /v2/users/storage
func (ph *ProfileHandler) GetStorageInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	lastBackup := time.Now().UTC().Format(time.RFC3339)
	info := StorageInfoData{
		TotalUsedBytes:     2576980378,  // ~2.4 GB
		TotalCapacityBytes: 5368709120,  // 5 GB
		PhotosVideosBytes:  1932735283,  // ~1.8 GB
		DocumentsBytes:     440401920,   // ~420 MB
		VoiceMessagesBytes: 188743680,   // ~180 MB
		OtherBytes:         44040192,    // ~42 MB
		CacheBytes:         163577856,   // ~156 MB
		AutoDownload:       "wifi",
		MediaQuality:       "standard",
		KeepMedia:          "forever",
		UseLessData:        false,
		AutoBackup:         "daily",
		IncludeMedia:       true,
		LastBackupDate:     &lastBackup,
	}

	writeJSON(w, http.StatusOK, info)
}

// GetAccountInfo handles GET /v2/users/account
func (ph *ProfileHandler) GetAccountInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	info := AccountInfoData{
		Phone:                   "+15551234567",
		Email:                   "",
		DID:                     "did:cardano:abc1...xyz",
		PasskeyCount:            2,
		TwoFactorEnabled:        true,
		ActiveSessionCount:      3,
		RecoveryPhraseSetUp:     true,
		TrustedRecoveryContacts: 2,
	}

	writeJSON(w, http.StatusOK, info)
}

// MARK: - Default Settings Factories

func defaultPersonaPrivacySettings() PersonaPrivacySettingsData {
	return PersonaPrivacySettingsData{
		LastSeenVisibility:       "contacts",
		OnlineStatusVisibility:   "contacts",
		ProfilePictureVisibility: "everyone",
		BioVisibility:            "everyone",
		StatusMessageVisibility:  "contacts",
		WhoCanMessage:            "contacts",
		WhoCanCall:               "contacts",
		WhoCanAddToGroups:        "contacts",
		RequireApproval:          false,
		SendReadReceipts:         true,
		SendTypingIndicators:     true,
		Searchable:               true,
		ShowInSuggestions:        true,
		AllowContactSharing:      false,
		AllowLinkingDiscovery:    false,
		ShowSharedTrustScore:     true,
		AllowCrossPersonaForward: false,
	}
}

func defaultPersonaNotifSettings() PersonaNotifSettingsData {
	return PersonaNotifSettingsData{
		Enabled:                   true,
		QuietHoursEnabled:         false,
		QuietHoursStart:           "22:00",
		QuietHoursEnd:             "07:00",
		QuietHoursTimezone:        "UTC",
		QuietHoursAllowExceptions: true,
		MessagesMode:              "all",
		CallsMode:                 "all",
		GroupActivityMode:         "all",
		ContactRequests:           true,
		SoundEnabled:              true,
		SoundID:                   "default",
		VibrationEnabled:          true,
		ShowContent:               true,
		ShowSender:                true,
		ShowPersonaName:           true,
	}
}

func defaultPersonaFeatureSettings() PersonaFeatureSettingsData {
	return PersonaFeatureSettingsData{
		VoiceCalls:               true,
		VideoCalls:               true,
		ScreenSharing:            false,
		VoiceMessages:            true,
		FileSharing:              true,
		LocationSharing:          false,
		DisappearingMessages:     false,
		ScheduledMessages:        true,
		SilentMessages:           true,
		MaxGroupSize:             256,
		MaxFileSizeMB:            100,
		VoiceMessageDurationSecs: 300,
	}
}

// MARK: - Enhanced Deletion with Recovery

// DeletePersonaEnhanced handles DELETE /v2/users/personas/{id} with recovery options
func (ph *ProfileHandler) DeletePersonaEnhanced(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only DELETE is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	if personaID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Persona ID is required", "")
		return
	}

	var req DeletePersonaRequest
	// Try to decode body, use defaults if empty
	json.NewDecoder(r.Body).Decode(&req)

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]

	// Cannot delete last persona
	activeCount := 0
	for _, p := range personas {
		if p.DeletionState == nil {
			activeCount++
		}
	}
	if activeCount <= 1 {
		writeV2Error(w, http.StatusConflict, "CANNOT_DELETE_LAST", "You must have at least one persona", "")
		return
	}

	idx := -1
	for i, p := range personas {
		if p.ID == personaID {
			idx = i
			break
		}
	}

	if idx == -1 {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	if req.RecoveryPeriodDays > 0 {
		// Soft delete with recovery
		recoveryExpires := time.Now().UTC().AddDate(0, 0, req.RecoveryPeriodDays).Format(time.RFC3339)
		personas[idx].DeletionState = &PersonaDeletionStateData{
			DeletedAt:            now,
			RecoveryExpiresAt:    &recoveryExpires,
			ArchiveConversations: req.ArchiveConversations,
			NotifyContacts:       req.NotifyContacts,
			IsRecoverable:        true,
		}

		// If it was default, promote the next active persona
		if personas[idx].IsDefault {
			personas[idx].IsDefault = false
			for i := range personas {
				if i != idx && personas[i].DeletionState == nil {
					personas[i].IsDefault = true
					break
				}
			}
		}

		ph.personas[userID] = personas
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"deleted":             true,
			"persona_id":          personaID,
			"recoverable":         true,
			"recovery_expires_at": recoveryExpires,
		})
	} else {
		// Hard delete (immediate)
		wasDefault := personas[idx].IsDefault
		personas = append(personas[:idx], personas[idx+1:]...)

		if wasDefault && len(personas) > 0 {
			personas[0].IsDefault = true
		}

		ph.personas[userID] = personas
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"deleted":     true,
			"persona_id":  personaID,
			"recoverable": false,
		})
	}
}

// RecoverPersona handles POST /v2/users/personas/{id}/recover
func (ph *ProfileHandler) RecoverPersona(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	if personaID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Persona ID is required", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]
	idx := -1
	for i, p := range personas {
		if p.ID == personaID {
			idx = i
			break
		}
	}

	if idx == -1 {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	if personas[idx].DeletionState == nil || !personas[idx].DeletionState.IsRecoverable {
		writeV2Error(w, http.StatusConflict, "NOT_RECOVERABLE", "Persona cannot be recovered", "")
		return
	}

	// Check if recovery period has expired
	if personas[idx].DeletionState.RecoveryExpiresAt != nil {
		expiresAt, err := time.Parse(time.RFC3339, *personas[idx].DeletionState.RecoveryExpiresAt)
		if err == nil && time.Now().UTC().After(expiresAt) {
			writeV2Error(w, http.StatusConflict, "RECOVERY_EXPIRED", "The recovery period for this persona has expired", "")
			return
		}
	}

	// Restore persona
	personas[idx].DeletionState = nil
	personas[idx].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	ph.personas[userID] = personas

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"recovered": true,
		"persona":   personas[idx],
		"warning":   "Persona restored. You may need to re-grant access to contacts.",
	})
}

// MARK: - Access Grants

// GrantAccess handles POST /v2/users/personas/{id}/access
func (ph *ProfileHandler) GrantAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	if personaID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Persona ID is required", "")
		return
	}

	var req GrantAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", "")
		return
	}

	if req.ContactID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_CONTACT", "Contact ID is required", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]
	idx := -1
	for i, p := range personas {
		if p.ID == personaID {
			idx = i
			break
		}
	}

	if idx == -1 {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	grant := AccessGrant{
		ID:                  fmt.Sprintf("grant-%d", time.Now().UnixNano()),
		ContactID:           req.ContactID,
		PersonaID:           personaID,
		GrantedAt:           now,
		GrantedBy:           personaID,
		CanView:             req.CanView,
		CanMessage:          req.CanMessage,
		CanCall:             req.CanCall,
		CanSeeOtherPersonas: req.CanSeeOtherPersonas,
		ExpiresAt:           req.ExpiresAt,
		Revocable:           true,
	}

	personas[idx].AccessGrants = append(personas[idx].AccessGrants, grant)
	if !containsString(personas[idx].SelectedContactIDs, req.ContactID) {
		personas[idx].SelectedContactIDs = append(personas[idx].SelectedContactIDs, req.ContactID)
	}
	ph.personas[userID] = personas

	writeJSON(w, http.StatusCreated, grant)
}

// RevokeAccess handles DELETE /v2/users/personas/access/{grantId}
func (ph *ProfileHandler) RevokeAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only DELETE is allowed", "")
		return
	}

	grantID := r.URL.Query().Get("grant_id")
	if grantID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Grant ID is required", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]
	found := false
	for i := range personas {
		for j, g := range personas[i].AccessGrants {
			if g.ID == grantID {
				personas[i].AccessGrants = append(personas[i].AccessGrants[:j], personas[i].AccessGrants[j+1:]...)
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Access grant not found", "")
		return
	}

	ph.personas[userID] = personas
	writeJSON(w, http.StatusOK, map[string]interface{}{"revoked": true, "grant_id": grantID})
}

// GetVisibilityMatrix handles GET /v2/users/personas/visibility-matrix
func (ph *ProfileHandler) GetVisibilityMatrix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.RLock()
	personas := ph.personas[userID]
	ph.mu.RUnlock()

	// Build matrix: for each unique contact, which personas they can see
	contactMap := make(map[string]map[string]bool) // contactID -> personaID -> visible
	for _, p := range personas {
		if p.DeletionState != nil {
			continue
		}
		if p.Visibility == "all" {
			// All contacts can see
			for _, cID := range p.SelectedContactIDs {
				if contactMap[cID] == nil {
					contactMap[cID] = make(map[string]bool)
				}
				contactMap[cID][p.ID] = true
			}
		}
		for _, grant := range p.AccessGrants {
			if contactMap[grant.ContactID] == nil {
				contactMap[grant.ContactID] = make(map[string]bool)
			}
			contactMap[grant.ContactID][p.ID] = grant.CanView
		}
	}

	var entries []map[string]interface{}
	for cID, visibility := range contactMap {
		entries = append(entries, map[string]interface{}{
			"contact_id":        cID,
			"contact_name":      "Contact " + cID,
			"persona_visibility": visibility,
		})
	}

	if entries == nil {
		entries = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"entries": entries})
}

// MARK: - Persona Switching

// ValidatePersonaSwitch handles POST /v2/users/personas/validate-switch
func (ph *ProfileHandler) ValidatePersonaSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", "")
		return
	}

	var req ValidateSwitchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", "")
		return
	}

	if req.ToPersonaID == "" || req.ContactID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_FIELDS", "to_persona_id and contact_id are required", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.RLock()
	personas := ph.personas[userID]
	ph.mu.RUnlock()

	// Check if contact knows both personas
	contactKnowsFrom := false
	contactKnowsTo := false
	for _, p := range personas {
		if req.FromPersonaID != nil && p.ID == *req.FromPersonaID {
			contactKnowsFrom = containsString(p.SelectedContactIDs, req.ContactID) || p.Visibility == "all"
		}
		if p.ID == req.ToPersonaID {
			contactKnowsTo = containsString(p.SelectedContactIDs, req.ContactID) || p.Visibility == "all"
		}
	}

	contactKnowsLink := contactKnowsFrom && contactKnowsTo
	requiresConfirmation := !contactKnowsLink && req.FromPersonaID != nil

	var warningMessage *string
	if requiresConfirmation {
		msg := "Contact does not know these personas are linked. Switching may reveal your identity."
		warningMessage = &msg
	}

	resp := map[string]interface{}{
		"from_persona_id":       req.FromPersonaID,
		"to_persona_id":         req.ToPersonaID,
		"contact_id":            req.ContactID,
		"contact_knows_link":    contactKnowsLink,
		"requires_confirmation": requiresConfirmation,
		"warning_message":       warningMessage,
	}

	writeJSON(w, http.StatusOK, resp)
}

// MARK: - Per-Persona Settings

// GetPersonaPrivacySettings handles GET /v2/users/personas/{id}/settings/privacy
func (ph *ProfileHandler) GetPersonaPrivacySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	persona := ph.findPersona(r, personaID)
	if persona == nil {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	writeJSON(w, http.StatusOK, persona.PrivacySettings)
}

// UpdatePersonaPrivacySettings handles PUT /v2/users/personas/{id}/settings/privacy
func (ph *ProfileHandler) UpdatePersonaPrivacySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PUT is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	var settings PersonaPrivacySettingsData
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", "")
		return
	}

	if ph.updatePersonaField(r, personaID, func(p *Persona) { p.PrivacySettings = settings }) {
		writeJSON(w, http.StatusOK, settings)
	} else {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
	}
}

// GetPersonaNotificationSettings handles GET /v2/users/personas/{id}/settings/notifications
func (ph *ProfileHandler) GetPersonaNotificationSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	persona := ph.findPersona(r, personaID)
	if persona == nil {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	writeJSON(w, http.StatusOK, persona.NotificationSettings)
}

// UpdatePersonaNotificationSettings handles PUT /v2/users/personas/{id}/settings/notifications
func (ph *ProfileHandler) UpdatePersonaNotificationSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PUT is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	var settings PersonaNotifSettingsData
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", "")
		return
	}

	if ph.updatePersonaField(r, personaID, func(p *Persona) { p.NotificationSettings = settings }) {
		writeJSON(w, http.StatusOK, settings)
	} else {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
	}
}

// GetPersonaFeatureSettings handles GET /v2/users/personas/{id}/settings/features
func (ph *ProfileHandler) GetPersonaFeatureSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	persona := ph.findPersona(r, personaID)
	if persona == nil {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	writeJSON(w, http.StatusOK, persona.FeatureSettings)
}

// UpdatePersonaFeatureSettings handles PUT /v2/users/personas/{id}/settings/features
func (ph *ProfileHandler) UpdatePersonaFeatureSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PUT is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	var settings PersonaFeatureSettingsData
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", "")
		return
	}

	if ph.updatePersonaField(r, personaID, func(p *Persona) { p.FeatureSettings = settings }) {
		writeJSON(w, http.StatusOK, settings)
	} else {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
	}
}

// MARK: - Persona Badges

// AddPersonaBadge handles POST /v2/users/personas/{id}/badges
func (ph *ProfileHandler) AddPersonaBadge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	if personaID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Persona ID is required", "")
		return
	}

	var req AddBadgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]
	idx := -1
	for i, p := range personas {
		if p.ID == personaID {
			idx = i
			break
		}
	}

	if idx == -1 {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Persona not found", "")
		return
	}

	badge := PersonaBadgeData{
		ID:         fmt.Sprintf("badge-%d", time.Now().UnixNano()),
		Type:       req.Type,
		IssuedAt:   time.Now().UTC().Format(time.RFC3339),
		Issuer:     req.Issuer,
		Verifiable: req.Verifiable,
		Proof:      req.Proof,
	}

	personas[idx].Badges = append(personas[idx].Badges, badge)
	ph.personas[userID] = personas

	writeJSON(w, http.StatusCreated, badge)
}

// RemovePersonaBadge handles DELETE /v2/users/personas/{id}/badges/{badgeId}
func (ph *ProfileHandler) RemovePersonaBadge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only DELETE is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	badgeID := r.URL.Query().Get("badge_id")
	if personaID == "" || badgeID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Persona ID and Badge ID are required", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]
	found := false
	for i, p := range personas {
		if p.ID == personaID {
			for j, b := range p.Badges {
				if b.ID == badgeID {
					personas[i].Badges = append(personas[i].Badges[:j], personas[i].Badges[j+1:]...)
					found = true
					break
				}
			}
			break
		}
	}

	if !found {
		writeV2Error(w, http.StatusNotFound, "NOT_FOUND", "Badge not found", "")
		return
	}

	ph.personas[userID] = personas
	writeJSON(w, http.StatusOK, map[string]interface{}{"removed": true, "badge_id": badgeID})
}

// MARK: - Persona Conversations (Isolated)

// GetPersonaConversations handles GET /v2/users/personas/{id}/conversations
func (ph *ProfileHandler) GetPersonaConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	personaID := r.URL.Query().Get("id")
	if personaID == "" {
		writeV2Error(w, http.StatusBadRequest, "MISSING_ID", "Persona ID is required", "")
		return
	}

	ph.mu.RLock()
	conversations := ph.conversations[personaID]
	ph.mu.RUnlock()

	if conversations == nil {
		conversations = []PersonaConversationData{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  conversations,
		"count": len(conversations),
	})
}

// GetPersonaLimits handles GET /v2/users/persona-limits
func (ph *ProfileHandler) GetPersonaLimits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", "")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	maxAllowed := ph.getMaxPersonas(userID)

	ph.mu.RLock()
	profile := ph.profiles[userID]
	currentCount := len(ph.personas[userID])
	ph.mu.RUnlock()

	trustLevel := "unverified"
	if profile != nil {
		trustLevel = strings.ToLower(profile.TrustLevel)
	}

	allowCustom := trustLevel == "member" || trustLevel == "basic" || trustLevel == "trusted" || trustLevel == "verified" || trustLevel == "elite"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"max_personas":          maxAllowed,
		"current_count":         currentCount,
		"remaining":             maxAllowed - currentCount,
		"trust_level":           trustLevel,
		"allow_custom_categories": allowCustom,
	})
}

// MARK: - Helper Functions

func (ph *ProfileHandler) findPersona(r *http.Request, personaID string) *Persona {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.RLock()
	defer ph.mu.RUnlock()

	for _, p := range ph.personas[userID] {
		if p.ID == personaID {
			return &p
		}
	}
	return nil
}

func (ph *ProfileHandler) updatePersonaField(r *http.Request, personaID string, update func(*Persona)) bool {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	ph.mu.Lock()
	defer ph.mu.Unlock()

	personas := ph.personas[userID]
	for i, p := range personas {
		if p.ID == personaID {
			update(&personas[i])
			personas[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			ph.personas[userID] = personas
			return true
		}
	}
	return false
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// writeJSON helper to write JSON responses
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
