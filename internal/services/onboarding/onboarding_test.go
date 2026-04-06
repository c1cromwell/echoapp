package onboarding

import (
	"fmt"
	"testing"
	"time"
)

func newTestOnboardingService() *OnboardingService {
	phone := NewPhoneVerificationService()
	passkey := NewPasskeyService()
	recovery := NewRecoveryService()
	return NewOnboardingService(phone, passkey, recovery)
}

func TestStartSession(t *testing.T) {
	svc := newTestOnboardingService()

	t.Run("phone registration", func(t *testing.T) {
		session, err := svc.StartSession(RegistrationPhone)
		if err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}
		if session.ID == "" {
			t.Error("session ID is empty")
		}
		if session.RegistrationMethod != RegistrationPhone {
			t.Errorf("method = %s, want phone", session.RegistrationMethod)
		}
		if session.CurrentStep != StepCarousel {
			t.Errorf("step = %s, want carousel", session.CurrentStep)
		}
		if session.Steps[StepCarousel] != StatusActive {
			t.Error("carousel should be active")
		}
	})

	t.Run("verifiable ID registration", func(t *testing.T) {
		session, err := svc.StartSession(RegistrationVerifiableID)
		if err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}
		if session.RegistrationMethod != RegistrationVerifiableID {
			t.Errorf("method = %s, want verifiable_id", session.RegistrationMethod)
		}
	})
}

func TestGetSession(t *testing.T) {
	svc := newTestOnboardingService()

	t.Run("existing session", func(t *testing.T) {
		created, _ := svc.StartSession(RegistrationPhone)
		session, err := svc.GetSession(created.ID)
		if err != nil {
			t.Fatalf("GetSession failed: %v", err)
		}
		if session.ID != created.ID {
			t.Errorf("ID mismatch")
		}
	})

	t.Run("nonexistent session", func(t *testing.T) {
		_, err := svc.GetSession("nonexistent")
		if err != ErrSessionNotFound {
			t.Errorf("expected ErrSessionNotFound, got %v", err)
		}
	})

	t.Run("expired session", func(t *testing.T) {
		session, _ := svc.StartSession(RegistrationPhone)
		svc.mu.Lock()
		svc.sessions[session.ID].CreatedAt = time.Now().Add(-25 * time.Hour)
		svc.mu.Unlock()

		_, err := svc.GetSession(session.ID)
		if err != ErrSessionExpired {
			t.Errorf("expected ErrSessionExpired, got %v", err)
		}
	})
}

func TestCarouselFlow(t *testing.T) {
	svc := newTestOnboardingService()

	t.Run("complete carousel", func(t *testing.T) {
		session, _ := svc.StartSession(RegistrationPhone)
		err := svc.CompleteCarousel(session.ID)
		if err != nil {
			t.Fatalf("CompleteCarousel failed: %v", err)
		}

		s, _ := svc.GetSession(session.ID)
		if !s.CarouselViewed {
			t.Error("carousel should be marked as viewed")
		}
		if s.CurrentStep != StepWelcome {
			t.Errorf("step = %s, want welcome", s.CurrentStep)
		}
		if s.Steps[StepCarousel] != StatusCompleted {
			t.Error("carousel should be completed")
		}
	})

	t.Run("skip carousel", func(t *testing.T) {
		session, _ := svc.StartSession(RegistrationPhone)
		err := svc.SkipCarousel(session.ID)
		if err != nil {
			t.Fatalf("SkipCarousel failed: %v", err)
		}

		s, _ := svc.GetSession(session.ID)
		if s.Steps[StepCarousel] != StatusSkipped {
			t.Error("carousel should be skipped")
		}
		if s.CurrentStep != StepWelcome {
			t.Errorf("step = %s, want welcome", s.CurrentStep)
		}
	})
}

