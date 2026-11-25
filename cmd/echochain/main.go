// Command echochain is the operator's tool for Echo's Besu chain.
//
//	echochain address  -key FILE                     print the address of a key (PEM or hex)
//	echochain newkey   -out FILE                     create a key file (hex), refusing to overwrite
//	echochain deploy   -rpc URL -owner FILE -signer ADDRESS [-name EchoAnchor|EchoIdentity]
//	                                                 deploy a contract; prints its address
//	echochain deploy   -rpc URL -owner FILE -name EchoReserve -reserved ADDRESS
//	                                                 deploy the reserve: the whole fixed supply at
//	                                                 one address, with no way to move or add to it
//	echochain reserve  -rpc URL -contract ADDRESS    print what a reserve contract holds
//	echochain validators -rpc URL                    print the current validator addresses
//	echochain trust-tier -rpc URL -identity ADDRESS -commitment HEX
//	                                                 print the block a trust-tier commitment was
//	                                                 anchored in; exit 1 if it is not on the chain
//	echochain username -rpc URL -identity ADDRESS -username NAME -did DID
//	                                                 report who owns a username; exit 1 unless DID does
//	echochain checkpoint -rpc URL [-out FILE]        write down the newest block with the validators'
//	                                                 signatures, for publishing outside Echo
//	echochain checkpoint-verify -in FILE -validators A,B,C,D [-rpc URL]
//	                                                 check a checkpoint's signatures offline; with
//	                                                 -rpc, also that the chain still holds that
//	                                                 block (exit 3 if history was rewritten)
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
		fail("usage: echochain address|newkey|deploy|validators|trust-tier|username|reserve|checkpoint|checkpoint-verify|prove [flags]")
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	rpcURL := fs.String("rpc", "http://localhost:8545", "Besu JSON-RPC endpoint")
	keyFile := fs.String("key", "", "key file (PEM or hex)")
	ownerFile := fs.String("owner", "", "contract owner key file")
	signer := fs.String("signer", "", "address allowed to write to the contract")
	out := fs.String("out", "", "file to write")
	name := fs.String("name", besu.ContractAnchor, "contract to deploy: EchoAnchor, EchoIdentity or EchoReserve")
	reserved := fs.String("reserved", "", "for EchoReserve: the one address that holds the supply")
	contract := fs.String("contract", "", "anchor contract address")
	root := fs.String("root", "", "Merkle root (64 hex characters)")
	identity := fs.String("identity", "", "identity contract address")
	commitment := fs.String("commitment", "", "trust-tier commitment (64 hex characters)")
	username := fs.String("username", "", "username")
	did := fs.String("did", "", "DID")
	in := fs.String("in", "", "file to read")
	validatorList := fs.String("validators", "", "comma-separated validator addresses you trust")
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
	case "checkpoint":
		cp, err := besu.TakeCheckpoint(ctx, *rpcURL)
		check(err)
		data, err := json.MarshalIndent(cp, "", "  ")
		check(err)
		if *out != "" {
			check(os.WriteFile(*out, append(data, '\n'), 0o644))
			fmt.Fprintf(os.Stderr, "checkpoint of block %d (%s) written to %s\n", cp.Number, cp.Hash.Hex(), *out)
		} else {
			fmt.Println(string(data))
		}
	case "checkpoint-verify":
		if *in == "" || *validatorList == "" {
			fail("checkpoint-verify needs -in FILE and -validators A,B,C,D")
		}
		data, err := os.ReadFile(*in)
		check(err)
		var cp besu.Checkpoint
		check(json.Unmarshal(data, &cp))
		var trusted []common.Address
		for _, v := range strings.Split(*validatorList, ",") {
			if v = strings.TrimSpace(v); v != "" {
				if !common.IsHexAddress(v) {
					fail("not an address: " + v)
				}
				trusted = append(trusted, common.HexToAddress(v))
			}
		}
		signers, err := besu.VerifyCheckpoint(&cp, trusted)
		check(err)
		fmt.Printf("block %d (%s) at %s: signed by %d of %d trusted validators\n", cp.Number, cp.Hash.Hex(), cp.Time.Format(time.RFC3339), len(signers), len(trusted))
		// -rpc has a default, so only compare with a chain when it was given on purpose.
		rpcGiven := false
		fs.Visit(func(f *flag.Flag) { rpcGiven = rpcGiven || f.Name == "rpc" })
		if rpcGiven {
			if err := besu.CompareWithChain(ctx, &cp, *rpcURL); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(3)
			}
			fmt.Println("the chain still holds this block")
		}
	case "reserve":
		if !common.IsHexAddress(*contract) {
			fail("reserve needs -contract ADDRESS")
		}
		r, err := besu.ReadReserve(ctx, *rpcURL, common.HexToAddress(*contract))
		check(err)
		fmt.Printf("reserved=%s\ntotal_supply=%s\nbalance_of_reserved=%s\ndecimals=%d\n", r.Reserved.Hex(), r.TotalSupply, r.Balance, r.Decimals)
	case "deploy":
		if *name == besu.ContractReserve {
			if !common.IsHexAddress(*reserved) {
				fail("deploying EchoReserve needs -reserved ADDRESS")
			}
			*signer = *reserved // the reserve's only constructor argument
		}
		if !common.IsHexAddress(*signer) {
			fail("deploy needs -signer ADDRESS")
		}
		owner, err := besu.LoadKeyFile(*ownerFile)
		check(err)
		addr, receipt, err := besu.Deploy(ctx, *rpcURL, *name, owner, common.HexToAddress(*signer))
		check(err)
		role := "writer"
		if *name == besu.ContractReserve {
			role = "reserved address"
		}
		fmt.Fprintf(os.Stderr, "deployed %s in block %d (tx %s); %s %s\n", *name, receipt.BlockNumber, receipt.TxHash.Hex(), role, *signer)
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
	case "trust-tier", "username":
		if !common.IsHexAddress(*identity) {
			fail(os.Args[1] + " needs -identity ADDRESS")
		}
		// Read-only: the anchor contract address is not used, but the client requires one.
		client, err := besu.New(besu.Config{RPCURLs: []string{*rpcURL}, AnchorContract: common.HexToAddress(*identity), IdentityContract: common.HexToAddress(*identity)})
		check(err)
		if os.Args[1] == "trust-tier" {
			block, err := client.TrustTierAnchoredAt(ctx, *commitment)
			check(err)
			if block == 0 {
				fail("commitment is not anchored")
			}
			fmt.Println(block)
			return
		}
		registered, owned, err := client.UsernameOwnedBy(ctx, *username, *did)
		check(err)
		switch {
		case !registered:
			fail("username is not registered")
		case !owned:
			fail("username is registered to a different DID")
		}
		fmt.Println("owned")
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
