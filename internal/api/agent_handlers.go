package api

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/agentbus"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

func (h *V3Handlers) handleAgent(w http.ResponseWriter, r *http.Request) {
	if h.Agent == nil {
		WriteError(w, http.StatusServiceUnavailable, "AGENT_UNAVAILABLE", "agent bus not configured", r.Header.Get("X-Request-ID"))
		return
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v3/agent"), "/")
	switch {
	case path == "resolve" && r.Method == http.MethodPost:
		h.handleAgentResolve(w, r)
	case path == "send" && r.Method == http.MethodPost:
		h.handleAgentSend(w, r)
	case path == "grants" && r.Method == http.MethodPost:
		h.handleAgentGrantCreate(w, r)
	case path == "grants" && r.Method == http.MethodGet:
		h.handleAgentGrantList(w, r)
	case strings.HasPrefix(path, "grants/") && r.Method == http.MethodDelete:
		h.handleAgentGrantRevoke(w, r, strings.TrimPrefix(path, "grants/"))
	case path == "confirm" && r.Method == http.MethodPost:
		h.handleAgentConfirm(w, r)
	case path == "signal" && r.Method == http.MethodPost:
		h.handleAgentSignal(w, r)
	case path == "search" && r.Method == http.MethodPost:
		h.handleAgentSearch(w, r)
	case path == "recovery/shares" && r.Method == http.MethodPost:
		h.handleAgentShareMeta(w, r)
	case path == "rotate-key" && r.Method == http.MethodPost:
		h.handleAgentRotate(w, r)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown agent route", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) readBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(strings.NewReader(string(body)))
	return body, nil
}

func (h *V3Handlers) userFromPasskey(r *http.Request, body []byte) (string, error) {
	did := r.Header.Get(headerSenderDID)
	sigB64 := r.Header.Get(headerSignature)
	ts := r.Header.Get(headerTimestamp)
	if did == "" || sigB64 == "" || ts == "" {
		return "", errors.New("missing passkey headers")
	}
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return "", errors.New("bad timestamp")
	}
	if skew := time.Now().Unix() - tsInt; skew > maxClockSkewSeconds || skew < -maxClockSkewSeconds {
		return "", errors.New("stale timestamp")
	}
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		sig, err = base64.URLEncoding.DecodeString(sigB64)
		if err != nil {
			return "", errors.New("bad signature")
		}
	}
	msg := canonicalSigningString(r.Method, r.URL.Path, ts, body)
	var pubs []*ecdsa.PublicKey
	if h.ResolveDeviceKeys != nil {
		resolved, err := h.ResolveDeviceKeys(r.Context(), did)
		if err != nil || len(resolved) == 0 {
			return "", errors.New("unknown key")
		}
		pubs = resolved
	} else if pub := h.Agent.DeviceKey(did); pub != nil {
		pubs = []*ecdsa.PublicKey{pub}
	}
	if len(pubs) == 0 {
		return "", errors.New("unknown key")
	}
	for _, pub := range pubs {
		if didkey.VerifyECDSAP256SHA256(pub, msg, sig) == nil {
			return did, nil
		}
	}
	return "", errors.New("bad signature")
}

func (h *V3Handlers) grantFromRequest(r *http.Request, verb string) (*agentbus.Grant, int, string) {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return nil, http.StatusUnauthorized, "AGENT_GRANT_REQUIRED"
	}
	if !strings.HasPrefix(auth, "Grant ") {
		return nil, http.StatusUnauthorized, "AGENT_GRANT_REQUIRED"
	}
	token := strings.TrimSpace(strings.TrimPrefix(auth, "Grant "))
	if token == "" {
		return nil, http.StatusUnauthorized, "AGENT_GRANT_REQUIRED"
	}
	g, err := h.Agent.Lookup(token, verb, time.Now())
	if err != nil {
		if errors.Is(err, agentbus.ErrExpired) {
			return nil, http.StatusUnauthorized, "GRANT_EXPIRED"
		}
		if errors.Is(err, agentbus.ErrVerb) {
			return nil, http.StatusForbidden, "VERB_MISMATCH"
		}
		return nil, http.StatusUnauthorized, "GRANT_INVALID"
	}
	return g, 0, ""
}

