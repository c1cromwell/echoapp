// Package besu is the gateway's client for Echo's own chain: a private Hyperledger Besu network
// with QBFT consensus (docs/BESU_MIGRATION_PLAN.md).
//
// Finality: QBFT never forks. A transaction whose receipt exists is final; there is no
// confirmation depth to wait for.
package besu

import (
	"context"
	"crypto/ecdsa"
	_ "embed"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

//go:embed contract/EchoAnchor.abi
var anchorABIJSON string

//go:embed contract/EchoAnchor.bin
var anchorBinHex string

//go:embed contract/EchoIdentity.abi
var identityABIJSON string

//go:embed contract/EchoIdentity.bin
var identityBinHex string

//go:embed contract/EchoReserve.abi
var reserveABIJSON string

//go:embed contract/EchoReserve.bin
var reserveBinHex string

var (
	anchorABI   = mustABI(anchorABIJSON)
	identityABI = mustABI(identityABIJSON)
	reserveABI  = mustABI(reserveABIJSON)
)

// Contract names accepted by Deploy.
const (
	ContractAnchor   = "EchoAnchor"
	ContractIdentity = "EchoIdentity"
	// ContractReserve holds the fixed supply at one address. Its constructor argument is that
	// address (passed to Deploy as `signer`); it has no writer.
	ContractReserve = "EchoReserve"
)

func mustABI(raw string) abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(raw))
	if err != nil {
		panic("besu: embedded contract ABI is invalid: " + err.Error())
	}
	return parsed
}

// Config configures a Client.
type Config struct {
	// RPCURLs are the JSON-RPC endpoints of Echo's validators. Each call uses the one that is
	// furthest ahead among those that answer, so the gateway keeps working while any is up.
	RPCURLs []string
	// AnchorContract is the deployed EchoAnchor address.
	AnchorContract common.Address
	// IdentityContract is the deployed EchoIdentity address. Optional: without it the identity
	// methods return ErrNoIdentityContract.
	IdentityContract common.Address
	// Signer signs the gateway's transactions. Its address must be an authorised signer of the
	// contract. Nil gives a read-only client.
	Signer *ecdsa.PrivateKey
	// ReceiptTimeout bounds the wait for a transaction to be included. Blocks are normally
	// seconds apart; straight after every validator has been restarted, the first block has been
	// seen to take more than 30 seconds. Default 90 seconds.
	ReceiptTimeout time.Duration
}

// Client talks to the chain. Safe for concurrent use.
type Client struct {
	cfg Config

	txMu sync.Mutex // serialises transactions so nonces are issued in order

	mu      sync.Mutex // guards conns and chainID
	conns   map[string]*ethclient.Client
	chainID *big.Int
}

// New validates cfg. It does not contact the chain; endpoints are dialled on first use.
func New(cfg Config) (*Client, error) {
	urls := make([]string, 0, len(cfg.RPCURLs))
	for _, u := range cfg.RPCURLs {
		if u = strings.TrimSpace(u); u != "" {
			urls = append(urls, u)
		}
	}
	if len(urls) == 0 {
		return nil, errors.New("besu: no RPC URL configured")
	}
	if cfg.AnchorContract == (common.Address{}) {
		return nil, errors.New("besu: anchor contract address is not configured")
	}
	cfg.RPCURLs = urls
	if cfg.ReceiptTimeout <= 0 {
		cfg.ReceiptTimeout = 90 * time.Second
	}
	return &Client{cfg: cfg, conns: map[string]*ethclient.Client{}}, nil
}

// SignerAddress is the address the gateway writes from, or the zero address if read-only.
func (c *Client) SignerAddress() common.Address {
	if c.cfg.Signer == nil {
		return common.Address{}
	}
	return crypto.PubkeyToAddress(c.cfg.Signer.PublicKey)
}

// AnchorContract is the EchoAnchor address this client writes to.
func (c *Client) AnchorContract() common.Address { return c.cfg.AnchorContract }

