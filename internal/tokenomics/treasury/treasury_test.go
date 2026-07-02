package treasury

import (
	"testing"

	"github.com/thechadcromwell/echoapp/internal/tokenomics/genesis"
)

func TestSeedPacaSwap(t *testing.T) {
	m := NewManager(genesis.DefaultFounderDIDs())
	if err := m.SeedPacaSwap("tx_seed_1", 1_000_000); err != nil {
		t.Fatal(err)
	}
	st := m.GetState()
	if !st.PacaSwapSeeded {
		t.Error("expected paca swap seeded")
	}
	if st.PacaSwapSeedTx != "tx_seed_1" {
		t.Errorf("tx hash: got %s", st.PacaSwapSeedTx)
	}
}

func TestMultisigProposal_3of5(t *testing.T) {
	founders := genesis.DefaultFounderDIDs()
	m := NewManager(founders)
	p, err := m.CreateProposal("prop1", "LP incentive", "pool:echo-dag", "lp_mining", 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	for _, did := range founders[:2] {
		if _, err := m.SignProposal(p.ID, did); err != nil {
			t.Fatal(err)
		}
	}
	_, err = m.ExecuteProposal(p.ID, "tx_exec")
	if err != ErrInsufficientSignatures {
		t.Errorf("expected insufficient signatures, got %v", err)
	}
	if _, err := m.SignProposal(p.ID, founders[2]); err != nil {
		t.Fatal(err)
	}
	executed, err := m.ExecuteProposal(p.ID, "tx_exec")
	if err != nil {
		t.Fatal(err)
	}
	if !executed.Executed {
		t.Error("expected executed")
	}
}
