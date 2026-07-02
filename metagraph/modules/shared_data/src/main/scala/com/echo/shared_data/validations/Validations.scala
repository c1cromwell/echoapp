package com.echo.shared_data.validations

import com.echo.shared_data.types._

/**
 * Shared validation logic used across L0 and L1 layers.
 * Mirrors constraints from Go backend internal/metagraph/transactions.go.
 *
 * T0–T7 invariants enforced here (WO-217):
 *   - Zero PII on any blockchain is a hard system invariant.
 *   - MerkleRootUpdate fields must be T5 hash commitments only (hex SHA-256).
 *   - TrustCommitmentUpdate must be T6 H(score|nonce) only (hex SHA-256).
 *   - No email addresses, phone numbers, or plain-text DID strings in data payloads.
 *   - No message content, ciphertext blobs, or user behavioral data beyond T5/T6/T7.
 */
object Validations {

  val SchemaVersion = "3.2.0"
  val SupportedVersions: Set[String] = Set("3.2.0", "3.1.0")

  // --- T0–T7 PII rejection patterns (WO-217) ---

  /** RFC 5322 simplified email pattern. */
  private val EmailPattern   = """[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}""".r

  /** E.164 / common phone patterns: +1234567890 or digits-only ≥ 10 chars. */
  private val PhonePattern   = """(?:\+?\d[\d\s\-().]{8,}\d)""".r

  /**
   * DID method prefixes that must never appear as plain-text data fields.
   * A DID is only valid as the authenticated *sender* field of the transaction
   * envelope — never embedded inside a payload (T0/T1 personal identifier).
   */
  private val DIDPrefixPattern = """did:[a-z0-9]+:""".r

  /**
   * Reject a string that contains PII: email, phone, or a plain DID embedded
   * in a data payload field.  Returns Left(reason) on violation.
   *
   * Used by all update validators that accept free-form string fields.
   */
  def rejectPII(fieldName: String, value: String): Either[String, Unit] = {
    if (EmailPattern.findFirstIn(value).isDefined)
      Left(s"T0-PII: $fieldName contains an email address — zero PII on-chain invariant violated")
    else if (PhonePattern.findFirstIn(value).isDefined)
      Left(s"T0-PII: $fieldName contains a phone number — zero PII on-chain invariant violated")
    else if (DIDPrefixPattern.findFirstIn(value).isDefined)
      Left(s"T0-PII: $fieldName contains a plain-text DID — embed DIDs only in the transaction sender envelope, not payload fields")
    else
      Right(())
  }

  /**
   * Reject a string that is not a valid 64-character lowercase hex SHA-256 hash.
   * T5 and T6 fields must be pure hash commitments — no content, no identifiers.
   */
  def requireHex64(fieldName: String, value: String): Either[String, Unit] =
    Either.cond(
      value.length == 64 && value.forall(c => c.isDigit || ('a' to 'f').contains(c)),
      (),
      s"T5/T6: $fieldName must be exactly 64 lowercase hex chars (SHA-256); got length ${value.length}"
    )

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

  /**
   * T6: TrustCommitmentUpdate.commitment must be a pure H(score|nonce) SHA-256 hex.
   * No PII, no content — only the 64-char hash is permitted.
   */
  def validateTrustCommitment(update: TrustCommitmentUpdate): Either[String, Unit] =
    for {
      _ <- requireHex64("commitment", update.commitment)
      _ <- Either.cond(update.epoch > 0, (), "Epoch must be positive")
    } yield ()

  /**
   * T5: MerkleRootUpdate.root must be a pure SHA-256 Merkle root — 64 lowercase hex chars.
   * No PII allowed in root or any associated string field (WO-217 zero-PII invariant).
   */
  def validateMerkleRoot(update: MerkleRootUpdate): Either[String, Unit] =
    for {
      _ <- requireHex64("root", update.root)
      _ <- Either.cond(update.leafCount > 0, (), "Leaf count must be positive")
    } yield ()

  def validateRewardClaim(update: RewardClaimUpdate): Either[String, Unit] =
    for {
      _ <- Either.cond(update.amount > 0, (), "Reward amount must be positive")
      _ <- StakingTiers.get(update.tier).toRight(s"Unknown tier: ${update.tier}")
    } yield ()

  /** WO-214: fixed 1B supply — reject all post-genesis minting. */
  def validateMint(update: MintUpdate): Either[String, Unit] =
    Left(s"Mint rejected: fixed supply — no minting after genesis (amount=${update.amount}, pool=${update.pool})")

  /** WO-225: 3-of-5 founder revocation must credit Future Team pool. */
  def validateFounderRevocation(update: FounderRevocationUpdate): Either[String, Unit] =
    for {
      _ <- Either.cond(update.amount > 0, (), "Revocation amount must be positive")
      _ <- Either.cond(update.revokerDids.distinct.size >= 3, (), "Revocation requires 3 distinct founder signatures")
      _ <- Either.cond(update.destinationPool == "future_team", (), "Revoked tokens must credit future_team pool")
    } yield ()

  /**
   * T7: TokenLockUpdate fields must contain no PII.
   * tierName is a controlled vocabulary — validate against known tiers.
   */
  def validateTokenLockPII(update: TokenLockUpdate): Either[String, Unit] =
    for {
      _ <- rejectPII("tierName", update.tierName)
    } yield ()

  /**
   * Reject any EchoUpdate that embeds PII in free-form string fields.
   * Called by the Data L1 dispatcher before structural validation.
   */
  def rejectUpdatePII(update: EchoUpdate): Either[String, Unit] = update match {
    case u: MerkleRootUpdate       => rejectPII("root", u.root)
    case u: TrustCommitmentUpdate  => rejectPII("commitment", u.commitment)
    case u: TokenLockUpdate        => validateTokenLockPII(u)
    case u: StakeDelegationUpdate  =>
      for {
        _ <- rejectPII("tokenLockTxId", u.tokenLockTxId)
        _ <- rejectPII("validatorDid",  u.validatorDid)
      } yield ()
    case u: RewardClaimUpdate      => rejectPII("tier", u.tier)
    case _: WithdrawLockUpdate     => Right(())
    case _: MintUpdate             => Right(())
    case u: FounderRevocationUpdate =>
      for {
        _ <- rejectPII("targetFounderDid", u.targetFounderDid)
        _ <- rejectPII("destinationPool", u.destinationPool)
      } yield ()
  }
}
