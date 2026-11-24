// Command echobesu is the Besu spike client (docs/BESU_MIGRATION_PLAN.md). It is throwaway code
// that answers four questions against a live private Besu network:
//
//	deploy   can Echo's anchoring rules live in a contract, written to only by the gateway's key?
//	anchor   how long from submitting a Merkle root to finality?
//	reject   is a root from an unlisted key refused?
//	prove    can a self-contained proof be exported ...
//	verify   ... and checked with the chain switched off, INCLUDING the validators' signatures?
package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	rpcURL = flag.String("rpc", "http://localhost:8545", "Besu JSON-RPC endpoint")
	dir    = flag.String("dir", ".", "spike directory (holds build/ and .network/)")
)

// anchoredAtSlot is the storage slot of EchoAnchor.anchoredAt (see the contract).
const anchoredAtSlot = 2

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fail(errors.New("usage: echobesu [-rpc URL] [-dir DIR] deploy | anchor [root] | reject | prove <root> <file> | verify <file> <validator,validator,...>"))
	}
	var err error
	switch flag.Arg(0) {
	case "deploy":
		err = deploy()
	case "anchor":
		err = anchor(flag.Arg(1))
	case "reject":
		err = reject()
	case "prove":
		err = prove(flag.Arg(1), flag.Arg(2))
	case "verify":
		err = verify(flag.Arg(1), flag.Arg(2))
	default:
		err = fmt.Errorf("unknown command %q", flag.Arg(0))
	}
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

// --- keys and contract plumbing ---------------------------------------------------------------

// loadOrCreateKey keeps throwaway spike keys under .network/.
func loadOrCreateKey(name string) (*ecdsa.PrivateKey, error) {
	path := filepath.Join(*dir, ".network", name+".key")
	if raw, err := os.ReadFile(path); err == nil {
		return crypto.HexToECDSA(strings.TrimSpace(string(raw)))
	}
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return key, os.WriteFile(path, []byte(hex.EncodeToString(crypto.FromECDSA(key))), 0o600)
}

func contractABI() (abi.ABI, error) {
	raw, err := os.ReadFile(filepath.Join(*dir, "build", "EchoAnchor.abi"))
	if err != nil {
		return abi.ABI{}, err
	}
	return abi.JSON(bytes.NewReader(raw))
}

func contractAddress() (common.Address, error) {
	raw, err := os.ReadFile(filepath.Join(*dir, ".network", "contract.address"))
	if err != nil {
		return common.Address{}, fmt.Errorf("contract not deployed yet (run deploy): %w", err)
	}
	return common.HexToAddress(strings.TrimSpace(string(raw))), nil
}

// txOpts signs with key and pays nothing: the network runs with a zero base fee.
func txOpts(ctx context.Context, c *ethclient.Client, key *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	chainID, err := c.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	opts, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return nil, err
	}
	opts.Context = ctx
	opts.GasPrice = big.NewInt(0)
	return opts, nil
}

func deploy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	c, err := ethclient.DialContext(ctx, *rpcURL)
	if err != nil {
		return err
	}
	owner, err := loadOrCreateKey("owner")
	if err != nil {
		return err
	}
	gateway, err := loadOrCreateKey("gateway")
	if err != nil {
		return err
	}
	parsed, err := contractABI()
	if err != nil {
		return err
	}
	binHex, err := os.ReadFile(filepath.Join(*dir, "build", "EchoAnchor.bin"))
	if err != nil {
		return err
	}
	opts, err := txOpts(ctx, c, owner)
	if err != nil {
		return err
	}
	gatewayAddr := crypto.PubkeyToAddress(gateway.PublicKey)
	addr, tx, _, err := bind.DeployContract(opts, parsed, common.FromHex(strings.TrimSpace(string(binHex))), c, gatewayAddr)
	if err != nil {
		return err
	}
	receipt, err := bind.WaitMined(ctx, c, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return errors.New("deployment reverted")
	}
	fmt.Printf("deployed EchoAnchor at %s in block %d (gas %d); authorised signer %s\n", addr.Hex(), receipt.BlockNumber, receipt.GasUsed, gatewayAddr.Hex())
	return os.WriteFile(filepath.Join(*dir, ".network", "contract.address"), []byte(addr.Hex()), 0o644)
}

func randomRoot() (common.Hash, error) {
	var h common.Hash
	_, err := rand.Read(h[:])
	return h, err
}

