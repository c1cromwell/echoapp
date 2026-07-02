package datasov

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrNotEnabled is returned when the user has not opted into data sovereignty.
var ErrNotEnabled = errors.New("data sovereignty layer not enabled")

// Contribution is an anonymized stats bundle (WO-249 stub).
type Contribution struct {
	ID        string    `json:"id"`
	DID       string    `json:"did"`
	StatsHash string    `json:"stats_hash"`
	CreatedAt time.Time `json:"created_at"`
}

// Service tracks opt-in state and queued contributions (WO-248/249 MVP).
type Service struct {
	mu            sync.Mutex
	optIn         map[string]bool
	contributions []Contribution
}

// NewService creates an in-memory data sovereignty service.
func NewService() *Service {
	return &Service{optIn: make(map[string]bool)}
}

// SetOptIn records whether a DID participates in the sovereignty layer.
func (s *Service) SetOptIn(did string, enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.optIn[did] = enabled
}

// IsOptedIn returns opt-in status for a DID.
func (s *Service) IsOptedIn(did string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.optIn[did]
}

// Contribute stores an anonymized stats hash when opted in.
func (s *Service) Contribute(did, statsHash string) (Contribution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.optIn[did] {
		return Contribution{}, ErrNotEnabled
	}
	c := Contribution{
		ID:        fmt.Sprintf("ds_%d", time.Now().UnixNano()),
		DID:       did,
		StatsHash: statsHash,
		CreatedAt: time.Now().UTC(),
	}
	s.contributions = append(s.contributions, c)
	return c, nil
}

// Query returns aggregate contribution count (privacy-preserving stub).
func (s *Service) Query() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]interface{}{
		"contribution_count": len(s.contributions),
		"status":             "stub",
	}
}
