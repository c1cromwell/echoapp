package onboarding

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name+tag@domain.co.uk", true},
		{"a@b.cc", true},
		{"", false},
		{"notanemail", false},
		{"@domain.com", false},
		{"user@", false},
		{"user@.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if result := ValidateEmail(tt.email); result != tt.valid {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, result, tt.valid)
			}
		})
	}
}

func TestAddEmail(t *testing.T) {
	svc := NewRecoveryService()

	t.Run("valid email", func(t *testing.T) {
		m, err := svc.AddEmail("user1", "user@example.com")
		if err != nil {
			t.Fatalf("AddEmail failed: %v", err)
		}
		if m.Type != RecoveryEmail {
			t.Errorf("type = %s, want email", m.Type)
		}
		if m.Value != "user@example.com" {
			t.Errorf("value = %s, want user@example.com", m.Value)
		}
		if m.Verified {
			t.Error("email should not be verified initially")
		}
		if !m.Primary {
			t.Error("first method should be primary")
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		_, err := svc.AddEmail("user1", "invalid")
		if err != ErrRecoveryEmailInvalid {
			t.Errorf("expected ErrRecoveryEmailInvalid, got %v", err)
		}
	})

	t.Run("empty user ID", func(t *testing.T) {
		_, err := svc.AddEmail("", "user@example.com")
		if err != ErrRecoveryEmailInvalid {
			t.Errorf("expected ErrRecoveryEmailInvalid, got %v", err)
		}
	})

	t.Run("second method not primary", func(t *testing.T) {
		svc := NewRecoveryService()
		svc.AddEmail("user2", "first@example.com")
		m, _ := svc.AddEmail("user2", "second@example.com")
		if m.Primary {
			t.Error("second method should not be primary")
		}
	})
}

func TestAddWallet(t *testing.T) {
	svc := NewRecoveryService()

	t.Run("valid wallet", func(t *testing.T) {
		m, err := svc.AddWallet("user1", "0x1234567890abcdef1234567890abcdef12345678")
		if err != nil {
			t.Fatalf("AddWallet failed: %v", err)
		}
		if m.Type != RecoveryWallet {
			t.Errorf("type = %s, want wallet", m.Type)
		}
		if !m.Verified {
			t.Error("wallet should be verified immediately")
		}
		if m.VerifiedAt == nil {
			t.Error("verifiedAt should be set")
		}
	})

	t.Run("empty wallet", func(t *testing.T) {
		_, err := svc.AddWallet("user1", "")
		if err != ErrRecoveryWalletInvalid {
			t.Errorf("expected ErrRecoveryWalletInvalid, got %v", err)
		}
	})

	t.Run("short wallet address", func(t *testing.T) {
		_, err := svc.AddWallet("user1", "0x123")
		if err != ErrRecoveryWalletInvalid {
			t.Errorf("expected ErrRecoveryWalletInvalid, got %v", err)
		}
	})
}

func TestAddTrustedContact(t *testing.T) {
	svc := NewRecoveryService()

	t.Run("valid contact", func(t *testing.T) {
		m, err := svc.AddTrustedContact("user1", "did:echo:contact123")
		if err != nil {
			t.Fatalf("AddTrustedContact failed: %v", err)
		}
		if m.Type != RecoveryContact {
			t.Errorf("type = %s, want trusted_contact", m.Type)
		}
		if m.Verified {
			t.Error("contact should not be verified initially")
		}
	})

	t.Run("empty contact", func(t *testing.T) {
		_, err := svc.AddTrustedContact("user1", "")
		if err != ErrRecoveryContactInvalid {
			t.Errorf("expected ErrRecoveryContactInvalid, got %v", err)
		}
	})

	t.Run("empty user", func(t *testing.T) {
		_, err := svc.AddTrustedContact("", "did:echo:contact123")
		if err != ErrRecoveryContactInvalid {
			t.Errorf("expected ErrRecoveryContactInvalid, got %v", err)
		}
	})
}

func TestVerifyMethod(t *testing.T) {
	svc := NewRecoveryService()

	t.Run("verify email", func(t *testing.T) {
		m, _ := svc.AddEmail("user1", "user@example.com")
		err := svc.VerifyMethod(m.ID)
		if err != nil {
			t.Fatalf("VerifyMethod failed: %v", err)
		}

		retrieved, _ := svc.GetMethod(m.ID)
		if !retrieved.Verified {
			t.Error("method should be verified")
		}
		if retrieved.VerifiedAt == nil {
			t.Error("verifiedAt should be set")
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		err := svc.VerifyMethod("nonexistent")
		if err != ErrRecoveryMethodInvalid {
			t.Errorf("expected ErrRecoveryMethodInvalid, got %v", err)
		}
	})
}

func TestGetUserMethods(t *testing.T) {
	svc := NewRecoveryService()

	// No methods initially
	methods := svc.GetUserMethods("user1")
	if len(methods) != 0 {
		t.Errorf("expected 0 methods, got %d", len(methods))
	}

	svc.AddEmail("user1", "email@test.com")
	svc.AddWallet("user1", "0x1234567890abcdef1234567890abcdef12345678")

	methods = svc.GetUserMethods("user1")
	if len(methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(methods))
	}
}

func TestRemoveMethod(t *testing.T) {
	svc := NewRecoveryService()
	m, _ := svc.AddEmail("user1", "remove@test.com")

	t.Run("valid remove", func(t *testing.T) {
		err := svc.RemoveMethod(m.ID, "user1")
		if err != nil {
			t.Fatalf("RemoveMethod failed: %v", err)
		}
		methods := svc.GetUserMethods("user1")
		if len(methods) != 0 {
			t.Errorf("expected 0 methods after remove, got %d", len(methods))
		}
	})

	t.Run("wrong user", func(t *testing.T) {
		m2, _ := svc.AddEmail("user2", "other@test.com")
		err := svc.RemoveMethod(m2.ID, "user1")
		if err != ErrRecoveryMethodInvalid {
			t.Errorf("expected ErrRecoveryMethodInvalid, got %v", err)
		}
	})

	t.Run("nonexistent method", func(t *testing.T) {
		err := svc.RemoveMethod("nonexistent", "user1")
		if err != ErrRecoveryMethodInvalid {
			t.Errorf("expected ErrRecoveryMethodInvalid, got %v", err)
		}
	})
}

func TestHasRecoveryMethod(t *testing.T) {
	svc := NewRecoveryService()

	if svc.HasRecoveryMethod("user1") {
		t.Error("should have no methods initially")
	}

	svc.AddEmail("user1", "user@test.com")
	if !svc.HasRecoveryMethod("user1") {
		t.Error("should have method after adding")
	}
}

func TestHasVerifiedRecoveryMethod(t *testing.T) {
	svc := NewRecoveryService()

	if svc.HasVerifiedRecoveryMethod("user1") {
		t.Error("should have no verified methods initially")
	}

	m, _ := svc.AddEmail("user1", "user@test.com")
	if svc.HasVerifiedRecoveryMethod("user1") {
		t.Error("unverified email should not count")
	}

	svc.VerifyMethod(m.ID)
	if !svc.HasVerifiedRecoveryMethod("user1") {
		t.Error("should have verified method after verification")
	}
}

func TestWalletAutoVerified(t *testing.T) {
	svc := NewRecoveryService()
	svc.AddWallet("user1", "0x1234567890abcdef1234567890abcdef12345678")

	if !svc.HasVerifiedRecoveryMethod("user1") {
		t.Error("wallet should be auto-verified")
	}
}
