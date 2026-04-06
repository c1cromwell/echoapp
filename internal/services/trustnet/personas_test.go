package trustnet

import (
	"testing"
)

func newPersonaTestSetup() (*PersonaService, *CircleService) {
	circles := NewCircleService()
	personas := NewPersonaService(circles)
	return personas, circles
}

func TestCreatePersona(t *testing.T) {
	t.Run("first persona becomes default", func(t *testing.T) {
		svc, _ := newPersonaTestSetup()
		p, err := svc.CreatePersona("user1", "Professional", "Alex Echo", "Product Lead", "briefcase", false)
		if err != nil {
			t.Fatalf("CreatePersona failed: %v", err)
		}
		if !p.IsDefault {
			t.Error("first persona should be default")
		}
	})

	t.Run("second persona not default", func(t *testing.T) {
		svc, _ := newPersonaTestSetup()
		svc.CreatePersona("user1", "Professional", "Alex Echo", "Product Lead", "briefcase", false)
		p, err := svc.CreatePersona("user1", "Personal", "Alex", "Just me", "house", false)
		if err != nil {
			t.Fatalf("CreatePersona failed: %v", err)
		}
		if p.IsDefault {
			t.Error("second persona should not be default")
		}
	})

	t.Run("duplicate name rejected", func(t *testing.T) {
		svc, _ := newPersonaTestSetup()
		svc.CreatePersona("user1", "Professional", "Alex Echo", "Lead", "briefcase", false)
		_, err := svc.CreatePersona("user1", "Professional", "Different Name", "Diff", "briefcase", false)
		if err != ErrPersonaDuplicate {
			t.Errorf("expected ErrPersonaDuplicate, got %v", err)
		}
	})

	t.Run("empty name rejected", func(t *testing.T) {
		svc, _ := newPersonaTestSetup()
		_, err := svc.CreatePersona("user1", "", "Alex", "Desc", "icon", false)
		if err != ErrPersonaNotFound {
			t.Errorf("expected error for empty name, got %v", err)
		}
	})
}

func TestGetUserPersonas(t *testing.T) {
	svc, _ := newPersonaTestSetup()
	svc.CreatePersona("user1", "Professional", "Alex Echo", "Lead", "briefcase", false)
	svc.CreatePersona("user1", "Personal", "Alex", "Just me", "house", false)
	svc.CreatePersona("user1", "Gaming", "NightOwl42", "Gaming", "controller", false)

	personas := svc.GetUserPersonas("user1")
	if len(personas) != 3 {
		t.Errorf("personas = %d, want 3", len(personas))
	}
}

func TestGetDefaultPersona(t *testing.T) {
	svc, _ := newPersonaTestSetup()
	svc.CreatePersona("user1", "Professional", "Alex Echo", "Lead", "briefcase", false)
	svc.CreatePersona("user1", "Personal", "Alex", "Just me", "house", false)

	def := svc.GetDefaultPersona("user1")
	if def == nil {
		t.Fatal("default persona should exist")
	}
	if def.Name != "Professional" {
		t.Errorf("default persona = %s, want Professional", def.Name)
	}
}

func TestRemovePersona(t *testing.T) {
	t.Run("remove non-default", func(t *testing.T) {
		svc, _ := newPersonaTestSetup()
		svc.CreatePersona("user1", "Professional", "Alex Echo", "Lead", "briefcase", false)
		p, _ := svc.CreatePersona("user1", "Gaming", "NightOwl", "Gaming", "controller", false)

		err := svc.RemovePersona(p.ID, "user1")
		if err != nil {
			t.Fatalf("RemovePersona failed: %v", err)
		}

		personas := svc.GetUserPersonas("user1")
		if len(personas) != 1 {
			t.Errorf("personas = %d, want 1", len(personas))
		}
	})

	t.Run("cannot remove default", func(t *testing.T) {
		svc, _ := newPersonaTestSetup()
		p, _ := svc.CreatePersona("user1", "Professional", "Alex Echo", "Lead", "briefcase", false)

		err := svc.RemovePersona(p.ID, "user1")
		if err != ErrPersonaDefaultOnly {
			t.Errorf("expected ErrPersonaDefaultOnly, got %v", err)
		}
	})

	t.Run("wrong user cannot remove", func(t *testing.T) {
		svc, _ := newPersonaTestSetup()
		svc.CreatePersona("user1", "Professional", "Alex", "Lead", "briefcase", false)
		p, _ := svc.CreatePersona("user1", "Gaming", "Night", "Gaming", "controller", false)

		err := svc.RemovePersona(p.ID, "user2")
		if err != ErrPersonaNotFound {
			t.Errorf("expected ErrPersonaNotFound, got %v", err)
		}
	})
}

