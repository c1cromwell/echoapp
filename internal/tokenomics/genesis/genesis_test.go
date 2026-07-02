package genesis

import (
	"testing"
	"time"
)

func TestBuildSnapshot_PoolTotals(t *testing.T) {
	snap, err := BuildSnapshot(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}
	if snap.TotalSupply != TotalSupply {
		t.Errorf("total supply: got %d want %d", snap.TotalSupply, TotalSupply)
	}
	var sum int64
	for _, p := range snap.Pools {
		sum += p.Amount
	}
	if sum != TotalSupply {
		t.Errorf("pool sum %d != %d", sum, TotalSupply)
	}
	if snap.Treasury.PacaSwapSeed+snap.Treasury.OperationalReserve+snap.Treasury.MultisigReserve != TreasuryAmount {
		t.Error("treasury sub-allocations do not sum to treasury pool")
	}
}

func TestBuildSnapshot_FounderAmounts(t *testing.T) {
	snap, err := BuildSnapshot(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}
	if len(snap.Founders) != 5 {
		t.Fatalf("expected 5 founders, got %d", len(snap.Founders))
	}
	if snap.Founders[0].Amount != FounderCEOAmount {
		t.Errorf("CEO amount: got %d want %d", snap.Founders[0].Amount, FounderCEOAmount)
	}
	for i := 1; i < 5; i++ {
		if snap.Founders[i].Amount != FounderMemberAmount {
			t.Errorf("founder %d amount: got %d want %d", i+1, snap.Founders[i].Amount, FounderMemberAmount)
		}
	}
}

func TestYearlyEmissionCaps_SumToPool(t *testing.T) {
	var sum int64
	for _, cap := range YearlyEmissionCaps {
		sum += cap
	}
	if sum != CommunityRewardsAmount {
		t.Errorf("yearly caps sum %d != community pool %d", sum, CommunityRewardsAmount)
	}
}

func TestIsFounderDID(t *testing.T) {
	dids := DefaultFounderDIDs()
	if !IsFounderDID(dids[0]) {
		t.Error("expected first founder DID to match")
	}
	if IsFounderDID("did:key:z6MkNotFounder") {
		t.Error("unexpected founder match")
	}
}
