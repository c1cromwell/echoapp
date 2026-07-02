// Package tokenomics wires genesis, treasury, quests, emission, VIP, and founder revocation (WO-206/214/215/225/271).
package tokenomics

import (
	"context"
	"time"

	"github.com/thechadcromwell/echoapp/internal/infra"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/emission"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/founder"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/genesis"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/quests"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/treasury"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/vip"
	"github.com/thechadcromwell/echoapp/internal/wallet"
)

// Service is the top-level tokenomics facade used by API handlers.
type Service struct {
	GenesisDate time.Time
	Genesis     *genesis.Snapshot
	Treasury    *treasury.Manager
	Quests      *quests.Service
	Emission    *emission.Tracker
	VIP         *vip.SubscriptionService
	Revocation  *founder.Coordinator
	Wallet      *wallet.WalletService
}

// Config holds tokenomics service dependencies.
type Config struct {
	GenesisDate  time.Time
	CurrencyL1   *metagraph.MetagraphClient
	Redis        *infra.RedisClient
	QuestStore   quests.Store
	Wallet       *wallet.WalletService
	ClaimedToday func() int64
	BackendDID   string
}

// NewService constructs the tokenomics service bundle.
func NewService(cfg Config) (*Service, error) {
	if cfg.GenesisDate.IsZero() {
		cfg.GenesisDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	snap, err := genesis.BuildSnapshot(cfg.GenesisDate)
	if err != nil {
		return nil, err
	}

	store := cfg.QuestStore
	if store == nil {
		store = quests.NewMemStore()
	}
	questSvc := quests.NewService(store)

	treasuryMgr := treasury.NewManager(genesis.DefaultFounderDIDs())
	_ = treasuryMgr.SeedPacaSwap("genesis_pacaswap_seed", 0)

	backendDID := cfg.BackendDID
	if backendDID == "" {
		backendDID = "did:key:z6MkEchoBackend"
	}
	vipSvc := vip.NewSubscriptionService(backendDID).WithCurrencyL1(cfg.CurrencyL1).WithRedis(cfg.Redis)

	claimed := cfg.ClaimedToday
	if claimed == nil {
		claimed = func() int64 { return 0 }
	}

	return &Service{
		GenesisDate: cfg.GenesisDate,
		Genesis:     snap,
		Treasury:    treasuryMgr,
		Quests:      questSvc,
		Emission:    emission.NewTracker(cfg.GenesisDate, claimed),
		VIP:         vipSvc,
		Revocation:  founder.NewCoordinator(),
		Wallet:      cfg.Wallet,
	}, nil
}

// VestingInfo returns founder vesting for authenticated DID (WO-226).
func (s *Service) VestingInfo(ctx context.Context, did string) (*wallet.VestingState, error) {
	if !genesis.IsFounderDID(did) {
		return nil, wallet.ErrVestingNotFound
	}
	if s.Wallet == nil {
		return nil, wallet.ErrVestingNotFound
	}
	state, err := s.Wallet.GetWalletState(ctx, did)
	if err != nil {
		return nil, err
	}
	if state.Vesting == nil {
		return nil, wallet.ErrVestingNotFound
	}
	return state.Vesting, nil
}
