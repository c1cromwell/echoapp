package api

import (
	"context"
	"github.com/thechadcromwell/echoapp/internal/features"
	"net/http"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/tokenomics"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/leaderboard"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/models"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/quests"
)

// TrustScoreLookup is the subset of database.Database needed for Echo Score.
type TrustScoreLookup interface {
	GetTrustScore(ctx context.Context, did string) (*database.TrustScore, error)
	GetUserByDID(ctx context.Context, did string) (*database.User, error)
}

// TokenomicsHandlers serves the gamification endpoints: quests, leaderboard, streaks and the
// weekly pack, all in points. (The name predates the removal of the token.)
type TokenomicsHandlers struct {
	Service *tokenomics.Service
	Trust   TrustScoreLookup
}

func (h *TokenomicsHandlers) handleQuestsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	catalog, err := h.Service.Quests.ListCatalog(r.Context(), didStr)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "QUESTS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	// Rewards are points. Quests that needed a token (staking, delegation, stake-weighted
	// governance) or that rewarded buying VIP are not offered.
	starter, advanced := make([]questPointsEntry, 0), make([]questPointsEntry, 0)
	for _, e := range catalog {
		if quests.IsTokenQuest(e.Definition) {
			continue
		}
		entry := questPointsEntry{
			ID: e.ID, Title: e.Title, Description: e.Description, Action: e.Action, RequiredCount: e.RequiredCount,
			Points: e.RewardECHO / features.PointsPerUnit, Badge: e.Badge, Tier: e.Tier, MinTrustTier: e.MinTrustTier,
			CompletedAt: e.CompletedAt, RewardClaimed: e.RewardClaimed, Progress: e.Progress,
		}
		if e.Tier == quests.TierAdvanced {
			advanced = append(advanced, entry)
		} else {
			starter = append(starter, entry)
		}
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"starter":    starter,
		"advanced":   advanced,
		"unit":       "points",
		"request_id": r.Header.Get("X-Request-ID"),
	})

}