// IdentityContract is the EchoIdentity address this client writes to, or zero if not configured.
func (c *Client) IdentityContract() common.Address { return c.cfg.IdentityContract }

// conn returns a connection to the validator that is furthest ahead. A validator that has just
// restarted answers before it has caught up; reading the account's next transaction number or a
// root's status from it would give a stale answer. Ties go to the earlier URL.
func (c *Client) conn(ctx context.Context) (*ethclient.Client, error) {
	type result struct {
		ec     *ethclient.Client
		height uint64
		err    error
	}
	results := make([]result, len(c.cfg.RPCURLs))
	var wg sync.WaitGroup
	for i, url := range c.cfg.RPCURLs {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			ec, err := c.dial(ctx, url)
			if err != nil {
				results[i].err = fmt.Errorf("%s: %w", url, err)
				return
			}
			probe, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			height, err := ec.BlockNumber(probe)
			if err != nil {
				results[i].err = fmt.Errorf("%s: %w", url, err)
				return
			}
			results[i] = result{ec: ec, height: height}
		}(i, url)
	}
	wg.Wait()

	var best *result
	var firstErr error
	for i := range results {
		r := &results[i]
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
			}
			continue
		}
		if best == nil || r.height > best.height {
			best = r
		}
	}
	if best == nil {
		return nil, fmt.Errorf("besu: no RPC endpoint is reachable: %w", firstErr)
	}
	return best.ec, nil
}

// dial returns the cached connection for url, creating it on first use.
func (c *Client) dial(ctx context.Context, url string) (*ethclient.Client, error) {
	c.mu.Lock()
	ec := c.conns[url]
	c.mu.Unlock()
	if ec != nil {
		return ec, nil
	}
	dialled, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if existing := c.conns[url]; existing != nil {
		dialled.Close()
		return existing, nil
	}
	c.conns[url] = dialled
	return dialled, nil
}

var errNoSigner = errors.New("besu: no signing key configured")

func (c *Client) transactOpts(ctx context.Context, ec *ethclient.Client) (*bind.TransactOpts, error) {
	if c.cfg.Signer == nil {
		return nil, errNoSigner
	}
	c.mu.Lock()
	chainID := c.chainID
	c.mu.Unlock()
	if chainID == nil {
		id, err := ec.ChainID(ctx)
		if err != nil {
			return nil, fmt.Errorf("besu: chain id: %w", err)
		}
		chainID = id
		c.mu.Lock()
		c.chainID = id
		c.mu.Unlock()
	}
	opts, err := bind.NewKeyedTransactorWithChainID(c.cfg.Signer, chainID)
	if err != nil {
		return nil, err
	}
	opts.Context = ctx
	// Echo's network has a zero base fee and the gateway is its only writer.
	opts.GasPrice = big.NewInt(0)
	return opts, nil
}

// Receipt is what the chain recorded for one write.
type Receipt struct {
	TxHash      common.Hash
	BlockNumber uint64
	BlockHash   common.Hash
}

// Ref is the reference callers store for a write: its transaction hash.
func (r Receipt) Ref() string { return r.TxHash.Hex() }

