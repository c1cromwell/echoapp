package infra

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(map[string]RateLimitConfig{
		"message_send": {MaxRequests: 60, Window: time.Minute},
	})

	for i := 0; i < 59; i++ {
		if err := rl.Check("did:dag:user1", "message_send"); err != nil {
			t.Fatalf("request %d should be allowed, got: %v", i, err)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(map[string]RateLimitConfig{
		"message_send": {MaxRequests: 5, Window: time.Minute},
	})

	for i := 0; i < 5; i++ {
		rl.Check("did:dag:user1", "message_send")
	}

	err := rl.Check("did:dag:user1", "message_send")
	if err != ErrRateLimitExceeded {
		t.Errorf("expected rate limit exceeded, got: %v", err)
	}
}

func TestRateLimiter_SeparateUsers(t *testing.T) {
	rl := NewRateLimiter(map[string]RateLimitConfig{
		"message_send": {MaxRequests: 3, Window: time.Minute},
	})

	// User 1 exhausts their limit
	for i := 0; i < 3; i++ {
		rl.Check("did:dag:user1", "message_send")
	}

	// User 2 should still be allowed
	if err := rl.Check("did:dag:user2", "message_send"); err != nil {
		t.Error("user2 should not be rate limited by user1's usage")
	}
}

func TestRateLimiter_SeparateActions(t *testing.T) {
	rl := NewRateLimiter(map[string]RateLimitConfig{
		"message_send": {MaxRequests: 3, Window: time.Minute},
		"api_request":  {MaxRequests: 100, Window: time.Minute},
	})

	// Exhaust message_send
	for i := 0; i < 3; i++ {
		rl.Check("did:dag:user1", "message_send")
	}

	// api_request should still work
	if err := rl.Check("did:dag:user1", "api_request"); err != nil {
		t.Error("different action should not be rate limited")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(map[string]RateLimitConfig{
		"test_action": {MaxRequests: 2, Window: 10 * time.Millisecond},
	})

	rl.Check("did:dag:user1", "test_action")
	rl.Check("did:dag:user1", "test_action")

	err := rl.Check("did:dag:user1", "test_action")
	if err != ErrRateLimitExceeded {
		t.Fatal("should be rate limited")
	}

	// Wait for window to expire
	time.Sleep(15 * time.Millisecond)

	if err := rl.Check("did:dag:user1", "test_action"); err != nil {
		t.Error("should be allowed after window reset")
	}
}

func TestRateLimiter_UnknownActionAllows(t *testing.T) {
	rl := NewRateLimiter(map[string]RateLimitConfig{})
	if err := rl.Check("did:dag:user1", "unknown_action"); err != nil {
		t.Error("unknown action should be allowed")
	}
}

func TestRateLimiter_Remaining(t *testing.T) {
	rl := NewRateLimiter(map[string]RateLimitConfig{
		"message_send": {MaxRequests: 10, Window: time.Minute},
	})

	remaining := rl.Remaining("did:dag:user1", "message_send")
	if remaining != 10 {
		t.Errorf("expected 10 remaining, got %d", remaining)
	}

	rl.Check("did:dag:user1", "message_send")
	rl.Check("did:dag:user1", "message_send")

	remaining = rl.Remaining("did:dag:user1", "message_send")
	if remaining != 8 {
		t.Errorf("expected 8 remaining, got %d", remaining)
	}
}

func TestDefaultRateLimits(t *testing.T) {
	limits := DefaultRateLimits()
	expected := []string{"api_request", "message_send", "reward_claim", "data_submission", "websocket_msg"}
	for _, action := range expected {
		if _, ok := limits[action]; !ok {
			t.Errorf("missing default limit for %s", action)
		}
	}

	// Verify specific values from architecture spec
	if limits["api_request"].MaxRequests != 100 {
		t.Errorf("api_request should be 100/min")
	}
	if limits["message_send"].MaxRequests != 60 {
		t.Errorf("message_send should be 60/min")
	}
}
