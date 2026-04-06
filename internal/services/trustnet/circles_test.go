package trustnet

import (
	"testing"
	"time"
)

func TestAddContact(t *testing.T) {
	t.Run("add to valid tier", func(t *testing.T) {
		svc := NewCircleService()
		c, err := svc.AddContact("user1", "contact1", CircleAcquaintance)
		if err != nil {
			t.Fatalf("AddContact failed: %v", err)
		}
		if c.Tier != CircleAcquaintance {
			t.Errorf("tier = %s, want acquaintance", c.Tier)
		}
		if c.UserDID != "user1" || c.ContactDID != "contact1" {
			t.Error("incorrect DIDs on contact")
		}
	})

	t.Run("cannot add self", func(t *testing.T) {
		svc := NewCircleService()
		_, err := svc.AddContact("user1", "user1", CircleAcquaintance)
		if err != ErrCannotAddSelf {
			t.Errorf("expected ErrCannotAddSelf, got %v", err)
		}
	})

	t.Run("invalid tier", func(t *testing.T) {
		svc := NewCircleService()
		_, err := svc.AddContact("user1", "contact1", CircleTier("invalid"))
		if err != ErrInvalidCircleTier {
			t.Errorf("expected ErrInvalidCircleTier, got %v", err)
		}
	})

	t.Run("duplicate same tier", func(t *testing.T) {
		svc := NewCircleService()
		svc.AddContact("user1", "contact1", CircleAcquaintance)
		_, err := svc.AddContact("user1", "contact1", CircleAcquaintance)
		if err != ErrContactAlreadyInCircle {
			t.Errorf("expected ErrContactAlreadyInCircle, got %v", err)
		}
	})

	t.Run("move to different tier", func(t *testing.T) {
		svc := NewCircleService()
		svc.AddContact("user1", "contact1", CircleAcquaintance)
		c, err := svc.AddContact("user1", "contact1", CircleTrusted)
		if err != nil {
			t.Fatalf("move tier failed: %v", err)
		}
		if c.Tier != CircleTrusted {
			t.Errorf("tier = %s, want trusted", c.Tier)
		}
		if c.PromotedAt == nil {
			t.Error("PromotedAt should be set")
		}
	})

	t.Run("circle capacity enforced", func(t *testing.T) {
		svc := NewCircleService()
		// Fill inner circle (limit 15)
		for i := 0; i < 15; i++ {
			_, err := svc.AddContact("user1", contactDID(i), CircleInner)
			if err != nil {
				t.Fatalf("add contact %d failed: %v", i, err)
			}
		}
		_, err := svc.AddContact("user1", "overflow", CircleInner)
		if err != ErrCircleFull {
			t.Errorf("expected ErrCircleFull, got %v", err)
		}
	})

	t.Run("public circle unlimited", func(t *testing.T) {
		svc := NewCircleService()
		for i := 0; i < 300; i++ {
			_, err := svc.AddContact("user1", contactDID(i), CirclePublic)
			if err != nil {
				t.Fatalf("add public contact %d failed: %v", i, err)
			}
		}
	})
}

func TestPromoteContact(t *testing.T) {
	t.Run("promote to higher tier", func(t *testing.T) {
		svc := NewCircleService()
		svc.AddContact("user1", "contact1", CircleAcquaintance)
		c, err := svc.PromoteContact("user1", "contact1", CircleTrusted)
		if err != nil {
			t.Fatalf("promote failed: %v", err)
		}
		if c.Tier != CircleTrusted {
			t.Errorf("tier = %s, want trusted", c.Tier)
		}
	})

	t.Run("promote unknown contact", func(t *testing.T) {
		svc := NewCircleService()
		_, err := svc.PromoteContact("user1", "unknown", CircleTrusted)
		if err != ErrContactNotFound {
			t.Errorf("expected ErrContactNotFound, got %v", err)
		}
	})

	t.Run("promote to full tier", func(t *testing.T) {
		svc := NewCircleService()
		for i := 0; i < 15; i++ {
			svc.AddContact("user1", contactDID(i), CircleInner)
		}
		svc.AddContact("user1", "overflow", CircleAcquaintance)
		_, err := svc.PromoteContact("user1", "overflow", CircleInner)
		if err != ErrCircleFull {
			t.Errorf("expected ErrCircleFull, got %v", err)
		}
	})
}

func TestRemoveContact(t *testing.T) {
	svc := NewCircleService()
	svc.AddContact("user1", "contact1", CircleAcquaintance)

	err := svc.RemoveContact("user1", "contact1")
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	_, err = svc.GetContact("user1", "contact1")
	if err != ErrContactNotFound {
		t.Error("contact should be removed")
	}
}

