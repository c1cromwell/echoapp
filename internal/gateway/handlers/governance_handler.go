package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thechadcromwell/echoapp/internal/governance"
)

// GovernanceHandler handles governance-related HTTP endpoints.
type GovernanceHandler struct {
	service *governance.GovernanceService
}

// NewGovernanceHandler creates a new governance handler.
func NewGovernanceHandler(service *governance.GovernanceService) *GovernanceHandler {
	return &GovernanceHandler{service: service}
}

// GetVotingPower handles GET /api/v1/governance/voting-power?did=xxx
func (h *GovernanceHandler) GetVotingPower(w http.ResponseWriter, r *http.Request) {
	did := r.URL.Query().Get("did")
	if did == "" {
		writeError(w, http.StatusBadRequest, "did is required")
		return
	}

	power, err := h.service.GetVotingPower(r.Context(), did)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, power)
}

// ListActiveProposals handles GET /api/v1/governance/proposals/active
func (h *GovernanceHandler) ListActiveProposals(w http.ResponseWriter, r *http.Request) {
	proposals, err := h.service.ListActiveProposals(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, proposals)
}

// CreateProposal handles POST /api/v1/governance/proposals
func (h *GovernanceHandler) CreateProposal(w http.ResponseWriter, r *http.Request) {
	var req governance.CreateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	proposal, err := h.service.CreateProposal(r.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case governance.ErrInvalidProposalType:
			status = http.StatusBadRequest
		case governance.ErrInvalidThreshold:
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, proposal)
}

// SubmitVote handles POST /api/v1/governance/votes
func (h *GovernanceHandler) SubmitVote(w http.ResponseWriter, r *http.Request) {
	var req governance.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.SubmitVote(r.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case governance.ErrCannotVote:
			status = http.StatusForbidden
		case governance.ErrProposalNotFound, governance.ErrProposalNotActive:
			status = http.StatusNotFound
		case governance.ErrAlreadyVoted:
			status = http.StatusConflict
		case governance.ErrInvalidVoteValue:
			status = http.StatusBadRequest
		case governance.ErrProposalExpired:
			status = http.StatusGone
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetProposalTally handles GET /api/v1/governance/proposals/:id/tally
func (h *GovernanceHandler) GetProposalTally(w http.ResponseWriter, r *http.Request) {
	proposalID := r.URL.Query().Get("proposalId")
	if proposalID == "" {
		writeError(w, http.StatusBadRequest, "proposalId is required")
		return
	}

	tally, err := h.service.GetProposalTally(r.Context(), proposalID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, tally)
}
