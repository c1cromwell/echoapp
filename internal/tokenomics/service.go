// Package tokenomics holds the in-app score: quests, the leaderboard, streaks and the weekly pack.
// Everything is counted in points. Points are a score, not a token: they hold no value and convert
// to nothing. (The package name predates the removal of the token; see docs/BESU_MIGRATION_PLAN.md.)
package tokenomics

import (
	"github.com/thechadcromwell/echoapp/internal/tokenomics/leaderboard"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/packs"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/quests"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/streaks"
)

// Service is the facade the API handlers use.
type Service struct {
	Quests      *quests.Service
	Leaderboard *leaderboard.Service
	Streaks     *streaks.Service
	Packs       *packs.Service
}

// Config holds the service's dependencies.
type Config struct {
	// QuestStore persists quest completion; in memory when nil.
	QuestStore quests.Store
}

// NewService constructs the service.
func NewService(cfg Config) (*Service, error) {
	store := cfg.QuestStore
	if store == nil {
		store = quests.NewMemStore()
	}
	return &Service{
		Quests:      quests.NewService(store),
		Leaderboard: leaderboard.NewService(nil),
		Streaks:     streaks.NewService(nil),
		Packs:       packs.NewService(),
	}, nil
}