// resolveGrant accepts a send grant or a read_metadata grant. A search grant
// is a verb mismatch. Expired and missing grants stay unauthorized.
func (h *V3Handlers) resolveGrant(r *http.Request) (*agentbus.Grant, int, string) {
	g, code, errCode := h.grantFromRequest(r, agentbus.VerbSend)
	if g != nil || errCode != "VERB_MISMATCH" {
		return g, code, errCode
	}
	return h.grantFromRequest(r, agentbus.VerbReadMetadata)
}

func (h *V3Handlers) handleAgentGrantCreate(w http.ResponseWriter, r *http.Request) {
	body, err := h.readBody(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid body", r.Header.Get("X-Request-ID"))
		return
	}
	userDID, err := h.userFromPasskey(r, body)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_INVALID_SIGNATURE", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		AgentDID    string `json:"agent_did"`
		AgentName   string `json:"agent_name"`
		Verb        string `json:"verb"`
		ThreadID    string `json:"thread_id"`
		ExpiresAt   string `json:"expires_at"`
		EmitSignals bool   `json:"emit_signals"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.AgentDID == "" || req.Verb == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "agent_did and verb required", r.Header.Get("X-Request-ID"))
		return
	}
	exp, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "expires_at must be RFC3339", r.Header.Get("X-Request-ID"))
		return
	}
	g, err := h.Agent.Issue(agentbus.Grant{
		UserDID:     userDID,
		AgentDID:    req.AgentDID,
		AgentName:   req.AgentName,
		Verb:        req.Verb,
		ThreadID:    req.ThreadID,
		ExpiresAt:   exp,
		EmitSignals: req.EmitSignals,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, "GRANT_REJECTED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"grant_id":     g.ID,
		"token":        g.Token,
		"agent_did":    g.AgentDID,
		"agent_name":   g.AgentName,
		"verb":         g.Verb,
		"thread_id":    g.ThreadID,
		"expires_at":   g.ExpiresAt.UTC().Format(time.RFC3339),
		"emit_signals": g.EmitSignals,
	})
}

func (h *V3Handlers) handleAgentGrantList(w http.ResponseWriter, r *http.Request) {
	body, _ := h.readBody(r)
	userDID, err := h.userFromPasskey(r, body)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_INVALID_SIGNATURE", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	list := h.Agent.List(userDID)
	out := make([]map[string]interface{}, 0, len(list))
	for _, g := range list {
		out = append(out, map[string]interface{}{
			"grant_id":     g.ID,
			"agent_did":    g.AgentDID,
			"agent_name":   g.AgentName,
			"verb":         g.Verb,
			"thread_id":    g.ThreadID,
			"expires_at":   g.ExpiresAt.UTC().Format(time.RFC3339),
			"emit_signals": g.EmitSignals,
		})
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"grants": out})
}

func (h *V3Handlers) handleAgentGrantRevoke(w http.ResponseWriter, r *http.Request, id string) {
	body, _ := h.readBody(r)
	userDID, err := h.userFromPasskey(r, body)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_INVALID_SIGNATURE", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if !h.Agent.Revoke(userDID, id) {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "grant not found", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"revoked": true})
}

func (h *V3Handlers) handleAgentResolve(w http.ResponseWriter, r *http.Request) {
	g, code, errCode := h.resolveGrant(r)
	if g == nil {
		WriteError(w, code, errCode, "agent grant required", r.Header.Get("X-Request-ID"))
		return
	}
	body, err := h.readBody(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid body", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		Name string `json:"name"`
		Room string `json:"room"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid JSON", r.Header.Get("X-Request-ID"))
		return
	}
	if req.Name == "" && req.Room == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "name or room required", r.Header.Get("X-Request-ID"))
		return
	}
	var peer string
	var thread string
	if req.Room != "" {
		parts := strings.Split(req.Room, ":")
		if len(parts) != 3 || parts[0] != "dm" {
			WriteError(w, http.StatusBadRequest, "INVALID_ROOM", "room must be dm:{did}:{did}", r.Header.Get("X-Request-ID"))
			return
		}
		thread = agentbus.DirectThread(parts[1], parts[2])
		if parts[1] == g.UserDID {
			peer = parts[2]
		} else if parts[2] == g.UserDID {
			peer = parts[1]
		} else {
			WriteError(w, http.StatusForbidden, "ROOM_DENIED", "grant holder is not in the room", r.Header.Get("X-Request-ID"))
			return
		}
	} else {
		if h.Contacts == nil {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "username directory unavailable", r.Header.Get("X-Request-ID"))
			return
		}
		hits, err := h.Contacts.SearchByUsername(r.Context(), g.UserDID, req.Name)
		if err != nil || len(hits) == 0 {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "name not found", r.Header.Get("X-Request-ID"))
			return
		}
		peer, _ = hits[0]["did"].(string)
		if peer == "" {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "name not found", r.Header.Get("X-Request-ID"))
			return
		}
		thread = agentbus.DirectThread(g.UserDID, peer)
	}
	if !h.Agent.AllowsThread(g, thread) {
		WriteError(w, http.StatusForbidden, "THREAD_DENIED", "thread not in grant", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"did":       peer,
		"thread_id": thread,
	})
}

