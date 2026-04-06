package rewards

import (
	"testing"
	"time"
)

func TestAntiGaming_VelocityCheck(t *testing.T) {
	detector := NewAntiGamingDetector()
	now := time.Now()

	// Submit 10 claims — all should succeed
	for i := 0; i < MaxClaimsPerHour; i++ {
		err := detector.CheckAndRecord(ClaimEvent{
			DID:        "did:dag:user1",
			RewardType: "messaging",
			Amount:     int64(100 + i), // Different amounts to avoid duplicate detection
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
		})
		if err != nil {
			t.Fatalf("claim %d should succeed, got: %v", i, err)
		}
	}

	// 11th claim should fail
	err := detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "messaging",
		Amount:     200,
		Timestamp:  now.Add(11 * time.Minute),
	})
	if err != ErrVelocityExceeded {
		t.Errorf("expected ErrVelocityExceeded, got: %v", err)
	}
}

func TestAntiGaming_DuplicateDetection(t *testing.T) {
	detector := NewAntiGamingDetector()
	now := time.Now()

	// First claim
	err := detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "referral",
		Amount:     500,
		Timestamp:  now,
	})
	if err != nil {
		t.Fatalf("first claim should succeed: %v", err)
	}

	// Same type + amount within 60 seconds → duplicate
	err = detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "referral",
		Amount:     500,
		Timestamp:  now.Add(30 * time.Second),
	})
	if err != ErrDuplicateClaim {
		t.Errorf("expected ErrDuplicateClaim, got: %v", err)
	}
}

func TestAntiGaming_DuplicateAllowedAfterWindow(t *testing.T) {
	detector := NewAntiGamingDetector()
	now := time.Now()

	err := detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "referral",
		Amount:     500,
		Timestamp:  now,
	})
	if err != nil {
		t.Fatalf("first claim should succeed: %v", err)
	}

	// Same type + amount but 61 seconds later → allowed
	err = detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "referral",
		Amount:     500,
		Timestamp:  now.Add(61 * time.Second),
	})
	if err != nil {
		t.Errorf("claim after 61s should succeed, got: %v", err)
	}
}

func TestAntiGaming_DifferentTypeNotDuplicate(t *testing.T) {
	detector := NewAntiGamingDetector()
	now := time.Now()

	_ = detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "messaging",
		Amount:     500,
		Timestamp:  now,
	})

	// Different type, same amount, within window → NOT a duplicate
	err := detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "staking",
		Amount:     500,
		Timestamp:  now.Add(10 * time.Second),
	})
	if err != nil {
		t.Errorf("different type should not trigger duplicate: %v", err)
	}
}

func TestAntiGaming_DifferentAmountNotDuplicate(t *testing.T) {
	detector := NewAntiGamingDetector()
	now := time.Now()

	_ = detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "messaging",
		Amount:     500,
		Timestamp:  now,
	})

	// Same type, different amount, within window → NOT a duplicate
	err := detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "messaging",
		Amount:     600,
		Timestamp:  now.Add(10 * time.Second),
	})
	if err != nil {
		t.Errorf("different amount should not trigger duplicate: %v", err)
	}
}

func TestAntiGaming_VelocityResetsAfterWindow(t *testing.T) {
	detector := NewAntiGamingDetector()
	baseTime := time.Now().Add(-2 * time.Hour) // 2 hours ago

	// Fill up with 10 old claims
	for i := 0; i < MaxClaimsPerHour; i++ {
		_ = detector.CheckAndRecord(ClaimEvent{
			DID:        "did:dag:user1",
			RewardType: "messaging",
			Amount:     int64(100 + i),
			Timestamp:  baseTime.Add(time.Duration(i) * time.Minute),
		})
	}

	// New claim 2 hours later — old claims should be pruned
	err := detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "messaging",
		Amount:     999,
		Timestamp:  time.Now(),
	})
	if err != nil {
		t.Errorf("claim after velocity window should succeed: %v", err)
	}
}

func TestAntiGaming_DifferentUsersIndependent(t *testing.T) {
	detector := NewAntiGamingDetector()
	now := time.Now()

	// Fill up user1
	for i := 0; i < MaxClaimsPerHour; i++ {
		_ = detector.CheckAndRecord(ClaimEvent{
			DID:        "did:dag:user1",
			RewardType: "messaging",
			Amount:     int64(100 + i),
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
		})
	}

	// user2 should still be able to claim
	err := detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user2",
		RewardType: "messaging",
		Amount:     100,
		Timestamp:  now,
	})
	if err != nil {
		t.Errorf("different user should not be rate limited: %v", err)
	}
}

func TestAntiGaming_RecentClaimCount(t *testing.T) {
	detector := NewAntiGamingDetector()
	now := time.Now()

	for i := 0; i < 5; i++ {
		_ = detector.CheckAndRecord(ClaimEvent{
			DID:        "did:dag:user1",
			RewardType: "messaging",
			Amount:     int64(100 + i),
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
		})
	}

	count := detector.RecentClaimCount("did:dag:user1")
	if count != 5 {
		t.Errorf("expected 5 recent claims, got %d", count)
	}

	count = detector.RecentClaimCount("did:dag:user2")
	if count != 0 {
		t.Errorf("expected 0 for unknown user, got %d", count)
	}
}

func TestAntiGaming_Reset(t *testing.T) {
	detector := NewAntiGamingDetector()

	_ = detector.CheckAndRecord(ClaimEvent{
		DID:        "did:dag:user1",
		RewardType: "messaging",
		Amount:     100,
		Timestamp:  time.Now(),
	})

	detector.Reset()

	count := detector.RecentClaimCount("did:dag:user1")
	if count != 0 {
		t.Errorf("expected 0 after reset, got %d", count)
	}
}
