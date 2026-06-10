package metagraph

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestTessellationDataHashUsesStructFieldOrder(t *testing.T) {
	u := TrustTierCommitmentUpdate{
		SubjectDID: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Commitment: strings.Repeat("a", 64),
		AnchoredAt: 1700000000000,
	}
	canonical, _, err := tessellationDataHash(u)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"subjectDID":"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK","commitment":"` + strings.Repeat("a", 64) + `","anchoredAt":1700000000000}`
	if string(canonical) != want {
		t.Fatalf("canonical JSON mismatch\n got:  %s\n want: %s", canonical, want)
	}
}

func TestTessellationDataHashDeterministic(t *testing.T) {
	u := TrustTierCommitmentUpdate{
		SubjectDID: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Commitment: strings.Repeat("b", 64),
		AnchoredAt: 1700000000000,
	}
	_, h1, err := tessellationDataHash(u)
	if err != nil {
		t.Fatal(err)
	}
	_, h2, err := tessellationDataHash(u)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 || len(h1) != 64 {
		t.Fatalf("hash: %s", h1)
	}
}

func TestLoadTessellationSigningFromSecp256k1PEM(t *testing.T) {
	pemPath := genTestSecp256k1KeyPEM(t)
	pem, err := os.ReadFile(pemPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadTessellationSigningFromPEM(pem)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Secp256k1Private == nil {
		t.Fatal("missing private key")
	}
	if len(cfg.PublicHex) != 128 {
		t.Fatalf("proof id len: %d", len(cfg.PublicHex))
	}
}

func TestSubmitSignedDataPostsToDataEndpoint(t *testing.T) {
	pemPath := genTestSecp256k1KeyPEM(t)
	pem, err := os.ReadFile(pemPath)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := LoadTessellationSigningFromPEM(pem)
	if err != nil {
		t.Fatal(err)
	}

	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/data" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "bad route", http.StatusBadRequest)
			return
		}
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hash":"abc123"}`))
	}))
	defer srv.Close()

	update := TrustTierCommitmentUpdate{
		SubjectDID: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		Commitment: strings.Repeat("a", 64),
		AnchoredAt: 1700000000000,
	}
	client := NewMetagraphClient(MetagraphConfig{IdentitySigner: &signer})
	hash, err := client.SubmitSignedData(t.Context(), srv.URL, update, signer)
	if err != nil {
		t.Fatal(err)
	}
	if hash != "abc123" {
		t.Fatalf("hash: %q", hash)
	}

	var envelope signedDataUpdate
	if err := json.Unmarshal(gotBody, &envelope); err != nil {
		t.Fatalf("unmarshal body: %v (%s)", err, string(gotBody))
	}
	if !strings.HasPrefix(string(envelope.Value), `{"subjectDID":`) {
		t.Fatalf("value JSON: %s", envelope.Value)
	}
}

func genTestSecp256k1KeyPEM(t *testing.T) string {
	t.Helper()
	path := t.TempDir() + "/key.pem"
	if err := exec.Command("openssl", "ecparam", "-name", "secp256k1", "-genkey", "-noout", "-out", path).Run(); err != nil {
		t.Skipf("openssl unavailable: %v", err)
	}
	return path
}