type agentSendBody struct {
	To         string `json:"to"`
	Ciphertext []byte `json:"ciphertext"`
	Text       string `json:"text"`
	ExpiresAt  string `json:"expires_at"`
	RenderHint string `json:"render_hint"`
}

func (h *V3Handlers) handleAgentSend(w http.ResponseWriter, r *http.Request) {
	body, err := h.readBody(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid body", r.Header.Get("X-Request-ID"))
		return
	}
	var req agentSendBody
	if err := json.Unmarshal(body, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid JSON", r.Header.Get("X-Request-ID"))
		return
	}
	if req.Text != "" {
		WriteError(w, http.StatusBadRequest, "PLAINTEXT_REFUSED", "text is not accepted", r.Header.Get("X-Request-ID"))
		return
	}
	g, code, errCode := h.grantFromRequest(r, agentbus.VerbSend)
	if g == nil && errCode == "VERB_MISMATCH" {
		// pay / forward / rotate grants hit the confirmation gate instead of send
		for _, verb := range []string{agentbus.VerbPay, agentbus.VerbForward, agentbus.VerbRotateKey} {
			alt, _, _ := h.grantFromRequest(r, verb)
			if alt != nil {
				h.openConfirmation(w, r, alt, req)
				return
			}
		}
	}
	if g == nil {
		WriteError(w, code, errCode, "agent grant required", r.Header.Get("X-Request-ID"))
		return
	}
	h.enqueueSealed(w, r, g, req, false)
}

func (h *V3Handlers) openConfirmation(w http.ResponseWriter, r *http.Request, g *agentbus.Grant, req agentSendBody) {
	if len(req.Ciphertext) == 0 || req.To == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "to and ciphertext required", r.Header.Get("X-Request-ID"))
		return
	}
	if !agentbus.ValidHint(req.RenderHint) {
		WriteError(w, http.StatusBadRequest, "INVALID_HINT", "render_hint refused", r.Header.Get("X-Request-ID"))
		return
	}
	thread := agentbus.DirectThread(g.UserDID, req.To)
	if !h.Agent.AllowsThread(g, thread) {
		WriteError(w, http.StatusForbidden, "THREAD_DENIED", "thread not in grant", r.Header.Get("X-Request-ID"))
		return
	}
	ch, err := h.Agent.OpenChallenge(agentbus.Challenge{
		UserDID:    g.UserDID,
		AgentDID:   g.AgentDID,
		Verb:       g.Verb,
		To:         req.To,
		ThreadID:   thread,
		Ciphertext: req.Ciphertext,
		ExpiresAt:  req.ExpiresAt,
		RenderHint: req.RenderHint,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "CHALLENGE_FAILED", "could not open confirmation", r.Header.Get("X-Request-ID"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":              "needs_confirmation",
		"code":               "NEEDS_CONFIRMATION",
		"challenge_id":       ch.ID,
		"needs_confirmation": true,
	})
}

