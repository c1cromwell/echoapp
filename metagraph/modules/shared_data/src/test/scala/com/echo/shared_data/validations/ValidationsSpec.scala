package com.echo.shared_data.validations

import com.echo.shared_data.types._
import org.scalatest.funspec.AnyFunSpec
import org.scalatest.matchers.should.Matchers

/**
 * Pure-function tests for the Currency / Data L1 validators.
 *
 * These tests do NOT touch the Tessellation SDK — they verify the rules
 * exactly as they will run inside the L1 dispatcher. Coverage target
 * (per WO-277): ≥90% line coverage of `Validations.scala`.
 */
final class ValidationsSpec extends AnyFunSpec with Matchers {

  // -- validateTokenLock ---------------------------------------------------

  describe("Validations.validateTokenLock") {

    val validTier1 = TokenLockUpdate(amount = 10000000000L,        tierName = "Tier 1", lockDays = 30)
    val validTier2 = TokenLockUpdate(amount = 100000000000L,       tierName = "Tier 2", lockDays = 90)
    val validTier3 = TokenLockUpdate(amount = 1000000000000L,      tierName = "Tier 3", lockDays = 180)
    val validTier4 = TokenLockUpdate(amount = 10000000000000L,     tierName = "Tier 4", lockDays = 270)
    val validTier5 = TokenLockUpdate(amount = 100000000000000L,    tierName = "Tier 5", lockDays = 365)

    it("accepts the minimum amount + duration for each of the 5 tiers") {
      Seq(validTier1, validTier2, validTier3, validTier4, validTier5).foreach { u =>
        Validations.validateTokenLock(u) shouldBe Right(())
      }
    }

    it("accepts amounts above the tier minimum") {
      val above = validTier1.copy(amount = validTier1.amount * 2)
      Validations.validateTokenLock(above) shouldBe Right(())
    }

    it("accepts lock durations longer than the tier minimum") {
      val longer = validTier3.copy(lockDays = 365)
      Validations.validateTokenLock(longer) shouldBe Right(())
    }

    it("rejects amounts below the tier minimum") {
      val low = validTier2.copy(amount = validTier2.amount - 1L)
      val r   = Validations.validateTokenLock(low)
      r.isLeft shouldBe true
      r.left.toOption.get should include("below minimum")
    }

    it("rejects lock durations below the tier minimum") {
      val short = validTier4.copy(lockDays = 30)
      val r     = Validations.validateTokenLock(short)
      r.isLeft shouldBe true
      r.left.toOption.get should include("below minimum")
    }

    it("rejects unknown tier names") {
      val unknown = TokenLockUpdate(amount = 1L, tierName = "Tier 99", lockDays = 1)
      val r       = Validations.validateTokenLock(unknown)
      r.isLeft shouldBe true
      r.left.toOption.get should include("Unknown tier")
    }

    it("rejects unknown tier names that look like typos") {
      val typo = TokenLockUpdate(amount = 10000000000L, tierName = "tier 1", lockDays = 30)
      Validations.validateTokenLock(typo).isLeft shouldBe true
    }
  }

  // -- validateTrustCommitment --------------------------------------------

  describe("Validations.validateTrustCommitment") {
    val valid = TrustCommitmentUpdate(
      commitment = "a" * 64,
      epoch      = 42L
    )

    it("accepts a 64-char lowercase hex commitment with positive epoch") {
      Validations.validateTrustCommitment(valid) shouldBe Right(())
    }

    it("rejects uppercase hex — submission must be normalised to lowercase before sending (WO-217)") {
      Validations.validateTrustCommitment(valid.copy(commitment = "A" * 64)).isLeft shouldBe true
    }

    it("rejects commitments shorter than 64 chars") {
      Validations.validateTrustCommitment(valid.copy(commitment = "a" * 63)).isLeft shouldBe true
    }

    it("rejects commitments longer than 64 chars") {
      Validations.validateTrustCommitment(valid.copy(commitment = "a" * 65)).isLeft shouldBe true
    }

    it("rejects non-hex characters") {
      Validations.validateTrustCommitment(valid.copy(commitment = "g" * 64)).isLeft shouldBe true
    }

    it("rejects non-positive epochs") {
      Validations.validateTrustCommitment(valid.copy(epoch = 0L)).isLeft   shouldBe true
      Validations.validateTrustCommitment(valid.copy(epoch = -1L)).isLeft  shouldBe true
    }
  }

  // -- validateMerkleRoot --------------------------------------------------

  describe("Validations.validateMerkleRoot") {
    val valid = MerkleRootUpdate(
      root      = "f" * 64,
      leafCount = 1
    )

    it("accepts a 64-char hex root with positive leaf count") {
      Validations.validateMerkleRoot(valid) shouldBe Right(())
    }

    it("rejects roots whose length is not 64") {
      Validations.validateMerkleRoot(valid.copy(root = "f" * 63)).isLeft shouldBe true
      Validations.validateMerkleRoot(valid.copy(root = "f" * 65)).isLeft shouldBe true
    }

    it("rejects non-hex roots") {
      Validations.validateMerkleRoot(valid.copy(root = "z" * 64)).isLeft shouldBe true
    }

    it("rejects zero or negative leaf counts") {
      Validations.validateMerkleRoot(valid.copy(leafCount = 0)).isLeft  shouldBe true
      Validations.validateMerkleRoot(valid.copy(leafCount = -1)).isLeft shouldBe true
    }
  }

