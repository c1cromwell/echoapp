# Echo Authentication & Identity Blueprint Review

## Executive Summary

This review analyzes the User Authentication & Identity Verification blueprint, identifying gaps and proposing enhancements focused on:
1. **Maximum Decentralization** - Reducing reliance on centralized services
2. **ECHO Token Incentives** - Gamified rewards tied to Trust Score
3. **Trust Score Gamification** - Engaging mechanics to drive authentic behavior

**Overall Assessment**: 🟢 Strong Foundation with Enhancement Opportunities

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    AUTHENTICATION ARCHITECTURE OVERVIEW                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     DECENTRALIZED IDENTITY LAYER                     │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐        │   │
│  │  │   DID     │  │ Verifiable│  │   Trust   │  │    ZKP    │        │   │
│  │  │  (Prism)  │  │Credentials│  │   Score   │  │  Proofs   │        │   │
│  │  └───────────┘  └───────────┘  └───────────┘  └───────────┘        │   │
│  │                        │                                             │   │
│  │                        ▼                                             │   │
│  │              ┌─────────────────┐                                    │   │
│  │              │     CARDANO     │ ← Immutable Identity Anchor        │   │
│  │              └─────────────────┘                                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      DEVICE SECURITY LAYER                           │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐                       │   │
│  │  │  Passkey  │  │  Secure   │  │  Device   │                       │   │
│  │  │ (WebAuthn)│  │  Enclave  │  │   Trust   │                       │   │
│  │  └───────────┘  └───────────┘  └───────────┘                       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      INCENTIVE LAYER (NEW)                           │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐        │   │
│  │  │   ECHO    │  │   Trust   │  │Achievement│  │  Staking  │        │   │
│  │  │  Rewards  │  │Multipliers│  │  Badges   │  │  Bonuses  │        │   │
│  │  └───────────┘  └───────────┘  └───────────┘  └───────────┘        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Part 1: Blueprint Strengths ✅

### What's Working Well

| Area | Implementation | Assessment |
|------|----------------|------------|
| **DID Architecture** | Atala PRISM on Cardano | ✅ Excellent - W3C compliant, decentralized |
| **Passkey Security** | iOS Secure Enclave | ✅ Excellent - Hardware-backed, never transmitted |
| **Trust Scoring** | Multi-factor 0-100 scale | ✅ Good - Comprehensive components |
| **ZKP Integration** | zk-SNARKs for privacy | ✅ Excellent - Privacy-preserving |
| **Multi-Device Support** | QR-based registration | ✅ Good - User-friendly flow |
| **Credential Revocation** | On-chain registry | ✅ Good - Transparent and auditable |

---

## Part 2: Identified Gaps 🔴

### Gap 1: Centralization Risk in Verification Services (HIGH)

**Current State**: Heavy reliance on centralized third-party services (Prove, Daon, Alloy, Darwinium).

**Risk**: Single points of failure, privacy concerns, vendor lock-in.

**Recommendation**: Implement decentralized verification alternatives.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│              DECENTRALIZED VERIFICATION HIERARCHY                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  TIER 1: Fully Decentralized (Preferred)                                    │
│  ─────────────────────────────────────────                                  │
│  • Web of Trust Attestations (peer vouching)                                │
│  • Self-Sovereign Identity Proofs                                           │
│  • Cross-Chain Identity Bridges (Polygon ID, etc.)                          │
│  • Decentralized Oracle Attestations (Chainlink)                            │
│                                                                              │
│  TIER 2: Federated Decentralized                                            │
│  ─────────────────────────────────────────                                  │
│  • Multiple independent verification providers                              │
│  • Threshold verification (2-of-3 providers must agree)                     │
│  • Provider reputation tracking on-chain                                    │
│                                                                              │
│  TIER 3: Centralized with Decentralized Anchoring                           │
│  ─────────────────────────────────────────                                  │
│  • Third-party services (Prove, Daon, Alloy)                                │
│  • Results anchored to Cardano immediately                                  │
│  • Provider accountability through staking                                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Gap 2: Trust Score Manipulation Prevention (HIGH)

**Current State**: Trust score components are defined but no anti-gaming measures.

**Risk**: Users could artificially inflate scores through fake interactions.

**Recommendation**: Implement anti-sybil and anti-gaming mechanisms.

### Gap 3: Account Recovery Decentralization (MEDIUM)

**Current State**: "Passkey reset with identity verification or account recovery" - details missing.

**Risk**: Centralized recovery could be a backdoor for account takeover.

**Recommendation**: Implement decentralized social recovery.

### Gap 4: Trust Score Decay (MEDIUM)

**Current State**: No mention of score decay for inactive accounts.

**Risk**: Stale accounts could maintain high trust despite inactivity.

**Recommendation**: Implement time-based trust decay with reactivation incentives.

### Gap 5: ECHO Incentive Depth (MEDIUM)

**Current State**: Only 100 ECHO for verification mentioned.

**Opportunity**: Expand gamified incentives throughout the trust journey.

### Gap 6: Reputation Portability (LOW)

**Current State**: DIDs are portable but no cross-app reputation.

**Opportunity**: Enable trust score visibility across ecosystem apps.

---

## Part 3: Enhanced Trust Score with ECHO Incentives

### 3.1 Gamified Trust Score Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    GAMIFIED TRUST SCORE SYSTEM                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  BASE TRUST SCORE (0-100)                                                   │
│  ═══════════════════════                                                    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  VERIFICATION (0-30 pts)    │  ECHO REWARDS                         │   │
│  │  ──────────────────────────────────────────────────────────────────│   │
│  │  • Passkey Created: +5      │  +10 ECHO                             │   │
│  │  • Phone Verified: +5       │  +15 ECHO                             │   │
│  │  • Email Verified: +5       │  +10 ECHO                             │   │
│  │  • KYC-Lite: +10            │  +50 ECHO                             │   │
│  │  • Full KYC: +15            │  +100 ECHO                            │   │
│  │  • Apple Digital ID: +15    │  +100 ECHO                            │   │
│  │  • Org Verified: +20        │  +200 ECHO + Badge                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  BEHAVIOR (0-30 pts)        │  ECHO REWARDS                         │   │
│  │  ──────────────────────────────────────────────────────────────────│   │
│  │  • Account Age (1pt/mo)     │  +5 ECHO/month (up to 25)             │   │
│  │  • Messages (1pt/100)       │  +0.5 ECHO/message (capped)           │   │
│  │  • Contacts (1pt/10)        │  +2 ECHO/verified contact             │   │
│  │  • Groups (1pt/5)           │  +5 ECHO/active group                 │   │
│  │  • Daily Login Streak       │  +1-10 ECHO (streak bonus)            │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  ON-CHAIN (0-30 pts)        │  ECHO REWARDS                         │   │
│  │  ──────────────────────────────────────────────────────────────────│   │
│  │  • Transactions (1pt/tx)    │  +1 ECHO/successful tx                │   │
│  │  • Staking (1pt/100 ECHO)   │  8-15% APY based on tier              │   │
│  │  • Governance Votes         │  +10 ECHO/vote + influence            │   │
│  │  • Referrals                │  +50 ECHO + 10% of referee earnings   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  PENALTIES (-20 to 0 pts)   │  ECHO PENALTIES                       │   │
│  │  ──────────────────────────────────────────────────────────────────│   │
│  │  • Spam Report: -2pts       │  -10 ECHO per report                  │   │
│  │  • Fraud Report: -5pts      │  -50 ECHO + stake slash               │   │
│  │  • Blocked: -1pt            │  Warning only                         │   │
│  │  • Inactivity: -1pt/month   │  0 (just score decay)                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  TRUST MULTIPLIER (Applied to all ECHO earnings)                            │
│  ═══════════════════════════════════════════════                            │
│  • Score 0-20 (Newcomer):      1.0x                                         │
│  • Score 21-40 (Basic):        1.1x                                         │
│  • Score 41-60 (Trusted):      1.25x                                        │
│  • Score 61-80 (Verified):     1.5x                                         │
│  • Score 81-100 (Elite):       2.0x                                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Trust Score Smart Contract

