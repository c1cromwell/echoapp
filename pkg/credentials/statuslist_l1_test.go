package credentials

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestStatusListMarkRevokedAsyncPublish(t *testing.T) {
	var posts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions" {
			t.Errorf("path %s", r.URL.Path)
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
		StatusListPublishInterval:  time.Hour,
	}
	p := NewStatusListPublisher(cfg)
	p.Start()
	defer p.Stop()

	p.AllocateIndex("cred-1")
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
}

func TestStatusListCoalescedSignalsSingleWorker(t *testing.T) {
	var posts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := MetagraphConfig{
		IdentityL1URL:            srv.URL,
		IssuerDID:                "did:key:zIssuer",
		EnableAnchor:             true,
		Timeout:                  3 * time.Second,
		MaxRetries:               3,
		RetryBackoff:             5 * time.Millisecond,
		StatusListPublishInterval: time.Hour,
	}
	p := NewStatusListPublisher(cfg)
	p.Start()
	defer p.Stop()

	p.AllocateIndex("a")
	p.AllocateIndex("b")
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
