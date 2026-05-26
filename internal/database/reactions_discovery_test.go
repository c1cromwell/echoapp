package database

import (
	"context"
	"testing"
)

func TestMemoryDB_Reactions(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	if err := db.AddReaction(ctx, "m1", "did:a", "👍"); err != nil {
		t.Fatal(err)
	}
	if err := db.AddReaction(ctx, "m1", "did:b", "👍"); err != nil {
		t.Fatal(err)
	}
	rows, err := db.GetReactions(ctx, "m1")
	if err != nil || len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d (err=%v)", len(rows), err)
	}

	// Re-react replaces (no duplicate row for the same reactor).
	if err := db.AddReaction(ctx, "m1", "did:a", "❤️"); err != nil {
		t.Fatal(err)
	}
	rows, _ = db.GetReactions(ctx, "m1")
	if len(rows) != 2 {
		t.Fatalf("re-react must replace, not add; want 2 rows, got %d", len(rows))
	}

	if err := db.RemoveReaction(ctx, "m1", "did:a"); err != nil {
		t.Fatal(err)
	}
	rows, _ = db.GetReactions(ctx, "m1")
	if len(rows) != 1 || rows[0].ReactorDID != "did:b" {
		t.Fatalf("after removal want only did:b, got %+v", rows)
	}
}

func TestMemoryDB_DiscoveryIndex(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	if err := db.PutDiscoveryKey(ctx, "oprf-key-1", "did:alice"); err != nil {
		t.Fatal(err)
	}
	if err := db.PutDiscoveryKey(ctx, "oprf-key-2", "did:bob"); err != nil {
		t.Fatal(err)
	}
	// Upsert: same key, new DID overwrites.
	if err := db.PutDiscoveryKey(ctx, "oprf-key-1", "did:alice2"); err != nil {
		t.Fatal(err)
	}

	idx, err := db.AllDiscoveryKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx) != 2 {
		t.Fatalf("want 2 keys, got %d", len(idx))
	}
	if idx["oprf-key-1"] != "did:alice2" {
		t.Fatalf("upsert should overwrite, got %q", idx["oprf-key-1"])
	}

	// Returned map is a copy — mutating it must not affect the store.
	idx["oprf-key-1"] = "tampered"
	again, _ := db.AllDiscoveryKeys(ctx)
	if again["oprf-key-1"] != "did:alice2" {
		t.Fatal("AllDiscoveryKeys must return a defensive copy")
	}
}
