package identity

import "testing"

func TestRegisterUser(t *testing.T) {
	service := NewIdentityService()
	user, err := service.RegisterUser("hash123")
	if err != nil {
		t.Errorf("RegisterUser failed: %v", err)
	}
	if user.PhoneHash != "hash123" {
		t.Errorf("PhoneHash mismatch")
	}
	if user.VerificationLevel != 0 {
		t.Errorf("Initial verification should be 0")
	}
}

func TestGetUser(t *testing.T) {
	service := NewIdentityService()
	registered, _ := service.RegisterUser("hash123")
	retrieved, err := service.GetUser(registered.ID)
	if err != nil {
		t.Errorf("GetUser failed: %v", err)
	}
	if retrieved.ID != registered.ID {
		t.Errorf("User ID mismatch")
	}
}

func TestGetUserNotFound(t *testing.T) {
	service := NewIdentityService()
	_, err := service.GetUser("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Should return ErrUserNotFound")
	}
}
