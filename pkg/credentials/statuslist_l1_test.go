package credentials

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestStatusListMarkRevokedAsyncPublish(t *testing.T) {
	var posts atomic.Int32
	var lastBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions" {
			t.Errorf("path %s", r.URL.Path)
		}
		var err error
		lastBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		posts.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := MetagraphConfig{
		IdentityL1URL:             srv.URL,
		IssuerDID:                 "did:key:z6MkTestIssuer",
		EnableAnchor:              true,
		Timeout:                   3 * time.Second,
		MaxRetries:                3,
		RetryBackoff:              5 * time.Millisecond,
		StatusListPublishInterval: time.Hour,
	}
	p := NewStatusListPublisher(cfg)
	p.Start()
	defer p.Stop()

	if _, err := p.AllocateIndex("cred-1"); err != nil {
		t.Fatal(err)
	}
	if !p.MarkRevoked("cred-1") {
		t.Fatal("expected MarkRevoked true")
	}
	if !p.PublishPending() {
		t.Fatal("expected pending before L1 ack")
	}

	deadline := time.After(2 * time.Second)
	for posts.Load() < 1 {
		select {
		case <-deadline:
			t.Fatalf("no POST observed, posts=%d", posts.Load())
		case <-time.After(5 * time.Millisecond):
		}
	}
	if p.PublishPending() {
		t.Fatal("expected dirty cleared after successful publish")
	}
	const wantHexLen = 32768
	var payload statusList2021BatchWire
	if err := json.Unmarshal(lastBody, &payload); err != nil {
		t.Fatalf("decode posted JSON: %v", err)
	}
	if len(payload.BitVector) != wantHexLen {
		t.Fatalf("bitVector len = %d, want %d", len(payload.BitVector), wantHexLen)
	}
	for _, c := range payload.BitVector {
		if c >= '0' && c <= '9' || c >= 'a' && c <= 'f' {
			continue
		}
		t.Fatalf("bitVector must be lowercase hex, bad rune %q in %s…", c, payload.BitVector[:16])
	}
}

func TestStatusListCoalescedSignalsSingleWorker(t *testing.T) {
	var posts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := MetagraphConfig{
		IdentityL1URL:             srv.URL,
		IssuerDID:                 "did:key:zIssuer",
		EnableAnchor:              true,
		Timeout:                   3 * time.Second,
		MaxRetries:                3,
		RetryBackoff:              5 * time.Millisecond,
		StatusListPublishInterval: time.Hour,
	}
	p := NewStatusListPublisher(cfg)
	p.Start()
	defer p.Stop()

	if _, err := p.AllocateIndex("a"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AllocateIndex("b"); err != nil {
		t.Fatal(err)
	}
	p.MarkRevoked("a")
	p.MarkRevoked("b")

	deadline := time.After(2 * time.Second)
	for posts.Load() < 1 {
		select {
		case <-deadline:
			t.Fatalf("timeout, posts=%d", posts.Load())
		case <-time.After(5 * time.Millisecond):
		}
	}
}
