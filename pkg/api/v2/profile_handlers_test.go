package v2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProfileHandler(t *testing.T) {
	ph := NewProfileHandler()
	if ph == nil {
		t.Fatal("NewProfileHandler() returned nil")
	}
	if ph.personas == nil {
		t.Fatal("personas map not initialized")
	}
	if ph.profiles == nil {
		t.Fatal("profiles map not initialized")
	}
}

func TestGetFullProfile(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/profile/full", nil)
	w := httptest.NewRecorder()

	ph.GetFullProfile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var profile ProfileData
	if err := json.NewDecoder(w.Body).Decode(&profile); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if profile.DisplayName != "Alex Echo" {
		t.Errorf("expected display name 'Alex Echo', got '%s'", profile.DisplayName)
	}
	if profile.TrustScore != 72 {
		t.Errorf("expected trust score 72, got %d", profile.TrustScore)
	}
	if profile.TrustLevel != "Trusted" {
		t.Errorf("expected trust level 'Trusted', got '%s'", profile.TrustLevel)
	}
}

func TestGetFullProfileMethodNotAllowed(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodPost, "/v2/users/profile/full", nil)
	w := httptest.NewRecorder()

	ph.GetFullProfile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestUpdateFullProfile(t *testing.T) {
	ph := NewProfileHandler()

	name := "Updated Name"
	bio := "New bio text"
	body, _ := json.Marshal(UpdateProfileRequest{
		DisplayName: &name,
		Bio:         &bio,
	})

	req := httptest.NewRequest(http.MethodPut, "/v2/users/profile", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	ph.UpdateFullProfile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var profile ProfileData
	json.NewDecoder(w.Body).Decode(&profile)

	if profile.DisplayName != "Updated Name" {
		t.Errorf("expected display name 'Updated Name', got '%s'", profile.DisplayName)
	}
	if profile.Bio != "New bio text" {
		t.Errorf("expected bio 'New bio text', got '%s'", profile.Bio)
	}
}

func TestUpdateProfileBioTooLong(t *testing.T) {
	ph := NewProfileHandler()

	// Bio limit is 500, so 600 bytes should fail
	longBio := string(make([]byte, 600))
	body, _ := json.Marshal(UpdateProfileRequest{
		Bio: &longBio,
	})

	req := httptest.NewRequest(http.MethodPut, "/v2/users/profile", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	ph.UpdateFullProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for long bio, got %d", w.Code)
	}
}

func TestUpdateProfileBioAtLimit(t *testing.T) {
	ph := NewProfileHandler()

	// 500 chars should be accepted
	bio500 := string(make([]byte, 500))
	for i := range bio500 {
		bio500 = bio500[:i] + "a" + bio500[i+1:]
	}
	body, _ := json.Marshal(UpdateProfileRequest{
		Bio: &bio500,
	})

	req := httptest.NewRequest(http.MethodPut, "/v2/users/profile", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	ph.UpdateFullProfile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for 500-char bio, got %d", w.Code)
	}
}

func TestUpdateProfileInvalidUsername(t *testing.T) {
	ph := NewProfileHandler()

	shortUsername := "ab"
	body, _ := json.Marshal(UpdateProfileRequest{
		Username: &shortUsername,
	})

	req := httptest.NewRequest(http.MethodPut, "/v2/users/profile", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	ph.UpdateFullProfile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for short username, got %d", w.Code)
	}
}

func TestCheckUsernameAvailability(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/check-username?username=newuser", nil)
	w := httptest.NewRecorder()

	ph.CheckUsernameAvailability(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)

	if result["available"] != true {
		t.Error("expected username to be available")
	}
}

func TestCheckUsernameMissing(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/check-username", nil)
	w := httptest.NewRecorder()

	ph.CheckUsernameAvailability(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestListPersonasEmpty(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	w := httptest.NewRecorder()

	ph.ListPersonas(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)

	count := int(result["count"].(float64))
	if count != 0 {
		t.Errorf("expected 0 personas, got %d", count)
	}
}

func TestCreatePersona(t *testing.T) {
	ph := NewProfileHandler()

	body, _ := json.Marshal(CreatePersonaRequest{
		Type:               "professional",
		Name:               "Professional",
		DisplayName:        "Alex Echo",
		Bio:                "Product Lead @ Echo",
		UseMainAvatar:      true,
		Visibility:         "all",
		SelectedContactIDs: []string{},
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	ph.CreatePersona(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var persona Persona
	json.NewDecoder(w.Body).Decode(&persona)

	if persona.Name != "Professional" {
		t.Errorf("expected name 'Professional', got '%s'", persona.Name)
	}
	if persona.Type != "professional" {
		t.Errorf("expected type 'professional', got '%s'", persona.Type)
	}
	if !persona.IsDefault {
		t.Error("first persona should be default")
	}
}

func TestCreatePersonaWithUsername(t *testing.T) {
	ph := NewProfileHandler()

	body, _ := json.Marshal(CreatePersonaRequest{
		Type:        "professional",
		Name:        "Pro",
		DisplayName: "Alex Echo",
		Username:    "alex_pro",
		Visibility:  "all",
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	ph.CreatePersona(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var persona Persona
	json.NewDecoder(w.Body).Decode(&persona)

	if persona.Username != "alex_pro" {
		t.Errorf("expected username 'alex_pro', got '%s'", persona.Username)
	}
}

func TestCreatePersonaDuplicateUsername(t *testing.T) {
	ph := NewProfileHandler()

	// Create first persona with username
	body1, _ := json.Marshal(CreatePersonaRequest{
		Type:        "professional",
		Name:        "Pro",
		DisplayName: "Alex",
		Username:    "unique_name",
		Visibility:  "all",
	})
	req1 := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body1))
	req1.Header.Set("X-User-ID", "test-user")
	w1 := httptest.NewRecorder()
	ph.CreatePersona(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first create failed: %d", w1.Code)
	}

	// Create second with same username
	body2, _ := json.Marshal(CreatePersonaRequest{
		Type:        "personal",
		Name:        "Personal",
		DisplayName: "Alex P",
		Username:    "unique_name",
		Visibility:  "all",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body2))
	req2.Header.Set("X-User-ID", "test-user")
	w2 := httptest.NewRecorder()
	ph.CreatePersona(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected status 409 for duplicate username, got %d", w2.Code)
	}
}

func TestCreatePersonaMissingFields(t *testing.T) {
	ph := NewProfileHandler()

	body, _ := json.Marshal(CreatePersonaRequest{
		Type: "professional",
		Name: "",
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ph.CreatePersona(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreatePersonaInvalidType(t *testing.T) {
	ph := NewProfileHandler()

	body, _ := json.Marshal(CreatePersonaRequest{
		Type:        "invalid",
		Name:        "Test",
		DisplayName: "Test",
		Visibility:  "all",
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ph.CreatePersona(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreatePersonaNewTypes(t *testing.T) {
	newTypes := []string{"dating", "creative", "anonymous"}

	for _, pt := range newTypes {
		ph := NewProfileHandler()

		body, _ := json.Marshal(CreatePersonaRequest{
			Type:        pt,
			Name:        "Test " + pt,
			DisplayName: "Test",
			Visibility:  "all",
		})

		req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
		req.Header.Set("X-User-ID", "test-user")
		w := httptest.NewRecorder()

		ph.CreatePersona(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201 for type '%s', got %d", pt, w.Code)
		}

		var persona Persona
		json.NewDecoder(w.Body).Decode(&persona)

		if persona.Type != pt {
			t.Errorf("expected type '%s', got '%s'", pt, persona.Type)
		}
	}
}

func TestCreatePersonaMaxLimitUnverified(t *testing.T) {
	ph := NewProfileHandler()
	// No profile set → default "unverified" → limit 2

	for i := 0; i < 2; i++ {
		body, _ := json.Marshal(CreatePersonaRequest{
			Type:        "personal",
			Name:        "Test",
			DisplayName: "Test",
			Visibility:  "all",
		})
		req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
		req.Header.Set("X-User-ID", "test-user")
		w := httptest.NewRecorder()
		ph.CreatePersona(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("creating persona %d failed with status %d", i+1, w.Code)
		}
	}

	// Third should fail (unverified limit is 2)
	body, _ := json.Marshal(CreatePersonaRequest{
		Type:        "gaming",
		Name:        "Extra",
		DisplayName: "Extra",
		Visibility:  "all",
	})
	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.CreatePersona(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409 for unverified limit, got %d", w.Code)
	}
}

func TestCreatePersonaMaxLimitTrusted(t *testing.T) {
	ph := NewProfileHandler()

	// Set up profile with "Trusted" trust level (limit 7)
	ph.mu.Lock()
	ph.profiles["trust-user"] = &ProfileData{
		DisplayName: "Trust User",
		TrustLevel:  "Trusted",
		TrustScore:  72,
	}
	ph.mu.Unlock()

	for i := 0; i < 7; i++ {
		body, _ := json.Marshal(CreatePersonaRequest{
			Type:        "personal",
			Name:        "Test",
			DisplayName: "Test",
			Visibility:  "all",
		})
		req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
		req.Header.Set("X-User-ID", "trust-user")
		w := httptest.NewRecorder()
		ph.CreatePersona(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("creating persona %d failed with status %d", i+1, w.Code)
		}
	}

	// Eighth should fail
	body, _ := json.Marshal(CreatePersonaRequest{
		Type:        "gaming",
		Name:        "Extra",
		DisplayName: "Extra",
		Visibility:  "all",
	})
	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "trust-user")
	w := httptest.NewRecorder()
	ph.CreatePersona(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409 for trusted limit (7), got %d", w.Code)
	}
}

func TestTrustLevelLimits(t *testing.T) {
	tests := []struct {
		level string
		limit int
	}{
		{"unverified", 2},
		{"newcomer", 3},
		{"member", 5},
		{"basic", 5},
		{"trusted", 7},
		{"verified", 10},
		{"elite", 10},
	}

	for _, tt := range tests {
		ph := NewProfileHandler()
		ph.mu.Lock()
		ph.profiles["user"] = &ProfileData{TrustLevel: tt.level}
		ph.mu.Unlock()

		max := ph.getMaxPersonas("user")
		if max != tt.limit {
			t.Errorf("trust level '%s': expected limit %d, got %d", tt.level, tt.limit, max)
		}
	}
}

func TestDeletePersona(t *testing.T) {
	ph := NewProfileHandler()

	// Create a persona first
	body, _ := json.Marshal(CreatePersonaRequest{
		Type:        "professional",
		Name:        "Pro",
		DisplayName: "Pro Me",
		Visibility:  "all",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	createReq.Header.Set("X-User-ID", "test-user")
	cw := httptest.NewRecorder()
	ph.CreatePersona(cw, createReq)

	var created Persona
	json.NewDecoder(cw.Body).Decode(&created)

	// Delete it
	deleteReq := httptest.NewRequest(http.MethodDelete, "/v2/users/personas?id="+created.ID, nil)
	deleteReq.Header.Set("X-User-ID", "test-user")
	dw := httptest.NewRecorder()
	ph.DeletePersona(dw, deleteReq)

	if dw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", dw.Code)
	}

	// Verify it's gone
	listReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	listReq.Header.Set("X-User-ID", "test-user")
	lw := httptest.NewRecorder()
	ph.ListPersonas(lw, listReq)

	var result map[string]interface{}
	json.NewDecoder(lw.Body).Decode(&result)

	count := int(result["count"].(float64))
	if count != 0 {
		t.Errorf("expected 0 personas after delete, got %d", count)
	}
}

func TestDeletePersonaNotFound(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodDelete, "/v2/users/personas?id=nonexistent", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	ph.DeletePersona(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestSetDefaultPersona(t *testing.T) {
	ph := NewProfileHandler()

	// Create two personas
	for _, name := range []string{"Pro", "Personal"} {
		body, _ := json.Marshal(CreatePersonaRequest{
			Type:        "professional",
			Name:        name,
			DisplayName: name,
			Visibility:  "all",
		})
		req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
		req.Header.Set("X-User-ID", "test-user")
		w := httptest.NewRecorder()
		ph.CreatePersona(w, req)
	}

	// Get second persona ID
	listReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	listReq.Header.Set("X-User-ID", "test-user")
	lw := httptest.NewRecorder()
	ph.ListPersonas(lw, listReq)

	var result map[string]interface{}
	json.NewDecoder(lw.Body).Decode(&result)
	data := result["data"].([]interface{})
	secondPersona := data[1].(map[string]interface{})
	secondID := secondPersona["id"].(string)

	// Set second as default
	setReq := httptest.NewRequest(http.MethodPost, "/v2/users/personas/default?id="+secondID, nil)
	setReq.Header.Set("X-User-ID", "test-user")
	sw := httptest.NewRecorder()
	ph.SetDefaultPersona(sw, setReq)

	if sw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", sw.Code)
	}

	// Verify
	listReq2 := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	listReq2.Header.Set("X-User-ID", "test-user")
	lw2 := httptest.NewRecorder()
	ph.ListPersonas(lw2, listReq2)

	var result2 map[string]interface{}
	json.NewDecoder(lw2.Body).Decode(&result2)
	data2 := result2["data"].([]interface{})

	first := data2[0].(map[string]interface{})
	second := data2[1].(map[string]interface{})

	if first["is_default"].(bool) {
		t.Error("first persona should no longer be default")
	}
	if !second["is_default"].(bool) {
		t.Error("second persona should be default")
	}
}

func TestSetDefaultPersonaNotFound(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/default?id=nonexistent", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	ph.SetDefaultPersona(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestUpdatePersona(t *testing.T) {
	ph := NewProfileHandler()

	// Create persona
	body, _ := json.Marshal(CreatePersonaRequest{
		Type:        "professional",
		Name:        "Pro",
		DisplayName: "Alex",
		Visibility:  "all",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	createReq.Header.Set("X-User-ID", "test-user")
	cw := httptest.NewRecorder()
	ph.CreatePersona(cw, createReq)

	var created Persona
	json.NewDecoder(cw.Body).Decode(&created)

	// Update it
	newName := "Updated Pro"
	newBio := "New bio"
	updateBody, _ := json.Marshal(UpdatePersonaRequest{
		Name: &newName,
		Bio:  &newBio,
	})

	updateReq := httptest.NewRequest(http.MethodPut, "/v2/users/personas?id="+created.ID, bytes.NewReader(updateBody))
	updateReq.Header.Set("X-User-ID", "test-user")
	uw := httptest.NewRecorder()
	ph.UpdatePersona(uw, updateReq)

	if uw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", uw.Code)
	}

	var updated Persona
	json.NewDecoder(uw.Body).Decode(&updated)

	if updated.Name != "Updated Pro" {
		t.Errorf("expected name 'Updated Pro', got '%s'", updated.Name)
	}
	if updated.Bio != "New bio" {
		t.Errorf("expected bio 'New bio', got '%s'", updated.Bio)
	}
}

func TestGetNotificationSettings(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/settings/notifications", nil)
	w := httptest.NewRecorder()

	ph.GetNotificationSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var settings NotificationSettingsData
	json.NewDecoder(w.Body).Decode(&settings)

	if !settings.MessageNotifications {
		t.Error("expected message notifications to be enabled")
	}
	if settings.RingSound != "Reflection" {
		t.Errorf("expected ring sound 'Reflection', got '%s'", settings.RingSound)
	}
}

func TestGetPrivacySettings(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/settings/privacy", nil)
	w := httptest.NewRecorder()

	ph.GetPrivacySettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var settings PrivacySettingsData
	json.NewDecoder(w.Body).Decode(&settings)

	if settings.FindByUsername != "everyone" {
		t.Errorf("expected find by username 'everyone', got '%s'", settings.FindByUsername)
	}
	if !settings.ReadReceipts {
		t.Error("expected read receipts to be enabled")
	}
}

func TestGetAppearanceSettings(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/settings/appearance", nil)
	w := httptest.NewRecorder()

	ph.GetAppearanceSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var settings AppearanceSettingsData
	json.NewDecoder(w.Body).Decode(&settings)

	if settings.Theme != "light" {
		t.Errorf("expected theme 'light', got '%s'", settings.Theme)
	}
	if settings.AccentColor != "indigo" {
		t.Errorf("expected accent color 'indigo', got '%s'", settings.AccentColor)
	}
}

func TestGetStorageInfo(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/storage", nil)
	w := httptest.NewRecorder()

	ph.GetStorageInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var info StorageInfoData
	json.NewDecoder(w.Body).Decode(&info)

	if info.TotalCapacityBytes != 5368709120 {
		t.Errorf("expected 5GB capacity, got %d", info.TotalCapacityBytes)
	}
	if info.AutoBackup != "daily" {
		t.Errorf("expected auto backup 'daily', got '%s'", info.AutoBackup)
	}
}

func TestGetAccountInfo(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/account", nil)
	w := httptest.NewRecorder()

	ph.GetAccountInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var info AccountInfoData
	json.NewDecoder(w.Body).Decode(&info)

	if info.PasskeyCount != 2 {
		t.Errorf("expected 2 passkeys, got %d", info.PasskeyCount)
	}
	if !info.TwoFactorEnabled {
		t.Error("expected 2FA to be enabled")
	}
	if info.ActiveSessionCount != 3 {
		t.Errorf("expected 3 active sessions, got %d", info.ActiveSessionCount)
	}
}

func TestDeleteDefaultPersonaPromotesNext(t *testing.T) {
	ph := NewProfileHandler()

	// Create two personas
	for _, name := range []string{"First", "Second"} {
		body, _ := json.Marshal(CreatePersonaRequest{
			Type:        "personal",
			Name:        name,
			DisplayName: name,
			Visibility:  "all",
		})
		req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
		req.Header.Set("X-User-ID", "test-user")
		w := httptest.NewRecorder()
		ph.CreatePersona(w, req)
	}

	// Get first persona ID (default)
	listReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	listReq.Header.Set("X-User-ID", "test-user")
	lw := httptest.NewRecorder()
	ph.ListPersonas(lw, listReq)

	var result map[string]interface{}
	json.NewDecoder(lw.Body).Decode(&result)
	data := result["data"].([]interface{})
	firstID := data[0].(map[string]interface{})["id"].(string)

	// Delete the default
	deleteReq := httptest.NewRequest(http.MethodDelete, "/v2/users/personas?id="+firstID, nil)
	deleteReq.Header.Set("X-User-ID", "test-user")
	dw := httptest.NewRecorder()
	ph.DeletePersona(dw, deleteReq)

	// Verify second is now default
	listReq2 := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	listReq2.Header.Set("X-User-ID", "test-user")
	lw2 := httptest.NewRecorder()
	ph.ListPersonas(lw2, listReq2)

	var result2 map[string]interface{}
	json.NewDecoder(lw2.Body).Decode(&result2)
	data2 := result2["data"].([]interface{})

	if len(data2) != 1 {
		t.Fatalf("expected 1 persona, got %d", len(data2))
	}

	remaining := data2[0].(map[string]interface{})
	if !remaining["is_default"].(bool) {
		t.Error("remaining persona should become default after deleting the default")
	}
}

// ===== Enhanced Deletion with Recovery =====

func TestDeletePersonaEnhancedSoftDelete(t *testing.T) {
	ph := NewProfileHandler()

	// Create two personas (need at least 2 for deletion)
	for _, name := range []string{"First", "Second"} {
		body, _ := json.Marshal(CreatePersonaRequest{
			Type:        "personal",
			Name:        name,
			DisplayName: name,
			Visibility:  "all",
		})
		req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
		req.Header.Set("X-User-ID", "test-user")
		w := httptest.NewRecorder()
		ph.CreatePersona(w, req)
	}

	// Get second persona ID
	listReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	listReq.Header.Set("X-User-ID", "test-user")
	lw := httptest.NewRecorder()
	ph.ListPersonas(lw, listReq)

	var result map[string]interface{}
	json.NewDecoder(lw.Body).Decode(&result)
	data := result["data"].([]interface{})
	secondID := data[1].(map[string]interface{})["id"].(string)

	// Soft delete with 30-day recovery
	deleteBody, _ := json.Marshal(DeletePersonaRequest{
		ArchiveConversations: true,
		NotifyContacts:       false,
		RecoveryPeriodDays:   30,
	})
	deleteReq := httptest.NewRequest(http.MethodDelete, "/v2/users/personas/enhanced?id="+secondID, bytes.NewReader(deleteBody))
	deleteReq.Header.Set("X-User-ID", "test-user")
	dw := httptest.NewRecorder()
	ph.DeletePersonaEnhanced(dw, deleteReq)

	if dw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", dw.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(dw.Body).Decode(&resp)

	if resp["recoverable"] != true {
		t.Error("expected recoverable to be true")
	}
	if resp["recovery_expires_at"] == nil {
		t.Error("expected recovery_expires_at to be set")
	}
}

func TestDeletePersonaEnhancedHardDelete(t *testing.T) {
	ph := NewProfileHandler()

	// Create two personas
	for _, name := range []string{"First", "Second"} {
		body, _ := json.Marshal(CreatePersonaRequest{
			Type:        "personal",
			Name:        name,
			DisplayName: name,
			Visibility:  "all",
		})
		req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
		req.Header.Set("X-User-ID", "test-user")
		w := httptest.NewRecorder()
		ph.CreatePersona(w, req)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	listReq.Header.Set("X-User-ID", "test-user")
	lw := httptest.NewRecorder()
	ph.ListPersonas(lw, listReq)

	var result map[string]interface{}
	json.NewDecoder(lw.Body).Decode(&result)
	data := result["data"].([]interface{})
	secondID := data[1].(map[string]interface{})["id"].(string)

	// Hard delete (recovery_period_days = 0)
	deleteBody, _ := json.Marshal(DeletePersonaRequest{
		RecoveryPeriodDays: 0,
	})
	deleteReq := httptest.NewRequest(http.MethodDelete, "/v2/users/personas/enhanced?id="+secondID, bytes.NewReader(deleteBody))
	deleteReq.Header.Set("X-User-ID", "test-user")
	dw := httptest.NewRecorder()
	ph.DeletePersonaEnhanced(dw, deleteReq)

	if dw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", dw.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(dw.Body).Decode(&resp)

	if resp["recoverable"] != false {
		t.Error("expected recoverable to be false for hard delete")
	}

	// Verify persona count
	listReq2 := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	listReq2.Header.Set("X-User-ID", "test-user")
	lw2 := httptest.NewRecorder()
	ph.ListPersonas(lw2, listReq2)

	var result2 map[string]interface{}
	json.NewDecoder(lw2.Body).Decode(&result2)
	count := int(result2["count"].(float64))
	if count != 1 {
		t.Errorf("expected 1 persona after hard delete, got %d", count)
	}
}

func TestDeletePersonaEnhancedCannotDeleteLast(t *testing.T) {
	ph := NewProfileHandler()

	// Create only one persona
	body, _ := json.Marshal(CreatePersonaRequest{
		Type:        "personal",
		Name:        "Only",
		DisplayName: "Only One",
		Visibility:  "all",
	})
	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.CreatePersona(w, req)

	var created Persona
	json.NewDecoder(w.Body).Decode(&created)

	// Try to delete the last persona
	deleteReq := httptest.NewRequest(http.MethodDelete, "/v2/users/personas/enhanced?id="+created.ID, nil)
	deleteReq.Header.Set("X-User-ID", "test-user")
	dw := httptest.NewRecorder()
	ph.DeletePersonaEnhanced(dw, deleteReq)

	if dw.Code != http.StatusConflict {
		t.Errorf("expected status 409 for deleting last persona, got %d", dw.Code)
	}
}

// ===== Persona Recovery =====

func TestRecoverPersona(t *testing.T) {
	ph := NewProfileHandler()

	// Create two personas
	for _, name := range []string{"First", "Second"} {
		body, _ := json.Marshal(CreatePersonaRequest{
			Type:        "personal",
			Name:        name,
			DisplayName: name,
			Visibility:  "all",
		})
		req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
		req.Header.Set("X-User-ID", "test-user")
		w := httptest.NewRecorder()
		ph.CreatePersona(w, req)
	}

	// Get second persona ID
	listReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas", nil)
	listReq.Header.Set("X-User-ID", "test-user")
	lw := httptest.NewRecorder()
	ph.ListPersonas(lw, listReq)

	var result map[string]interface{}
	json.NewDecoder(lw.Body).Decode(&result)
	data := result["data"].([]interface{})
	secondID := data[1].(map[string]interface{})["id"].(string)

	// Soft delete
	deleteBody, _ := json.Marshal(DeletePersonaRequest{
		RecoveryPeriodDays: 30,
	})
	deleteReq := httptest.NewRequest(http.MethodDelete, "/v2/users/personas/enhanced?id="+secondID, bytes.NewReader(deleteBody))
	deleteReq.Header.Set("X-User-ID", "test-user")
	dw := httptest.NewRecorder()
	ph.DeletePersonaEnhanced(dw, deleteReq)

	// Recover it
	recoverReq := httptest.NewRequest(http.MethodPost, "/v2/users/personas/recover?id="+secondID, nil)
	recoverReq.Header.Set("X-User-ID", "test-user")
	rw := httptest.NewRecorder()
	ph.RecoverPersona(rw, recoverReq)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rw.Body).Decode(&resp)

	if resp["recovered"] != true {
		t.Error("expected recovered to be true")
	}
}

func TestRecoverPersonaNotRecoverable(t *testing.T) {
	ph := NewProfileHandler()

	// Create persona (no soft-delete, so not recoverable)
	body, _ := json.Marshal(CreatePersonaRequest{
		Type:        "personal",
		Name:        "Active",
		DisplayName: "Active",
		Visibility:  "all",
	})
	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.CreatePersona(w, req)

	var created Persona
	json.NewDecoder(w.Body).Decode(&created)

	// Try to recover a non-deleted persona
	recoverReq := httptest.NewRequest(http.MethodPost, "/v2/users/personas/recover?id="+created.ID, nil)
	recoverReq.Header.Set("X-User-ID", "test-user")
	rw := httptest.NewRecorder()
	ph.RecoverPersona(rw, recoverReq)

	if rw.Code != http.StatusConflict {
		t.Errorf("expected status 409 for non-recoverable persona, got %d", rw.Code)
	}
}

func TestRecoverPersonaNotFound(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/recover?id=nonexistent", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.RecoverPersona(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// ===== Access Grants =====

// helper to create a persona and return its ID
func createTestPersona(t *testing.T, ph *ProfileHandler, userID, name string) string {
	t.Helper()
	body, _ := json.Marshal(CreatePersonaRequest{
		Type:        "professional",
		Name:        name,
		DisplayName: name,
		Visibility:  "selected",
	})
	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas", bytes.NewReader(body))
	req.Header.Set("X-User-ID", userID)
	w := httptest.NewRecorder()
	ph.CreatePersona(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create persona '%s': status %d", name, w.Code)
	}

	var persona Persona
	json.NewDecoder(w.Body).Decode(&persona)
	return persona.ID
}

func TestGrantAccess(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	grantBody, _ := json.Marshal(GrantAccessRequest{
		ContactID:           "contact-1",
		CanView:             true,
		CanMessage:          true,
		CanCall:             false,
		CanSeeOtherPersonas: false,
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/access?id="+personaID, bytes.NewReader(grantBody))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GrantAccess(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var grant AccessGrant
	json.NewDecoder(w.Body).Decode(&grant)

	if grant.ContactID != "contact-1" {
		t.Errorf("expected contact 'contact-1', got '%s'", grant.ContactID)
	}
	if !grant.CanView {
		t.Error("expected can_view to be true")
	}
	if !grant.CanMessage {
		t.Error("expected can_message to be true")
	}
	if grant.CanCall {
		t.Error("expected can_call to be false")
	}
}

func TestGrantAccessMissingContact(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	grantBody, _ := json.Marshal(GrantAccessRequest{
		ContactID: "",
		CanView:   true,
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/access?id="+personaID, bytes.NewReader(grantBody))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GrantAccess(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing contact, got %d", w.Code)
	}
}

func TestGrantAccessPersonaNotFound(t *testing.T) {
	ph := NewProfileHandler()

	grantBody, _ := json.Marshal(GrantAccessRequest{
		ContactID: "contact-1",
		CanView:   true,
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/access?id=nonexistent", bytes.NewReader(grantBody))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GrantAccess(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRevokeAccess(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	// Grant access first
	grantBody, _ := json.Marshal(GrantAccessRequest{
		ContactID:  "contact-1",
		CanView:    true,
		CanMessage: true,
	})
	grantReq := httptest.NewRequest(http.MethodPost, "/v2/users/personas/access?id="+personaID, bytes.NewReader(grantBody))
	grantReq.Header.Set("X-User-ID", "test-user")
	gw := httptest.NewRecorder()
	ph.GrantAccess(gw, grantReq)

	var grant AccessGrant
	json.NewDecoder(gw.Body).Decode(&grant)

	// Revoke
	revokeReq := httptest.NewRequest(http.MethodDelete, "/v2/users/personas/access?grant_id="+grant.ID, nil)
	revokeReq.Header.Set("X-User-ID", "test-user")
	rw := httptest.NewRecorder()
	ph.RevokeAccess(rw, revokeReq)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rw.Body).Decode(&resp)

	if resp["revoked"] != true {
		t.Error("expected revoked to be true")
	}
}

func TestRevokeAccessNotFound(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodDelete, "/v2/users/personas/access?grant_id=nonexistent", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.RevokeAccess(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// ===== Visibility Matrix =====

func TestGetVisibilityMatrixEmpty(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/personas/visibility-matrix", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GetVisibilityMatrix(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	entries := resp["entries"].([]interface{})
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty user, got %d", len(entries))
	}
}

func TestGetVisibilityMatrixWithGrants(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	// Grant access to a contact
	grantBody, _ := json.Marshal(GrantAccessRequest{
		ContactID:  "contact-1",
		CanView:    true,
		CanMessage: true,
	})
	grantReq := httptest.NewRequest(http.MethodPost, "/v2/users/personas/access?id="+personaID, bytes.NewReader(grantBody))
	grantReq.Header.Set("X-User-ID", "test-user")
	gw := httptest.NewRecorder()
	ph.GrantAccess(gw, grantReq)

	// Get matrix
	req := httptest.NewRequest(http.MethodGet, "/v2/users/personas/visibility-matrix", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GetVisibilityMatrix(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	entries := resp["entries"].([]interface{})
	if len(entries) == 0 {
		t.Error("expected at least 1 visibility entry")
	}
}

// ===== Persona Switching =====

func TestValidatePersonaSwitch(t *testing.T) {
	ph := NewProfileHandler()

	// Create two personas with selected visibility
	personaID1 := createTestPersona(t, ph, "test-user", "Pro")
	personaID2 := createTestPersona(t, ph, "test-user", "Personal")

	// Grant contact-1 access to both personas
	for _, pid := range []string{personaID1, personaID2} {
		grantBody, _ := json.Marshal(GrantAccessRequest{
			ContactID: "contact-1",
			CanView:   true,
		})
		req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/access?id="+pid, bytes.NewReader(grantBody))
		req.Header.Set("X-User-ID", "test-user")
		w := httptest.NewRecorder()
		ph.GrantAccess(w, req)
	}

	// Validate switch
	switchBody, _ := json.Marshal(ValidateSwitchRequest{
		FromPersonaID: &personaID1,
		ToPersonaID:   personaID2,
		ContactID:     "contact-1",
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/validate-switch", bytes.NewReader(switchBody))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.ValidatePersonaSwitch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["to_persona_id"] != personaID2 {
		t.Errorf("expected to_persona_id '%s', got '%v'", personaID2, resp["to_persona_id"])
	}
}

func TestValidatePersonaSwitchMissingFields(t *testing.T) {
	ph := NewProfileHandler()

	switchBody, _ := json.Marshal(ValidateSwitchRequest{
		ToPersonaID: "",
		ContactID:   "",
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/validate-switch", bytes.NewReader(switchBody))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.ValidatePersonaSwitch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// ===== Per-Persona Settings =====

func TestGetPersonaPrivacySettings(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	req := httptest.NewRequest(http.MethodGet, "/v2/users/personas/settings/privacy?id="+personaID, nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GetPersonaPrivacySettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var settings PersonaPrivacySettingsData
	json.NewDecoder(w.Body).Decode(&settings)

	// Check defaults
	if !settings.SendReadReceipts {
		t.Error("expected default send_read_receipts to be true")
	}
}

func TestGetPersonaPrivacySettingsNotFound(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/personas/settings/privacy?id=nonexistent", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GetPersonaPrivacySettings(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestUpdatePersonaPrivacySettings(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	settings := PersonaPrivacySettingsData{
		LastSeenVisibility: "nobody",
		SendReadReceipts:   false,
		Searchable:         false,
	}
	body, _ := json.Marshal(settings)

	req := httptest.NewRequest(http.MethodPut, "/v2/users/personas/settings/privacy?id="+personaID, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.UpdatePersonaPrivacySettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify the update persisted
	getReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas/settings/privacy?id="+personaID, nil)
	getReq.Header.Set("X-User-ID", "test-user")
	gw := httptest.NewRecorder()
	ph.GetPersonaPrivacySettings(gw, getReq)

	var updated PersonaPrivacySettingsData
	json.NewDecoder(gw.Body).Decode(&updated)

	if updated.LastSeenVisibility != "nobody" {
		t.Errorf("expected last_seen 'nobody', got '%s'", updated.LastSeenVisibility)
	}
	if updated.SendReadReceipts {
		t.Error("expected send_read_receipts to be false after update")
	}
}

func TestGetPersonaNotificationSettings(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	req := httptest.NewRequest(http.MethodGet, "/v2/users/personas/settings/notifications?id="+personaID, nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GetPersonaNotificationSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestUpdatePersonaNotificationSettings(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	settings := PersonaNotifSettingsData{
		Enabled:           false,
		QuietHoursEnabled: true,
		QuietHoursStart:   "23:00",
		QuietHoursEnd:     "08:00",
	}
	body, _ := json.Marshal(settings)

	req := httptest.NewRequest(http.MethodPut, "/v2/users/personas/settings/notifications?id="+personaID, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.UpdatePersonaNotificationSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetPersonaFeatureSettings(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	req := httptest.NewRequest(http.MethodGet, "/v2/users/personas/settings/features?id="+personaID, nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GetPersonaFeatureSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestUpdatePersonaFeatureSettings(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	settings := PersonaFeatureSettingsData{
		VoiceCalls:    false,
		VideoCalls:    false,
		FileSharing:   true,
		MaxGroupSize:  50,
		MaxFileSizeMB: 100,
	}
	body, _ := json.Marshal(settings)

	req := httptest.NewRequest(http.MethodPut, "/v2/users/personas/settings/features?id="+personaID, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.UpdatePersonaFeatureSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify
	getReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas/settings/features?id="+personaID, nil)
	getReq.Header.Set("X-User-ID", "test-user")
	gw := httptest.NewRecorder()
	ph.GetPersonaFeatureSettings(gw, getReq)

	var updated PersonaFeatureSettingsData
	json.NewDecoder(gw.Body).Decode(&updated)

	if updated.VoiceCalls {
		t.Error("expected voice_calls to be false")
	}
	if updated.MaxGroupSize != 50 {
		t.Errorf("expected max_group_size 50, got %d", updated.MaxGroupSize)
	}
}

// ===== Persona Badges =====

func TestAddPersonaBadge(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	badgeBody, _ := json.Marshal(AddBadgeRequest{
		Type:       "professional_certified",
		Issuer:     "EchoVerify",
		Verifiable: true,
		Proof:      "proof-hash-123",
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/badges?id="+personaID, bytes.NewReader(badgeBody))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.AddPersonaBadge(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var badge PersonaBadgeData
	json.NewDecoder(w.Body).Decode(&badge)

	if badge.Type != "professional_certified" {
		t.Errorf("expected badge type 'professional_certified', got '%s'", badge.Type)
	}
	if badge.Issuer != "EchoVerify" {
		t.Errorf("expected issuer 'EchoVerify', got '%s'", badge.Issuer)
	}
	if !badge.Verifiable {
		t.Error("expected badge to be verifiable")
	}
	if badge.ID == "" {
		t.Error("expected badge ID to be set")
	}
}

func TestAddPersonaBadgeNotFound(t *testing.T) {
	ph := NewProfileHandler()

	badgeBody, _ := json.Marshal(AddBadgeRequest{
		Type:   "professional_certified",
		Issuer: "EchoVerify",
	})

	req := httptest.NewRequest(http.MethodPost, "/v2/users/personas/badges?id=nonexistent", bytes.NewReader(badgeBody))
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.AddPersonaBadge(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRemovePersonaBadge(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	// Add a badge first
	badgeBody, _ := json.Marshal(AddBadgeRequest{
		Type:   "gamer_rank",
		Issuer: "GameVerify",
	})
	addReq := httptest.NewRequest(http.MethodPost, "/v2/users/personas/badges?id="+personaID, bytes.NewReader(badgeBody))
	addReq.Header.Set("X-User-ID", "test-user")
	aw := httptest.NewRecorder()
	ph.AddPersonaBadge(aw, addReq)

	var badge PersonaBadgeData
	json.NewDecoder(aw.Body).Decode(&badge)

	// Remove it
	removeReq := httptest.NewRequest(http.MethodDelete, "/v2/users/personas/badges?id="+personaID+"&badge_id="+badge.ID, nil)
	removeReq.Header.Set("X-User-ID", "test-user")
	rw := httptest.NewRecorder()
	ph.RemovePersonaBadge(rw, removeReq)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rw.Body).Decode(&resp)

	if resp["removed"] != true {
		t.Error("expected removed to be true")
	}
}

func TestRemovePersonaBadgeNotFound(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	req := httptest.NewRequest(http.MethodDelete, "/v2/users/personas/badges?id="+personaID+"&badge_id=nonexistent", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.RemovePersonaBadge(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// ===== Persona Conversations =====

func TestGetPersonaConversationsEmpty(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	req := httptest.NewRequest(http.MethodGet, "/v2/users/personas/conversations?id="+personaID, nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GetPersonaConversations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	count := int(resp["count"].(float64))
	if count != 0 {
		t.Errorf("expected 0 conversations, got %d", count)
	}
}

// ===== Persona Limits =====

func TestGetPersonaLimitsUnverified(t *testing.T) {
	ph := NewProfileHandler()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/persona-limits", nil)
	req.Header.Set("X-User-ID", "new-user")
	w := httptest.NewRecorder()
	ph.GetPersonaLimits(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	maxPersonas := int(resp["max_personas"].(float64))
	if maxPersonas != 2 {
		t.Errorf("expected max_personas 2 for unverified, got %d", maxPersonas)
	}

	trustLevel := resp["trust_level"].(string)
	if trustLevel != "unverified" {
		t.Errorf("expected trust_level 'unverified', got '%s'", trustLevel)
	}

	allowCustom := resp["allow_custom_categories"].(bool)
	if allowCustom {
		t.Error("expected allow_custom_categories to be false for unverified")
	}
}

func TestGetPersonaLimitsTrusted(t *testing.T) {
	ph := NewProfileHandler()

	ph.mu.Lock()
	ph.profiles["trusted-user"] = &ProfileData{
		TrustLevel: "Trusted",
		TrustScore: 72,
	}
	ph.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/v2/users/persona-limits", nil)
	req.Header.Set("X-User-ID", "trusted-user")
	w := httptest.NewRecorder()
	ph.GetPersonaLimits(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	maxPersonas := int(resp["max_personas"].(float64))
	if maxPersonas != 7 {
		t.Errorf("expected max_personas 7 for trusted, got %d", maxPersonas)
	}

	allowCustom := resp["allow_custom_categories"].(bool)
	if !allowCustom {
		t.Error("expected allow_custom_categories to be true for trusted")
	}
}

func TestGetPersonaLimitsWithPersonas(t *testing.T) {
	ph := NewProfileHandler()

	// Create some personas
	createTestPersona(t, ph, "test-user", "Pro")
	createTestPersona(t, ph, "test-user", "Personal")

	req := httptest.NewRequest(http.MethodGet, "/v2/users/persona-limits", nil)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()
	ph.GetPersonaLimits(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	currentCount := int(resp["current_count"].(float64))
	if currentCount != 2 {
		t.Errorf("expected current_count 2, got %d", currentCount)
	}

	remaining := int(resp["remaining"].(float64))
	if remaining != 0 {
		t.Errorf("expected remaining 0 for unverified with 2 personas, got %d", remaining)
	}
}

// ===== Persona Default Settings on Create =====

func TestPersonaCreatedWithDefaultSettings(t *testing.T) {
	ph := NewProfileHandler()
	personaID := createTestPersona(t, ph, "test-user", "Pro")

	// Check privacy settings have defaults
	privReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas/settings/privacy?id="+personaID, nil)
	privReq.Header.Set("X-User-ID", "test-user")
	pw := httptest.NewRecorder()
	ph.GetPersonaPrivacySettings(pw, privReq)

	var privacy PersonaPrivacySettingsData
	json.NewDecoder(pw.Body).Decode(&privacy)

	if !privacy.SendReadReceipts {
		t.Error("expected default send_read_receipts to be true")
	}
	if !privacy.SendTypingIndicators {
		t.Error("expected default send_typing_indicators to be true")
	}
	if !privacy.Searchable {
		t.Error("expected default searchable to be true")
	}

	// Check notification settings
	notifReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas/settings/notifications?id="+personaID, nil)
	notifReq.Header.Set("X-User-ID", "test-user")
	nw := httptest.NewRecorder()
	ph.GetPersonaNotificationSettings(nw, notifReq)

	var notif PersonaNotifSettingsData
	json.NewDecoder(nw.Body).Decode(&notif)

	if !notif.Enabled {
		t.Error("expected default notifications enabled to be true")
	}

	// Check feature settings
	featReq := httptest.NewRequest(http.MethodGet, "/v2/users/personas/settings/features?id="+personaID, nil)
	featReq.Header.Set("X-User-ID", "test-user")
	fw := httptest.NewRecorder()
	ph.GetPersonaFeatureSettings(fw, featReq)

	var features PersonaFeatureSettingsData
	json.NewDecoder(fw.Body).Decode(&features)

	if !features.VoiceCalls {
		t.Error("expected default voice_calls to be true")
	}
	if !features.FileSharing {
		t.Error("expected default file_sharing to be true")
	}
}
