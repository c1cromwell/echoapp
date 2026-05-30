package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/credentials/oidc4vc"
	"github.com/thechadcromwell/echoapp/pkg/passport"
	"github.com/thechadcromwell/echoapp/pkg/passport/disclosure"
)

type passportPresentBeginRequest struct {
	CredentialRefID  string   `json:"credential_ref_id"`
	RedirectURI      string   `json:"redirect_uri,omitempty"`
	DisclosureFields []string `json:"disclosure_fields,omitempty"`
}

type passportPresentAcceptRequest struct {
	SessionID   string                          `json:"session_id"`
	VPToken     string                          `json:"vp_token,omitempty"`
	State       string                          `json:"state,omitempty"`
	CallbackURL string                          `json:"callback_url,omitempty"`
	Submission  *oidc4vc.PresentationSubmission `json:"presentation_submission,omitempty"`
}

func (rt *Router) handlePassport(w http.ResponseWriter, r *http.Request) {
	if rt.Passport == nil {
		WriteError(w, http.StatusServiceUnavailable, "PASSPORT_UNAVAILABLE", "Passport service not configured", r.Header.Get("X-Request-ID"))
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/passport/")
	switch {
	case path == "credentials" || strings.HasPrefix(path, "credentials/"):
		rt.handlePassportCredentials(w, r, path)
	case path == "present/begin" || path == "present/accept":
		rt.handlePassportPresent(w, r, path)
	case path == "sync":
		rt.handlePassportSync(w, r)
	case strings.HasPrefix(path, "recovery/"):
		rt.handlePassportRecovery(w, r, recoveryActionFromPath(path))
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown passport path", r.Header.Get("X-Request-ID"))
	}
}

func (rt *Router) handlePassportCredentials(w http.ResponseWriter, r *http.Request, path string) {
	holderDID, ok := rt.authenticatedHolderDID(w, r)
	if !ok {
		return
	}

	if path == "credentials" {
		switch r.Method {
		case http.MethodGet:
			refs, err := rt.Passport.List(r.Context(), holderDID)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
				return
			}
			WriteJSON(w, http.StatusOK, map[string]interface{}{
				"credentials": refs,
				"count":       len(refs),
			})
		case http.MethodPost:
			var req passport.RegisterRefRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
				return
			}
			ref, err := rt.Passport.Register(r.Context(), holderDID, req)
			if err != nil {
				if errors.Is(err, passport.ErrDuplicateHash) {
					WriteError(w, http.StatusConflict, "DUPLICATE_CREDENTIAL", err.Error(), r.Header.Get("X-Request-ID"))
					return
				}
				WriteError(w, http.StatusBadRequest, "REGISTER_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
				return
			}
			WriteJSON(w, http.StatusCreated, ref)
		default:
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET and POST are allowed", r.Header.Get("X-Request-ID"))
		}
		return
	}

	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	refID := strings.TrimPrefix(path, "credentials/")
	if refID == "" || strings.Contains(refID, "/") {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown credential path", r.Header.Get("X-Request-ID"))
		return
	}
	ref, err := rt.Passport.Get(r.Context(), holderDID, refID)
	if err != nil {
		if errors.Is(err, passport.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "Credential ref not found", r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusInternalServerError, "GET_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, ref)
}

func (rt *Router) handlePassportPresent(w http.ResponseWriter, r *http.Request, path string) {
	if rt.OIDCVerifier == nil {
		WriteError(w, http.StatusServiceUnavailable, "OIDC4VC_DISABLED", "OIDC4VC verifier is not enabled", r.Header.Get("X-Request-ID"))
		return
	}
	holderDID, ok := rt.authenticatedHolderDID(w, r)
	if !ok {
		return
	}

	switch path {
	case "present/begin":
		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
			return
		}
		var req passportPresentBeginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		ref, err := rt.Passport.Get(r.Context(), holderDID, req.CredentialRefID)
		if err != nil || ref == nil {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "Credential ref not found", r.Header.Get("X-Request-ID"))
			return
		}
		if ref.RevocationStatus == "revoked" {
			WriteError(w, http.StatusConflict, "CREDENTIAL_REVOKED", "Credential is revoked", r.Header.Get("X-Request-ID"))
			return
		}
		redirectURI := req.RedirectURI
		if redirectURI == "" {
			redirectURI = "echo-passport://present/callback"
		}
		presReq, err := rt.OIDCVerifier.BeginPresentation(ref.CredentialType, redirectURI)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "PRESENTATION_REQUEST_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		expiresAt := time.Now().Add(5 * time.Minute)
		rt.storePassportPresentSession(presReq.State, passportPresentSession{
			State:            presReq.State,
			HolderDID:        holderDID,
			CredentialRefID:  ref.RefID,
			CredentialType:   ref.CredentialType,
			RedirectURI:      redirectURI,
			DisclosureFields: req.DisclosureFields,
			ExpiresAt:        expiresAt,
		})
		base := strings.TrimSuffix(rt.OIDCVerifierBaseURL, "/")
		verifierURL := base + "/verification/ui?state=" + url.QueryEscape(presReq.State) +
			"&credential_type=" + url.QueryEscape(ref.CredentialType)
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"session_id":           presReq.State,
			"credential_ref_id":    ref.RefID,
			"credential_type":      ref.CredentialType,
			"verifier_url":         verifierURL,
			"expires_at":           expiresAt.Format(time.RFC3339),
			"presentation_request": presReq,
		})
	case "present/accept":
		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
			return
		}
		var req passportPresentAcceptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		sessionID := req.SessionID
		if sessionID == "" {
			sessionID = req.State
		}
		sess, ok := rt.loadPassportPresentSession(sessionID)
		if !ok || sess.HolderDID != holderDID {
			WriteError(w, http.StatusBadRequest, "INVALID_SESSION", "Unknown or expired presentation session", r.Header.Get("X-Request-ID"))
			return
		}
		vpToken := req.VPToken
		state := req.State
		submission := req.Submission
		if req.CallbackURL != "" {
			parsed, err := url.Parse(req.CallbackURL)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "INVALID_CALLBACK", "Malformed callback URL", r.Header.Get("X-Request-ID"))
				return
			}
			q := parsed.Query()
			if vpToken == "" {
				vpToken = q.Get("vp_token")
			}
			if state == "" {
				state = q.Get("state")
			}
		}
		if vpToken == "" || state == "" {
			WriteError(w, http.StatusBadRequest, "MISSING_VP", "vp_token and state are required", r.Header.Get("X-Request-ID"))
			return
		}
		if len(sess.DisclosureFields) > 0 {
			if err := disclosure.ValidatePresentation(vpToken, sess.DisclosureFields); err != nil {
				WriteError(w, http.StatusBadRequest, "DISCLOSURE_VIOLATION", err.Error(), r.Header.Get("X-Request-ID"))
				return
			}
		}
		if submission == nil {
			submission = defaultPresentationSubmission(sess.CredentialType)
		}
		result, err := rt.OIDCVerifier.AcceptPresentation(r.Context(), vpToken, submission, state)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "PRESENTATION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		if !result.IsValid {
			msg := result.Error
			if msg == "" {
				msg = "credential verification failed"
			}
			WriteError(w, http.StatusBadRequest, "PRESENTATION_FAILED", msg, r.Header.Get("X-Request-ID"))
			return
		}
		rt.deletePassportPresentSession(sessionID)
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"verified":          result.IsValid,
			"holder_did":        result.HolderDID,
			"credential_ref_id": sess.CredentialRefID,
			"credential_type":   sess.CredentialType,
			"credentials":       result.Credentials,
		})
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown present path", r.Header.Get("X-Request-ID"))
	}
}