func TestPhoneEntryFlow(t *testing.T) {
	svc := newTestOnboardingService()
	session, _ := svc.StartSession(RegistrationPhone)
	svc.CompleteCarousel(session.ID)

	t.Run("start phone entry", func(t *testing.T) {
		err := svc.StartPhoneEntry(session.ID)
		if err != nil {
			t.Fatalf("StartPhoneEntry failed: %v", err)
		}
		s, _ := svc.GetSession(session.ID)
		if s.CurrentStep != StepPhoneEntry {
			t.Errorf("step = %s, want phone_entry", s.CurrentStep)
		}
	})

	t.Run("submit phone", func(t *testing.T) {
		v, err := svc.SubmitPhone(session.ID, "+15551234567")
		if err != nil {
			t.Fatalf("SubmitPhone failed: %v", err)
		}
		if v == nil {
			t.Fatal("verification is nil")
		}

		s, _ := svc.GetSession(session.ID)
		if s.PhoneNumber != "+15551234567" {
			t.Errorf("phone = %s, want +15551234567", s.PhoneNumber)
		}
		if s.CurrentStep != StepOTP {
			t.Errorf("step = %s, want otp_verification", s.CurrentStep)
		}
	})

	t.Run("verify OTP", func(t *testing.T) {
		// Get the code from phone service
		v, _ := svc.phone.GetVerification("+15551234567")
		err := svc.VerifyOTP(session.ID, v.Code)
		if err != nil {
			t.Fatalf("VerifyOTP failed: %v", err)
		}

		s, _ := svc.GetSession(session.ID)
		if s.CurrentStep != StepPasskey {
			t.Errorf("step = %s, want passkey_setup", s.CurrentStep)
		}
		if s.Steps[StepOTP] != StatusCompleted {
			t.Error("OTP step should be completed")
		}
	})
}

func TestPhoneEntryStepNotReady(t *testing.T) {
	svc := newTestOnboardingService()
	session, _ := svc.StartSession(RegistrationPhone)

	// Try to start phone entry without completing carousel
	err := svc.StartPhoneEntry(session.ID)
	if err != ErrStepNotReady {
		t.Errorf("expected ErrStepNotReady, got %v", err)
	}
}

func TestPasskeyFlow(t *testing.T) {
	svc := newTestOnboardingService()
	session := advanceToPasskey(t, svc)

	t.Run("setup and complete passkey", func(t *testing.T) {
		challenge, err := svc.SetupPasskey(session.ID)
		if err != nil {
			t.Fatalf("SetupPasskey failed: %v", err)
		}

		err = svc.CompletePasskey(session.ID, challenge.ID, "cred-id", []byte("public-key-data"), PasskeyFaceID, "iPhone 15")
		if err != nil {
			t.Fatalf("CompletePasskey failed: %v", err)
		}

		s, _ := svc.GetSession(session.ID)
		if s.CurrentStep != StepRecovery {
			t.Errorf("step = %s, want recovery_setup", s.CurrentStep)
		}
		if s.DID == "" {
			t.Error("DID should be generated")
		}
		if s.Steps[StepPasskey] != StatusCompleted {
			t.Error("passkey step should be completed")
		}
	})

	t.Run("skip passkey", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToPasskey(t, svc)

		err := svc.SkipPasskey(session.ID)
		if err != nil {
			t.Fatalf("SkipPasskey failed: %v", err)
		}

		s, _ := svc.GetSession(session.ID)
		if s.Steps[StepPasskey] != StatusSkipped {
			t.Error("passkey should be skipped")
		}
		if s.CurrentStep != StepRecovery {
			t.Errorf("step = %s, want recovery_setup", s.CurrentStep)
		}
		if len(s.SkippedSteps) != 1 || s.SkippedSteps[0] != StepPasskey {
			t.Error("passkey should be in skipped steps")
		}
	})
}

