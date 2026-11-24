package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs against a real database: DATABASE_URL=postgres://... go test ./internal/database -run QueuePrivacy
//
// What the message queue may hold about a message, before and after it is delivered.
func TestQueuePrivacyPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	db := &PostgresDB{pool: pool}

	const bob = "did:key:zQueuePrivacyBob"
	cleanup := func() { _, _ = pool.Exec(ctx, `DELETE FROM message_queue WHERE recipient_did = $1`, bob) }
	cleanup()
	defer cleanup()

	enqueue := func(id string) {
		t.Helper()
		if err := db.Enqueue(ctx, &QueuedMessage{
			MessageID: id, ConversationID: "dm:alice:bob", SenderDID: "did:key:zAlice", RecipientDID: bob,
			Payload: []byte(`{"from":"did:key:zAlice","payload":"ciphertext"}`),
		}); err != nil {
			t.Fatal(err)
		}
	}
	enqueue("qp-1")
	enqueue("qp-2")

	// The recipient gets the messages exactly as they were queued, oldest first.
	got, err := db.Dequeue(ctx, bob, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].MessageID != "qp-1" || got[1].MessageID != "qp-2" {
		t.Fatalf("dequeued %+v", got)
	}
	if string(got[0].Payload) != `{"from":"did:key:zAlice","payload":"ciphertext"}` {
		t.Fatalf("delivered payload was altered: %s", got[0].Payload)
	}

	// What is left behind names the recipient and nothing else.
	var sender, conversation, status string
	var payloadLen int
	var deliveredAt *time.Time
	if err := pool.QueryRow(ctx, `SELECT sender_did, conversation_id, octet_length(encrypted_payload), status, delivered_at
		FROM message_queue WHERE message_id = 'qp-1'`).Scan(&sender, &conversation, &payloadLen, &status, &deliveredAt); err != nil {
		t.Fatal(err)
	}
	if sender != "" || conversation != "" || payloadLen != 0 {
		t.Fatalf("after delivery the row still holds sender=%q conversation=%q payload=%d bytes", sender, conversation, payloadLen)
	}
	if status != "delivered" || deliveredAt == nil {
		t.Fatalf("status=%q delivered_at=%v", status, deliveredAt)
	}
	// Delivery status still works from the stub.
	if meta, err := db.GetMessageMeta(ctx, "qp-1"); err != nil || meta.Status != "delivered" {
		t.Fatalf("message meta after delivery: %+v, %v", meta, err)
	}
	// A second dequeue returns nothing: delivered messages are not handed out again.
	if again, _ := db.Dequeue(ctx, bob, 10); len(again) != 0 {
		t.Fatalf("delivered messages were dequeued again: %d", len(again))
	}

	// The stub is deleted a day after delivery; a fresh one is not.
	if _, err := pool.Exec(ctx, `UPDATE message_queue SET delivered_at = NOW() - INTERVAL '25 hours' WHERE message_id = 'qp-1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.PurgeExpired(ctx); err != nil {
		t.Fatal(err)
	}
	var left int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM message_queue WHERE recipient_did = $1`, bob).Scan(&left); err != nil {
		t.Fatal(err)
	}
	if left != 1 {
		t.Fatalf("%d rows left after purge, want 1 (the stub delivered less than a day ago)", left)
	}

	// An undelivered message is kept for the offline retention period, not the old 30 days.
	enqueue("qp-3")
	var ttl time.Duration
	var expires, queued time.Time
	if err := pool.QueryRow(ctx, `SELECT expires_at, queued_at FROM message_queue WHERE message_id = 'qp-3'`).Scan(&expires, &queued); err != nil {
		t.Fatal(err)
	}
	ttl = expires.Sub(queued)
	if ttl < 13*24*time.Hour || ttl > 15*24*time.Hour {
		t.Fatalf("undelivered retention is %s, want 14 days", ttl)
	}
}
