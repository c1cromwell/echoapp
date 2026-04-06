# Go Backend Implementation Spec (v3.0)

## Aligned To

| Document | Version |
|----------|---------|
| PRD | v2.4 |
| Data Layer Architecture | v3.3 |
| Backend Architecture | v2.1 |
| ECHO Tokenomics | v1.0 |
| OpenAPI | v2 |

---

## 1. Project Structure

```
echo-backend/
├── cmd/
│   └── echo-server/
│       └── main.go                    # Entry point, service wiring
├── internal/
│   ├── gateway/                       # Port 8000 — API Gateway
│   │   ├── router.go                  # Chi/Gin router, middleware chain
│   │   ├── middleware/
│   │   │   ├── auth.go                # Bearer token validation
│   │   │   ├── ratelimit.go           # Per-DID rate limiting
│   │   │   ├── tls.go                 # TLS 1.3 + cert pinning
│   │   │   └── cors.go
│   │   └── handlers/                  # HTTP handlers (thin — delegate to services)
│   │       ├── auth_handler.go
│   │       ├── message_handler.go
│   │       ├── conversation_handler.go
│   │       ├── user_handler.go
│   │       ├── identity_handler.go
│   │       ├── token_handler.go       # Wallet: balance, stake, delegate, rewards
│   │       ├── evidence_handler.go    # Digital Evidence fingerprinting
│   │       └── group_handler.go
│   │
│   ├── identity/                      # Port 8001 — Identity Service
│   │   ├── service.go                 # DID management, credential caching
│   │   ├── did.go                     # Cardano DID operations
│   │   ├── passkey.go                 # WebAuthn passkey registration
│   │   └── models.go
│   │
│   ├── relay/                         # Port 8002 — Message Relay Service
│   │   ├── service.go                 # Core relay logic
│   │   ├── connections.go             # WebSocket connection manager
│   │   ├── offline_queue.go           # Redis + PG offline message queue
│   │   ├── apns.go                    # Apple Push Notification
│   │   ├── anchoring_batcher.go       # Merkle tree batching → Data L1
│   │   └── models.go
│   │
│   ├── trust/                         # Port 8003 — Trust Service
│   │   ├── service.go                 # Trust score computation
│   │   ├── tier.go                    # Tier calculation, multipliers
│   │   └── models.go
│   │
│   ├── rewards/                       # Port 8004 — Rewards Service
│   │   ├── service.go                 # Reward validation + batching
│   │   ├── emission.go                # 10-year declining emission curve
│   │   ├── daily_cap.go               # Per-DID daily cap tracking
│   │   ├── anti_gaming.go             # Velocity checks, pattern detection
│   │   └── models.go
│   │
│   ├── wallet/                        # Wallet API (new — supports iOS Wallet tab)
│   │   ├── service.go                 # Balance aggregation, position queries
│   │   ├── staking.go                 # TokenLock construction + submission
│   │   ├── delegation.go              # StakeDelegation construction + submission
│   │   ├── vesting.go                 # Founder vesting position queries
│   │   ├── swap.go                    # PacaSwap integration (Phase 3+)
│   │   └── models.go
│   │
│   ├── metagraph/                     # Port 8006 — Metagraph Gateway
│   │   ├── client.go                  # L1/L0 submission (v3 transaction types)
│   │   ├── snapshot_listener.go       # Snapshot event subscription
│   │   ├── fee_manager.go             # FeeTransaction automation
│   │   ├── v3_types.go                # Tessellation v3 type definitions
│   │   └── models.go
│   │
│   ├── evidence/                      # Port 8010 — Digital Evidence
│   │   ├── client.go                  # Constellation Digital Evidence API client
│   │   ├── service.go                 # Fingerprint submission orchestration
│   │   └── models.go
│   │
│   ├── cardano/                       # Cardano Integration
│   │   ├── client.go                  # DID registration, credential queries
│   │   ├── trust_tier.go              # Trust tier UTXO queries
│   │   └── models.go
│   │
│   ├── logging/                       # Port 8009 — IPFS Log Publisher
│   │   ├── publisher.go               # Batch encrypt → IPFS → record CID
│   │   └── models.go
│   │
│   └── infra/                         # Shared infrastructure
│       ├── circuit_breaker.go
│       ├── redis.go
│       ├── postgres.go
│       ├── nats.go
│       └── config.go
│
├── pkg/                               # Public packages
│   ├── crypto/
│   │   ├── merkle.go                  # Merkle tree construction
│   │   ├── commitment.go              # H(H(plaintext) || nonce)
│   │   └── aes.go                     # AES-256-GCM for log encryption
│   └── types/
│       └── echo_types.go              # Shared domain types
│
├── configs/
│   ├── config.yaml                    # Service configuration
│   └── genesis.yaml                   # Token genesis parameters
│
├── migrations/                        # PostgreSQL migrations
│   ├── 001_initial.sql
│   ├── 002_offline_queue.sql
│   ├── 003_wallet_cache.sql
│   └── 004_evidence_audit.sql
│
├── Dockerfile
├── docker-compose.yaml               # Local dev (Go + Redis + PG + NATS)
├── go.mod
└── go.sum
```

