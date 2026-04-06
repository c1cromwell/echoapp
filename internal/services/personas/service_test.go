package personas

import (
	"strings"
	"testing"
	"time"
)

func TestCreateMasterIdentity(t *testing.T) {
	ps := NewPersonaService()

	master, err := ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key_123", "trusted")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if master.UserID != "user1" {
		t.Errorf("expected userID user1, got %s", master.UserID)
	}
	if master.DID != "did:echo:user1" {
		t.Errorf("expected DID did:echo:user1, got %s", master.DID)
	}

	_, err = ps.CreateMasterIdentity("user1", "did:echo:user1_dup", "key_dup", "trusted")
	if err == nil {
		t.Error("expected error for duplicate master identity, got none")
	}
}

func TestCreatePersona(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")

	persona, err := ps.CreatePersona(
		"user1",
		"John Professional",
		"jprofessional",
		"avatar.jpg",
		"I'm a professional",
		CategoryProfessional,
		"trusted",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if persona.DisplayName != "John Professional" {
		t.Errorf("expected display name 'John Professional', got %s", persona.DisplayName)
	}
	if persona.Category != CategoryProfessional {
		t.Errorf("expected category professional, got %s", persona.Category)
	}

	for i := 1; i < 7; i++ {
		_, err := ps.CreatePersona(
			"user1",
			"Persona "+string(rune(i)),
			"persona_"+string(rune(i)),
			"avatar.jpg",
			"Persona "+string(rune(i)),
			CategoryPersonal,
			"trusted",
		)
		if err != nil {
			t.Fatalf("failed to create persona %d: %v", i+1, err)
		}
	}

	_, err = ps.CreatePersona(
		"user1",
		"Persona 8",
		"persona_8",
		"avatar.jpg",
		"Too many",
		CategoryPersonal,
		"trusted",
	)
	if err != ErrPersonaLimitExceeded {
		t.Errorf("expected ErrPersonaLimitExceeded, got %v", err)
	}

	_, err = ps.CreatePersona(
		"user1",
		"John Professional 2",
		"jprofessional",
		"avatar2.jpg",
		"Another professional",
		CategoryPersonal,
		"trusted",
	)
	if err != ErrDuplicateUsername {
		t.Errorf("expected ErrDuplicateUsername, got %v", err)
	}
}

func TestCreatePersonaWithTrustLimits(t *testing.T) {

	tests := []struct {
		name        string
		trustLevel  string
		maxPersonas int
	}{
		{"unverified", "unverified", 2},
		{"newcomer", "newcomer", 3},
		{"member", "member", 5},
		{"trusted", "trusted", 7},
		{"verified", "verified", 10},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ps := NewPersonaService()
			userID := "user_" + test.name
			ps.CreateMasterIdentity(userID, "did:echo:"+userID, "key", test.trustLevel)

			for i := 0; i < test.maxPersonas; i++ {
				_, err := ps.CreatePersona(
					userID,
					"Persona "+string(rune(i)),
					"persona_"+string(rune(i))+"_"+test.name,
					"avatar.jpg",
					"Test",
					CategoryPersonal,
					test.trustLevel,
				)
				if err != nil {
					t.Fatalf("failed to create persona %d: %v", i, err)
				}
			}

			_, err := ps.CreatePersona(
				userID,
				"Persona Excess",
				"persona_excess_"+test.name,
				"avatar.jpg",
				"Should fail",
				CategoryPersonal,
				test.trustLevel,
			)
			if err != ErrPersonaLimitExceeded {
				t.Errorf("expected ErrPersonaLimitExceeded, got %v", err)
			}
		})
	}
}

