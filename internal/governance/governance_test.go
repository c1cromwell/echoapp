package governance

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// --- Mock implementations ---

type mockTrust struct {
	tiers map[string]int
}

func (m *mockTrust) GetTrustTier(_ context.Context, did string) (int, error) {
	tier, ok := m.tiers[did]
	if !ok {
		return 0, fmt.Errorf("unknown did: %s", did)
	}
	return tier, nil
}

type mockStake struct {
	stakes map[string]int64
}

func (m *mockStake) GetTotalStaked(_ context.Context, did string) (int64, error) {
	return m.stakes[did], nil
}

type mockProposalStore struct {
	proposals map[string]*Proposal
	votes     map[string]map[string]*VoteRecord // proposalID -> did -> record
}

func newMockStore() *mockProposalStore {
	return &mockProposalStore{
		proposals: make(map[string]*Proposal),
		votes:     make(map[string]map[string]*VoteRecord),
	}
}

func (m *mockProposalStore) CreateProposal(_ context.Context, p *Proposal) error {
	m.proposals[p.ID] = p
	return nil
}

func (m *mockProposalStore) GetProposal(_ context.Context, id string) (*Proposal, error) {
	p, ok := m.proposals[id]
	if !ok {
		return nil, ErrProposalNotFound
	}
	return p, nil
}

func (m *mockProposalStore) ListActiveProposals(_ context.Context) ([]Proposal, error) {
	var active []Proposal
	for _, p := range m.proposals {
		if p.Status == StatusActive {
			active = append(active, *p)
		}
	}
	return active, nil
}

func (m *mockProposalStore) RecordVote(_ context.Context, v *VoteRecord) error {
	if m.votes[v.ProposalID] == nil {
		m.votes[v.ProposalID] = make(map[string]*VoteRecord)
	}
	m.votes[v.ProposalID][v.DID] = v

	// Update cached tally on proposal
	p := m.proposals[v.ProposalID]
	if p != nil {
		switch v.Value {
		case VoteFor:
			p.Tally = updateTally(p, v.Weight, 0, 0)
		case VoteAgainst:
			p.Tally = updateTally(p, 0, v.Weight, 0)
		case VoteAbstain:
			p.Tally = updateTally(p, 0, 0, v.Weight)
		}
	}
	return nil
}

func updateTally(p *Proposal, forW, againstW, abstainW int64) *ProposalTally {
	tally := p.Tally
	if tally == nil {
		tally = &ProposalTally{ProposalID: p.ID}
	}
	tally.ForWeight += forW
	tally.AgainstWeight += againstW
	tally.AbstainWeight += abstainW
	tally.TotalWeight = tally.ForWeight + tally.AgainstWeight + tally.AbstainWeight
	tally.VoterCount++
	if tally.TotalWeight > 0 {
		tally.ForPercent = float64(tally.ForWeight*100) / float64(tally.TotalWeight)
	}
	tally.Passed = CheckThresholdPassed(p.Threshold, tally.ForWeight, tally.AgainstWeight, tally.TotalWeight)
	return tally
}

func (m *mockProposalStore) HasVoted(_ context.Context, did, proposalID string) (bool, error) {
	if m.votes[proposalID] == nil {
		return false, nil
	}
	_, ok := m.votes[proposalID][did]
	return ok, nil
}

func (m *mockProposalStore) GetTally(_ context.Context, proposalID string) (*ProposalTally, error) {
	p, ok := m.proposals[proposalID]
	if !ok {
		return nil, ErrProposalNotFound
	}
	if p.Tally != nil {
		return p.Tally, nil
	}
	return &ProposalTally{ProposalID: proposalID}, nil
}

func (m *mockProposalStore) UpdateProposalStatus(_ context.Context, id, status string) error {
	p, ok := m.proposals[id]
	if !ok {
		return ErrProposalNotFound
	}
	p.Status = status
	return nil
}

// --- Weight Calculator Tests ---

func TestCalculateWeight_Tier5(t *testing.T) {
	// 100,000 staked × 2.0x = 200,000
	weight := CalculateWeight(100000, 5)
	if weight != 200000 {
		t.Errorf("expected 200000, got %d", weight)
	}
}

