package messaging

import "github.com/thechadcromwell/echoapp/internal/database"

// ReactionCount is an aggregated reaction for a message.
type ReactionCount struct {
	Emoji    string   `json:"emoji"`
	Count    int      `json:"count"`
	Reactors []string `json:"reactors"` // DIDs that reacted with this emoji
}

// AggregateReactions groups raw reaction rows by emoji with stable ordering
// (descending count, then emoji) so responses are deterministic. The durable
// store is database.ReactionStore; this is the read-model projection over it.
func AggregateReactions(rows []database.ReactionRow) []ReactionCount {
	byEmoji := make(map[string][]string)
	for _, r := range rows {
		byEmoji[r.Emoji] = append(byEmoji[r.Emoji], r.ReactorDID)
	}

	out := make([]ReactionCount, 0, len(byEmoji))
	for emoji, reactors := range byEmoji {
		sortStrings(reactors)
		out = append(out, ReactionCount{Emoji: emoji, Count: len(reactors), Reactors: reactors})
	}
	sortReactions(out)
	return out
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

func sortReactions(r []ReactionCount) {
	for i := 1; i < len(r); i++ {
		for j := i; j > 0 && less(r[j], r[j-1]); j-- {
			r[j-1], r[j] = r[j], r[j-1]
		}
	}
}

func less(a, b ReactionCount) bool {
	if a.Count != b.Count {
		return a.Count > b.Count // higher count first
	}
	return a.Emoji < b.Emoji
}
