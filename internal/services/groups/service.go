package groups

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrGroupNotFound          = errors.New("group not found")
	ErrMemberNotFound         = errors.New("member not found")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrGroupFull              = errors.New("group is full")
	ErrAlreadyMember          = errors.New("user is already a member")
	ErrInsufficientTrustLevel = errors.New("insufficient trust level")
	ErrExceedsCreationLimit   = errors.New("exceeds creation limit")
	ErrInvalidGroupType       = errors.New("invalid group type")
	ErrCooldownNotMet         = errors.New("member cooldown period not met")
)

// GroupService manages group operations
type GroupService struct {
	groups       map[string]*Group
	memberships  map[string][]*GroupMember // groupID -> members
	memberLookup map[string]*GroupMember   // memberID:groupID -> member
}

// NewGroupService creates a new group service
func NewGroupService() *GroupService {
	return &GroupService{
		groups:       make(map[string]*Group),
		memberships:  make(map[string][]*GroupMember),
		memberLookup: make(map[string]*GroupMember),
	}
}

// GroupProfile represents group profile information
type GroupProfile struct {
	Name          string
	Description   string
	Avatar        string
	Category      GroupCategory
	Tags          []string
	Rules         string
	MaxMembers    int
	MaxAdmins     int
	MaxModerators int
}

// CreateGroup creates a new group
func (gs *GroupService) CreateGroup(groupID, ownerID string, groupType GroupType, profile GroupProfile, requirements VerificationRequirements) (*Group, error) {
	// Validate group doesn't already exist
	if _, exists := gs.groups[groupID]; exists {
		return nil, fmt.Errorf("group %s already exists", groupID)
	}

	// Validate group type
	if groupType != GroupTypePublic && groupType != GroupTypePrivate && groupType != GroupTypeSecret {
		return nil, ErrInvalidGroupType
	}

	// Create group
	group := &Group{
		GroupID:        groupID,
		OwnerID:        ownerID,
		Type:           groupType,
		Name:           profile.Name,
		Description:    profile.Description,
		Avatar:         profile.Avatar,
		Category:       profile.Category,
		Tags:           profile.Tags,
		Rules:          profile.Rules,
		CreatedAt:      time.Now(),
		Requirements:   requirements,
		MaxMembers:     profile.MaxMembers,
		CurrentMembers: 0,
		MaxAdmins:      profile.MaxAdmins,
		MaxModerators:  profile.MaxModerators,
		Settings: GroupSettings{
			AllowInvites:       true,
			AllowSearch:        groupType == GroupTypePublic,
			AllowFileSharing:   true,
			AllowVoiceMessages: true,
			AllowVideo:         true,
		},
		Permissions: NewDefaultPermissionMatrix(),
		Governance: GovernanceSettings{
			EnableVoting:        true,
			VotingThreshold:     0.5,
			ModerationThreshold: 3,
			AppealPeriodDays:    7,
			EnablePublicVoting:  false,
		},
		Stats: GroupStatistics{
			VerifiedMemberPercentage: 0,
			AverageTrustScore:        0,
			ActivityLevel:            ActivityLevelLow,
			SecurityRating:           SecurityRatingA,
		},
	}

	gs.groups[groupID] = group
	gs.memberships[groupID] = make([]*GroupMember, 0)

	// Add owner as member
	owner := &GroupMember{
		MemberID:          ownerID,
		GroupID:           groupID,
		DisplayName:       "Owner",
		Role:              GroupRoleOwner,
		Permissions:       DefaultPermissions(GroupRoleOwner),
		TrustScore:        100,
		JoinedAt:          time.Now(),
		LastActiveAt:      time.Now(),
		NotificationLevel: NotificationAll,
		ShowTrustScore:    true,
	}

	gs.memberships[groupID] = append(gs.memberships[groupID], owner)
	gs.memberLookup[ownerID+":"+groupID] = owner
	group.CurrentMembers = 1

	return group, nil
}

// GetGroup retrieves a group by ID
func (gs *GroupService) GetGroup(groupID string) (*Group, error) {
	group, exists := gs.groups[groupID]
	if !exists {
		return nil, ErrGroupNotFound
	}
	return group, nil
}