func TestCalculateWeight_Tier1_ZeroGovernance(t *testing.T) {
	weight := CalculateWeight(1000000, 1)
	if weight != 0 {
		t.Errorf("tier 1 should have zero governance power, got %d", weight)
	}
}

func TestCalculateWeight_Tier3_Standard(t *testing.T) {
	// 50,000 staked × 1.0x = 50,000
	weight := CalculateWeight(50000, 3)
	if weight != 50000 {
		t.Errorf("expected 50000, got %d", weight)
	}
}

func TestCalculateWeight_Tier2_Half(t *testing.T) {
	// 10,000 staked × 0.5x = 5,000
	weight := CalculateWeight(10000, 2)
	if weight != 5000 {
		t.Errorf("expected 5000, got %d", weight)
	}
}

func TestCalculateWeight_Tier4(t *testing.T) {
	// 20,000 staked × 1.5x = 30,000
	weight := CalculateWeight(20000, 4)
	if weight != 30000 {
		t.Errorf("expected 30000, got %d", weight)
	}
}

func TestCalculateWeight_InvalidTier(t *testing.T) {
	weight := CalculateWeight(100000, 6)
	if weight != 0 {
		t.Errorf("invalid tier should return 0, got %d", weight)
	}
}

func TestCanVote_Tier2WithStake(t *testing.T) {
	if !CanVote(2, 1000) {
		t.Error("tier 2 with stake should be able to vote")
	}
}

func TestCanVote_Tier1Rejected(t *testing.T) {
	if CanVote(1, 1000000) {
		t.Error("tier 1 should not be able to vote regardless of stake")
	}
}

func TestCanVote_NoStakeRejected(t *testing.T) {
	if CanVote(5, 0) {
		t.Error("zero stake should not be able to vote regardless of tier")
	}
}

func TestTierMultiplierFloat(t *testing.T) {
	tests := []struct {
		tier     int
		expected float64
	}{
		{1, 0.0},
		{2, 0.5},
		{3, 1.0},
		{4, 1.5},
		{5, 2.0},
		{6, 0.0},
	}
	for _, tc := range tests {
		got := TierMultiplierFloat(tc.tier)
		if got != tc.expected {
			t.Errorf("tier %d: expected %f, got %f", tc.tier, tc.expected, got)
		}
	}
}

// --- Threshold Tests ---

func TestCheckThreshold_SimpleMajority(t *testing.T) {
	if !CheckThresholdPassed(ThresholdSimpleMajority, 5100, 4900, 10000) {
		t.Error("simple majority should pass when for > against")
	}
	if CheckThresholdPassed(ThresholdSimpleMajority, 4900, 5100, 10000) {
		t.Error("simple majority should fail when for < against")
	}
}

func TestCheckThreshold_Supermajority67(t *testing.T) {
	if !CheckThresholdPassed(ThresholdSupermajority67, 6700, 3300, 10000) {
		t.Error("67% threshold should pass at exactly 67%")
	}
	if CheckThresholdPassed(ThresholdSupermajority67, 6600, 3400, 10000) {
		t.Error("67% threshold should fail at 66%")
	}
}

func TestCheckThreshold_Supermajority75(t *testing.T) {
	if !CheckThresholdPassed(ThresholdSupermajority75, 7500, 2500, 10000) {
		t.Error("75% threshold should pass at exactly 75%")
	}
	if CheckThresholdPassed(ThresholdSupermajority75, 7400, 2600, 10000) {
		t.Error("75% threshold should fail at 74%")
	}
}

func TestCheckThreshold_ZeroTotal(t *testing.T) {
	if CheckThresholdPassed(ThresholdSimpleMajority, 0, 0, 0) {
		t.Error("zero total should not pass")
	}
}

// --- Validation Tests ---

func TestValidateVoteValue(t *testing.T) {
	if !ValidateVoteValue("for") {
		t.Error("'for' should be valid")
	}
	if !ValidateVoteValue("against") {
		t.Error("'against' should be valid")
	}
	if !ValidateVoteValue("abstain") {
		t.Error("'abstain' should be valid")
	}
	if ValidateVoteValue("maybe") {
		t.Error("'maybe' should be invalid")
	}
}

