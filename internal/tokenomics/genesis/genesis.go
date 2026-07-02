package genesis

import (
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// Pool describes a genesis allocation bucket.
type Pool struct {
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
}

// FounderPosition is a genesis TokenLock for a founder DID.
type FounderPosition struct {
	DID         string    `json:"did"`
	Role        string    `json:"role"`
	Amount      int64     `json:"amount"`
	CliffMonths int       `json:"cliffMonths"`
	VestMonths  int       `json:"vestMonths"`
	GenesisDate time.Time `json:"genesisDate"`
	LockedUntil time.Time `json:"lockedUntil"`
	VestingType string    `json:"vestingType"`
	ExplorerURL string    `json:"explorerUrl,omitempty"`
}

// Snapshot is the full genesis deployment manifest (WO-214).
type Snapshot struct {
	GenesisDate time.Time         `json:"genesisDate"`
	TotalSupply int64             `json:"totalSupply"`
	Pools       []Pool            `json:"pools"`
	Treasury    TreasuryGenesis   `json:"treasury"`
	Founders    []FounderPosition `json:"founders"`
	Emission    EmissionGenesis   `json:"emission"`
}

// TreasuryGenesis captures treasury sub-allocations at genesis.
type TreasuryGenesis struct {
	Total              int64 `json:"total"`
	PacaSwapSeed       int64 `json:"pacaSwapSeed"`
	OperationalReserve int64 `json:"operationalReserve"`
	MultisigReserve    int64 `json:"multisigReserve"`
}

// EmissionGenesis documents the community rewards curve.
type EmissionGenesis struct {
	PoolTotal       int64   `json:"poolTotal"`
	YearlyCaps      []int64 `json:"yearlyCaps"`
	NoMintAfterInit bool    `json:"noMintAfterGenesis"`
}

// DefaultFounderDIDs returns founder DIDs from ECHO_FOUNDER_DIDS (comma-separated).
// First DID is CEO (100M); next four are 20M each. Placeholders used when unset.
func DefaultFounderDIDs() []string {
	raw := strings.TrimSpace(os.Getenv("ECHO_FOUNDER_DIDS"))
	if raw == "" {
		return []string{
			"did:key:z6MkFounderCEO",
			"did:key:z6MkFounder2",
			"did:key:z6MkFounder3",
			"did:key:z6MkFounder4",
			"did:key:z6MkFounder5",
		}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// IsFounderDID reports whether did matches a configured genesis founder.
func IsFounderDID(did string) bool {
	for _, f := range DefaultFounderDIDs() {
		if f == did {
			return true
		}
	}
	return false
}

// BuildSnapshot constructs the WO-214 genesis manifest.
func BuildSnapshot(genesisDate time.Time) (*Snapshot, error) {
	dids := DefaultFounderDIDs()
	if len(dids) < 5 {
		return nil, fmt.Errorf("genesis: need 5 founder DIDs, got %d", len(dids))
	}

	founders := make([]FounderPosition, 0, 5)
	roles := []string{"CEO", "Founder 2", "Founder 3", "Founder 4", "Founder 5"}
	amounts := []int64{FounderCEOAmount, FounderMemberAmount, FounderMemberAmount, FounderMemberAmount, FounderMemberAmount}
	lockEnd := genesisDate.AddDate(0, FounderVestMonths, 0)

	for i := 0; i < 5; i++ {
		founders = append(founders, FounderPosition{
			DID:         dids[i],
			Role:        roles[i],
			Amount:      amounts[i],
			CliffMonths: FounderCliffMonths,
			VestMonths:  FounderVestMonths,
			GenesisDate: genesisDate,
			LockedUntil: lockEnd,
			VestingType: "founder",
			ExplorerURL: "https://dagexplorer.io/address/" + dids[i],
		})
	}

	var founderTotal int64
	for _, f := range founders {
		founderTotal += f.Amount
	}
	if founderTotal != FoundersAmount {
		return nil, fmt.Errorf("genesis: founder total %d != pool %d", founderTotal, FoundersAmount)
	}

	pools := []Pool{
		{Name: PoolCommunityRewards, Amount: CommunityRewardsAmount},
		{Name: PoolTreasury, Amount: TreasuryAmount},
		{Name: PoolFounders, Amount: FoundersAmount},
		{Name: PoolFutureTeam, Amount: FutureTeamAmount},
		{Name: PoolEcosystem, Amount: EcosystemAmount},
	}

	var poolSum int64
	for _, p := range pools {
		poolSum += p.Amount
	}
	if poolSum != TotalSupply {
		return nil, fmt.Errorf("genesis: pool sum %d != total supply %d", poolSum, TotalSupply)
	}

	return &Snapshot{
		GenesisDate: genesisDate,
		TotalSupply: TotalSupply,
		Pools:       pools,
		Treasury: TreasuryGenesis{
			Total:              TreasuryAmount,
			PacaSwapSeed:       TreasuryPacaSwapSeed,
			OperationalReserve: TreasuryOperational,
			MultisigReserve:    TreasuryMultisigReserve,
		},
		Founders: founders,
		Emission: EmissionGenesis{
			PoolTotal:       CommunityRewardsAmount,
			YearlyCaps:      append([]int64(nil), YearlyEmissionCaps...),
			NoMintAfterInit: true,
		},
	}, nil
}

// FounderTokenLockUpdates returns Currency L1 TokenLock payloads for genesis deploy.
func FounderTokenLockUpdates(genesisDate time.Time) ([]metagraph.CurrencyL1Transaction, error) {
	snap, err := BuildSnapshot(genesisDate)
	if err != nil {
		return nil, err
	}
	lockDays := FounderVestMonths * 30
	out := make([]metagraph.CurrencyL1Transaction, 0, len(snap.Founders))
	for _, f := range snap.Founders {
		out = append(out, metagraph.CurrencyL1Transaction{
			Type: "tokenLock",
			TokenLock: &metagraph.TokenLock{
				BaseTx: metagraph.BaseTx{
					SenderDID: f.DID,
					Layer:     metagraph.CurrencyL1,
				},
				Amount:       big.NewInt(f.Amount),
				LockDuration: lockDays,
				UnlocksAt:    f.LockedUntil,
				TierName:     "Founder",
			},
		})
	}
	return out, nil
}
