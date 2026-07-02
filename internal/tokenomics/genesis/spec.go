// Package genesis defines the WO-214 fixed-supply genesis allocation.
package genesis

const DatumPerECHO int64 = 100_000_000

// Pool names for the five genesis allocation buckets.
const (
	PoolCommunityRewards = "community_rewards"
	PoolTreasury         = "treasury"
	PoolFounders         = "founders"
	PoolFutureTeam       = "future_team"
	PoolEcosystem        = "ecosystem"
)

// Fixed 1B ECHO total supply (datum units).
const TotalSupply int64 = 1_000_000_000 * DatumPerECHO

// Pool allocations (datum).
const (
	CommunityRewardsAmount int64 = 400_000_000 * DatumPerECHO
	TreasuryAmount         int64 = 220_000_000 * DatumPerECHO
	FoundersAmount         int64 = 180_000_000 * DatumPerECHO
	FutureTeamAmount       int64 = 100_000_000 * DatumPerECHO
	EcosystemAmount        int64 = 100_000_000 * DatumPerECHO
)

// Treasury sub-allocations at genesis (datum).
const (
	TreasuryPacaSwapSeed    int64 = 80_000_000 * DatumPerECHO
	TreasuryOperational     int64 = 50_000_000 * DatumPerECHO
	TreasuryMultisigReserve int64 = 90_000_000 * DatumPerECHO
)

// Founder vesting schedule (Currency L1 enforced).
const (
	FounderCliffMonths = 12
	FounderVestMonths  = 48
)

// Founder lock amounts (datum).
const (
	FounderCEOAmount    int64 = 100_000_000 * DatumPerECHO
	FounderMemberAmount int64 = 20_000_000 * DatumPerECHO
)

// Yearly community emission caps (datum) — years 1–10.
var YearlyEmissionCaps = []int64{
	80_000_000 * DatumPerECHO,
	64_000_000 * DatumPerECHO,
	52_000_000 * DatumPerECHO,
	44_000_000 * DatumPerECHO,
	36_000_000 * DatumPerECHO,
	28_000_000 * DatumPerECHO,
	24_000_000 * DatumPerECHO,
	24_000_000 * DatumPerECHO,
	24_000_000 * DatumPerECHO,
	24_000_000 * DatumPerECHO,
}
