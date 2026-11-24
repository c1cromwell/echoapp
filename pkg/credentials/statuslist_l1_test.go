package credentials

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeChain records what the publisher anchors.
type fakeChain struct {
	mu        sync.Mutex
	lists     [][]byte
	sequences []int64
	issuers   []string
	issuances []string
	failNext  int // number of PublishStatusList calls to fail before succeeding
}

func (f *fakeChain) PublishCredentialIssuance(_ context.Context, credentialID, subjectDID, issuerDID, kind string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.issuances = append(f.issuances, credentialID+"|"+subjectDID+"|"+issuerDID+"|"+kind)
	return "0xtx", nil
}

func (f *fakeChain) PublishStatusList(_ context.Context, issuerDID string, sequence int64, list []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNext > 0 {
		f.failNext--
		return errors.New("chain unavailable")
	}
	f.lists = append(f.lists, append([]byte(nil), list...))
	f.sequences = append(f.sequences, sequence)
	f.issuers = append(f.issuers, issuerDID)
	return nil
}

func (f *fakeChain) published() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.lists)
}

func testStatusConfig(chain ChainPublisher) MetagraphConfig {
	return MetagraphConfig{
		Chain:                     chain,
		IssuerDID:                 "did:key:z6MkTestIssuer",
		EnableAnchor:              true,
		Timeout:                   3 * time.Second,
		MaxRetries:                3,
		RetryBackoff:              5 * time.Millisecond,
		StatusListPublishInterval: time.Hour,
	}
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for !cond() {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %s", what)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestStatusListMarkRevokedAsyncPublish(t *testing.T) {
	chain := &fakeChain{}
	p := NewStatusListPublisher(testStatusConfig(chain))
	p.Start()
	defer p.Stop()

	idx, err := p.AllocateIndex("cred-1")
	if err != nil {
		t.Fatal(err)
	}
	if !p.MarkRevoked("cred-1") {
		t.Fatal("expected MarkRevoked true")
	}
	if !p.PublishPending() {
		t.Fatal("expected pending before the chain acknowledges")
	}
	waitFor(t, "a publication", func() bool { return chain.published() >= 1 })
	waitFor(t, "pending to clear", func() bool { return !p.PublishPending() })

	chain.mu.Lock()
	defer chain.mu.Unlock()
	list := chain.lists[0]
	if len(list) != statusListBitCount/8 {
		t.Fatalf("list is %d bytes, want %d", len(list), statusListBitCount/8)
	}
	// The revoked credential's bit, and only that bit, is set.
	set := 0
	for _, b := range list {
		for ; b != 0; b &= b - 1 {
			set++
		}
	}
	if set != 1 || list[idx/8] == 0 {
		t.Fatalf("%d bits set; byte for index %d = %#x", set, idx, list[idx/8])
	}
	if chain.issuers[0] != "did:key:z6MkTestIssuer" {
		t.Fatalf("issuer = %q", chain.issuers[0])
	}
}

func TestStatusListCoalescedSignalsSingleWorker(t *testing.T) {
	chain := &fakeChain{}
	p := NewStatusListPublisher(testStatusConfig(chain))
	p.Start()
	defer p.Stop()

	if _, err := p.AllocateIndex("a"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AllocateIndex("b"); err != nil {
		t.Fatal(err)
	}
	p.MarkRevoked("a")
	p.MarkRevoked("b")
	waitFor(t, "a publication", func() bool { return chain.published() >= 1 })
}

// A later publication must carry a higher sequence, or the chain would refuse it as stale.
func TestStatusListSequenceIncreases(t *testing.T) {
	chain := &fakeChain{}
	p := NewStatusListPublisher(testStatusConfig(chain))
	for _, id := range []string{"a", "b", "c"} {
		if _, err := p.AllocateIndex(id); err != nil {
			t.Fatal(err)
		}
		p.MarkRevoked(id)
		if err := p.Publish(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	chain.mu.Lock()
	defer chain.mu.Unlock()
	if len(chain.sequences) != 3 {
		t.Fatalf("%d publications, want 3", len(chain.sequences))
	}
	for i := 1; i < len(chain.sequences); i++ {
		if chain.sequences[i] <= chain.sequences[i-1] {
			t.Fatalf("sequences %v do not increase", chain.sequences)
		}
	}
}

// A failed publication stays pending and is retried; nothing is marked published that was not.
func TestStatusListRetriesAfterChainFailure(t *testing.T) {
	chain := &fakeChain{failNext: 2}
	p := NewStatusListPublisher(testStatusConfig(chain))
	p.Start()
	defer p.Stop()
	if _, err := p.AllocateIndex("a"); err != nil {
		t.Fatal(err)
	}
	p.MarkRevoked("a")
	waitFor(t, "a publication after two failures", func() bool { return chain.published() >= 1 })
	waitFor(t, "pending to clear", func() bool { return !p.PublishPending() })
}

func TestStatusListPublisherOffWithoutChain(t *testing.T) {
	cfg := testStatusConfig(nil)
	if NewStatusListPublisher(cfg) != nil {
		t.Fatal("publisher created with no chain")
	}
	cfg = testStatusConfig(&fakeChain{})
	cfg.EnableAnchor = false
	if NewStatusListPublisher(cfg) != nil {
		t.Fatal("publisher created with anchoring disabled")
	}
}