func anchor(rootHex string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	c, err := ethclient.DialContext(ctx, *rpcURL)
	if err != nil {
		return err
	}
	root := common.HexToHash(rootHex)
	if rootHex == "" {
		if root, err = randomRoot(); err != nil {
			return err
		}
	}
	gateway, err := loadOrCreateKey("gateway")
	if err != nil {
		return err
	}
	parsed, err := contractABI()
	if err != nil {
		return err
	}
	addr, err := contractAddress()
	if err != nil {
		return err
	}
	opts, err := txOpts(ctx, c, gateway)
	if err != nil {
		return err
	}
	start := time.Now()
	tx, err := bind.NewBoundContract(addr, parsed, c, c, c).Transact(opts, "anchorRoot", root, uint32(1))
	if err != nil {
		return err
	}
	receipt, err := bind.WaitMined(ctx, c, tx)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return errors.New("anchorRoot reverted")
	}
	// QBFT has immediate finality: a block that exists is final, there are no forks to wait out.
	fmt.Printf("anchored %s in block %d, %s after submit (gas %d)\n", root.Hex(), receipt.BlockNumber, time.Since(start).Round(time.Millisecond), receipt.GasUsed)
	return nil
}

// reject submits a root from a key that is not an authorised signer and requires it to fail.
func reject() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	c, err := ethclient.DialContext(ctx, *rpcURL)
	if err != nil {
		return err
	}
	stranger, err := crypto.GenerateKey()
	if err != nil {
		return err
	}
	parsed, err := contractABI()
	if err != nil {
		return err
	}
	addr, err := contractAddress()
	if err != nil {
		return err
	}
	root, err := randomRoot()
	if err != nil {
		return err
	}
	opts, err := txOpts(ctx, c, stranger)
	if err != nil {
		return err
	}
	// Skip gas estimation (it would refuse to build the transaction) so the chain itself decides.
	opts.GasLimit = 200_000
	bound := bind.NewBoundContract(addr, parsed, c, c, c)
	tx, err := bound.Transact(opts, "anchorRoot", root, uint32(1))
	if err != nil {
		return fmt.Errorf("could not submit the unauthorised transaction: %w", err)
	}
	receipt, err := bind.WaitMined(ctx, c, tx)
	if err != nil {
		return err
	}
	var out []interface{}
	if err := bound.Call(&bind.CallOpts{Context: ctx}, &out, "anchoredAt", root); err != nil {
		return err
	}
	stored := out[0].(*big.Int)
	if receipt.Status == types.ReceiptStatusSuccessful || stored.Sign() != 0 {
		return fmt.Errorf("UNAUTHORISED ROOT WAS ACCEPTED (status %d, stored %s)", receipt.Status, stored)
	}
	fmt.Printf("unauthorised signer %s: transaction mined in block %d and reverted; root not stored\n", crypto.PubkeyToAddress(stranger.PublicKey).Hex(), receipt.BlockNumber)
	return nil
}

// --- self-contained proof --------------------------------------------------------------------

// Bundle is everything a verifier needs. Nothing in it is trusted; the verifier's only trust
// anchor is the validator addresses it is given separately.
type Bundle struct {
	Root         common.Hash     `json:"root"`
	Contract     common.Address  `json:"contract"`
	HeaderRLP    hexutil.Bytes   `json:"header_rlp"`
	AccountProof []hexutil.Bytes `json:"account_proof"`
	StorageProof []hexutil.Bytes `json:"storage_proof"`
}

func storageKey(root common.Hash) common.Hash {
	return crypto.Keccak256Hash(root[:], common.LeftPadBytes(big.NewInt(anchoredAtSlot).Bytes(), 32))
}

