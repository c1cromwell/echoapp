package messaging

import (
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func find(list []ReactionCount, emoji string) (ReactionCount, bool) {
	for _, r := range list {
		if r.Emoji == emoji {
			return r, true
		}
	}
	return ReactionCount{}, false
}

func TestAggregateReactions_GroupsAndCounts(t *testing.T) {
	rows := []database.ReactionRow{
		{MessageID: "m1", ReactorDID: "did:alice", Emoji: "👍"},
		{MessageID: "m1", ReactorDID: "did:bob", Emoji: "👍"},
		{MessageID: "m1", ReactorDID: "did:carol", Emoji: "❤️"},
	}
	list := AggregateReactions(rows)
	if thumbs, ok := find(list, "👍"); !ok || thumbs.Count != 2 {
		t.Fatalf("👍 should have count 2, got %+v ok=%v", thumbs, ok)
	}
	if heart, ok := find(list, "❤️"); !ok || heart.Count != 1 {
		t.Fatalf("❤️ should have count 1, got %+v ok=%v", heart, ok)
	}
}

func TestAggregateReactions_OrderingDeterministic(t *testing.T) {
	rows := []database.ReactionRow{
		{ReactorDID: "d1", Emoji: "🙂"},
		{ReactorDID: "d2", Emoji: "👍"},
		{ReactorDID: "d3", Emoji: "👍"},
	}
	list := AggregateReactions(rows)
	if list[0].Emoji != "👍" || list[0].Count != 2 {
		t.Fatalf("expected 👍(2) first, got %+v", list)
	}
	// Reactors within an emoji are sorted.
	if list[0].Reactors[0] != "d2" || list[0].Reactors[1] != "d3" {
		t.Fatalf("reactors should be sorted, got %+v", list[0].Reactors)
	}
}

func TestAggregateReactions_Empty(t *testing.T) {
	if got := AggregateReactions(nil); len(got) != 0 {
		t.Fatalf("nil rows should aggregate to empty, got %d", len(got))
	}
}