```go
// metagraph/trust_score.go
package metagraph

import (
    "time"
    "math"
)

// TrustScoreConfig defines the scoring parameters (stored on-chain for transparency)
type TrustScoreConfig struct {
    // Verification weights
    VerificationWeights struct {
        PasskeyCreated    int `json:"passkey_created"`     // 5
        PhoneVerified     int `json:"phone_verified"`      // 5
        EmailVerified     int `json:"email_verified"`      // 5
        KYCLite           int `json:"kyc_lite"`            // 10
        FullKYC           int `json:"full_kyc"`            // 15
        AppleDigitalID    int `json:"apple_digital_id"`    // 15
        OrgVerified       int `json:"org_verified"`        // 20
    }
    
    // Behavior weights
    BehaviorWeights struct {
        AccountAgePerMonth     int `json:"account_age_per_month"`     // 1, max 5
        MessagesPerHundred     int `json:"messages_per_hundred"`      // 1, max 5
        ContactsPerTen         int `json:"contacts_per_ten"`          // 1, max 5
        GroupsPerFive          int `json:"groups_per_five"`           // 1, max 5
        DailyLoginStreak       int `json:"daily_login_streak"`        // 1, max 10
    }
    
    // On-chain weights
    OnChainWeights struct {
        TransactionsPerTx      int `json:"transactions_per_tx"`       // 1, max 10
        StakingPer100ECHO      int `json:"staking_per_100_echo"`      // 1, max 10
        GovernancePerVote      int `json:"governance_per_vote"`       // 1, max 10
    }
    
    // Penalty weights
    PenaltyWeights struct {
        SpamReport             int `json:"spam_report"`               // -2, max -10
        FraudReport            int `json:"fraud_report"`              // -5, max -20
        BlockedByUser          int `json:"blocked_by_user"`           // -1, max -10
        InactivityPerMonth     int `json:"inactivity_per_month"`      // -1, max -20
    }
    
    // Trust multipliers
    TrustMultipliers map[string]float64 `json:"trust_multipliers"`
    
    // ECHO rewards
    ECHORewards map[string]int64 `json:"echo_rewards"`
    
    // Anti-gaming parameters
    AntiGaming struct {
        MinMessageInterval     time.Duration `json:"min_message_interval"`     // 1 second
        MaxDailyMessages       int           `json:"max_daily_messages"`       // 1000
        MinContactAge          time.Duration `json:"min_contact_age"`          // 24 hours
        VouchingCooldown       time.Duration `json:"vouching_cooldown"`        // 7 days
        SybilDetectionEnabled  bool          `json:"sybil_detection_enabled"`
    }
}

// TrustScoreState represents a user's current trust state
type TrustScoreState struct {
    UserDID              string                    `json:"user_did"`
    CurrentScore         int                       `json:"current_score"`
    ScoreBreakdown       TrustScoreBreakdown       `json:"score_breakdown"`
    TrustLevel           TrustLevel                `json:"trust_level"`
    Multiplier           float64                   `json:"multiplier"`
    LastUpdated          time.Time                 `json:"last_updated"`
    LastActivity         time.Time                 `json:"last_activity"`
    StreakDays           int                       `json:"streak_days"`
    LifetimeECHOEarned   int64                     `json:"lifetime_echo_earned"`
    Achievements         []Achievement             `json:"achievements"`
    PendingPenalties     []PendingPenalty          `json:"pending_penalties"`
    MetagraphStateHash   string                    `json:"metagraph_state_hash"`
}

type TrustScoreBreakdown struct {
    Verification int `json:"verification"`
    Behavior     int `json:"behavior"`
    OnChain      int `json:"on_chain"`
    Penalties    int `json:"penalties"`
    Bonus        int `json:"bonus"` // From achievements, streaks
}

type TrustLevel string

const (
    TrustLevelNewcomer TrustLevel = "newcomer"  // 0-20
    TrustLevelBasic    TrustLevel = "basic"     // 21-40
    TrustLevelTrusted  TrustLevel = "trusted"   // 41-60
    TrustLevelVerified TrustLevel = "verified"  // 61-80
    TrustLevelElite    TrustLevel = "elite"     // 81-100
)

// CalculateTrustScore computes the current trust score with anti-gaming checks
func CalculateTrustScore(
    state *TrustScoreState,
    config *TrustScoreConfig,
    activities *UserActivities,
) (*TrustScoreResult, error) {
    
    result := &TrustScoreResult{
        PreviousScore: state.CurrentScore,
        ECHORewards:   make([]ECHOReward, 0),
    }
    
    // 1. Calculate verification score
    verificationScore := calculateVerificationScore(activities.Verifications, config)
    
    // 2. Calculate behavior score with anti-gaming
    behaviorScore, behaviorValid := calculateBehaviorScoreWithAntiGaming(
        activities.Behaviors,
        config,
        state.LastActivity,
    )
    
    if !behaviorValid {
        result.Warnings = append(result.Warnings, "Suspicious activity detected - behavior score capped")
    }
    
    // 3. Calculate on-chain score
    onChainScore := calculateOnChainScore(activities.OnChainActions, config)
    
    // 4. Calculate penalties
    penaltyScore := calculatePenalties(activities.Reports, config)
    
    // 5. Calculate inactivity decay
    decayPenalty := calculateInactivityDecay(state.LastActivity, config)
    
    // 6. Calculate streak bonus
    streakBonus := calculateStreakBonus(state.StreakDays)
    
    // 7. Sum all components
    rawScore := verificationScore + behaviorScore + onChainScore + penaltyScore + decayPenalty + streakBonus
    
    // 8. Clamp to 0-100
    finalScore := clamp(rawScore, 0, 100)
    
    // 9. Determine trust level and multiplier
    trustLevel := getTrustLevel(finalScore)
    multiplier := config.TrustMultipliers[string(trustLevel)]
    
    // 10. Calculate ECHO rewards
    echoRewards := calculateECHORewards(activities, config, multiplier)
    
    result.NewScore = finalScore
    result.TrustLevel = trustLevel
    result.Multiplier = multiplier
    result.Breakdown = TrustScoreBreakdown{
        Verification: verificationScore,
        Behavior:     behaviorScore,
        OnChain:      onChainScore,
        Penalties:    penaltyScore + decayPenalty,
        Bonus:        streakBonus,
    }
    result.ECHORewards = echoRewards
    
    return result, nil
}

// Anti-gaming behavior score calculation
func calculateBehaviorScoreWithAntiGaming(
    behaviors *BehaviorActivities,
    config *TrustScoreConfig,
    lastActivity time.Time,
) (int, bool) {
    
    score := 0
    valid := true
    
    // Account age (hard to game)
    months := int(time.Since(behaviors.AccountCreated).Hours() / 24 / 30)
    score += min(months * config.BehaviorWeights.AccountAgePerMonth, 5)
    
    // Messages with rate limiting check
    if behaviors.MessageCount > 0 {
        // Check for suspicious patterns
        if behaviors.AvgMessageInterval < config.AntiGaming.MinMessageInterval {
            valid = false
            // Cap message score for suspicious activity
            score += 2
        } else {
            messageScore := behaviors.MessageCount / 100
            score += min(messageScore * config.BehaviorWeights.MessagesPerHundred, 5)
        }
    }
    
    // Contacts with verification check
    verifiedContacts := 0
    for _, contact := range behaviors.Contacts {
        // Only count contacts that have been active for minimum period
        if time.Since(contact.AddedAt) >= config.AntiGaming.MinContactAge {
            // Only count contacts with minimum trust score (anti-sybil)
            if contact.TrustScore >= 20 {
                verifiedContacts++
            }
        }
    }
    contactScore := verifiedContacts / 10
    score += min(contactScore * config.BehaviorWeights.ContactsPerTen, 5)
    
    // Groups with activity check
    activeGroups := 0
    for _, group := range behaviors.Groups {
        // Only count groups where user has been active
        if group.UserMessageCount >= 10 && group.MemberCount >= 3 {
            activeGroups++
        }
    }
    groupScore := activeGroups / 5
    score += min(groupScore * config.BehaviorWeights.GroupsPerFive, 5)
    
    return score, valid
}

// Calculate ECHO rewards with multiplier
func calculateECHORewards(
    activities *UserActivities,
    config *TrustScoreConfig,
    multiplier float64,
) []ECHOReward {
    
    rewards := make([]ECHOReward, 0)
    
    // Verification rewards (one-time)
    for _, verification := range activities.NewVerifications {
        baseReward := config.ECHORewards[verification.Type]
        adjustedReward := int64(float64(baseReward) * multiplier)
        
        rewards = append(rewards, ECHOReward{
            Type:       "verification",
            Reason:     verification.Type,
            BaseAmount: baseReward,
            Multiplier: multiplier,
            FinalAmount: adjustedReward,
            Timestamp:  time.Now(),
        })
    }
    
    // Daily activity rewards
    if activities.DailyActive {
        baseReward := config.ECHORewards["daily_activity"]
        adjustedReward := int64(float64(baseReward) * multiplier)
        
        rewards = append(rewards, ECHOReward{
            Type:        "daily",
            Reason:      "daily_activity",
            BaseAmount:  baseReward,
            Multiplier:  multiplier,
            FinalAmount: adjustedReward,
            Timestamp:   time.Now(),
        })
    }
    
    // Streak bonuses
    if activities.StreakMilestone > 0 {
        baseReward := getStreakReward(activities.StreakMilestone)
        adjustedReward := int64(float64(baseReward) * multiplier)
        
        rewards = append(rewards, ECHOReward{
            Type:        "streak",
            Reason:      fmt.Sprintf("%d_day_streak", activities.StreakMilestone),
            BaseAmount:  baseReward,
            Multiplier:  multiplier,
            FinalAmount: adjustedReward,
            Timestamp:   time.Now(),
        })
    }
    
    return rewards
}

func getStreakReward(days int) int64 {
    switch {
    case days >= 365:
        return 1000 // 1000 ECHO for 1 year streak
    case days >= 180:
        return 500
    case days >= 90:
        return 200
    case days >= 30:
        return 100
    case days >= 7:
        return 25
    case days >= 3:
        return 10
    default:
        return 0
    }
}
```

