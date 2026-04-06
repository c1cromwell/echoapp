package auth

import (
	"testing"
	"time"
)

func TestAuthRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewAuthRateLimiter()

	cfg := AuthRateLimitConfig{
		Name:   "test",
		Limit:  5,
		Window: time.Minute,
	}

	for i := 0; i < 5; i++ {
		if err := rl.Check("key-1", cfg); err != nil {
			t.Fatalf("request %d should be allowed: %v", i, err)
		}
	}
}

func TestAuthRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewAuthRateLimiter()

	cfg := AuthRateLimitConfig{
		Name:   "test",
		Limit:  3,
		Window: time.Minute,
	}

	for i := 0; i < 3; i++ {
		rl.Check("key-1", cfg)
	}

	err := rl.Check("key-1", cfg)
	if err == nil {
		t.Error("should be rate limited")
	}
	if err.Code != ErrCodeGlobalRateLimit {
		t.Errorf("expected AUTH_012, got %s", err.Code)
	}
	if err.RetryAfter == nil {
		t.Error("should include retry_after")
	}
}

func TestAuthRateLimiter_SeparateKeys(t *testing.T) {
	rl := NewAuthRateLimiter()

	cfg := AuthRateLimitConfig{
		Name:   "test",
		Limit:  2,
		Window: time.Minute,
	}

	// Exhaust key-1
	rl.Check("key-1", cfg)
	rl.Check("key-1", cfg)

	// key-2 should still be allowed
	if err := rl.Check("key-2", cfg); err != nil {
		t.Error("different key should not be rate limited")
	}
}

func TestAuthRateLimiter_Remaining(t *testing.T) {
	rl := NewAuthRateLimiter()

	cfg := AuthRateLimitConfig{
		Name:   "test",
		Limit:  10,
		Window: time.Minute,
	}

	if rl.Remaining("key-1", cfg) != 10 {
		t.Error("should have 10 remaining initially")
	}

	rl.Check("key-1", cfg)
	rl.Check("key-1", cfg)

	if rl.Remaining("key-1", cfg) != 8 {
		t.Errorf("expected 8 remaining, got %d", rl.Remaining("key-1", cfg))
	}
}

func TestAuthRateLimiter_Reset(t *testing.T) {
	rl := NewAuthRateLimiter()

	cfg := AuthRateLimitConfig{
		Name:   "test",
		Limit:  2,
		Window: time.Minute,
	}

	rl.Check("key-1", cfg)
	rl.Check("key-1", cfg)

	rl.Reset("key-1")

	// Should be allowed again
	if err := rl.Check("key-1", cfg); err != nil {
		t.Error("should be allowed after reset")
	}
}

func TestAuthRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewAuthRateLimiter()

	cfg := AuthRateLimitConfig{
		Name:   "test",
		Limit:  2,
		Window: 10 * time.Millisecond,
	}

	rl.Check("key-1", cfg)
	rl.Check("key-1", cfg)

	// Wait for window to expire
	time.Sleep(15 * time.Millisecond)

	if err := rl.Check("key-1", cfg); err != nil {
		t.Error("should be allowed after window expiry")
	}
}

func TestFormatRateLimitKey(t *testing.T) {
	key := FormatRateLimitKey("otp", "abc123")
	if key != "rl:otp:abc123" {
		t.Errorf("expected rl:otp:abc123, got %s", key)
	}
}

func TestDefaultRateLimitConfigs(t *testing.T) {
	configs := []AuthRateLimitConfig{
		RateLimitOTPSend, RateLimitOTPVerify, RateLimitLoginIP,
		RateLimitLoginAccount, RateLimitRefresh, RateLimitStepUp,
		RateLimitRecovery, RateLimitAPIUnverified, RateLimitAPIVerified,
	}

	for _, cfg := range configs {
		if cfg.Name == "" {
			t.Error("config should have a name")
		}
		if cfg.Limit <= 0 {
			t.Errorf("config %s should have positive limit", cfg.Name)
		}
		if cfg.Window <= 0 {
			t.Errorf("config %s should have positive window", cfg.Name)
		}
	}

	// Verified users should have higher limits than unverified
	if RateLimitAPIVerified.Limit <= RateLimitAPIUnverified.Limit {
		t.Error("verified API limit should be higher than unverified")
	}
}
