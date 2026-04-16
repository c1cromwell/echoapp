# ECHO Tokenomics — Implementation Spec Correlation & Update

## Document Info

| Field | Value |
|-------|-------|
| Version | 1.0 |
| Date | April 15, 2026 |
| Purpose | Resolve all tokenomics discrepancies across PRD, Blueprint IMPL, Scala IMPL, Go Backend IMPL, iOS Frontend IMPL, and Data Layer Architecture |
| Authoritative Source | PRD v2.5.1 changelog (March 31, 2026): "resolved tokenomics conflict (auto-scaling model adopted, daily caps removed)" + ECHO_TOKENOMICS_v2.md (March 26, 2026) for founder allocation |
| Affected Files | SCALA_METAGRAPH_IMPL_v2.md, GO_BACKEND_IMPL_v4.md, ECHO_TOKENOMICS_v2.md, Echo_ECHO_Tokenomics_Founder_Allocation_and_Token_Launch_PRD.md, Echo_PRD2_5_1_Features_Combined_Documents_FIXED.md, DATA_LAYER_ARCHITECTURE_v3_4.md, IOS_IMPL_v4_2.md, Echo_Frontend_2_5_1_FIXED.md |

---

## Part 1: Discrepancy Audit

### CONFLICT 1: Daily Caps vs. Auto-Scaling Reward Model (CRITICAL)

The PRD v2.5.1 changelog explicitly states: **"auto-scaling model adopted, daily caps removed."** This is the authoritative resolution. However, multiple implementation files still contain the daily-cap model.

| File | Current State | Required State |
|------|--------------|----------------|
| **SCALA_METAGRAPH_IMPL_v2.md** | Contains `DailyCaps` object with per-tier caps (Tier 1=0, Tier 2=5, Tier 3=25, Tier 4=50, Tier 5=100). `RewardClaimValidator` calls `validateDailyCap()`. Project structure includes `DailyCapTracker.scala`. | Remove `DailyCaps` object entirely. Replace `validateDailyCap()` with `validateAutoScaleRate()`. Remove `DailyCapTracker.scala` from project structure, replace with `AutoScaleRateTracker.scala`. |
| **GO_BACKEND_IMPL_v4.md** | Has `DailyCapState` and `DailyCapEntry` structs in data model. | Replace with `AutoScaleState` struct containing `CurrentRate`, `DailyBudget`, `DailyUsed`, `TotalActivityWeight`. |
| **ECHO_TOKENOMICS_v2.md** | iOS wallet code snippet: `.updateDailyCap(did: currentDID)` in `claimRewards()` function. Section 1.1 mentions "daily caps" in community rewards description. | Change AtomicAction to `.recordNetworkActivity(did: currentDID)`. Remove "daily caps" language from Section 1.1. |
| **Echo_PRD2_5_1_Features_Combined_Documents_FIXED.md** | Contains BOTH models simultaneously. Old section has AC-TOK-002.4 with per-tier daily caps (Tier 1: 10 ECHO/day, Tier 2: 25, etc.). Newer section has auto-scaling. "Rate of Issuance" section discusses daily cap system as if active. | Remove all daily-cap AC criteria. Keep only auto-scaling version. Rewrite "Rate of Issuance" section to use auto-scaling math. |
| **DATA_LAYER_ARCHITECTURE_v3_4.md** | Section 3.3 Reward Claim: "Validate claim against daily caps" | Change to: "Validate claim against annual emission budget and auto-scaled network rate" |
| **Echo_ECHO_Tokenomics_Founder_Allocation_and_Token_Launch_PRD.md** | AtomicAction terminology: "verify tier + claim reward + update daily cap" | Change to: "verify tier + claim reward + record network activity" |

### CONFLICT 2: Founder Allocation Model (CRITICAL)

