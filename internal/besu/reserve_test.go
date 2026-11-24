package besu

import (
	"context"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// The reserve contract must offer no way to move or create units. This reads the compiled
// interface itself, so it fails if such a function is ever added.
func TestReserveContractCanOnlyBeRead(t *testing.T) {
	if len(reserveABI.Methods) == 0 {
		t.Fatal("reserve ABI has no methods; the embedded artefact is wrong")
	}
	for name, m := range reserveABI.Methods {
		if m.StateMutability != "view" && m.StateMutability != "pure" {
			t.Errorf("reserve contract has a function that changes state: %s (%s)", name, m.StateMutability)
		}
		lower := strings.ToLower(name)
		for _, forbidden := range []string{"transfer", "mint", "approve", "burn", "owner", "withdraw", "set"} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("reserve contract exposes %q", name)
			}
		}
	}
	// The bytecode must not contain the ERC-20 transfer selectors, however a function is named.
	for name, selector := range map[string]string{"transfer(address,uint256)": "a9059cbb", "transferFrom(address,address,uint256)": "23b872dd", "approve(address,uint256)": "095ea7b3"} {
		if strings.Contains(strings.ToLower(reserveBinHex), selector) {
			t.Errorf("reserve bytecode contains the selector for %s", name)
		}
	}
}

// Live: deploy a reserve and read it back. Needs BESU_RPC_URLS and a throwaway deployer key is
// generated here; nothing is reused.
func TestLiveReserveHoldsTheWholeSupplyAtOneAddress(t *testing.T) {
	urls := os.Getenv("BESU_RPC_URLS")
	if urls == "" {
		t.Skip("live chain not configured (BESU_RPC_URLS)")
	}
	rpcURL := strings.Split(urls, ",")[0]
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	deployer, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	reservedKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	reserved := crypto.PubkeyToAddress(reservedKey.PublicKey)

	address, _, err := Deploy(ctx, rpcURL, ContractReserve, deployer, reserved)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ReadReserve(ctx, rpcURL, address)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := new(big.Int).SetString("100000000000000000", 10) // 1,000,000,000 x 10^8
	if r.TotalSupply.Cmp(want) != 0 || r.Decimals != 8 {
		t.Fatalf("supply = %s with %d decimals, want %s with 8", r.TotalSupply, r.Decimals, want)
	}
	if r.Reserved != reserved || r.Balance.Cmp(want) != 0 {
		t.Fatalf("reserved = %s holding %s; want %s holding the whole supply", r.Reserved.Hex(), r.Balance, reserved.Hex())
	}
	for _, other := range []common.Address{crypto.PubkeyToAddress(deployer.PublicKey), common.HexToAddress("0x01"), address} {
		balance, err := ReserveBalanceOf(ctx, rpcURL, address, other)
		if err != nil {
			t.Fatal(err)
		}
		if balance.Sign() != 0 {
			t.Errorf("%s holds %s; only the reserved address may hold anything", other.Hex(), balance)
		}
	}
	// A reserve at the zero address is refused.
	if _, _, err := Deploy(ctx, rpcURL, ContractReserve, deployer, common.Address{}); err == nil {
		t.Error("a reserve was deployed with no reserved address")
	}
}
