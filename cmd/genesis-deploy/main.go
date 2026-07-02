// genesis-deploy prints or submits the WO-214 genesis token allocation manifest.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/thechadcromwell/echoapp/internal/tokenomics/genesis"
)

func main() {
	genesisFlag := flag.String("genesis-date", "2026-01-01", "genesis date (YYYY-MM-DD)")
	dryRun := flag.Bool("dry-run", true, "print manifest only; do not submit to Currency L1")
	flag.Parse()

	t, err := time.Parse("2006-01-02", *genesisFlag)
	if err != nil {
		log.Fatalf("invalid genesis date: %v", err)
	}

	snap, err := genesis.BuildSnapshot(t)
	if err != nil {
		log.Fatalf("build snapshot: %v", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		log.Fatalf("encode: %v", err)
	}

	if *dryRun {
		fmt.Fprintln(os.Stderr, "dry-run: no Currency L1 submissions")
		return
	}

	txs, err := genesis.FounderTokenLockUpdates(t)
	if err != nil {
		log.Fatalf("founder locks: %v", err)
	}
	fmt.Fprintf(os.Stderr, "would submit %d founder TokenLock transactions\n", len(txs))
}
