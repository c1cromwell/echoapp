package groups

import "time"

// GroupType represents the privacy level of a group
type GroupType string

const (
	GroupTypePublic  GroupType = "public"
	GroupTypePrivate GroupType = "private"
	GroupTypeSecret  GroupType = "secret"
)

// GroupRole represents a user's role within a group
type GroupRole string

const (
	GroupRoleOwner          GroupRole = "owner"
	GroupRoleAdmin          GroupRole = "admin"
	GroupRoleModerator      GroupRole = "moderator"
	GroupRoleVerifiedMember GroupRole = "verified_member"
	GroupRoleMember         GroupRole = "member"
	GroupRoleRestricted     GroupRole = "restricted"
	GroupRolePending        GroupRole = "pending"
)

// TrustLevel represents user's verification level
type TrustLevel string

const (
	TrustLevelUnverified TrustLevel = "unverified"
	TrustLevelNewcomer   TrustLevel = "newcomer"
	TrustLevelMember     TrustLevel = "member"
	TrustLevelTrusted    TrustLevel = "trusted"
	TrustLevelVerified   TrustLevel = "verified"
)

// GroupCategory represents group topic categories
type GroupCategory string

const (
	CategoryTechnology    GroupCategory = "technology"
	CategoryBusiness      GroupCategory = "business"
	CategoryFinance       GroupCategory = "finance"
	CategoryGaming        GroupCategory = "gaming"
	CategoryEntertainment GroupCategory = "entertainment"
	CategoryEducation     GroupCategory = "education"
	CategoryLocal         GroupCategory = "local"
	CategoryOther         GroupCategory = "other"
)

// ActivityLevel represents group activity
type ActivityLevel string

const (
	ActivityLevelLow      ActivityLevel = "low"
	ActivityLevelMedium   ActivityLevel = "medium"
	ActivityLevelHigh     ActivityLevel = "high"
	ActivityLevelVeryHigh ActivityLevel = "very_high"
)

// SecurityRating represents group security assessment
type SecurityRating string

const (
	SecurityRatingA SecurityRating = "A"
	SecurityRatingB SecurityRating = "B"
	SecurityRatingC SecurityRating = "C"
	SecurityRatingD SecurityRating = "D"
	SecurityRatingF SecurityRating = "F"
)

// ApprovalMode represents how members are approved
type ApprovalMode string

const (
	ApprovalModeAuto   ApprovalMode = "auto"
	ApprovalModeManual ApprovalMode = "manual"
	ApprovalModeVote   ApprovalMode = "vote"
)

// NotificationLevel represents user notification preferences
type NotificationLevel string

const (
	NotificationAll      NotificationLevel = "all"
	NotificationMentions NotificationLevel = "mentions"
	NotificationNone     NotificationLevel = "none"
)

// Permission represents an action that can be performed
type Permission string

const (
	PermissionPost           Permission = "post"
	PermissionComment        Permission = "comment"
	PermissionReact          Permission = "react"
	PermissionInvite         Permission = "invite"
	PermissionManageMembers  Permission = "manage_members"
	PermissionManageRoles    Permission = "manage_roles"
	PermissionDeleteMessages Permission = "delete_messages"
	PermissionModerate       Permission = "moderate"
	PermissionArchive        Permission = "archive"
	PermissionChangeSettings Permission = "change_settings"
	PermissionViewAnalytics  Permission = "view_analytics"
)

// Group represents a user group
type Group struct {
	// Identity
	GroupID string
	OwnerID string
	Type    GroupType

	// Profile
	Name        string
	Description string
	Avatar      string
	Category    GroupCategory
	Tags        []string
	Rules       string
	CreatedAt   time.Time

	// Verification requirements
	Requirements VerificationRequirements

	// Capacity
	MaxMembers     int
	CurrentMembers int
	MaxAdmins      int
	MaxModerators  int

	// Settings
	Settings GroupSettings

	// Permissions
	Permissions PermissionMatrix

	// Governance
	Governance GovernanceSettings

	// Statistics
	Stats GroupStatistics

	// Blockchain anchor
	CreationTxHash   string
	LastUpdateTxHash string
	SnapshotID       string
}

// VerificationRequirements defines entry criteria for the group
type VerificationRequirements struct {
	MinimumTrustScore   int
	MinimumTrustLevel   TrustLevel
	RequiredBadges      []string
	RequiredCredentials []string
	ApprovalMode        ApprovalMode
	CooldownPeriodHours int
}