func TestPersonaVisibility(t *testing.T) {
	svc, circles := newPersonaTestSetup()

	// Create personas for user1
	pro, _ := svc.CreatePersona("user1", "Professional", "Alex Echo", "Lead", "briefcase", true)
	personal, _ := svc.CreatePersona("user1", "Personal", "Alex", "Just me", "house", false)
	gaming, _ := svc.CreatePersona("user1", "Gaming", "NightOwl42", "Gaming", "controller", false)
	svc.CreatePersona("user1", "Family", "Dad", "Family", "family", false)

	t.Run("inner circle sees all personas", func(t *testing.T) {
		circles.AddContact("user1", "innerFriend", CircleInner)
		visible := svc.GetVisiblePersonas("user1", "innerFriend")
		if len(visible) != 4 {
			t.Errorf("inner should see all 4, got %d", len(visible))
		}
	})

	t.Run("trusted sees default plus selected", func(t *testing.T) {
		circles.AddContact("user1", "trustedFriend", CircleTrusted)
		// Grant access to Professional and Personal (Professional is default, already visible)
		svc.SetPersonaVisibility("user1", "trustedFriend", []string{pro.ID, personal.ID})

		visible := svc.GetVisiblePersonas("user1", "trustedFriend")
		if len(visible) != 2 {
			t.Errorf("trusted should see default + selected, got %d", len(visible))
		}

		// Verify it's the right personas
		names := make(map[string]bool)
		for _, p := range visible {
			names[p.Name] = true
		}
		if !names["Professional"] || !names["Personal"] {
			t.Error("should see Professional and Personal")
		}
		if names["Gaming"] || names["Family"] {
			t.Error("should not see Gaming or Family")
		}
	})

	t.Run("trusted with no explicit settings sees default only", func(t *testing.T) {
		circles.AddContact("user1", "trustedNoGrant", CircleTrusted)
		// No explicit visibility set
		visible := svc.GetVisiblePersonas("user1", "trustedNoGrant")
		if len(visible) != 1 {
			t.Errorf("trusted with no grants should see default only, got %d", len(visible))
		}
		if visible[0].Name != "Professional" {
			t.Errorf("should see default persona, got %s", visible[0].Name)
		}
	})

	t.Run("acquaintance sees default only", func(t *testing.T) {
		circles.AddContact("user1", "acquaintance1", CircleAcquaintance)
		visible := svc.GetVisiblePersonas("user1", "acquaintance1")
		if len(visible) != 1 {
			t.Errorf("acquaintance should see 1, got %d", len(visible))
		}
		if visible[0].Name != "Professional" {
			t.Errorf("should see default, got %s", visible[0].Name)
		}
	})

	t.Run("stranger sees nothing", func(t *testing.T) {
		visible := svc.GetVisiblePersonas("user1", "stranger")
		// Public tier has SeeDefaultPersona: false in spec, but we set it to true for minimal
		// Actually, per spec, public sees default persona only... let me check
		// Spec says Acquaintance = "Default only" and there's no public tier in spec
		// Our CirclePublic has SeeDefaultPersona: false, MinimalProfile: true
		if len(visible) != 0 {
			t.Errorf("stranger should see no personas, got %d", len(visible))
		}
	})

	t.Run("toggle persona visibility", func(t *testing.T) {
		circles.AddContact("user1", "toggleFriend", CircleTrusted)

		// Toggle gaming on
		err := svc.TogglePersonaForContact("user1", "toggleFriend", gaming.ID, true)
		if err != nil {
			t.Fatalf("TogglePersonaForContact failed: %v", err)
		}

		visible := svc.GetVisiblePersonas("user1", "toggleFriend")
		names := make(map[string]bool)
		for _, p := range visible {
			names[p.Name] = true
		}
		if !names["Gaming"] {
			t.Error("gaming should be visible after toggle on")
		}

		// Toggle it off
		svc.TogglePersonaForContact("user1", "toggleFriend", gaming.ID, false)
		visible = svc.GetVisiblePersonas("user1", "toggleFriend")
		names = make(map[string]bool)
		for _, p := range visible {
			names[p.Name] = true
		}
		if names["Gaming"] {
			t.Error("gaming should not be visible after toggle off")
		}
	})

	t.Run("toggle nonexistent persona fails", func(t *testing.T) {
		err := svc.TogglePersonaForContact("user1", "friend", "nonexistent", true)
		if err != ErrPersonaNotFound {
			t.Errorf("expected ErrPersonaNotFound, got %v", err)
		}
	})
}