Three different founder allocation models exist across files. The authoritative model is from ECHO_TOKENOMICS_v2.md (March 26, 2026) and the Blueprint IMPL: **CEO 10% (100M), co-founders 2% each (20M each)**.

| File | Current Model | Required Model |
|------|--------------|----------------|
| **ECHO_TOKENOMICS.md (v1)** | 15% founders (150M). Unequal split: CEO 6% (60M), CTO 3.75% (37.5M), Scala 2.25% (22.5M), Growth 1.5% (15M), Design 1.5% (15M). Treasury 25% (250M). | **SUPERSEDED by v2. Should be archived or clearly marked as v1 (superseded).** |
| **Echo_ECHO_Tokenomics_Founder_Allocation_and_Token_Launch_PRD.md** | 18% founders (180M). **EQUAL split: 5 founders × 36M each (3.6% each)**. 5-of-5 unanimous approval for treasury and departures. | **INCORRECT.** Must be updated to: CEO 100M (10%), co-founders 20M each (2%). Change 5-of-5 to 3-of-5 for treasury multi-sig and departure revocations. |
| **ECHO_TOKENOMICS_v2.md** | CEO 10% (100M), co-founders 2% each (20M). 3-of-5 multi-sig. | **CORRECT — this is authoritative.** |
| **Blueprint IMPL** | CEO 10% (100M), co-founders 2% each (20M). 3-of-5 multi-sig. | **CORRECT — aligned with v2.** |

### CONFLICT 3: Multi-Sig Threshold (MEDIUM)

| File | Current | Required |
|------|---------|----------|
| **Echo_ECHO_Tokenomics_Founder_Allocation_and_Token_Launch_PRD.md** | 5-of-5 unanimous for treasury and founder departures (AC-TOK-004.5, AC-TOK-004.6, AC-TOK-005.1, AC-TOK-005.2) | 3-of-5 threshold (consistent with v2 tokenomics and Blueprint IMPL) |
| **ECHO_TOKENOMICS_v2.md** | 3-of-5 | Correct |
| **Blueprint IMPL** | 3-of-5 | Correct |

### CONFLICT 4: AtomicAction Description Inconsistency (LOW)

| File | Current Description | Required Description |
|------|-------------------|---------------------|
| **Echo_ECHO_Tokenomics_Founder_Allocation_and_Token_Launch_PRD.md** | "verify tier + claim reward + update daily cap" | "verify tier + claim reward + record network activity" |
| **Blueprint IMPL** | "verify tier + claim reward + record network activity" | Correct |
| **Echo_PRD2_5_1_Features_Combined_Documents_FIXED.md** (old section) | "verify the user's current trust tier on-chain, apply the correct multiplier, record the claim against the daily cap, and update the cap counter" | "verify the user's current trust tier on-chain, apply the correct multiplier, record the claim against the network daily total, and update the auto-scale rate" |

### CONFLICT 5: Scala Daily Cap Values vs. Old PRD Daily Cap Values (NOW MOOT)

Both are wrong because daily caps have been removed entirely, but for the record:

- Old PRD daily caps: Tier 1=10, Tier 2=25, Tier 3=50, Tier 4=100, Tier 5=150
- Scala IMPL daily caps: Tier 1=0, Tier 2=5, Tier 3=25, Tier 4=50, Tier 5=100

These are both removed in the corrected spec below.

### ISSUE 6: ECHO_TOKENOMICS_v2.md Footer Error (LOW)

The footer reads: *"ECHO Tokenomics & Founder Allocation v1.0 / March 7, 2026"*
Should read: *"ECHO Tokenomics & Founder Allocation v2.0 / March 26, 2026"*

---

## Part 2: Corrected Implementation Specs

### 2.1 Scala Metagraph — RewardClaimValidator (REPLACES existing Section 4.3)

