package onboarding

import (
	"testing"
	"time"
)

func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{"+15551234567", true},
		{"+442071234567", true},
		{"+8613800138000", true},
		{"+1234567", true},    // minimum length
		{"5551234567", false},  // missing +
		{"+0551234567", false}, // starts with 0
		{"", false},
		{"+1", false},         // too short
		{"hello", false},
		{"+1abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			result := ValidatePhoneNumber(tt.phone)
			if result != tt.valid {
				t.Errorf("ValidatePhoneNumber(%q) = %v, want %v", tt.phone, result, tt.valid)
			}
		})
	}
}

func TestSendCode(t *testing.T) {
	svc := NewPhoneVerificationService()

	t.Run("valid phone", func(t *testing.T) {
		v, err := svc.SendCode("+15551234567")
		if err != nil {
			t.Fatalf("SendCode failed: %v", err)
		}
		if v.PhoneNumber != "+15551234567" {
			t.Errorf("phone = %s, want +15551234567", v.PhoneNumber)
		}
		if len(v.Code) != OTPLength {
			t.Errorf("code length = %d, want %d", len(v.Code), OTPLength)
		}
		if v.Verified {
			t.Error("should not be verified initially")
		}
		if v.ExpiresAt.Before(time.Now()) {
			t.Error("expiry should be in the future")
		}
	})

	t.Run("invalid phone", func(t *testing.T) {
		_, err := svc.SendCode("notaphone")
		if err != ErrInvalidPhoneNumber {
			t.Errorf("expected ErrInvalidPhoneNumber, got %v", err)
		}
	})

	t.Run("cooldown enforced", func(t *testing.T) {
		svc := NewPhoneVerificationService()
		svc.SendCode("+15559999999")
		_, err := svc.SendCode("+15559999999")
		if err != ErrOTPAlreadySent {
			t.Errorf("expected ErrOTPAlreadySent, got %v", err)
		}
	})

	t.Run("resend after cooldown", func(t *testing.T) {
		svc := NewPhoneVerificationService()
		svc.SendCode("+15558888888")

		// Simulate cooldown expiry
		svc.mu.Lock()
		svc.verifications["+15558888888"].CreatedAt = time.Now().Add(-1 * time.Minute)
		svc.mu.Unlock()

		v, err := svc.SendCode("+15558888888")
		if err != nil {
			t.Fatalf("resend after cooldown failed: %v", err)
		}
		if v.Resends != 1 {
			t.Errorf("resends = %d, want 1", v.Resends)
		}
	})

	t.Run("max resends enforced", func(t *testing.T) {
		svc := NewPhoneVerificationService()

		// Manually set up a verification that's hit the resend limit
		svc.mu.Lock()
		svc.verifications["+15557777777"] = &PhoneVerification{
			PhoneNumber: "+15557777777",
			Code:        "123456",
			CreatedAt:   time.Now().Add(-1 * time.Minute),
			ExpiresAt:   time.Now().Add(4 * time.Minute),
			Resends:     MaxResends,
		}
		svc.mu.Unlock()

		_, err := svc.SendCode("+15557777777")
		if err != ErrOTPRateLimited {
			t.Errorf("expected ErrOTPRateLimited, got %v", err)
		}
	})
}

