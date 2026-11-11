//go:build integration

package database

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run with: DATABASE_URL=postgres://... go test -tags integration ./internal/database/ -run PurgeAccount
func TestPurgeAccountPostgresIntegration(t *testing.T) {
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

	const gone = "did:key:zPurgeIntegrationGone"
	const stays = "did:key:zPurgeIntegrationStays"
	cleanup := func() {
		for _, did := range []string{gone, stays} {
			_, _ = db.PurgeAccount(ctx, did)
		}
		_, _ = pool.Exec(ctx, `DELETE FROM message_queue WHERE message_id LIKE 'purge-it-%'`)
	}
	cleanup()
	defer cleanup()

	seed := []struct {
		q    string
		args []interface{}
	}{
		{`INSERT INTO users (user_id, did, username) VALUES ($1, $2, $3)`, []interface{}{"purge-it-gone", gone, "purge_it_gone"}},
		{`INSERT INTO users (user_id, did, username) VALUES ($1, $2, $3)`, []interface{}{"purge-it-stays", stays, "purge_it_stays"}},
		{`INSERT INTO user_devices (device_id, user_id, did, device_label, public_key, apns_token) VALUES ('purge-it-dev', 'purge-it-gone', $1, 'phone', 'pk', 'apns-token')`, []interface{}{gone}},
		{`INSERT INTO identity_device (did, public_key_hex, device_label) VALUES ($1, $2, 'phone')`, []interface{}{gone, "04" + strings.Repeat("ab", 64)}},
		{`INSERT INTO identity_device (did, public_key_hex, device_label) VALUES ($1, $2, 'phone')`, []interface{}{stays, "04" + strings.Repeat("cd", 64)}},
		{`INSERT INTO sms_recovery (did, phone_hash) VALUES ($1, 'hash-gone')`, []interface{}{gone}},
		{`INSERT INTO contacts (id, owner_did, contact_did) VALUES ('purge-it-c1', $1, $2)`, []interface{}{gone, stays}},
		{`INSERT INTO contacts (id, owner_did, contact_did) VALUES ('purge-it-c2', $2, $1)`, []interface{}{gone, stays}},
		{`INSERT INTO user_prefs (did, prefs) VALUES ($1, '{}')`, []interface{}{gone}},
		{`INSERT INTO user_prefs (did, prefs) VALUES ($1, '{}')`, []interface{}{stays}},
		{`INSERT INTO message_queue (message_id, conversation_id, sender_did, recipient_did, encrypted_payload, status, expires_at)
		  VALUES ('purge-it-in', 'c', $2, $1, 'x', 'queued', NOW() + interval '1 day')`, []interface{}{gone, stays}},
		{`INSERT INTO message_queue (message_id, conversation_id, sender_did, recipient_did, encrypted_payload, status, expires_at)
		  VALUES ('purge-it-out', 'c', $1, $2, 'x', 'queued', NOW() + interval '1 day')`, []interface{}{gone, stays}},
	}
	for _, s := range seed {
		if _, err := pool.Exec(ctx, s.q, s.args...); err != nil {
			t.Fatalf("seed %q: %v", s.q, err)
		}
	}

	removed, err := db.PurgeAccount(ctx, gone)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	for _, table := range []string{"users", "user_devices", "identity_device", "sms_recovery", "contacts", "user_prefs", "message_queue"} {
		if removed[table] == 0 {
			t.Errorf("expected rows removed from %s, got %v", table, removed)
		}
	}

	count := func(q string, args ...interface{}) int {
		var n int
		if err := pool.QueryRow(ctx, q, args...).Scan(&n); err != nil {
			t.Fatalf("count %q: %v", q, err)
		}
		return n
	}
	// Everything personal to the deleted account is gone…
	for _, q := range []string{
		`SELECT count(*) FROM users WHERE did = $1`,
		`SELECT count(*) FROM user_devices WHERE did = $1`,
		`SELECT count(*) FROM identity_device WHERE did = $1`,
		`SELECT count(*) FROM sms_recovery WHERE did = $1`,
		`SELECT count(*) FROM contacts WHERE owner_did = $1 OR contact_did = $1`,
		`SELECT count(*) FROM user_prefs WHERE did = $1`,
		`SELECT count(*) FROM message_queue WHERE recipient_did = $1`,
	} {
		if n := count(q, gone); n != 0 {
			t.Errorf("%s -> %d rows left for the deleted account", q, n)
		}
	}
	// …the other account is untouched, and a message the deleted account had
	// already sent to them is still deliverable.
	if n := count(`SELECT count(*) FROM users WHERE did = $1`, stays); n != 1 {
		t.Errorf("bystander user row count = %d", n)
	}
	if n := count(`SELECT count(*) FROM identity_device WHERE did = $1`, stays); n != 1 {
		t.Errorf("bystander device row count = %d", n)
	}
	if n := count(`SELECT count(*) FROM user_prefs WHERE did = $1`, stays); n != 1 {
		t.Errorf("bystander prefs row count = %d", n)
	}
	if n := count(`SELECT count(*) FROM message_queue WHERE message_id = 'purge-it-out'`); n != 1 {
		t.Errorf("in-flight message to the bystander should remain, got %d", n)
	}

	// Purging again is a no-op, not an error.
	if again, err := db.PurgeAccount(ctx, gone); err != nil || len(again) != 0 {
		t.Errorf("second purge: removed=%v err=%v", again, err)
	}
}