---

## 2. Core Types (Tessellation v3)

```go
// metagraph/v3_types.go

package metagraph

import "time"

// ── Tessellation v3 Transaction Primitives ──

// TokenLockRequest locks ECHO in user's wallet. Tokens remain in wallet but are non-spendable.
type TokenLockRequest struct {
    DID           string `json:"did"`
    Amount        int64  `json:"amount"`         // ECHO amount (smallest unit)
    Tier          string `json:"tier"`            // bronze/silver/gold/platinum
    DurationDays  int    `json:"durationDays"`    // 30/90/180/365
    VestingType   string `json:"vestingType,omitempty"` // "founder" for vesting locks, empty for user staking
    CliffMonths   int    `json:"cliffMonths,omitempty"` // 12 for founders, 0 for users
    VestMonths    int    `json:"vestMonths,omitempty"`  // 48 for founders, 0 for users
}

// StakeDelegationRequest delegates locked tokens to an L1 validator.
type StakeDelegationRequest struct {
    DelegatorDID string `json:"delegatorDid"`
    ValidatorID  string `json:"validatorId"`
    StakeID      string `json:"stakeId"`      // TokenLock position to delegate
    Amount       int64  `json:"amount"`
}

// WithdrawLockRequest initiates unstaking with 14-day cooldown.
type WithdrawLockRequest struct {
    DID     string `json:"did"`
    StakeID string `json:"stakeId"`
    Amount  int64  `json:"amount"`
    // L1 validation enforces 14-day cooldown; founder locks enforce cliff/vest schedule
}

// AllowSpendRequest creates time-limited spend approval (Phase 5 marketplace).
type AllowSpendRequest struct {
    OwnerDID   string    `json:"ownerDid"`
    SpenderDID string    `json:"spenderDid"`
    Amount     int64     `json:"amount"`
    ExpiresAt  time.Time `json:"expiresAt"` // Never unlimited
}

// SpendTransactionRequest executes against an AllowSpend approval.
type SpendTransactionRequest struct {
    AllowanceID string `json:"allowanceId"`
    Amount      int64  `json:"amount"`
}

// FeeTransactionRequest pays snapshot fees from treasury DAG reserves.
type FeeTransactionRequest struct {
    DAGAmount int64 `json:"dagAmount"`
}

// AtomicActionRequest bundles multiple transactions — all succeed or all fail.
type AtomicActionRequest struct {
    Actions       []TransactionAction `json:"actions"`
    SchemaVersion int                 `json:"schemaVersion"`
}

// TransactionAction is one action inside an AtomicAction bundle.
type TransactionAction struct {
    Type            string               `json:"type"` // "verify_tier", "claim_reward", "update_cap", "token_lock", etc.
    VerifyTier      *VerifyTierAction     `json:"verifyTier,omitempty"`
    ClaimReward     *ClaimRewardAction    `json:"claimReward,omitempty"`
    UpdateCap       *UpdateCapAction      `json:"updateCap,omitempty"`
    TokenLock       *TokenLockRequest     `json:"tokenLock,omitempty"`
    StakeDelegation *StakeDelegationRequest `json:"stakeDelegation,omitempty"`
}

type VerifyTierAction struct {
    DID          string `json:"did"`
    ExpectedTier int    `json:"expectedTier"`
}

type ClaimRewardAction struct {
    DID        string   `json:"did"`
    RewardType string   `json:"rewardType"` // messaging, referral, staking, payment_rail
    Amount     int64    `json:"amount"`
}

type UpdateCapAction struct {
    DID        string `json:"did"`
    RewardType string `json:"rewardType"`
    Amount     int64  `json:"amountUsed"`
}
```

