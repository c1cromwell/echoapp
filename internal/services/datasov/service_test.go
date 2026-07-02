package datasov

import "testing"

func TestService_OptInAndContribute(t *testing.T) {
	s := NewService()
	if _, err := s.Contribute("did:key:alice", "abc123"); err != ErrNotEnabled {
		t.Fatalf("expected not enabled, got %v", err)
	}
	s.SetOptIn("did:key:alice", true)
	c, err := s.Contribute("did:key:alice", "abc123hash")
	if err != nil || c.StatsHash != "abc123hash" {
		t.Fatalf("contribute failed: %+v %v", c, err)
	}
}
