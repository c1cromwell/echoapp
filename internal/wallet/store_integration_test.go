package wallet

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs against a real database: DATABASE_URL=postgres://... go test ./internal/wallet -run PG
func TestPGStoreIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	const alice, mallory = "did:key:zRecoveryItAlice", "did:key:zRecoveryItMallory"
	const addr1, addr2 = "DAG4recoveryItAddressOne000000000000000", "DAG4recoveryItAddressTwo000000000000000"
	cleanup := func() { _, _ = pool.Exec(ctx, `DELETE FROM wallet_accounts WHERE did IN ($1, $2)`, alice, mallory) }
	cleanup()
	defer cleanup()

	store := NewPGStore(pool)
	if _, err := store.AccountFor(ctx, addr1); !errors.Is(err, ErrNotLinked) {
		t.Fatalf("unlinked address: %v", err)
	}
	if err := store.Link(ctx, alice, addr1); err != nil {
		t.Fatal(err)
	}
	if err := store.Link(ctx, alice, addr1); err != nil {
		t.Fatalf("linking again for the same account: %v", err)
	}
	if did, err := store.AccountFor(ctx, addr1); err != nil || did != alice {
		t.Fatalf("AccountFor: %q %v", did, err)
	}
	if err := store.Link(ctx, mallory, addr1); !errors.Is(err, ErrLinkedElsewhere) {
		t.Fatalf("second account claiming the address: %v", err)
	}
	if did, _ := store.AccountFor(ctx, addr1); did != alice {
		t.Fatalf("after the refused claim the address belongs to %q", did)
	}
	// An account that links a new address gives up the old one.
	if err := store.Link(ctx, alice, addr2); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AccountFor(ctx, addr1); !errors.Is(err, ErrNotLinked) {
		t.Fatalf("the replaced address still resolves: %v", err)
	}
	if did, _ := store.AccountFor(ctx, addr2); did != alice {
		t.Fatalf("new address belongs to %q", did)
	}
}

func TestMemStore(t *testing.T) {
	ctx := context.Background()
	m := NewMemStore()
	if err := m.Link(ctx, "a", "x"); err != nil {
		t.Fatal(err)
	}
	if err := m.Link(ctx, "b", "x"); !errors.Is(err, ErrLinkedElsewhere) {
		t.Fatalf("second account: %v", err)
	}
	if err := m.Link(ctx, "a", "y"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AccountFor(ctx, "x"); !errors.Is(err, ErrNotLinked) {
		t.Fatalf("replaced address still resolves: %v", err)
	}
	if err := m.Link(ctx, "", "z"); err == nil {
		t.Fatal("empty account accepted")
	}
}