```scala
// modules/currency_l1/src/main/scala/echo/currency/validation/RewardClaimValidator.scala

package echo.currency.validation

/**
 * Validates reward claims using the AUTO-SCALING rate model.
 * NO per-user daily caps exist. Every message always earns.
 * The per-message rate auto-scales based on total daily network activity
 * so the annual emission budget is always fully distributed but never exceeded.
 *
 * Rate = Daily Budget ÷ Total Daily Network Activity Weight
 * Each message contributes 1 × sender's trust tier multiplier to the activity weight.
 *
 * Authoritative: PRD v2.5.1 — "auto-scaling model adopted, daily caps removed"
 */
object RewardClaimValidator {

  def validate(claim: RewardClaimTransaction, state: LedgerState, emission: EmissionState): ValidationResult = {
    for {
      _ <- validateRewardType(claim.rewardType)
      _ <- validateTrustTier(claim.did, claim.trustTier, state)
      _ <- validateAutoScaleRate(claim, state, emission)
      _ <- validateEmissionBudget(claim.amount, emission)
      _ <- validateTierMultiplier(claim.amount, claim.trustTier, claim.rewardType)
      _ <- validateAntiGaming(claim, state)
    } yield ValidTransaction
  }

  /**
   * Auto-scale rate validation: ensures the claimed amount does not exceed
   * the auto-scaled rate for this message given current network activity.
   * 
   * Formula: currentRate = dailyBudget / totalDailyActivityWeight
   * Max claim = currentRate × trustTierMultiplier × 1.01 (1% rounding tolerance)
   *
   * No per-user cap — every message always earns. The rate itself adjusts
   * to keep total daily emissions within the daily budget.
   */
  private def validateAutoScaleRate(claim: RewardClaimTransaction, state: LedgerState, emission: EmissionState) = {
    claim.rewardType match {
      case "messaging" =>
        val autoScaleState = state.autoScaleState
        val currentRate = autoScaleState.currentRate
        val tierMultiplier = TrustTier.multiplier(claim.trustTier)
        val maxExpected = (currentRate * tierMultiplier * 1.01).toLong
        if (claim.amount > maxExpected)
          Left(s"Amount exceeds auto-scaled rate: max=$maxExpected, claimed=${claim.amount}, rate=$currentRate")
        else Right(())
      case "referral" =>
        // Referral rewards are fixed (50 ECHO each), exempt from auto-scaling
        val expectedAmount = 50_00000000L // 50 ECHO with 8 decimal places
        if (claim.amount > expectedAmount)
          Left(s"Referral reward exceeds fixed amount: max=$expectedAmount, claimed=${claim.amount}")
        else Right(())
      case _ => Right(())
    }
  }

  /** Global emission budget check: don't exceed this year's allocation. */
  private def validateEmissionBudget(amount: Long, emission: EmissionState) = {
    if (emission.remainingThisYear < amount)
      Left(s"Emission budget exhausted: remaining=${emission.remainingThisYear}")
    else Right(())
  }

  /** Trust tier multiplier validation. */
  private def validateTierMultiplier(amount: Long, tier: Int, rewardType: String) = {
    rewardType match {
      case "referral" => Right(()) // Referrals are fixed, not multiplied
      case _ =>
        val baseRate = RewardType.baseAmount(rewardType)
        val multiplier = TrustTier.multiplier(tier)
        val maxExpected = (baseRate * multiplier * 1.01).toLong
        if (amount > maxExpected)
          Left(s"Amount exceeds tier multiplier: max=$maxExpected, claimed=$amount")
        else Right(())
    }
  }

  /** Anti-gaming: reject suspicious patterns. */
  private def validateAntiGaming(claim: RewardClaimTransaction, state: LedgerState) = {
    val recentClaims = state.recentClaims(claim.did, hours = 1)

    // Velocity check: max 10 claims per hour
    if (recentClaims.size >= 10)
      return Left("Velocity limit: max 10 claims per hour")

    // Duplicate check: same type + amount within 60 seconds
    val isDuplicate = recentClaims.exists { c =>
      c.rewardType == claim.rewardType &&
      c.amount == claim.amount &&
      c.timestamp.plusSeconds(60).isAfter(claim.timestamp)
    }
    if (isDuplicate)
      return Left("Duplicate claim detected")

    Right(())
  }

  private def validateRewardType(rewardType: String) =
    if (RewardType.isValid(rewardType)) Right(())
    else Left(s"Unknown reward type: $rewardType")

  private def validateTrustTier(did: String, claimedTier: Int, state: LedgerState) = {
    val actualTier = state.trustTier(did)
    if (claimedTier != actualTier)
      Left(s"Trust tier mismatch: claimed=$claimedTier, actual=$actualTier")
    else Right(())
  }
}

// NOTE: The DailyCaps object has been REMOVED entirely.
// Daily caps were superseded by the auto-scaling rate model per PRD v2.5.1.
```