---

## Part 4: Decentralized Verification Alternatives

### 4.1 Web of Trust Implementation

```go
// trust/web_of_trust.go
package trust

import (
    "context"
    "time"
)

// WebOfTrustAttestation represents a peer vouching for another user
type WebOfTrustAttestation struct {
    AttesterDID     string    `json:"attester_did"`
    SubjectDID      string    `json:"subject_did"`
    AttestationType string    `json:"attestation_type"` // "vouch", "endorse", "verify"
    Confidence      int       `json:"confidence"`       // 1-10
    Context         string    `json:"context"`          // "professional", "personal", "community"
    CreatedAt       time.Time `json:"created_at"`
    ExpiresAt       time.Time `json:"expires_at"`
    Signature       string    `json:"signature"`
    MetagraphTxHash string    `json:"metagraph_tx_hash"`
}

// WebOfTrustConfig defines the rules for trust propagation
type WebOfTrustConfig struct {
    // Minimum trust score to vouch for others
    MinVouchingTrustScore int `json:"min_vouching_trust_score"` // 50
    
    // Maximum vouches per user per period
    MaxVouchesPerWeek int `json:"max_vouches_per_week"` // 5
    
    // Trust score boost from vouches
    VouchBoostPoints map[string]int `json:"vouch_boost_points"`
    // e.g., {"vouch": 1, "endorse": 2, "verify": 5}
    
    // Maximum boost from web of trust
    MaxWebOfTrustBoost int `json:"max_web_of_trust_boost"` // 15
    
    // Decay rate for vouches
    VouchDecayDays int `json:"vouch_decay_days"` // 180 (6 months)
    
    // Minimum attestations for verification level
    MinAttestationsForTrusted   int `json:"min_attestations_trusted"`   // 3
    MinAttestationsForVerified  int `json:"min_attestations_verified"`  // 10
    
    // ECHO rewards for vouching
    ECHORewardPerVouch int64 `json:"echo_reward_per_vouch"` // 5 ECHO
    ECHORewardWhenVouched int64 `json:"echo_reward_when_vouched"` // 10 ECHO
}

// WebOfTrustService manages decentralized trust attestations
type WebOfTrustService struct {
    config    *WebOfTrustConfig
    metagraph *MetagraphClient
    cardano   *CardanoClient
}

// CreateAttestation allows a user to vouch for another
func (s *WebOfTrustService) CreateAttestation(
    ctx context.Context,
    attesterDID string,
    subjectDID string,
    attestationType string,
    confidence int,
    attestationContext string,
) (*WebOfTrustAttestation, []ECHOReward, error) {
    
    // 1. Verify attester meets minimum trust requirements
    attesterScore, err := s.getTrustScore(ctx, attesterDID)
    if err != nil {
        return nil, nil, err
    }
    
    if attesterScore.CurrentScore < s.config.MinVouchingTrustScore {
        return nil, nil, ErrInsufficientTrustToVouch
    }
    
    // 2. Check vouch rate limiting
    recentVouches, err := s.getRecentVouches(ctx, attesterDID, 7*24*time.Hour)
    if err != nil {
        return nil, nil, err
    }
    
    if len(recentVouches) >= s.config.MaxVouchesPerWeek {
        return nil, nil, ErrVouchLimitReached
    }
    
    // 3. Prevent self-vouching and circular vouching
    if attesterDID == subjectDID {
        return nil, nil, ErrCannotVouchSelf
    }
    
    hasCircular, err := s.detectCircularVouching(ctx, attesterDID, subjectDID)
    if err != nil {
        return nil, nil, err
    }
    if hasCircular {
        return nil, nil, ErrCircularVouchingDetected
    }
    
    // 4. Create attestation
    attestation := &WebOfTrustAttestation{
        AttesterDID:     attesterDID,
        SubjectDID:      subjectDID,
        AttestationType: attestationType,
        Confidence:      confidence,
        Context:         attestationContext,
        CreatedAt:       time.Now(),
        ExpiresAt:       time.Now().Add(time.Duration(s.config.VouchDecayDays) * 24 * time.Hour),
    }
    
    // 5. Sign attestation
    attestation.Signature, err = s.signAttestation(ctx, attesterDID, attestation)
    if err != nil {
        return nil, nil, err
    }
    
    // 6. Submit to metagraph
    txHash, err := s.metagraph.SubmitAttestation(ctx, attestation)
    if err != nil {
        return nil, nil, err
    }
    attestation.MetagraphTxHash = txHash
    
    // 7. Calculate ECHO rewards
    rewards := []ECHOReward{
        {
            Type:        "web_of_trust",
            Reason:      "vouched_for_user",
            Recipient:   attesterDID,
            BaseAmount:  s.config.ECHORewardPerVouch,
            Multiplier:  attesterScore.Multiplier,
            FinalAmount: int64(float64(s.config.ECHORewardPerVouch) * attesterScore.Multiplier),
        },
        {
            Type:        "web_of_trust",
            Reason:      "received_vouch",
            Recipient:   subjectDID,
            BaseAmount:  s.config.ECHORewardWhenVouched,
            Multiplier:  1.0, // No multiplier for receiving vouches
            FinalAmount: s.config.ECHORewardWhenVouched,
        },
    }
    
    // 8. Submit rewards to Currency L1
    for _, reward := range rewards {
        if err := s.metagraph.SubmitReward(ctx, &reward); err != nil {
            // Log but don't fail - attestation is already recorded
            log.Error("Failed to submit vouch reward", "error", err)
        }
    }
    
    return attestation, rewards, nil
}

// CalculateWebOfTrustBoost calculates trust boost from attestations
func (s *WebOfTrustService) CalculateWebOfTrustBoost(
    ctx context.Context,
    userDID string,
) (int, error) {
    
    attestations, err := s.getActiveAttestations(ctx, userDID)
    if err != nil {
        return 0, err
    }
    
    totalBoost := 0
    uniqueAttesters := make(map[string]bool)
    
    for _, attestation := range attestations {
        // Only count unique attesters
        if uniqueAttesters[attestation.AttesterDID] {
            continue
        }
        uniqueAttesters[attestation.AttesterDID] = true
        
        // Get attester's trust score (weight attestation by attester credibility)
        attesterScore, err := s.getTrustScore(ctx, attestation.AttesterDID)
        if err != nil {
            continue
        }
        
        // Calculate weighted boost
        baseBoost := s.config.VouchBoostPoints[attestation.AttestationType]
        weightedBoost := float64(baseBoost) * (float64(attesterScore.CurrentScore) / 100.0)
        
        // Apply confidence factor
        confidenceFactor := float64(attestation.Confidence) / 10.0
        finalBoost := int(weightedBoost * confidenceFactor)
        
        totalBoost += finalBoost
    }
    
    // Cap at maximum
    if totalBoost > s.config.MaxWebOfTrustBoost {
        totalBoost = s.config.MaxWebOfTrustBoost
    }
    
    return totalBoost, nil
}
```

