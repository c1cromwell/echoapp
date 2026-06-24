package groups

import (
	"errors"
	"testing"
)

func TestAuthorizeAction_OnlyAdminsManageMembers(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{Name: "Test Group", MaxMembers: 100}
	requirements := VerificationRequirements{MinimumTrustScore: 0, ApprovalMode: ApprovalModeAuto}
	gs.CreateGroup("group_1", "owner", GroupTypePublic, profile, requirements, TrustLevelVerified)
	// A regular member (no manage-members permission).
	gs.AddMember("group_1", "member", 50, TrustLevelMember, false)

	// Owner may manage members.
	if err := gs.AuthorizeAction("group_1", "owner", PermissionManageMembers); err != nil {
		t.Fatalf("owner should be authorized to manage members: %v", err)
	}

	// A regular member may NOT.
	if err := gs.AuthorizeAction("group_1", "member", PermissionManageMembers); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("regular member must be unauthorized, got %v", err)
	}

	// A non-member may NOT (and we don't leak group internals).
	if err := gs.AuthorizeAction("group_1", "stranger", PermissionManageMembers); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("non-member must be unauthorized, got %v", err)
	}

	// Empty actor (unauthenticated) is unauthorized.
	if err := gs.AuthorizeAction("group_1", "", PermissionManageMembers); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("empty actor must be unauthorized, got %v", err)
	}
}

func TestAuthorizeAction_PromotedAdminCanManage(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{Name: "Test Group", MaxMembers: 100}
	requirements := VerificationRequirements{MinimumTrustScore: 0, ApprovalMode: ApprovalModeAuto}
	gs.CreateGroup("group_1", "owner", GroupTypePublic, profile, requirements, TrustLevelVerified)
	gs.AddMember("group_1", "member", 50, TrustLevelMember, false)

	// Promote to admin → now authorized.
	if _, err := gs.UpdateMemberRole("group_1", "member", GroupRoleAdmin); err != nil {
		t.Fatalf("promote failed: %v", err)
	}
	if err := gs.AuthorizeAction("group_1", "member", PermissionManageMembers); err != nil {
		t.Fatalf("promoted admin should be authorized: %v", err)
	}
}
