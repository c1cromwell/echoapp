package api

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/ledger"
)

// A message that has to wait for an offline recipient is stored under the recipient alone: the
// queue row records neither who sent it nor which conversation it belongs to.
func TestQueuedMessageRecordsRecipientOnly(t *testing.T) {
	db := database.NewMemoryDB()
	h := NewHub()
	h.SetDurableOfflineQueue(db)

	envelope, _ := json.Marshal(WSMessage{Type: "text", From: "did:key:zAlice", To: "did:key:zBob", ConversationID: "dm:did:key:zAlice:did:key:zBob"})
	if h.deliverOrQueue("did:key:zBob", envelope, "did:key:zAlice", "dm:did:key:zAlice:did:key:zBob", false) {
		t.Fatal("recipient is offline; the message should have been queued")
	}
	rows, err := db.Dequeue(context.Background(), "did:key:zBob", 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("queued rows: %d, %v", len(rows), err)
	}
	if rows[0].SenderDID != "" {
		t.Errorf("queue row records the sender: %q", rows[0].SenderDID)
	}
	if rows[0].ConversationID != "" {
		t.Errorf("queue row records the conversation: %q", rows[0].ConversationID)
	}
	if rows[0].RecipientDID != "did:key:zBob" {
		t.Errorf("recipient = %q", rows[0].RecipientDID)
	}
}

// A deployment that exists to retain records (Comply) can opt back in, explicitly.
func TestQueuedMessageKeepsMetadataOnlyWhenAskedTo(t *testing.T) {
	t.Setenv("ECHO_RETAIN_MESSAGE_METADATA", "true")
	db := database.NewMemoryDB()
	h := NewHub()
	h.SetDurableOfflineQueue(db)
	h.deliverOrQueue("did:key:zBob", []byte(`{"type":"text"}`), "did:key:zAlice", "dm:x", false)
	rows, _ := db.Dequeue(context.Background(), "did:key:zBob", 10)
	if len(rows) != 1 || rows[0].SenderDID != "did:key:zAlice" || rows[0].ConversationID != "dm:x" {
		t.Fatalf("retention mode did not keep metadata: %+v", rows)
	}
}

// An anchoring confirmation names a message. It is delivered live or not at all: queued for an
// offline sender it would be a stored record tying that sender to that message.
func TestAnchorConfirmationIsNeverQueued(t *testing.T) {
	db := database.NewMemoryDB()
	h := NewHub()
	h.SetDurableOfflineQueue(db)
	if h.PublishConfirmation("did:key:zAlice", ledger.AnchorConfirmation{Type: "confirmation", MessageID: "m-1"}) {
		t.Fatal("the sender is offline; nothing can have been delivered")
	}
	if rows, _ := db.Dequeue(context.Background(), "did:key:zAlice", 10); len(rows) != 0 {
		t.Fatalf("a confirmation was queued for an offline sender: %s", rows[0].Payload)
	}
}

// A sealed message arrives with no message id. The id under which its proof is kept for good
// must not be built from the conversation, which can name both parties.
func TestCommitmentIDNeverContainsTheConversation(t *testing.T) {
	src, err := readFileForTest("ws.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(src, `msg.ConversationID + ":" + msg.Timestamp`) {
		t.Fatal("the commitment id is derived from the conversation id again")
	}
}

func readFileForTest(name string) (string, error) {
	b, err := os.ReadFile(name)
	return string(b), err
}
