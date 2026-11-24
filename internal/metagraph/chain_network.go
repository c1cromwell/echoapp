package metagraph

import (
	"fmt"
	"os"
	"strings"
)

// Chain network names accepted in ECHO_CHAIN_NETWORK.
const (
	// ChainNetworkSovereign is Echo's own self-run chain (default).
	ChainNetworkSovereign = "sovereign"
	// ChainNetworkConstellation is Constellation's public network.
	ChainNetworkConstellation = "constellation"
)

// ParseChainNetwork validates an ECHO_CHAIN_NETWORK value. Empty means the
// default (sovereign); anything other than the two known names is an error.
func ParseChainNetwork(raw string) (string, error) {
	switch v := strings.ToLower(strings.TrimSpace(raw)); v {
	case "":
		return ChainNetworkSovereign, nil
	case ChainNetworkSovereign, ChainNetworkConstellation:
		return v, nil
	default:
		return "", fmt.Errorf("invalid ECHO_CHAIN_NETWORK %q: must be %q or %q",
			raw, ChainNetworkSovereign, ChainNetworkConstellation)
	}
}

// ChainNetworkFromEnv reads and validates ECHO_CHAIN_NETWORK.
func ChainNetworkFromEnv() (string, error) {
	return ParseChainNetwork(os.Getenv("ECHO_CHAIN_NETWORK"))
}
