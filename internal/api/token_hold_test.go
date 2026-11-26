package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/tokenomics/quests"
)

// tokenWords must not appear in anything the gamification API says: there is no token.
var tokenWords = []string{"echo earned", "reward_echo", "redeemable", "transferable", "custody", "staking", "stake", "wallet", "token", "datum"}

func assertNoTokenWords(t *testing.T, what, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, w := range tokenWords {
		if strings.Contains(lower, w) {
			t.Errorf("%s mentions %q but there is no token: %s", what, w, body)
		}
	}
}

func TestTokenHold_TokenRoutesDoNotExist(t *testing.T) {
	rt := NewRouter(nil)
	rt.Tokenomics = &TokenomicsHandlers{}
	for _, path := range []string{
		"/v1/tokens/emission/status", "/v1/tokens/vesting", "/v1/tokens/governance/voting-power",
		"/v1/tokens/vip/allow-spend", "/v1/admin/founder-revocation/initiate",
	} {
		rec := httptest.NewRecorder()
		rt.handleTokenomicsSubroutes(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404: there is no token", path, rec.Code)
		}
	}
}

func TestTokenHold_WalletAndRewardRoutesNotRegistered(t *testing.T) {
	h := &V3Handlers{}
	mux := http.NewServeMux()
	h.RegisterV3Routes(mux)
	for _, path := range []string{"/v3/wallet", "/v3/wallet/stake", "/v3/wallet/claim", "/v3/rewards/claim", "/v3/rewards/daily-stats"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404: there is no token", path, rec.Code)
		}
	}
}

func TestTokenHold_GamificationStatusSpeaksOfPointsOnly(t *testing.T) {
	h := &TokenomicsHandlers{}
	rec := httptest.NewRecorder()
	h.handleGamificationStatus(rec, httptest.NewRequest(http.MethodGet, "/v1/gamification/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["unit"] != "points" {
		t.Errorf("unit = %v, want points", body["unit"])
	}
	for _, gone := range []string{"custody_mode", "redeemable", "transferable"} {
		if _, present := body[gone]; present {
			t.Errorf("%s is still in the response", gone)
		}
	}
	// The disclaimer itself says points have no value and cannot be transferred or redeemed;
	// check the token-specific words only.
	lower := strings.ToLower(rec.Body.String())
	for _, w := range []string{"echo earned", "custody", "wallet", "token", "staking"} {
		if strings.Contains(lower, w) {
			t.Errorf("status mentions %q: %s", w, rec.Body.String())
		}
	}
}

func TestTokenHold_NoQuestDependsOnTheToken(t *testing.T) {
	offered := 0
	for _, d := range append(append([]quests.Definition{}, quests.StarterQuests...), quests.AdvancedQuests...) {
		if quests.IsTokenQuest(d) {
			continue
		}
		offered++
		raw, _ := json.Marshal(questPointsEntry{ID: d.ID, Title: d.Title, Description: d.Description, Action: d.Action, Points: d.RewardECHO / 100_000_000, Badge: d.Badge, Tier: d.Tier})
		assertNoTokenWords(t, "quest "+d.ID, string(raw))
		if strings.Contains(strings.ToLower(d.Description), "vip") || strings.Contains(strings.ToLower(d.Title), "vip") {
			t.Errorf("quest %s rewards buying VIP; a purchase must earn nothing", d.ID)
		}
	}
	if offered == 0 {
		t.Fatal("no quests left once token quests are removed")
	}
}