---

## 3. Wallet Service (New)

```go
// wallet/service.go

package wallet

import (
    "context"
    "echo-backend/internal/metagraph"
    "echo-backend/internal/rewards"
)

// WalletService aggregates on-chain and cached data for the iOS Wallet tab.
type WalletService struct {
    metagraph  *metagraph.MetagraphClient
    rewards    *rewards.Service
    redis      *Redis
}

// GetWalletState returns the complete wallet view for a DID.
func (s *WalletService) GetWalletState(ctx context.Context, did string) (*WalletState, error) {
    // 1. Balance from metagraph (cached 5s TTL)
    balance, err := s.getBalance(ctx, did)
    if err != nil {
        return nil, err
    }

    // 2. TokenLock positions (staking)
    locks, err := s.getTokenLocks(ctx, did)
    if err != nil {
        return nil, err
    }

    // 3. StakeDelegation positions
    delegations, err := s.getDelegations(ctx, did)
    if err != nil {
        return nil, err
    }

    // 4. Pending rewards (from rewards service cache)
    pending, err := s.rewards.GetPending(ctx, did)
    if err != nil {
        return nil, err
    }

    // 5. Daily reward progress
    dailyCaps, err := s.rewards.GetDailyCaps(ctx, did)
    if err != nil {
        return nil, err
    }

    // 6. Founder vesting (if applicable)
    var vesting *VestingState
    for _, lock := range locks {
        if lock.VestingType == "founder" {
            vesting = s.computeVestingState(lock)
            break
        }
    }

    return &WalletState{
        DID:            did,
        TotalBalance:   balance.Total,
        Available:      balance.Available,
        Staked:         sumLocks(locks),
        PendingRewards: pending.Total,
        Locks:          locks,
        Delegations:    delegations,
        DailyRewards:   dailyCaps,
        Vesting:        vesting,
    }, nil
}

// StakeEcho constructs and submits a TokenLock transaction.
func (s *WalletService) StakeEcho(ctx context.Context, req StakeRequest) (*StakeResult, error) {
    // Validate tier
    tier, err := ValidateTier(req.Tier)
    if err != nil {
        return nil, err
    }

    // Submit TokenLock via metagraph gateway
    txHash, err := s.metagraph.SubmitCurrencyL1(metagraph.CurrencyL1Transaction{
        Type: "token_lock",
        TokenLock: &metagraph.TokenLockRequest{
            DID:          req.DID,
            Amount:       req.Amount,
            Tier:         tier.Name,
            DurationDays: tier.DurationDays,
        },
    })
    if err != nil {
        return nil, err
    }

    return &StakeResult{TxHash: txHash, Tier: tier}, nil
}

// DelegateToValidator constructs and submits a StakeDelegation transaction.
func (s *WalletService) DelegateToValidator(ctx context.Context, req DelegateRequest) (*DelegateResult, error) {
    txHash, err := s.metagraph.SubmitCurrencyL1(metagraph.CurrencyL1Transaction{
        Type: "stake_delegation",
        StakeDelegation: &metagraph.StakeDelegationRequest{
            DelegatorDID: req.DID,
            ValidatorID:  req.ValidatorID,
            StakeID:      req.StakeID,
            Amount:       req.Amount,
        },
    })
    if err != nil {
        return nil, err
    }

    return &DelegateResult{TxHash: txHash}, nil
}

// ClaimRewards constructs and submits an AtomicAction for reward claiming.
func (s *WalletService) ClaimRewards(ctx context.Context, did string, types []string) (*ClaimResult, error) {
    // Build atomic action: verify tier + claim each type + update caps
    actions := []metagraph.TransactionAction{
        {Type: "verify_tier", VerifyTier: &metagraph.VerifyTierAction{DID: did}},
    }

    for _, rewardType := range types {
        pending, _ := s.rewards.GetPendingByType(ctx, did, rewardType)
        if pending > 0 {
            actions = append(actions,
                metagraph.TransactionAction{
                    Type:        "claim_reward",
                    ClaimReward: &metagraph.ClaimRewardAction{DID: did, RewardType: rewardType, Amount: pending},
                },
                metagraph.TransactionAction{
                    Type:      "update_cap",
                    UpdateCap: &metagraph.UpdateCapAction{DID: did, RewardType: rewardType, AmountUsed: pending},
                },
            )
        }
    }

    txHash, err := s.metagraph.SubmitAtomicAction(actions)
    if err != nil {
        return nil, err
    }

    // Clear pending cache
    s.rewards.ClearPending(ctx, did, types)

    return &ClaimResult{TxHash: txHash}, nil
}

// Unstake constructs and submits a WithdrawLock transaction.
func (s *WalletService) Unstake(ctx context.Context, req UnstakeRequest) (*UnstakeResult, error) {
    txHash, err := s.metagraph.SubmitCurrencyL1(metagraph.CurrencyL1Transaction{
        Type: "withdraw_lock",
        WithdrawLock: &metagraph.WithdrawLockRequest{
            DID:     req.DID,
            StakeID: req.StakeID,
            Amount:  req.Amount,
        },
    })
    if err != nil {
        return nil, err
    }

    return &UnstakeResult{
        TxHash:          txHash,
        CooldownEndDate: time.Now().Add(14 * 24 * time.Hour),
    }, nil
}

// GetValidators returns active L1 validators with performance metrics.
func (s *WalletService) GetValidators(ctx context.Context) ([]ValidatorInfo, error) {
    validators, err := s.metagraph.QueryValidators()
    if err != nil {
        return nil, err
    }

    var result []ValidatorInfo
    for _, v := range validators {
        result = append(result, ValidatorInfo{
            ID:             v.ID,
            Address:        v.Address,
            Uptime:         v.UptimePercent,
            Commission:     v.CommissionPercent,
            TotalDelegated: v.TotalDelegated,
            DelegatorCount: v.DelegatorCount,
            Layer:          v.Layer, // "currency_l1" or "data_l1"
            EstimatedAPR:   s.calculateAPR(v),
        })
    }

    return result, nil
}
```

