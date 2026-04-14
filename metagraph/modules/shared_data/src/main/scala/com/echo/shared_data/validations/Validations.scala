package com.echo.shared_data.validations

import com.echo.shared_data.types._

/**
 * Shared validation logic used across L0 and L1 layers.
 * Mirrors constraints from Go backend internal/metagraph/transactions.go.
 */
object Validations {

  val SchemaVersion = "3.2.0"
  val SupportedVersions: Set[String] = Set("3.2.0", "3.1.0")

  /** Default staking tiers (amounts in datum: 1 ECHO = 1e8) */
  val StakingTiers: Map[String, (Long, Int)] = Map(
    "Tier 1" -> (10000000000L, 30),      // 100 ECHO, 30 days
    "Tier 2" -> (100000000000L, 90),     // 1,000 ECHO, 90 days
    "Tier 3" -> (1000000000000L, 180),   // 10,000 ECHO, 180 days
    "Tier 4" -> (10000000000000L, 270),  // 100,000 ECHO, 270 days
    "Tier 5" -> (100000000000000L, 365)  // 1,000,000 ECHO, 365 days
  )

  def validateTokenLock(update: TokenLockUpdate): Either[String, Unit] =
    for {
      tier <- StakingTiers.get(update.tierName).toRight(s"Unknown tier: ${update.tierName}")
      (minStake, expectedDays) = tier
      _ <- Either.cond(update.amount >= minStake, (), s"Amount ${update.amount} below minimum ${minStake} for ${update.tierName}")
      _ <- Either.cond(update.lockDays >= expectedDays, (), s"Lock duration ${update.lockDays} below minimum ${expectedDays} days for ${update.tierName}")
    } yield ()

  def validateTrustCommitment(update: TrustCommitmentUpdate): Either[String, Unit] =
    for {
      _ <- Either.cond(update.commitment.length == 64, (), s"Commitment must be 64 hex chars (SHA-256), got ${update.commitment.length}")
      _ <- Either.cond(update.commitment.forall(c => c.isDigit || ('a' to 'f').contains(c.toLower)), (), "Commitment must be hex-encoded")
      _ <- Either.cond(update.epoch > 0, (), "Epoch must be positive")
    } yield ()

  def validateMerkleRoot(update: MerkleRootUpdate): Either[String, Unit] =
    for {
      _ <- Either.cond(update.root.length == 64, (), s"Root must be 64 hex chars (SHA-256), got ${update.root.length}")
      _ <- Either.cond(update.root.forall(c => c.isDigit || ('a' to 'f').contains(c.toLower)), (), "Root must be hex-encoded")
      _ <- Either.cond(update.leafCount > 0, (), "Leaf count must be positive")
    } yield ()

  def validateRewardClaim(update: RewardClaimUpdate): Either[String, Unit] =
    for {
      _ <- Either.cond(update.amount > 0, (), "Reward amount must be positive")
      _ <- StakingTiers.get(update.tier).toRight(s"Unknown tier: ${update.tier}")
    } yield ()
}
