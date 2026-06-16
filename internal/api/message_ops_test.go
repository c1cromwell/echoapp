package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/comply"
)

func postJSON(mux http.Handler, path, did string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := withDID(httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b)), did)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// --- Edit (WO-25, hybrid) ---

func TestMessageEdit_NonRetained_NoServerHistory_FansOut(t *testing.T) {
	db := database.NewMemoryDB()
	pub := &fakeSignalPublisher{}
	mux := v3Mux(signalsRouter(db, pub))
	enqueue(t, db, "m1", "c1", "did:alice", "did:bob")

	rec := postJSON(mux, "/v3/messages/m1/edit", "did:alice",
		map[string]any{"conversation_id": "c1", "ciphertext": []byte("new-ct")})
	if rec.Code != http.StatusOK {
		t.Fatalf("edit want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Retained bool `json:"retained"`
		Version  int  `json:"version"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Retained || resp.Version != 0 {
		t.Fatalf("non-retained edit should not version server-side, got %+v", resp)
	}
	// No server-side history.
	hist, _ := db.GetEditHistory(context.Background(), "m1")
	if len(hist) != 0 {
		t.Fatalf("expected no server history for non-retained convo, got %d", len(hist))
	}
	// Edit fanned out to peer.
	sent := pub.all()
	if len(sent) != 1 || sent[0].to != "did:bob" || sent[0].msg.Type != "edit" {
		t.Fatalf("expected one edit signal to did:bob, got %+v", sent)
	}
	var sig EditSignal
	_ = json.Unmarshal(sent[0].msg.Payload, &sig)
	if string(sig.Ciphertext) != "new-ct" || sig.MessageID != "m1" {
		t.Fatalf("edit signal payload wrong: %+v", sig)
	}
}

func TestMessageEdit_Retained_StoresImmutableVersions(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(signalsRouter(db, &fakeSignalPublisher{}))
	enqueue(t, db, "m1", "c1", "did:alice", "did:bob")

	// Flag the conversation for retention (Comply / litigation hold).
	if rec := postJSON(mux, "/v3/conversations/c1/retention", "did:admin",
		map[string]any{"retained": true}); rec.Code != http.StatusOK {
		t.Fatalf("retention set want 200, got %d", rec.Code)
	}

	for i := 1; i <= 2; i++ {
		rec := postJSON(mux, "/v3/messages/m1/edit", "did:alice",
			map[string]any{"conversation_id": "c1", "ciphertext": []byte(fmt.Sprintf("v%d", i))})
		var resp struct {
			Retained bool `json:"retained"`
			Version  int  `json:"version"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if !resp.Retained || resp.Version != i {
			t.Fatalf("edit %d: expected retained version %d, got %+v", i, i, resp)
		}
	}

	// History endpoint returns both immutable versions.
	hrec := httptest.NewRecorder()
	mux.ServeHTTP(hrec, httptest.NewRequest(http.MethodGet, "/v3/messages/m1/history", nil))
	if hrec.Code != http.StatusOK {
		t.Fatalf("history want 200, got %d", hrec.Code)
	}
	var hresp struct {
		Versions []database.MessageEdit `json:"versions"`
	}
	_ = json.Unmarshal(hrec.Body.Bytes(), &hresp)
	if len(hresp.Versions) != 2 || hresp.Versions[0].Version != 1 || hresp.Versions[1].Version != 2 {
		t.Fatalf("expected 2 ordered versions, got %+v", hresp.Versions)
	}
}

// --- Delete (WO-84) ---

func TestMessageDelete_TombstonesAndFansOut(t *testing.T) {
	db := database.NewMemoryDB()
	pub := &fakeSignalPublisher{}
	mux := v3Mux(signalsRouter(db, pub))
	enqueue(t, db, "m1", "c1", "did:alice", "did:bob")

	rec := postJSON(mux, "/v3/messages/m1/delete", "did:alice",
		map[string]any{"conversation_id": "c1"})
	if rec.Code != http.StatusOK {
		t.Fatalf("delete want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if deleted, _ := db.IsMessageDeleted(context.Background(), "m1"); !deleted {
		t.Fatal("expected message tombstoned")
	}
	sent := pub.all()
	if len(sent) != 1 || sent[0].msg.Type != "delete" || sent[0].to != "did:bob" {
		t.Fatalf("expected one delete signal to did:bob, got %+v", sent)
	}
}

func TestMessageDelete_RetainedPreservesHistory(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(signalsRouter(db, &fakeSignalPublisher{}))
	enqueue(t, db, "m1", "c1", "did:alice", "did:bob")
	_ = db.SetConversationRetention(context.Background(), "c1", true)
	_, _ = db.AppendEditVersion(context.Background(), &database.MessageEdit{MessageID: "m1", ConversationID: "c1", EditorDID: "did:alice", Ciphertext: []byte("v1")})

	if rec := postJSON(mux, "/v3/messages/m1/delete", "did:alice",
		map[string]any{"conversation_id": "c1"}); rec.Code != http.StatusOK {
		t.Fatalf("delete want 200, got %d", rec.Code)
	}
	hist, _ := db.GetEditHistory(context.Background(), "m1")
	if len(hist) != 1 {
		t.Fatalf("retained delete must preserve edit history, got %d", len(hist))
	}
}

// --- Pins (WO-59) ---

func TestMessagePin_LimitAndListAndUnpin(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(signalsRouter(db, &fakeSignalPublisher{}))

	// Pin 5 distinct messages -> OK.
	for i := 0; i < database.MaxPinnedPerConversation; i++ {
		mid := fmt.Sprintf("m%d", i)
		if rec := postJSON(mux, "/v3/messages/"+mid+"/pin", "did:alice",
			map[string]any{"conversation_id": "c1"}); rec.Code != http.StatusOK {
			t.Fatalf("pin %s want 200, got %d", mid, rec.Code)
		}
	}
	// 6th -> 409.
	if rec := postJSON(mux, "/v3/messages/m5/pin", "did:alice",
		map[string]any{"conversation_id": "c1"}); rec.Code != http.StatusConflict {
		t.Fatalf("6th pin want 409, got %d", rec.Code)
	}

	// List returns 5.
	lrec := httptest.NewRecorder()
	mux.ServeHTTP(lrec, httptest.NewRequest(http.MethodGet, "/v3/conversations/c1/pins", nil))
	var lresp struct {
		Pins []database.PinnedMessage `json:"pins"`
	}
	_ = json.Unmarshal(lrec.Body.Bytes(), &lresp)
	if len(lresp.Pins) != database.MaxPinnedPerConversation {
		t.Fatalf("expected %d pins, got %d", database.MaxPinnedPerConversation, len(lresp.Pins))
	}

	// Unpin one frees a slot.
	if rec := postJSON(mux, "/v3/messages/m0/unpin", "did:alice",
		map[string]any{"conversation_id": "c1"}); rec.Code != http.StatusOK {
		t.Fatalf("unpin want 200, got %d", rec.Code)
	}
	if rec := postJSON(mux, "/v3/messages/m5/pin", "did:alice",
		map[string]any{"conversation_id": "c1"}); rec.Code != http.StatusOK {
		t.Fatalf("pin after unpin want 200, got %d", rec.Code)
	}
}

// --- Disappearing messages config ---

func TestDisappearing_SetGetAndFanOut(t *testing.T) {
	db := database.NewMemoryDB()
	pub := &fakeSignalPublisher{}
	mux := v3Mux(signalsRouter(db, pub))

	// Set a 24h TTL and fan out to the peer.
	rec := postJSON(mux, "/v3/conversations/c1/disappearing", "did:alice",
		map[string]any{"ttl_seconds": 86400, "peer_did": "did:bob"})
	if rec.Code != http.StatusOK {
		t.Fatalf("set disappearing want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ttl, _ := db.GetDisappearingTTL(context.Background(), "c1"); ttl != 86400 {
		t.Fatalf("expected stored ttl 86400, got %d", ttl)
	}
	sent := pub.all()
	if len(sent) != 1 || sent[0].to != "did:bob" || sent[0].msg.Type != "disappearing_config" {
		t.Fatalf("expected disappearing_config signal to did:bob, got %+v", sent)
	}

	// GET reflects it.
	grec := httptest.NewRecorder()
	mux.ServeHTTP(grec, httptest.NewRequest(http.MethodGet, "/v3/conversations/c1/disappearing", nil))
	var gresp struct {
		TTLSeconds int `json:"ttl_seconds"`
	}
	_ = json.Unmarshal(grec.Body.Bytes(), &gresp)
	if gresp.TTLSeconds != 86400 {
		t.Fatalf("GET expected 86400, got %d", gresp.TTLSeconds)
	}

	// Turning it off (0) clears it.
	_ = postJSON(mux, "/v3/conversations/c1/disappearing", "did:alice", map[string]any{"ttl_seconds": 0})
	if ttl, _ := db.GetDisappearingTTL(context.Background(), "c1"); ttl != 0 {
		t.Fatalf("expected ttl cleared, got %d", ttl)
	}
}

func TestMessageDelete_BlockedByComplyPolicy(t *testing.T) {
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, "tok")
	mux := v3Mux(complySignalsRouter(db, &fakeSignalPublisher{}, svc))
	enqueue(t, db, "m1", "c1", "did:alice", "did:bob")

	_, err := svc.CreateRetentionPolicy(context.Background(), comply.CreatePolicyInput{
		OrgDID:         "did:org:acme",
		PolicyType:     database.PolicyLitigationHold,
		ConversationID: "c1",
		CreatedByDID:   "did:admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	rec := postJSON(mux, "/v3/messages/m1/delete", "did:alice", map[string]any{"conversation_id": "c1"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("delete under hold want 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisappearing_BlockedByComplyPermanentPolicy(t *testing.T) {
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, "tok")
	mux := v3Mux(complySignalsRouter(db, &fakeSignalPublisher{}, svc))

	_, err := svc.CreateRetentionPolicy(context.Background(), comply.CreatePolicyInput{
		OrgDID:         "did:org:acme",
		PolicyType:     database.PolicyPermanent,
		ConversationID: "c1",
		CreatedByDID:   "did:admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	rec := postJSON(mux, "/v3/conversations/c1/disappearing", "did:alice",
		map[string]any{"ttl_seconds": 3600, "peer_did": "did:bob"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("disappearing under permanent policy want 403, got %d: %s", rec.Code, rec.Body.String())
	}
}