### 4.2 Decentralized Social Recovery

```go
// recovery/social_recovery.go
package recovery

import (
    "context"
    "time"
)

// SocialRecoveryConfig defines the recovery parameters
type SocialRecoveryConfig struct {
    // Minimum guardians required
    MinGuardians int `json:"min_guardians"` // 3
    
    // Maximum guardians allowed
    MaxGuardians int `json:"max_guardians"` // 7
    
    // Threshold for recovery approval
    RecoveryThreshold int `json:"recovery_threshold"` // 2 of 3, 3 of 5, etc.
    
    // Cooldown after guardian change
    GuardianChangeCooldown time.Duration `json:"guardian_change_cooldown"` // 7 days
    
    // Recovery request timeout
    RecoveryTimeout time.Duration `json:"recovery_timeout"` // 72 hours
    
    // Minimum guardian trust score
    MinGuardianTrustScore int `json:"min_guardian_trust_score"` // 60
    
    // ECHO reward for guardians who participate in recovery
    ECHORewardForRecoveryParticipation int64 `json:"echo_reward_recovery"` // 25 ECHO
}

// Guardian represents a trusted contact for account recovery
type Guardian struct {
    DID           string    `json:"did"`
    AddedAt       time.Time `json:"added_at"`
    Relationship  string    `json:"relationship"` // "friend", "family", "colleague"
    Confirmed     bool      `json:"confirmed"`
    ConfirmedAt   *time.Time `json:"confirmed_at"`
}

// RecoveryRequest represents a pending account recovery
type RecoveryRequest struct {
    ID              string            `json:"id"`
    UserDID         string            `json:"user_did"`
    NewDeviceKey    string            `json:"new_device_key"`
    RequestedAt     time.Time         `json:"requested_at"`
    ExpiresAt       time.Time         `json:"expires_at"`
    Status          RecoveryStatus    `json:"status"`
    Approvals       []RecoveryApproval `json:"approvals"`
    RequiredCount   int               `json:"required_count"`
    MetagraphTxHash string            `json:"metagraph_tx_hash"`
}

type RecoveryStatus string

const (
    RecoveryStatusPending   RecoveryStatus = "pending"
    RecoveryStatusApproved  RecoveryStatus = "approved"
    RecoveryStatusRejected  RecoveryStatus = "rejected"
    RecoveryStatusExpired   RecoveryStatus = "expired"
    RecoveryStatusCompleted RecoveryStatus = "completed"
)

type RecoveryApproval struct {
    GuardianDID string    `json:"guardian_did"`
    Decision    string    `json:"decision"` // "approve", "reject"
    Timestamp   time.Time `json:"timestamp"`
    Signature   string    `json:"signature"`
}

// SocialRecoveryService manages decentralized account recovery
type SocialRecoveryService struct {
    config    *SocialRecoveryConfig
    metagraph *MetagraphClient
    cardano   *CardanoClient
    notifier  *NotificationService
}

// AddGuardian adds a trusted contact as recovery guardian
func (s *SocialRecoveryService) AddGuardian(
    ctx context.Context,
    userDID string,
    guardianDID string,
    relationship string,
) error {
    
    // 1. Verify guardian meets trust requirements
    guardianScore, err := s.getTrustScore(ctx, guardianDID)
    if err != nil {
        return err
    }
    
    if guardianScore.CurrentScore < s.config.MinGuardianTrustScore {
        return ErrGuardianInsufficientTrust
    }
    
    // 2. Check current guardian count
    guardians, err := s.getGuardians(ctx, userDID)
    if err != nil {
        return err
    }
    
    if len(guardians) >= s.config.MaxGuardians {
        return ErrMaxGuardiansReached
    }
    
    // 3. Check cooldown
    lastChange, err := s.getLastGuardianChange(ctx, userDID)
    if err != nil {
        return err
    }
    
    if time.Since(lastChange) < s.config.GuardianChangeCooldown {
        return ErrGuardianChangeCooldown
    }
    
    // 4. Create guardian record (requires guardian confirmation)
    guardian := &Guardian{
        DID:          guardianDID,
        AddedAt:      time.Now(),
        Relationship: relationship,
        Confirmed:    false,
    }
    
    // 5. Submit to metagraph
    if err := s.metagraph.AddGuardian(ctx, userDID, guardian); err != nil {
        return err
    }
    
    // 6. Notify guardian for confirmation
    s.notifier.SendGuardianRequest(ctx, guardianDID, userDID)
    
    return nil
}

// InitiateRecovery starts the recovery process
func (s *SocialRecoveryService) InitiateRecovery(
    ctx context.Context,
    userDID string,
    newDevicePublicKey string,
) (*RecoveryRequest, error) {
    
    // 1. Get confirmed guardians
    guardians, err := s.getConfirmedGuardians(ctx, userDID)
    if err != nil {
        return nil, err
    }
    
    if len(guardians) < s.config.MinGuardians {
        return nil, ErrInsufficientGuardians
    }
    
    // 2. Calculate required approvals
    requiredApprovals := s.calculateThreshold(len(guardians))
    
    // 3. Create recovery request
    request := &RecoveryRequest{
        ID:            generateRecoveryID(),
        UserDID:       userDID,
        NewDeviceKey:  newDevicePublicKey,
        RequestedAt:   time.Now(),
        ExpiresAt:     time.Now().Add(s.config.RecoveryTimeout),
        Status:        RecoveryStatusPending,
        Approvals:     make([]RecoveryApproval, 0),
        RequiredCount: requiredApprovals,
    }
    
    // 4. Submit to metagraph
    txHash, err := s.metagraph.CreateRecoveryRequest(ctx, request)
    if err != nil {
        return nil, err
    }
    request.MetagraphTxHash = txHash
    
    // 5. Notify all guardians
    for _, guardian := range guardians {
        s.notifier.SendRecoveryRequest(ctx, guardian.DID, request)
    }
    
    return request, nil
}

// ApproveRecovery allows a guardian to approve recovery
func (s *SocialRecoveryService) ApproveRecovery(
    ctx context.Context,
    recoveryID string,
    guardianDID string,
    decision string,
    signature string,
) (*RecoveryRequest, []ECHOReward, error) {
    
    // 1. Get recovery request
    request, err := s.getRecoveryRequest(ctx, recoveryID)
    if err != nil {
        return nil, nil, err
    }
    
    // 2. Verify request is pending
    if request.Status != RecoveryStatusPending {
        return nil, nil, ErrRecoveryNotPending
    }
    
    // 3. Verify not expired
    if time.Now().After(request.ExpiresAt) {
        request.Status = RecoveryStatusExpired
        s.metagraph.UpdateRecoveryStatus(ctx, request)
        return nil, nil, ErrRecoveryExpired
    }
    
    // 4. Verify guardian is authorized
    isGuardian, err := s.isGuardian(ctx, request.UserDID, guardianDID)
    if err != nil || !isGuardian {
        return nil, nil, ErrNotAuthorizedGuardian
    }
    
    // 5. Verify signature
    if err := s.verifyGuardianSignature(ctx, guardianDID, recoveryID, decision, signature); err != nil {
        return nil, nil, err
    }
    
    // 6. Record approval
    approval := RecoveryApproval{
        GuardianDID: guardianDID,
        Decision:    decision,
        Timestamp:   time.Now(),
        Signature:   signature,
    }
    request.Approvals = append(request.Approvals, approval)
    
    // 7. Check if threshold reached
    approvalCount := 0
    for _, a := range request.Approvals {
        if a.Decision == "approve" {
            approvalCount++
        }
    }
    
    rewards := make([]ECHOReward, 0)
    
    if approvalCount >= request.RequiredCount {
        request.Status = RecoveryStatusApproved
        
        // 8. Execute recovery (update DID document with new device key)
        if err := s.executeRecovery(ctx, request); err != nil {
            return nil, nil, err
        }
        
        request.Status = RecoveryStatusCompleted
        
        // 9. Reward participating guardians
        for _, a := range request.Approvals {
            if a.Decision == "approve" {
                reward := ECHOReward{
                    Type:        "social_recovery",
                    Reason:      "recovery_participation",
                    Recipient:   a.GuardianDID,
                    BaseAmount:  s.config.ECHORewardForRecoveryParticipation,
                    Multiplier:  1.0,
                    FinalAmount: s.config.ECHORewardForRecoveryParticipation,
                }
                rewards = append(rewards, reward)
                s.metagraph.SubmitReward(ctx, &reward)
            }
        }
    }
    
    // 10. Update metagraph
    s.metagraph.UpdateRecoveryRequest(ctx, request)
    
    return request, rewards, nil
}

func (s *SocialRecoveryService) calculateThreshold(guardianCount int) int {
    // Simple majority + 1 for security
    return (guardianCount / 2) + 1
}
```

