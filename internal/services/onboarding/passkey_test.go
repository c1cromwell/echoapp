package onboarding

import (
	"testing"
)

func TestCreateChallenge(t *testing.T) {
	svc := NewPasskeyService()

	t.Run("valid challenge", func(t *testing.T) {
		c, err := svc.CreateChallenge("user1")
		if err != nil {
			t.Fatalf("CreateChallenge failed: %v", err)
		}
		if c.ID == "" {
			t.Error("challenge ID is empty")
		}
		if c.UserID != "user1" {
			t.Errorf("userID = %s, want user1", c.UserID)
		}
		if len(c.Challenge) != 32 {
			t.Errorf("challenge length = %d, want 32", len(c.Challenge))
		}
		if c.Used {
			t.Error("challenge should not be used initially")
		}
	})

	t.Run("empty user ID", func(t *testing.T) {
		_, err := svc.CreateChallenge("")
		if err != ErrPasskeyInvalidData {
			t.Errorf("expected ErrPasskeyInvalidData, got %v", err)
		}
	})

	t.Run("unique challenges", func(t *testing.T) {
		c1, _ := svc.CreateChallenge("user2")
		c2, _ := svc.CreateChallenge("user3")
		if c1.ID == c2.ID {
			t.Error("challenges should have unique IDs")
		}
	})
}

func TestRegisterCredential(t *testing.T) {
	svc := NewPasskeyService()

	t.Run("valid registration", func(t *testing.T) {
		c, _ := svc.CreateChallenge("user1")
		pubKey := []byte("test-public-key-data-32bytes!!!!!")

		cred, err := svc.RegisterCredential(c.ID, "cred-123", pubKey, PasskeyFaceID, "iPhone 15 Pro")
		if err != nil {
			t.Fatalf("RegisterCredential failed: %v", err)
		}
		if cred.CredentialID != "cred-123" {
			t.Errorf("credentialID = %s, want cred-123", cred.CredentialID)
		}
		if cred.PasskeyType != PasskeyFaceID {
			t.Errorf("type = %s, want face_id", cred.PasskeyType)
		}
		if cred.DeviceInfo != "iPhone 15 Pro" {
			t.Errorf("device = %s, want iPhone 15 Pro", cred.DeviceInfo)
		}
	})

	t.Run("invalid challenge ID", func(t *testing.T) {
		_, err := svc.RegisterCredential("bad-challenge", "cred-456", []byte("key"), PasskeyFaceID, "")
		if err != ErrPasskeyInvalidData {
			t.Errorf("expected ErrPasskeyInvalidData, got %v", err)
		}
	})

	t.Run("challenge already used", func(t *testing.T) {
		svc := NewPasskeyService()
		c, _ := svc.CreateChallenge("user2")
		svc.RegisterCredential(c.ID, "cred-first", []byte("key1"), PasskeyFaceID, "")

		_, err := svc.RegisterCredential(c.ID, "cred-second", []byte("key2"), PasskeyFaceID, "")
		if err != ErrPasskeyAlreadyExists {
			t.Errorf("expected ErrPasskeyAlreadyExists, got %v", err)
		}
	})

	t.Run("empty public key", func(t *testing.T) {
		svc := NewPasskeyService()
		c, _ := svc.CreateChallenge("user3")
		_, err := svc.RegisterCredential(c.ID, "cred-789", []byte{}, PasskeyFaceID, "")
		if err != ErrPasskeyInvalidData {
			t.Errorf("expected ErrPasskeyInvalidData, got %v", err)
		}
	})

	t.Run("empty credential ID", func(t *testing.T) {
		svc := NewPasskeyService()
		c, _ := svc.CreateChallenge("user4")
		_, err := svc.RegisterCredential(c.ID, "", []byte("key"), PasskeyFaceID, "")
		if err != ErrPasskeyInvalidData {
			t.Errorf("expected ErrPasskeyInvalidData, got %v", err)
		}
	})

	t.Run("user already has passkey", func(t *testing.T) {
		svc := NewPasskeyService()
		c1, _ := svc.CreateChallenge("user5")
		svc.RegisterCredential(c1.ID, "cred-a", []byte("key-a"), PasskeyFaceID, "")

		c2, _ := svc.CreateChallenge("user5")
		_, err := svc.RegisterCredential(c2.ID, "cred-b", []byte("key-b"), PasskeyTouchID, "")
		if err != ErrPasskeyAlreadyExists {
			t.Errorf("expected ErrPasskeyAlreadyExists, got %v", err)
		}
	})
}

func TestGetCredential(t *testing.T) {
	svc := NewPasskeyService()
	c, _ := svc.CreateChallenge("user1")
	svc.RegisterCredential(c.ID, "cred-lookup", []byte("key"), PasskeyFaceID, "")

	t.Run("existing credential", func(t *testing.T) {
		cred, err := svc.GetCredential("cred-lookup")
		if err != nil {
			t.Fatalf("GetCredential failed: %v", err)
		}
		if cred.CredentialID != "cred-lookup" {
			t.Errorf("credentialID = %s, want cred-lookup", cred.CredentialID)
		}
	})

	t.Run("nonexistent credential", func(t *testing.T) {
		_, err := svc.GetCredential("nonexistent")
		if err != ErrPasskeyInvalidData {
			t.Errorf("expected ErrPasskeyInvalidData, got %v", err)
		}
	})
}

func TestHasPasskey(t *testing.T) {
	svc := NewPasskeyService()

	if svc.HasPasskey("user1") {
		t.Error("user without passkey should return false")
	}

	c, _ := svc.CreateChallenge("user1")
	svc.RegisterCredential(c.ID, "cred-1", []byte("key"), PasskeyFaceID, "")

	if !svc.HasPasskey("user1") {
		t.Error("user with passkey should return true")
	}
}

func TestGetUserCredentials(t *testing.T) {
	svc := NewPasskeyService()

	// No credentials yet
	creds := svc.GetUserCredentials("user1")
	if len(creds) != 0 {
		t.Errorf("expected 0 credentials, got %d", len(creds))
	}

	c, _ := svc.CreateChallenge("user1")
	svc.RegisterCredential(c.ID, "cred-1", []byte("key"), PasskeyFaceID, "Device1")

	creds = svc.GetUserCredentials("user1")
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(creds))
	}
	if creds[0].DeviceInfo != "Device1" {
		t.Errorf("device = %s, want Device1", creds[0].DeviceInfo)
	}
}

func TestGenerateDID(t *testing.T) {
	did := GenerateDID([]byte("test-public-key"))

	if did == "" {
		t.Error("DID should not be empty")
	}
	if len(did) < 10 {
		t.Error("DID is too short")
	}
	if did[:9] != "did:echo:" {
		t.Errorf("DID prefix = %s, want did:echo:", did[:9])
	}

	// Same key should produce same DID
	did2 := GenerateDID([]byte("test-public-key"))
	if did != did2 {
		t.Error("same key should produce same DID")
	}

	// Different key should produce different DID
	did3 := GenerateDID([]byte("different-key"))
	if did == did3 {
		t.Error("different keys should produce different DIDs")
	}
}

func TestPasskeyTypes(t *testing.T) {
	types := []PasskeyType{PasskeyFaceID, PasskeyTouchID, PasskeyFingerprint, PasskeyPIN}
	for _, pt := range types {
		if pt == "" {
			t.Error("passkey type should not be empty")
		}
	}
}