func (h *V3Handlers) recipientTrusts(r *http.Request, recipient, sender string) bool {
	if h.Agent.IsContact(recipient, sender) {
		return true
	}
	if h.Contacts == nil {
		return false
	}
	list, err := h.Contacts.GetContacts(r.Context(), recipient)
	if err != nil {
		return false
	}
	for _, c := range list {
		if c != nil && c.ContactDID == sender && !c.Blocked {
			return true
		}
	}
	return false
}

func (h *V3Handlers) enqueueSealed(w http.ResponseWriter, r *http.Request, g *agentbus.Grant, req agentSendBody, fromConfirm bool) {
	if len(req.Ciphertext) == 0 || req.To == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "to and ciphertext required", r.Header.Get("X-Request-ID"))
		return
	}
	if !agentbus.ValidHint(req.RenderHint) {
		WriteError(w, http.StatusBadRequest, "INVALID_HINT", "render_hint refused", r.Header.Get("X-Request-ID"))
		return
	}
	thread := agentbus.DirectThread(g.UserDID, req.To)
	if !h.Agent.AllowsThread(g, thread) {
		WriteError(w, http.StatusForbidden, "THREAD_DENIED", "thread not in grant", r.Header.Get("X-Request-ID"))
		return
	}
	untrusted := !h.recipientTrusts(r, req.To, g.UserDID)
	d := agentbus.Delivery{
		To:         req.To,
		ThreadID:   thread,
		Ciphertext: req.Ciphertext,
		ExpiresAt:  req.ExpiresAt,
		RenderHint: req.RenderHint,
		Untrusted:  untrusted,
		FromUser:   g.UserDID,
		AgentDID:   g.AgentDID,
	}
	h.Agent.Enqueue(d)
	if h.Signals != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"ciphertext":  req.Ciphertext,
			"expires_at":  req.ExpiresAt,
			"render_hint": req.RenderHint,
			"untrusted":   untrusted,
			"thread_id":   thread,
		})
		h.Signals.PublishSignal(req.To, WSMessage{
			Type:           "sealed_text",
			From:           g.UserDID,
			To:             req.To,
			ConversationID: thread,
			Payload:        payload,
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
		})
	}
	_ = fromConfirm
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"queued":    true,
		"thread_id": thread,
		"untrusted": untrusted,
	})
}

func (h *V3Handlers) handleAgentConfirm(w http.ResponseWriter, r *http.Request) {
	body, err := h.readBody(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid body", r.Header.Get("X-Request-ID"))
		return
	}
	if strings.HasPrefix(strings.ToLower(r.Header.Get("Authorization")), "grant ") && r.Header.Get(headerSignature) == "" {
		WriteError(w, http.StatusUnauthorized, "USER_SIGNATURE_REQUIRED", "agent grant cannot confirm", r.Header.Get("X-Request-ID"))
		return
	}
	userDID, err := h.userFromPasskey(r, body)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_INVALID_SIGNATURE", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		ChallengeID string `json:"challenge_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.ChallengeID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "challenge_id required", r.Header.Get("X-Request-ID"))
		return
	}
	ch, err := h.Agent.ConsumeChallenge(req.ChallengeID, userDID)
	if err != nil {
		if errors.Is(err, agentbus.ErrReplay) {
			WriteError(w, http.StatusConflict, "CHALLENGE_REPLAY", "challenge already used", r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "challenge not found", r.Header.Get("X-Request-ID"))
		return
	}
	if ch.Verb == agentbus.VerbRotateKey {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"rotated":    true,
			"generation": h.Agent.KeyGeneration(userDID),
		})
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"queued":    true,
		"thread_id": ch.ThreadID,
	})
}

