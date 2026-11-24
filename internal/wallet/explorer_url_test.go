package wallet

import "testing"

func TestExplorerURLOnlyOnConstellation(t *testing.T) {
	t.Setenv("ECHO_CHAIN_NETWORK", "sovereign")
	if got := explorerURL("did:key:abc"); got != "" {
		t.Fatalf("sovereign network: explorer URL = %q, want empty", got)
	}
	t.Setenv("ECHO_CHAIN_NETWORK", "")
	if got := explorerURL("did:key:abc"); got != "" {
		t.Fatalf("default network: explorer URL = %q, want empty", got)
	}
	t.Setenv("ECHO_CHAIN_NETWORK", "constellation")
	if got := explorerURL("did:key:abc"); got != "https://dagexplorer.io/address/did:key:abc" {
		t.Fatalf("constellation network: explorer URL = %q", got)
	}
}