func TestGetPersona(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	created, _ := ps.CreatePersona("user1", "Test", "test", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	retrieved, err := ps.GetPersona(created.PersonaID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrieved.PersonaID != created.PersonaID {
		t.Errorf("expected persona ID %s, got %s", created.PersonaID, retrieved.PersonaID)
	}

	_, err = ps.GetPersona("nonexistent")
	if err != ErrPersonaNotFound {
		t.Errorf("expected ErrPersonaNotFound, got %v", err)
	}
}

func TestGetUserPersonas(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	ps.CreateMasterIdentity("user2", "did:echo:user2", "master_key", "trusted")

	for i := 0; i < 3; i++ {
		ps.CreatePersona("user1", "User1 Persona "+string(rune(i)), "user1_persona_"+string(rune(i)), "avatar.jpg", "bio", CategoryPersonal, "trusted")
	}

	for i := 0; i < 2; i++ {
		ps.CreatePersona("user2", "User2 Persona "+string(rune(i)), "user2_persona_"+string(rune(i)), "avatar.jpg", "bio", CategoryPersonal, "trusted")
	}

	user1Personas, err := ps.GetUserPersonas("user1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(user1Personas) != 3 {
		t.Errorf("expected 3 personas for user1, got %d", len(user1Personas))
	}

	user2Personas, err := ps.GetUserPersonas("user2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(user2Personas) != 2 {
		t.Errorf("expected 2 personas for user2, got %d", len(user2Personas))
	}
}

func TestDeletePersona(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	persona, _ := ps.CreatePersona("user1", "Test", "test", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	err := ps.DeletePersona(persona.PersonaID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = ps.GetPersona(persona.PersonaID)
	if err != ErrPersonaNotFound {
		t.Errorf("expected ErrPersonaNotFound after deletion, got %v", err)
	}

	master, _ := ps.GetMasterIdentity("user1")
	for _, pid := range master.Personas {
		if pid == persona.PersonaID {
			t.Error("persona still in master's persona list after deletion")
		}
	}
}

func TestGrantAccess(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	persona, _ := ps.CreatePersona("user1", "Test", "test", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	permissions := AccessPermissions{
		CanView:             true,
		CanMessage:          true,
		CanCall:             false,
		CanSeeOtherPersonas: false,
	}

	grant, err := ps.GrantAccess(persona.PersonaID, "contact1", permissions)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if grant.ContactID != "contact1" {
		t.Errorf("expected contact ID contact1, got %s", grant.ContactID)
	}
	if !grant.Permissions.CanView {
		t.Error("expected CanView permission to be true")
	}

	if len(persona.AccessListIDs) != 1 {
		t.Errorf("expected 1 grant in access list, got %d", len(persona.AccessListIDs))
	}
}

func TestRevokeAccess(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	persona, _ := ps.CreatePersona("user1", "Test", "test", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	grant, _ := ps.GrantAccess(persona.PersonaID, "contact1", AccessPermissions{CanView: true})

	err := ps.RevokeAccess(grant.GrantID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(persona.AccessListIDs) != 0 {
		t.Errorf("expected 0 grants after revocation, got %d", len(persona.AccessListIDs))
	}
}

func TestCheckAccess(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	persona, _ := ps.CreatePersona("user1", "Test", "test", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	allowed, grant, _ := ps.CheckAccess("contact1", persona.PersonaID)
	if allowed {
		t.Error("expected contact1 to not have access")
	}

	permissions := AccessPermissions{CanView: true, CanMessage: true}
	ps.GrantAccess(persona.PersonaID, "contact1", permissions)

	allowed, grant, _ = ps.CheckAccess("contact1", persona.PersonaID)
	if !allowed {
		t.Error("expected contact1 to have access")
	}
	if grant == nil {
		t.Error("expected grant to be returned")
	}
	if !grant.Permissions.CanView {
		t.Error("expected CanView permission in grant")
	}
}

func TestCheckAccessExpiry(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	persona, _ := ps.CreatePersona("user1", "Test", "test", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	permissions := AccessPermissions{CanView: true}
	grant, _ := ps.GrantAccess(persona.PersonaID, "contact1", permissions)

	expiredTime := time.Now().Add(-1 * time.Hour)
	grant.ExpiresAt = &expiredTime

	allowed, _, err := ps.CheckAccess("contact1", persona.PersonaID)
	if allowed {
		t.Error("expected expired grant to deny access")
	}
	if err != ErrGrantExpired {
		t.Errorf("expected ErrGrantExpired, got %v", err)
	}
}

func TestGetContactVisiblePersonas(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")

	public, _ := ps.CreatePersona("user1", "Public", "public", "avatar.jpg", "bio", CategoryPersonal, "trusted")
	public.DefaultVisibility = VisibilityEveryone

	contactOnly, _ := ps.CreatePersona("user1", "Contact Only", "contact_only", "avatar.jpg", "bio", CategoryPersonal, "trusted")
	contactOnly.DefaultVisibility = VisibilityContacts

	private, _ := ps.CreatePersona("user1", "Private", "private", "avatar.jpg", "bio", CategoryPersonal, "trusted")
	private.DefaultVisibility = VisibilityNobody

	ps.GrantAccess(contactOnly.PersonaID, "contact1", AccessPermissions{CanView: true})

	visible, err := ps.GetContactVisiblePersonas("user1", "contact1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(visible) != 2 {
		t.Errorf("expected 2 visible personas, got %d", len(visible))
	}

	visible2, _ := ps.GetContactVisiblePersonas("user1", "contact_other")
	if len(visible2) != 1 {
		t.Errorf("expected 1 visible persona (public only) for contact_other, got %d", len(visible2))
	}
}

func TestUpdatePersonaProfile(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	persona, _ := ps.CreatePersona("user1", "Original", "original", "avatar.jpg", "original bio", CategoryPersonal, "trusted")

	updated, err := ps.UpdatePersonaProfile(persona.PersonaID, "Updated", "updated bio", "away")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.DisplayName != "Updated" {
		t.Errorf("expected display name 'Updated', got %s", updated.DisplayName)
	}
	if updated.Bio != "updated bio" {
		t.Errorf("expected bio 'updated bio', got %s", updated.Bio)
	}
	if updated.Status != "away" {
		t.Errorf("expected status 'away', got %s", updated.Status)
	}
}

func TestPrivacySettings(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	persona, _ := ps.CreatePersona("user1", "Test", "test", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	newSettings := DefaultPrivacySettings()
	newSettings.Searchable = false
	newSettings.WhoCanMessage = VisibilityNobody
	newSettings.SendReadReceipts = false

	updated, err := ps.UpdatePrivacySettings(persona.PersonaID, newSettings)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Privacy.Searchable {
		t.Error("expected Searchable to be false")
	}
	if updated.Privacy.WhoCanMessage != VisibilityNobody {
		t.Errorf("expected WhoCanMessage to be VisibilityNobody, got %s", updated.Privacy.WhoCanMessage)
	}
	if updated.Privacy.SendReadReceipts {
		t.Error("expected SendReadReceipts to be false")
	}
}

func TestNotificationSettings(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	persona, _ := ps.CreatePersona("user1", "Test", "test", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	newSettings := DefaultNotificationSettings()
	newSettings.Enabled = false
	newSettings.Messages = NotificationMentions
	newSettings.SoundEnabled = false

	updated, err := ps.UpdateNotificationSettings(persona.PersonaID, newSettings)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Notifications.Enabled {
		t.Error("expected Enabled to be false")
	}
	if updated.Notifications.Messages != NotificationMentions {
		t.Errorf("expected Messages to be NotificationMentions, got %s", updated.Notifications.Messages)
	}
	if updated.Notifications.SoundEnabled {
		t.Error("expected SoundEnabled to be false")
	}
}

func TestPublicPersonaInfo(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")
	persona, _ := ps.CreatePersona("user1", "Test", "test", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	publicInfo, err := ps.PublicPersonaInfo(persona.PersonaID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if publicInfo.PersonaID != persona.PersonaID {
		t.Errorf("expected persona ID %s, got %s", persona.PersonaID, publicInfo.PersonaID)
	}
	if publicInfo.DisplayName != "Test" {
		t.Error("expected display name in public info")
	}

	if publicInfo.SigningKey != "" {
		t.Error("expected SigningKey to not be in public info")
	}
	if publicInfo.EncryptionKey != "" {
		t.Error("expected EncryptionKey to not be in public info")
	}
}

func TestPersonaKeyDerivation(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "trusted")

	persona1, _ := ps.CreatePersona("user1", "First", "first", "avatar.jpg", "bio", CategoryPersonal, "trusted")
	persona2, _ := ps.CreatePersona("user1", "Second", "second", "avatar.jpg", "bio", CategoryPersonal, "trusted")

	if persona1.SigningKey == persona2.SigningKey {
		t.Error("expected different signing keys for different personas")
	}
	if persona1.EncryptionKey == persona2.EncryptionKey {
		t.Error("expected different encryption keys for different personas")
	}

	if !strings.Contains(persona1.DerivationPath, "m/867530") {
		t.Errorf("expected derivation path to start with m/867530, got %s", persona1.DerivationPath)
	}
}

func TestMultipleCategories(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "verified")

	categories := []PersonaCategory{
		CategoryProfessional,
		CategoryPersonal,
		CategoryFamily,
		CategoryGaming,
		CategoryDating,
		CategoryCreative,
		CategoryAnonymous,
		CategoryCustom,
	}

	for i, cat := range categories {
		persona, err := ps.CreatePersona(
			"user1",
			"Persona "+string(rune(i)),
			"persona_"+string(rune(i)),
			"avatar.jpg",
			"bio",
			cat,
			"verified",
		)
		if err != nil {
			t.Fatalf("failed to create %s persona: %v", cat, err)
		}
		if persona.Category != cat {
			t.Errorf("expected category %s, got %s", cat, persona.Category)
		}
	}
}

func TestCustomCategoryRestriction(t *testing.T) {
	ps := NewPersonaService()
	ps.CreateMasterIdentity("user1", "did:echo:user1", "master_key", "unverified")

	_, err := ps.CreatePersona("user1", "Custom", "custom", "avatar.jpg", "bio", CategoryCustom, "unverified")
	if err != ErrInvalidPersonaCategory {
		t.Errorf("expected ErrInvalidPersonaCategory for unverified user with custom category, got %v", err)
	}

	ps.CreateMasterIdentity("user2", "did:echo:user2", "master_key", "member")
	persona, err := ps.CreatePersona("user2", "Custom", "custom2", "avatar.jpg", "bio", CategoryCustom, "member")
	if err != nil {
		t.Fatalf("expected no error for member with custom category, got %v", err)
	}
	if persona.Category != CategoryCustom {
		t.Error("expected custom category to be created for member")
	}
}