---

## 4. Wallet Models

```go
// wallet/models.go

package wallet

import "time"

type WalletState struct {
    DID            string           `json:"did"`
    TotalBalance   int64            `json:"totalBalance"`
    Available      int64            `json:"available"`
    Staked         int64            `json:"staked"`
    PendingRewards int64            `json:"pendingRewards"`
    Locks          []TokenLockPos   `json:"locks"`
    Delegations    []DelegationPos  `json:"delegations"`
    DailyRewards   *DailyCapState   `json:"dailyRewards"`
    Vesting        *VestingState    `json:"vesting,omitempty"` // Founders only
}

type TokenLockPos struct {
    ID           string    `json:"id"`
    Amount       int64     `json:"amount"`
    Tier         string    `json:"tier"`
    LockedUntil  time.Time `json:"lockedUntil"`
    VestingType  string    `json:"vestingType,omitempty"` // "founder" or ""
    DelegatedTo  string    `json:"delegatedTo,omitempty"`
}

type DelegationPos struct {
    ID          string        `json:"id"`
    StakeID     string        `json:"stakeId"`
    ValidatorID string        `json:"validatorId"`
    Validator   ValidatorInfo `json:"validator"`
    Amount      int64         `json:"amount"`
    Since       time.Time     `json:"since"`
}

type ValidatorInfo struct {
    ID             string  `json:"id"`
    Address        string  `json:"address"`
    Uptime         float64 `json:"uptimePercent"`
    Commission     float64 `json:"commissionPercent"`
    TotalDelegated int64   `json:"totalDelegated"`
    DelegatorCount int     `json:"delegatorCount"`
    Layer          string  `json:"layer"` // currency_l1, data_l1
    EstimatedAPR   float64 `json:"estimatedApr"`
}

type VestingState struct {
    Role             string    `json:"role"`
    TotalAllocated   int64     `json:"totalAllocated"`
    Vested           int64     `json:"vested"`
    Locked           int64     `json:"locked"`
    Withdrawable     int64     `json:"withdrawable"`
    NextUnlockAmount int64     `json:"nextUnlockAmount"`
    NextUnlockDate   time.Time `json:"nextUnlockDate"`
    CliffDate        time.Time `json:"cliffDate"`
    CliffCompleted   bool      `json:"cliffCompleted"`
    VestingPercent   float64   `json:"vestingPercent"`
    ExplorerURL      string    `json:"explorerUrl"`
}

type DailyCapState struct {
    Messaging   DailyCapEntry `json:"messaging"`
    Referrals   DailyCapEntry `json:"referrals"`
    Staking     DailyCapEntry `json:"staking"`
    PaymentRail DailyCapEntry `json:"paymentRail"`
}

type DailyCapEntry struct {
    Earned int64 `json:"earned"`
    Cap    int64 `json:"cap"`
}

type StakeRequest struct {
    DID    string `json:"did"`
    Amount int64  `json:"amount"`
    Tier   string `json:"tier"`
}

type DelegateRequest struct {
    DID         string `json:"did"`
    ValidatorID string `json:"validatorId"`
    StakeID     string `json:"stakeId"`
    Amount      int64  `json:"amount"`
}

type UnstakeRequest struct {
    DID     string `json:"did"`
    StakeID string `json:"stakeId"`
    Amount  int64  `json:"amount"`
}

type StakingTier struct {
    Name         string  `json:"name"`
    DurationDays int     `json:"durationDays"`
    APR          float64 `json:"apr"`
}

var StakingTiers = map[string]StakingTier{
    "bronze":   {Name: "bronze", DurationDays: 30, APR: 5.0},
    "silver":   {Name: "silver", DurationDays: 90, APR: 8.0},
    "gold":     {Name: "gold", DurationDays: 180, APR: 12.0},
    "platinum": {Name: "platinum", DurationDays: 365, APR: 15.0},
}
```

