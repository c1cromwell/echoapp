package com.echo.shared_data.types

import io.circe.parser.decode
import io.circe.syntax._
import org.scalatest.funspec.AnyFunSpec
import org.scalatest.matchers.should.Matchers

final class IdentityUpdateCodecSpec extends AnyFunSpec with Matchers {

  private val subject = "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
  private val issuer  = "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH"
  private val member  = "did:key:z6MkrJVnaZkeFzdQOGNvL7TbMaMt4G2zHFp2hPBn9Az3vS5K"
  private val pubHex  = "04" + ("c" * 128)

  describe("IdentityUpdate.decoder") {
    it("decodes DeviceKeyRegistrationUpdate JSON") {
      val json =
        s"""{"subjectDID":"$subject","publicKeyHex":"$pubHex","deviceLabel":"watch","addedAt":1700000000123}"""
      decode[IdentityUpdate](json) shouldBe Right(
        DeviceKeyRegistrationUpdate(subject, pubHex, "watch", 1700000000123L)
      )
    }

    it("decodes VCIssuanceUpdate when schemaVersion is present") {
      val json =
        """{"credentialId":"urn:x","subjectDID":"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK","issuerDID":"did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH","credentialType":"TrustTierCredential","issuedAt":1,"schemaVersion":"v1"}"""
      decode[IdentityUpdate](json).isRight shouldBe true
    }

    it("round-trips VCIssuanceUpdate via encoder + decoder (WO-272 / WO-277)") {
      val u: IdentityUpdate = VCIssuanceUpdate(
        credentialId   = "urn:uuid:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
        subjectDID     = subject,
        issuerDID      = issuer,
        credentialType = "TrustTierCredential",
        issuedAt       = 1700000000999L,
        schemaVersion  = "v1.0.0"
      )
      val json = u.asJson.noSpaces
      decode[IdentityUpdate](json) shouldBe Right(u)
    }

    it("round-trips TrustTierCommitmentUpdate via encoder + decoder") {
      val u: IdentityUpdate = TrustTierCommitmentUpdate(
        subjectDID = subject,
        commitment = "a" * 64,
        anchoredAt = 1700000000123L
      )
      val json = u.asJson.noSpaces
      decode[IdentityUpdate](json) shouldBe Right(u)
    }

    it("round-trips StatusList2021BatchUpdate via encoder + decoder") {
      val u: IdentityUpdate = StatusList2021BatchUpdate(
        issuerOrgDID = issuer,
        bitVector    = "0" * StatusList2021Vector.ExpectedHexLength,
        publishedAt  = 1700000000456L,
        sequence     = 9L
      )
      val json = u.asJson.noSpaces
      decode[IdentityUpdate](json) shouldBe Right(u)
    }

    it("round-trips EchoOrgRoleCredentialUpdate via encoder + decoder") {
      val u: IdentityUpdate = EchoOrgRoleCredentialUpdate(
        credentialId = "urn:uuid:role-codec-1",
        issuerOrgDID = issuer,
        memberDID    = member,
        role         = "moderator",
        expiry       = 1900000000000L,
        issuedAt     = 1700000000000L
      )
      val json = u.asJson.noSpaces
      decode[IdentityUpdate](json) shouldBe Right(u)
    }
  }
}