### 2.2 Scala Metagraph — AutoScaleRateTracker (REPLACES DailyCapTracker)

```scala
// modules/currency_l1/src/main/scala/echo/currency/state/AutoScaleRateTracker.scala

package echo.currency.state

import java.time.{Instant, LocalDate, ZoneOffset}

/**
 * Tracks the auto-scaling reward rate for the current day.
 *
 * Rate = Daily Budget ÷ Total Daily Activity Weight
 * Activity weight per message = 1.0 × sender's trust tier multiplier
 *
 * As network activity grows, the per-message rate declines — but every
 * message always earns something. Unused budget from low-activity days
 * rolls forward within the same calendar year.
 *
 * Replaces DailyCapTracker (removed per PRD v2.5.1).
 */
class AutoScaleRateTracker(emissionTracker: EmissionTracker) {

  case class AutoScaleState(
    date: LocalDate,
    totalActivityWeight: Double,    // Sum of all trust-tier-weighted messages today
    budgetUsedToday: Long,          // Total ECHO distributed today
    currentRate: Long,              // Current per-message base rate (8 decimal places)
    rolloverBudget: Long            // Unused budget rolled forward from previous days this year
  )

  private var state: AutoScaleState = AutoScaleState(
    date = LocalDate.now(ZoneOffset.UTC),
    totalActivityWeight = 0.0,
    budgetUsedToday = 0L,
    currentRate = calculateInitialRate(),
    rolloverBudget = 0L
  )

  /** Calculate the current auto-scaled rate for a message claim. */
  def currentRate: Long = {
    maybeRolloverDay()
    state.currentRate
  }

  /** Record a reward claim and recalculate the auto-scale rate. */
  def recordClaim(amount: Long, trustTierMultiplier: Double): Unit = {
    maybeRolloverDay()
    state = state.copy(
      totalActivityWeight = state.totalActivityWeight + trustTierMultiplier,
      budgetUsedToday = state.budgetUsedToday + amount,
      currentRate = recalculateRate(state.totalActivityWeight + trustTierMultiplier)
    )
  }

  /** Get the effective daily budget including rollover from low-activity days. */
  def effectiveDailyBudget: Long = emissionTracker.dailyBudget + state.rolloverBudget

  /** Get the remaining budget for today. */
  def remainingToday: Long = {
    val effective = effectiveDailyBudget
    val remaining = effective - state.budgetUsedToday
    if (remaining < 0) 0L else remaining
  }

  /** Recalculate rate based on current activity weight. */
  private def recalculateRate(activityWeight: Double): Long = {
    if (activityWeight <= 0) calculateInitialRate()
    else (effectiveDailyBudget.toDouble / activityWeight).toLong
  }

  /** Initial rate when no activity has occurred (target: 0.1 ECHO/message). */
  private def calculateInitialRate(): Long = {
    val targetRate = 0_10000000L // 0.1 ECHO with 8 decimal places
    targetRate
  }

  /** Roll over unused budget at day boundary. */
  private def maybeRolloverDay(): Unit = {
    val today = LocalDate.now(ZoneOffset.UTC)
    if (today.isAfter(state.date)) {
      val unused = effectiveDailyBudget - state.budgetUsedToday
      val rollover = if (unused > 0) unused else 0L
      state = AutoScaleState(
        date = today,
        totalActivityWeight = 0.0,
        budgetUsedToday = 0L,
        currentRate = calculateInitialRate(),
        rolloverBudget = rollover // Rolls forward within same calendar year
      )
    }
  }

  /** Public state for DAG Explorer and API queries. */
  def publicState: Map[String, Any] = Map(
    "date" -> state.date.toString,
    "totalActivityWeight" -> state.totalActivityWeight,
    "currentRate" -> state.currentRate,
    "dailyBudget" -> emissionTracker.dailyBudget,
    "effectiveDailyBudget" -> effectiveDailyBudget,
    "budgetUsedToday" -> state.budgetUsedToday,
    "remainingToday" -> remainingToday
  )
}
```

