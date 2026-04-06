package personas

import "time"

type PersonaCategory string

const (
	CategoryProfessional PersonaCategory = "professional"
	CategoryPersonal     PersonaCategory = "personal"
	CategoryFamily       PersonaCategory = "family"
	CategoryGaming       PersonaCategory = "gaming"
	CategoryDating       PersonaCategory = "dating"
	CategoryCreative     PersonaCategory = "creative"
	CategoryAnonymous    PersonaCategory = "anonymous"
	CategoryCustom       PersonaCategory = "custom"
)

type VisibilityLevel string

const (
	VisibilityEveryone VisibilityLevel = "everyone"
	VisibilityContacts VisibilityLevel = "contacts"
	VisibilityNobody   VisibilityLevel = "nobody"
)

type NotificationLevel string

const (
	NotificationAll      NotificationLevel = "all"
	NotificationMentions NotificationLevel = "mentions"
	NotificationNone     NotificationLevel = "none"
)

type MasterIdentity struct {
	UserID     string
	DID        string
	MasterKey  string
	TrustScore int

	CoreVerification VerificationStatus
	Personas         []string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type VerificationStatus struct {
	KYCVerified   bool
	PhoneVerified bool
	EmailVerified bool
}

type Persona struct {
	PersonaID string
	MasterDID string
	UserID    string

	DisplayName  string
	Username     string
	Avatar       string
	Bio          string
	Category     PersonaCategory
	Status       string
	CreatedAt    time.Time
	LastActiveAt time.Time

	SigningKey     string
	EncryptionKey  string
	DerivationPath string

	AccessListIDs     []string
	DefaultVisibility VisibilityLevel
	Discoverability   bool

	Privacy       PersonaPrivacySettings
	Notifications PersonaNotificationSettings
	Features      PersonaFeatureSettings

	Badges       []string
	Credentials  []string
	MessageCount int
	ContactCount int
}

type AccessGrant struct {
	GrantID   string
	ContactID string
	PersonaID string
	GrantedAt time.Time
	GrantedBy string
	ExpiresAt *time.Time
	Revocable bool

	Permissions AccessPermissions
}

type AccessPermissions struct {
	CanView             bool
	CanMessage          bool
	CanCall             bool
	CanSeeOtherPersonas bool
}

type PersonaPrivacySettings struct {
	LastSeenVisibility       VisibilityLevel
	OnlineStatusVisibility   VisibilityLevel
	ProfilePictureVisibility VisibilityLevel
	BioVisibility            VisibilityLevel
	StatusVisibility         VisibilityLevel

	WhoCanMessage             VisibilityLevel
	WhoCanCall                VisibilityLevel
	WhoCanAddToGroups         VisibilityLevel
	RequireApprovalForContact bool

	SendReadReceipts     bool
	SendTypingIndicators bool

	Searchable          bool
	ShowInSuggestions   bool
	AllowContactSharing bool

	AllowLinkingDiscovery    bool
	ShowSharedTrustScore     bool
	AllowCrossPersonaForward bool
}

type PersonaNotificationSettings struct {
	Enabled bool

	Messages        NotificationLevel
	Calls           NotificationLevel
	GroupActivity   NotificationLevel
	ContactRequests bool

	SoundEnabled     bool
	VibrationEnabled bool

	ShowContent     bool
	ShowSender      bool
	ShowPersonaName bool
}

type PersonaFeatureSettings struct {
	VoiceCalls           bool
	VideoCalls           bool
	ScreenSharing        bool
	VoiceMessages        bool
	FileSharing          bool
	LocationSharing      bool
	DisappearingMessages bool
	ScheduledMessages    bool
	SilentMessages       bool

	MaxGroupSize         int
	MaxFileSize          int
	VoiceMessageDuration int
}

var PersonaLimits = map[string]struct {
	MaxPersonas      int
	CustomCategories bool
	BadgeSlots       int
}{
	"unverified": {MaxPersonas: 2, CustomCategories: false, BadgeSlots: 1},
	"newcomer":   {MaxPersonas: 3, CustomCategories: false, BadgeSlots: 2},
	"member":     {MaxPersonas: 5, CustomCategories: true, BadgeSlots: 3},
	"trusted":    {MaxPersonas: 7, CustomCategories: true, BadgeSlots: 5},
	"verified":   {MaxPersonas: 10, CustomCategories: true, BadgeSlots: -1},
}

func DefaultPrivacySettings() PersonaPrivacySettings {
	return PersonaPrivacySettings{
		LastSeenVisibility:       VisibilityContacts,
		OnlineStatusVisibility:   VisibilityContacts,
		ProfilePictureVisibility: VisibilityEveryone,
		BioVisibility:            VisibilityEveryone,
		StatusVisibility:         VisibilityContacts,

		WhoCanMessage:             VisibilityContacts,
		WhoCanCall:                VisibilityContacts,
		RequireApprovalForContact: false,

		SendReadReceipts:     true,
		SendTypingIndicators: true,

		Searchable:        true,
		ShowInSuggestions: true,

		AllowLinkingDiscovery:    false,
		ShowSharedTrustScore:     true,
		AllowCrossPersonaForward: false,
	}
}

func DefaultNotificationSettings() PersonaNotificationSettings {
	return PersonaNotificationSettings{
		Enabled:          true,
		Messages:         NotificationAll,
		Calls:            NotificationAll,
		GroupActivity:    NotificationMentions,
		ContactRequests:  true,
		SoundEnabled:     true,
		VibrationEnabled: true,
		ShowContent:      true,
		ShowSender:       true,
		ShowPersonaName:  true,
	}
}

func DefaultFeatureSettings() PersonaFeatureSettings {
	return PersonaFeatureSettings{
		VoiceCalls:           true,
		VideoCalls:           true,
		ScreenSharing:        true,
		VoiceMessages:        true,
		FileSharing:          true,
		LocationSharing:      false,
		DisappearingMessages: true,
		ScheduledMessages:    true,
		SilentMessages:       true,

		MaxGroupSize:         500,
		MaxFileSize:          100,
		VoiceMessageDuration: 120,
	}
}
