package messaging

import "testing"

func find(list []ReactionCount, emoji string) (ReactionCount, bool) {
	for _, r := range list {
		if r.Emoji == emoji {
			return r, true
		}
	}
	return ReactionCount{}, false
}

func TestReactionStore_AddReplaceRemove(t *testing.T) {
	s := NewReactionStore()

	s.Add("m1", "did:alice", "👍")
	s.Add("m1", "did:bob", "👍")
	s.Add("m1", "did:carol", "❤️")

	list := s.List("m1")
	if thumbs, ok := find(list, "👍"); !ok || thumbs.Count != 2 {
		t.Fatalf("👍 should have count 2, got %+v ok=%v", thumbs, ok)
	}
	if heart, ok := find(list, "❤️"); !ok || heart.Count != 1 {
		t.Fatalf("❤️ should have count 1, got %+v ok=%v", heart, ok)
	}

	// Re-reacting replaces the same reactor's emoji (no double counting).
	s.Add("m1", "did:alice", "❤️")
	list = s.List("m1")
	if thumbs, _ := find(list, "👍"); thumbs.Count != 1 {
		t.Fatalf("after alice switches, 👍 count should be 1, got %d", thumbs.Count)
	}
	if heart, _ := find(list, "❤️"); heart.Count != 2 {
		t.Fatalf("after alice switches, ❤️ count should be 2, got %d", heart.Count)
	}

	// Remove drops the reactor's reaction; removing both ❤️ reactors clears it.
	s.Remove("m1", "did:alice")
	s.Remove("m1", "did:carol")
	list = s.List("m1")
	if _, ok := find(list, "❤️"); ok {
		t.Fatalf("after removing both ❤️ reactors, ❤️ should be gone, got %+v", list)
	}
	if thumbs, _ := find(list, "👍"); thumbs.Count != 1 {
		t.Fatalf("👍 should remain at 1 (bob), got %d", thumbs.Count)
	}
}

func TestReactionStore_OrderingDeterministic(t *testing.T) {
	s := NewReactionStore()
	s.Add("m", "d1", "🙂")
	s.Add("m", "d2", "👍")
	s.Add("m", "d3", "👍")
	list := s.List("m")
	// Higher count first.
	if list[0].Emoji != "👍" || list[0].Count != 2 {
		t.Fatalf("expected 👍(2) first, got %+v", list)
	}
}

func TestReactionStore_EmptyMessage(t *testing.T) {
	s := NewReactionStore()
	if got := s.List("nope"); len(got) != 0 {
		t.Fatalf("empty message should have no reactions, got %d", len(got))
	}
}
