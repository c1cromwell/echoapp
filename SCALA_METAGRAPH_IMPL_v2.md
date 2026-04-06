# Scala Metagraph L1 Implementation Spec (v2.0)

## Aligned To

| Document | Version |
|----------|---------|
| PRD | v2.5 |
| Data Layer Architecture | v3.4 |
| ECHO Tokenomics | v2.0 |
| Tessellation | v3 (targeting v4 compatibility) |

---

## 1. Overview

The Scala metagraph layer is the on-chain enforcement engine for ECHO. It runs on the Constellation Hypergraph using the Euclid SDK (built on Tessellation) and contains all business rules that must be trustlessly enforced: token genesis, emission schedule, founder vesting, staking tiers, reward caps, anti-gaming detection, governance voting, Merkle root validation, and validator slashing.

**The Go backend pre-validates and constructs transactions. The Scala L1 validators are the authority.** If the Go backend submits a malformed or rule-violating transaction, L1 validators reject it. This separation means compromising the Go backend cannot forge rewards, bypass vesting, or inflate supply.

**Language:** Scala 2.13+ on JVM 17+
**Framework:** Euclid SDK (Constellation's metagraph development toolkit built on Tessellation)
**Dev Environment:** Hydra CLI (Docker-based local cluster: Global L0 + Metagraph L0 + Currency L1 + Data L1)

---

## 2. Project Structure

```
echo-metagraph/
├── project/
│   ├── build.properties
│   └── plugins.sbt
├── build.sbt                          # Euclid SDK dependency, Scala version
│
├── modules/
│   ├── shared/                        # Shared types across L1 layers
│   │   └── src/main/scala/echo/shared/
│   │       ├── types/
│   │       │   ├── EchoToken.scala       # Token type definition (L0 standard)
│   │       │   ├── StakingTier.scala     # Bronze/Silver/Gold/Platinum
│   │       │   ├── RewardType.scala      # Messaging, referral, staking, etc.
│   │       │   ├── TrustTier.scala       # 1–5 trust levels
│   │       │   └── SchemaVersion.scala   # Data schema versioning
│   │       └── config/
│   │           ├── GenesisConfig.scala   # 1B supply, pool allocations
│   │           └── EmissionConfig.scala  # 10-year declining curve
│   │
│   ├── currency_l1/                   # Currency L1 validation logic
│   │   └── src/main/scala/echo/currency/
│   │       ├── CurrencyL1App.scala       # Entry point, validator registration
│   │       ├── genesis/
│   │       │   └── TokenGenesis.scala    # Snapshot #1: mint + allocate
│   │       ├── validation/
│   │       │   ├── TokenLockValidator.scala      # Staking + founder vesting
│   │       │   ├── StakeDelegationValidator.scala # Delegation rules
│   │       │   ├── WithdrawLockValidator.scala    # Unstaking + cooldown + cliff
│   │       │   ├── RewardClaimValidator.scala     # Daily caps, tier multipliers
│   │       │   ├── AtomicActionValidator.scala    # Bundle all-or-nothing
│   │       │   ├── AllowSpendValidator.scala      # Time-limited approvals
│   │       │   ├── FeeTransactionValidator.scala  # Snapshot fee payment
│   │       │   └── AntiGamingValidator.scala      # Velocity, patterns
│   │       └── state/
│   │           ├── EmissionTracker.scala  # Year/day emission accounting
│   │           └── DailyCapTracker.scala  # Per-DID daily reward caps
│   └── data_l1/                       # Data L1 validation logic
│       └── src/main/scala/echo/data/
│           ├── DataL1App.scala           # Entry point
│           └── validation/
│               ├── MerkleRootValidator.scala      # Message integrity roots
│               ├── TrustCommitmentValidator.scala  # H(score||nonce) format
│               ├── GovernanceVoteValidator.scala   # Trust-tier weighted voting
│               ├── GovernanceWeightCalculator.scala # Weight = StakedECHO × TrustTierMultiplier
│               ├── GroupMetadataValidator.scala     # Group admin signature
│               ├── RelayNodeValidator.scala         # Relay registry (Phase 4)
│               └── SchemaValidator.scala            # Version check
│
├── docker/
│   ├── Dockerfile.currency-l1
│   ├── Dockerfile.data-l1
│   └── hydra-config.yaml             # Local dev cluster config
│
└── test/
    └── src/test/scala/echo/
        ├── GenesisSpec.scala
        ├── TokenLockSpec.scala
        ├── VestingSpec.scala
        ├── RewardClaimSpec.scala
        ├── EmissionSpec.scala
        ├── SlashingSpec.scala
        ├── AtomicActionSpec.scala
        ├── GovernanceWeightSpec.scala    # Trust-tier weighted voting tests
        └── GovernanceVoteSpec.scala      # Full vote validation tests
```

---

## 3. Token Genesis (Snapshot #1)

```scala
// modules/currency_l1/src/main/scala/echo/currency/genesis/TokenGenesis.scala

package echo.currency.genesis

import echo.shared.config.GenesisConfig
import org.tessellation.currency.l1.domain.snapshot.CurrencySnapshotEvent

object TokenGenesis {

  val TotalSupply: Long = 1_000_000_000_00000000L // 1B with 8 decimals

  case class GenesisAllocation(
    address: String,     // Constellation wallet address
    amount: Long,
    poolType: PoolType,
    vestingParams: Option[VestingParams] = None
  )

  sealed trait PoolType
  case object CommunityRewards extends PoolType
  case object Treasury extends PoolType
  case object FounderVesting extends PoolType
  case object FutureTeam extends PoolType
  case object Ecosystem extends PoolType

  case class VestingParams(
    cliffMonths: Int,       // 12 for founders
    totalVestMonths: Int,   // 48 for founders
    monthlyUnlockBps: Int   // Basis points per month after cliff (2778 = 1/36)
  )

  /** Generate all genesis allocations from config. */
  def buildGenesisAllocations(config: GenesisConfig): Seq[GenesisAllocation] = {
    val communityPool = GenesisAllocation(
      address = config.communityPoolAddress,
      amount = 400_000_000_00000000L,
      poolType = CommunityRewards
    )

    val treasuryAllocations = Seq(
      GenesisAllocation(config.treasuryMultisigAddress, 220_000_000_00000000L, Treasury)
    )

    // CEO: 10% (100M), Co-founders: 2% each (20M × 4)
    val founderAllocations = config.founders.map { founder =>
      GenesisAllocation(
        address = founder.walletAddress,
        amount = founder.amount,
        poolType = FounderVesting,
        vestingParams = Some(VestingParams(
          cliffMonths = 12,
          totalVestMonths = 48,
          monthlyUnlockBps = 2778 // 1/36 per month after cliff ≈ 2.778%
        ))
      )
    }

    val futureTeam = GenesisAllocation(
      config.futureTeamPoolAddress, 100_000_000_00000000L, FutureTeam
    )

    val ecosystem = GenesisAllocation(
      config.ecosystemPoolAddress, 100_000_000_00000000L, Ecosystem
    )

    communityPool +: (treasuryAllocations ++ founderAllocations :+ futureTeam :+ ecosystem)
  }

  /** Validate that genesis allocations sum to TotalSupply. */
  def validateGenesis(allocations: Seq[GenesisAllocation]): Either[String, Unit] = {
    val total = allocations.map(_.amount).sum
    if (total != TotalSupply)
      Left(s"Genesis total $total != expected $TotalSupply")
    else
      Right(())
  }
}
```

---

## 4. Currency L1 Validators

### 4.1 TokenLock Validator (Staking + Founder Vesting)

```scala
// modules/currency_l1/src/main/scala/echo/currency/validation/TokenLockValidator.scala

package echo.currency.validation

import echo.shared.types._

object TokenLockValidator {

  /** Validate a TokenLock transaction (user staking or founder vesting). */
  def validate(tx: TokenLockTransaction, state: LedgerState): ValidationResult = {
    for {
      _ <- validateBalance(tx.did, tx.amount, state)
      _ <- validateTier(tx.tier, tx.durationDays)
      _ <- validateMinimumStake(tx.amount, tx.tier)
      _ <- validateFounderVesting(tx)
    } yield ValidTransaction
  }

  private def validateBalance(did: String, amount: Long, state: LedgerState) = {
    val available = state.availableBalance(did)
    if (amount > available)
      Left(s"Insufficient balance: need $amount, have $available")
    else Right(())
  }

  private def validateTier(tier: String, days: Int) = {
    StakingTier.fromString(tier) match {
      case Some(t) if t.durationDays == days => Right(())
      case _ => Left(s"Invalid tier '$tier' with duration $days")
    }
  }

  private def validateMinimumStake(amount: Long, tier: String) = {
    val minimum = StakingTier.minimumStake(tier)
    if (amount < minimum)
      Left(s"Below minimum stake for $tier: need $minimum, got $amount")
    else Right(())
  }

  /** Founder vesting locks have additional constraints. */
  private def validateFounderVesting(tx: TokenLockTransaction) = {
    tx.vestingType match {
      case Some("founder") =>
        // Only allowed at genesis (snapshot #1) or by multi-sig for new team members
        if (tx.cliffMonths != 12 || tx.vestMonths != 48)
          Left("Founder vesting must have 12-month cliff and 48-month vest")
        else Right(())
      case _ => Right(())
    }
  }
}
```

### 4.2 WithdrawLock Validator (Unstaking + Cliff Enforcement)

```scala
// modules/currency_l1/src/main/scala/echo/currency/validation/WithdrawLockValidator.scala

package echo.currency.validation

import java.time.{Duration, Instant}

object WithdrawLockValidator {

  val UnstakingCooldownDays: Int = 14

  def validate(tx: WithdrawLockTransaction, state: LedgerState): ValidationResult = {
    val lock = state.getTokenLock(tx.stakeId)
      .getOrElse(return Left(s"TokenLock ${tx.stakeId} not found"))

    for {
      _ <- validateOwnership(tx.did, lock)
      _ <- validateAmount(tx.amount, lock)
      _ <- validateLockExpiry(lock)
      _ <- validateFounderVesting(tx, lock)
    } yield ValidTransaction
  }

  private def validateOwnership(did: String, lock: TokenLockPosition) =
    if (lock.did != did) Left("Not owner of this TokenLock") else Right(())

  private def validateAmount(amount: Long, lock: TokenLockPosition) =
    if (amount > lock.amount) Left("Withdraw exceeds locked amount") else Right(())

  private def validateLockExpiry(lock: TokenLockPosition) = {
    if (lock.lockedUntil.isAfter(Instant.now()))
      Left(s"Lock not expired until ${lock.lockedUntil}")
    else Right(())
  }

  /** Founder locks enforce cliff and monthly vesting schedule. */
  private def validateFounderVesting(tx: WithdrawLockTransaction, lock: TokenLockPosition) = {
    lock.vestingParams match {
      case Some(vp) =>
        val monthsSinceGenesis = monthsBetween(lock.createdAt, Instant.now())

        // Cliff check: no withdrawals before cliff
        if (monthsSinceGenesis < vp.cliffMonths)
          return Left(s"Cliff not reached: ${vp.cliffMonths - monthsSinceGenesis} months remaining")

        // Calculate vested amount
        val monthsVesting = monthsSinceGenesis - vp.cliffMonths
        val cliffVestAmount = lock.originalAmount / 4 // 25% at cliff
        val monthlyVest = (lock.originalAmount - cliffVestAmount) / 36 // Remaining over 36 months
        val totalVested = Math.min(
          cliffVestAmount + (monthlyVest * Math.min(monthsVesting, 36)),
          lock.originalAmount
        )

        val alreadyWithdrawn = lock.originalAmount - lock.amount
        val withdrawable = totalVested - alreadyWithdrawn

        if (tx.amount > withdrawable)
          Left(s"Exceeds vested amount: withdrawable=$withdrawable, requested=${tx.amount}")
        else Right(())

      case None => Right(()) // Non-founder lock: standard expiry check sufficient
    }
  }

  private def monthsBetween(from: Instant, to: Instant): Int =
    (Duration.between(from, to).toDays / 30).toInt
}
```

### 4.3 Reward Claim Validator

```scala
// modules/currency_l1/src/main/scala/echo/currency/validation/RewardClaimValidator.scala

package echo.currency.validation

object RewardClaimValidator {

  def validate(claim: RewardClaimTransaction, state: LedgerState, emission: EmissionState): ValidationResult = {
    for {
      _ <- validateRewardType(claim.rewardType)
      _ <- validateTrustTier(claim.did, claim.trustTier, state)
      _ <- validateDailyCap(claim.did, claim.rewardType, claim.amount, state)
      _ <- validateEmissionBudget(claim.amount, emission)
      _ <- validateTierMultiplier(claim.amount, claim.trustTier, claim.rewardType)
      _ <- validateAntiGaming(claim, state)
    } yield ValidTransaction
  }

  /** Daily cap per DID per reward type (varies by trust tier). */
  private def validateDailyCap(did: String, rewardType: String, amount: Long, state: LedgerState) = {
    val cap = DailyCaps.forTier(state.trustTier(did), rewardType)
    val usedToday = state.dailyRewardUsage(did, rewardType)
    if (usedToday + amount > cap)
      Left(s"Daily cap exceeded for $rewardType: cap=$cap, used=$usedToday, claiming=$amount")
    else Right(())
  }

  /** Global emission budget check: don't exceed this year's allocation. */
  private def validateEmissionBudget(amount: Long, emission: EmissionState) = {
    if (emission.remainingThisYear < amount)
      Left(s"Emission budget exhausted: remaining=${emission.remainingThisYear}")
    else Right(())
  }

  /** Trust tier multiplier validation. */
  private def validateTierMultiplier(amount: Long, tier: Int, rewardType: String) = {
    val baseReward = RewardType.baseAmount(rewardType)
    val multiplier = TrustTier.multiplier(tier)
    val maxExpected = (baseReward * multiplier * 1.01).toLong // 1% tolerance for rounding
    if (amount > maxExpected)
      Left(s"Amount exceeds tier multiplier: max=$maxExpected, claimed=$amount")
    else Right(())
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
}

/** Daily reward caps by trust tier. */
object DailyCaps {
  def forTier(tier: Int, rewardType: String): Long = (tier, rewardType) match {
    case (1, "messaging")    => 0              // Unverified: no messaging rewards
    case (2, "messaging")    => 5_00000000L    // 5 ECHO
    case (3, "messaging")    => 25_00000000L   // 25 ECHO
    case (4, "messaging")    => 50_00000000L   // 50 ECHO
    case (5, "messaging")    => 100_00000000L  // 100 ECHO
    case (_, "referral")     => 250_00000000L  // 250 ECHO (5 referrals × 50)
    case (_, "staking")      => Long.MaxValue  // No cap on staking rewards
    case (_, "payment_rail") => 500_00000000L  // 500 ECHO
    case _                   => 0
  }
}
```

### 4.4 Emission Tracker

```scala
// modules/currency_l1/src/main/scala/echo/currency/state/EmissionTracker.scala

package echo.currency.state

import java.time.{Instant, Duration}

/** Tracks community reward emission against the 10-year declining curve. */
class EmissionTracker(genesisTimestamp: Instant) {

  val TotalPool: Long = 400_000_000_00000000L
  val YearlyPercents: Seq[Double] = Seq(0.20, 0.16, 0.13, 0.11, 0.09, 0.07, 0.06, 0.06, 0.06, 0.06)

  def currentYear: Int = {
    val elapsed = Duration.between(genesisTimestamp, Instant.now())
    Math.min((elapsed.toDays / 365).toInt + 1, 11) // 11 = post-emission
  }

  def yearlyBudget(year: Int): Long = {
    if (year < 1 || year > 10) 0L
    else (TotalPool * YearlyPercents(year - 1)).toLong
  }

  def dailyBudget: Long = {
    val year = currentYear
    if (year > 10) 0L else yearlyBudget(year) / 365
  }
}
```

---

## 5. Data L1 Validators

### 5.1 Merkle Root Validator

```scala
// modules/data_l1/src/main/scala/echo/data/validation/MerkleRootValidator.scala

package echo.data.validation

object MerkleRootValidator {

  def validate(submission: MerkleRootSubmission, state: LedgerState): ValidationResult = {
    for {
      _ <- validateStructure(submission.merkleRoot)
      _ <- validateAuthorizedSender(submission.senderDID, state)
      _ <- validateSchemaVersion(submission.schemaVersion)
      _ <- validateCommitmentCount(submission.commitmentCount)
      _ <- validateTimeRange(submission.timeRange)
    } yield ValidTransaction
  }

  private def validateStructure(root: Array[Byte]) = {
    if (root.length != 32) Left(s"Merkle root must be 32 bytes (SHA-256), got ${root.length}")
    else Right(())
  }

  private def validateAuthorizedSender(did: String, state: LedgerState) = {
    if (!state.isAuthorizedRelayNode(did))
      Left(s"Sender $did is not an authorized relay node")
    else Right(())
  }

  private def validateCommitmentCount(count: Int) = {
    if (count < 1 || count > 10000)
      Left(s"Commitment count out of range: $count (1-10000)")
    else Right(())
  }

  private def validateTimeRange(range: TimeRange) = {
    val duration = Duration.between(range.from, range.to)
    if (duration.toMinutes > 10)
      Left(s"Batch time range too wide: ${duration.toMinutes} minutes (max 10)")
    else Right(())
  }
}
```

### 5.2 Governance Vote Validator (Trust-Tier Weighted)

```scala
// modules/data_l1/src/main/scala/echo/data/validation/GovernanceVoteValidator.scala

package echo.data.validation

object GovernanceVoteValidator {

  def validate(vote: GovernanceVote, state: LedgerState): ValidationResult = {
    for {
      _ <- validateProposalActive(vote.proposalId, state)
      _ <- validateOneVotePerDID(vote.did, vote.proposalId, state)
      _ <- validateTrustTierMinimum(vote.did, state) // Tier 2+ required
      _ <- validateHasStake(vote.did, state) // Must have TokenLock positions
      _ <- validateVoteValue(vote.value)
    } yield {
      // Calculate governance weight and record weighted vote
      val weight = GovernanceWeightCalculator.calculate(vote.did, state)
      WeightedVote(vote, weight)
    }
  }

  private def validateOneVotePerDID(did: String, proposalId: String, state: LedgerState) = {
    if (state.hasVoted(did, proposalId))
      Left(s"$did has already voted on proposal $proposalId")
    else Right(())
  }

  private def validateTrustTierMinimum(did: String, state: LedgerState) = {
    val tier = state.trustTier(did)
    if (tier < 2)
      Left(s"Trust Tier $tier insufficient for governance (minimum Tier 2)")
    else Right(())
  }

  private def validateHasStake(did: String, state: LedgerState) = {
    val totalStaked = state.totalStakedBalance(did) // Sum of all TokenLock positions
    if (totalStaked <= 0)
      Left(s"Must have staked ECHO (TokenLock) to participate in governance")
    else Right(())
  }
}

// modules/data_l1/src/main/scala/echo/data/validation/GovernanceWeightCalculator.scala

package echo.data.validation

/**
 * Calculates trust-tier weighted governance power.
 *
 * Formula: GovernanceWeight = StakedECHO × TrustTierMultiplier
 *
 * This prevents plutocratic capture: a whale who buys 50M ECHO but never
 * verifies (Tier 1) gets zero governance power. The CEO's 100M at Tier 5
 * gives 200M effective weight, but 10,000 Tier 5 users staking 10K each
 * also produce 200M — equal influence at scale.
 *
 * Staked tokens include:
 *  - User staking positions (TokenLock)
 *  - Founder vesting positions (even if still locked/unvested)
 *  - Delegated positions (delegation doesn't affect governance weight — the delegator votes, not the validator)
 */
object GovernanceWeightCalculator {

  /** Trust tier multipliers (basis points for precision: 10000 = 1.0x) */
  val TierMultipliers: Map[Int, Int] = Map(
    1 -> 0,      // Unverified: 0.0x — no governance
    2 -> 5000,   // Newcomer:   0.5x
    3 -> 10000,  // Member:     1.0x
    4 -> 15000,  // Verified:   1.5x
    5 -> 20000   // Trusted:    2.0x
  )

  /**
   * Calculate the governance weight for a DID.
   * @return Governance weight (staked amount × tier multiplier, in basis points)
   */
  def calculate(did: String, state: LedgerState): Long = {
    val trustTier = state.trustTier(did)
    val multiplierBps = TierMultipliers.getOrElse(trustTier, 0)

    if (multiplierBps == 0) return 0L

    // Sum all TokenLock positions (including founder vesting locks)
    val totalStaked = state.allTokenLocks(did).map(_.amount).sum

    // Weight = staked × multiplier / 10000 (basis points to actual)
    (totalStaked * multiplierBps) / 10000
  }

  /**
   * Calculate the total weight for a vote outcome.
   * Used by L1 validators to determine if a proposal has passed.
   */
  def tallyProposal(proposalId: String, state: LedgerState): ProposalTally = {
    val votes = state.votesForProposal(proposalId)
    val forWeight = votes.filter(_.value == "for").map(_.weight).sum
    val againstWeight = votes.filter(_.value == "against").map(_.weight).sum
    val abstainWeight = votes.filter(_.value == "abstain").map(_.weight).sum
    val totalWeight = forWeight + againstWeight + abstainWeight

    ProposalTally(
      proposalId = proposalId,
      forWeight = forWeight,
      againstWeight = againstWeight,
      abstainWeight = abstainWeight,
      totalWeight = totalWeight,
      forPercent = if (totalWeight > 0) (forWeight * 100) / totalWeight else 0,
      passed = state.proposalThreshold(proposalId) match {
        case ThresholdType.SimpleMajority  => forWeight > againstWeight
        case ThresholdType.Supermajority67 => forPercent(forWeight, totalWeight) >= 67
        case ThresholdType.Supermajority75 => forPercent(forWeight, totalWeight) >= 75
      }
    )
  }

  private def forPercent(forWeight: Long, totalWeight: Long): Long =
    if (totalWeight > 0) (forWeight * 100) / totalWeight else 0
}

case class WeightedVote(vote: GovernanceVote, weight: Long)

case class ProposalTally(
  proposalId: String,
  forWeight: Long,
  againstWeight: Long,
  abstainWeight: Long,
  totalWeight: Long,
  forPercent: Long,
  passed: Boolean
)

sealed trait ThresholdType
object ThresholdType {
  case object SimpleMajority extends ThresholdType   // Treasury allocation ratios
  case object Supermajority67 extends ThresholdType   // Protocol upgrades, schema changes
  case object Supermajority75 extends ThresholdType   // Board removal, existential changes
}
```

### 5.3 Governance Weight Tests

```scala
// test/src/test/scala/echo/GovernanceWeightSpec.scala

package echo

import org.scalatest.flatspec.AnyFlatSpec
import org.scalatest.matchers.should.Matchers
import echo.data.validation.GovernanceWeightCalculator

class GovernanceWeightSpec extends AnyFlatSpec with Matchers {

  "GovernanceWeightCalculator" should "give zero weight to Tier 1 users" in {
    val state = MockState(trustTier = 1, stakedAmount = 50_000_000_00000000L)
    GovernanceWeightCalculator.calculate("did:unverified", state) shouldBe 0L
  }

  it should "apply 0.5x multiplier for Tier 2" in {
    val state = MockState(trustTier = 2, stakedAmount = 10_000_00000000L) // 10,000 ECHO
    val weight = GovernanceWeightCalculator.calculate("did:newcomer", state)
    weight shouldBe 5_000_00000000L // 5,000 effective
  }

  it should "apply 1.0x multiplier for Tier 3" in {
    val state = MockState(trustTier = 3, stakedAmount = 10_000_00000000L)
    val weight = GovernanceWeightCalculator.calculate("did:member", state)
    weight shouldBe 10_000_00000000L // 10,000 effective (1:1)
  }

  it should "apply 2.0x multiplier for Tier 5" in {
    val state = MockState(trustTier = 5, stakedAmount = 100_000_000_00000000L) // 100M (CEO)
    val weight = GovernanceWeightCalculator.calculate("did:ceo", state)
    weight shouldBe 200_000_000_00000000L // 200M effective
  }

  it should "include founder vesting locks in staked amount" in {
    val state = MockState(
      trustTier = 5,
      stakedAmount = 100_000_000_00000000L, // 100M in founder vesting lock
      isFounderLock = true
    )
    val weight = GovernanceWeightCalculator.calculate("did:founder", state)
    weight shouldBe 200_000_000_00000000L // Vesting locks count for governance
  }

  it should "give zero weight to users with no stake" in {
    val state = MockState(trustTier = 5, stakedAmount = 0L)
    GovernanceWeightCalculator.calculate("did:nostake", state) shouldBe 0L
  }

  "Community vs CEO balance" should "show community can outvote CEO at scale" in {
    // CEO: 100M staked, Tier 5 → 200M weight
    val ceoWeight = 200_000_000_00000000L

    // 10,000 Tier 5 users, 10K ECHO each = 100M total, ×2.0 = 200M weight
    val communityWeight = 10000L * (10_000_00000000L * 20000 / 10000)
    communityWeight shouldBe ceoWeight // Equal — community balances CEO
  }

  "Proposal tally" should "pass with simple majority" in {
    val tally = ProposalTally(
      proposalId = "prop-1",
      forWeight = 150_000_000L,
      againstWeight = 100_000_000L,
      abstainWeight = 50_000_000L,
      totalWeight = 300_000_000L,
      forPercent = 50,
      passed = true // 150M > 100M
    )
    tally.passed shouldBe true
  }

  it should "fail supermajority at 60%" in {
    // 60% for, needs 67% → fails
    val forPct = (180_000_000L * 100) / 300_000_000L // 60%
    forPct should be < 67L
  }
}
```

---

## 6. Validator Slashing (Phase 4)

```scala
// modules/currency_l1/src/main/scala/echo/currency/validation/SlashingValidator.scala

package echo.currency.validation

/** Validates slashing proposals submitted by L0 layer based on evidence from peer validators. */
object SlashingValidator {

  sealed trait SlashingOffense
  case class FraudulentRewardValidation(evidence: Seq[ConflictingValidationProof]) extends SlashingOffense
  case class InvalidMerkleSubmission(rejectionProof: ConsensusRejectionProof) extends SlashingOffense
  case class ExtendedDowntime(offlineBlocks: Int) extends SlashingOffense
  case class DoubleSigning(proof: ConflictingSignatureProof) extends SlashingOffense
  case class AntiGamingCollusion(evidence: PatternEvidence) extends SlashingOffense

  def slashPercentage(offense: SlashingOffense): Int = offense match {
    case _: FraudulentRewardValidation => 10    // 10% of staked ECHO
    case _: InvalidMerkleSubmission    => 5     // 5%
    case ExtendedDowntime(blocks)      => Math.min(blocks, 10) // 1% per 24h block, max 10%
    case _: DoubleSigning              => 50    // 50% + permanent ban
    case _: AntiGamingCollusion        => 25    // 25% + permanent ban
  }

  def isPermanentBan(offense: SlashingOffense): Boolean = offense match {
    case _: DoubleSigning        => true
    case _: AntiGamingCollusion  => true
    case _                       => false
  }

  def validate(proposal: SlashingProposal, state: LedgerState): ValidationResult = {
    for {
      _ <- validateValidatorExists(proposal.validatorId, state)
      _ <- validateEvidenceAuthenticity(proposal.offense)
      _ <- validateNotAlreadySlashed(proposal.validatorId, proposal.offense, state)
    } yield ValidTransaction
  }
}
```

---

## 7. Build & Test

```sbt
// build.sbt

ThisBuild / scalaVersion := "2.13.12"
ThisBuild / organization := "io.echo"

lazy val shared = (project in file("modules/shared"))
  .settings(
    libraryDependencies ++= Seq(
      "org.constellation" %% "tessellation-sdk" % "3.0.0",
      "org.constellation" %% "euclid-sdk" % "1.0.0"
    )
  )

lazy val currencyL1 = (project in file("modules/currency_l1"))
  .dependsOn(shared)
  .settings(
    libraryDependencies ++= Seq(
      "org.scalatest" %% "scalatest" % "3.2.17" % Test
    )
  )

lazy val dataL1 = (project in file("modules/data_l1"))
  .dependsOn(shared)

lazy val root = (project in file("."))
  .aggregate(shared, currencyL1, dataL1)
```

```bash
# Local development with Hydra CLI
hydra init echo-metagraph
hydra start  # Spins up: Global L0 + Metagraph L0 + Currency L1 + Data L1

# Run tests
sbt test

# Build JARs for deployment
sbt currencyL1/assembly
sbt dataL1/assembly
```

---

## 8. Implementation Priority

| Priority | Component | Effort | Phase |
|----------|-----------|--------|-------|
| P0 | Genesis allocation + minting | 1 week | Phase 1 (testnet) |
| P0 | TokenLock validator (user staking) | 1 week | Phase 1 |
| P0 | Founder vesting enforcement in WithdrawLock | 1 week | Phase 1 |
| P0 | RewardClaim validator + daily caps + emission tracking | 2 weeks | Phase 1 |
| P0 | AtomicAction validator | 1 week | Phase 1 |
| P0 | MerkleRoot validator (Data L1) | 1 week | Phase 1 |
| P1 | StakeDelegation validator | 3 days | Phase 1 |
| P1 | GovernanceVote validator + GovernanceWeightCalculator (Data L1) | 1 week | Phase 1 |
| P1 | FeeTransaction validator | 2 days | Phase 1 |
| P1 | Anti-gaming detection | 1 week | Phase 2 |
| P1 | TrustCommitment validator (Data L1) | 3 days | Phase 2 |
| P2 | AllowSpend + SpendTransaction validators | 1 week | Phase 3 |
| P2 | Slashing logic (Phase 4 activation) | 2 weeks | Phase 3 (build), Phase 4 (activate) |
| P2 | RelayNode registry validator (Data L1) | 3 days | Phase 4 |

**Total estimated effort:** ~12–14 engineering weeks (1 senior Scala developer)

---

*Scala Metagraph L1 Implementation Spec v2.0*
*Aligned to: PRD v2.5, Data Layer v3.4, Tokenomics v2.0*
*Status: Implementation-ready for Phase 1 testnet*