func TestValidateProposalType(t *testing.T) {
	if !ValidateProposalType("protocol_upgrade") {
		t.Error("protocol_upgrade should be valid")
	}
	if ValidateProposalType("random") {
		t.Error("random should be invalid")
	}
}

func TestValidateThreshold(t *testing.T) {
	if !ValidateThreshold("simple_majority") {
		t.Error("simple_majority should be valid")
	}
	if ValidateThreshold("unanimous") {
		t.Error("unanimous should be invalid")
	}
}

// --- BuildTally Tests ---

func TestBuildTally(t *testing.T) {
	tally := BuildTally("prop1", 7000, 2000, 1000, 10, ThresholdSupermajority67)
	if tally.TotalWeight != 10000 {
		t.Errorf("expected total 10000, got %d", tally.TotalWeight)
	}
	if tally.ForPercent != 70.0 {
		t.Errorf("expected 70%%, got %f", tally.ForPercent)
	}
	if !tally.Passed {
		t.Error("70% should pass 67% threshold")
	}
	if tally.VoterCount != 10 {
		t.Errorf("expected 10 voters, got %d", tally.VoterCount)
	}
}

// --- Service Integration Tests (with mocks) ---

func newTestService() (*GovernanceService, *mockProposalStore) {
	trust := &mockTrust{tiers: map[string]int{
		"did:dag:voter1": 3,
		"did:dag:voter2": 5,
		"did:dag:tier1":  1,
		"did:dag:nostake": 4,
	}}
	stake := &mockStake{stakes: map[string]int64{
		"did:dag:voter1":  50000,
		"did:dag:voter2":  100000,
		"did:dag:tier1":   1000000,
		"did:dag:nostake":  0,
	}}
	store := newMockStore()
	svc := NewGovernanceService(nil, trust, stake, store) // nil metagraph for unit tests
	return svc, store
}

func TestGetVotingPower_Tier3(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	power, err := svc.GetVotingPower(ctx, "did:dag:voter1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if power.TrustTier != 3 {
		t.Errorf("expected tier 3, got %d", power.TrustTier)
	}
	if power.Weight != 50000 { // 50000 × 1.0x
		t.Errorf("expected weight 50000, got %d", power.Weight)
	}
	if power.Multiplier != 1.0 {
		t.Errorf("expected multiplier 1.0, got %f", power.Multiplier)
	}
	if !power.CanVote {
		t.Error("tier 3 with stake should be able to vote")
	}
}

func TestGetVotingPower_Tier5(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	power, err := svc.GetVotingPower(ctx, "did:dag:voter2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if power.Weight != 200000 { // 100000 × 2.0x
		t.Errorf("expected weight 200000, got %d", power.Weight)
	}
}

func TestGetVotingPower_Tier1CannotVote(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	power, err := svc.GetVotingPower(ctx, "did:dag:tier1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if power.CanVote {
		t.Error("tier 1 should not be able to vote")
	}
	if power.Weight != 0 {
		t.Errorf("tier 1 weight should be 0, got %d", power.Weight)
	}
}

func TestGetVotingPower_NoStakeCannotVote(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	power, err := svc.GetVotingPower(ctx, "did:dag:nostake")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if power.CanVote {
		t.Error("zero stake should not be able to vote")
	}
}

