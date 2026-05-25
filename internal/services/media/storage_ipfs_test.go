package media

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// fakeKuboMFS mocks the kubo MFS HTTP API endpoints the IPFS backend uses.
func fakeKuboMFS() (*httptest.Server, *sync.Map) {
	var store sync.Map // path -> []byte
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v0/files/write", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("arg")
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		f, _, err := r.FormFile("data")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer f.Close()
		buf := make([]byte, 0)
		tmp := make([]byte, 1024)
		for {
			n, rerr := f.Read(tmp)
			buf = append(buf, tmp[:n]...)
			if rerr != nil {
				break
			}
		}
		store.Store(path, buf)
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/v0/files/read", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("arg")
		v, ok := store.Load(path)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`file does not exist`))
			return
		}
		_, _ = w.Write(v.([]byte))
	})

	mux.HandleFunc("/api/v0/files/rm", func(w http.ResponseWriter, r *http.Request) {
		store.Delete(r.URL.Query().Get("arg"))
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/v0/files/stat", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := store.Load(r.URL.Query().Get("arg")); !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"Hash":"QmTestCID123","Size":0}`))
	})

	return httptest.NewServer(mux), &store
}

func TestIPFSStorage_RoundTrip(t *testing.T) {
	srv, _ := fakeKuboMFS()
	defer srv.Close()

	st, err := NewIPFSStorage(IPFSConfig{APIURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	data := []byte("encrypted-media-blob")

	cid, err := st.Store(ctx, "media/abc123", data)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	if cid != "QmTestCID123" {
		t.Fatalf("Store should return the CID from files/stat, got %q", cid)
	}
	got, err := st.Retrieve(ctx, "media/abc123")
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("round-trip mismatch: got %q want %q", got, data)
	}

	if err := st.Delete(ctx, "media/abc123"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := st.Retrieve(ctx, "media/abc123"); err == nil {
		t.Fatal("retrieve after delete should error")
	}
}

func TestIPFSStorage_RejectsTraversalKeys(t *testing.T) {
	st, _ := NewIPFSStorage(IPFSConfig{APIURL: "http://unused"})
	ctx := context.Background()
	for _, bad := range []string{"", "../escape", "a/../../etc", "/"} {
		if _, err := st.Store(ctx, bad, []byte("x")); err == nil {
			t.Fatalf("key %q should be rejected", bad)
		}
	}
}

func TestNewIPFSStorage_RequiresURL(t *testing.T) {
	if _, err := NewIPFSStorage(IPFSConfig{}); err == nil {
		t.Fatal("expected error when APIURL is empty")
	}
}

// IPFSStorage must satisfy the StorageBackend interface.
var _ StorageBackend = (*IPFSStorage)(nil)
