package metagraph

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// A Merkle root signed by a key that is not on the chain's DATA_AUTHORIZED_SIGNERS
// list must never be finalized. Runs only against a live chain that enforces the
// list (`make sovereign-up`):
//
//	ECHO_SIGNER_TEST_DATA_L1=http://localhost:9400 \
//	ECHO_SIGNER_TEST_METAGRAPH_L0=http://localhost:9200 go test ./internal/metagraph -run SignerAllowlist -v
func TestSignerAllowlistRejectsUnknownSigner(t *testing.T) {
	dataL1 := os.Getenv("ECHO_SIGNER_TEST_DATA_L1")
	l0 := os.Getenv("ECHO_SIGNER_TEST_METAGRAPH_L0")
	if dataL1 == "" || l0 == "" {
		t.Skip("set ECHO_SIGNER_TEST_DATA_L1 and ECHO_SIGNER_TEST_METAGRAPH_L0 to run against a live chain")
	}

	priv, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	proofID, err := tessellationProofIDFromSecp256k1(priv.PubKey())
	if err != nil {
		t.Fatal(err)
	}
	stranger := IdentitySigningConfig{Secp256k1Private: priv, PublicHex: proofID}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		t.Fatal(err)
	}
	root := hex.EncodeToString(raw)

	client := NewMetagraphClient(MetagraphConfig{})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Data L1 checks only that the signature is valid, so it may accept the update
	// into its pool. The allowlist is enforced where it counts: Metagraph L0.
	if _, err := client.SubmitSignedData(ctx, dataL1, DataL1MerkleRootUpdate{Root: root, LeafCount: 1}, stranger); err != nil {
		t.Logf("Data L1 refused the update outright: %v", err)
	}

	wait := 90 * time.Second
	if v := os.Getenv("ECHO_SIGNER_TEST_WAIT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			wait = d
		}
	}
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("%s/data-application/merkle-roots/%s", strings.TrimRight(l0, "/"), root))
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if strings.Contains(string(body), `"finalized":true`) {
				t.Fatalf("root %s signed by an unlisted key was finalized: %s", root, body)
			}
		}
		time.Sleep(5 * time.Second)
	}
}