func TestCreateProposal(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	proposal, err := svc.CreateProposal(ctx, CreateProposalRequest{
		Title:       "Increase staking rewards",
		Description: "Proposal to increase staking APR by 2%",
		Type:        ProposalTypeParameterChange,
		Threshold:   ThresholdSimpleMajority,
		CreatedBy:   "did:dag:voter1",
		EndsAt:      time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proposal.Status != StatusActive {
		t.Errorf("expected status active, got %s", proposal.Status)
	}
	if proposal.Title != "Increase staking rewards" {
		t.Errorf("unexpected title: %s", proposal.Title)
	}
}

func TestCreateProposal_InvalidType(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.CreateProposal(ctx, CreateProposalRequest{
		Title:     "Bad",
		Type:      "invalid",
		Threshold: ThresholdSimpleMajority,
		EndsAt:    time.Now().Add(time.Hour),
	})
	if err != ErrInvalidProposalType {
		t.Errorf("expected ErrInvalidProposalType, got %v", err)
	}
}

func TestCreateProposal_InvalidThreshold(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.CreateProposal(ctx, CreateProposalRequest{
		Title:     "Bad",
		Type:      ProposalTypeProtocolUpgrade,
		Threshold: "unanimous",
		EndsAt:    time.Now().Add(time.Hour),
	})
	if err != ErrInvalidThreshold {
		t.Errorf("expected ErrInvalidThreshold, got %v", err)
	}
}

func TestListActiveProposals(t *testing.T) {
	svc, store := newTestService()
	ctx := context.Background()

	// Create 2 active, 1 passed
	store.proposals["p1"] = &Proposal{ID: "p1", Status: StatusActive, EndsAt: time.Now().Add(time.Hour)}
	store.proposals["p2"] = &Proposal{ID: "p2", Status: StatusActive, EndsAt: time.Now().Add(time.Hour)}
	store.proposals["p3"] = &Proposal{ID: "p3", Status: StatusPassed}

	proposals, err := svc.ListActiveProposals(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(proposals) != 2 {
		t.Errorf("expected 2 active proposals, got %d", len(proposals))
	}
}

// --- FinalizeExpiredProposals Tests ---

func TestFinalizeExpiredProposals(t *testing.T) {
	store := newMockStore()
	ctx := context.Background()

	// One expired with enough for-votes
	store.proposals["p1"] = &Proposal{
		ID:        "p1",
		Status:    StatusActive,
		Threshold: ThresholdSimpleMajority,
		EndsAt:    time.Now().Add(-1 * time.Hour),
		Tally: &ProposalTally{
			ProposalID:    "p1",
			ForWeight:     6000,
			AgainstWeight: 4000,
			TotalWeight:   10000,
		},
	}
	// One expired that fails
	store.proposals["p2"] = &Proposal{
		ID:        "p2",
		Status:    StatusActive,
		Threshold: ThresholdSupermajority75,
		EndsAt:    time.Now().Add(-1 * time.Hour),
		Tally: &ProposalTally{
			ProposalID:    "p2",
			ForWeight:     5000,
			AgainstWeight: 5000,
			TotalWeight:   10000,
		},
	}
	// One still active (not expired)
	store.proposals["p3"] = &Proposal{
		ID:     "p3",
		Status: StatusActive,
		EndsAt: time.Now().Add(24 * time.Hour),
	}

	finalized, err := FinalizeExpiredProposals(ctx, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if finalized != 2 {
		t.Errorf("expected 2 finalized, got %d", finalized)
	}
	if store.proposals["p1"].Status != StatusPassed {
		t.Errorf("p1 should be passed, got %s", store.proposals["p1"].Status)
	}
	if store.proposals["p2"].Status != StatusFailed {
		t.Errorf("p2 should be failed, got %s", store.proposals["p2"].Status)
	}
	if store.proposals["p3"].Status != StatusActive {
		t.Errorf("p3 should still be active, got %s", store.proposals["p3"].Status)
	}
}

// --- Anti-Plutocratic Scenario Test ---

func TestAntiPlutocraticScenario(t *testing.T) {
	// A Tier 1 whale with 50M ECHO gets zero governance power.
	// 10,000 Tier 5 community members staking 10K each can match the CEO.
	whaleWeight := CalculateWeight(50_000_000, 1)
	if whaleWeight != 0 {
		t.Errorf("tier 1 whale should have zero weight, got %d", whaleWeight)
	}

	// CEO at tier 5 with 100M
	ceoWeight := CalculateWeight(100_000_000, 5)
	// 100M × 2.0 = 200M
	if ceoWeight != 200_000_000 {
		t.Errorf("CEO weight should be 200M, got %d", ceoWeight)
	}

	// 10,000 members at tier 5 with 10K each
	memberWeight := CalculateWeight(10_000, 5)
	totalCommunity := memberWeight * 10_000
	if totalCommunity != ceoWeight {
		t.Errorf("10K tier-5 members with 10K ECHO each (%d) should equal CEO weight (%d)", totalCommunity, ceoWeight)
	}
}
