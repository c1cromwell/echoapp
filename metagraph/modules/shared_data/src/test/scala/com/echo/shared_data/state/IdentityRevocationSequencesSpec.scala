package com.echo.shared_data.state

import org.scalatest.funspec.AnyFunSpec
import org.scalatest.matchers.should.Matchers

final class IdentityRevocationSequencesSpec extends AnyFunSpec with Matchers {

  describe("IdentityRevocationSequences") {
    it("tracks baselines per issuer org DID") {
      IdentityRevocationSequences.resetForTests()
      IdentityRevocationSequences.previousFor("did:key:zOrgA") shouldBe 0L
      IdentityRevocationSequences.recordPublished("did:key:zOrgA", 7L)
      IdentityRevocationSequences.previousFor("did:key:zOrgA") shouldBe 7L
      IdentityRevocationSequences.previousFor("did:key:zOrgB") shouldBe 0L
    }
  }
}