func TestRemoveContactNotFound(t *testing.T) {
	svc := NewCircleService()
	err := svc.RemoveContact("user1", "unknown")
	if err != ErrContactNotFound {
		t.Errorf("expected ErrContactNotFound, got %v", err)
	}
}

func TestGetCircle(t *testing.T) {
	svc := NewCircleService()
	svc.AddContact("user1", "c1", CircleAcquaintance)
	svc.AddContact("user1", "c2", CircleAcquaintance)
	svc.AddContact("user1", "c3", CircleTrusted)

	acquaintances := svc.GetCircle("user1", CircleAcquaintance)
	if len(acquaintances) != 2 {
		t.Errorf("acquaintance circle = %d, want 2", len(acquaintances))
	}

	trusted := svc.GetCircle("user1", CircleTrusted)
	if len(trusted) != 1 {
		t.Errorf("trusted circle = %d, want 1", len(trusted))
	}
}

func TestGetAllContacts(t *testing.T) {
	svc := NewCircleService()
	svc.AddContact("user1", "c1", CircleAcquaintance)
	svc.AddContact("user1", "c2", CircleTrusted)
	svc.AddContact("user1", "c3", CircleInner)

	all := svc.GetAllContacts("user1")
	if len(all) != 3 {
		t.Errorf("all contacts = %d, want 3", len(all))
	}
}

func TestGetPermissions(t *testing.T) {
	svc := NewCircleService()
	svc.AddContact("user1", "inner1", CircleInner)
	svc.AddContact("user1", "trusted1", CircleTrusted)
	svc.AddContact("user1", "acq1", CircleAcquaintance)

	t.Run("inner circle full permissions", func(t *testing.T) {
		p := svc.GetPermissions("user1", "inner1")
		if !p.SeeAllPersonas {
			t.Error("inner should see all personas")
		}
		if !p.SeeOnlineStatus || !p.SeeReadReceipts {
			t.Error("inner should see online status and read receipts")
		}
		if !p.DirectCall || !p.AddToGroups {
			t.Error("inner should have direct call and add to groups")
		}
		if !p.FullProfile {
			t.Error("inner should have full profile")
		}
	})

	t.Run("trusted permissions", func(t *testing.T) {
		p := svc.GetPermissions("user1", "trusted1")
		if p.SeeAllPersonas {
			t.Error("trusted should not see all personas")
		}
		if !p.SeeSelectedPersonas {
			t.Error("trusted should see selected personas")
		}
		if !p.SeeOnlineStatus || !p.SeeReadReceipts {
			t.Error("trusted should see online status and read receipts")
		}
		if p.AddToGroups {
			t.Error("trusted should not directly add to groups")
		}
		if !p.AskToAddToGroups {
			t.Error("trusted should ask to add to groups")
		}
	})

	t.Run("acquaintance limited permissions", func(t *testing.T) {
		p := svc.GetPermissions("user1", "acq1")
		if p.SeeAllPersonas || p.SeeSelectedPersonas {
			t.Error("acquaintance should only see default persona")
		}
		if !p.SeeDefaultPersona {
			t.Error("acquaintance should see default persona")
		}
		if p.SeeOnlineStatus || p.SeeReadReceipts {
			t.Error("acquaintance should not see online status or read receipts")
		}
		if p.DirectCall || p.AddToGroups || p.AskToAddToGroups {
			t.Error("acquaintance should not have call or group permissions")
		}
	})

	t.Run("stranger gets public permissions", func(t *testing.T) {
		p := svc.GetPermissions("user1", "stranger")
		if p.SeeOnlineStatus || p.DirectCall {
			t.Error("public should not have inner permissions")
		}
		if !p.MinimalProfile {
			t.Error("public should have minimal profile")
		}
	})
}

func TestRecordInteraction(t *testing.T) {
	svc := NewCircleService()
	svc.AddContact("user1", "contact1", CircleAcquaintance)

	svc.RecordInteraction("user1", "contact1", 10, 2)
	c, _ := svc.GetContact("user1", "contact1")
	if c.MessagesExchanged != 10 {
		t.Errorf("messages = %d, want 10", c.MessagesExchanged)
	}
	if c.CallCount != 2 {
		t.Errorf("calls = %d, want 2", c.CallCount)
	}
	if c.LastInteraction == nil {
		t.Error("LastInteraction should be set")
	}

	// Accumulates
	svc.RecordInteraction("user1", "contact1", 5, 1)
	c, _ = svc.GetContact("user1", "contact1")
	if c.MessagesExchanged != 15 {
		t.Errorf("messages = %d, want 15", c.MessagesExchanged)
	}
}

