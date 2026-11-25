package besu

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// fixtureCheckpoint builds a checkpoint from the signed header inside the recorded proof.
func fixtureCheckpoint(t *testing.T) (*Checkpoint, []common.Address) {
	t.Helper()
	p, validators := loadFixture(t)
	var header types.Header
	if err := rlp.DecodeBytes(p.HeaderRLP, &header); err != nil {
		t.Fatal(err)
	}
	cp := &Checkpoint{ChainID: 20261, Number: header.Number.Uint64(), HeaderRLP: p.HeaderRLP}
	if err := cp.fillFromHeader(); err != nil {
		t.Fatal(err)
	}
	return cp, validators
}

func TestCheckpointVerifiesOffline(t *testing.T) {
	cp, validators := fixtureCheckpoint(t)
	signers, err := VerifyCheckpoint(cp, validators)
	if err != nil {
		t.Fatalf("a checkpoint of a real signed block must verify: %v", err)
	}
	if len(signers) < Quorum(len(validators)) {
		t.Fatalf("%d signers", len(signers))
	}
}

func TestCheckpointRejectsTamperingAndStrangers(t *testing.T) {
	good, validators := fixtureCheckpoint(t)
	copyOf := func() *Checkpoint { c := *good; c.HeaderRLP = append([]byte(nil), good.HeaderRLP...); return &c }

	wrongNumber := copyOf()
	wrongNumber.Number++
	wrongHash := copyOf()
	wrongHash.Hash[0] ^= 1
	wrongTime := copyOf()
	wrongTime.Time = wrongTime.Time.Add(time.Second)
	flipped := copyOf()
	flipped.HeaderRLP[len(flipped.HeaderRLP)/3] ^= 1

	for name, cp := range map[string]*Checkpoint{
		"a different block number":      wrongNumber,
		"a different hash":              wrongHash,
		"a different time":              wrongTime,
		"a header with one bit changed": flipped,
	} {
		if _, err := VerifyCheckpoint(cp, validators); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	// Signed by validators the checker does not trust.
	stranger := []common.Address{common.HexToAddress("0x1"), common.HexToAddress("0x2"), common.HexToAddress("0x3"), common.HexToAddress("0x4")}
	if _, err := VerifyCheckpoint(good, stranger); err == nil {
		t.Error("a checkpoint signed by untrusted validators was accepted")
	}
	// Too few of the trusted set signed: trusting seven means five must sign.
	wider := append(append([]common.Address{}, validators...), common.HexToAddress("0x5"), common.HexToAddress("0x6"), common.HexToAddress("0x7"))
	if _, err := VerifyCheckpoint(good, wider); err == nil {
		t.Error("a checkpoint without a quorum of the trusted set was accepted")
	}
	if _, err := VerifyCheckpoint(good, nil); err == nil {
		t.Error("verifying with no trusted validators was accepted")
	}
}

// Live: a checkpoint taken from one validator verifies, and every other validator reports the
// same block at that height. A checkpoint altered to name a different block is reported as a
// rewrite of history.
func TestLiveCheckpointMatchesEveryValidator(t *testing.T) {
	urls, validators := os.Getenv("BESU_RPC_URLS"), os.Getenv("BESU_VALIDATORS")
	if urls == "" || validators == "" {
		t.Skip("live chain not configured (BESU_RPC_URLS, BESU_VALIDATORS)")
	}
	trusted, err := ParseAddresses(validators)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	nodes := strings.Split(urls, ",")
	cp, err := TakeCheckpoint(ctx, nodes[0])
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyCheckpoint(cp, trusted); err != nil {
		t.Fatalf("live checkpoint does not verify: %v", err)
	}
	for _, node := range nodes {
		// A validator a block behind has not got there yet; give it a moment.
		var err error
		for i := 0; i < 20; i++ {
			if err = CompareWithChain(ctx, cp, node); err == nil || !strings.Contains(err.Error(), "has no block") {
				break
			}
			time.Sleep(time.Second)
		}
		if err != nil {
			t.Errorf("%s: %v", node, err)
		}
	}
	forged := *cp
	forged.Hash[5] ^= 0xff
	if err := CompareWithChain(ctx, &forged, nodes[0]); !errors.Is(err, ErrHistoryRewritten) {
		t.Fatalf("a checkpoint naming a different block: err = %v, want ErrHistoryRewritten", err)
	}
	otherChain := *cp
	otherChain.ChainID++
	if err := CompareWithChain(ctx, &otherChain, nodes[0]); err == nil {
		t.Fatal("a checkpoint for another chain was accepted")
	}
}
