package groups

import "testing"

func TestMutualGroups(t *testing.T) {
	gs := NewGroupService()
	profile := GroupProfile{Name: "Echo Builders", MaxMembers: 50}
	req := VerificationRequirements{MinimumTrustScore: 0, ApprovalMode: ApprovalModeAuto}

	_, _ = gs.CreateGroup("grp_a", "did:key:alice", GroupTypePublic, profile, req, TrustLevelVerified)
	_, _ = gs.CreateGroup("grp_b", "did:key:bob", GroupTypePrivate, GroupProfile{Name: "VIP Lounge", MaxMembers: 10}, req, TrustLevelVerified)
	_, _ = gs.AddMember("grp_a", "did:key:bob", 50, TrustLevelMember, true)
	_, _ = gs.AddMember("grp_b", "did:key:alice", 50, TrustLevelMember, true)

	mutual := gs.MutualGroups("did:key:alice", "did:key:bob")
	if len(mutual) != 2 {
		t.Fatalf("expected 2 mutual groups, got %d", len(mutual))
	}

	none := gs.MutualGroups("did:key:alice", "did:key:stranger")
	if len(none) != 0 {
		t.Fatalf("expected 0 mutual groups with stranger, got %d", len(none))
	}
}