### 2.3 Scala Metagraph — Updated Project Structure (Section 2)

Replace the project structure references to `DailyCapTracker.scala` with `AutoScaleRateTracker.scala`:

```
├── currency_l1/
│   └── src/main/scala/echo/currency/
│       ├── validation/
│       │   ├── RewardClaimValidator.scala     # Auto-scale rate validation, tier multipliers
│       │   ...
│       └── state/
│           ├── EmissionTracker.scala           # Year/day emission accounting
│           └── AutoScaleRateTracker.scala      # Per-day auto-scaling rate state (REPLACES DailyCapTracker)
```

### 2.4 Go Backend — Updated Data Model (REPLACES DailyCapState in Section 4)

```go
// models/token.go — CORRECTED data model

// AutoScaleState replaces DailyCapState.
// Per PRD v2.5.1: auto-scaling model adopted, daily caps removed.
type AutoScaleState struct {
    CurrentRate         int64   `json:"currentRate"`         // Current per-message base rate (8 decimals)
    DailyBudget         int64   `json:"dailyBudget"`         // Today's emission budget from annual curve
    EffectiveDailyBudget int64  `json:"effectiveDailyBudget"` // Including rollover from low-activity days
    BudgetUsedToday     int64   `json:"budgetUsedToday"`     // Total distributed today
    RemainingToday      int64   `json:"remainingToday"`      // Budget remaining today
    TotalActivityWeight float64 `json:"totalActivityWeight"` // Sum of tier-weighted messages today
    LastUpdated         string  `json:"lastUpdated"`         // ISO timestamp of last rate recalculation
}

// REMOVED: DailyCapState, DailyCapEntry structs
// These were part of the daily-cap model which has been superseded.
```

### 2.5 Go Backend — Updated Reward Claim Service

```go
// rewards/claim.go — CORRECTED reward claim pre-validation

package rewards

import (
    "fmt"
    "time"
)

// PreValidateRewardClaim validates a reward claim against the auto-scaling
// rate model before submitting to Currency L1 for on-chain validation.
//
// IMPORTANT: This replaces the previous daily-cap pre-validation.
// Per PRD v2.5.1: auto-scaling model adopted, daily caps removed.
func PreValidateRewardClaim(
    claim RewardClaim,
    emission *EmissionSchedule,
    autoScaleRate int64,
    trustTierMultiplier float64,
) error {
    // 1. Check annual emission budget
    if emission.RemainingToday(claim.ClaimedToday) <= 0 {
        return fmt.Errorf("daily emission budget exhausted")
    }

    // 2. Validate against auto-scaled rate (messaging only)
    if claim.RewardType == "messaging" {
        maxReward := int64(float64(autoScaleRate) * trustTierMultiplier * 1.01) // 1% rounding tolerance
        if claim.Amount > maxReward {
            return fmt.Errorf("amount %d exceeds auto-scaled rate max %d", claim.Amount, maxReward)
        }
    }

    // 3. Validate referral fixed amount
    if claim.RewardType == "referral" {
        fixedAmount := int64(50_00000000) // 50 ECHO
        if claim.Amount > fixedAmount {
            return fmt.Errorf("referral amount %d exceeds fixed %d", claim.Amount, fixedAmount)
        }
    }

    // 4. Anti-gaming velocity check (pre-validation; L1 is authoritative)
    // Backend rate-limits to 10 claims per hour per DID
    // This is enforced via Redis rate limiter in the API handler

    return nil
}
```

