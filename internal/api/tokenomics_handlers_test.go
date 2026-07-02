package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