---

## 5. Emission Curve Service

```go
// rewards/emission.go

package rewards

import "time"

// EmissionSchedule manages the 10-year declining emission curve for community rewards.
// 400M ECHO total, declining annually. Enforced by Currency L1 Scala validation.
// Go backend pre-validates to avoid rejected submissions.
type EmissionSchedule struct {
    GenesisDate   time.Time
    TotalPool     int64 // 400_000_000_00000000 (with 8 decimal places)
    YearlyPercent []float64
}

func NewEmissionSchedule(genesis time.Time) *EmissionSchedule {
    return &EmissionSchedule{
        GenesisDate:   genesis,
        TotalPool:     400_000_000_00000000,
        YearlyPercent: []float64{0.20, 0.16, 0.13, 0.11, 0.09, 0.07, 0.06, 0.06, 0.06, 0.06},
    }
}

// YearlyBudget returns the emission budget for a given year (1-indexed).
func (e *EmissionSchedule) YearlyBudget(year int) int64 {
    if year < 1 || year > 10 {
        return 0 // No emission after year 10
    }
    return int64(float64(e.TotalPool) * e.YearlyPercent[year-1])
}

// DailyBudget returns the daily emission budget for the current year.
func (e *EmissionSchedule) DailyBudget() int64 {
    year := e.currentYear()
    if year > 10 {
        return 0
    }
    return e.YearlyBudget(year) / 365
}

// RemainingToday returns how much of today's budget remains after claimed rewards.
func (e *EmissionSchedule) RemainingToday(claimedToday int64) int64 {
    daily := e.DailyBudget()
    remaining := daily - claimedToday
    if remaining < 0 {
        return 0
    }
    return remaining
}

func (e *EmissionSchedule) currentYear() int {
    elapsed := time.Since(e.GenesisDate)
    return int(elapsed.Hours()/8760) + 1
}
```

---

## 6. Fee Manager (FeeTransaction Automation)

```go
// metagraph/fee_manager.go

package metagraph

import (
    "context"
    "log"
    "time"
)

// FeeManager automatically pays snapshot fees from treasury DAG reserves.
type FeeManager struct {
    client         *MetagraphClient
    dagBalance     int64 // Cached treasury DAG balance
    checkInterval  time.Duration
    lowBalanceAlert func(balance int64)
}

func NewFeeManager(client *MetagraphClient) *FeeManager {
    return &FeeManager{
        client:        client,
        checkInterval: 1 * time.Hour,
    }
}

// Run starts the automated fee payment loop.
func (f *FeeManager) Run(ctx context.Context) {
    ticker := time.NewTicker(f.checkInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := f.checkAndPayFees(ctx); err != nil {
                log.Printf("Fee payment error: %v", err)
            }
        }
    }
}

func (f *FeeManager) checkAndPayFees(ctx context.Context) error {
    // 1. Query pending snapshot fees
    pending, err := f.client.QueryPendingFees()
    if err != nil {
        return err
    }

    if pending <= 0 {
        return nil
    }

    // 2. Check treasury DAG balance
    balance, err := f.client.QueryDAGBalance(TreasuryAddress)
    if err != nil {
        return err
    }

    if balance < pending*2 { // Alert if balance below 2x pending fees
        if f.lowBalanceAlert != nil {
            f.lowBalanceAlert(balance)
        }
    }

    // 3. Submit FeeTransaction
    _, err = f.client.SubmitCurrencyL1(CurrencyL1Transaction{
        Type:           "fee_transaction",
        FeeTransaction: &FeeTransactionRequest{DAGAmount: pending},
    })
    return err
}
```