---

## Part 5: Achievement & Badge System

### 5.1 Decentralized Achievement NFTs

```go
// achievements/achievement_system.go
package achievements

import (
    "context"
    "time"
)

// Achievement represents a trust-building milestone
type Achievement struct {
    ID              string          `json:"id"`
    Name            string          `json:"name"`
    Description     string          `json:"description"`
    Category        AchievementCat  `json:"category"`
    Tier            AchievementTier `json:"tier"`
    Requirements    []Requirement   `json:"requirements"`
    ECHOReward      int64           `json:"echo_reward"`
    TrustBonus      int             `json:"trust_bonus"`
    BadgeImageURI   string          `json:"badge_image_uri"` // IPFS URI
    IsNFT           bool            `json:"is_nft"`          // Mint as NFT on Cardano
}

type AchievementCat string

const (
    CatVerification AchievementCat = "verification"
    CatCommunity    AchievementCat = "community"
    CatOnChain      AchievementCat = "on_chain"
    CatTrust        AchievementCat = "trust"
    CatStreak       AchievementCat = "streak"
    CatSpecial      AchievementCat = "special"
)

type AchievementTier string

const (
    TierBronze   AchievementTier = "bronze"
    TierSilver   AchievementTier = "silver"
    TierGold     AchievementTier = "gold"
    TierPlatinum AchievementTier = "platinum"
    TierDiamond  AchievementTier = "diamond"
)

// Pre-defined achievements
var Achievements = []Achievement{
    // Verification Achievements
    {
        ID:          "verified_human",
        Name:        "Verified Human",
        Description: "Complete identity verification",
        Category:    CatVerification,
        Tier:        TierGold,
        Requirements: []Requirement{
            {Type: "verification", Value: "high_assurance"},
        },
        ECHOReward:    100,
        TrustBonus:    5,
        BadgeImageURI: "ipfs://Qm.../verified_human.png",
        IsNFT:         true,
    },
    {
        ID:          "early_adopter",
        Name:        "Early Adopter",
        Description: "Join Echo in the first month",
        Category:    CatSpecial,
        Tier:        TierPlatinum,
        Requirements: []Requirement{
            {Type: "account_age", Operator: "before", Value: "2025-03-01"},
        },
        ECHOReward:    250,
        TrustBonus:    10,
        BadgeImageURI: "ipfs://Qm.../early_adopter.png",
        IsNFT:         true,
    },
    
    // Community Achievements
    {
        ID:          "social_butterfly",
        Name:        "Social Butterfly",
        Description: "Add 50 verified contacts",
        Category:    CatCommunity,
        Tier:        TierSilver,
        Requirements: []Requirement{
            {Type: "verified_contacts", Operator: ">=", Value: 50},
        },
        ECHOReward:    50,
        TrustBonus:    3,
        BadgeImageURI: "ipfs://Qm.../social_butterfly.png",
        IsNFT:         false,
    },
    {
        ID:          "trusted_connector",
        Name:        "Trusted Connector",
        Description: "Successfully vouch for 10 users who maintain good standing",
        Category:    CatCommunity,
        Tier:        TierGold,
        Requirements: []Requirement{
            {Type: "successful_vouches", Operator: ">=", Value: 10},
        },
        ECHOReward:    100,
        TrustBonus:    5,
        BadgeImageURI: "ipfs://Qm.../trusted_connector.png",
        IsNFT:         true,
    },
    {
        ID:          "guardian_angel",
        Name:        "Guardian Angel",
        Description: "Serve as recovery guardian for 5 users",
        Category:    CatCommunity,
        Tier:        TierGold,
        Requirements: []Requirement{
            {Type: "guardian_count", Operator: ">=", Value: 5},
        },
        ECHOReward:    75,
        TrustBonus:    4,
        BadgeImageURI: "ipfs://Qm.../guardian_angel.png",
        IsNFT:         true,
    },
    
    // On-Chain Achievements
    {
        ID:          "diamond_hands",
        Name:        "Diamond Hands",
        Description: "Stake 10,000+ ECHO for 6 months",
        Category:    CatOnChain,
        Tier:        TierDiamond,
        Requirements: []Requirement{
            {Type: "staked_amount", Operator: ">=", Value: 10000},
            {Type: "stake_duration", Operator: ">=", Value: 180},
        },
        ECHOReward:    500,
        TrustBonus:    10,
        BadgeImageURI: "ipfs://Qm.../diamond_hands.png",
        IsNFT:         true,
    },
    {
        ID:          "governance_guru",
        Name:        "Governance Guru",
        Description: "Participate in 50 governance votes",
        Category:    CatOnChain,
        Tier:        TierPlatinum,
        Requirements: []Requirement{
            {Type: "governance_votes", Operator: ">=", Value: 50},
        },
        ECHOReward:    200,
        TrustBonus:    8,
        BadgeImageURI: "ipfs://Qm.../governance_guru.png",
        IsNFT:         true,
    },
    
    // Streak Achievements
    {
        ID:          "week_warrior",
        Name:        "Week Warrior",
        Description: "7-day activity streak",
        Category:    CatStreak,
        Tier:        TierBronze,
        Requirements: []Requirement{
            {Type: "streak_days", Operator: ">=", Value: 7},
        },
        ECHOReward:    25,
        TrustBonus:    1,
        BadgeImageURI: "ipfs://Qm.../week_warrior.png",
        IsNFT:         false,
    },
    {
        ID:          "monthly_master",
        Name:        "Monthly Master",
        Description: "30-day activity streak",
        Category:    CatStreak,
        Tier:        TierSilver,
        Requirements: []Requirement{
            {Type: "streak_days", Operator: ">=", Value: 30},
        },
        ECHOReward:    100,
        TrustBonus:    3,
        BadgeImageURI: "ipfs://Qm.../monthly_master.png",
        IsNFT:         false,
    },
    {
        ID:          "yearly_legend",
        Name:        "Yearly Legend",
        Description: "365-day activity streak",
        Category:    CatStreak,
        Tier:        TierDiamond,
        Requirements: []Requirement{
            {Type: "streak_days", Operator: ">=", Value: 365},
        },
        ECHOReward:    1000,
        TrustBonus:    15,
        BadgeImageURI: "ipfs://Qm.../yearly_legend.png",
        IsNFT:         true, // Rare achievement minted as NFT
    },
    
    // Trust Achievements
    {
        ID:          "trust_elite",
        Name:        "Trust Elite",
        Description: "Achieve and maintain 90+ trust score for 30 days",
        Category:    CatTrust,
        Tier:        TierPlatinum,
        Requirements: []Requirement{
            {Type: "trust_score", Operator: ">=", Value: 90},
            {Type: "trust_duration", Operator: ">=", Value: 30},
        },
        ECHOReward:    300,
        TrustBonus:    10,
        BadgeImageURI: "ipfs://Qm.../trust_elite.png",
        IsNFT:         true,
    },
    {
        ID:          "clean_record",
        Name:        "Clean Record",
        Description: "Maintain zero reports for 1 year",
        Category:    CatTrust,
        Tier:        TierGold,
        Requirements: []Requirement{
            {Type: "report_count", Operator: "==", Value: 0},
            {Type: "account_age", Operator: ">=", Value: 365},
        },
        ECHOReward:    150,
        TrustBonus:    5,
        BadgeImageURI: "ipfs://Qm.../clean_record.png",
        IsNFT:         true,
    },
}

// AchievementService manages achievement tracking and rewards
type AchievementService struct {
    achievements map[string]Achievement
    metagraph    *MetagraphClient
    cardano      *CardanoClient // For NFT minting
}

// CheckAndAwardAchievements evaluates user progress and awards earned achievements
func (s *AchievementService) CheckAndAwardAchievements(
    ctx context.Context,
    userDID string,
    userStats *UserStats,
) ([]AwardedAchievement, error) {
    
    awarded := make([]AwardedAchievement, 0)
    
    // Get user's existing achievements
    existing, err := s.getUserAchievements(ctx, userDID)
    if err != nil {
        return nil, err
    }
    existingMap := make(map[string]bool)
    for _, a := range existing {
        existingMap[a.AchievementID] = true
    }
    
    // Check each achievement
    for _, achievement := range s.achievements {
        // Skip if already earned
        if existingMap[achievement.ID] {
            continue
        }
        
        // Check requirements
        if s.meetsRequirements(achievement.Requirements, userStats) {
            // Award achievement
            award, err := s.awardAchievement(ctx, userDID, achievement)
            if err != nil {
                log.Error("Failed to award achievement", "achievement", achievement.ID, "error", err)
                continue
            }
            awarded = append(awarded, *award)
        }
    }
    
    return awarded, nil
}

// awardAchievement grants an achievement to a user
func (s *AchievementService) awardAchievement(
    ctx context.Context,
    userDID string,
    achievement Achievement,
) (*AwardedAchievement, error) {
    
    award := &AwardedAchievement{
        UserDID:       userDID,
        AchievementID: achievement.ID,
        AwardedAt:     time.Now(),
        ECHOReward:    achievement.ECHOReward,
        TrustBonus:    achievement.TrustBonus,
    }
    
    // 1. Record achievement on metagraph
    txHash, err := s.metagraph.RecordAchievement(ctx, award)
    if err != nil {
        return nil, err
    }
    award.MetagraphTxHash = txHash
    
    // 2. If NFT achievement, mint on Cardano
    if achievement.IsNFT {
        nftTx, err := s.cardano.MintAchievementNFT(ctx, userDID, achievement)
        if err != nil {
            log.Error("Failed to mint achievement NFT", "error", err)
            // Don't fail - achievement is recorded, NFT can be minted later
        } else {
            award.NFTTxHash = nftTx
        }
    }
    
    // 3. Submit ECHO reward
    reward := &ECHOReward{
        Type:        "achievement",
        Reason:      achievement.ID,
        Recipient:   userDID,
        BaseAmount:  achievement.ECHOReward,
        Multiplier:  1.0, // Achievements use flat rewards
        FinalAmount: achievement.ECHOReward,
    }
    s.metagraph.SubmitReward(ctx, reward)
    
    // 4. Update trust score with bonus
    s.metagraph.ApplyTrustBonus(ctx, userDID, achievement.TrustBonus, "achievement:"+achievement.ID)
    
    return award, nil
}
```