func prove(rootHex, out string) error {
	if rootHex == "" || out == "" {
		return errors.New("usage: prove <root> <file>")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	c, err := ethclient.DialContext(ctx, *rpcURL)
	if err != nil {
		return err
	}
	addr, err := contractAddress()
	if err != nil {
		return err
	}
	root := common.HexToHash(rootHex)
	number, err := c.BlockNumber(ctx)
	if err != nil {
		return err
	}
	blockHex := hexutil.EncodeUint64(number)

	var raw hexutil.Bytes
	if err := c.Client().CallContext(ctx, &raw, "debug_getRawHeader", blockHex); err != nil {
		return fmt.Errorf("debug_getRawHeader: %w", err)
	}
	var proof struct {
		AccountProof []hexutil.Bytes `json:"accountProof"`
		StorageProof []struct {
			Proof []hexutil.Bytes `json:"proof"`
		} `json:"storageProof"`
	}
	if err := c.Client().CallContext(ctx, &proof, "eth_getProof", addr, []string{storageKey(root).Hex()}, blockHex); err != nil {
		return fmt.Errorf("eth_getProof: %w", err)
	}
	if len(proof.StorageProof) != 1 {
		return errors.New("eth_getProof returned no storage proof")
	}
	body, err := json.MarshalIndent(Bundle{Root: root, Contract: addr, HeaderRLP: raw, AccountProof: proof.AccountProof, StorageProof: proof.StorageProof[0].Proof}, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("proof for %s at block %d written to %s (%d bytes)\n", root.Hex(), number, out, len(body))
	return os.WriteFile(out, body, 0o644)
}

// qbftExtra is the RLP layout of a QBFT block header's extraData.
type qbftExtra struct {
	Vanity     []byte
	Validators []common.Address
	Vote       rlp.RawValue
	Round      rlp.RawValue
	Seals      [][]byte
}

// sealHash is what each validator signs when it commits a block: the header hash computed with
// the commit seals removed from extraData (the round number stays).
func sealHash(header *types.Header, extra qbftExtra) (common.Hash, error) {
	stripped, err := rlp.EncodeToBytes([]interface{}{extra.Vanity, extra.Validators, extra.Vote, extra.Round, [][]byte{}})
	if err != nil {
		return common.Hash{}, err
	}
	h := types.CopyHeader(header)
	h.Extra = stripped
	return h.Hash(), nil
}

func proofDB(nodes []hexutil.Bytes) *memorydb.Database {
	db := memorydb.New()
	for _, n := range nodes {
		_ = db.Put(crypto.Keccak256(n), n)
	}
	return db
}

// verify makes no network call.
func verify(file, validatorList string) error {
	if file == "" || validatorList == "" {
		return errors.New("usage: verify <file> <validator,validator,...>")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var b Bundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return err
	}
	trusted := map[common.Address]bool{}
	for _, v := range strings.Split(validatorList, ",") {
		trusted[common.HexToAddress(strings.TrimSpace(v))] = true
	}

	// 1. The header was committed by a quorum of the trusted validators.
	var header types.Header
	if err := rlp.DecodeBytes(b.HeaderRLP, &header); err != nil {
		return fmt.Errorf("decode header: %w", err)
	}
	var extra qbftExtra
	if err := rlp.DecodeBytes(header.Extra, &extra); err != nil {
		return fmt.Errorf("decode QBFT extraData: %w", err)
	}
	digest, err := sealHash(&header, extra)
	if err != nil {
		return err
	}
	signers := map[common.Address]bool{}
	for _, seal := range extra.Seals {
		pub, err := crypto.SigToPub(digest[:], seal)
		if err != nil {
			return fmt.Errorf("bad commit seal: %w", err)
		}
		signer := crypto.PubkeyToAddress(*pub)
		if !trusted[signer] {
			return fmt.Errorf("commit seal from %s, which is not a trusted validator", signer.Hex())
		}
		signers[signer] = true
	}
	// QBFT needs more than two thirds: ceil(2n/3).
	quorum := (2*len(trusted) + 2) / 3
	if len(signers) < quorum {
		return fmt.Errorf("only %d of %d trusted validators signed; need %d", len(signers), len(trusted), quorum)
	}

	// 2. The contract's storage root is in that header's state.
	accountRLP, err := trie.VerifyProof(header.Root, crypto.Keccak256(b.Contract[:]), proofDB(b.AccountProof))
	if err != nil || accountRLP == nil {
		return fmt.Errorf("account proof does not verify against the header's state root: %v", err)
	}
	var account struct {
		Nonce       uint64
		Balance     *big.Int
		StorageRoot common.Hash
		CodeHash    []byte
	}
	if err := rlp.DecodeBytes(accountRLP, &account); err != nil {
		return fmt.Errorf("decode account: %w", err)
	}

	// 3. The root is recorded in the contract's storage.
	key := storageKey(b.Root)
	valueRLP, err := trie.VerifyProof(account.StorageRoot, crypto.Keccak256(key[:]), proofDB(b.StorageProof))
	if err != nil {
		return fmt.Errorf("storage proof does not verify: %v", err)
	}
	if valueRLP == nil {
		return fmt.Errorf("root %s is NOT anchored at block %d (proof of absence)", b.Root.Hex(), header.Number)
	}
	var value []byte
	if err := rlp.DecodeBytes(valueRLP, &value); err != nil {
		return fmt.Errorf("decode storage value: %w", err)
	}
	packed := new(big.Int).SetBytes(value)
	anchoredBlock := new(big.Int).Rsh(packed, 32)
	leafCount := new(big.Int).And(packed, big.NewInt(0xffffffff))

	fmt.Printf("VERIFIED offline: root %s\n", b.Root.Hex())
	fmt.Printf("  anchored in block %s with %s leaves, contract %s\n", anchoredBlock, leafCount, b.Contract.Hex())
	fmt.Printf("  proven at block %d (%s), signed by %d of %d trusted validators (quorum %d)\n",
		header.Number, time.Unix(int64(header.Time), 0).UTC().Format(time.RFC3339), len(signers), len(trusted), quorum)
	return nil
}