### 2.6 Go Backend — Updated Emission Status API Response

```go
// api/handlers/token_handlers.go — Updated emission status endpoint

// GET /tokens/emission/status
// Returns current emission state including auto-scaled rate.
type EmissionStatusResponse struct {
    CurrentYear           int     `json:"currentYear"`
    AnnualBudget          int64   `json:"annualBudget"`
    DistributedThisYear   int64   `json:"distributedThisYear"`
    RemainingThisYear     int64   `json:"remainingThisYear"`
    DailyBudget           int64   `json:"dailyBudget"`
    EffectiveDailyBudget  int64   `json:"effectiveDailyBudget"`
    DistributedToday      int64   `json:"distributedToday"`
    CurrentAutoScaledRate int64   `json:"currentAutoScaledRate"` // Per-message base rate
    TotalActivityToday    float64 `json:"totalActivityToday"`    // Tier-weighted activity
    PoolRemaining         int64   `json:"poolRemaining"`         // Total remaining in 400M pool
}
```

### 2.7 iOS Frontend — Updated AtomicAction in WalletViewModel

Replace the `claimRewards()` function in all iOS code references:

```swift
// WalletViewModel.swift — CORRECTED reward claim

// Claim rewards via AtomicAction (verify tier + claim + record network activity)
// NOTE: .updateDailyCap removed — daily caps replaced by auto-scaling model (PRD v2.5.1)
func claimRewards() async throws {
    try await stargazer.submitAtomicAction([
        .verifyTrustTier(did: currentDID),
        .claimRewards(did: currentDID, types: dailyRewards.claimableTypes),
        .recordNetworkActivity(did: currentDID)  // Records against auto-scale rate
    ])
    await loadWallet()  // Refresh
}
```

### 2.8 Data Layer Architecture — Updated Reward Claim Flow (REPLACES Section 3.3)

```
1. iOS App → POST /tokens/rewards/claim (type, evidence)

2. Go Backend (Rewards Service)
   ├─ Validate claim against annual emission budget and auto-scaled network rate
   ├─ Apply trust tier reward multiplier: Tier 1 (1.0x), Tier 2 (1.2x), Tier 3 (1.5x), Tier 4 (2.0x), Tier 5 (3.0x)
   ├─ Pre-validate against anti-gaming rules (velocity checks, repeat claims)
   └─ Add to reward batch queue

3. Batch Processing (every 30 seconds)
   ├─ Construct reward batch transaction
   └─ Submit to Currency L1

4. Currency L1
   ├─ Validate each reward (auto-scale rate, emission budget, eligibility, signature)
   ├─ Update token balances
   ├─ Update auto-scale rate state (recalculate rate based on new activity weight)
   └─ Package into L1 block

5. Metagraph L0 → Global L0 → Finality

6. Confirmation
   ├─ Backend cache updated with new balance and current auto-scaled rate
   └─ Push balance update to iOS via WebSocket
```

---

## Part 3: Corrected Founder Allocation for PRD

The standalone tokenomics PRD file (Echo_ECHO_Tokenomics_Founder_Allocation_and_Token_Launch_PRD.md) must be corrected to match the authoritative v2 model.

### 3.1 REQ-TOK-004 — Corrected Acceptance Criteria

