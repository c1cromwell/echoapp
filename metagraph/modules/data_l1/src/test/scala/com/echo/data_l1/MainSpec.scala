package com.echo.data_l1

import com.echo.shared_data.types._
import org.scalatest.funspec.AnyFunSpec
import org.scalatest.matchers.should.Matchers

/**
 * Wired-validator integration spec for the Data L1 application.
 *
 * WO-277 acceptance criterion #3: "At least one spec demonstrates a
 * wired validator rejecting an invalid update at the L1 application
 * layer (not just calling Validations directly)."
 *
 * `Main.dispatch` is the same function the Tessellation
 * `BaseDataApplicationL1Service` callback delegates to, so this proves
 * the wiring — not just the rule library.
 */
final class MainSpec extends AnyFunSpec with Matchers {

  describe("Data L1 Main.dispatch") {

    it("accepts a well-formed MerkleRoot update") {
      val u = MerkleRootUpdate(root = "a" * 64, leafCount = 1)
      Main.dispatch(u) shouldBe Right(())
    }

    it("rejects a MerkleRoot update whose root is not 64 hex chars (wired rejection)") {
      val u = MerkleRootUpdate(root = "a" * 10, leafCount = 1)
      Main.dispatch(u).isLeft shouldBe true
    }

    it("accepts a well-formed TrustCommitment update") {
      val u = TrustCommitmentUpdate(commitment = "f" * 64, epoch = 1L)
      Main.dispatch(u) shouldBe Right(())
    }

    it("rejects a TrustCommitment update with a non-positive epoch (wired rejection)") {
      val u = TrustCommitmentUpdate(commitment = "f" * 64, epoch = 0L)
      Main.dispatch(u).isLeft shouldBe true
    }

    it("rejects update types that don't belong on Data L1 (TokenLock)") {
      val u = TokenLockUpdate(amount = 1L, tierName = "Tier 1", lockDays = 30)
      val r = Main.dispatch(u)
      r.isLeft shouldBe true
      r.left.toOption.get should include("Data L1 does not accept")
    }
  }
}
