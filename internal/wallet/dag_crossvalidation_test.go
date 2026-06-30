package wallet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// walletVector mirrors ios/Echo/Resources/wallet-sdk/testvector.json, produced
// by the embedded dag4.js bundle (gen-vector.js). This test is the M0 de-risk
// gate: it proves the iOS signing bundle and the Go backend agree byte-for-byte
// on DAG address derivation and secp256k1 signature verification, WITHOUT a live
// metagraph. If dag4 or the bundle changes incompatibly, this fails.
type walletVector struct {
	Mnemonic   string `json:"mnemonic"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Address    string `json:"address"`
	Message    string `json:"message"`
	Signature  string `json:"signature"`
}

func loadWalletVector(t *testing.T) walletVector {
	t.Helper()
	path := filepath.Join("..", "..", "ios", "Echo", "Resources", "wallet-sdk", "testvector.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("wallet test vector missing (run scripts/build-wallet-sdk.sh): %v", err)
	}
	var v walletVector
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("decode vector: %v", err)
	}
	return v
}

func TestDagAddressMatchesDag4Bundle(t *testing.T) {
	v := loadWalletVector(t)
	got := DagAddressFromPubKey(v.PublicKey)
	if got != v.Address {
		t.Fatalf("Go address %q != dag4 bundle address %q (pubkey %s)", got, v.Address, v.PublicKey)
	}
}

func TestVerifyDag4BundleSignature(t *testing.T) {
	v := loadWalletVector(t)
	if !VerifyDagMessageSignature(v.PublicKey, v.Message, v.Signature) {
		t.Fatalf("Go failed to verify dag4 bundle signature over %q", v.Message)
	}
	// Negative control: a tampered message must NOT verify.
	if VerifyDagMessageSignature(v.PublicKey, v.Message+"x", v.Signature) {
		t.Fatal("tampered message unexpectedly verified")
	}
}

func TestDagAddressDeterministicAndFormat(t *testing.T) {
	// Format guard independent of the vector file.
	pub := "0453f06ad396d382ff1db457e6d2b608c04be2678bbd12207625e581f1e030c4c4c8f9db9094424eea5f7868b846301a8e8857e2c01583714316a043edd192798b"
	a := DagAddressFromPubKey(pub)
	if a == "" || a[:3] != "DAG" || len(a) != 40 {
		t.Fatalf("unexpected address shape: %q (len %d)", a, len(a))
	}
	if DagAddressFromPubKey(pub) != a {
		t.Fatal("derivation not deterministic")
	}
}
