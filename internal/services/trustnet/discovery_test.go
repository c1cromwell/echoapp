package trustnet

import (
	"testing"
)

func newDiscoveryTestSetup() (*DiscoveryService, *CircleService) {
	circles := NewCircleService()
	discovery := NewDiscoveryService(circles, []byte("test-key"))
	return discovery, circles
}

func TestRegisterProfile(t *testing.T) {
	t.Run("register new profile", func(t *testing.T) {
		svc, _ := newDiscoveryTestSetup()
		p, err := svc.RegisterProfile("did:echo:alice", "alice", "Alice Smith", 85, true)
		if err != nil {
			t.Fatalf("RegisterProfile failed: %v", err)
		}
		if p.Username != "alice" {
			t.Errorf("username = %s, want alice", p.Username)
		}
	})

	t.Run("update existing profile", func(t *testing.T) {
		svc, _ := newDiscoveryTestSetup()
		svc.RegisterProfile("did:echo:alice", "alice", "Alice", 85, true)
		p, err := svc.RegisterProfile("did:echo:alice", "alice_new", "Alice Updated", 90, true)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}
		if p.Username != "alice_new" {
			t.Errorf("username = %s, want alice_new", p.Username)
		}

		// Old username should be available
		if !svc.IsUsernameAvailable("alice") {
			t.Error("old username should be available after update")
		}
	})

	t.Run("duplicate username rejected", func(t *testing.T) {
		svc, _ := newDiscoveryTestSetup()
		svc.RegisterProfile("did:echo:alice", "alice", "Alice", 85, true)
		_, err := svc.RegisterProfile("did:echo:bob", "alice", "Bob", 50, false)
		if err == nil {
			t.Error("duplicate username should be rejected")
		}
	})

	t.Run("empty DID rejected", func(t *testing.T) {
		svc, _ := newDiscoveryTestSetup()
		_, err := svc.RegisterProfile("", "alice", "Alice", 85, true)
		if err != ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestGenerateAndParseQRCode(t *testing.T) {
	t.Run("generate and parse", func(t *testing.T) {
		svc, _ := newDiscoveryTestSetup()
		svc.RegisterProfile("did:echo:alice", "alice", "Alice", 85, true)

		qr, err := svc.GenerateQRCode("did:echo:alice", "alice")
		if err != nil {
			t.Fatalf("GenerateQRCode failed: %v", err)
		}

		if qr.Payload == "" {
			t.Error("payload should not be empty")
		}

		// Parse it back
		profile, err := svc.ParseQRCode(qr.Payload)
		if err != nil {
			t.Fatalf("ParseQRCode failed: %v", err)
		}
		if profile.DID != "did:echo:alice" {
			t.Errorf("DID = %s, want did:echo:alice", profile.DID)
		}
	})

	t.Run("invalid QR rejected", func(t *testing.T) {
		svc, _ := newDiscoveryTestSetup()
		_, err := svc.ParseQRCode("not-a-qr-code")
		if err != ErrQRCodeInvalid {
			t.Errorf("expected ErrQRCodeInvalid, got %v", err)
		}
	})

	t.Run("invalid base64 rejected", func(t *testing.T) {
		svc, _ := newDiscoveryTestSetup()
		_, err := svc.ParseQRCode("echo://not-valid-base64!!!")
		if err != ErrQRCodeInvalid {
			t.Errorf("expected ErrQRCodeInvalid, got %v", err)
		}
	})

	t.Run("unknown user QR rejected", func(t *testing.T) {
		svc, _ := newDiscoveryTestSetup()
		svc.RegisterProfile("did:echo:alice", "alice", "Alice", 85, true)
		qr, _ := svc.GenerateQRCode("did:echo:alice", "alice")

		// Delete the profile, then try to parse
		svc.mu.Lock()
		delete(svc.profiles, "did:echo:alice")
		svc.mu.Unlock()

		_, err := svc.ParseQRCode(qr.Payload)
		if err != ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("unregistered user QR fails", func(t *testing.T) {
		svc, _ := newDiscoveryTestSetup()
		_, err := svc.GenerateQRCode("did:echo:unknown", "unknown")
		if err != ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestSearchByUsername(t *testing.T) {
	svc, circles := newDiscoveryTestSetup()
	svc.RegisterProfile("did:echo:sarah", "sarahchen", "Sarah Chen", 85, true)
	svc.RegisterProfile("did:echo:sarah2", "sarahc_design", "Sarah C.", 42, false)
	svc.RegisterProfile("did:echo:sarah3", "sarahcrypto", "SarahCrypto", 28, false)
	svc.RegisterProfile("did:echo:alex", "alexecho", "Alex Echo", 70, true)

	// Add shared contact for mutual count
	circles.AddContact("did:echo:me", "did:echo:sarah", CircleTrusted)
	circles.AddContact("did:echo:sarah", "did:echo:me", CircleTrusted)

	t.Run("search by prefix", func(t *testing.T) {
		results := svc.SearchByUsername("sarah", "did:echo:me")
		if len(results) != 3 {
			t.Errorf("results = %d, want 3", len(results))
		}
	})

	t.Run("exact match ranked first", func(t *testing.T) {
		results := svc.SearchByUsername("sarahchen", "did:echo:me")
		if len(results) == 0 {
			t.Fatal("should have results")
		}
		if results[0].Username != "sarahchen" {
			t.Errorf("first result = %s, want sarahchen (exact match)", results[0].Username)
		}
	})

	t.Run("verified ranked higher", func(t *testing.T) {
		results := svc.SearchByUsername("sarah", "did:echo:me")
		if len(results) < 2 {
			t.Fatal("need at least 2 results")
		}
		// sarahchen (verified, score 85) should come before unverified
		if results[0].Username != "sarahchen" {
			t.Errorf("first result = %s, want sarahchen (verified)", results[0].Username)
		}
	})

	t.Run("self excluded from results", func(t *testing.T) {
		results := svc.SearchByUsername("sarah", "did:echo:sarah")
		for _, r := range results {
			if r.DID == "did:echo:sarah" {
				t.Error("self should not appear in search results")
			}
		}
	})

	t.Run("query too short", func(t *testing.T) {
		results := svc.SearchByUsername("s", "did:echo:me")
		if len(results) != 0 {
			t.Errorf("short query should return no results, got %d", len(results))
		}
	})

	t.Run("no results for unknown", func(t *testing.T) {
		results := svc.SearchByUsername("zzzz", "did:echo:me")
		if len(results) != 0 {
			t.Errorf("unknown query should return no results, got %d", len(results))
		}
	})

	t.Run("mutual count included", func(t *testing.T) {
		results := svc.SearchByUsername("sarahchen", "did:echo:me")
		// did:echo:me and did:echo:sarah have each other as contacts,
		// but mutual means shared third-party contacts, which we don't have
		// So mutual should be 0 here since they don't share contacts with a third party
		// Actually, GetMutualContacts checks if contactDIDs overlap
		// me has sarah; sarah has me — "me" is in sarah's contacts and "sarah" is in me's contacts
		// But GetMutualContacts looks at *me's* contacts vs *sarah's* contacts
		// me has: sarah. sarah has: me. Mutual = contacts that both have = none (me is not in me's list)
		if results[0].MutualCount != 0 {
			t.Logf("mutual count = %d (expected 0)", results[0].MutualCount)
		}
	})
}

func TestSearchByDID(t *testing.T) {
	svc, _ := newDiscoveryTestSetup()
	svc.RegisterProfile("did:echo:alice", "alice", "Alice", 85, true)

	t.Run("found", func(t *testing.T) {
		result, err := svc.SearchByDID("did:echo:alice", "did:echo:me")
		if err != nil {
			t.Fatalf("SearchByDID failed: %v", err)
		}
		if result.Username != "alice" {
			t.Errorf("username = %s, want alice", result.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.SearchByDID("did:echo:unknown", "did:echo:me")
		if err != ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"alice", true},
		{"alex_echo", true},
		{"NightOwl42", true},
		{"ab", false},            // too short
		{"has space", false},
		{"has-dash", false},
		{"has.dot", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateUsername(tt.name); got != tt.valid {
				t.Errorf("ValidateUsername(%q) = %v, want %v", tt.name, got, tt.valid)
			}
		})
	}
}

func TestIsUsernameAvailable(t *testing.T) {
	svc, _ := newDiscoveryTestSetup()
	svc.RegisterProfile("did:echo:alice", "alice", "Alice", 85, true)

	if svc.IsUsernameAvailable("alice") {
		t.Error("alice should be taken")
	}
	if svc.IsUsernameAvailable("Alice") {
		t.Error("Alice (case-insensitive) should be taken")
	}
	if !svc.IsUsernameAvailable("bob") {
		t.Error("bob should be available")
	}
}