---

## Part 6: Complete ECHO Incentive Schedule

### 6.1 Comprehensive Reward Table

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ECHO TOKEN INCENTIVE SCHEDULE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ONE-TIME REWARDS                                                            │
│  ════════════════                                                            │
│                                                                              │
│  Action                              │ Base ECHO │ Trust Multiplier Applied │
│  ────────────────────────────────────┼───────────┼─────────────────────────│
│  Create Account + Passkey            │    10     │ ✗ (no trust yet)        │
│  Verify Phone Number                 │    15     │ ✗                       │
│  Verify Email                        │    10     │ ✗                       │
│  Complete KYC-Lite                   │    50     │ ✗                       │
│  Complete Full KYC                   │   100     │ ✗                       │
│  Apple Digital ID Verification       │   100     │ ✗                       │
│  Organization Verification           │   200     │ ✗                       │
│  Add Recovery Guardians (3+)         │    25     │ ✓                       │
│  First Successful Referral           │    75     │ ✓                       │
│  First Governance Vote               │    20     │ ✓                       │
│  First Stake (any amount)            │    25     │ ✓                       │
│                                                                              │
│  RECURRING REWARDS (Daily Caps Apply)                                        │
│  ════════════════════════════════════                                        │
│                                                                              │
│  Action                              │ ECHO/Unit │ Daily Cap │ Trust Mult. │
│  ────────────────────────────────────┼───────────┼───────────┼────────────│
│  Send Message                        │    0.5    │    50     │ ✓          │
│  Add Verified Contact                │    2.0    │    20     │ ✓          │
│  Active Group Participation          │    5.0    │    25     │ ✓          │
│  Daily Login                         │    2.0    │     2     │ ✓          │
│  Successful Transaction              │    1.0    │    10     │ ✓          │
│  Vouch for User (valid)              │    5.0    │    25     │ ✓          │
│  Receive Vouch                       │   10.0    │   N/A     │ ✗          │
│  Recovery Guardian Participation     │   25.0    │   N/A     │ ✗          │
│                                                                              │
│  STREAK BONUSES                                                              │
│  ══════════════                                                              │
│                                                                              │
│  Streak Duration                     │ Bonus ECHO │ Additional Benefit      │
│  ────────────────────────────────────┼────────────┼───────────────────────│
│  3 Days                              │     10     │ -                       │
│  7 Days                              │     25     │ +0.1x earning boost     │
│  14 Days                             │     50     │ +0.15x earning boost    │
│  30 Days                             │    100     │ +0.2x earning boost     │
│  90 Days                             │    200     │ +0.25x earning boost    │
│  180 Days                            │    500     │ +0.3x earning boost     │
│  365 Days                            │   1000     │ +0.5x earning boost     │
│                                                                              │
│  STAKING REWARDS                                                             │
│  ═══════════════                                                             │
│                                                                              │
│  Tier         │ Min Stake  │ Lock Period │ Base APY │ Trust Bonus           │
│  ─────────────┼────────────┼─────────────┼──────────┼──────────────────────│
│  Bronze       │    100     │   30 days   │    8%    │ +1 trust point        │
│  Silver       │    500     │   60 days   │   10%    │ +3 trust points       │
│  Gold         │  2,000     │   90 days   │   12%    │ +5 trust points       │
│  Platinum     │  5,000     │  180 days   │   14%    │ +7 trust points       │
│  Diamond      │ 10,000     │  365 days   │   15%    │ +10 trust points      │
│                                                                              │
│  GOVERNANCE PARTICIPATION                                                    │
│  ════════════════════════                                                    │
│                                                                              │
│  Action                              │ ECHO Reward │ Trust Impact           │
│  ────────────────────────────────────┼─────────────┼───────────────────────│
│  Vote on Proposal                    │     10      │ +1 point               │
│  Submit Proposal (if passes)         │    100      │ +5 points              │
│  Delegate Votes                      │      5      │ +0.5 points            │
│  Maintain Delegation 30+ days        │     25      │ +2 points              │
│                                                                              │
│  REFERRAL PROGRAM                                                            │
│  ════════════════                                                            │
│                                                                              │
│  Event                               │ Referrer    │ Referee                │
│  ────────────────────────────────────┼─────────────┼───────────────────────│
│  Referee Signs Up                    │     25      │    25                  │
│  Referee Completes Verification      │     50      │   (included above)     │
│  Referee Reaches Trust 50+           │     25      │    10                  │
│  Referee's First Stake               │     10      │     -                  │
│  Ongoing: 10% of Referee Earnings    │  Lifetime   │     -                  │
│                                                                              │
│  PENALTY DEDUCTIONS                                                          │
│  ══════════════════                                                          │
│                                                                              │
│  Violation                           │ ECHO Loss   │ Trust Impact           │
│  ────────────────────────────────────┼─────────────┼───────────────────────│
│  Spam Report (confirmed)             │    -10      │ -2 points              │
│  Fraud Report (confirmed)            │    -50      │ -5 points              │
│  Stake Slashing (bad actor)          │  -10% stake │ -20 points             │
│  Vouch Abuse (serial bad vouches)    │    -25      │ -5 points              │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Part 7: Security Enhancements

