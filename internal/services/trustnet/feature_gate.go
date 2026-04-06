package trustnet

import (
	"fmt"
	"sync"
)

// FeatureTier represents the trust level required for features
type FeatureTier string

const (
	FeatureTierUnverified FeatureTier = "unverified"
	FeatureTierNewcomer   FeatureTier = "newcomer"
	FeatureTierMember     FeatureTier = "member"
	FeatureTierTrusted    FeatureTier = "trusted"
	FeatureTierVerified   FeatureTier = "verified"
)

// Feature defines a trust-gated feature
type Feature struct {
	Name           string
	RequiredTier   FeatureTier
	RequiredScore  float64
	RateLimitDaily int // 0 = unlimited
}

// TrustTierThresholds maps score ranges to feature tiers
var TrustTierThresholds = map[FeatureTier]struct{ Min, Max float64 }{
	FeatureTierUnverified: {Min: 0, Max: 19},
	FeatureTierNewcomer:   {Min: 20, Max: 39},
	FeatureTierMember:     {Min: 40, Max: 59},
	FeatureTierTrusted:    {Min: 60, Max: 79},
	FeatureTierVerified:   {Min: 80, Max: 100},
}

// DefaultFeatures defines trust gating for core features
var DefaultFeatures = map[string]Feature{
	"messaging": {
		Name:          "messaging",
		RequiredTier:  FeatureTierUnverified,
		RequiredScore: 0,
	},
	"create_group": {
		Name:          "create_group",
		RequiredTier:  FeatureTierNewcomer,
		RequiredScore: 20,
	},
	"group_size_50": {
		Name:          "group_size_50",
		RequiredTier:  FeatureTierMember,
		RequiredScore: 40,
	},
	"group_size_200": {
		Name:          "group_size_200",
		RequiredTier:  FeatureTierTrusted,
		RequiredScore: 60,
	},
	"unlimited_groups": {
		Name:          "unlimited_groups",
		RequiredTier:  FeatureTierVerified,
		RequiredScore: 80,
	},
	"voice_calls": {
		Name:          "voice_calls",
		RequiredTier:  FeatureTierMember,
		RequiredScore: 40,
	},
	"video_calls": {
		Name:          "video_calls",
		RequiredTier:  FeatureTierTrusted,
		RequiredScore: 60,
	},
	"endorse_others": {
		Name:          "endorse_others",
		RequiredTier:  FeatureTierTrusted,
		RequiredScore: 60,
	},
	"file_transfer": {
		Name:          "file_transfer",
		RequiredTier:  FeatureTierMember,
		RequiredScore: 40,
	},
	"schedule_messages": {
		Name:          "schedule_messages",
		RequiredTier:  FeatureTierMember,
		RequiredScore: 40,
	},
	"disappearing_messages": {
		Name:          "disappearing_messages",
		RequiredTier:  FeatureTierTrusted,
		RequiredScore: 60,
	},
	"broadcast_channel_creation": {
		Name:          "broadcast_channel_creation",
		RequiredTier:  FeatureTierMember,
		RequiredScore: 40,
	},
	"broadcast_20_channels": {
		Name:          "broadcast_20_channels",
		RequiredTier:  FeatureTierMember,
		RequiredScore: 40,
	},
	"broadcast_unlimited": {
		Name:          "broadcast_unlimited",
		RequiredTier:  FeatureTierVerified,
		RequiredScore: 80,
	},
}

// FeatureGateService manages trust-based feature access
type FeatureGateService struct {
	mu               sync.RWMutex
	features         map[string]Feature
	customRules      map[string]map[string]interface{} // feature -> user-specific overrides
	featureUsage     map[string]map[string]int         // feature -> userDID -> count today
	usageLastResetAt map[string]map[string]int64       // feature -> userDID -> unix timestamp
}

// NewFeatureGateService creates a new feature gate service
func NewFeatureGateService() *FeatureGateService {
	fgs := &FeatureGateService{
		features:         make(map[string]Feature),
		customRules:      make(map[string]map[string]interface{}),
		featureUsage:     make(map[string]map[string]int),
		usageLastResetAt: make(map[string]map[string]int64),
	}

	// Initialize with default features
	for name, feature := range DefaultFeatures {
		fgs.features[name] = feature
	}

	return fgs
}

