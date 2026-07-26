package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func syncRouter(db database.DB) *Router {
	return &Router{V3: &V3Handlers{DB: db}}
}

type pullResp struct {
	Entries    []database.SyncEntry `json:"entries"`
	NextCursor int64                `json:"next_cursor"`
}

func pushEntry(t *testing.T, mux http.Handler, did, deviceID, body string) int64 {
	t.Helper()
	rec := postJSON(mux, "/v3/sync/push", did,
		map[string]any{"target_device_id": deviceID, "entry_type": "history", "ciphertext": []byte(body)})
	if rec.Code != http.StatusOK {
		t.Fatalf("push want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var r struct {
		Seq int64 `json:"seq"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	return r.Seq
}

func pull(t *testing.T, mux http.Handler, did, deviceID string, after int64) pullResp {
	t.Helper()
	url := fmt.Sprintf("/v3/sync/pull?device_id=%s&after=%d", deviceID, after)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, withDID(httptest.NewRequest(http.MethodGet, url, nil), did))
	if rec.Code != http.StatusOK {
		t.Fatalf("pull want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var r pullResp
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	return r
}

// TestSync_PushPullCursor: entries are addressed per device, ordered by seq, and
// a cursor pulls only newer entries (WO-CA3).
func TestSync_PushPullCursor(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(syncRouter(db))

	s1 := pushEntry(t, mux, "did:alice", "ipad", "h1")
	s2 := pushEntry(t, mux, "did:alice", "ipad", "h2")
	if s1 != 1 || s2 != 2 {
		t.Fatalf("seq should be monotonic 1,2; got %d,%d", s1, s2)
	}

	// Fresh device pulls all history from cursor 0.
	r := pull(t, mux, "did:alice", "ipad", 0)
	if len(r.Entries) != 2 || string(r.Entries[0].Ciphertext) != "h1" || r.NextCursor != 2 {
		t.Fatalf("expected 2 ordered entries + cursor 2, got %+v", r)
	}

	// Incremental pull after the cursor returns nothing new...
	if r2 := pull(t, mux, "did:alice", "ipad", 2); len(r2.Entries) != 0 {
		t.Fatalf("expected no new entries after cursor 2, got %d", len(r2.Entries))
	}
	// ...until another entry is appended.
	pushEntry(t, mux, "did:alice", "ipad", "h3")
	if r3 := pull(t, mux, "did:alice", "ipad", 2); len(r3.Entries) != 1 || string(r3.Entries[0].Ciphertext) != "h3" {
		t.Fatalf("expected only h3 after cursor 2, got %+v", r3)
	}
}

// TestSync_PerDeviceIsolation: each device has its own stream/cursor.
func TestSync_PerDeviceIsolation(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(syncRouter(db))
	pushEntry(t, mux, "did:alice", "ipad", "for-ipad")
	pushEntry(t, mux, "did:alice", "watch", "for-watch")

	ipad := pull(t, mux, "did:alice", "ipad", 0)
	if len(ipad.Entries) != 1 || string(ipad.Entries[0].Ciphertext) != "for-ipad" {
		t.Fatalf("ipad stream wrong: %+v", ipad.Entries)
	}
	watch := pull(t, mux, "did:alice", "watch", 0)
	if len(watch.Entries) != 1 || string(watch.Entries[0].Ciphertext) != "for-watch" {
		t.Fatalf("watch stream wrong: %+v", watch.Entries)
	}
}

// TestSync_ControllerScoping: a different account cannot see another's stream.
func TestSync_ControllerScoping(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(syncRouter(db))
	pushEntry(t, mux, "did:alice", "ipad", "secret")

	// Bob pulls device id "ipad" under his own DID — empty (scoped by controller DID).
	if r := pull(t, mux, "did:bob", "ipad", 0); len(r.Entries) != 0 {
		t.Fatalf("cross-account pull must be empty, got %+v", r.Entries)
	}
}

// TestSync_RevokeStopsSyncAndPurges: revoking a device closes its stream — push and
// pull are rejected and entries are gone (the "revoke stops sync" gate).
func TestSync_RevokeStopsSyncAndPurges(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(syncRouter(db))
	pushEntry(t, mux, "did:alice", "ipad", "h1")

	if rec := postJSON(mux, "/v3/sync/revoke", "did:alice",
		map[string]any{"target_device_id": "ipad"}); rec.Code != http.StatusOK {
		t.Fatalf("revoke want 200, got %d", rec.Code)
	}

	// Pull is now forbidden.
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, withDID(httptest.NewRequest(http.MethodGet, "/v3/sync/pull?device_id=ipad&after=0", nil), "did:alice"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("pull after revoke want 403, got %d", rec.Code)
	}
	// Push is now forbidden.
	if prec := postJSON(mux, "/v3/sync/push", "did:alice",
		map[string]any{"target_device_id": "ipad", "ciphertext": []byte("h2")}); prec.Code != http.StatusForbidden {
		t.Fatalf("push after revoke want 403, got %d", prec.Code)
	}
}

// TestSync_Head returns the latest seq for the stream.
func TestSync_Head(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(syncRouter(db))
	pushEntry(t, mux, "did:alice", "ipad", "h1")
	pushEntry(t, mux, "did:alice", "ipad", "h2")

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, withDID(httptest.NewRequest(http.MethodGet, "/v3/sync/head?device_id=ipad", nil), "did:alice"))
	var r struct {
		Seq int64 `json:"seq"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	if r.Seq != 2 {
		t.Fatalf("expected head seq 2, got %d", r.Seq)
	}
}

// TestSync_AckTrimsAppliedEntries: client ack deletes seq <= through_seq (Wave S1).
func TestSync_AckTrimsAppliedEntries(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(syncRouter(db))
	pushEntry(t, mux, "did:alice", "ipad", "h1")
	pushEntry(t, mux, "did:alice", "ipad", "h2")
	pushEntry(t, mux, "did:alice", "ipad", "h3")

	if rec := postJSON(mux, "/v3/sync/ack", "did:alice",
		map[string]any{"device_id": "ipad", "through_seq": 2}); rec.Code != http.StatusOK {
		t.Fatalf("ack want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	r := pull(t, mux, "did:alice", "ipad", 0)
	if len(r.Entries) != 1 || string(r.Entries[0].Ciphertext) != "h3" {
		t.Fatalf("expected only h3 after ack through 2, got %+v", r.Entries)
	}
}
