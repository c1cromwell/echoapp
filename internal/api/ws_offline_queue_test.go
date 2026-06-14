package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/groups"
)

func TestWSOfflineQueue_EnqueueAndDequeue(t *testing.T) {
	q := newWSOfflineQueue()
	q.Enqueue("did:key:bob", []byte(`{"type":"text"}`), wsOfflineRetention)
	q.Enqueue("did:key:bob", []byte(`{"type":"text2"}`), wsOfflineRetention)

	got := q.DequeueAll("did:key:bob")
	if len(got.Blobs) != 2 {
		t.Fatalf("dequeued %d blobs, want 2", len(got.Blobs))
	}
	if string(got.Blobs[0]) != `{"type":"text"}` {
		t.Fatalf("first = %s", got.Blobs[0])
	}
	if len(q.DequeueAll("did:key:bob").Blobs) != 0 {
		t.Fatal("queue should be empty after dequeue")
	}
}

func TestFlushOffline_DeliversOnReconnect(t *testing.T) {
	h := NewHub()
	bob := &Client{hub: h, userID: "did:key:bob", send: make(chan []byte, 8)}
	h.offlineQueue.Enqueue("did:key:bob", []byte(`{"queued":true}`), wsOfflineRetention)

	h.flushOffline(bob)

	select {
	case data := <-bob.send:
		if string(data) != `{"queued":true}` {
			t.Fatalf("payload = %s", data)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for flushed message")
	}
}

func TestRouteGroupText_QueuesWhenOffline(t *testing.T) {
	gs := groups.NewGroupService()
	_, _ = gs.CreateGroup("grp-off", "did:key:alice", groups.GroupTypePrivate, groups.GroupProfile{
		Name: "Offline", MaxMembers: 10,
	}, groups.VerificationRequirements{ApprovalMode: groups.ApprovalModeAuto})
	_, _ = gs.AddMember("grp-off", "did:key:bob", 10, groups.TrustLevelMember, true)

	h := NewHub()
	h.SetGroupMemberLister(gs)
	alice := &Client{hub: h, userID: "did:key:alice", send: make(chan []byte, 1)}

	payload, _ := json.Marshal(map[string]string{"message_id": "m-off"})
	alice.routeGroupText(WSMessage{
		Type:           "text",
		ConversationID: "group:grp-off",
		Payload:        payload,
	})

	queued := h.offlineQueue.DequeueAll("did:key:bob")
	if len(queued.Blobs) != 1 {
		t.Fatalf("queued %d messages for offline bob, want 1", len(queued.Blobs))
	}
}
