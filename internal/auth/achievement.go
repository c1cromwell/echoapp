package auth

import (
	"context"
	"fmt"
	"time"
)

// BadgeType represents different achievement badge types
type BadgeType string

const (
	// Verification badges
	BadgePhoneVerified BadgeType = "phone_verified"
	BadgeEmailVerified BadgeType = "email_verified"
	BadgeKYCVerified   BadgeType = "kyc_verified"
	BadgeAppleIDLinked BadgeType = "apple_id_linked"
	BadgeOrgVerified   BadgeType = "org_verified"

	// Trust badges
	BadgeTrustedUser     BadgeType = "trusted_user"     // 41+ trust score
	BadgeVerifiedUser    BadgeType = "verified_user"    // 61+ trust score
	BadgeEliteUser       BadgeType = "elite_user"       // 81+ trust score
	BadgeTrustAmbassador BadgeType = "trust_ambassador" // 10+ active attestations

	// Activity badges
	BadgeEarlyAdopter     BadgeType = "early_adopter"     // First week user
	BadgeSocialButterfly  BadgeType = "social_butterfly"  // 50+ contacts
	BadgeCommunityBuilder BadgeType = "community_builder" // 5+ groups created
	BadgeDailyStreakFire  BadgeType = "daily_streak_fire" // 30+ day streak
	BadgeWeeklyEngagement BadgeType = "weekly_engagement" // Active 4+ weeks

	// On-chain badges
	BadgeBlockchainBorn     BadgeType = "blockchain_born"     // Made first transaction
	BadgeStakeholder        BadgeType = "stakeholder"         // 1000+ ECHO staked
	BadgeGovernanceVoter    BadgeType = "governance_voter"    // Voted 3+ times
	BadgeReferralMaster     BadgeType = "referral_master"     // 5+ referrals
	BadgeMultiChainExplorer BadgeType = "multichain_explorer" // Used 3+ chains

	// Gamification badges
	BadgePointsCollector BadgeType = "points_collector" // 1000+ total points
	BadgeStreakMaster    BadgeType = "streak_master"    // 50+ day streak
	BadgeRewardRunner    BadgeType = "reward_runner"    // Earned 5000+ ECHO

	// Web of Trust badges
	BadgeVouchPartner    BadgeType = "vouch_partner"    // Vouched 3+ users
	BadgeEndorsement     BadgeType = "endorsement"      // Endorsed 3+ users
	BadgeVerificationPro BadgeType = "verification_pro" // Verified 5+ users
	BadgeTrustNetwork    BadgeType = "trust_network"    // Part of 5+ verification chains
)

// AchievementBadge represents an earned badge
type AchievementBadge struct {
	BadgeID    string
	BadgeType  BadgeType
	UserDID    string
	EarnedAt   time.Time
	ExpiresAt  *time.Time
	UnlockedBy string // Description of what unlocked this badge
	IsActive   bool
}

// AchievementLevel represents progress tiers
type AchievementLevel string

const (
	LevelBronze   AchievementLevel = "bronze"
	LevelSilver   AchievementLevel = "silver"
	LevelGold     AchievementLevel = "gold"
	LevelPlatinum AchievementLevel = "platinum"
)

// BadgeDefinition defines badge unlock criteria
type BadgeDefinition struct {
	Type           BadgeType
	Name           string
	Description    string
	Category       string // verification, trust, activity, onchain, gamification, webtrust
	BaseECHOReward int64
	LevelBonus     map[AchievementLevel]int64 // Bonus rewards by tier
	UnlockCriteria func(context.Context, *TrustScoreSnapshot, []*AchievementBadge) (bool, string, error)
	ExpiryDays     *int // nil = permanent
}

// BadgeEarnEvent tracks when a badge is earned
type BadgeEarnEvent struct {
	BadgeID    string
	UserDID    string
	BadgeType  BadgeType
	EarnedAt   time.Time
	ECHOReward int64
	EventType  string // "earned", "revoked", "claimed"
}

