package trustnet

import (
	"math"
	"sync"
	"time"
)

const (
	// DailyDecayRate is the daily trust score decay rate (0.2%)
	DailyDecayRate = 0.002

	// DecayFloorPercent is the minimum score as a percentage of peak (60%)
	DecayFloorPercent = 0.60

	// HistoryRetentionDays is how many days of history to retain
	HistoryRetentionDays = 90
)

// TrustScoreSnapshot represents a point-in-time trust score record
type TrustScoreSnapshot struct {
	UserDID       string
	Score         float64
	PeakScore     float64
	Components    ScoreComponents
	RecordedAt    time.Time
	DecayApplied  float64
	ActivityBonus float64
}

// ScoreComponents breaks down the trust score into weighted parts
type ScoreComponents struct {
	Verification float64 // 30% weight, max 30
	Network      float64 // 25% weight, max 25
	Behavior     float64 // 25% weight, max 25
	Transactions float64 // 20% weight, max 20
}

// Total returns the sum of all components
func (sc ScoreComponents) Total() float64 {
	return sc.Verification + sc.Network + sc.Behavior + sc.Transactions
}

// TrustDecayService manages trust score decay and history
type TrustDecayService struct {
	mu        sync.RWMutex
	scores    map[string]float64              // userDID -> current score
	peaks     map[string]float64              // userDID -> peak score
	history   map[string][]TrustScoreSnapshot // userDID -> history
	lastActivity map[string]time.Time         // userDID -> last activity time
}

// NewTrustDecayService creates a new trust decay service
func NewTrustDecayService() *TrustDecayService {
	return &TrustDecayService{
		scores:       make(map[string]float64),
		peaks:        make(map[string]float64),
		history:      make(map[string][]TrustScoreSnapshot),
		lastActivity: make(map[string]time.Time),
	}
}

// SetScore sets a user's trust score and updates peak
func (s *TrustDecayService) SetScore(userDID string, score float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	s.scores[userDID] = score
	if score > s.peaks[userDID] {
		s.peaks[userDID] = score
	}
}

// GetScore returns the current trust score for a user
func (s *TrustDecayService) GetScore(userDID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scores[userDID]
}

// GetPeakScore returns the peak trust score for a user
func (s *TrustDecayService) GetPeakScore(userDID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.peaks[userDID]
}

// RecordActivity records user activity to pause decay
func (s *TrustDecayService) RecordActivity(userDID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActivity[userDID] = time.Now()
}

// ApplyDecay applies daily decay to a user's score based on inactivity
func (s *TrustDecayService) ApplyDecay(userDID string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	score := s.scores[userDID]
	if score == 0 {
		return 0
	}

	lastAct, hasActivity := s.lastActivity[userDID]
	if !hasActivity {
		// No recorded activity, apply full decay
		return s.applyDecayLocked(userDID, score, 1)
	}

	// Calculate days of inactivity
	daysSinceActivity := time.Since(lastAct).Hours() / 24
	if daysSinceActivity < 1 {
		return 0 // Active today, no decay
	}

	daysToDecay := int(daysSinceActivity)
	return s.applyDecayLocked(userDID, score, daysToDecay)
}

// applyDecayLocked applies decay for a number of days (must hold write lock)
func (s *TrustDecayService) applyDecayLocked(userDID string, score float64, days int) float64 {
	peak := s.peaks[userDID]
	floor := peak * DecayFloorPercent

	totalDecay := 0.0
	for i := 0; i < days; i++ {
		decay := score * DailyDecayRate
		score -= decay
		totalDecay += decay

		if score < floor {
			score = floor
			break
		}
	}

	s.scores[userDID] = math.Round(score*100) / 100 // round to 2 decimal places
	return math.Round(totalDecay*100) / 100
}

// ApplyDecayAll applies decay to all tracked users
func (s *TrustDecayService) ApplyDecayAll() map[string]float64 {
	s.mu.Lock()

	users := make([]string, 0, len(s.scores))
	for uid := range s.scores {
		users = append(users, uid)
	}
	s.mu.Unlock()

	results := make(map[string]float64)
	for _, uid := range users {
		decay := s.ApplyDecay(uid)
		if decay > 0 {
			results[uid] = decay
		}
	}
	return results
}

// RecordSnapshot saves a point-in-time snapshot of the trust score
func (s *TrustDecayService) RecordSnapshot(userDID string, components ScoreComponents) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := TrustScoreSnapshot{
		UserDID:    userDID,
		Score:      s.scores[userDID],
		PeakScore:  s.peaks[userDID],
		Components: components,
		RecordedAt: time.Now(),
	}

	s.history[userDID] = append(s.history[userDID], snapshot)
	s.pruneHistoryLocked(userDID)
}

// GetHistory returns the trust score history for a user
func (s *TrustDecayService) GetHistory(userDID string) []TrustScoreSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := s.history[userDID]
	result := make([]TrustScoreSnapshot, len(history))
	copy(result, history)
	return result
}

// GetHistorySince returns history entries since a specific time
func (s *TrustDecayService) GetHistorySince(userDID string, since time.Time) []TrustScoreSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []TrustScoreSnapshot
	for _, snap := range s.history[userDID] {
		if !snap.RecordedAt.Before(since) {
			result = append(result, snap)
		}
	}
	return result
}

// GetScoreChange returns the net score change over a time period
func (s *TrustDecayService) GetScoreChange(userDID string, since time.Time) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := s.history[userDID]
	if len(history) == 0 {
		return 0
	}

	// Find earliest snapshot at or after 'since'
	var earliest *TrustScoreSnapshot
	for i := range history {
		if !history[i].RecordedAt.Before(since) {
			earliest = &history[i]
			break
		}
	}

	if earliest == nil {
		return 0
	}

	current := s.scores[userDID]
	return math.Round((current-earliest.Score)*100) / 100
}

// pruneHistoryLocked removes snapshots older than retention period
func (s *TrustDecayService) pruneHistoryLocked(userDID string) {
	cutoff := time.Now().AddDate(0, 0, -HistoryRetentionDays)
	history := s.history[userDID]

	firstValid := 0
	for i, snap := range history {
		if !snap.RecordedAt.Before(cutoff) {
			firstValid = i
			break
		}
		if i == len(history)-1 {
			firstValid = len(history)
		}
	}

	if firstValid > 0 {
		s.history[userDID] = history[firstValid:]
	}
}

// CalculateDecayFloor returns the decay floor for a user
func (s *TrustDecayService) CalculateDecayFloor(userDID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.peaks[userDID] * DecayFloorPercent
}