### 7.1 Anti-Sybil Measures

```go
// security/anti_sybil.go
package security

import (
    "context"
    "time"
)

// SybilDetector identifies potential fake/duplicate accounts
type SybilDetector struct {
    config    *SybilConfig
    metagraph *MetagraphClient
    ml        *MLClient // Optional ML-based detection
}

type SybilConfig struct {
    // Behavioral thresholds
    MaxAccountsPerDevice     int           `json:"max_accounts_per_device"`     // 2
    MaxAccountsPerIP         int           `json:"max_accounts_per_ip"`         // 5
    MinTimeBetweenAccounts   time.Duration `json:"min_time_between_accounts"`   // 24h
    
    // Network analysis
    MaxClusterSize           int           `json:"max_cluster_size"`            // 10
    MinClusterDiversity      float64       `json:"min_cluster_diversity"`       // 0.7
    
    // Interaction patterns
    MinUniqueIPsForTrust     int           `json:"min_unique_ips_for_trust"`    // 3
    SuspiciousMessageRate    float64       `json:"suspicious_message_rate"`     // 100/min
    
    // Verification requirements for high-risk actions
    RequireVerificationFor   []string      `json:"require_verification_for"`
    // e.g., ["large_transfer", "high_stake", "governance_vote"]
}

// AnalyzeAccount checks for sybil indicators
func (d *SybilDetector) AnalyzeAccount(
    ctx context.Context,
    userDID string,
) (*SybilAnalysis, error) {
    
    analysis := &SybilAnalysis{
        UserDID:    userDID,
        Timestamp:  time.Now(),
        RiskLevel:  RiskLevelLow,
        Indicators: make([]SybilIndicator, 0),
    }
    
    // 1. Device fingerprint analysis
    deviceRisk, indicators := d.analyzeDevicePatterns(ctx, userDID)
    analysis.Indicators = append(analysis.Indicators, indicators...)
    
    // 2. Network graph analysis
    networkRisk, indicators := d.analyzeNetworkPatterns(ctx, userDID)
    analysis.Indicators = append(analysis.Indicators, indicators...)
    
    // 3. Behavioral analysis
    behaviorRisk, indicators := d.analyzeBehaviorPatterns(ctx, userDID)
    analysis.Indicators = append(analysis.Indicators, indicators...)
    
    // 4. Cross-reference with known sybil clusters
    clusterRisk, indicators := d.checkKnownClusters(ctx, userDID)
    analysis.Indicators = append(analysis.Indicators, indicators...)
    
    // 5. Calculate overall risk score
    analysis.RiskScore = (deviceRisk + networkRisk + behaviorRisk + clusterRisk) / 4
    analysis.RiskLevel = d.getRiskLevel(analysis.RiskScore)
    
    // 6. Determine restrictions
    if analysis.RiskLevel >= RiskLevelMedium {
        analysis.Restrictions = d.getRestrictions(analysis.RiskLevel)
    }
    
    return analysis, nil
}

func (d *SybilDetector) analyzeNetworkPatterns(
    ctx context.Context,
    userDID string,
) (float64, []SybilIndicator) {
    
    indicators := make([]SybilIndicator, 0)
    riskScore := 0.0
    
    // Get user's contact graph
    contacts, _ := d.metagraph.GetContacts(ctx, userDID)
    
    // Check for suspicious clustering
    cluster := d.findConnectedCluster(ctx, userDID, contacts)
    
    if len(cluster) > d.config.MaxClusterSize {
        indicators = append(indicators, SybilIndicator{
            Type:        "large_cluster",
            Description: "Part of unusually large tightly-connected cluster",
            Severity:    0.7,
        })
        riskScore += 0.3
    }
    
    // Check cluster diversity (age, verification, activity patterns)
    diversity := d.calculateClusterDiversity(ctx, cluster)
    if diversity < d.config.MinClusterDiversity {
        indicators = append(indicators, SybilIndicator{
            Type:        "low_diversity_cluster",
            Description: "Cluster has suspiciously similar account characteristics",
            Severity:    0.8,
        })
        riskScore += 0.4
    }
    
    // Check for reciprocal vouching rings
    vouchRings := d.detectVouchingRings(ctx, userDID)
    if len(vouchRings) > 0 {
        indicators = append(indicators, SybilIndicator{
            Type:        "vouch_ring",
            Description: "Detected circular vouching pattern",
            Severity:    0.9,
        })
        riskScore += 0.5
    }
    
    return riskScore, indicators
}

// Restrictions based on risk level
func (d *SybilDetector) getRestrictions(riskLevel RiskLevel) []Restriction {
    switch riskLevel {
    case RiskLevelMedium:
        return []Restriction{
            {Type: "earning_cap", Value: "50%"},      // 50% of normal earning cap
            {Type: "vouch_disabled", Value: "true"},  // Cannot vouch for others
            {Type: "verification_required", Value: "kyc_lite"},
        }
    case RiskLevelHigh:
        return []Restriction{
            {Type: "earning_cap", Value: "10%"},
            {Type: "vouch_disabled", Value: "true"},
            {Type: "stake_disabled", Value: "true"},
            {Type: "verification_required", Value: "full_kyc"},
        }
    case RiskLevelCritical:
        return []Restriction{
            {Type: "account_suspended", Value: "true"},
            {Type: "appeal_required", Value: "true"},
        }
    default:
        return nil
    }
}
```

