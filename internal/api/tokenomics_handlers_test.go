package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/rewards"
	"github.com/thechadcromwell/echoapp/internal/tokenomics"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/quests"
)

func TestEmissionStatus(t *testing.T) {
	svc, err := tokenomics.NewService(tokenomics.Config{
		GenesisDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	h := &TokenomicsHandlers{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/v1/tokens/emission/status", nil)
	w := httptest.NewRecorder()
	h.handleEmissionStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["current_year"] == nil {
		t.Error("expected current_year")
	}
}

func TestQuestClaim_AlreadyClaimed(t *testing.T) {
	store := quests.NewMemStore()
	svc, err := tokenomics.NewService(tokenomics.Config{
		GenesisDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		QuestStore:  store,
	})
	if err != nil {
		t.Fatal(err)
	}
	did := "did:key:z6MkQuestTest"
	_ = svc.Quests.MarkComplete(context.Background(), did, "identity_builder")
	_, _ = svc.Quests.Claim(context.Background(), did, "identity_builder")

	h := &TokenomicsHandlers{Service: svc}
	req := httptest.NewRequest(http.MethodPost, "/v1/gamification/quests/identity_builder/claim", nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, did))
	w := httptest.NewRecorder()
	h.handleQuestClaim(w, req, "identity_builder")
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEmissionSchedule_Alignment(t *testing.T) {
	em := rewards.NewEmissionSchedule(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if em.YearlyBudget(1) == 0 {
		t.Error("expected positive year 1 budget")
	}
}

func TestGamificationStatus_IncludesEchoScoreWhenAuthed(t *testing.T) {
	t.Setenv("ECHO_WALLET_REAL_FUNDS", "")
	svc, err := tokenomics.NewService(tokenomics.Config{
		GenesisDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	db := database.NewMemoryDB()
	if err := db.CreateUser(context.Background(), &database.User{DID: "did:alice", Username: "alice", TrustTier: 3}); err != nil {
		t.Fatal(err)
	}
	h := &TokenomicsHandlers{Service: svc, Trust: db}
	req := httptest.NewRequest(http.MethodGet, "/v1/gamification/status", nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:alice"))
	w := httptest.NewRecorder()
	h.handleGamificationStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	score, ok := resp["echo_score"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected echo_score, got %+v", resp)
	}
	if score["tier"].(float64) != 3 || score["level"] != "member" {
		t.Errorf("echo_score: %+v", score)
	}
	unlock, ok := score["next_unlock"].(map[string]interface{})
	if !ok || unlock["tier"].(float64) != 4 {
		t.Errorf("next_unlock: %+v", score["next_unlock"])
	}
}

func TestGamificationStatus_InterimNonRedeemable(t *testing.T) {
	t.Setenv("ECHO_WALLET_REAL_FUNDS", "")
	svc, err := tokenomics.NewService(tokenomics.Config{
		GenesisDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	h := &TokenomicsHandlers{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/v1/gamification/status", nil)
	w := httptest.NewRecorder()
	h.handleGamificationStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["custody_mode"] != "interim" {
		t.Errorf("custody_mode: got %v want interim", resp["custody_mode"])
	}
	if resp["redeemable"] != false || resp["transferable"] != false {
		t.Errorf("expected non-redeemable/non-transferable, got %+v", resp)
	}
}

func TestActivityPing_IncrementsStreakIdempotentSameDay(t *testing.T) {
	svc, err := tokenomics.NewService(tokenomics.Config{
		GenesisDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	h := &TokenomicsHandlers{Service: svc}
	ctx := context.WithValue(context.Background(), ContextKeyUserID, "did:alice")

	ping := func() map[string]interface{} {
		req := httptest.NewRequest(http.MethodPost, "/v1/gamification/activity/ping", nil).WithContext(ctx)
		w := httptest.NewRecorder()
		h.handleActivityPing(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d body %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		return resp["streak"].(map[string]interface{})
	}

	s1 := ping()
	s2 := ping() // same UTC day → idempotent
	if s1["current_days"].(float64) != 1 || s2["current_days"].(float64) != 1 {
		t.Errorf("same-day pings should stay at 1 day: %v then %v", s1["current_days"], s2["current_days"])
	}
}

func TestLeaderboard_EmptyOK(t *testing.T) {
	svc, err := tokenomics.NewService(tokenomics.Config{
		GenesisDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	h := &TokenomicsHandlers{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/v1/gamification/leaderboard?window=weekly", nil)
	w := httptest.NewRecorder()
	h.handleLeaderboard(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["redeemable"] != false {
		t.Errorf("leaderboard should carry redeemable=false, got %+v", resp["redeemable"])
	}
}

func TestWeeklyPack_GetThenOpen(t *testing.T) {
	svc, err := tokenomics.NewService(tokenomics.Config{
		GenesisDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	h := &TokenomicsHandlers{Service: svc}
	ctx := context.WithValue(context.Background(), ContextKeyUserID, "did:alice")

	get := httptest.NewRequest(http.MethodGet, "/v1/gamification/weekly-pack", nil).WithContext(ctx)
	gw := httptest.NewRecorder()
	h.handleWeeklyPack(gw, get)
	if gw.Code != http.StatusOK {
		t.Fatalf("GET status %d body %s", gw.Code, gw.Body.String())
	}
	var preview map[string]interface{}
	if err := json.Unmarshal(gw.Body.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	pack := preview["pack"].(map[string]interface{})
	if pack["opened"] != false {
		t.Errorf("preview should be unopened: %+v", pack)
	}

	post := httptest.NewRequest(http.MethodPost, "/v1/gamification/weekly-pack", nil).WithContext(ctx)
	pw := httptest.NewRecorder()
	h.handleWeeklyPack(pw, post)
	if pw.Code != http.StatusOK {
		t.Fatalf("POST status %d body %s", pw.Code, pw.Body.String())
	}
	var opened map[string]interface{}
	if err := json.Unmarshal(pw.Body.Bytes(), &opened); err != nil {
		t.Fatal(err)
	}
	if opened["pack"].(map[string]interface{})["opened"] != true {
		t.Errorf("open should set opened: %+v", opened)
	}
}
