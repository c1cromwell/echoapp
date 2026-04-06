package rewards

import (
	"context"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	internalrewards "github.com/thechadcromwell/echoapp/internal/rewards"
)

func newTestService() *Service {
	db := database.NewMemoryDB()
	emission := internalrewards.NewEmissionSchedule(time.Now().AddDate(-1, 0, 0))
	return NewService(db, emission)
}

func TestClaim_MessageReward(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	req := ClaimRequest{
		DID:          "did:alice",
		RewardType:   "messaging",
		TrustTier:    3,
		MessageCount: 10,
	}

	result, err := svc.Claim(ctx, req)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if result.Status != "pending_l1" {
		t.Errorf("expected pending_l1, got %s", result.Status)
	}
	if result.Amount <= 0 {
		t.Errorf("expected positive reward amount, got %d", result.Amount)
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0x multiplier for tier 3, got %f", result.Multiplier)
	}
}

func TestClaim_ReferralReward(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	req := ClaimRequest{
		DID:        "did:alice",
		RewardType: "referral",
		TrustTier:  4,
	}

	result, err := svc.Claim(ctx, req)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	expected := int64(75_00000000) // 50 * 1.5
	if result.Amount != expected {
		t.Errorf("expected %d, got %d", expected, result.Amount)
	}
}

func TestClaim_Tier1Rejected(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	req := ClaimRequest{
		DID:          "did:alice",
		RewardType:   "messaging",
		TrustTier:    1,
		MessageCount: 5,
	}

	_, err := svc.Claim(ctx, req)
	if err != ErrInsufficientTier {
		t.Errorf("expected ErrInsufficientTier, got %v", err)
	}
}

func TestClaim_Tier2HalfMultiplier(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	req := ClaimRequest{
		DID:          "did:alice",
		RewardType:   "messaging",
		TrustTier:    2,
		MessageCount: 1,
	}

	result, err := svc.Claim(ctx, req)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if result.Multiplier != 0.5 {
		t.Errorf("expected 0.5x multiplier for tier 2, got %f", result.Multiplier)
	}
}

func TestClaim_Tier5DoubleMultiplier(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	req := ClaimRequest{
		DID:          "did:alice",
		RewardType:   "messaging",
		TrustTier:    5,
		MessageCount: 1,
	}

	result, err := svc.Claim(ctx, req)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if result.Multiplier != 2.0 {
		t.Errorf("expected 2.0x multiplier for tier 5, got %f", result.Multiplier)
	}
}

func TestClaim_AntiGaming_Velocity(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	// Make 10 claims (max per hour)
	for i := 0; i < 10; i++ {
		req := ClaimRequest{
			DID:          "did:alice",
			RewardType:   "messaging",
			TrustTier:    3,
			MessageCount: i + 1,
		}
		_, err := svc.Claim(ctx, req)
		if err != nil {
			t.Fatalf("Claim %d: %v", i, err)
		}
	}

	// 11th should fail
	req := ClaimRequest{
		DID:          "did:alice",
		RewardType:   "messaging",
		TrustTier:    3,
		MessageCount: 11,
	}
	_, err := svc.Claim(ctx, req)
	if err == nil {
		t.Errorf("expected velocity error on 11th claim")
	}
}

func TestClaim_UnknownRewardType(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	req := ClaimRequest{
		DID:        "did:alice",
		RewardType: "unknown",
		TrustTier:  3,
	}

	_, err := svc.Claim(ctx, req)
	if err != ErrNoPendingRewards {
		t.Errorf("expected ErrNoPendingRewards for unknown type, got %v", err)
	}
}

func TestGetPending(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	pending, err := svc.GetPending(ctx, "did:alice", 3)
	if err != nil {
		t.Fatalf("GetPending: %v", err)
	}
	if pending.TrustTier != 3 {
		t.Errorf("expected tier 3, got %d", pending.TrustTier)
	}
	if pending.Multiplier != 1.0 {
		t.Errorf("expected 1.0x multiplier, got %f", pending.Multiplier)
	}
	if pending.TotalPending <= 0 {
		t.Errorf("expected positive pending amount")
	}
}

func TestGetDailyStats(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	stats, err := svc.GetDailyStats(ctx)
	if err != nil {
		t.Fatalf("GetDailyStats: %v", err)
	}
	if stats.DailyBudget <= 0 {
		t.Errorf("expected positive daily budget")
	}
	if stats.Date == "" {
		t.Errorf("expected non-empty date")
	}
}

func TestAutoScaleRate(t *testing.T) {
	svc := newTestService()

	// No messages: rate should be at base
	rate0 := svc.AutoScaleRate()
	if rate0 != BaseRatePerMessage {
		t.Errorf("expected base rate %d, got %d", BaseRatePerMessage, rate0)
	}

	// Record many messages to trigger decay
	for i := 0; i < 200; i++ {
		svc.RecordMessage()
	}

	rate200 := svc.AutoScaleRate()
	if rate200 >= rate0 {
		t.Errorf("expected decay: rate at 200 msgs (%d) should be less than base (%d)", rate200, rate0)
	}
	if rate200 <= 0 {
		t.Errorf("rate should never reach zero (min floor), got %d", rate200)
	}
}

func TestAutoScaleRate_MinFloor(t *testing.T) {
	svc := newTestService()

	svc.mu.Lock()
	svc.todayMessages = 100000
	svc.mu.Unlock()

	rate := svc.AutoScaleRate()
	minRate := int64(float64(BaseRatePerMessage) * MinRateMultiplier)
	if rate < minRate {
		t.Errorf("rate %d should not go below min floor %d", rate, minRate)
	}
}

func TestTrustMultipliers(t *testing.T) {
	expected := map[int]float64{
		1: 0.0,
		2: 0.5,
		3: 1.0,
		4: 1.5,
		5: 2.0,
	}
	for tier, mult := range expected {
		if TrustMultiplier[tier] != mult {
			t.Errorf("tier %d: expected multiplier %f, got %f", tier, mult, TrustMultiplier[tier])
		}
	}
}

func TestDailyReset(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	req := ClaimRequest{DID: "did:alice", RewardType: "messaging", TrustTier: 3, MessageCount: 1}
	svc.Claim(ctx, req)

	stats, _ := svc.GetDailyStats(ctx)
	if stats.TotalDistributed == 0 {
		t.Errorf("expected non-zero distributed after claim")
	}

	// Simulate next day
	svc.mu.Lock()
	svc.lastResetDate = "2000-01-01"
	svc.mu.Unlock()

	stats, _ = svc.GetDailyStats(ctx)
	if stats.TotalDistributed != 0 {
		t.Errorf("expected 0 after daily reset, got %d", stats.TotalDistributed)
	}
}
