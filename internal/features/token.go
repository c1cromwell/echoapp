// Package features holds product switches that more than one package needs to agree on.
package features

import (
	"os"
	"strings"
)

// TokenEnabled reports whether ECHO token features are switched on.
//
// The token is on hold (docs/BESU_MIGRATION_PLAN.md §0): by default there is no wallet, balance,
// staking, token reward or any other token surface in the API, and nothing accrues toward one.
// Echo Score, streaks, quests and the leaderboard remain, scored in points that are not a token
// and are not convertible to one. Setting ECHO_TOKEN_ENABLED=true restores the token surface for
// development; it must not be set in a release environment while the hold stands.
func TokenEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ECHO_TOKEN_ENABLED")), "true")
}

// PointsPerUnit converts the ledger's internal 8-decimal amounts to whole points for display.
const PointsPerUnit = 100_000_000
