package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/groups"
)

type fakeGroupKeyPublisher struct {
	msgs []WSMessage
}

func (f *fakeGroupKeyPublisher) PublishSignal(to string, msg WSMessage) bool {
	f.msgs = append(f.msgs, msg)
	return true
}

func TestGroupKeyDistribute_FansOutToMembers(t *testing.T) {
	gs := groups.NewGroupService()
	group, err := gs.CreateGroup("grp-1", "did:key:admin", groups.GroupTypePrivate, groups.GroupProfile{
		Name: "Test Group", MaxMembers: 10,
	}, groups.VerificationRequirements{ApprovalMode: groups.ApprovalModeAuto}, groups.TrustLevelVerified)
	if err != nil {
		t.Fatal(err)
	}
	_ = group
	_, _ = gs.AddMember("grp-1", "did:key:member1", 10, groups.TrustLevelMember, true)
	_, _ = gs.AddMember("grp-1", "did:key:member2", 10, groups.TrustLevelMember, true)

	pub := &fakeGroupKeyPublisher{}
	h := &V3Handlers{Groups: gs, Signals: pub}

	body, _ := json.Marshal(map[string]interface{}{
		"group_id": "grp-1",
		"version":  1,
		"packages": []map[string]interface{}{
			{"to": "did:key:member1", "payload": []byte("pkg-1")},
			{"to": "did:key:member2", "payload": []byte("pkg-2")},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/groups/key/distribute", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:admin"))
	rec := httptest.NewRecorder()
	h.handleGroupKeyDistribute(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(pub.msgs) != 2 {
		t.Fatalf("delivered %d signals, want 2", len(pub.msgs))
	}
	if pub.msgs[0].Type != "group_key" || pub.msgs[0].To != "did:key:member1" {
		t.Fatalf("first signal = %+v", pub.msgs[0])
	}
}

func TestGroupKeyDistribute_RequiresAdmin(t *testing.T) {
	gs := groups.NewGroupService()
	_, _ = gs.CreateGroup("grp-2", "did:key:admin", groups.GroupTypePrivate, groups.GroupProfile{
		Name: "Locked", MaxMembers: 10,
	}, groups.VerificationRequirements{ApprovalMode: groups.ApprovalModeAuto}, groups.TrustLevelVerified)
	_, _ = gs.AddMember("grp-2", "did:key:member", 5, groups.TrustLevelMember, false)

	h := &V3Handlers{Groups: gs, Signals: &fakeGroupKeyPublisher{}}
	body, _ := json.Marshal(map[string]interface{}{
		"group_id": "grp-2",
		"version":  1,
		"packages": []map[string]interface{}{
			{"to": "did:key:member", "payload": []byte("pkg")},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/groups/key/distribute", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:member"))
	rec := httptest.NewRecorder()
	h.handleGroupKeyDistribute(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", rec.Code)
	}
}