---

## 7. Digital Evidence Service

```go
// evidence/service.go

package evidence

import "context"

// EvidenceService orchestrates Digital Evidence fingerprint operations.
type EvidenceService struct {
    client *DigitalEvidenceClient
    redis  *Redis
}

// FingerprintMedia creates a SHA-256 fingerprint for media content (VIP+ users).
func (s *EvidenceService) FingerprintMedia(ctx context.Context, req MediaFingerprintReq) (*FingerprintResult, error) {
    fp := Fingerprint{
        ContentHash: req.ContentHash,
        Metadata: map[string]string{
            "type":      "media",
            "messageId": req.MessageID,
            "source":    "echo_backend",
        },
        SourceDID: req.SenderDID,
    }

    return s.client.SubmitFingerprint(fp)
}

// FingerprintAuditBatch creates a fingerprint for an IPFS audit log batch (Org tier).
func (s *EvidenceService) FingerprintAuditBatch(ctx context.Context, req AuditBatchReq) (*FingerprintResult, error) {
    fp := Fingerprint{
        ContentHash: req.BatchHash,
        Metadata: map[string]string{
            "type":       "audit_batch",
            "ipfsCid":    req.IPFSCID,
            "entryCount": fmt.Sprintf("%d", req.EntryCount),
            "timeFrom":   req.TimeFrom.Format(time.RFC3339),
            "timeTo":     req.TimeTo.Format(time.RFC3339),
        },
        SourceDID: "echo_platform",
    }

    return s.client.SubmitFingerprint(fp)
}

// VerifyFingerprint checks on-chain verification status.
func (s *EvidenceService) VerifyFingerprint(ctx context.Context, eventID string) (*VerificationResult, error) {
    return s.client.VerifyFingerprint(eventID)
}

type MediaFingerprintReq struct {
    ContentHash string `json:"contentHash"` // SHA-256 hex
    MessageID   string `json:"messageId"`
    SenderDID   string `json:"senderDid"`
}

type AuditBatchReq struct {
    BatchHash  string    `json:"batchHash"`
    IPFSCID    string    `json:"ipfsCid"`
    EntryCount int       `json:"entryCount"`
    TimeFrom   time.Time `json:"timeFrom"`
    TimeTo     time.Time `json:"timeTo"`
}
```

---

## 8. Genesis Configuration

```yaml
# configs/genesis.yaml
# Token genesis parameters — used by Scala Currency L1 at snapshot #1
# Go backend reads this for pre-validation and display

genesis:
  total_supply: 1_000_000_000
  decimal_places: 8

  allocations:
    community_rewards:
      amount: 400_000_000
      type: emission_pool
      emission_years: 10
      year_percentages: [20, 16, 13, 11, 9, 7, 6, 6, 6, 6]

    treasury:
      amount: 250_000_000
      type: multi_sig
      signers: 5
      threshold: 3
      initial_distributions:
        pacaswap_liquidity: 100_000_000
        operational_reserve: 50_000_000
        locked_reserve: 100_000_000

    founders:
      amount: 150_000_000
      type: vesting_locks
      cliff_months: 12
      vest_months: 48
      allocations:
        - role: "CEO / Visionary / Product"
          share_percent: 40
          amount: 60_000_000
          did: "" # Set at genesis
        - role: "CTO / Lead iOS Engineer"
          share_percent: 25
          amount: 37_500_000
          did: "" # Set at genesis
        - role: "Scala / Blockchain Lead"
          share_percent: 15
          amount: 22_500_000
          did: "" # Set at genesis
        - role: "Head of Growth / Community"
          share_percent: 10
          amount: 15_000_000
          did: "" # Set at genesis
        - role: "Head of Design / UX"
          share_percent: 10
          amount: 15_000_000
          did: "" # Set at genesis

    future_team:
      amount: 100_000_000
      type: reserved_pool
      release_mechanism: multi_sig
      revert_to_treasury_after_years: 3

    ecosystem:
      amount: 100_000_000
      type: governance_pool
      release_mechanism: governance_vote

  staking_tiers:
    bronze:  { duration_days: 30,  apr: 5.0 }
    silver:  { duration_days: 90,  apr: 8.0 }
    gold:    { duration_days: 180, apr: 12.0 }
    platinum: { duration_days: 365, apr: 15.0 }

  unstaking_cooldown_days: 14
```