// GroupSettings contains configurable group options
type GroupSettings struct {
	AllowInvites       bool
	AllowSearch        bool
	IsArchived         bool
	MessageRetention   int // days
	AllowFileSharing   bool
	AllowVoiceMessages bool
	AllowVideo         bool
}

// PermissionMatrix defines role-based permissions
type PermissionMatrix struct {
	Owner          []Permission
	Admin          []Permission
	Moderator      []Permission
	VerifiedMember []Permission
	Member         []Permission
	Restricted     []Permission
}

// GovernanceSettings defines voting and moderation rules
type GovernanceSettings struct {
	EnableVoting        bool
	VotingThreshold     float64 // percentage
	ModerationThreshold int     // number of reports
	AppealPeriodDays    int
	EnablePublicVoting  bool
}

// GroupStatistics contains aggregated group metrics
type GroupStatistics struct {
	VerifiedMemberPercentage float64
	AverageTrustScore        int
	ActivityLevel            ActivityLevel
	SecurityRating           SecurityRating
	MessageCount             int
	ActiveMembersToday       int
}

// GroupMember represents a user's membership in a group
type GroupMember struct {
	MemberID    string
	GroupID     string
	DisplayName string
	Avatar      string
	PersonaID   string

	// Role and permissions
	Role        GroupRole
	Permissions []Permission

	// Trust info
	TrustScore int
	TrustLevel TrustLevel
	Badges     []string
	VerifiedAt time.Time

	// Status
	JoinedAt     time.Time
	LastActiveAt time.Time
	MessageCount int
	WarningCount int
	IsMuted      bool
	MutedUntil   *time.Time
	IsBanned     bool

	// Settings
	NotificationLevel NotificationLevel
	Nickname          string
	ShowTrustScore    bool
}

// CreationLimits defines per-trust-level creation constraints
var CreationLimits = map[TrustLevel]struct {
	MaxGroupsOwned  int
	MaxGroupSize    int
	CanCreatePublic bool
}{
	TrustLevelUnverified: {MaxGroupsOwned: 0, MaxGroupSize: 0, CanCreatePublic: false},
	TrustLevelNewcomer:   {MaxGroupsOwned: 1, MaxGroupSize: 20, CanCreatePublic: false},
	TrustLevelMember:     {MaxGroupsOwned: 3, MaxGroupSize: 100, CanCreatePublic: true},
	TrustLevelTrusted:    {MaxGroupsOwned: 10, MaxGroupSize: 500, CanCreatePublic: true},
	TrustLevelVerified:   {MaxGroupsOwned: 25, MaxGroupSize: 5000, CanCreatePublic: true},
}

// DefaultPermissions provides default permissions per role
func DefaultPermissions(role GroupRole) []Permission {
	permissions := map[GroupRole][]Permission{
		GroupRoleOwner: {
			PermissionPost, PermissionComment, PermissionReact,
			PermissionInvite, PermissionManageMembers, PermissionManageRoles,
			PermissionDeleteMessages, PermissionModerate, PermissionArchive,
			PermissionChangeSettings, PermissionViewAnalytics,
		},
		GroupRoleAdmin: {
			PermissionPost, PermissionComment, PermissionReact,
			PermissionInvite, PermissionManageMembers, PermissionManageRoles,
			PermissionDeleteMessages, PermissionModerate, PermissionViewAnalytics,
		},
		GroupRoleModerator: {
			PermissionPost, PermissionComment, PermissionReact,
			PermissionInvite, PermissionDeleteMessages, PermissionModerate,
		},
		GroupRoleVerifiedMember: {
			PermissionPost, PermissionComment, PermissionReact, PermissionInvite,
		},
		GroupRoleMember: {
			PermissionPost, PermissionComment, PermissionReact,
		},
		GroupRoleRestricted: {},
		GroupRolePending:    {PermissionComment, PermissionReact},
	}
	return permissions[role]
}

// NewDefaultPermissionMatrix creates a default permission matrix
func NewDefaultPermissionMatrix() PermissionMatrix {
	return PermissionMatrix{
		Owner:          DefaultPermissions(GroupRoleOwner),
		Admin:          DefaultPermissions(GroupRoleAdmin),
		Moderator:      DefaultPermissions(GroupRoleModerator),
		VerifiedMember: DefaultPermissions(GroupRoleVerifiedMember),
		Member:         DefaultPermissions(GroupRoleMember),
		Restricted:     DefaultPermissions(GroupRoleRestricted),
	}
}