// AchievementConfig holds achievement system configuration
type AchievementConfig struct {
	BadgeDefinitions map[BadgeType]*BadgeDefinition
	MaxBadgesPerUser int
	BadgeExpirySets  map[BadgeType]int // Days until expiry, nil = permanent
}

// AchievementService manages badges and achievements
type AchievementService struct {
	config *AchievementConfig
}

// NewAchievementService creates a new achievement service
func NewAchievementService(config *AchievementConfig) *AchievementService {
	if config == nil {
		config = getDefaultAchievementConfig()
	}
	return &AchievementService{config: config}
}

// CheckBadgeUnlock evaluates if a user unlocks a badge
func (s *AchievementService) CheckBadgeUnlock(
	ctx context.Context,
	badgeType BadgeType,
	trustScore *TrustScoreSnapshot,
	currentBadges []*AchievementBadge,
) (bool, string, error) {

	definition, exists := s.config.BadgeDefinitions[badgeType]
	if !exists {
		return false, "", fmt.Errorf("badge type not found: %s", badgeType)
	}

	// Check if already earned
	for _, badge := range currentBadges {
		if badge.BadgeType == badgeType && badge.IsActive {
			return false, "already earned", nil
		}
	}

	// Check unlock criteria
	return definition.UnlockCriteria(ctx, trustScore, currentBadges)
}

// EarnBadge records that a user earned a badge
func (s *AchievementService) EarnBadge(
	ctx context.Context,
	userDID string,
	badgeType BadgeType,
	level AchievementLevel,
) (*AchievementBadge, *ECHOReward, error) {

	definition, exists := s.config.BadgeDefinitions[badgeType]
	if !exists {
		return nil, nil, fmt.Errorf("badge type not found: %s", badgeType)
	}

	badge := &AchievementBadge{
		BadgeID:   generateBadgeID(),
		BadgeType: badgeType,
		UserDID:   userDID,
		EarnedAt:  time.Now(),
		IsActive:  true,
	}

	// Calculate expiry if defined
	if expiryDays := definition.ExpiryDays; expiryDays != nil {
		expiryTime := time.Now().AddDate(0, 0, *expiryDays)
		badge.ExpiresAt = &expiryTime
	}

	// Calculate ECHO reward with level bonus
	baseReward := definition.BaseECHOReward
	levelBonus := definition.LevelBonus[level]
	totalReward := baseReward + levelBonus

	reward := &ECHOReward{
		Amount:     totalReward,
		Source:     fmt.Sprintf("badge:%s", badgeType),
		Multiplier: 1.0,
		EarnedAt:   time.Now(),
	}

	return badge, reward, nil
}

// RevokeBadge removes a badge from a user
func (s *AchievementService) RevokeBadge(
	ctx context.Context,
	badgeID string,
) error {
	// This would be implemented with actual database
	// For now, just validate the badge ID format
	if badgeID == "" {
		return fmt.Errorf("badge ID cannot be empty")
	}
	return nil
}

// GetBadgeStats returns statistics about user's badges
func (s *AchievementService) GetBadgeStats(
	ctx context.Context,
	badges []*AchievementBadge,
) map[string]interface{} {

	stats := make(map[string]interface{})

	active := 0
	expired := 0
	categoryCount := make(map[string]int)

	for _, badge := range badges {
		if !badge.IsActive {
			continue
		}

		// Check expiry
		if badge.ExpiresAt != nil && time.Now().After(*badge.ExpiresAt) {
			expired++
			continue
		}

		active++

		// Count by category
		definition, exists := s.config.BadgeDefinitions[badge.BadgeType]
		if exists {
			categoryCount[definition.Category]++
		}
	}

	stats["total_badges"] = len(badges)
	stats["active_badges"] = active
	stats["expired_badges"] = expired
	stats["categories"] = categoryCount

	return stats
}

