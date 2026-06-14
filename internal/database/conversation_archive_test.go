package database

import (
	"context"
	"testing"
)

func TestMemoryDB_ConversationArchive(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	archived, err := db.IsConversationArchived(ctx, "c1")
	if err != nil {
		t.Fatalf("IsConversationArchived: %v", err)
	}
	if archived {
		t.Fatal("new conversation should not be archived")
	}

	if err := db.SetConversationArchived(ctx, "c1", true); err != nil {
		t.Fatalf("SetConversationArchived true: %v", err)
	}
	archived, err = db.IsConversationArchived(ctx, "c1")
	if err != nil || !archived {
		t.Fatalf("expected archived=true, got archived=%v err=%v", archived, err)
	}

	if err := db.SetConversationArchived(ctx, "c1", false); err != nil {
		t.Fatalf("SetConversationArchived false: %v", err)
	}
	archived, err = db.IsConversationArchived(ctx, "c1")
	if err != nil || archived {
		t.Fatalf("expected archived=false after unarchive, got archived=%v err=%v", archived, err)
	}
}