func (h *TokenomicsHandlers) handleQuestClaim(w http.ResponseWriter, r *http.Request, questID string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	txHash, err := h.Service.Quests.Claim(r.Context(), didStr, questID)
	if err != nil {
		code := http.StatusBadRequest
		errCode := "CLAIM_ERROR"
		if err == quests.ErrAlreadyClaimed {
			code = http.StatusConflict
			errCode = "ALREADY_CLAIMED"
		}
		WriteError(w, code, errCode, err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	// R0 gamification accrual (value-free): count the claimed reward toward the
	// leaderboard and the claim as daily activity for the streak. Best-effort —
	// never fails the claim itself.
	h.accrueGamification(r, didStr, questReward(questID))
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"quest_id":   questID,
		"tx_hash":    txHash,
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

// handleLeaderboard serves the usage leaderboard (requirements §Mechanic 1).
// GET /v1/gamification/leaderboard?window=weekly|monthly
func (h *TokenomicsHandlers) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil || h.Service.Leaderboard == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	window := leaderboard.Window(r.URL.Query().Get("window"))
	if !window.Valid() {
		window = leaderboard.WindowWeekly
	}
	snap, err := h.Service.Leaderboard.Top(r.Context(), window, leaderboard.DefaultLimit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LEADERBOARD_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	// Scores are shown in whole points.
	for i := range snap.Entries {
		snap.Entries[i].Score /= features.PointsPerUnit
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"snapshot":   snap,
		"unit":       "points",
		"request_id": r.Header.Get("X-Request-ID"),
	})

}

// questPointsEntry is a quest as served while the token is on hold: the same quest, with its
// reward expressed in points and no token-denominated field.
type questPointsEntry struct {
	ID            string      `json:"questId"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	Action        string      `json:"action"`
	RequiredCount int         `json:"requiredCount"`
	Points        int64       `json:"points"`
	Badge         string      `json:"badge"`
	Tier          quests.Tier `json:"tier"`
	MinTrustTier  int         `json:"minTrustTier,omitempty"`
	CompletedAt   string      `json:"completedAt,omitempty"`
	RewardClaimed bool        `json:"rewardClaimed"`
	Progress      int         `json:"progress"`
}

// handleStreak serves the authenticated user's daily-activity streak
// (requirements §Mechanic 4). GET /v1/gamification/streak
func (h *TokenomicsHandlers) handleStreak(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil || h.Service.Streaks == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	st, err := h.Service.Streaks.Get(r.Context(), didStr)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "STREAK_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"streak":     st,
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

// handleGamificationStatus reports the launch custody/compliance posture so the
// client can render the authoritative "no cash value" disclaimer. While custody
// is interim (ECHO_WALLET_REAL_FUNDS unset) the earned token is non-redeemable
// and non-transferable — see docs/legal/CORPORATE_STRUCTURE_AND_COMPLIANCE.md.
// GET /v1/gamification/status
func (h *TokenomicsHandlers) handleGamificationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	body := map[string]interface{}{"request_id": r.Header.Get("X-Request-ID")}
	// Points are a score. They are not a token, hold no value and convert to nothing.
	body["unit"] = "points"
	body["disclaimer"] = "Points are a score for activity in the app. They have no cash value and cannot be transferred, redeemed or exchanged for anything."
	if snap, ok := h.echoScoreForRequest(r); ok {
		body["echo_score"] = snap
	}
	WriteJSON(w, http.StatusOK, body)
}

func (h *TokenomicsHandlers) echoScoreForRequest(r *http.Request) (models.EchoScoreSnapshot, bool) {
	did, _ := r.Context().Value(ContextKeyUserID).(string)
	if did == "" {
		return models.EchoScoreSnapshot{}, false
	}
	if h.Trust != nil {
		if ts, err := h.Trust.GetTrustScore(r.Context(), did); err == nil && ts != nil {
			return models.SnapshotFromScore(int(ts.Score + 0.5)), true
		}
		if user, err := h.Trust.GetUserByDID(r.Context(), did); err == nil && user != nil && user.TrustTier > 0 {
			return models.SnapshotFromTier(user.TrustTier), true
		}
	}
	tier := TrustTierFromContext(r.Context(), 0)
	if tier <= 0 {
		return models.EchoScoreSnapshot{}, false
	}
	return models.SnapshotFromTier(tier), true
}

// handleActivityPing records a daily-activity signal for the streak (requirements
// §Mechanic 4). Clients call this on app open. POST /v1/gamification/activity/ping
func (h *TokenomicsHandlers) handleActivityPing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil || h.Service.Streaks == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	st, err := h.Service.Streaks.RecordActivity(r.Context(), didStr, time.Now())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "STREAK_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"streak":     st,
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

// handleWeeklyPack serves GET preview / POST open for this ISO week's pack.
func (h *TokenomicsHandlers) handleWeeklyPack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil || h.Service.Packs == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	didStr, _ := r.Context().Value(ContextKeyUserID).(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	streakDays := 0
	if h.Service.Streaks != nil {
		if st, err := h.Service.Streaks.Get(r.Context(), didStr); err == nil && st != nil {
			streakDays = st.CurrentDays
		}
	}
	now := time.Now()
	var pack interface{}
	if r.Method == http.MethodPost {
		pack = h.Service.Packs.Open(r.Context(), didStr, streakDays, now)
	} else {
		pack = h.Service.Packs.Preview(didStr, streakDays, now)
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"pack":       pack,
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

// questReward returns the ECHO datum reward for a quest id, 0 if unknown.
func questReward(questID string) int64 {
	if def, ok := quests.ByID(questID); ok {
		return def.RewardECHO
	}
	return 0
}

// accrueGamification records an earning event toward the leaderboard (amountDatum,
// if > 0) and counts it as daily activity for the streak. Best-effort: errors are
// swallowed so a gamification hiccup never fails the originating action.
func (h *TokenomicsHandlers) accrueGamification(r *http.Request, did string, amountDatum int64) {
	if h.Service == nil || did == "" {
		return
	}
	now := time.Now()
	if h.Service.Leaderboard != nil && amountDatum > 0 {
		tier := TrustTierFromContext(r.Context(), 1)
		_ = h.Service.Leaderboard.RecordEarning(r.Context(), did, tier, amountDatum, now)
	}
	if h.Service.Streaks != nil {
		_, _ = h.Service.Streaks.RecordActivity(r.Context(), did, now)
	}
}

// handleTokenomicsSubroutes is called from Router.handleV1 for tokenomics paths.
func (rt *Router) handleTokenomicsSubroutes(w http.ResponseWriter, r *http.Request) {
	if rt.Tokenomics == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	path := r.URL.Path
	switch {
	case path == "/v1/gamification/leaderboard":
		rt.Tokenomics.handleLeaderboard(w, r)
	case path == "/v1/gamification/streak":
		rt.Tokenomics.handleStreak(w, r)
	case path == "/v1/gamification/activity/ping":
		rt.Tokenomics.handleActivityPing(w, r)
	case path == "/v1/gamification/status":
		rt.Tokenomics.handleGamificationStatus(w, r)
	case path == "/v1/gamification/weekly-pack":
		rt.Tokenomics.handleWeeklyPack(w, r)
	case path == "/v1/gamification/quests":
		rt.Tokenomics.handleQuestsList(w, r)
	case strings.HasPrefix(path, "/v1/gamification/quests/") && strings.HasSuffix(path, "/claim"):
		questID := strings.TrimSuffix(strings.TrimPrefix(path, "/v1/gamification/quests/"), "/claim")
		questID = strings.Trim(questID, "/")
		rt.Tokenomics.handleQuestClaim(w, r, questID)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Not found", r.Header.Get("X-Request-ID"))
	}
}