func TestRecoveryFlow(t *testing.T) {
	t.Run("email recovery", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToRecovery(t, svc)

		method, err := svc.SetupRecoveryEmail(session.ID, "backup@example.com")
		if err != nil {
			t.Fatalf("SetupRecoveryEmail failed: %v", err)
		}
		if method.Type != RecoveryEmail {
			t.Errorf("type = %s, want email", method.Type)
		}

		s, _ := svc.GetSession(session.ID)
		if s.CurrentStep != StepProfile {
			t.Errorf("step = %s, want profile_setup", s.CurrentStep)
		}
	})

	t.Run("wallet recovery", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToRecovery(t, svc)

		method, err := svc.SetupRecoveryWallet(session.ID, "0x1234567890abcdef1234567890abcdef12345678")
		if err != nil {
			t.Fatalf("SetupRecoveryWallet failed: %v", err)
		}
		if method.Type != RecoveryWallet {
			t.Errorf("type = %s, want wallet", method.Type)
		}
	})

	t.Run("trusted contact recovery", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToRecovery(t, svc)

		method, err := svc.SetupRecoveryContact(session.ID, "did:echo:trusted-friend")
		if err != nil {
			t.Fatalf("SetupRecoveryContact failed: %v", err)
		}
		if method.Type != RecoveryContact {
			t.Errorf("type = %s, want trusted_contact", method.Type)
		}
	})

	t.Run("skip recovery", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToRecovery(t, svc)

		err := svc.SkipRecovery(session.ID)
		if err != nil {
			t.Fatalf("SkipRecovery failed: %v", err)
		}

		s, _ := svc.GetSession(session.ID)
		if s.Steps[StepRecovery] != StatusSkipped {
			t.Error("recovery should be skipped")
		}
		if s.CurrentStep != StepProfile {
			t.Errorf("step = %s, want profile_setup", s.CurrentStep)
		}
	})
}

func TestProfileFlow(t *testing.T) {
	t.Run("complete profile", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToProfile(t, svc)

		profile := &Profile{
			DisplayName: "Alex",
			Username:    "alex_echo",
			Bio:         "Hello world",
		}
		err := svc.SetupProfile(session.ID, profile)
		if err != nil {
			t.Fatalf("SetupProfile failed: %v", err)
		}

		s, _ := svc.GetSession(session.ID)
		if s.CurrentStep != StepTrustDash {
			t.Errorf("step = %s, want trust_dashboard", s.CurrentStep)
		}
		if s.Profile.DisplayName != "Alex" {
			t.Errorf("display name = %s, want Alex", s.Profile.DisplayName)
		}
	})

	t.Run("display name required", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToProfile(t, svc)

		err := svc.SetupProfile(session.ID, &Profile{})
		if err != ErrDisplayNameRequired {
			t.Errorf("expected ErrDisplayNameRequired, got %v", err)
		}
	})

	t.Run("display name too long", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToProfile(t, svc)

		longName := ""
		for i := 0; i < MaxDisplayNameLen+1; i++ {
			longName += "a"
		}
		err := svc.SetupProfile(session.ID, &Profile{DisplayName: longName})
		if err != ErrDisplayNameTooLong {
			t.Errorf("expected ErrDisplayNameTooLong, got %v", err)
		}
	})

	t.Run("bio too long", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToProfile(t, svc)

		longBio := ""
		for i := 0; i < MaxBioLen+1; i++ {
			longBio += "a"
		}
		err := svc.SetupProfile(session.ID, &Profile{DisplayName: "Alex", Bio: longBio})
		if err != ErrBioTooLong {
			t.Errorf("expected ErrBioTooLong, got %v", err)
		}
	})

	t.Run("invalid username", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToProfile(t, svc)

		err := svc.SetupProfile(session.ID, &Profile{DisplayName: "Alex", Username: "has spaces"})
		if err != ErrUsernameInvalid {
			t.Errorf("expected ErrUsernameInvalid, got %v", err)
		}
	})

	t.Run("username taken", func(t *testing.T) {
		svc := newTestOnboardingService()

		// First user takes the name
		s1 := advanceToProfile(t, svc)
		svc.SetupProfile(s1.ID, &Profile{DisplayName: "First", Username: "taken_name"})

		// Second user tries same name
		s2 := advanceToProfile(t, svc)
		err := svc.SetupProfile(s2.ID, &Profile{DisplayName: "Second", Username: "taken_name"})
		if err != ErrUsernameTaken {
			t.Errorf("expected ErrUsernameTaken, got %v", err)
		}
	})

	t.Run("optional username", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToProfile(t, svc)

		err := svc.SetupProfile(session.ID, &Profile{DisplayName: "NoUsername"})
		if err != nil {
			t.Fatalf("profile without username should succeed, got %v", err)
		}
	})

	t.Run("skip profile", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToProfile(t, svc)

		err := svc.SkipProfile(session.ID)
		if err != nil {
			t.Fatalf("SkipProfile failed: %v", err)
		}

		s, _ := svc.GetSession(session.ID)
		if s.CurrentStep != StepTrustDash {
			t.Errorf("step = %s, want trust_dashboard", s.CurrentStep)
		}
	})
}

