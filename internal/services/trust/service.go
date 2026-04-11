package trust

import (
	"context"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/auth"
)

type ScoreRecord struct {
	DID        string               `json:"did"`
	Score      int                  `json:"score"`
	Tier       int                  `json:"tier"`
	Level      auth.TrustScoreLevel `json:"level"`
	Multiplier float64              `json:"multiplier"`
	UpdatedAt  time.Time            `json:"updatedAt"`
	ExpiresAt  time.Time            `json:"expiresAt"`
}

type Report struct {
	ReporterDID string    `json:"reporterDid"`
	TargetDID   string    `json:"targetDid"`
	ReportType  string    `json:"reportType"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

type Service struct {
	mu       sync.RWMutex
	scorer   *auth.TrustScoreService
	cache    map[string]*ScoreRecord
	reports  []Report
	cacheTTL time.Duration
}

func NewService() *Service {
	config := DefaultConfig()
	return &Service{
		scorer:   auth.NewTrustScoreService(config),
		cache:    make(map[string]*ScoreRecord),
		cacheTTL: 60 * time.Second,
	}
}

func (s *Service) GetScore(ctx context.Context, did string) (*ScoreRecord, error) {
	s.mu.RLock()
	cached, ok := s.cache[did]
	s.mu.RUnlock()

	if ok && time.Now().Before(cached.ExpiresAt) {
		return cached, nil
	}

	vp := s.getVerificationPoints(did)
	bp := s.getBehaviorPoints(did)
	ocp := s.getOnChainPoints(did)
	pp := s.getPenaltyPoints(did)

	snapshot, err := s.scorer.CalculateTrustScore(ctx, vp, bp, ocp, pp)
	if err != nil {
		return nil, err
	}

	record := &ScoreRecord{
		DID:        did,
		Score:      snapshot.Score,
		Tier:       tierFromLevel(snapshot.Level),
		Level:      snapshot.Level,
		Multiplier: snapshot.EarningMultiplier,
		UpdatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(s.cacheTTL),
	}

	s.mu.Lock()
	s.cache[did] = record
	s.mu.Unlock()

	return record, nil
}

func (s *Service) GetScoreBatch(ctx context.Context, dids []string) ([]*ScoreRecord, error) {
	records := make([]*ScoreRecord, 0, len(dids))
	for _, did := range dids {
		record, err := s.GetScore(ctx, did)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *Service) SubmitReport(report Report) error {
	report.Timestamp = time.Now()
	s.mu.Lock()
	s.reports = append(s.reports, report)
	delete(s.cache, report.TargetDID)
	s.mu.Unlock()
	return nil
}

func (s *Service) GetReports(targetDID string) []Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Report
	for _, r := range s.reports {
		if r.TargetDID == targetDID {
			result = append(result, r)
		}
	}
	return result
}

func (s *Service) InvalidateCache(did string) {
	s.mu.Lock()
	delete(s.cache, did)
	s.mu.Unlock()
}

func (s *Service) getVerificationPoints(did string) *auth.VerificationPoints {
	return &auth.VerificationPoints{PasskeyCreated: 1}
}

func (s *Service) getBehaviorPoints(did string) *auth.BehaviorPoints {
	return &auth.BehaviorPoints{}
}

func (s *Service) getOnChainPoints(did string) *auth.OnChainPoints {
	return &auth.OnChainPoints{}
}

func (s *Service) getPenaltyPoints(did string) *auth.PenaltyPoints {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pp := &auth.PenaltyPoints{}
	for _, r := range s.reports {
		if r.TargetDID == did {
			switch r.ReportType {
			case "spam":
				pp.SpamReports++
			case "fraud":
				pp.FraudReports++
			case "abuse":
				pp.SpamReports++
			}
		}
	}
	return pp
}

func tierFromLevel(level auth.TrustScoreLevel) int {
	switch level {
	case auth.TrustLevelElite:
		return 5
	case auth.TrustLevelVerified:
		return 4
	case auth.TrustLevelTrusted:
		return 3
	case auth.TrustLevelBasic:
		return 2
	default:
		return 1
	}
}

func DefaultConfig() auth.TrustScoreConfig {
	config := auth.TrustScoreConfig{}
	config.VerificationWeights.PasskeyCreated = 5
	config.VerificationWeights.PhoneVerified = 5
	config.VerificationWeights.EmailVerified = 5
	config.VerificationWeights.KYCLite = 10
	config.VerificationWeights.FullKYC = 15
	config.VerificationWeights.AppleDigitalID = 15
	config.VerificationWeights.OrgVerified = 20
	config.VerificationWeights.MaxVerification = 70
	config.BehaviorWeights.AccountAgePerMonth = 1
	config.BehaviorWeights.MessagesPerHundred = 1
	config.BehaviorWeights.ContactsPerTen = 3
	config.BehaviorWeights.GroupsPerFive = 1
	config.BehaviorWeights.DailyLoginStreak = 1
	config.BehaviorWeights.MaxBehavior = 80
	config.OnChainWeights.TransactionsPerTx = 1
	config.OnChainWeights.StakingPer100ECHO = 1
	config.OnChainWeights.GovernancePerVote = 1
	config.OnChainWeights.MaxOnChain = 70
	config.PenaltyWeights.SpamReport = 2
	config.PenaltyWeights.FraudReport = 5
	config.PenaltyWeights.BlockedByUser = 1
	config.PenaltyWeights.InactivityPerMonth = 1
	config.PenaltyWeights.MaxPenalty = 20
	config.ECHORewards.VerificationRewards = map[string]int64{
		"passkey": 10, "phone": 15, "email": 10,
		"kyc_lite": 50, "kyc_full": 100, "apple_id": 100, "org": 200,
	}
	config.ECHORewards.BehaviorRewards = map[string]int64{
		"daily_login": 1, "contact": 2, "group": 5,
	}
	config.ECHORewards.OnChainRewards = map[string]int64{
		"transaction": 1, "staking": 1, "governance": 10,
	}
	config.ECHORewards.PenaltyReductions = map[string]int64{
		"spam": 10, "fraud": 50,
	}
	config.MultiplierTiers = map[auth.TrustScoreLevel]float64{
		auth.TrustLevelNewcomer: 1.0,
		auth.TrustLevelBasic:    1.1,
		auth.TrustLevelTrusted:  1.25,
		auth.TrustLevelVerified: 1.5,
		auth.TrustLevelElite:    2.0,
	}
	return config
}
