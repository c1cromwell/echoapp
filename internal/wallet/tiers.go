package wallet

import "fmt"

// ChainTier maps API tier names to Currency L1 tier vocabulary.
type ChainTier struct {
	ScalaName string
	LockDays  int
	APR       float64
	MinDatum  int64
}

var chainTiers = map[string]ChainTier{
	"bronze":   {ScalaName: "Tier 1", LockDays: 30, APR: 5.0, MinDatum: 10_000_000_000},
	"silver":   {ScalaName: "Tier 2", LockDays: 90, APR: 8.0, MinDatum: 100_000_000_000},
	"gold":     {ScalaName: "Tier 3", LockDays: 180, APR: 10.0, MinDatum: 1_000_000_000_000},
	"platinum": {ScalaName: "Tier 5", LockDays: 365, APR: 15.0, MinDatum: 100_000_000_000_000},
}

// ResolveChainTier returns L1 tier metadata for an API tier name.
func ResolveChainTier(name string) (ChainTier, error) {
	t, ok := chainTiers[name]
	if !ok {
		return ChainTier{}, fmt.Errorf("%w: %s", ErrInvalidTier, name)
	}
	return t, nil
}

// DatumPerECHO is the on-chain / API integer scale (1 ECHO = 1e8 datum).
const DatumPerECHO = int64(100_000_000)