func (h *V3Handlers) handleAgentSignal(w http.ResponseWriter, r *http.Request) {
	g, code, errCode := h.grantFromRequest(r, agentbus.VerbSend)
	if g == nil {
		WriteError(w, code, errCode, "agent grant required", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		To   string `json:"to"`
		Type string `json:"type"`
	}
	body, _ := h.readBody(r)
	_ = json.Unmarshal(body, &req)
	if !g.EmitSignals {
		WriteJSON(w, http.StatusOK, map[string]interface{}{"published": false})
		return
	}
	if h.Signals != nil && req.To != "" {
		thread := agentbus.DirectThread(g.UserDID, req.To)
		payload, _ := json.Marshal(map[string]string{
			"conversation_id": thread,
			"state":           "start",
		})
		h.Signals.PublishSignal(req.To, WSMessage{
			Type:           "typing",
			From:           g.UserDID,
			To:             req.To,
			ConversationID: thread,
			Payload:        payload,
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
		})
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"published": true})
}

func (h *V3Handlers) handleAgentSearch(w http.ResponseWriter, r *http.Request) {
	g, code, errCode := h.grantFromRequest(r, agentbus.VerbSearch)
	if g == nil {
		WriteError(w, code, errCode, "agent grant required", r.Header.Get("X-Request-ID"))
		return
	}
	WriteError(w, http.StatusBadRequest, "SEARCH_EDGE_ONLY", "search runs on the user device", r.Header.Get("X-Request-ID"))
}

func (h *V3Handlers) handleAgentShareMeta(w http.ResponseWriter, r *http.Request) {
	body, err := h.readBody(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid body", r.Header.Get("X-Request-ID"))
		return
	}
	if bytesContainShare(body) {
		WriteError(w, http.StatusBadRequest, "SHARE_BYTES_REFUSED", "share material is not stored", r.Header.Get("X-Request-ID"))
		return
	}
	userDID, err := h.userFromPasskey(r, body)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_INVALID_SIGNATURE", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		Index     int    `json:"index"`
		HolderDID string `json:"holder_did"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.HolderDID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "holder_did required", r.Header.Get("X-Request-ID"))
		return
	}
	h.Agent.PutShareMeta(agentbus.ShareMeta{Index: req.Index, HolderDID: req.HolderDID, Status: "held"})
	if err := h.Agent.Reconstruct(); err == nil {
		WriteError(w, http.StatusInternalServerError, "RECONSTRUCT_LEAK", "server must not reconstruct", r.Header.Get("X-Request-ID"))
		return
	}
	_ = userDID
	WriteJSON(w, http.StatusOK, map[string]interface{}{"stored": "metadata"})
}

func bytesContainShare(body []byte) bool {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(body, &probe); err != nil {
		return false
	}
	for _, key := range []string{"share", "share_bytes", "secret"} {
		if _, ok := probe[key]; ok {
			return true
		}
	}
	return false
}

func (h *V3Handlers) handleAgentRotate(w http.ResponseWriter, r *http.Request) {
	g, code, errCode := h.grantFromRequest(r, agentbus.VerbRotateKey)
	if g == nil {
		WriteError(w, code, errCode, "agent grant required", r.Header.Get("X-Request-ID"))
		return
	}
	before := h.Agent.KeyGeneration(g.UserDID)
	ch, err := h.Agent.OpenChallenge(agentbus.Challenge{
		UserDID:  g.UserDID,
		AgentDID: g.AgentDID,
		Verb:     agentbus.VerbRotateKey,
		ThreadID: g.ThreadID,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "CHALLENGE_FAILED", "could not open confirmation", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Agent.KeyGeneration(g.UserDID) != before {
		WriteError(w, http.StatusInternalServerError, "ROTATED_WITHOUT_USER", "grant rotated a key", r.Header.Get("X-Request-ID"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":              "needs_confirmation",
		"code":               "NEEDS_CONFIRMATION",
		"challenge_id":       ch.ID,
		"needs_confirmation": true,
		"generation":         before,
	})
}