// CalculateAchievementLevel determines a user's achievement tier
func (s *AchievementService) CalculateAchievementLevel(
	ctx context.Context,
	badges []*AchievementBadge,
	totalECHOEarned int64,
) AchievementLevel {

	activeCount := 0
	for _, b := range badges {
		if b.IsActive && (b.ExpiresAt == nil || time.Now().Before(*b.ExpiresAt)) {
			activeCount++
		}
	}

	// Determine level based on badges and earnings
	switch {
	case activeCount >= 15 && totalECHOEarned >= 10000:
		return LevelPlatinum
	case activeCount >= 10 && totalECHOEarned >= 5000:
		return LevelGold
	case activeCount >= 5 && totalECHOEarned >= 2000:
		return LevelSilver
	default:
		return LevelBronze
	}
}

// GetBadgesForCategory returns all badges in a category
func (s *AchievementService) GetBadgesForCategory(
	ctx context.Context,
	category string,
) []*BadgeDefinition {

	var badges []*BadgeDefinition
	for _, def := range s.config.BadgeDefinitions {
		if def.Category == category {
			badges = append(badges, def)
		}
	}
	return badges
}

// EstimateNextBadges suggests which badges a user could earn
func (s *AchievementService) EstimateNextBadges(
	ctx context.Context,
	trustScore *TrustScoreSnapshot,
	currentBadges []*AchievementBadge,
	limit int,
) []BadgeType {

	var nextBadges []BadgeType
	earnedTypes := make(map[BadgeType]bool)

	for _, badge := range currentBadges {
		if badge.IsActive {
			earnedTypes[badge.BadgeType] = true
		}
	}

	// Check each badge definition
	for badgeType, definition := range s.config.BadgeDefinitions {
		if earnedTypes[badgeType] {
			continue // Already earned
		}

		// Check if could earn
		canEarn, _, _ := definition.UnlockCriteria(ctx, trustScore, currentBadges)
		if canEarn && len(nextBadges) < limit {
			nextBadges = append(nextBadges, badgeType)
		}
	}

	return nextBadges
}

