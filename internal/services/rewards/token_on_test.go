package rewards

import (
	"os"
	"testing"
)

// These tests exercise the token ledger, which is switched off by default while the token is on
// hold (internal/features). They run with it switched on.
func TestMain(m *testing.M) {
	os.Setenv("ECHO_TOKEN_ENABLED", "true")
	os.Setenv("ECHO_WALLET_GENESIS_AUTO", "1")
	os.Exit(m.Run())
}
