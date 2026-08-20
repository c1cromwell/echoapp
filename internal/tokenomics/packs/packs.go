// Package packs implements the weekly reward-pack ritual: a deterministic
// open based on streak already earned — not a prize draw.
package packs

import (
	"context"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/tokenomics/leaderboard"
)

// Item is one standing entry inside a pack (badge / streak label).
type Item struct {
	Kind   string `json:"kind"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// Pack is the current ISO week's pack for a DID.
type Pack struct {
	WeekKey  string     `json:"week_key"`
	Label    string     `json:"label"`
	Opened   bool       `json:"opened"`
	Items    []Item     `json:"items"`
	OpenedAt *time.Time `json:"opened_at,omitempty"`
}

// Service issues weekly packs. Opened-set is in-memory (R0, same as streaks).
type Service struct {
	mu     sync.Mutex
	opened map[string]time.Time // did|weekKey → opened at
}

func NewService() *Service {
	return &Service{opened: make(map[string]time.Time)}
}

func WeekKey(t time.Time) string {
	return leaderboard.BucketKey(leaderboard.WindowWeekly, t)
}

func openKey(did, week string) string { return did + "|" + week }

// Preview builds the pack for did at t without marking it opened.
func (s *Service) Preview(did string, streakDays int, t time.Time) Pack {
	week := WeekKey(t)
	p := build(streakDays, week)
	s.mu.Lock()
	if at, ok := s.opened[openKey(did, week)]; ok {
		p.Opened = true
		p.OpenedAt = &at
	}
	s.mu.Unlock()
	return p
}

// Open marks this week's pack opened. Idempotent.
func (s *Service) Open(_ context.Context, did string, streakDays int, t time.Time) Pack {
	p := s.Preview(did, streakDays, t)
	if p.Opened {
		return p
	}
	now := t.UTC()
	s.mu.Lock()
	s.opened[openKey(did, p.WeekKey)] = now
	s.mu.Unlock()
	p.Opened = true
	p.OpenedAt = &now
	return p
}

func build(streakDays int, week string) Pack {
	label := "Weekly pack"
	items := []Item{{
		Kind:   "activity",
		Title:  "Week recorded",
		Detail: "Your activity this week is counted toward standing.",
	}}
	switch {
	case streakDays >= 30:
		label = "Committed pack"
		items = append(items, Item{
			Kind:   "badge",
			Title:  "Committed",
			Detail: "30-day streak standing.",
		})
	case streakDays >= 7:
		label = "Week One pack"
		items = append(items, Item{
			Kind:   "badge",
			Title:  "Week One",
			Detail: "Seven-day streak standing.",
		})
	}
	return Pack{WeekKey: week, Label: label, Items: items}
}
