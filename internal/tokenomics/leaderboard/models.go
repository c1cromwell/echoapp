// Package leaderboard implements the usage-leaderboard gamification mechanic
// (requirements §Mechanic 1). It ranks users by ECHO earned within a rolling
// window. This is a display-only reputation ranking for the value-free launch
// (Wave R0): the ECHO it counts is earned into the interim, non-redeemable,
// non-transferable wallet, so appearing here confers standing, not money.
package leaderboard

import "time"

// Window is a leaderboard ranking period.
type Window string

const (
	WindowWeekly  Window = "weekly"
	WindowMonthly Window = "monthly"
)

// Valid reports whether w is a recognized window.
func (w Window) Valid() bool {
	return w == WindowWeekly || w == WindowMonthly
}

// MinTrustTier is the minimum trust tier required to appear on the leaderboard
// (requirements §Mechanic 1: "min Trust Tier 2 to appear"). Lower-tier users
// still earn ECHO — they just aren't ranked, which blunts sybil/gaming.
const MinTrustTier = 2

// Entry is one ranked participant within a window.
type Entry struct {
	DID       string `json:"did"`
	TrustTier int    `json:"trust_tier"`
	// Score is ECHO earned in the window, in datum (8-decimal base units).
	Score int64 `json:"score"`
	Rank  int   `json:"rank"`
}

// Snapshot is a ranked slice for a given window bucket.
type Snapshot struct {
	Window    Window    `json:"window"`
	BucketKey string    `json:"bucket_key"`
	Entries   []Entry   `json:"entries"`
	UpdatedAt time.Time `json:"updated_at"`
}
