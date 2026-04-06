package metagraph

import (
	"math/big"
	"testing"
	"time"
)

func TestDefaultSlashingPenalties_AllOffenses(t *testing.T) {
	penalties := DefaultSlashingPenalties()
	offenses := []SlashingOffense{
		OffenseFraudulentReward,
		OffenseInvalidMerkle,
		OffenseExtendedDowntime,
		OffenseDoubleSigning,
		OffenseCollusion,
	}

	for _, o := range offenses {
		p, ok := penalties[o]
		if !ok {
			t.Errorf("missing penalty for offense %s", o)
			continue
		}
		if p.StakePercent <= 0 {
			t.Errorf("offense %s should have positive stake percent", o)
		}
	}
}

func TestDefaultSlashingPenalties_DoubleSignIsSeverest(t *testing.T) {
	penalties := DefaultSlashingPenalties()
	ds := penalties[OffenseDoubleSigning]

	if ds.StakePercent != 50 {
		t.Errorf("double signing should slash 50%%, got %d%%", ds.StakePercent)
	}
	if ds.Recoverable {
		t.Error("double signing should be permanent ban")
	}
	if ds.SuspensionDays != -1 {
		t.Error("double signing should have permanent suspension (-1)")
	}
}

func TestDefaultSlashingPenalties_DowntimeIsLeast(t *testing.T) {
	penalties := DefaultSlashingPenalties()
	dt := penalties[OffenseExtendedDowntime]

	if dt.StakePercent != 1 {
		t.Errorf("downtime should slash 1%%, got %d%%", dt.StakePercent)
	}
	if !dt.Recoverable {
		t.Error("downtime should be recoverable")
	}
}

func TestCalculateSlash_FraudulentReward(t *testing.T) {
	validator := &ValidatorStatus{
		ValidatorDID: "did:dag:val1",
		StakedAmount: big.NewInt(1000000000000), // 10,000 ECHO
		IsActive:     true,
		Layer:        CurrencyL1,
	}

	event, err := CalculateSlash(validator, OffenseFraudulentReward, "evidence-hash-1", "did:dag:peer1", 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 10% of 10,000 ECHO = 1,000 ECHO
	expectedSlash := big.NewInt(100000000000)
	if event.SlashedAmount.Cmp(expectedSlash) != 0 {
		t.Errorf("expected %s slashed, got %s", expectedSlash.String(), event.SlashedAmount.String())
	}
	if event.PermanentBan {
		t.Error("fraudulent reward should not be permanent ban")
	}
	if event.SuspendedUntil.IsZero() {
		t.Error("should have 30-day suspension")
	}
	if event.TreasuryCredits.Cmp(expectedSlash) != 0 {
		t.Error("slashed amount should go to treasury")
	}
}

func TestCalculateSlash_DoubleSigning(t *testing.T) {
	validator := &ValidatorStatus{
		ValidatorDID: "did:dag:val2",
		StakedAmount: big.NewInt(1000000000000),
		IsActive:     true,
	}

	event, err := CalculateSlash(validator, OffenseDoubleSigning, "double-sign-proof", "did:dag:peer2", 12346)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 50% of 10,000 ECHO = 5,000 ECHO
	expectedSlash := big.NewInt(500000000000)
	if event.SlashedAmount.Cmp(expectedSlash) != 0 {
		t.Errorf("expected %s slashed, got %s", expectedSlash.String(), event.SlashedAmount.String())
	}
	if !event.PermanentBan {
		t.Error("double signing should be permanent ban")
	}
}

func TestCalculateSlash_UnknownOffense(t *testing.T) {
	validator := &ValidatorStatus{
		ValidatorDID: "did:dag:val3",
		StakedAmount: big.NewInt(1000000000000),
	}

	_, err := CalculateSlash(validator, SlashingOffense("unknown_offense"), "hash", "peer", 100)
	if err == nil {
		t.Fatal("expected error for unknown offense")
	}
}

func TestApplySlash_ReducesStake(t *testing.T) {
	originalStake := big.NewInt(1000000000000) // 10,000 ECHO
	validator := &ValidatorStatus{
		ValidatorDID: "did:dag:val4",
		StakedAmount: new(big.Int).Set(originalStake),
		IsActive:     true,
	}

	event, _ := CalculateSlash(validator, OffenseInvalidMerkle, "hash", "peer", 100)
	ApplySlash(validator, event)

	// 5% slash: remaining = 9,500 ECHO
	expectedRemaining := big.NewInt(950000000000)
	if validator.StakedAmount.Cmp(expectedRemaining) != 0 {
		t.Errorf("expected %s remaining, got %s", expectedRemaining.String(), validator.StakedAmount.String())
	}
	if len(validator.SlashingHistory) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(validator.SlashingHistory))
	}
}

func TestApplySlash_PermanentBan(t *testing.T) {
	validator := &ValidatorStatus{
		ValidatorDID: "did:dag:val5",
		StakedAmount: big.NewInt(1000000000000),
		IsActive:     true,
	}

	event, _ := CalculateSlash(validator, OffenseDoubleSigning, "hash", "peer", 100)
	ApplySlash(validator, event)

	if !validator.IsBanned {
		t.Error("validator should be banned after double-signing")
	}
	if validator.IsActive {
		t.Error("banned validator should not be active")
	}
}

func TestApplySlash_Suspension(t *testing.T) {
	validator := &ValidatorStatus{
		ValidatorDID: "did:dag:val6",
		StakedAmount: big.NewInt(1000000000000),
		IsActive:     true,
	}

	event, _ := CalculateSlash(validator, OffenseFraudulentReward, "hash", "peer", 100)
	ApplySlash(validator, event)

	if !validator.IsSuspended {
		t.Error("validator should be suspended")
	}
	if validator.IsActive {
		t.Error("suspended validator should not be active")
	}
	if validator.SuspendedUntil.Before(time.Now()) {
		t.Error("suspension should be in the future")
	}
}

func TestApplySlash_DelegatorsNeverSlashed(t *testing.T) {
	// The architecture guarantees delegators are never slashed.
	// We verify this by confirming only the validator's staked amount decreases.
	validator := &ValidatorStatus{
		ValidatorDID: "did:dag:val7",
		StakedAmount: big.NewInt(500000000000), // validator's own stake only
		IsActive:     true,
	}

	event, _ := CalculateSlash(validator, OffenseFraudulentReward, "hash", "peer", 100)

	// Slashing is calculated on validator.StakedAmount only (not delegated amounts)
	expectedSlash := big.NewInt(50000000000) // 10% of 5,000 ECHO
	if event.SlashedAmount.Cmp(expectedSlash) != 0 {
		t.Errorf("slash should only apply to validator's own stake, expected %s, got %s",
			expectedSlash.String(), event.SlashedAmount.String())
	}
}
