package com.echo.shared_data.types

import io.circe.parser.decode
import org.scalatest.funspec.AnyFunSpec
import org.scalatest.matchers.should.Matchers

final class IdentityUpdateCodecSpec extends AnyFunSpec with Matchers {

  private val subject = "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
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
  }
}