func TestCheckAutoPromotion(t *testing.T) {
	t.Run("acquaintance to trusted", func(t *testing.T) {
		svc := NewCircleService()
		svc.AddContact("user1", "contact1", CircleAcquaintance)

		svc.mu.Lock()
		c := svc.contacts["user1"]["contact1"]
		c.MessagesExchanged = 50
		c.CallCount = 5
		c.AddedAt = time.Now().Add(-31 * 24 * time.Hour)
		svc.mu.Unlock()

		tier, err := svc.CheckAutoPromotion("user1", "contact1")
		if err != nil {
			t.Fatalf("CheckAutoPromotion failed: %v", err)
		}
		if tier == nil || *tier != CircleTrusted {
			t.Error("should suggest promotion to trusted")
		}
	})

	t.Run("trusted to inner", func(t *testing.T) {
		svc := NewCircleService()
		svc.AddContact("user1", "contact1", CircleTrusted)

		svc.mu.Lock()
		c := svc.contacts["user1"]["contact1"]
		c.MessagesExchanged = 200
		c.CallCount = 20
		c.AddedAt = time.Now().Add(-91 * 24 * time.Hour)
		svc.mu.Unlock()

		tier, err := svc.CheckAutoPromotion("user1", "contact1")
		if err != nil {
			t.Fatalf("CheckAutoPromotion failed: %v", err)
		}
		if tier == nil || *tier != CircleInner {
			t.Error("should suggest promotion to inner")
		}
	})

	t.Run("not enough interaction", func(t *testing.T) {
		svc := NewCircleService()
		svc.AddContact("user1", "contact1", CircleAcquaintance)

		tier, err := svc.CheckAutoPromotion("user1", "contact1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tier != nil {
			t.Error("should not suggest promotion")
		}
	})

	t.Run("contact not found", func(t *testing.T) {
		svc := NewCircleService()
		_, err := svc.CheckAutoPromotion("user1", "unknown")
		if err != ErrContactNotFound {
			t.Errorf("expected ErrContactNotFound, got %v", err)
		}
	})
}

func TestCountCircle(t *testing.T) {
	svc := NewCircleService()
	svc.AddContact("user1", "c1", CircleAcquaintance)
	svc.AddContact("user1", "c2", CircleAcquaintance)
	svc.AddContact("user1", "c3", CircleTrusted)

	if n := svc.CountCircle("user1", CircleAcquaintance); n != 2 {
		t.Errorf("acquaintance count = %d, want 2", n)
	}
	if n := svc.CountCircle("user1", CircleTrusted); n != 1 {
		t.Errorf("trusted count = %d, want 1", n)
	}
	if n := svc.CountCircle("user1", CircleInner); n != 0 {
		t.Errorf("inner count = %d, want 0", n)
	}
}

func TestHasContact(t *testing.T) {
	svc := NewCircleService()
	svc.AddContact("user1", "contact1", CircleAcquaintance)

	if !svc.HasContact("user1", "contact1") {
		t.Error("should have contact1")
	}
	if svc.HasContact("user1", "unknown") {
		t.Error("should not have unknown")
	}
}

func TestMutualContacts(t *testing.T) {
	svc := NewCircleService()

	// user1 has contacts: A, B, C
	svc.AddContact("user1", "A", CircleAcquaintance)
	svc.AddContact("user1", "B", CircleTrusted)
	svc.AddContact("user1", "C", CircleInner)

	// user2 has contacts: B, C, D
	svc.AddContact("user2", "B", CircleAcquaintance)
	svc.AddContact("user2", "C", CircleTrusted)
	svc.AddContact("user2", "D", CircleInner)

	mutual := svc.GetMutualContacts("user1", "user2")
	if len(mutual) != 2 {
		t.Errorf("mutual = %d, want 2", len(mutual))
	}

	// Verify B and C are the mutuals
	mutualSet := make(map[string]bool)
	for _, did := range mutual {
		mutualSet[did] = true
	}
	if !mutualSet["B"] || !mutualSet["C"] {
		t.Error("mutual contacts should be B and C")
	}

	if n := svc.CountMutualContacts("user1", "user2"); n != 2 {
		t.Errorf("count mutual = %d, want 2", n)
	}

	// No mutual contacts
	if n := svc.CountMutualContacts("user1", "nobody"); n != 0 {
		t.Errorf("count mutual with nobody = %d, want 0", n)
	}
}

func contactDID(i int) string {
	return "contact-" + string(rune('a'+i%26)) + string(rune('0'+i/26))
}

func BenchmarkAddContact(b *testing.B) {
	svc := NewCircleService()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.AddContact("user1", contactDID(i), CirclePublic)
	}
}
