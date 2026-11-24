package messaging

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs against a real database: DATABASE_URL=postgres://... go test ./internal/services/messaging -run DeliveryKeysPostgres
func TestDeliveryKeysPostgresIntegration(t *testing.T) {
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
	migration, err := os.ReadFile("../../../migrations/038_delivery_keys.sql")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ { // twice: the migration must be safe to re-run
		if _, err := pool.Exec(ctx, string(migration)); err != nil {
			t.Fatalf("apply migration: %v", err)
		}
	}
	const bob, mallory = "did:key:zDeliveryKeyItBob", "did:key:zDeliveryKeyItMallory"
	cleanup := func() { _, _ = pool.Exec(ctx, `DELETE FROM delivery_keys WHERE did IN ($1, $2)`, bob, mallory) }
	cleanup()
	defer cleanup()

	store := NewPostgresDeliveryKeyStore(pool)
	hash := func(b byte) string { return HashDeliveryKey(bytes.Repeat([]byte{b, 0x5a}, 16)) }

	if err := store.Register(ctx, bob, hash(1), false); err != nil {
		t.Fatal(err)
	}
	if err := store.Register(ctx, bob, hash(1), false); err != nil {
		t.Fatalf("registering the same key twice: %v", err)
	}
	if did, ok, err := store.Resolve(ctx, hash(1)); err != nil || !ok || did != bob {
		t.Fatalf("resolve: %q %v %v", did, ok, err)
	}
	if err := store.Register(ctx, mallory, hash(1), false); err == nil {
		t.Fatal("another account took over a registered delivery key")
	}
	if did, _, _ := store.Resolve(ctx, hash(1)); did != bob {
		t.Fatalf("after the takeover attempt the key delivers to %q", did)
	}
	if err := store.Register(ctx, bob, "nope", false); err != ErrBadDeliveryKeyHash {
		t.Fatalf("malformed hash: %v", err)
	}
	// Rotation keeps the two newest.
	if _, err := pool.Exec(ctx, `UPDATE delivery_keys SET created_at = NOW() - INTERVAL '2 hours' WHERE key_hash = $1`, hash(1)); err != nil {
		t.Fatal(err)
	}
	if err := store.Register(ctx, bob, hash(2), false); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE delivery_keys SET created_at = NOW() - INTERVAL '1 hour' WHERE key_hash = $1`, hash(2)); err != nil {
		t.Fatal(err)
	}
	if err := store.Register(ctx, bob, hash(3), false); err != nil {
		t.Fatal(err)
	}
	for b, want := range map[byte]bool{1: false, 2: true, 3: true} {
		if _, ok, _ := store.Resolve(ctx, hash(b)); ok != want {
			t.Errorf("key %d valid = %v, want %v", b, ok, want)
		}
	}
}
