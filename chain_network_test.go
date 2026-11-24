package main

import (
	"strings"
	"testing"
)

// An unknown ECHO_CHAIN_NETWORK must stop startup before anything is wired.
func TestStartRejectsUnknownChainNetwork(t *testing.T) {
	t.Setenv("ECHO_CHAIN_NETWORK", "mainnet")
	err := NewServer(ServerConfig{Port: "0"}).Start()
	if err == nil || !strings.Contains(err.Error(), "ECHO_CHAIN_NETWORK") {
		t.Fatalf("Start() error = %v, want an ECHO_CHAIN_NETWORK error", err)
	}
}
