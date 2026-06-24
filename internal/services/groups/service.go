package groups

import (
	"context"
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

// GroupService manages group operations backed by a GroupStore.
type GroupService struct {
	store GroupStore
}

// GroupServiceOption configures GroupService construction.
type GroupServiceOption func(*GroupService)

// WithStore sets a custom persistence backend (default: in-memory).
func WithStore(store GroupStore) GroupServiceOption {
	return func(gs *GroupService) {
		gs.store = store
	}
}

// NewGroupService creates a group service. When Postgres is available, pass
// WithStore(NewPostgresStore(pool)) from main.
func NewGroupService(opts ...GroupServiceOption) *GroupService {
	gs := &GroupService{store: newMemoryStore()}
	for _, opt := range opts {
		opt(gs)
	}
	return gs
}

func (gs *GroupService) ctx() context.Context {
	return context.Background()
}

// CreateGroup creates a new group
func (gs *GroupService) CreateGroup(groupID, ownerID string, groupType GroupType, profile GroupProfile, requirements VerificationRequirements) (*Group, error) {
	exists, err := gs.store.GroupExists(gs.ctx(), groupID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("group %s already exists", groupID)
	}

	if groupType != GroupTypePublic && groupType != GroupTypePrivate && groupType != GroupTypeSecret {
		return nil, ErrInvalidGroupType
	}

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
		CurrentMembers: 1,
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
			ActivityLevel:  ActivityLevelLow,
			SecurityRating: SecurityRatingA,
		},
	}

	if err := gs.store.SaveGroup(gs.ctx(), group); err != nil {
		return nil, err
	}

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
	if err := gs.store.SaveMember(gs.ctx(), owner); err != nil {
		return nil, err
	}

	return group, nil
}

// GetGroup retrieves a group by ID
func (gs *GroupService) GetGroup(groupID string) (*Group, error) {
	return gs.store.GetGroup(gs.ctx(), groupID)
}

// AddMember adds a user to a group
func (gs *GroupService) AddMember(groupID, memberID string, trustScore int, trustLevel TrustLevel, isVerified bool) (*GroupMember, error) {
	group, err := gs.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	if _, err := gs.GetMember(groupID, memberID); err == nil {
		return nil, ErrAlreadyMember
	} else if !errors.Is(err, ErrMemberNotFound) {
		return nil, err
	}

	if group.CurrentMembers >= group.MaxMembers {
		return nil, ErrGroupFull
	}

	if trustScore < group.Requirements.MinimumTrustScore {
		return nil, ErrInsufficientTrustLevel
	}

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

	if err := gs.store.SaveMember(gs.ctx(), member); err != nil {
		return nil, err
	}

	group.CurrentMembers++
	if err := gs.store.SaveGroup(gs.ctx(), group); err != nil {
		return nil, err
	}

	return member, nil
}

// GetMember retrieves a group member
func (gs *GroupService) GetMember(groupID, memberID string) (*GroupMember, error) {
	return gs.store.GetMember(gs.ctx(), groupID, memberID)
}

// RemoveMember removes a user from a group
func (gs *GroupService) RemoveMember(groupID, memberID string) error {
	group, err := gs.GetGroup(groupID)
	if err != nil {
		return err
	}

	if err := gs.store.DeleteMember(gs.ctx(), groupID, memberID); err != nil {
		return err
	}

	if group.CurrentMembers > 0 {
		group.CurrentMembers--
	}
	return gs.store.SaveGroup(gs.ctx(), group)
}

// UpdateMemberRole updates a member's role
func (gs *GroupService) UpdateMemberRole(groupID, memberID string, newRole GroupRole) (*GroupMember, error) {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return nil, err
	}

	member.Role = newRole
	member.Permissions = DefaultPermissions(newRole)
	if err := gs.store.UpdateMember(gs.ctx(), member); err != nil {
		return nil, err
	}
	return member, nil
}

// GetGroupMembers retrieves all members of a group
func (gs *GroupService) GetGroupMembers(groupID string) ([]*GroupMember, error) {
	return gs.store.ListMembers(gs.ctx(), groupID)
}

// GroupMemberDIDs returns active member DIDs for WS group text fan-out (M2).
func (gs *GroupService) GroupMemberDIDs(groupID string) ([]string, error) {
	members, err := gs.GetGroupMembers(groupID)
	if err != nil {
		return nil, err
	}
	dids := make([]string, 0, len(members))
	for _, m := range members {
		if m.IsBanned {
			continue
		}
		dids = append(dids, m.MemberID)
	}
	return dids, nil
}

// IsGroupMember reports whether memberID belongs to groupID.
func (gs *GroupService) IsGroupMember(groupID, memberID string) (bool, error) {
	_, err := gs.GetMember(groupID, memberID)
	if err == ErrMemberNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
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

// AuthorizeAction returns ErrUnauthorized unless actorID holds `permission` in the
// group. Membership/role mutations (add, remove, role change, rekey) must be gated on
// this server-side so a non-admin can't drive them by calling the endpoint directly.
// A caller that is not a member of the group is treated as unauthorized.
func (gs *GroupService) AuthorizeAction(groupID, actorID string, permission Permission) error {
	if actorID == "" {
		return ErrUnauthorized
	}
	ok, err := gs.HasPermission(groupID, actorID, permission)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return ErrUnauthorized
		}
		return err
	}
	if !ok {
		return ErrUnauthorized
	}
	return nil
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

	return gs.store.UpdateMember(gs.ctx(), member)
}

// UnmuteUser unmutes a user
func (gs *GroupService) UnmuteUser(groupID, memberID string) error {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return err
	}

	member.IsMuted = false
	member.MutedUntil = nil

	return gs.store.UpdateMember(gs.ctx(), member)
}

// BanUser bans a user from the group
func (gs *GroupService) BanUser(groupID, memberID string) error {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return err
	}

	member.IsBanned = true
	return gs.store.UpdateMember(gs.ctx(), member)
}

// RecordWarning adds a warning to a member
func (gs *GroupService) RecordWarning(groupID, memberID string) (*GroupMember, error) {
	member, err := gs.GetMember(groupID, memberID)
	if err != nil {
		return nil, err
	}

	member.WarningCount++
	if err := gs.store.UpdateMember(gs.ctx(), member); err != nil {
		return nil, err
	}
	return member, nil
}
