package groups

import (
	"testing"
	"time"
)

func TestCreateGroup(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Tech Group",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 30,
		ApprovalMode:      ApprovalModeAuto,
	}

	group, err := gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	if group.GroupID != "group_1" {
		t.Errorf("Expected group_1, got %s", group.GroupID)
	}
	if group.CurrentMembers != 1 {
		t.Errorf("Expected 1 member, got %d", group.CurrentMembers)
	}
}

func TestAddMember(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Test Group",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 30,
		ApprovalMode:      ApprovalModeAuto,
	}
	gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)

	member, err := gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}
	if member.Role != GroupRoleMember {
		t.Errorf("Expected member role, got %s", member.Role)
	}
}

func TestRemoveMember(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Test Group",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 0,
		ApprovalMode:      ApprovalModeAuto,
	}
	gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)
	gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)

	err := gs.RemoveMember("group_1", "user_2")
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}

	_, err = gs.GetMember("group_1", "user_2")
	if err != ErrMemberNotFound {
		t.Error("Expected member to be removed")
	}
}

func TestUpdateMemberRole(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Test Group",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 0,
		ApprovalMode:      ApprovalModeAuto,
	}
	gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)
	gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)

	member, err := gs.UpdateMemberRole("group_1", "user_2", GroupRoleAdmin)
	if err != nil {
		t.Fatalf("UpdateMemberRole failed: %v", err)
	}
	if member.Role != GroupRoleAdmin {
		t.Errorf("Expected admin, got %s", member.Role)
	}
}

func TestHasPermission(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Test Group",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 0,
		ApprovalMode:      ApprovalModeAuto,
	}
	gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)
	gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)

	hasPost, _ := gs.HasPermission("group_1", "user_2", PermissionPost)
	if !hasPost {
		t.Error("Expected post permission")
	}

	hasManage, _ := gs.HasPermission("group_1", "user_2", PermissionManageRoles)
	if hasManage {
		t.Error("Did not expect manage_roles permission")
	}
}

func TestMuteAndUnmuteUser(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Test Group",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 0,
		ApprovalMode:      ApprovalModeAuto,
	}
	gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)
	gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)

	gs.MuteUser("group_1", "user_2", 1*time.Hour)
	member, _ := gs.GetMember("group_1", "user_2")
	if !member.IsMuted {
		t.Error("Expected user to be muted")
	}

	gs.UnmuteUser("group_1", "user_2")
	member, _ = gs.GetMember("group_1", "user_2")
	if member.IsMuted {
		t.Error("Expected user to be unmuted")
	}
}

func TestBanUser(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Test Group",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 0,
		ApprovalMode:      ApprovalModeAuto,
	}
	gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)
	gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)

	gs.BanUser("group_1", "user_2")
	member, _ := gs.GetMember("group_1", "user_2")
	if !member.IsBanned {
		t.Error("Expected user to be banned")
	}
}

func TestRecordWarning(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Test Group",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 0,
		ApprovalMode:      ApprovalModeAuto,
	}
	gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)
	gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)

	member, err := gs.RecordWarning("group_1", "user_2")
	if err != nil {
		t.Fatalf("RecordWarning failed: %v", err)
	}
	if member.WarningCount != 1 {
		t.Errorf("Expected 1 warning, got %d", member.WarningCount)
	}
}

func TestGetGroupMembers(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Test Group",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 0,
		ApprovalMode:      ApprovalModeAuto,
	}
	gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)
	gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)
	gs.AddMember("group_1", "user_3", 60, TrustLevelMember, false)

	members, err := gs.GetGroupMembers("group_1")
	if err != nil {
		t.Fatalf("GetGroupMembers failed: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}
}

func TestCreationLimits(t *testing.T) {
	limits := CreationLimits[TrustLevelMember]
	if limits.MaxGroupsOwned != 3 {
		t.Errorf("Expected 3 max groups, got %d", limits.MaxGroupsOwned)
	}
	if limits.MaxGroupSize != 100 {
		t.Errorf("Expected max size 100, got %d", limits.MaxGroupSize)
	}
	if !limits.CanCreatePublic {
		t.Error("Expected to create public groups")
	}
}

func TestDefaultPermissions(t *testing.T) {
	ownerPerms := DefaultPermissions(GroupRoleOwner)
	if len(ownerPerms) == 0 {
		t.Error("Owner should have permissions")
	}

	restrictedPerms := DefaultPermissions(GroupRoleRestricted)
	if len(restrictedPerms) != 0 {
		t.Error("Restricted should have no permissions")
	}
}

func TestErrorCases(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{
		Name:       "Test",
		MaxMembers: 100,
	}
	requirements := VerificationRequirements{
		MinimumTrustScore: 50,
		ApprovalMode:      ApprovalModeAuto,
	}
	gs.CreateGroup("group_1", "user_1", GroupTypePublic, profile, requirements)

	// Test adding duplicate member
	gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)
	_, err := gs.AddMember("group_1", "user_2", 50, TrustLevelMember, false)
	if err != ErrAlreadyMember {
		t.Errorf("Expected ErrAlreadyMember, got %v", err)
	}

	// Test insufficient trust
	_, err = gs.AddMember("group_1", "user_3", 30, TrustLevelMember, false)
	if err != ErrInsufficientTrustLevel {
		t.Errorf("Expected ErrInsufficientTrustLevel, got %v", err)
	}

	// Test group full with MaxMembers=1
	gsSmall := NewGroupService()
	profileSmall := GroupProfile{
		Name:       "Small Group",
		MaxMembers: 1,
	}
	requirementsSmall := VerificationRequirements{
		MinimumTrustScore: 0,
		ApprovalMode:      ApprovalModeAuto,
	}
	gsSmall.CreateGroup("small_group", "user_1", GroupTypePublic, profileSmall, requirementsSmall)
	_, err = gsSmall.AddMember("small_group", "user_2", 50, TrustLevelMember, false)
	if err != ErrGroupFull {
		t.Errorf("Expected ErrGroupFull, got %v", err)
	}
}
