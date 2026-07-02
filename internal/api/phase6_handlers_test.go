package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/broadcast_channels"
	"github.com/thechadcromwell/echoapp/internal/services/groups"
)

func TestBroadcastList_Empty(t *testing.T) {
	h := &V3Handlers{Broadcasts: broadcast_channels.NewChannelService()}
	req := httptest.NewRequest(http.MethodGet, "/v3/broadcasts/list", nil)
	rec := httptest.NewRecorder()
	h.handleBroadcastList(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGroupMute_RequiresAdmin(t *testing.T) {
	gs := groups.NewGroupService()
	_, _ = gs.CreateGroup(
		"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"did:key:owner",
		groups.GroupTypePrivate,
		groups.GroupProfile{Name: "g"},
		groups.VerificationRequirements{},
		groups.TrustLevelMember,
	)
	h := &V3Handlers{Groups: gs}
	body := `{"groupId":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","memberId":"did:key:bob","duration_hours":24}`
	req := httptest.NewRequest(http.MethodPost, "/v3/groups/members/mute", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	h.handleGroupMuteMember(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", rec.Code)
	}
}