// Helper function to get default achievement config
func getDefaultAchievementConfig() *AchievementConfig {
	return &AchievementConfig{
		MaxBadgesPerUser: 50,
		BadgeDefinitions: map[BadgeType]*BadgeDefinition{
			// Verification badges
			BadgePhoneVerified: {
				Type:           BadgePhoneVerified,
				Name:           "Phone Verified",
				Description:    "Verified phone number",
				Category:       "verification",
				BaseECHOReward: 10,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0},
				UnlockCriteria: verifyPhoneCriteria,
			},
			BadgeEmailVerified: {
				Type:           BadgeEmailVerified,
				Name:           "Email Verified",
				Description:    "Verified email address",
				Category:       "verification",
				BaseECHOReward: 10,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0},
				UnlockCriteria: verifyEmailCriteria,
			},
			BadgeKYCVerified: {
				Type:           BadgeKYCVerified,
				Name:           "KYC Verified",
				Description:    "Completed KYC verification",
				Category:       "verification",
				BaseECHOReward: 100,
				LevelBonus: map[AchievementLevel]int64{
					LevelBronze:   0,
					LevelSilver:   50,
					LevelGold:     100,
					LevelPlatinum: 150,
				},
				UnlockCriteria: verifyKYCCriteria,
			},
			BadgeAppleIDLinked: {
				Type:           BadgeAppleIDLinked,
				Name:           "Apple ID Linked",
				Description:    "Linked Apple ID for biometric auth",
				Category:       "verification",
				BaseECHOReward: 50,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 25},
				UnlockCriteria: verifyAppleIDCriteria,
			},
			BadgeOrgVerified: {
				Type:           BadgeOrgVerified,
				Name:           "Organization Verified",
				Description:    "Verified organization account",
				Category:       "verification",
				BaseECHOReward: 200,
				LevelBonus: map[AchievementLevel]int64{
					LevelBronze:   0,
					LevelSilver:   100,
					LevelGold:     200,
					LevelPlatinum: 300,
				},
				UnlockCriteria: verifyOrgCriteria,
			},

			// Trust badges
			BadgeTrustedUser: {
				Type:           BadgeTrustedUser,
				Name:           "Trusted User",
				Description:    "Achieved 41+ trust score",
				Category:       "trust",
				BaseECHOReward: 50,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 25},
				UnlockCriteria: verifyTrustedUserCriteria,
			},
			BadgeVerifiedUser: {
				Type:           BadgeVerifiedUser,
				Name:           "Verified User",
				Description:    "Achieved 61+ trust score",
				Category:       "trust",
				BaseECHOReward: 100,
				LevelBonus: map[AchievementLevel]int64{
					LevelBronze:   0,
					LevelSilver:   50,
					LevelGold:     100,
					LevelPlatinum: 150,
				},
				UnlockCriteria: verifyVerifiedUserCriteria,
			},
			BadgeEliteUser: {
				Type:           BadgeEliteUser,
				Name:           "Elite User",
				Description:    "Achieved 81+ trust score",
				Category:       "trust",
				BaseECHOReward: 200,
				LevelBonus: map[AchievementLevel]int64{
					LevelBronze:   0,
					LevelSilver:   100,
					LevelGold:     200,
					LevelPlatinum: 300,
				},
				UnlockCriteria: verifyEliteUserCriteria,
			},
			BadgeTrustAmbassador: {
				Type:           BadgeTrustAmbassador,
				Name:           "Trust Ambassador",
				Description:    "Created 10+ active attestations",
				Category:       "trust",
				BaseECHOReward: 150,
				LevelBonus: map[AchievementLevel]int64{
					LevelBronze:   0,
					LevelSilver:   75,
					LevelGold:     150,
					LevelPlatinum: 225,
				},
				UnlockCriteria: verifyTrustAmbassadorCriteria,
			},

			// Activity badges
			BadgeEarlyAdopter: {
				Type:           BadgeEarlyAdopter,
				Name:           "Early Adopter",
				Description:    "Joined in first week",
				Category:       "activity",
				BaseECHOReward: 25,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0},
				UnlockCriteria: verifyEarlyAdopterCriteria,
			},
			BadgeSocialButterfly: {
				Type:           BadgeSocialButterfly,
				Name:           "Social Butterfly",
				Description:    "Connected with 50+ contacts",
				Category:       "activity",
				BaseECHOReward: 50,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 25},
				UnlockCriteria: verifySocialButterflyCriteria,
			},
			BadgeCommunityBuilder: {
				Type:           BadgeCommunityBuilder,
				Name:           "Community Builder",
				Description:    "Created 5+ groups",
				Category:       "activity",
				BaseECHOReward: 100,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 50, LevelGold: 100},
				UnlockCriteria: verifyCommunityBuilderCriteria,
			},
			BadgeDailyStreakFire: {
				Type:           BadgeDailyStreakFire,
				Name:           "Daily Streak Fire",
				Description:    "30+ day login streak",
				Category:       "activity",
				BaseECHOReward: 75,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 40, LevelGold: 75},
				UnlockCriteria: verifyDailyStreakFireCriteria,
			},
			BadgeWeeklyEngagement: {
				Type:           BadgeWeeklyEngagement,
				Name:           "Weekly Engagement",
				Description:    "Active 4+ weeks",
				Category:       "activity",
				BaseECHOReward: 30,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 15},
				UnlockCriteria: verifyWeeklyEngagementCriteria,
			},

			// On-chain badges
			BadgeBlockchainBorn: {
				Type:           BadgeBlockchainBorn,
				Name:           "Blockchain Born",
				Description:    "Made first on-chain transaction",
				Category:       "onchain",
				BaseECHOReward: 25,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0},
				UnlockCriteria: verifyBlockchainBornCriteria,
			},
			BadgeStakeholder: {
				Type:           BadgeStakeholder,
				Name:           "Stakeholder",
				Description:    "Staked 1000+ ECHO",
				Category:       "onchain",
				BaseECHOReward: 100,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 50, LevelGold: 100},
				UnlockCriteria: verifyStakeholderCriteria,
			},
			BadgeGovernanceVoter: {
				Type:           BadgeGovernanceVoter,
				Name:           "Governance Voter",
				Description:    "Voted in 3+ governance proposals",
				Category:       "onchain",
				BaseECHOReward: 50,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 25, LevelGold: 50},
				UnlockCriteria: verifyGovernanceVoterCriteria,
			},
			BadgeReferralMaster: {
				Type:           BadgeReferralMaster,
				Name:           "Referral Master",
				Description:    "Referred 5+ users",
				Category:       "onchain",
				BaseECHOReward: 150,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 75, LevelGold: 150},
				UnlockCriteria: verifyReferralMasterCriteria,
			},
			BadgeMultiChainExplorer: {
				Type:           BadgeMultiChainExplorer,
				Name:           "MultiChain Explorer",
				Description:    "Used 3+ different blockchains",
				Category:       "onchain",
				BaseECHOReward: 75,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 40, LevelGold: 75},
				UnlockCriteria: verifyMultiChainExplorerCriteria,
			},

			// Gamification badges
			BadgePointsCollector: {
				Type:           BadgePointsCollector,
				Name:           "Points Collector",
				Description:    "Earned 1000+ total points",
				Category:       "gamification",
				BaseECHOReward: 50,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 25},
				UnlockCriteria: verifyPointsCollectorCriteria,
			},
			BadgeStreakMaster: {
				Type:           BadgeStreakMaster,
				Name:           "Streak Master",
				Description:    "50+ day streak",
				Category:       "gamification",
				BaseECHOReward: 100,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 50, LevelGold: 100},
				UnlockCriteria: verifyStreakMasterCriteria,
			},
			BadgeRewardRunner: {
				Type:           BadgeRewardRunner,
				Name:           "Reward Runner",
				Description:    "Earned 5000+ ECHO",
				Category:       "gamification",
				BaseECHOReward: 200,
				LevelBonus: map[AchievementLevel]int64{
					LevelBronze:   0,
					LevelSilver:   100,
					LevelGold:     200,
					LevelPlatinum: 300,
				},
				UnlockCriteria: verifyRewardRunnerCriteria,
			},

			// Web of Trust badges
			BadgeVouchPartner: {
				Type:           BadgeVouchPartner,
				Name:           "Vouch Partner",
				Description:    "Vouched for 3+ users",
				Category:       "webtrust",
				BaseECHOReward: 50,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 25},
				UnlockCriteria: verifyVouchPartnerCriteria,
			},
			BadgeEndorsement: {
				Type:           BadgeEndorsement,
				Name:           "Endorsement",
				Description:    "Endorsed 3+ users",
				Category:       "webtrust",
				BaseECHOReward: 75,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 40, LevelGold: 75},
				UnlockCriteria: verifyEndorsementCriteria,
			},
			BadgeVerificationPro: {
				Type:           BadgeVerificationPro,
				Name:           "Verification Pro",
				Description:    "Verified 5+ users",
				Category:       "webtrust",
				BaseECHOReward: 150,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 75, LevelGold: 150},
				UnlockCriteria: verifyVerificationProCriteria,
			},
			BadgeTrustNetwork: {
				Type:           BadgeTrustNetwork,
				Name:           "Trust Network",
				Description:    "Part of 5+ verification chains",
				Category:       "webtrust",
				BaseECHOReward: 100,
				LevelBonus:     map[AchievementLevel]int64{LevelBronze: 0, LevelSilver: 50, LevelGold: 100},
				UnlockCriteria: verifyTrustNetworkCriteria,
			},
		},
	}
}