func TestCheckUsernameAvailability(t *testing.T) {
	svc := newTestOnboardingService()

	t.Run("available", func(t *testing.T) {
		available, err := svc.CheckUsernameAvailability("new_user")
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if !available {
			t.Error("should be available")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := svc.CheckUsernameAvailability("ab") // too short
		if err != ErrUsernameInvalid {
			t.Errorf("expected ErrUsernameInvalid, got %v", err)
		}
	})

	t.Run("taken after registration", func(t *testing.T) {
		session := advanceToProfile(t, svc)
		svc.SetupProfile(session.ID, &Profile{DisplayName: "Test", Username: "claimed_name"})

		available, _ := svc.CheckUsernameAvailability("claimed_name")
		if available {
			t.Error("should not be available after claimed")
		}
	})
}

func TestSecuritySetup(t *testing.T) {
	svc := newTestOnboardingService()

	t.Run("after phone verification only", func(t *testing.T) {
		session := advanceToPasskey(t, svc)
		svc.SkipPasskey(session.ID)
		svc.SkipRecovery(session.ID)
		svc.SkipProfile(session.ID)

		items, err := svc.GetSecuritySetup(session.ID)
		if err != nil {
			t.Fatalf("GetSecuritySetup failed: %v", err)
		}
		if len(items) != 4 {
			t.Fatalf("expected 4 items, got %d", len(items))
		}
		// Phone should be completed
		if !items[0].Completed {
			t.Error("phone should be completed")
		}
		// Passkey should not be completed
		if items[1].Completed {
			t.Error("passkey should not be completed (was skipped)")
		}
		// Recovery should not be completed
		if items[2].Completed {
			t.Error("recovery should not be completed (was skipped)")
		}
	})

	t.Run("fully completed", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToRecovery(t, svc)
		svc.SetupRecoveryEmail(session.ID, "test@example.com")
		svc.SetupProfile(session.ID, &Profile{DisplayName: "Full"})

		items, _ := svc.GetSecuritySetup(session.ID)
		completed := 0
		for _, item := range items {
			if item.Completed {
				completed++
			}
		}
		// Phone + Passkey + Recovery = 3 completed (Identity verified is always incomplete here)
		if completed != 3 {
			t.Errorf("expected 3 completed items, got %d", completed)
		}
	})
}

func TestGetTrustScore(t *testing.T) {
	t.Run("minimal score", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToPasskey(t, svc)
		svc.SkipPasskey(session.ID)
		svc.SkipRecovery(session.ID)
		svc.SkipProfile(session.ID)

		score, level, err := svc.GetTrustScore(session.ID)
		if err != nil {
			t.Fatalf("GetTrustScore failed: %v", err)
		}
		if score != 5 { // Phone only
			t.Errorf("score = %d, want 5", score)
		}
		if level != "Newcomer" {
			t.Errorf("level = %s, want Newcomer", level)
		}
	})

	t.Run("full onboarding score", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToRecovery(t, svc)
		svc.SetupRecoveryEmail(session.ID, "test@example.com")

		score, _, err := svc.GetTrustScore(session.ID)
		if err != nil {
			t.Fatalf("GetTrustScore failed: %v", err)
		}
		if score != 15 { // Phone + Passkey + Recovery
			t.Errorf("score = %d, want 15", score)
		}
	})
}