// AddMember adds a user to a group
func (gs *GroupService) AddMember(groupID, memberID string, trustScore int, trustLevel TrustLevel, isVerified bool) (*GroupMember, error) {
	group, err := gs.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	// Check if user is already a member
	if _, err := gs.GetMember(groupID, memberID); err == nil {
		return nil, ErrAlreadyMember
	}

	// Check capacity
	if group.CurrentMembers >= group.MaxMembers {
		return nil, ErrGroupFull
	}

	// Check trust requirements
	if trustScore < group.Requirements.MinimumTrustScore {
		return nil, ErrInsufficientTrustLevel
	}

	// Determine role based on verification
	role := GroupRolePending
	if group.Requirements.ApprovalMode == ApprovalModeAuto {
		if isVerified && trustLevel != TrustLevelUnverified {
			role = GroupRoleVerifiedMember
		} else {
			role = GroupRoleMember
		}
	}

	member := &GroupMember{
		MemberID:          memberID,
		GroupID:           groupID,
		Role:              role,
		Permissions:       DefaultPermissions(role),
		TrustScore:        trustScore,
		TrustLevel:        trustLevel,
		JoinedAt:          time.Now(),
		LastActiveAt:      time.Now(),
		NotificationLevel: NotificationAll,
		ShowTrustScore:    !isVerified,
	}

	if isVerified {
		member.VerifiedAt = time.Now()
	}

	gs.memberships[groupID] = append(gs.memberships[groupID], member)
	gs.memberLookup[memberID+":"+groupID] = member
	group.CurrentMembers++

	return member, nil
}

// GetMember retrieves a group member
func (gs *GroupService) GetMember(groupID, memberID string) (*GroupMember, error) {
	key := memberID + ":" + groupID
	member, exists := gs.memberLookup[key]
	if !exists {
		return nil, ErrMemberNotFound
	}
	return member, nil
}

// RemoveMember removes a user from a group
func (gs *GroupService) RemoveMember(groupID, memberID string) error {
	group, err := gs.GetGroup(groupID)
	if err != nil {
		return err
	}

	members, exists := gs.memberships[groupID]
	if !exists {
		return ErrGroupNotFound
	}

	// Find and remove member
	foundIndex := -1
	for i, m := range members {
		if m.MemberID == memberID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return ErrMemberNotFound
	}

	// Remove from slice by creating a new slice without the element
	newMembers := make([]*GroupMember, 0, len(members)-1)
	for i, m := range members {
		if i != foundIndex {
			newMembers = append(newMembers, m)
		}
	}
	gs.memberships[groupID] = newMembers
	delete(gs.memberLookup, memberID+":"+groupID)
	group.CurrentMembers--

	return nil
}

// UpdateMemberRole updates a member's role
func (gs *GroupService) UpdateMemberRole(groupID, memberID string, newRole GroupRole) (*GroupMember, error) {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return nil, err
	}

	oldRole := member.Role
	member.Role = newRole
	member.Permissions = DefaultPermissions(newRole)

	// Update stored reference
	gs.memberLookup[memberID+":"+groupID] = member

	fmt.Printf("Member %s role changed from %s to %s\n", memberID, oldRole, newRole)

	return member, nil
}

// GetGroupMembers retrieves all members of a group
func (gs *GroupService) GetGroupMembers(groupID string) ([]*GroupMember, error) {
	members, exists := gs.memberships[groupID]
	if !exists {
		return nil, ErrGroupNotFound
	}
	return members, nil
}

// HasPermission checks if a member has a specific permission
func (gs *GroupService) HasPermission(groupID, memberID string, permission Permission) (bool, error) {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return false, err
	}

	for _, p := range member.Permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// MuteUser mutes a user for a specified duration
func (gs *GroupService) MuteUser(groupID, memberID string, duration time.Duration) error {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return err
	}

	member.IsMuted = true
	mutedUntil := time.Now().Add(duration)
	member.MutedUntil = &mutedUntil

	return nil
}

// UnmuteUser unmutes a user
func (gs *GroupService) UnmuteUser(groupID, memberID string) error {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return err
	}

	member.IsMuted = false
	member.MutedUntil = nil

	return nil
}

// BanUser bans a user from the group
func (gs *GroupService) BanUser(groupID, memberID string) error {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return err
	}

	member.IsBanned = true
	return nil
}

// RecordWarning adds a warning to a member
func (gs *GroupService) RecordWarning(groupID, memberID string) (*GroupMember, error) {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return nil, err
	}

	member.WarningCount++
	return member, nil
}
