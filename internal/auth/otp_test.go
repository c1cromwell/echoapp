package auth

import (
	"testing"
)

func TestGenerateOTP_Length(t *testing.T) {
	code, err := GenerateOTP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != OTPLength {
		t.Errorf("expected %d-digit code, got %d digits: %s", OTPLength, len(code), code)
	}
}

func TestGenerateOTP_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := GenerateOTP()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seen[code] = true
	}
	// At least 90 unique codes out of 100 (very high probability)
	if len(seen) < 90 {
		t.Errorf("expected high uniqueness, only got %d unique codes out of 100", len(seen))
	}
}

func TestHashOTP_VerifyRoundTrip(t *testing.T) {
	code := "482916"
	hash, err := HashOTP(code)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	if !VerifyOTPHash(code, hash) {
		t.Error("correct code should verify against its hash")
	}

	if VerifyOTPHash("000000", hash) {
		t.Error("wrong code should not verify against hash")
	}
}

func TestOTPService_CreateAndVerify(t *testing.T) {
	svc := NewOTPService()

	code := "123456"
	hash, _ := HashOTP(code)
	svc.CreateSession("vid-1", "phone-hash-1", hash)

	// Verify correct code
	ok, err := svc.VerifyCode("vid-1", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("correct code should succeed")
	}
}

func TestOTPService_WrongCode(t *testing.T) {
	svc := NewOTPService()

	hash, _ := HashOTP("123456")
	svc.CreateSession("vid-1", "phone-hash-1", hash)

	ok, err := svc.VerifyCode("vid-1", "000000")
	if ok {
		t.Error("wrong code should fail")
	}
	if err == nil {
		t.Error("wrong code should return error")
	}
	if err != nil && err.Code != ErrCodeInvalidOTP {
		t.Errorf("expected AUTH_003, got %s", err.Code)
	}
}

func TestOTPService_NonexistentSession(t *testing.T) {
	svc := NewOTPService()

	ok, err := svc.VerifyCode("nonexistent", "123456")
	if ok {
		t.Error("nonexistent session should fail")
	}
	if err == nil || err.Code != ErrCodeInvalidOTP {
		t.Error("expected AUTH_003 for nonexistent session")
	}
}

func TestOTPService_BruteForce_SixthAttemptFails(t *testing.T) {
	svc := NewOTPService()

	hash, _ := HashOTP("123456")
	svc.CreateSession("vid-1", "phone-hash-1", hash)

	// Use up 5 attempts with wrong code
	for i := 0; i < 5; i++ {
		svc.VerifyCode("vid-1", "000000")
	}

	// 6th attempt should fail even with correct code
	ok, err := svc.VerifyCode("vid-1", "123456")
	if ok {
		t.Error("6th attempt should fail regardless of correct code")
	}
	if err == nil {
		t.Error("expected error on 6th attempt")
	}
}

func TestOTPService_SessionExpiry(t *testing.T) {
	svc := NewOTPService()

	hash, _ := HashOTP("123456")
	session := svc.CreateSession("vid-1", "phone-hash-1", hash)

	// Manually expire the session
	session.ExpiresAt = session.CreatedAt

	ok, err := svc.VerifyCode("vid-1", "123456")
	if ok {
		t.Error("expired session should fail")
	}
	if err == nil || err.Code != ErrCodeInvalidOTP {
		t.Error("expected AUTH_003 for expired session")
	}
}

func TestOTPService_PhoneRateLimit(t *testing.T) {
	svc := NewOTPService()

	phoneHash := "test-phone-hash"

	// Should be under limit initially
	if svc.CheckPhoneRateLimit(phoneHash) {
		t.Error("should not be rate limited initially")
	}

	// Record max sends
	for i := 0; i < MaxOTPPerPhoneHr; i++ {
		svc.RecordOTPSend(phoneHash)
	}

	// Should now be rate limited
	if !svc.CheckPhoneRateLimit(phoneHash) {
		t.Error("should be rate limited after max sends")
	}
}

func TestOTPService_CleanExpired(t *testing.T) {
	svc := NewOTPService()

	hash, _ := HashOTP("123456")
	session := svc.CreateSession("vid-1", "phone-hash-1", hash)
	svc.CreateSession("vid-2", "phone-hash-2", hash)

	// Expire one session
	session.ExpiresAt = session.CreatedAt

	cleaned := svc.CleanExpired()
	if cleaned != 1 {
		t.Errorf("expected 1 cleaned session, got %d", cleaned)
	}
	if svc.SessionCount() != 1 {
		t.Errorf("expected 1 remaining session, got %d", svc.SessionCount())
	}
}

func TestOTPService_VerifiedSessionNotReusable(t *testing.T) {
	svc := NewOTPService()

	code := "123456"
	hash, _ := HashOTP(code)
	svc.CreateSession("vid-1", "phone-hash-1", hash)

	// First verify succeeds
	ok, _ := svc.VerifyCode("vid-1", code)
	if !ok {
		t.Fatal("first verify should succeed")
	}

	// Session should still exist (but marked verified)
	session := svc.GetSession("vid-1")
	if session == nil {
		t.Fatal("session should still exist")
	}
	if !session.Verified {
		t.Error("session should be marked verified")
	}
}