```
* AC-TOK-004.1: At genesis, the system shall create five founder TokenLock 
  positions: Founder 1 (CEO/Visionary/Product) receives 100M ECHO (10% of 
  total supply); Founders 2–5 (co-founders) receive 20M ECHO each (2% of 
  total supply each). Total founders: 180M ECHO (18%).

* AC-TOK-004.5: Pre-cliff departure: the entire TokenLock balance is returned 
  to the Future Team pool via 3-of-5 founder multi-sig revocation transaction.

* AC-TOK-004.6: Post-cliff departure: vested tokens are released; unvested 
  balance is returned to the Future Team pool via 3-of-5 founder multi-sig.
```

### 3.2 REQ-TOK-005 — Corrected Treasury Controls

```
* AC-TOK-005.1: The 220M treasury at genesis shall be subdivided as: 80M to 
  PacaSwap liquidity seeding (ECHO/DAG and ECHO/USDC pools), 50M to 
  operational reserve (bridged to stablecoins via Base bridge), and 90M locked 
  in a 3-of-5 founder multi-sig for Phase 5–6 operations.

* AC-TOK-005.2: During Phases 1–3, treasury disbursements require 3-of-5 
  founder multi-sig authorization. From Phase 4 onward, disbursements require 
  a governance vote passing the defined threshold for the disbursement type.
```

### 3.3 Founder Allocation Rationale (Corrected)

The corrected rationale for the PRD should match ECHO_TOKENOMICS_v2.md:

> The CEO's 10% of total supply (100M ECHO) reflects the totality of pre-team contributions: full product architecture, 5+ PRD versions, backend/iOS/API architecture documents, tokenomics design, governance model, and all strategic decisions before any co-founder joined. The co-founder 2% equal split (20M each) provides a clean, competitive offer that avoids internal allocation disputes. The insider total (founders 18% + future team 10% = 28%) stays below the industry average of 35–45% inclusive of VC allocation. Community + ecosystem retains 50% — the majority.

---

## Part 4: Corrected Metagraph L1 Validation Rules Table

This table appears in multiple files (Data Layer Architecture, Combined Blueprint, etc.) and should be consistent everywhere:

| Validation Rule | L1 Layer | Logic |
|----------------|----------|-------|
| Annual emission enforcement | Currency L1 | Reject reward claims that would cause Year-N total distributions to exceed the Year-N emission cap. Per-message rate auto-scales based on total daily network activity weight. **No per-user daily cap.** |
| Auto-scale rate validation | Currency L1 | Validate claimed amount does not exceed current auto-scaled rate × trust tier multiplier. Rate = Daily Budget ÷ Total Daily Activity Weight. |
| Trust-tier multiplier | Currency L1 | Apply correct multiplier based on cached trust tier; reject mismatched multipliers (1.0x/1.2x/1.5x/2.0x/3.0x) |
| Anti-gaming | Currency L1 | Detect and reject suspicious reward patterns (velocity checks: max 10 claims/hour, duplicate detection within 60s) |
| Founder vesting | Currency L1 | 12-month cliff, 1/36th monthly vesting, 14-day WithdrawLock cooldown. CEO: 100M, co-founders: 20M each. |
| Treasury controls | Currency L1 | 3-of-5 founder multi-sig (Phases 1–3), governance vote (Phase 4+) |

---

## Part 5: Summary of All Required Changes by File

### SCALA_METAGRAPH_IMPL_v2.md
1. **Remove** `DailyCaps` object entirely (Section 4.3)
2. **Remove** `validateDailyCap()` from `RewardClaimValidator` — replace with `validateAutoScaleRate()`
3. **Remove** `DailyCapTracker.scala` from project structure — replace with `AutoScaleRateTracker.scala`
4. **Add** `AutoScaleRateTracker` class (Section 4.4 alternative)
5. **Update** `RewardClaimValidator` to use auto-scaling validation (full replacement in Part 2.1 above)
6. **Update** file header comment in Section 1 to remove "reward caps" terminology, use "auto-scale rate validation"