type passportPresentSession struct {
	State            string
	HolderDID        string
	CredentialRefID  string
	CredentialType   string
	RedirectURI      string
	DisclosureFields []string
	ExpiresAt        time.Time
}

func (rt *Router) authenticatedHolderDID(w http.ResponseWriter, r *http.Request) (string, bool) {
	holderDID, _ := r.Context().Value(ContextKeyUserID).(string)
	if holderDID == "" || !strings.HasPrefix(holderDID, "did:key:") {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authenticated did:key required", r.Header.Get("X-Request-ID"))
		return "", false
	}
	return holderDID, true
}

func (rt *Router) storePassportPresentSession(state string, sess passportPresentSession) {
	rt.enrollmentVCMu.Lock()
	defer rt.enrollmentVCMu.Unlock()
	if rt.passportPresentSessions == nil {
		rt.passportPresentSessions = make(map[string]passportPresentSession)
	}
	rt.passportPresentSessions[state] = sess
}

func (rt *Router) loadPassportPresentSession(state string) (passportPresentSession, bool) {
	rt.enrollmentVCMu.Lock()
	defer rt.enrollmentVCMu.Unlock()
	sess, ok := rt.passportPresentSessions[state]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return passportPresentSession{}, false
	}
	return sess, true
}

func (rt *Router) deletePassportPresentSession(state string) {
	rt.enrollmentVCMu.Lock()
	defer rt.enrollmentVCMu.Unlock()
	delete(rt.passportPresentSessions, state)
}

func (rt *Router) handlePassportSync(w http.ResponseWriter, r *http.Request) {
	if rt.PassportSync == nil {
		WriteError(w, http.StatusServiceUnavailable, "PASSPORT_SYNC_UNAVAILABLE", "Passport sync not configured", r.Header.Get("X-Request-ID"))
		return
	}
	holderDID, ok := rt.authenticatedHolderDID(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPost:
		var req passport.PushSyncRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		meta, err := rt.PassportSync.Push(r.Context(), holderDID, req)
		if err != nil {
			if errors.Is(err, passport.ErrSyncHashMismatch) {
				WriteError(w, http.StatusBadRequest, "HASH_MISMATCH", err.Error(), r.Header.Get("X-Request-ID"))
				return
			}
			WriteError(w, http.StatusBadRequest, "SYNC_PUSH_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, meta)
	case http.MethodGet:
		record, err := rt.PassportSync.Pull(r.Context(), holderDID)
		if err != nil {
			if errors.Is(err, passport.ErrSyncNotFound) {
				WriteError(w, http.StatusNotFound, "NOT_FOUND", "No sync blob for holder", r.Header.Get("X-Request-ID"))
				return
			}
			WriteError(w, http.StatusInternalServerError, "SYNC_PULL_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"storage_uri":       record.StorageURI,
			"content_hash":      record.ContentHash,
			"byte_size":         record.ByteSize,
			"version":           record.Version,
			"updated_at":        record.UpdatedAt.Format(time.RFC3339),
			"ciphertext_base64": passport.EncodeBase64(record.Ciphertext),
		})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET and POST are allowed", r.Header.Get("X-Request-ID"))
	}
}