// Badge unlock criteria functions
func verifyPhoneCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Verification.PhoneVerifiedBool {
		return true, "phone verified", nil
	}
	return false, "phone not verified", nil
}

func verifyEmailCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Verification.EmailVerifiedBool {
		return true, "email verified", nil
	}
	return false, "email not verified", nil
}

func verifyKYCCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Verification.KYCLiteVerified > 0 || score.Verification.KYCFullVerified > 0 {
		return true, "kyc verified", nil
	}
	return false, "not kyc verified", nil
}

func verifyAppleIDCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Verification.AppleIDVerified {
		return true, "apple id linked", nil
	}
	return false, "apple id not linked", nil
}

func verifyOrgCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Verification.OrgVerified > 0 {
		return true, "org verified", nil
	}
	return false, "org not verified", nil
}

func verifyTrustedUserCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Score >= 41 {
		return true, "trust score 41+", nil
	}
	return false, fmt.Sprintf("trust score %d, need 41+", score.Score), nil
}

func verifyVerifiedUserCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Score >= 61 {
		return true, "trust score 61+", nil
	}
	return false, fmt.Sprintf("trust score %d, need 61+", score.Score), nil
}

func verifyEliteUserCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Score >= 81 {
		return true, "trust score 81+", nil
	}
	return false, fmt.Sprintf("trust score %d, need 81+", score.Score), nil
}

func verifyTrustAmbassadorCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	// Stub: would check Web of Trust attestation count
	return false, "requires database integration", nil
}

func verifyEarlyAdopterCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	// Stub: would check account creation date
	return false, "requires database integration", nil
}

func verifySocialButterflyCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Behavior.ContactCount >= 50 {
		return true, "50+ contacts", nil
	}
	return false, fmt.Sprintf("%d contacts, need 50+", score.Behavior.ContactCount), nil
}

func verifyCommunityBuilderCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Behavior.GroupsCreated >= 5 {
		return true, "5+ groups created", nil
	}
	return false, fmt.Sprintf("%d groups created, need 5+", score.Behavior.GroupsCreated), nil
}

func verifyDailyStreakFireCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Behavior.DailyLoginStreak >= 30 {
		return true, "30+ day streak", nil
	}
	return false, fmt.Sprintf("%d day streak, need 30+", score.Behavior.DailyLoginStreak), nil
}

func verifyWeeklyEngagementCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	// Stub: would check activity history
	return false, "requires database integration", nil
}

func verifyBlockchainBornCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.OnChain.TransactionCount > 0 {
		return true, "made on-chain transaction", nil
	}
	return false, "no transactions yet", nil
}

func verifyStakeholderCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.OnChain.StakedAmount >= 1000 {
		return true, "1000+ ECHO staked", nil
	}
	return false, fmt.Sprintf("%.0f ECHO staked, need 1000+", score.OnChain.StakedAmount), nil
}

func verifyGovernanceVoterCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.OnChain.GovernanceVoteCount >= 3 {
		return true, "3+ governance votes", nil
	}
	return false, fmt.Sprintf("%d votes, need 3+", score.OnChain.GovernanceVoteCount), nil
}

func verifyReferralMasterCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.OnChain.ReferralCount >= 5 {
		return true, "5+ referrals", nil
	}
	return false, fmt.Sprintf("%d referrals, need 5+", score.OnChain.ReferralCount), nil
}

func verifyMultiChainExplorerCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	// Stub: would check chain usage history
	return false, "requires database integration", nil
}

func verifyPointsCollectorCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	totalPoints := int64(score.Score) + score.OnChain.TransactionCount + int64(score.Behavior.ContactCount) + int64(score.Behavior.GroupsCreated)
	if totalPoints >= 1000 {
		return true, "1000+ total points", nil
	}
	return false, fmt.Sprintf("%d points, need 1000+", totalPoints), nil
}

func verifyStreakMasterCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	if score == nil {
		return false, "no trust score", nil
	}
	if score.Behavior.DailyLoginStreak >= 50 {
		return true, "50+ day streak", nil
	}
	return false, fmt.Sprintf("%d day streak, need 50+", score.Behavior.DailyLoginStreak), nil
}

func verifyRewardRunnerCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	// Stub: would check total ECHO earned
	return false, "requires database integration", nil
}

func verifyVouchPartnerCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	// Stub: would check Web of Trust vouches
	return false, "requires database integration", nil
}

func verifyEndorsementCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	// Stub: would check Web of Trust endorsements
	return false, "requires database integration", nil
}

func verifyVerificationProCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	// Stub: would check Web of Trust verifications
	return false, "requires database integration", nil
}

func verifyTrustNetworkCriteria(ctx context.Context, score *TrustScoreSnapshot, badges []*AchievementBadge) (bool, string, error) {
	// Stub: would check attestation chain participation
	return false, "requires database integration", nil
}

// Helper to generate badge IDs
func generateBadgeID() string {
	return fmt.Sprintf("badge_%d", time.Now().UnixNano())
}
