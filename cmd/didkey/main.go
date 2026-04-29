// didkey is a tiny CLI that derives a canonical did:key (per W3C did-key
// method, P-256) from a PEM-encoded public key. It is intended to be called
// by scripts/validate-phase1.sh so that the bash harness can use the same
// canonical derivation logic the backend uses to validate registrations.
//
// Usage:
//
//	go run ./cmd/didkey -pem path/to/public.pem
//	go run ./cmd/didkey -pem -                # read PEM from stdin
//
// Output (stdout):
//
//	did=did:key:z…
//	public_key_hex=04…
//
// Exit codes:
//
//	0 — success
//	1 — usage / IO / parse error
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

func main() {
	pemPath := flag.String("pem", "", "path to a PEM-encoded SubjectPublicKeyInfo, or '-' for stdin")
	flag.Parse()

	if *pemPath == "" {
		fmt.Fprintln(os.Stderr, "usage: didkey -pem <path|->")
		os.Exit(1)
	}

	var pemBytes []byte
	var err error
	if *pemPath == "-" {
		pemBytes, err = io.ReadAll(os.Stdin)
	} else {
		pemBytes, err = os.ReadFile(*pemPath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "read PEM: %v\n", err)
		os.Exit(1)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		fmt.Fprintln(os.Stderr, "no PEM block found")
		os.Exit(1)
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse PKIX public key: %v\n", err)
		os.Exit(1)
	}
	pub, ok := pubAny.(*ecdsa.PublicKey)
	if !ok {
		fmt.Fprintln(os.Stderr, "PEM does not contain an ECDSA public key")
		os.Exit(1)
	}
	if pub.Curve != elliptic.P256() {
		fmt.Fprintf(os.Stderr, "public key curve must be P-256, got %s\n", pub.Curve.Params().Name)
		os.Exit(1)
	}

	did, err := didkey.Derive(pub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "derive did:key: %v\n", err)
		os.Exit(1)
	}

	uncompressed := elliptic.Marshal(elliptic.P256(), pub.X, pub.Y) // 0x04 || X || Y
	fmt.Printf("did=%s\n", did)
	fmt.Printf("public_key_hex=%s\n", hex.EncodeToString(uncompressed))
}
