package api

import (
	"encoding/json"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/groups"
)

type groupFanoutHub struct {
	*Hub
	delivered map[string][][]byte
}

func newGroupFanoutHub(gs *groups.GroupService) *groupFanoutHub {
	h := NewHub()
	h.SetGroupMemberLister(gs)
	return &groupFanoutHub{Hub: h, delivered: make(map[string][][]byte)}
}

func (h *groupFanoutHub) captureSend(userID string, data []byte) {
	h.delivered[userID] = append(h.delivered[userID], data)
}

func TestRouteGroupText_FansOutToMembers(t *testing.T) {
	gs := groups.NewGroupService()
	_, err := gs.CreateGroup("grp-fan", "did:key:alice", groups.GroupTypePrivate, groups.GroupProfile{
		Name: "Fanout", MaxMembers: 10,
	}, groups.VerificationRequirements{ApprovalMode: groups.ApprovalModeAuto}, groups.TrustLevelVerified)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = gs.AddMember("grp-fan", "did:key:bob", 10, groups.TrustLevelMember, true)
	_, _ = gs.AddMember("grp-fan", "did:key:carol", 10, groups.TrustLevelMember, true)

	hub := newGroupFanoutHub(gs)
	alice := &Client{hub: hub.Hub, userID: "did:key:alice", send: make(chan []byte, 4)}
	bob := &Client{hub: hub.Hub, userID: "did:key:bob", send: make(chan []byte, 4)}
	carol := &Client{hub: hub.Hub, userID: "did:key:carol", send: make(chan []byte, 4)}
	hub.Hub.mu.Lock()
	hub.Hub.clients["did:key:bob"] = bob
	hub.Hub.clients["did:key:carol"] = carol
	hub.Hub.mu.Unlock()

	payload, _ := json.Marshal(map[string]string{"message_id": "m1", "group_ciphertext": "opaque"})
	msg := WSMessage{
		Type:           "text",
		From:           "did:key:alice",
		ConversationID: "group:grp-fan",
		Payload:        payload,
		Timestamp:      "2026-06-12T00:00:00Z",
	}
	alice.routeGroupText(msg)

	if len(bob.send) != 1 {
		t.Fatalf("bob deliveries = %d, want 1", len(bob.send))
	}
	if len(carol.send) != 1 {
		t.Fatalf("carol deliveries = %d, want 1", len(carol.send))
	}

	var out WSMessage
	if err := json.Unmarshal(<-bob.send, &out); err != nil {
		t.Fatal(err)
	}
	if out.To != "did:key:bob" || out.ConversationID != "group:grp-fan" {
		t.Fatalf("bob msg = %+v", out)
	}
}

func TestRouteGroupText_RejectsNonMember(t *testing.T) {
	gs := groups.NewGroupService()
	_, _ = gs.CreateGroup("grp-x", "did:key:admin", groups.GroupTypePrivate, groups.GroupProfile{
		Name: "X", MaxMembers: 10,
	}, groups.VerificationRequirements{ApprovalMode: groups.ApprovalModeAuto}, groups.TrustLevelVerified)

	hub := newGroupFanoutHub(gs)
	outsider := &Client{hub: hub.Hub, userID: "did:key:outsider", send: make(chan []byte, 1)}
	bob := &Client{hub: hub.Hub, userID: "did:key:bob", send: make(chan []byte, 1)}
	hub.Hub.mu.Lock()
	hub.Hub.clients["did:key:bob"] = bob
	hub.Hub.mu.Unlock()

	outsider.routeGroupText(WSMessage{
		Type:           "text",
		ConversationID: "group:grp-x",
		Payload:        json.RawMessage(`{"message_id":"m1"}`),
	})
	if len(bob.send) != 0 {
		t.Fatal("non-member should not fan out")
	}
}