func TestCompleteOnboarding(t *testing.T) {
	t.Run("successful completion", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToProfile(t, svc)
		svc.SetupProfile(session.ID, &Profile{DisplayName: "Done"})

		completed, err := svc.CompleteOnboarding(session.ID)
		if err != nil {
			t.Fatalf("CompleteOnboarding failed: %v", err)
		}
		if completed.CompletedAt == nil {
			t.Error("completedAt should be set")
		}
		if completed.CurrentStep != StepComplete {
			t.Errorf("step = %s, want complete", completed.CurrentStep)
		}
	})

	t.Run("requires phone verification", func(t *testing.T) {
		svc := newTestOnboardingService()
		session, _ := svc.StartSession(RegistrationPhone)
		svc.CompleteCarousel(session.ID)

		_, err := svc.CompleteOnboarding(session.ID)
		if err != ErrStepNotReady {
			t.Errorf("expected ErrStepNotReady, got %v", err)
		}
	})
}

func TestSkippedSteps(t *testing.T) {
	svc := newTestOnboardingService()
	session := advanceToPasskey(t, svc)
	svc.SkipPasskey(session.ID)
	svc.SkipRecovery(session.ID)

	skipped, err := svc.GetSkippedSteps(session.ID)
	if err != nil {
		t.Fatalf("GetSkippedSteps failed: %v", err)
	}
	if len(skipped) != 2 {
		t.Errorf("expected 2 skipped steps, got %d", len(skipped))
	}
}

func TestHasNudgeCard(t *testing.T) {
	t.Run("no skipped steps", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToRecovery(t, svc)
		svc.SetupRecoveryEmail(session.ID, "test@example.com")

		hasNudge, _ := svc.HasNudgeCard(session.ID)
		if hasNudge {
			t.Error("should not have nudge card with no skipped steps")
		}
	})

	t.Run("with skipped steps", func(t *testing.T) {
		svc := newTestOnboardingService()
		session := advanceToPasskey(t, svc)
		svc.SkipPasskey(session.ID)

		hasNudge, _ := svc.HasNudgeCard(session.ID)
		if !hasNudge {
			t.Error("should have nudge card with skipped steps")
		}
	})
}

func TestFullOnboardingFlow(t *testing.T) {
	svc := newTestOnboardingService()

	// Step 0: Start
	session, err := svc.StartSession(RegistrationPhone)
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	// Step 0: Carousel
	if err := svc.CompleteCarousel(session.ID); err != nil {
		t.Fatalf("carousel: %v", err)
	}

	// Step 1: Welcome -> Phone Entry
	if err := svc.StartPhoneEntry(session.ID); err != nil {
		t.Fatalf("phone entry: %v", err)
	}

	// Step 2: Phone number
	v, err := svc.SubmitPhone(session.ID, "+15551234567")
	if err != nil {
		t.Fatalf("submit phone: %v", err)
	}

	// Step 3: OTP
	if err := svc.VerifyOTP(session.ID, v.Code); err != nil {
		t.Fatalf("verify otp: %v", err)
	}

	// Step 4: Passkey
	challenge, err := svc.SetupPasskey(session.ID)
	if err != nil {
		t.Fatalf("setup passkey: %v", err)
	}
	if err := svc.CompletePasskey(session.ID, challenge.ID, "cred-final", []byte("final-key"), PasskeyFaceID, "iPhone"); err != nil {
		t.Fatalf("complete passkey: %v", err)
	}

	// Step 4B: Recovery
	_, err = svc.SetupRecoveryEmail(session.ID, "recovery@example.com")
	if err != nil {
		t.Fatalf("recovery: %v", err)
	}

	// Step 5: Profile
	if err := svc.SetupProfile(session.ID, &Profile{
		DisplayName: "Alex",
		Username:    "alex_echo",
		Bio:         "Privacy-first messaging",
	}); err != nil {
		t.Fatalf("profile: %v", err)
	}

	// Step 6: Trust Dashboard
	items, err := svc.GetSecuritySetup(session.ID)
	if err != nil {
		t.Fatalf("security setup: %v", err)
	}

	completedCount := 0
	for _, item := range items {
		if item.Completed {
			completedCount++
		}
	}
	if completedCount != 3 { // Phone + Passkey + Recovery
		t.Errorf("completed = %d, want 3", completedCount)
	}

	score, level, _ := svc.GetTrustScore(session.ID)
	if score != 15 {
		t.Errorf("trust score = %d, want 15", score)
	}
	if level != "Newcomer" {
		t.Errorf("level = %s, want Newcomer", level)
	}

	// Complete
	result, err := svc.CompleteOnboarding(session.ID)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if result.CompletedAt == nil {
		t.Error("completedAt should be set")
	}
	if result.DID == "" {
		t.Error("DID should be set")
	}

	// Nudge card should not show
	hasNudge, _ := svc.HasNudgeCard(session.ID)
	if hasNudge {
		t.Error("full flow should not trigger nudge card")
	}
}