  // -- validateRewardClaim -------------------------------------------------

  describe("Validations.validateRewardClaim") {
    it("accepts positive amount + known tier") {
      val u = RewardClaimUpdate(amount = 1L, tier = "Tier 1")
      Validations.validateRewardClaim(u) shouldBe Right(())
    }

    it("rejects non-positive amounts") {
      val u = RewardClaimUpdate(amount = 0L, tier = "Tier 2")
      Validations.validateRewardClaim(u).isLeft shouldBe true
    }

    it("rejects unknown tier names") {
      val u = RewardClaimUpdate(amount = 1L, tier = "Tier 0")
      Validations.validateRewardClaim(u).isLeft shouldBe true
    }
  }

  // -- StakingTiers table --------------------------------------------------

  describe("Validations.StakingTiers") {
    it("contains exactly the 5 production tiers") {
      Validations.StakingTiers.keySet shouldBe Set("Tier 1", "Tier 2", "Tier 3", "Tier 4", "Tier 5")
    }

    it("scales amounts and durations monotonically across tiers") {
      val ordered = Seq("Tier 1", "Tier 2", "Tier 3", "Tier 4", "Tier 5")
        .map(Validations.StakingTiers)
      ordered.map(_._1).sliding(2).foreach { case Seq(a, b) => a should be < b; case _ => () }
      ordered.map(_._2).sliding(2).foreach { case Seq(a, b) => a should be < b; case _ => () }
    }
  }

  // -- WO-217: T0–T7 PII rejection (rejectPII) --------------------------------

  describe("Validations.rejectPII (WO-217)") {

    it("accepts clean hex-only fields") {
      Validations.rejectPII("root", "a" * 64) shouldBe Right(())
    }

    it("rejects email addresses") {
      val result = Validations.rejectPII("data", "user@example.com")
      result shouldBe a [Left[_, _]]
      result.swap.getOrElse("") should include ("email")
    }

    it("rejects phone numbers — E.164 format") {
      val result = Validations.rejectPII("data", "+14155551234")
      result shouldBe a [Left[_, _]]
      result.swap.getOrElse("") should include ("phone")
    }

    it("rejects embedded DID strings") {
      val result = Validations.rejectPII("data", "did:key:zDnmrTest123")
      result shouldBe a [Left[_, _]]
      result.swap.getOrElse("") should include ("DID")
    }

    it("rejects did:prism strings") {
      val result = Validations.rejectPII("payload", "did:prism:abc123")
      result shouldBe a [Left[_, _]]
    }

    it("accepts normal alphanumeric free-form values") {
      Validations.rejectPII("tier", "Tier 3") shouldBe Right(())
    }
  }

  // -- WO-217: requireHex64 --------------------------------------------------

  describe("Validations.requireHex64 (WO-217)") {

    it("accepts a valid 64-char lowercase hex SHA-256") {
      val hash = "0" * 32 + "a" * 32
      Validations.requireHex64("root", hash) shouldBe Right(())
    }

    it("rejects a value that is too short") {
      Validations.requireHex64("root", "abc123") shouldBe a [Left[_, _]]
    }

    it("rejects a value with uppercase hex chars (normalisation required before submission)") {
      Validations.requireHex64("root", "A" * 64) shouldBe a [Left[_, _]]
    }

    it("rejects a value with non-hex characters") {
      Validations.requireHex64("root", "g" * 64) shouldBe a [Left[_, _]]
    }
  }

  // -- WO-217: rejectUpdatePII -----------------------------------------------

  describe("Validations.rejectUpdatePII (WO-217)") {

    it("rejects MerkleRootUpdate whose root embeds PII") {
      val bad = MerkleRootUpdate(root = "did:key:zBadPII", leafCount = 1)
      Validations.rejectUpdatePII(bad) shouldBe a [Left[_, _]]
    }

    it("accepts MerkleRootUpdate with clean hex root") {
      val good = MerkleRootUpdate(root = "b" * 64, leafCount = 5)
      Validations.rejectUpdatePII(good) shouldBe Right(())
    }

    it("rejects TrustCommitmentUpdate whose commitment embeds PII") {
      val bad = TrustCommitmentUpdate(commitment = "user@example.com", epoch = 1)
      Validations.rejectUpdatePII(bad) shouldBe a [Left[_, _]]
    }

    it("accepts TrustCommitmentUpdate with clean hex commitment") {
      val good = TrustCommitmentUpdate(commitment = "f" * 64, epoch = 42)
      Validations.rejectUpdatePII(good) shouldBe Right(())
    }

    it("rejects StakeDelegationUpdate with DID embedded in validatorDid field as data") {
      // validatorDid is expected to be an opaque identifier, not a PII DID in a data payload
      val bad = StakeDelegationUpdate(tokenLockTxId = "tx-001", validatorDid = "did:key:zBad", amount = 1000)
      Validations.rejectUpdatePII(bad) shouldBe a [Left[_, _]]
    }

    it("passes WithdrawLockUpdate unconditionally (no string fields to check)") {
      val ok = WithdrawLockUpdate(tokenLockTxId = "tx-001", amount = 500)
      Validations.rejectUpdatePII(ok) shouldBe Right(())
    }
  }
}