func TestVerifyCode(t *testing.T) {
	t.Run("correct code", func(t *testing.T) {
		svc := NewPhoneVerificationService()
		v, _ := svc.SendCode("+15551111111")

		err := svc.VerifyCode("+15551111111", v.Code)
		if err != nil {
			t.Fatalf("VerifyCode failed: %v", err)
		}

		if !svc.IsVerified("+15551111111") {
			t.Error("phone should be verified")
		}
	})

	t.Run("wrong code", func(t *testing.T) {
		svc := NewPhoneVerificationService()
		svc.SendCode("+15552222222")

		err := svc.VerifyCode("+15552222222", "000000")
		if err != ErrOTPInvalid {
			t.Errorf("expected ErrOTPInvalid, got %v", err)
		}
	})

	t.Run("expired code", func(t *testing.T) {
		svc := NewPhoneVerificationService()
		v, _ := svc.SendCode("+15553333333")

		// Simulate expiry
		svc.mu.Lock()
		svc.verifications["+15553333333"].ExpiresAt = time.Now().Add(-1 * time.Second)
		svc.mu.Unlock()

		err := svc.VerifyCode("+15553333333", v.Code)
		if err != ErrOTPExpired {
			t.Errorf("expected ErrOTPExpired, got %v", err)
		}
	})

	t.Run("max attempts exceeded", func(t *testing.T) {
		svc := NewPhoneVerificationService()
		svc.SendCode("+15554444444")

		for i := 0; i < MaxOTPAttempts; i++ {
			svc.VerifyCode("+15554444444", "000000")
		}

		err := svc.VerifyCode("+15554444444", "000000")
		if err != ErrOTPMaxAttempts {
			t.Errorf("expected ErrOTPMaxAttempts, got %v", err)
		}
	})

	t.Run("unknown phone", func(t *testing.T) {
		svc := NewPhoneVerificationService()
		err := svc.VerifyCode("+15550000000", "123456")
		if err != ErrOTPInvalid {
			t.Errorf("expected ErrOTPInvalid, got %v", err)
		}
	})

	t.Run("already verified is idempotent", func(t *testing.T) {
		svc := NewPhoneVerificationService()
		v, _ := svc.SendCode("+15556666666")
		svc.VerifyCode("+15556666666", v.Code)

		// Second verify should succeed
		err := svc.VerifyCode("+15556666666", v.Code)
		if err != nil {
			t.Errorf("re-verify should succeed, got %v", err)
		}
	})
}

func TestIsVerified(t *testing.T) {
	svc := NewPhoneVerificationService()

	if svc.IsVerified("+15550000000") {
		t.Error("unknown phone should not be verified")
	}

	v, _ := svc.SendCode("+15551234567")
	if svc.IsVerified("+15551234567") {
		t.Error("should not be verified before code entry")
	}

	svc.VerifyCode("+15551234567", v.Code)
	if !svc.IsVerified("+15551234567") {
		t.Error("should be verified after correct code")
	}
}

func TestGetVerification(t *testing.T) {
	svc := NewPhoneVerificationService()

	_, err := svc.GetVerification("+15550000000")
	if err != ErrOTPInvalid {
		t.Errorf("expected ErrOTPInvalid for unknown phone, got %v", err)
	}

	svc.SendCode("+15551234567")
	v, err := svc.GetVerification("+15551234567")
	if err != nil {
		t.Fatalf("GetVerification failed: %v", err)
	}
	if v.PhoneNumber != "+15551234567" {
		t.Errorf("phone = %s, want +15551234567", v.PhoneNumber)
	}
}

func TestTimeUntilResend(t *testing.T) {
	svc := NewPhoneVerificationService()

	// No verification = no wait
	if d := svc.TimeUntilResend("+15550000000"); d != 0 {
		t.Errorf("expected 0 for unknown phone, got %v", d)
	}

	svc.SendCode("+15551234567")
	d := svc.TimeUntilResend("+15551234567")
	if d <= 0 || d > OTPCooldown {
		t.Errorf("cooldown = %v, expected between 0 and %v", d, OTPCooldown)
	}
}

func TestOTPRandomness(t *testing.T) {
	// Generate multiple OTPs and verify they're not all the same
	svc := NewPhoneVerificationService()
	codes := make(map[string]bool)

	for i := 0; i < 10; i++ {
		phone := "+1555000" + string(rune('0'+i)) + "000"
		v, err := svc.SendCode(phone)
		if err != nil {
			t.Fatalf("SendCode failed: %v", err)
		}
		codes[v.Code] = true
	}

	// With 10 random 6-digit codes, extremely unlikely they're all the same
	if len(codes) < 2 {
		t.Error("OTP codes should be random, got all identical codes")
	}
}

func BenchmarkSendCode(b *testing.B) {
	svc := NewPhoneVerificationService()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		phone := "+1555" + string(rune('0'+i%10)) + "234567"
		svc.SendCode(phone)
	}
}