// transact sends one call to a contract and waits until it is final.
//
// Each attempt signs one transaction, then submits it to, and looks for it on, whichever validator
// is furthest ahead. Resubmitting the same signed transaction is harmless, and any validator can
// report the result.
//
// Another process may be writing with the same key (a second gateway replica). Both then pick the
// same transaction number and only one transaction can use it. The loser sees its number refused
// as already used while its own transaction has no receipt; it signs again with a fresh number.
// That cannot execute the call twice: a transaction whose number has been used by a different
// transaction can never be included.
func (c *Client) transact(ctx context.Context, contract common.Address, contractABI abi.ABI, method string, args ...interface{}) (Receipt, error) {
	c.txMu.Lock()
	defer c.txMu.Unlock()

	// A call that cannot even be encoded is a programming error; say so before touching the chain.
	if _, err := contractABI.Pack(method, args...); err != nil {
		return Receipt{}, fmt.Errorf("besu: %s: %w", method, err)
	}
	deadline := time.Now().Add(c.cfg.ReceiptTimeout)
	const maxSignings = 8
	var lastErr error
	for signing := 1; signing <= maxSignings; signing++ {
		tx, err := c.build(ctx, contract, contractABI, deadline, method, args...)
		if err != nil {
			return Receipt{}, fmt.Errorf("besu: %s: %w", method, describeRevert(contractABI, err))
		}
		receipt, superseded, err := c.submitAndWait(ctx, tx, deadline)
		switch {
		case err != nil && isNodeRejection(err):
			return Receipt{}, fmt.Errorf("besu: %s: %w", method, describeRevert(contractABI, err))
		case err != nil:
			return Receipt{}, fmt.Errorf("besu: %s: %w", method, err)
		case !superseded:
			if receipt.Status != types.ReceiptStatusSuccessful {
				return Receipt{}, fmt.Errorf("besu: %s: transaction %s reverted", method, tx.Hash().Hex())
			}
			return Receipt{TxHash: tx.Hash(), BlockNumber: receipt.BlockNumber.Uint64(), BlockHash: receipt.BlockHash}, nil
		}
		lastErr = fmt.Errorf("transaction number %d was used by another writer", tx.Nonce())
		// Let the other writer's transaction land before reading the next number.
		select {
		case <-ctx.Done():
			return Receipt{}, fmt.Errorf("besu: %s: %w", method, ctx.Err())
		case <-time.After(time.Duration(signing) * 300 * time.Millisecond):
		}
	}
	return Receipt{}, fmt.Errorf("besu: %s: gave up after %d signings: %v", method, maxSignings, lastErr)
}

