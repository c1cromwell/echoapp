package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/infra"
)

// TestSMSRecovery_OTPLockout verifies S6: after maxOTPAttempts wrong codes the
// session is locked (429) for its lifetime — even a subsequently-correct code
// is refused.
func TestSMSRecovery_OTPLockout(t *testing.T) {
	rt := &Router{AllowedOrigins: []string{"*"}} // no Redis → in-memory sessions
	token := "sess-lockout"
	if err := rt.putSMSSession(context.Background(), token, smsOTPSession{
		OTP:       "123456",
		DID:       "did:key:zX",
		ExpiresAt: time.Now().Add(time.Minute),
	}); err != nil {
		t.Fatal(err)
	}

	post := func(otp string) int {
		body, _ := json.Marshal(map[string]string{"session_token": token, "otp": otp})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/sms-recovery/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		rt.Handler().ServeHTTP(rec, req)
		return rec.Code
	}

	if c := post("000000"); c != http.StatusUnauthorized {
		t.Fatalf("attempt 1 want 401, got %d", c)
	}
	if c := post("000000"); c != http.StatusUnauthorized {
		t.Fatalf("attempt 2 want 401, got %d", c)
	}
	if c := post("000000"); c != http.StatusTooManyRequests {
		t.Fatalf("attempt 3 (lockout) want 429, got %d", c)
	}
	if c := post("123456"); c != http.StatusTooManyRequests {
		t.Fatalf("correct code after lockout must still be 429, got %d", c)
	}
}

// TestSMSRecovery_PerPhoneSendCap verifies S9: a phone number may request only a
// bounded number of codes per window.
func TestSMSRecovery_PerPhoneSendCap(t *testing.T) {
	rt := &Router{
		PublicRateLimiter: infra.NewRateLimiter(map[string]infra.RateLimitConfig{
			otpSendAction: {MaxRequests: 2, Window: time.Minute},
		}),
	}
	ctx := context.Background()
	phone := "+14155550000"
	ph := phoneHash(phone)

	for i := 0; i < 2; i++ {
		if _, _, err := rt.dispatchOTP(ctx, phone, ph, "did:key:z"); err != nil {
			t.Fatalf("send %d should succeed, got %v", i+1, err)
		}
	}
	if _, _, err := rt.dispatchOTP(ctx, phone, ph, "did:key:z"); !errors.Is(err, errOTPSendRateLimited) {
		t.Fatalf("3rd send should be rate limited, got %v", err)
	}
}
