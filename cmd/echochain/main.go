// Command echochain is the operator's tool for Echo's Besu chain.
//
//	echochain address  -key FILE                     print the address of a key (PEM or hex)
//	echochain newkey   -out FILE                     create a key file (hex), refusing to overwrite
//	echochain deploy   -rpc URL -owner FILE -signer ADDRESS
//	                                                 deploy the anchor contract; prints its address
//	echochain validators -rpc URL                    print the current validator addresses
//	echochain prove    -rpc URL -contract ADDRESS -root HEX [-out FILE]
//	                                                 export a proof bundle for an anchored root,
//	                                                 read straight from the chain (no gateway)
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/thechadcromwell/echoapp/internal/besu"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

func main() {
	if len(os.Args) < 2 {
		fail("usage: echochain address|newkey|deploy|validators|prove [flags]")
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	rpcURL := fs.String("rpc", "http://localhost:8545", "Besu JSON-RPC endpoint")
	keyFile := fs.String("key", "", "key file (PEM or hex)")
	ownerFile := fs.String("owner", "", "contract owner key file")
	signer := fs.String("signer", "", "address allowed to write to the contract")
	out := fs.String("out", "", "file to write")
	contract := fs.String("contract", "", "anchor contract address")
	root := fs.String("root", "", "Merkle root (64 hex characters)")
	_ = fs.Parse(os.Args[2:])
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	switch os.Args[1] {
	case "address":
		key, err := besu.LoadKeyFile(*keyFile)
		check(err)
		fmt.Println(crypto.PubkeyToAddress(key.PublicKey).Hex())
	case "newkey":
		if *out == "" {
			fail("newkey needs -out")
		}
		if _, err := os.Stat(*out); err == nil {
			fail(*out + " already exists; refusing to overwrite a key")
		}
		key, err := crypto.GenerateKey()
		check(err)
		check(os.WriteFile(*out, []byte(hex.EncodeToString(crypto.FromECDSA(key))+"\n"), 0o600))
		fmt.Println(crypto.PubkeyToAddress(key.PublicKey).Hex())
	case "deploy":
		if !common.IsHexAddress(*signer) {
			fail("deploy needs -signer ADDRESS")
		}
		owner, err := besu.LoadKeyFile(*ownerFile)
		check(err)
		addr, receipt, err := besu.Deploy(ctx, *rpcURL, owner, common.HexToAddress(*signer))
		check(err)
		fmt.Fprintf(os.Stderr, "deployed EchoAnchor in block %d (tx %s); writer %s\n", receipt.BlockNumber, receipt.TxHash.Hex(), *signer)
		fmt.Println(addr.Hex())
	case "validators":
		client, err := rpc.DialContext(ctx, *rpcURL)
		check(err)
		var validators []common.Address
		check(client.CallContext(ctx, &validators, "qbft_getValidatorsByBlockNumber", "latest"))
		parts := make([]string, len(validators))
		for i, v := range validators {
			parts[i] = v.Hex()
		}
		fmt.Println(strings.Join(parts, ","))
	case "prove":
		if !common.IsHexAddress(*contract) {
			fail("prove needs -contract ADDRESS")
		}
		client, err := besu.New(besu.Config{RPCURLs: []string{*rpcURL}, AnchorContract: common.HexToAddress(*contract)})
		check(err)
		anchor := metagraph.NewBesuAnchor(client)
		status, err := client.RootStatus(ctx, *root)
		check(err)
		evidence, err := anchor.ProveRoot(ctx, *root)
		check(err)
		now := time.Now().UTC()
		body, err := json.MarshalIndent(metagraph.ProofBundle{
			Version:     metagraph.ProofBundleVersion,
			Network:     metagraph.ChainNetworkBesu,
			MerkleRoot:  strings.TrimPrefix(strings.ToLower(*root), "0x"),
			LeafCount:   int(status.LeafCount),
			SubmittedAt: now,
			Finalized:   true,
			FinalizedAt: &now,
			Chain:       &evidence,
		}, "", "  ")
		check(err)
		if *out == "" {
			fmt.Println(string(body))
		} else {
			check(os.WriteFile(*out, append(body, '\n'), 0o644))
		}
	default:
		fail("unknown command " + os.Args[1])
	}
}

func check(err error) {
	if err != nil {
		fail(err.Error())
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "echochain:", msg)
	os.Exit(1)
}