// build prepares and signs one transaction on whichever validator is furthest ahead. Building asks
// that validator for the chain id, the next transaction number and a gas estimate; the gas
// estimate is where a refusal by the contract shows.
//
// The validator can go away between being chosen and answering (a restart drops the connection
// mid-request). That says nothing about the call, so it is tried again on whichever validator is
// then furthest ahead, until the deadline. An answer of "no" from a validator is returned at once.
func (c *Client) build(ctx context.Context, contract common.Address, contractABI abi.ABI, deadline time.Time, method string, args ...interface{}) (*types.Transaction, error) {
	for {
		tx, err := c.buildOnce(ctx, contract, contractABI, method, args...)
		if err == nil || isNodeRejection(err) || errors.Is(err, errNoSigner) {
			return tx, err
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("no validator could prepare the transaction within %s (last error: %w)", c.cfg.ReceiptTimeout, err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (c *Client) buildOnce(ctx context.Context, contract common.Address, contractABI abi.ABI, method string, args ...interface{}) (*types.Transaction, error) {
	ec, err := c.conn(ctx)
	if err != nil {
		return nil, err
	}
	opts, err := c.transactOpts(ctx, ec)
	if err != nil {
		return nil, err
	}
	opts.NoSend = true
	return bind.NewBoundContract(contract, contractABI, ec, ec, ec).Transact(opts, method, args...)
}

// submitAndWait submits one signed transaction and waits for its receipt. superseded is true when
// the transaction can no longer be included because its number was used by a different one.
//
// A validator can accept a transaction and restart before passing it on, after which the
// transaction exists nowhere. So it is submitted again every few seconds until it is seen in a
// block; a validator that already has it just says so.
func (c *Client) submitAndWait(ctx context.Context, tx *types.Transaction, deadline time.Time) (receipt *types.Receipt, superseded bool, err error) {
	const resubmitEvery = 4 * time.Second
	accepted := false // some validator has taken this transaction at least once
	var lastSubmit time.Time
	var lastErr error
	for {
		if live, connErr := c.conn(ctx); connErr != nil {
			lastErr = connErr
		} else {
			numberUsed := false
			if time.Since(lastSubmit) >= resubmitEvery {
				lastSubmit = time.Now()
				switch sendErr := live.SendTransaction(ctx, tx); {
				case sendErr == nil, isAlreadyKnown(sendErr):
					accepted = true
				case isNumberUsed(sendErr):
					numberUsed = true
					lastErr = sendErr
				case isNodeRejection(sendErr) && !accepted:
					return nil, false, sendErr // refused outright on first submission
				default:
					lastErr = sendErr // transport failure, or a refusal of a resubmission
				}
			}
			r, rcptErr := live.TransactionReceipt(ctx, tx.Hash())
			switch {
			case rcptErr == nil:
				return r, false, nil
			case !errors.Is(rcptErr, ethereum.NotFound):
				lastErr = rcptErr
			case numberUsed:
				// The same validator says the number is used and has no receipt for this
				// transaction: a different transaction took the number.
				return nil, true, nil
			}
		}
		if time.Now().After(deadline) {
			return nil, false, fmt.Errorf("transaction %s not included within %s (last error: %v)", tx.Hash().Hex(), c.cfg.ReceiptTimeout, lastErr)
		}
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// isNumberUsed reports a node refusing a transaction because its number is taken: either already
// used on the chain, or held by a different pending transaction from the same key.
func isNumberUsed(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "nonce too low") || strings.Contains(msg, "replacement transaction underpriced")
}

// isAlreadyKnown reports a node saying it already holds this exact transaction, which is what a
// resubmission to a second validator looks like.
func isAlreadyKnown(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already known") || strings.Contains(msg, "known transaction")
}

// isNodeRejection distinguishes a node that answered "no" (a JSON-RPC error) from a connection
// that failed. Only the second is worth retrying elsewhere.
func isNodeRejection(err error) bool {
	var rpcErr rpc.Error
	return errors.As(err, &rpcErr)
}

// describeRevert names the contract's custom error when the node reports one.
func describeRevert(contractABI abi.ABI, err error) error {
	var de interface{ ErrorData() interface{} }
	if !errors.As(err, &de) {
		return err
	}
	data, ok := de.ErrorData().(string)
	if !ok || len(data) < 10 {
		return err
	}
	selector := common.FromHex(data)[:4]
	for name, e := range contractABI.Errors {
		if string(e.ID[:4]) == string(selector) {
			return fmt.Errorf("rejected by contract: %s", name)
		}
	}
	return err
}

// Deploy publishes a new contract (ContractAnchor or ContractIdentity) owned by ownerKey, with
// signer as its only authorised writer. Used at genesis and by the local development network; not
// by the running gateway.
func Deploy(ctx context.Context, rpcURL, contract string, ownerKey *ecdsa.PrivateKey, signer common.Address) (common.Address, Receipt, error) {
	var (
		contractABI abi.ABI
		bin         string
	)
	switch contract {
	case ContractAnchor:
		contractABI, bin = anchorABI, anchorBinHex
	case ContractIdentity:
		contractABI, bin = identityABI, identityBinHex
	case ContractReserve:
		contractABI, bin = reserveABI, reserveBinHex
	default:
		return common.Address{}, Receipt{}, fmt.Errorf("besu: unknown contract %q", contract)
	}
	ec, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return common.Address{}, Receipt{}, err
	}
	defer ec.Close()
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return common.Address{}, Receipt{}, err
	}
	opts, err := bind.NewKeyedTransactorWithChainID(ownerKey, chainID)
	if err != nil {
		return common.Address{}, Receipt{}, err
	}
	opts.Context = ctx
	opts.GasPrice = big.NewInt(0)
	addr, tx, _, err := bind.DeployContract(opts, contractABI, common.FromHex(bin), ec, signer)
	if err != nil {
		return common.Address{}, Receipt{}, err
	}
	receipt, err := bind.WaitMined(ctx, ec, tx)
	if err != nil {
		return common.Address{}, Receipt{}, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return common.Address{}, Receipt{}, errors.New("besu: contract deployment reverted")
	}
	return addr, Receipt{TxHash: tx.Hash(), BlockNumber: receipt.BlockNumber.Uint64(), BlockHash: receipt.BlockHash}, nil
}
