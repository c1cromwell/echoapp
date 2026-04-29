package com.echo.l1

import com.echo.shared_data.types._
import org.scalatest.funspec.AnyFunSpec
import org.scalatest.matchers.should.Matchers

/** Wired-validator integration spec for the Currency L1 application. */
final class MainSpec extends AnyFunSpec with Matchers {

  describe("Currency L1 Main.dispatch") {

    it("accepts a TokenLock at the Tier 1 minimum") {
      val u = TokenLockUpdate(amount = 10000000000L, tierName = "Tier 1", lockDays = 30)
      Main.dispatch(u) shouldBe Right(())
    }

    it("rejects a TokenLock below the Tier 1 minimum (wired rejection)") {
      val u = TokenLockUpdate(amount = 1L, tierName = "Tier 1", lockDays = 30)
      Main.dispatch(u).isLeft shouldBe true
    }

    it("accepts a positive RewardClaim with a known tier") {
      val u = RewardClaimUpdate(amount = 100L, tier = "Tier 3")
      Main.dispatch(u) shouldBe Right(())
    }

    it("rejects a RewardClaim with an unknown tier (wired rejection)") {
      val u = RewardClaimUpdate(amount = 100L, tier = "Tier 999")
      Main.dispatch(u).isLeft shouldBe true
    }

    it("passes through StakeDelegationUpdate / WithdrawLockUpdate (no per-update invariants yet)") {
      Main.dispatch(StakeDelegationUpdate("tx", "did:key:zABC", 1L)) shouldBe Right(())
      Main.dispatch(WithdrawLockUpdate("tx", 1L))                    shouldBe Right(())
    }

    it("rejects update types that don't belong on Currency L1 (MerkleRoot)") {
      val u = MerkleRootUpdate(root = "a" * 64, leafCount = 1)
      Main.dispatch(u).isLeft shouldBe true
    }
  }
}
