package besu

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Reserve is what the EchoReserve contract records: a fixed supply, all of it at one address.
type Reserve struct {
	Reserved    common.Address // the single address holding the supply
	TotalSupply *big.Int       // in base units (8 decimals)
	Balance     *big.Int       // balance of Reserved; equal to TotalSupply by construction
	Decimals    uint8
}

// ReadReserve reads the reserve contract at address from rpcURL. Read-only.
func ReadReserve(ctx context.Context, rpcURL string, address common.Address) (Reserve, error) {
	ec, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return Reserve{}, err
	}
	defer ec.Close()
	contract := bind.NewBoundContract(address, reserveABI, ec, ec, ec)
	call := func(method string, args ...interface{}) (interface{}, error) {
		var out []interface{}
		if err := contract.Call(&bind.CallOpts{Context: ctx}, &out, method, args...); err != nil {
			return nil, fmt.Errorf("besu: read reserve %s: %w", method, err)
		}
		return out[0], nil
	}
	reserved, err := call("reserved")
	if err != nil {
		return Reserve{}, err
	}
	total, err := call("totalSupply")
	if err != nil {
		return Reserve{}, err
	}
	decimals, err := call("DECIMALS")
	if err != nil {
		return Reserve{}, err
	}
	balance, err := call("balanceOf", reserved.(common.Address))
	if err != nil {
		return Reserve{}, err
	}
	return Reserve{Reserved: reserved.(common.Address), TotalSupply: total.(*big.Int), Balance: balance.(*big.Int), Decimals: decimals.(uint8)}, nil
}

// ReserveBalanceOf reads one address's balance in the reserve contract.
func ReserveBalanceOf(ctx context.Context, rpcURL string, address, account common.Address) (*big.Int, error) {
	ec, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, err
	}
	defer ec.Close()
	var out []interface{}
	if err := bind.NewBoundContract(address, reserveABI, ec, ec, ec).Call(&bind.CallOpts{Context: ctx}, &out, "balanceOf", account); err != nil {
		return nil, fmt.Errorf("besu: read reserve balanceOf: %w", err)
	}
	return out[0].(*big.Int), nil
}