---

## Part 8: Summary & Implementation Roadmap

### 8.1 Enhancement Summary

| Category | Enhancement | Priority | Effort | ECHO Impact |
|----------|-------------|----------|--------|-------------|
| **Decentralization** | Web of Trust | High | 3 weeks | +5 ECHO/vouch |
| **Decentralization** | Social Recovery | High | 2 weeks | +25 ECHO/recovery |
| **Decentralization** | Federated Verification | Medium | 4 weeks | - |
| **Incentives** | Gamified Trust Score | High | 2 weeks | Comprehensive |
| **Incentives** | Achievement NFTs | Medium | 3 weeks | 25-1000 ECHO |
| **Incentives** | Streak System | High | 1 week | 10-1000 ECHO |
| **Security** | Anti-Sybil Detection | High | 3 weeks | Prevents abuse |
| **Security** | Trust Decay | Medium | 1 week | Maintains quality |
| **Features** | Governance Rewards | Medium | 2 weeks | 10-100 ECHO |

### 8.2 Implementation Phases

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    IMPLEMENTATION ROADMAP                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  PHASE 1: Core Trust & Incentives (Weeks 1-4)                               │
│  ──────────────────────────────────────────────                              │
│  □ Enhanced Trust Score Algorithm                                            │
│  □ Trust Multiplier System                                                   │
│  □ Basic ECHO Reward Distribution                                            │
│  □ Streak Tracking & Bonuses                                                │
│  □ Anti-Gaming Basic Measures                                               │
│                                                                              │
│  PHASE 2: Decentralized Trust (Weeks 5-8)                                   │
│  ──────────────────────────────────────────                                  │
│  □ Web of Trust Attestations                                                │
│  □ Social Recovery Implementation                                            │
│  □ Guardian Management                                                       │
│  □ Vouch-based Trust Boost                                                  │
│                                                                              │
│  PHASE 3: Gamification (Weeks 9-12)                                         │
│  ──────────────────────────────────────                                      │
│  □ Achievement System                                                        │
│  □ NFT Badge Minting                                                        │
│  □ Leaderboards (optional)                                                  │
│  □ Seasonal Challenges                                                       │
│                                                                              │
│  PHASE 4: Advanced Security (Weeks 13-16)                                   │
│  ─────────────────────────────────────────                                   │
│  □ ML-based Sybil Detection                                                 │
│  □ Network Graph Analysis                                                    │
│  □ Federated Verification Providers                                         │
│  □ Cross-chain Identity Bridges                                             │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 8.3 Key Metrics to Track

| Metric | Target | Purpose |
|--------|--------|---------|
| Verification Rate | >60% of users | Measure trust adoption |
| Avg Trust Score | >50 | Network health |
| ECHO Circulation | Controlled inflation | Economic health |
| Sybil Detection Rate | >95% accuracy | Security |
| Streak Retention | >40% 7-day | Engagement |
| Web of Trust Density | >3 vouches/user | Decentralization |
| Recovery Success Rate | >99% | UX quality |

---

## Document Information

| Field | Value |
|-------|-------|
| Version | 1.0 |
| Date | February 2025 |
| Status | Review Required |
| Focus Areas | Decentralization, ECHO Incentives, Trust Gamification |

---

*This enhancement document should be reviewed alongside the original blueprint and integrated into the implementation plan.*