func TestMinimalOnboardingFlow(t *testing.T) {
	svc := newTestOnboardingService()

	session, _ := svc.StartSession(RegistrationPhone)
	svc.SkipCarousel(session.ID)
	svc.StartPhoneEntry(session.ID)

	v, _ := svc.SubmitPhone(session.ID, "+15559876543")
	svc.VerifyOTP(session.ID, v.Code)

	svc.SkipPasskey(session.ID)
	svc.SkipRecovery(session.ID)
	svc.SkipProfile(session.ID)

	result, err := svc.CompleteOnboarding(session.ID)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if result.CompletedAt == nil {
		t.Error("completedAt should be set")
	}

	// Should show nudge card for skipped steps
	hasNudge, _ := svc.HasNudgeCard(session.ID)
	if !hasNudge {
		t.Error("skipped flow should trigger nudge card")
	}

	skipped, _ := svc.GetSkippedSteps(session.ID)
	if len(skipped) != 3 { // passkey + recovery + profile
		t.Errorf("expected 3 skipped steps, got %d", len(skipped))
	}
}

// phoneCounter generates unique phone numbers for tests
var phoneCounter int

func nextTestPhone() string {
	phoneCounter++
	return fmt.Sprintf("+1555%07d", phoneCounter)
}

// Test helpers to advance session to specific steps

func advanceToPasskey(t *testing.T, svc *OnboardingService) *OnboardingSession {
	t.Helper()
	session, _ := svc.StartSession(RegistrationPhone)
	svc.CompleteCarousel(session.ID)
	svc.StartPhoneEntry(session.ID)
	phone := nextTestPhone()
	v, err := svc.SubmitPhone(session.ID, phone)
	if err != nil {
		t.Fatalf("SubmitPhone failed: %v", err)
	}
	if err := svc.VerifyOTP(session.ID, v.Code); err != nil {
		t.Fatalf("VerifyOTP failed: %v", err)
	}
	s, _ := svc.GetSession(session.ID)
	return s
}

func advanceToRecovery(t *testing.T, svc *OnboardingService) *OnboardingSession {
	t.Helper()
	session := advanceToPasskey(t, svc)
	challenge, _ := svc.SetupPasskey(session.ID)
	credID := "cred-" + session.ID[:8]
	svc.CompletePasskey(session.ID, challenge.ID, credID, []byte("key-"+session.ID[:8]), PasskeyFaceID, "Test Device")
	s, _ := svc.GetSession(session.ID)
	return s
}

func advanceToProfile(t *testing.T, svc *OnboardingService) *OnboardingSession {
	t.Helper()
	session := advanceToRecovery(t, svc)
	svc.SetupRecoveryEmail(session.ID, session.ID[:8]+"@test.com")
	s, _ := svc.GetSession(session.ID)
	return s
}

func BenchmarkFullOnboarding(b *testing.B) {
	for i := 0; i < b.N; i++ {
		svc := newTestOnboardingService()
		session, _ := svc.StartSession(RegistrationPhone)
		svc.CompleteCarousel(session.ID)
		svc.StartPhoneEntry(session.ID)
		v, _ := svc.SubmitPhone(session.ID, "+15551234567")
		svc.VerifyOTP(session.ID, v.Code)
		svc.SkipPasskey(session.ID)
		svc.SkipRecovery(session.ID)
		svc.SkipProfile(session.ID)
		svc.CompleteOnboarding(session.ID)
	}
}
