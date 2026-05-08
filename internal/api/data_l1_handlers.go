package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
	"github.com/thechadcromwell/echoapp/internal/validation"
)

// dataL1MaxSupportedSchemaVersion is the newest batch schema the API accepts (WO-35).
const dataL1MaxSupportedSchemaVersion = 3

// handleDataL1MerkleRoots accepts POST /v1/data-l1/merkle-roots and forwards a
// MerkleRootUpdate to the local Data L1 (WO-230 Step 5). Unauthenticated for
// Phase-1 local testnet; restrict at the network layer in shared environments.
func (rt *Router) handleDataL1MerkleRoots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.DataL1 == nil {
		WriteError(w, http.StatusServiceUnavailable, "DATA_L1_NOT_CONFIGURED", "DATA_L1_URL is not set", r.Header.Get("X-Request-ID"))
		return
	}

	var body struct {
		Root          string `json:"root"`
		LeafCount     int    `json:"leafCount"`
		LeafSnake     int    `json:"leaf_count"`
		TimeFrom      string `json:"timeFrom,omitempty"`
		TimeTo        string `json:"timeTo,omitempty"`
		SchemaVersion int    `json:"schemaVersion,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	leaf := body.LeafCount
	if leaf == 0 {
		leaf = body.LeafSnake
	}
	root := strings.TrimSpace(strings.ToLower(body.Root))
	if len(root) != 64 || leaf <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "root must be 64 hex chars and leaf_count must be positive", r.Header.Get("X-Request-ID"))
		return
	}
	for _, c := range root {
		if c >= '0' && c <= '9' {
			continue
		}
		if c >= 'a' && c <= 'f' {
			continue
		}
		WriteError(w, http.StatusBadRequest, "INVALID_ROOT", "root must be lowercase hex", r.Header.Get("X-Request-ID"))
		return
	}

	rootBytes, err := hex.DecodeString(root)
	if err != nil || len(rootBytes) != 32 {
		WriteError(w, http.StatusBadRequest, "INVALID_ROOT", "root must decode to 32 bytes", r.Header.Get("X-Request-ID"))
		return
	}

	var tr validation.TimeRange
	if body.TimeFrom != "" || body.TimeTo != "" {
		if body.TimeFrom == "" || body.TimeTo == "" {
			WriteError(w, http.StatusBadRequest, "INVALID_TIME_RANGE", "timeFrom and timeTo must both be set when present", r.Header.Get("X-Request-ID"))
			return
		}
		fromT, err1 := time.Parse(time.RFC3339, body.TimeFrom)
		toT, err2 := time.Parse(time.RFC3339, body.TimeTo)
		if err1 != nil || err2 != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_TIME_RANGE", "timeFrom and timeTo must be RFC3339 timestamps", r.Header.Get("X-Request-ID"))
			return
		}
		tr = validation.TimeRange{From: fromT, To: toT}
	}

	sub := validation.DataL1Submission{
		MerkleRoot:      rootBytes,
		CommitmentCount: leaf,
		TimeRange:       tr,
		SchemaVersion:   body.SchemaVersion,
	}
	if err := validation.ValidateDataL1Submission(sub, dataL1MaxSupportedSchemaVersion); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	tx := metagraph.DataL1MerkleRootUpdate{Root: root, LeafCount: leaf}
	txID, err := rt.DataL1.SubmitDataL1(r.Context(), tx)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "DATA_L1_SUBMIT_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if txID == "" {
		txID = "accepted"
	}
	WriteJSON(w, http.StatusCreated, map[string]string{
		"tx_id":  txID,
		"txHash": txID,
	})
}