### GO_BACKEND_IMPL_v4.md
1. **Remove** `DailyCapState` and `DailyCapEntry` structs from data model
2. **Add** `AutoScaleState` struct (Part 2.4 above)
3. **Update** reward claim pre-validation to use auto-scale model (Part 2.5 above)
4. **Update** emission status API response to include auto-scaled rate fields (Part 2.6 above)

### ECHO_TOKENOMICS_v2.md
1. **Fix** iOS code snippet: change `.updateDailyCap(did:)` to `.recordNetworkActivity(did:)`
2. **Fix** Section 1.1 community rewards description: remove "with daily caps" language
3. **Fix** footer: change "v1.0 / March 7, 2026" to "v2.0 / March 26, 2026"

### Echo_ECHO_Tokenomics_Founder_Allocation_and_Token_Launch_PRD.md
1. **Fix** AC-TOK-004.1: change equal 36M split to CEO 100M + co-founders 20M each
2. **Fix** AC-TOK-004.5, AC-TOK-004.6: change 5-of-5 to 3-of-5
3. **Fix** AC-TOK-005.1, AC-TOK-005.2: change 5-of-5 to 3-of-5
4. **Fix** AtomicAction terminology: "update daily cap" → "record network activity"
5. **Fix** Founder Allocation Rationale section: replace equal-split rationale with CEO/co-founder rationale

### Echo_PRD2_5_1_Features_Combined_Documents_FIXED.md
1. **Remove** old daily-cap AC-TOK-002.4 with per-tier caps (Tier 1: 10, Tier 2: 25, etc.)
2. **Remove** AC-TOK-003.5 reference to "record the claim against the daily cap"
3. **Rewrite** "Rate of Issuance vs. Anti-Inflation Controls" section to use auto-scaling math
4. **Ensure** only the auto-scaling version of REQ-TOK-002 and REQ-TOK-003 remains

### DATA_LAYER_ARCHITECTURE_v3_4.md
1. **Update** Section 3.3 Reward Claim flow: "Validate claim against daily caps" → "Validate claim against annual emission budget and auto-scaled network rate"
2. **Update** Metagraph L1 Validation Rules table to match Part 4 above

### IOS_IMPL_v4_2.md / Echo_Frontend_2_5_1_FIXED.md
1. **Update** any remaining `.updateDailyCap()` references to `.recordNetworkActivity()`
2. **Verify** `DailyRewards` struct uses auto-scaling fields (the Frontend 2.5.1 file is already mostly correct with `currentAutoScaledRate` and `networkDailyBudget`)

---

## Part 6: Version Lineage (For Reference)

| Document | Version | Date | Key Changes |
|----------|---------|------|-------------|
| ECHO_TOKENOMICS.md | 1.0 | March 7, 2026 | Initial: 15% founders (unequal split), 25% treasury, daily caps |
| ECHO_TOKENOMICS_v2.md | 2.0 | March 26, 2026 | 18% founders (CEO 10%, co-founders 2%), 22% treasury, trust-tier governance |
| PRD v2.5 | 2.5 | March 26, 2026 | Adopted v2 tokenomics into PRD |
| PRD v2.5.1 | 2.5.1 | March 31, 2026 | **Resolved tokenomics conflict: auto-scaling adopted, daily caps removed** |
| Blueprint IMPL | — | Post-v2.5.1 | Aligned with auto-scaling + CEO 10% model |
| **This Update** | — | April 15, 2026 | Correlates all implementation files to authoritative PRD v2.5.1 + Tokenomics v2.0 |

---

*ECHO Tokenomics Implementation Spec Correlation & Update*
*April 15, 2026*
*Status: All implementation files require the changes documented above to achieve full consistency with PRD v2.5.1 and ECHO Tokenomics v2.0.*
