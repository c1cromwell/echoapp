package auth

import (
	"testing"
	"time"
)

func TestRecoveryService_InitiateRecoveryPhrase(t *testing.T) {
	rs := NewRecoveryService()

	session, err := rs.InitiateRecovery("user-1", RecoveryPhrase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Method != RecoveryPhrase {
		t.Errorf("expected method recovery_phrase, got %s", session.Method)
	}
	if session.Status != RecoveryPending {
		t.Errorf("expected pending status, got %s", session.Status)
	}
	if len(session.RequiredSteps) != 1 || session.RequiredSteps[0] != "sign_challenge" {
		t.Errorf("unexpected required steps: %v", session.RequiredSteps)
	}
	if session.ShardsRequired != 0 {
		t.Error("phrase recovery should not require shards")
	}
}

func TestRecoveryService_InitiateTrustedContacts(t *testing.T) {
	rs := NewRecoveryService()

	session, err := rs.InitiateRecovery("user-1", RecoveryTrustedContacts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.ShardsRequired != RecoveryShardThreshold {
		t.Errorf("expected %d shards required, got %d", RecoveryShardThreshold, session.ShardsRequired)
	}
	if len(session.RequiredSteps) != 2 {
		t.Errorf("trusted contacts should have 2 required steps, got %d", len(session.RequiredSteps))
	}
}

func TestRecoveryService_InitiatePhone(t *testing.T) {
	rs := NewRecoveryService()

	session, err := rs.InitiateRecovery("user-1", RecoveryPhone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(session.RequiredSteps) != 2 {
		t.Errorf("phone recovery should have 2 steps, got %d", len(session.RequiredSteps))
	}
	if session.RequiredSteps[0] != "verify_otp" {
		t.Errorf("first step should be verify_otp, got %s", session.RequiredSteps[0])
	}
}

func TestRecoveryService_InvalidMethod(t *testing.T) {
	rs := NewRecoveryService()

	_, err := rs.InitiateRecovery("user-1", RecoveryMethod("invalid"))
	if err == nil {
		t.Error("invalid method should return error")
	}
	if err.Code != ErrCodeRecoveryInvalid {
		t.Errorf("expected AUTH_011, got %s", err.Code)
	}
}

func TestRecoveryService_GetSession(t *testing.T) {
	rs := NewRecoveryService()

	session, _ := rs.InitiateRecovery("user-1", RecoveryPhrase)
	retrieved, err := rs.GetSession(session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != session.ID {
		t.Error("retrieved session should match")
	}
}

func TestRecoveryService_GetSessionNotFound(t *testing.T) {
	rs := NewRecoveryService()

	_, err := rs.GetSession("nonexistent")
	if err == nil {
		t.Error("should return error for nonexistent session")
	}
}

func TestRecoveryService_TrustedContactShards(t *testing.T) {
	rs := NewRecoveryService()

	session, _ := rs.InitiateRecovery("user-1", RecoveryTrustedContacts)

	// First shard - not enough
	met, err := rs.SubmitContactShard(session.ID, "contact-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if met {
		t.Error("1 shard should not meet threshold")
	}

	// Duplicate shard - ignored
	met, _ = rs.SubmitContactShard(session.ID, "contact-1")
	if met {
		t.Error("duplicate shard should not count")
	}

	// Second shard - threshold met (2-of-3)
	met, err = rs.SubmitContactShard(session.ID, "contact-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !met {
		t.Error("2 shards should meet threshold")
	}

	// Session should be in verifying state
	s, _ := rs.GetSession(session.ID)
	if s.Status != RecoveryVerifying {
		t.Errorf("expected verifying status, got %s", s.Status)
	}
}

func TestRecoveryService_ShardWrongMethod(t *testing.T) {
	rs := NewRecoveryService()

	session, _ := rs.InitiateRecovery("user-1", RecoveryPhrase)
	_, err := rs.SubmitContactShard(session.ID, "contact-1")
	if err == nil {
		t.Error("submitting shard to non-trusted-contacts session should fail")
	}
}

func TestRecoveryService_CompleteStep(t *testing.T) {
	rs := NewRecoveryService()

	session, _ := rs.InitiateRecovery("user-1", RecoveryPhrase)

	err := rs.CompleteStep(session.ID, "sign_challenge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !rs.IsComplete(session.ID) {
		t.Error("session should be complete after all steps done")
	}
}

func TestRecoveryService_CompleteStepInvalid(t *testing.T) {
	rs := NewRecoveryService()

	session, _ := rs.InitiateRecovery("user-1", RecoveryPhrase)

	err := rs.CompleteStep(session.ID, "nonexistent_step")
	if err == nil {
		t.Error("invalid step should return error")
	}
}

func TestRecoveryService_CompleteStepIdempotent(t *testing.T) {
	rs := NewRecoveryService()

	session, _ := rs.InitiateRecovery("user-1", RecoveryPhrase)

	rs.CompleteStep(session.ID, "sign_challenge")
	err := rs.CompleteStep(session.ID, "sign_challenge")
	if err != nil {
		t.Error("completing same step twice should be idempotent")
	}
}

func TestRecoveryService_MultiStepCompletion(t *testing.T) {
	rs := NewRecoveryService()

	session, _ := rs.InitiateRecovery("user-1", RecoveryPhone)

	// Complete first step only
	rs.CompleteStep(session.ID, "verify_otp")
	if rs.IsComplete(session.ID) {
		t.Error("should not be complete with only 1 of 2 steps")
	}

	// Complete second step
	rs.CompleteStep(session.ID, "register_passkey")
	if !rs.IsComplete(session.ID) {
		t.Error("should be complete after both steps")
	}
}

func TestRecoveryService_SessionExpiry(t *testing.T) {
	rs := NewRecoveryService()

	session, _ := rs.InitiateRecovery("user-1", RecoveryPhrase)

	// Force expire
	rs.mu.Lock()
	rs.sessions[session.ID].ExpiresAt = time.Now().Add(-time.Second)
	rs.mu.Unlock()

	_, err := rs.GetSession(session.ID)
	if err == nil {
		t.Error("expired session should return error")
	}
}

func TestRecoveryService_CleanExpired(t *testing.T) {
	rs := NewRecoveryService()

	s1, _ := rs.InitiateRecovery("user-1", RecoveryPhrase)
	rs.InitiateRecovery("user-2", RecoveryPhone)

	// Expire one session
	rs.mu.Lock()
	rs.sessions[s1.ID].ExpiresAt = time.Now().Add(-time.Second)
	rs.mu.Unlock()

	cleaned := rs.CleanExpired()
	if cleaned != 1 {
		t.Errorf("expected 1 cleaned, got %d", cleaned)
	}
	if rs.SessionCount() != 1 {
		t.Errorf("expected 1 remaining session, got %d", rs.SessionCount())
	}
}

func TestRecoveryService_SessionCount(t *testing.T) {
	rs := NewRecoveryService()

	if rs.SessionCount() != 0 {
		t.Error("should start with 0 sessions")
	}

	rs.InitiateRecovery("user-1", RecoveryPhrase)
	rs.InitiateRecovery("user-2", RecoveryPhone)

	if rs.SessionCount() != 2 {
		t.Errorf("expected 2 sessions, got %d", rs.SessionCount())
	}
}

func TestRecoveryConstants(t *testing.T) {
	if RecoverySessionTTL != 24*time.Hour {
		t.Errorf("expected 24hr TTL, got %v", RecoverySessionTTL)
	}
	if RecoveryShardThreshold != 2 {
		t.Errorf("expected 2-of-3 threshold, got %d", RecoveryShardThreshold)
	}
	if RecoveryMinContacts != 3 {
		t.Errorf("expected min 3 contacts, got %d", RecoveryMinContacts)
	}
}
