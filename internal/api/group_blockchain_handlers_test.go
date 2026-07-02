package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/groups"
)

func TestGroupBlockchainProof_NotFound(t *testing.T) {
	rt := &Router{V3: &V3Handlers{GroupAnchoring: groups.NewGroupAnchoringService(nil)}}
	req := httptest.NewRequest(http.MethodGet, "/v1/groups/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/blockchain-proof", nil)
	rec := httptest.NewRecorder()
	rt.handleGroupBlockchainProof(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}
