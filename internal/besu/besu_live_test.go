package besu

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Live tests run against a real chain (`make dev`, then `make chain-test`). They are skipped
// unless BESU_RPC_URLS, BESU_ANCHOR_CONTRACT, BESU_VALIDATORS and IDENTITY_SERVICE_KEY_PEM are set.
func liveClient(t *testing.T) (*Client, []common.Address) {
	t.Helper()
	urls, contract, validators, keyPath := os.Getenv("BESU_RPC_URLS"), os.Getenv("BESU_ANCHOR_CONTRACT"), os.Getenv("BESU_VALIDATORS"), os.Getenv("IDENTITY_SERVICE_KEY_PEM")
	if urls == "" || contract == "" || validators == "" || keyPath == "" {
		t.Skip("live chain not configured (BESU_RPC_URLS, BESU_ANCHOR_CONTRACT, BESU_VALIDATORS, IDENTITY_SERVICE_KEY_PEM)")
	}
	key, err := LoadKeyFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	trusted, err := ParseAddresses(validators)
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(Config{RPCURLs: strings.Split(urls, ","), AnchorContract: common.HexToAddress(contract),
		IdentityContract: common.HexToAddress(os.Getenv("BESU_IDENTITY_CONTRACT")), Signer: key})
	if err != nil {
		t.Fatal(err)
	}
	return c, trusted
}

func randomHash(t *testing.T) string {
	t.Helper()
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

func TestLiveAnchorProveVerify(t *testing.T) {
	c, trusted := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := randomHash(t)

	if _, err := c.RootStatus(ctx, root); !errors.Is(err, ErrNotAnchored) {
		t.Fatalf("status before anchoring: err = %v, want ErrNotAnchored", err)
	}
	if _, _, err := c.ProveRoot(ctx, root); !errors.Is(err, ErrNotAnchored) {
		t.Fatalf("proof before anchoring: err = %v, want ErrNotAnchored", err)
	}

	receipt, err := c.AnchorRoot(ctx, root, 7)
	if err != nil {
		t.Fatal(err)
	}
	status, err := c.RootStatus(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	if status.BlockNumber != receipt.BlockNumber || status.LeafCount != 7 {
		t.Fatalf("status = %+v, want block %d with 7 leaves", status, receipt.BlockNumber)
	}

	proof, _, err := c.ProveRoot(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	v, err := Verify(proof, trusted)
	if err != nil {
		t.Fatalf("a fresh proof must verify: %v", err)
	}
	if v.AnchoredBlock != receipt.BlockNumber || v.LeafCount != 7 || len(v.Signers) < v.QuorumRequired {
		t.Fatalf("verified = %+v", v)
	}

	// The contract refuses to anchor the same root twice.
	if _, err := c.AnchorRoot(ctx, root, 7); err == nil {
		t.Fatal("anchoring the same root twice must fail")
	}

	if os.Getenv("BESU_UPDATE_FIXTURE") == "1" {
		body, _ := json.MarshalIndent(struct {
			Validators []common.Address `json:"validators"`
			Proof      *Proof           `json:"proof"`
		}{trusted, proof}, "", "  ")
		if err := os.WriteFile("testdata/proof.json", append(body, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLiveUnlistedSignerIsRejected(t *testing.T) {
	c, _ := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	stranger, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	other, err := New(Config{RPCURLs: c.cfg.RPCURLs, AnchorContract: c.cfg.AnchorContract, Signer: stranger})
	if err != nil {
		t.Fatal(err)
	}
	root := randomHash(t)
	if _, err := other.AnchorRoot(ctx, root, 1); err == nil {
		t.Fatal("a key that is not an authorised signer anchored a root")
	} else {
		t.Logf("rejected as expected: %v", err)
	}
	if _, err := c.RootStatus(ctx, root); !errors.Is(err, ErrNotAnchored) {
		t.Fatalf("the rejected root is on the chain: %v", err)
	}
}

func TestLiveCommitment(t *testing.T) {
	c, _ := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := c.AnchorCommitment(ctx, randomHash(t), 3); err != nil {
		t.Fatal(err)
	}
}

func TestLiveFailsOverToAnotherEndpoint(t *testing.T) {
	c, _ := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// A dead endpoint first: the client must move on to a live one.
	withDead, err := New(Config{RPCURLs: append([]string{"http://127.0.0.1:9"}, c.cfg.RPCURLs...), AnchorContract: c.cfg.AnchorContract, Signer: c.cfg.Signer})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := withDead.AnchorRoot(ctx, randomHash(t), 1); err != nil {
		t.Fatalf("anchoring with one dead endpoint: %v", err)
	}
}

// Three separate clients share one signing key and write at the same time, as gateway replicas
// do. They pick the same transaction numbers; every write must still land exactly once.
func TestLiveSeveralWritersShareOneKey(t *testing.T) {
	c, _ := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	const writers, each = 3, 6
	roots := make([][]string, writers)
	errs := make(chan error, writers*each)
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		replica, err := New(Config{RPCURLs: c.cfg.RPCURLs, AnchorContract: c.cfg.AnchorContract, Signer: c.cfg.Signer})
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < each; i++ {
			roots[w] = append(roots[w], randomHash(t))
		}
		wg.Add(1)
		go func(w int, replica *Client) {
			defer wg.Done()
			for _, root := range roots[w] {
				if _, err := replica.AnchorRoot(ctx, root, 1); err != nil {
					errs <- fmt.Errorf("writer %d: %w", w, err)
				}
			}
		}(w, replica)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	for w := range roots {
		for _, root := range roots[w] {
			if _, err := c.RootStatus(ctx, root); err != nil {
				t.Errorf("writer %d: root %s is not on the chain: %v", w, root[:8], err)
			}
		}
	}
}
