package api

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/services/cloudstorage"
	"github.com/thechadcromwell/echoapp/internal/services/media"
)

// WireCloudStorage registers WO-46 integration routes.
func (h *V3Handlers) WireCloudStorage(mux *http.ServeMux) {
	mux.HandleFunc("/v3/integrations/cloud", h.handleCloudIntegrations)
	mux.HandleFunc("/v3/integrations/cloud/", h.handleCloudIntegrationByProvider)
}

func (h *V3Handlers) handleCloudIntegrations(w http.ResponseWriter, r *http.Request) {
	if h.CloudStorage == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "cloud storage not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	switch r.Method {
	case http.MethodGet:
		providers := h.CloudStorage.ListProviders(did)
		out := make([]string, len(providers))
		for i, p := range providers {
			out[i] = string(p)
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{"providers": out})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET only", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) handleCloudIntegrationByProvider(w http.ResponseWriter, r *http.Request) {
	if h.CloudStorage == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "cloud storage not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	provider := strings.TrimPrefix(r.URL.Path, "/v3/integrations/cloud/")
	if provider == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_PROVIDER", "provider required", r.Header.Get("X-Request-ID"))
		return
	}
	p := cloudstorage.Provider(provider)
	switch r.Method {
	case http.MethodPost:
		var req struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}
		if err := h.readJSON(r, &req); err != nil || req.AccessToken == "" {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "access_token required", r.Header.Get("X-Request-ID"))
			return
		}
		h.CloudStorage.SaveToken(did, cloudstorage.Token{
			Provider:    p,
			AccessToken: req.AccessToken,
			Refresh:     req.RefreshToken,
		})
		WriteJSON(w, http.StatusCreated, map[string]string{"provider": provider, "status": "connected"})
	case http.MethodDelete:
		if err := h.CloudStorage.Revoke(did, p); err != nil {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"provider": provider, "status": "revoked"})
	case http.MethodGet:
		redirect := r.URL.Query().Get("redirect_uri")
		if redirect == "" {
			redirect = "echo://oauth/cloud"
		}
		WriteJSON(w, http.StatusOK, map[string]string{
			"provider":      provider,
			"authorize_url": cloudstorage.AuthURL(p, redirect),
		})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET, POST, or DELETE", r.Header.Get("X-Request-ID"))
	}
}

// WireMediaResumable registers chunked upload resume routes (WO-21/34).
func (h *V3Handlers) WireMediaResumable(mux *http.ServeMux) {
	mux.HandleFunc("/v3/media/upload/init", h.handleMediaUploadInit)
	mux.HandleFunc("/v3/media/upload/chunk", h.handleMediaUploadChunk)
	mux.HandleFunc("/v3/media/upload/complete", h.handleMediaUploadComplete)
}

func (h *V3Handlers) handleMediaUploadInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Media == nil {
		WriteError(w, http.StatusServiceUnavailable, "MEDIA_UNAVAILABLE", "Media service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	sizeStr := r.Header.Get("X-Encrypted-Size")
	tierStr := r.Header.Get("X-Trust-Tier")
	size, _ := strconv.ParseInt(sizeStr, 10, 64)
	tier, _ := strconv.Atoi(tierStr)
	if tier == 0 {
		tier = 1
	}
	sess, err := h.Media.InitUpload(r.Context(), media.UploadRequest{
		UploaderDID:   h.getDID(r),
		ContentType:   r.Header.Get("Content-Type"),
		EncryptedSize: size,
		TrustTier:     tier,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, "UPLOAD_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"file_id":          sess.FileID,
		"total_chunks":     sess.ChunkCount,
		"chunk_size_bytes": media.ChunkSize,
		"content_type":     sess.ContentType,
	})
}

func (h *V3Handlers) handleMediaUploadChunk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PUT", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Media == nil {
		WriteError(w, http.StatusServiceUnavailable, "MEDIA_UNAVAILABLE", "Media service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	fileID := r.Header.Get("X-File-Id")
	indexStr := r.Header.Get("X-Chunk-Index")
	index, err := strconv.Atoi(indexStr)
	if fileID == "" || err != nil || index < 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_HEADERS", "X-File-Id and X-Chunk-Index required", r.Header.Get("X-Request-ID"))
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "could not read chunk", r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.Media.UploadChunk(r.Context(), fileID, index, data); err != nil {
		WriteError(w, http.StatusBadRequest, "UPLOAD_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"file_id": fileID, "chunk_index": index})
}

func (h *V3Handlers) handleMediaUploadComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Media == nil {
		WriteError(w, http.StatusServiceUnavailable, "MEDIA_UNAVAILABLE", "Media service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		FileID string `json:"file_id"`
	}
	if err := h.readJSON(r, &req); err != nil || req.FileID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "file_id required", r.Header.Get("X-Request-ID"))
		return
	}
	result, err := h.Media.CompleteUpload(r.Context(), req.FileID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "UPLOAD_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, result)
}

func (h *V3Handlers) handleMediaManifest(w http.ResponseWriter, r *http.Request, fileID string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Media == nil {
		WriteError(w, http.StatusServiceUnavailable, "MEDIA_UNAVAILABLE", "Media service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	sess, err := h.Media.Manifest(r.Context(), fileID)
	if err != nil {
		// Fall back to committed file metadata for completed uploads.
		_, meta, ferr := h.Media.Download(r.Context(), fileID, h.getDID(r))
		if ferr != nil || meta == nil {
			WriteError(w, http.StatusNotFound, "FILE_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"file_id":          fileID,
			"total_chunks":     meta.ChunkCount,
			"chunk_size_bytes": media.ChunkSize,
			"complete":         true,
		})
		return
	}
	received := 0
	for _, ok := range sess.Received {
		if ok {
			received++
		}
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"file_id":          fileID,
		"total_chunks":     sess.ChunkCount,
		"chunk_size_bytes": media.ChunkSize,
		"received_chunks":  received,
		"complete":         false,
	})
}

func (h *V3Handlers) handleMediaFilecoin(w http.ResponseWriter, r *http.Request, fileID string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Media == nil {
		WriteError(w, http.StatusServiceUnavailable, "MEDIA_UNAVAILABLE", "Media service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	deal := h.Media.FilecoinDealForFile(r.Context(), fileID)
	if deal == nil {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"file_id": fileID,
			"enabled": false,
		})
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"file_id":  fileID,
		"enabled":  true,
		"cid":      deal.CID,
		"deal_id":  deal.DealID,
		"status":   deal.Status,
		"provider": deal.Provider,
	})
}
