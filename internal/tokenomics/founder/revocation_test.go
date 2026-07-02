package founder

import (
	"testing"

	"github.com/thechadcromwell/echoapp/internal/tokenomics/genesis"
)

func TestRevocation_3of5_FutureTeamPool(t *testing.T) {
	founders := genesis.DefaultFounderDIDs()
	target := founders[1]
	coord := NewCoordinator()

	req, err := coord.Initiate("rev1", target, genesis.FounderMemberAmount/2)
	if err != nil {
		t.Fatal(err)
	}
	signers := []string{founders[0], founders[2], founders[3]}
	for _, s := range signers[:2] {
		if _, err := coord.Sign(req.ID, s); err != nil {
			t.Fatal(err)
		}
	}
	_, err = coord.Finalize(req.ID)
	if err != ErrInsufficientSignatures {
		t.Fatalf("expected insufficient sigs, got %v", err)
	}
	if _, err := coord.Sign(req.ID, signers[2]); err != nil {
		t.Fatal(err)
	}
	ev, err := coord.Finalize(req.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ev.DestinationPool != genesis.PoolFutureTeam {
		t.Errorf("expected future team pool, got %s", ev.DestinationPool)
	}
	if len(ev.RevokerDIDs) != 3 {
		t.Errorf("expected 3 revokers, got %d", len(ev.RevokerDIDs))
	}
	remaining := coord.RemainingLocked(target)
	expected := genesis.FounderMemberAmount - genesis.FounderMemberAmount/2
	if remaining != expected {
		t.Errorf("remaining locked: got %d want %d", remaining, expected)
	}
}
