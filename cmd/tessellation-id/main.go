// tessellation-id prints the Tessellation 4.x data-application proof id for a secp256k1 PEM key.
//
// Usage:
//
//	go run ./cmd/tessellation-id -pem path/to/secp256k1.pem
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

func main() {
	pemPath := flag.String("pem", "", "path to secp256k1 private key PEM")
	flag.Parse()
	if *pemPath == "" {
		fmt.Fprintln(os.Stderr, "usage: tessellation-id -pem <path>")
		os.Exit(1)
	}
	pemBytes, err := os.ReadFile(*pemPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read PEM: %v\n", err)
		os.Exit(1)
	}
	cfg, err := metagraph.LoadTessellationSigningFromPEM(pemBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("proof_id=%s\n", cfg.PublicHex)
}
