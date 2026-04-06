package trustnet

import (
	"testing"
	"time"
)

func newProfileTestSetup() (*ContactProfileService, *CircleService, *PersonaService, *DiscoveryService) {
	circles := NewCircleService()
	personas := NewPersonaService(circles)
	discovery := NewDiscoveryService(circles, []byte("test-key"))
	profiles := NewContactProfileService(circles, personas, discovery)
	return profiles, circles, personas, discovery
}

func TestGetContactProfile(t *testing.T) {
	t.Run("full profile for inner circle contact", func(t *testing.T) {
		svc, circles, personas, discovery := newProfileTestSetup()

		// Set up contact (both directions, as accept flow does)
		discovery.RegisterProfile("did:echo:sarah", "sarahchen", "Sarah Chen", 85, true)
		circles.AddContact("did:echo:me", "did:echo:sarah", CircleInner)
		circles.AddContact("did:echo:sarah", "did:echo:me", CircleInner)

		// Set up personas for sarah
		personas.CreatePersona("did:echo:sarah", "Professional", "Sarah Chen", "Engineer", "briefcase", true)
		personas.CreatePersona("did:echo:sarah", "Personal", "Sarah", "Personal", "house", false)

		// Add verifications
		svc.AddVerification("did:echo:sarah", VerificationBadge{
			Type:       "identity",
			Label:      "Identity Verified",
			Detail:     "Government ID",
			VerifiedAt: time.Now(),
		})
		svc.AddVerification("did:echo:sarah", VerificationBadge{
			Type:       "phone",
			Label:      "Phone Verified",
			Detail:     "+1 ******* 4567",
			VerifiedAt: time.Now(),
		})

		// Get profile
		view, err := svc.GetContactProfile("did:echo:me", "did:echo:sarah")
		if err != nil {
			t.Fatalf("GetContactProfile failed: %v", err)
		}

		if view.DisplayName != "Sarah Chen" {
			t.Errorf("display name = %s, want Sarah Chen", view.DisplayName)
		}
		if view.TrustScore != 85 {
			t.Errorf("trust score = %d, want 85", view.TrustScore)
		}
		if view.CircleTier != CircleInner {
			t.Errorf("circle = %s, want inner", view.CircleTier)
		}
		if len(view.Verifications) != 2 {
			t.Errorf("verifications = %d, want 2", len(view.Verifications))
		}
		// Inner circle should see all personas
		if len(view.VisiblePersonas) != 2 {
			t.Errorf("visible personas = %d, want 2", len(view.VisiblePersonas))
		}
		if view.ContactSince == nil {
			t.Error("ContactSince should be set for contacts")
		}
	})

	t.Run("stranger profile", func(t *testing.T) {
		svc, _, _, discovery := newProfileTestSetup()
		discovery.RegisterProfile("did:echo:stranger", "stranger", "Stranger", 30, false)

		view, err := svc.GetContactProfile("did:echo:me", "did:echo:stranger")
		if err != nil {
			t.Fatalf("GetContactProfile failed: %v", err)
		}

		if view.CircleTier != CirclePublic {
			t.Errorf("circle = %s, want public", view.CircleTier)
		}
		if view.ContactSince != nil {
			t.Error("ContactSince should be nil for non-contacts")
		}
	})

	t.Run("unknown user", func(t *testing.T) {
		svc, _, _, _ := newProfileTestSetup()
		_, err := svc.GetContactProfile("did:echo:me", "did:echo:nonexistent")
		if err != ErrContactNotFound {
			t.Errorf("expected ErrContactNotFound, got %v", err)
		}
	})
}

func TestMuteContact(t *testing.T) {
	svc, _, _, discovery := newProfileTestSetup()
	discovery.RegisterProfile("did:echo:bob", "bob", "Bob", 50, false)

	svc.MuteContact("did:echo:me", "did:echo:bob")
	if !svc.IsMuted("did:echo:me", "did:echo:bob") {
		t.Error("should be muted")
	}

	svc.UnmuteContact("did:echo:me", "did:echo:bob")
	if svc.IsMuted("did:echo:me", "did:echo:bob") {
		t.Error("should be unmuted")
	}
}

func TestBlockContact(t *testing.T) {
	svc, _, _, discovery := newProfileTestSetup()
	discovery.RegisterProfile("did:echo:bob", "bob", "Bob", 50, false)

	svc.BlockContact("did:echo:me", "did:echo:bob", "spam")
	if !svc.IsBlocked("did:echo:me", "did:echo:bob") {
		t.Error("should be blocked")
	}

	// Check it shows in profile
	view, _ := svc.GetContactProfile("did:echo:me", "did:echo:bob")
	if !view.IsBlocked {
		t.Error("profile view should show blocked")
	}

	svc.UnblockContact("did:echo:me", "did:echo:bob")
	if svc.IsBlocked("did:echo:me", "did:echo:bob") {
		t.Error("should be unblocked")
	}
}

func TestMuteBlockNotContact(t *testing.T) {
	svc, _, _, _ := newProfileTestSetup()

	// Should not error even if not a registered user
	if svc.IsBlocked("a", "b") {
		t.Error("unknown pair should not be blocked")
	}
	if svc.IsMuted("a", "b") {
		t.Error("unknown pair should not be muted")
	}
}

func TestSharedGroups(t *testing.T) {
	svc, circles, _, discovery := newProfileTestSetup()
	discovery.RegisterProfile("did:echo:sarah", "sarah", "Sarah", 85, true)
	circles.AddContact("did:echo:me", "did:echo:sarah", CircleTrusted)

	svc.SetSharedGroups("did:echo:me", "did:echo:sarah", []string{"Product Team", "Design Guild"})

	view, _ := svc.GetContactProfile("did:echo:me", "did:echo:sarah")
	if len(view.SharedGroups) != 2 {
		t.Errorf("shared groups = %d, want 2", len(view.SharedGroups))
	}
}

func TestGetVerifications(t *testing.T) {
	svc, _, _, _ := newProfileTestSetup()

	svc.AddVerification("user1", VerificationBadge{Type: "identity", Label: "ID Verified"})
	svc.AddVerification("user1", VerificationBadge{Type: "phone", Label: "Phone Verified"})

	badges := svc.GetVerifications("user1")
	if len(badges) != 2 {
		t.Errorf("badges = %d, want 2", len(badges))
	}

	// Verify it's a copy
	badges[0].Label = "modified"
	original := svc.GetVerifications("user1")
	if original[0].Label == "modified" {
		t.Error("GetVerifications should return a copy")
	}
}

func TestMutualCountInProfile(t *testing.T) {
	svc, circles, _, discovery := newProfileTestSetup()
	discovery.RegisterProfile("did:echo:sarah", "sarah", "Sarah", 85, true)

	// Both have charlie as a contact
	circles.AddContact("did:echo:me", "charlie", CircleTrusted)
	circles.AddContact("did:echo:sarah", "charlie", CircleAcquaintance)

	view, _ := svc.GetContactProfile("did:echo:me", "did:echo:sarah")
	if view.MutualCount != 1 {
		t.Errorf("mutual count = %d, want 1", view.MutualCount)
	}
}
