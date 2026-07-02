package media

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestPinataStorage_StoreRetrieve(t *testing.T) {
	var stored bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pinning/pinJSONToIPFS" {
			stored = true
			_, _ = w.Write([]byte(`{"IpfsHash":"QmTestCID"}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/ipfs/") {
			_, _ = w.Write([]byte("chunk-bytes"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	_ = os.Setenv("PINATA_API_KEY", "k")
	_ = os.Setenv("PINATA_API_SECRET", "s")
	_ = os.Setenv("PINATA_API_BASE", srv.URL)
	_ = os.Setenv("PINATA_GATEWAY", srv.URL)

	p, err := NewPinataStorage()
	if err != nil {
		t.Fatal(err)
	}

	cid, err := p.Store(context.Background(), "file-1/chunk-0", []byte("secret"))
	if err != nil || cid != "QmTestCID" || !stored {
		t.Fatalf("store failed: cid=%s err=%v stored=%v", cid, err, stored)
	}
	data, err := p.Retrieve(context.Background(), "file-1/chunk-0")
	if err != nil || string(data) != "chunk-bytes" {
		t.Fatalf("retrieve: %q %v", data, err)
	}
}