// CanAccessFeature checks if a user can access a feature
func (fgs *FeatureGateService) CanAccessFeature(userDID string, featureName string, currentScore float64) (bool, string) {
	fgs.mu.RLock()
	defer fgs.mu.RUnlock()

	feature, exists := fgs.features[featureName]
	if !exists {
		return false, fmt.Sprintf("feature %s not found", featureName)
	}

	// Check trust score requirement
	if currentScore < feature.RequiredScore {
		tier := fgs.getTierForScore(currentScore)
		requiredTier := feature.RequiredTier
		return false, fmt.Sprintf("requires trust tier %s (you have %s, score %.1f/%.0f)",
			requiredTier, tier, currentScore, feature.RequiredScore)
	}

	// Check rate limit if applicable
	if feature.RateLimitDaily > 0 {
		usage := fgs.featureUsage[featureName][userDID]
		if usage >= feature.RateLimitDaily {
			return false, fmt.Sprintf("daily limit reached (%d/%d)", usage, feature.RateLimitDaily)
		}
	}

	return true, ""
}

// RecordFeatureUsage increments the usage counter for a feature
func (fgs *FeatureGateService) RecordFeatureUsage(userDID string, featureName string) {
	fgs.mu.Lock()
	defer fgs.mu.Unlock()

	if fgs.featureUsage[featureName] == nil {
		fgs.featureUsage[featureName] = make(map[string]int)
	}

	fgs.featureUsage[featureName][userDID]++
}

// GetFeatureTier returns the trust tier for a given score
func (fgs *FeatureGateService) GetFeatureTier(score float64) FeatureTier {
	fgs.mu.RLock()
	defer fgs.mu.RUnlock()

	return fgs.getTierForScore(score)
}

// getTierForScore is internal helper to get tier from score
func (fgs *FeatureGateService) getTierForScore(score float64) FeatureTier {
	if score < 20 {
		return FeatureTierUnverified
	}
	if score < 40 {
		return FeatureTierNewcomer
	}
	if score < 60 {
		return FeatureTierMember
	}
	if score < 80 {
		return FeatureTierTrusted
	}
	return FeatureTierVerified
}

// GetFeatureInfo returns info about a feature
func (fgs *FeatureGateService) GetFeatureInfo(featureName string) (Feature, error) {
	fgs.mu.RLock()
	defer fgs.mu.RUnlock()

	feature, exists := fgs.features[featureName]
	if !exists {
		return Feature{}, ErrCircleNotFound // reuse error for simplicity
	}

	return feature, nil
}

// RegisterCustomFeature adds a custom feature with trust gating
func (fgs *FeatureGateService) RegisterCustomFeature(feature Feature) {
	fgs.mu.Lock()
	defer fgs.mu.Unlock()

	fgs.features[feature.Name] = feature
	fgs.featureUsage[feature.Name] = make(map[string]int)
	fgs.usageLastResetAt[feature.Name] = make(map[string]int64)
}

// SetFeatureOverride creates a user-specific override for a feature
func (fgs *FeatureGateService) SetFeatureOverride(featureName string, userDID string, override map[string]interface{}) {
	fgs.mu.Lock()
	defer fgs.mu.Unlock()

	if fgs.customRules[featureName] == nil {
		fgs.customRules[featureName] = make(map[string]interface{})
	}

	fgs.customRules[featureName][userDID] = override
}

// GetRemainingUsageToday returns how many times a user can still use a feature today
func (fgs *FeatureGateService) GetRemainingUsageToday(userDID string, featureName string) int {
	fgs.mu.RLock()
	defer fgs.mu.RUnlock()

	feature, exists := fgs.features[featureName]
	if !exists || feature.RateLimitDaily == 0 {
		return -1 // unlimited
	}

	used := fgs.featureUsage[featureName][userDID]
	remaining := feature.RateLimitDaily - used

	if remaining < 0 {
		return 0
	}

	return remaining
}

// GetFeaturesForTier returns all features accessible at a given tier
func (fgs *FeatureGateService) GetFeaturesForTier(tier FeatureTier) []Feature {
	fgs.mu.RLock()
	defer fgs.mu.RUnlock()

	var features []Feature
	scoreMin := TrustTierThresholds[tier].Min

	for _, feature := range fgs.features {
		if feature.RequiredScore <= scoreMin {
			features = append(features, feature)
		}
	}

	return features
}
