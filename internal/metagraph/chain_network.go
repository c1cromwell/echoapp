package metagraph

import (
	"fmt"
	"os"
	"strings"
)

// ChainNetworkBesu is Echo's own chain: a private Hyperledger Besu network. It is the only
// network the gateway anchors to (docs/BESU_MIGRATION_PLAN.md).
const ChainNetworkBesu = "besu"

// ParseChainNetwork validates an ECHO_CHAIN_NETWORK value. Empty means the default (besu);
// anything else is an error, so a stale setting from an earlier chain stops startup instead of
// being silently ignored.
func ParseChainNetwork(raw string) (string, error) {
	switch v := strings.ToLower(strings.TrimSpace(raw)); v {
	case "", ChainNetworkBesu:
		return ChainNetworkBesu, nil
	default:
		return "", fmt.Errorf("invalid ECHO_CHAIN_NETWORK %q: must be %q", raw, ChainNetworkBesu)
	}
}

// ChainNetworkFromEnv reads and validates ECHO_CHAIN_NETWORK.
func ChainNetworkFromEnv() (string, error) {
	return ParseChainNetwork(os.Getenv("ECHO_CHAIN_NETWORK"))
}