---

## 9. Database Migrations

```sql
-- migrations/003_wallet_cache.sql

-- Wallet balance cache (refreshed from metagraph every 5s)
CREATE TABLE wallet_balance_cache (
    did            TEXT PRIMARY KEY,
    total_balance  BIGINT NOT NULL DEFAULT 0,
    available      BIGINT NOT NULL DEFAULT 0,
    staked         BIGINT NOT NULL DEFAULT 0,
    pending_rewards BIGINT NOT NULL DEFAULT 0,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Staking positions (mirror of on-chain TokenLock state)
CREATE TABLE staking_positions (
    id            TEXT PRIMARY KEY,
    did           TEXT NOT NULL,
    amount        BIGINT NOT NULL,
    tier          TEXT NOT NULL,
    locked_until  TIMESTAMPTZ NOT NULL,
    vesting_type  TEXT, -- 'founder' or NULL
    delegated_to  TEXT, -- validator ID or NULL
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_staking_positions_did ON staking_positions(did);

-- Daily reward caps (reset at UTC midnight)
CREATE TABLE daily_reward_caps (
    did          TEXT NOT NULL,
    reward_type  TEXT NOT NULL, -- messaging, referral, staking, payment_rail
    earned_today BIGINT NOT NULL DEFAULT 0,
    cap          BIGINT NOT NULL,
    reset_at     TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (did, reward_type)
);

-- Validator directory (synced from metagraph)
CREATE TABLE validators (
    id              TEXT PRIMARY KEY,
    address         TEXT NOT NULL,
    layer           TEXT NOT NULL, -- currency_l1, data_l1
    uptime_percent  REAL NOT NULL DEFAULT 0,
    commission_pct  REAL NOT NULL DEFAULT 0,
    total_delegated BIGINT NOT NULL DEFAULT 0,
    delegator_count INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- migrations/004_evidence_audit.sql

-- Digital Evidence fingerprint records (Org tier)
CREATE TABLE evidence_fingerprints (
    event_id         TEXT PRIMARY KEY,
    content_hash     TEXT NOT NULL,
    source_type      TEXT NOT NULL, -- media, audit_batch, message, retention_proof
    message_id       TEXT,
    sender_did       TEXT,
    verification_url TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evidence_message ON evidence_fingerprints(message_id);
```

---

## 10. Build & Deploy

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o echo-server ./cmd/echo-server/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/echo-server /usr/local/bin/
COPY --from=builder /app/configs/ /etc/echo/
EXPOSE 8000 8001 8002 8003 8004 8005 8006 8007 8008 8009 8010
ENTRYPOINT ["echo-server"]
```

```yaml
# docker-compose.yaml (local dev)
services:
  echo-backend:
    build: .
    ports: ["8000-8010:8000-8010"]
    environment:
      - REDIS_URL=redis://redis:6379
      - POSTGRES_URL=postgres://echo:echo@postgres:5432/echo
      - NATS_URL=nats://nats:4222
      - METAGRAPH_DATA_L1=http://metagraph:9010
      - METAGRAPH_CURRENCY_L1=http://metagraph:9020
      - METAGRAPH_L0=http://metagraph:9000
      - DIGITAL_EVIDENCE_API_URL=https://api.digitalevidence.constellationnetwork.io
      - DIGITAL_EVIDENCE_API_KEY=${DE_API_KEY}
      - GENESIS_CONFIG=/etc/echo/genesis.yaml
    depends_on: [redis, postgres, nats]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: echo
      POSTGRES_USER: echo
      POSTGRES_PASSWORD: echo
    ports: ["5432:5432"]

  nats:
    image: nats:2-alpine
    ports: ["4222:4222"]
```

---

*Go Backend Implementation Spec v3.0*
*Aligned to: PRD v2.4, Data Layer v3.3, Tokenomics v1.0*
*Status: Implementation-ready for Phase 2*
