package com.echo.shared_data.validations

import com.echo.shared_data.types._
import org.scalatest.funspec.AnyFunSpec
import org.scalatest.matchers.should.Matchers

/**
 * Pure-function tests for the Identity Metagraph L1 validators.
 *
 * Mirrors the rules that the Identity L1 dispatcher invokes on every
 * incoming submission. Coverage target (per WO-272/WO-277): every
 * validator must have at least one happy-path and one rejection test
 * for each rule it enforces (authorized-sender, DID format, field
 * lengths, tier/role bounds, sequence monotonicity).
 */
final class IdentityValidationsSpec extends AnyFunSpec with Matchers {

  private val Now           = 1_700_000_000_000L
  private val IdentitySvc   = "did:key:z6MkjvBkt8ETnxXGBFPSGgYKb43q7oNHLX8BiYSPcXVG6gY6"
  private val Subject       = "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
  private val Issuer        = "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH"
  private val OrgIssuer     = "did:key:z6Mkm5K4XEQq8YjHK5xmM7VxN3aUuVeEEgkwZxBgcs2yK4Ra"
  private val Member        = "did:key:z6MkrJVnaZkeFzdQOGNvL7TbMaMt4G2zHFp2hPBn9Az3vS5K"
  private val UnauthSender  = "did:key:zUNAUTHORIZED_SENDER_FAKE_KEY"

  // helpers ---------------------------------------------------------------

  private def validVC = VCIssuanceUpdate(
    credentialId   = "urn:uuid:11111111-1111-1111-1111-111111111111",
    subjectDID     = Subject,
    issuerDID      = Issuer,
    credentialType = "TrustTierCredential",
    issuedAt       = Now - 1000L,
    schemaVersion  = "v1.0.0"
  )

  private def validTrustTier = TrustTierCommitmentUpdate(
    subjectDID = Subject,
    commitment = "a" * 64, // 32 bytes hex
    anchoredAt = Now - 500L
  )

  private def validStatusList = StatusList2021BatchUpdate(
    issuerOrgDID = OrgIssuer,
    bitVector    = "0" * StatusList2021Vector.ExpectedHexLength,
    publishedAt  = Now - 100L,
    sequence     = 1L
  )

  private def validOrgRole = EchoOrgRoleCredentialUpdate(
    credentialId = "urn:uuid:22222222-2222-2222-2222-222222222222",
    issuerOrgDID = OrgIssuer,
    memberDID    = Member,
    role         = "admin",
    expiry       = Now + 86_400_000L,
    issuedAt     = Now - 1000L
  )

  // -- VC issuance --------------------------------------------------------

  describe("validateVCIssuance") {
    it("accepts a well-formed update from the authorized sender") {
      IdentityValidations.validateVCIssuance(validVC, IdentitySvc, IdentitySvc, Now) shouldBe Right(())
    }

    it("rejects submissions from an unauthorized sender") {
      val r = IdentityValidations.validateVCIssuance(validVC, UnauthSender, IdentitySvc, Now)
      r.isLeft shouldBe true
      r.left.toOption.get should include("not the authorized")
    }

    it("rejects a non-did:key subject") {
      val u = validVC.copy(subjectDID = "did:web:example.com")
      IdentityValidations.validateVCIssuance(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects a non-did:key issuer") {
      val u = validVC.copy(issuerDID = "did:prism:cardano:abc")
      IdentityValidations.validateVCIssuance(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects empty credentialId") {
      val u = validVC.copy(credentialId = "")
      IdentityValidations.validateVCIssuance(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects an issuedAt that is too far in the future") {
      val u = validVC.copy(issuedAt = Now + 600_000L)
      IdentityValidations.validateVCIssuance(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects non-positive issuedAt") {
      val u = validVC.copy(issuedAt = 0L)
      IdentityValidations.validateVCIssuance(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects empty schemaVersion") {
      val u = validVC.copy(schemaVersion = "")
      IdentityValidations.validateVCIssuance(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects empty credentialType") {
      val u = validVC.copy(credentialType = "")
      IdentityValidations.validateVCIssuance(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }
  }

  // -- Trust tier commitment ----------------------------------------------

  describe("validateTrustTierCommitment") {
    it("accepts a well-formed 32-byte hex commitment from the authorized sender") {
      IdentityValidations.validateTrustTierCommitment(
        validTrustTier, IdentitySvc, IdentitySvc, Now
      ) shouldBe Right(())
    }

    it("rejects submissions from an unauthorized sender") {
      IdentityValidations.validateTrustTierCommitment(
        validTrustTier, UnauthSender, IdentitySvc, Now
      ).isLeft shouldBe true
    }

    it("rejects commitments that are not 64 hex chars") {
      val short = validTrustTier.copy(commitment = "a" * 63)
      IdentityValidations.validateTrustTierCommitment(short, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
      val long  = validTrustTier.copy(commitment = "a" * 65)
      IdentityValidations.validateTrustTierCommitment(long, IdentitySvc, IdentitySvc, Now).isLeft  shouldBe true
    }

    it("rejects uppercase hex (commitment must be lowercase)") {
      val u = validTrustTier.copy(commitment = "A" * 64)
      IdentityValidations.validateTrustTierCommitment(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects non-hex characters") {
      val u = validTrustTier.copy(commitment = "z" * 64)
      IdentityValidations.validateTrustTierCommitment(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects non-did:key subject") {
      val u = validTrustTier.copy(subjectDID = "did:web:example.com")
      IdentityValidations.validateTrustTierCommitment(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects non-positive anchoredAt") {
      val u = validTrustTier.copy(anchoredAt = 0L)
      IdentityValidations.validateTrustTierCommitment(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects anchoredAt too far in the future") {
      val u = validTrustTier.copy(anchoredAt = Now + 600_000L)
      IdentityValidations.validateTrustTierCommitment(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }
  }

  // -- StatusList2021 -----------------------------------------------------

  describe("validateStatusList2021") {
    it("accepts a 131,072-bit (32,768 hex chars) vector with monotonically increasing sequence") {
      IdentityValidations.validateStatusList2021(
        validStatusList, IdentitySvc, IdentitySvc, previousSequence = 0L, Now
      ) shouldBe Right(())
    }

    it("verifies the constant ExpectedHexLength is the W3C-mandated 32,768") {
      StatusList2021Vector.ExpectedHexLength shouldBe 32768
      StatusList2021Vector.ExpectedBitLength shouldBe 131072
    }

    it("rejects a vector that is not exactly 32,768 hex chars") {
      val short = validStatusList.copy(bitVector = "0" * (StatusList2021Vector.ExpectedHexLength - 1))
      IdentityValidations.validateStatusList2021(short, IdentitySvc, IdentitySvc, 0L, Now).isLeft shouldBe true
      val long  = validStatusList.copy(bitVector = "0" * (StatusList2021Vector.ExpectedHexLength + 1))
      IdentityValidations.validateStatusList2021(long,  IdentitySvc, IdentitySvc, 0L, Now).isLeft shouldBe true
    }

    it("rejects a vector that is not hex-encoded") {
      val u = validStatusList.copy(bitVector = "z" * StatusList2021Vector.ExpectedHexLength)
      IdentityValidations.validateStatusList2021(u, IdentitySvc, IdentitySvc, 0L, Now).isLeft shouldBe true
    }

    it("rejects a sequence equal to the previous sequence (monotonic)") {
      val r = IdentityValidations.validateStatusList2021(validStatusList, IdentitySvc, IdentitySvc, 1L, Now)
      r.isLeft shouldBe true
      r.left.toOption.get should include("must be greater than previous")
    }

    it("rejects a sequence less than the previous sequence (replay/regression)") {
      val u = validStatusList.copy(sequence = 1L)
      IdentityValidations.validateStatusList2021(u, IdentitySvc, IdentitySvc, 5L, Now).isLeft shouldBe true
    }

    it("rejects unauthorized senders") {
      IdentityValidations.validateStatusList2021(
        validStatusList, UnauthSender, IdentitySvc, 0L, Now
      ).isLeft shouldBe true
    }

    it("rejects non-did:key issuerOrgDID") {
      val u = validStatusList.copy(issuerOrgDID = "did:web:example.com")
      IdentityValidations.validateStatusList2021(u, IdentitySvc, IdentitySvc, 0L, Now).isLeft shouldBe true
    }
  }

  // -- EchoOrgRoleCredential ---------------------------------------------

  describe("validateEchoOrgRoleCredential") {
    it("accepts a well-formed org membership credential from the authorized sender") {
      IdentityValidations.validateEchoOrgRoleCredential(
        validOrgRole, IdentitySvc, IdentitySvc, Now
      ) shouldBe Right(())
    }

    it("rejects unauthorized senders") {
      IdentityValidations.validateEchoOrgRoleCredential(
        validOrgRole, UnauthSender, IdentitySvc, Now
      ).isLeft shouldBe true
    }

    it("accepts every recognized role: owner, admin, moderator, member") {
      OrgRole.All.foreach { role =>
        val u = validOrgRole.copy(role = role)
        IdentityValidations.validateEchoOrgRoleCredential(u, IdentitySvc, IdentitySvc, Now) shouldBe Right(())
      }
    }

    it("rejects an unknown role to prevent privilege escalation via typos") {
      val u = validOrgRole.copy(role = "superadmin")
      val r = IdentityValidations.validateEchoOrgRoleCredential(u, IdentitySvc, IdentitySvc, Now)
      r.isLeft shouldBe true
      r.left.toOption.get should include("not a recognized org role")
    }

    it("rejects non-did:key issuerOrgDID and memberDID") {
      val u1 = validOrgRole.copy(issuerOrgDID = "did:web:example.com")
      IdentityValidations.validateEchoOrgRoleCredential(u1, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
      val u2 = validOrgRole.copy(memberDID = "did:web:example.com")
      IdentityValidations.validateEchoOrgRoleCredential(u2, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects an expiry that is in the past") {
      val u = validOrgRole.copy(expiry = Now - 1L, issuedAt = Now - 1000L)
      IdentityValidations.validateEchoOrgRoleCredential(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects expiry <= issuedAt") {
      val u = validOrgRole.copy(expiry = validOrgRole.issuedAt)
      IdentityValidations.validateEchoOrgRoleCredential(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects empty credentialId") {
      val u = validOrgRole.copy(credentialId = "")
      IdentityValidations.validateEchoOrgRoleCredential(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects empty role") {
      val u = validOrgRole.copy(role = "")
      IdentityValidations.validateEchoOrgRoleCredential(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }
  }

  // -- Device key registration -------------------------------------------

  private def validDeviceKey = DeviceKeyRegistrationUpdate(
    subjectDID   = Subject,
    publicKeyHex = "04" + ("a" * 128),
    deviceLabel  = "ipad",
    addedAt      = Now - 100L
  )

  describe("validateDeviceKeyRegistration") {
    it("accepts a well-formed update from the authorized sender") {
      IdentityValidations.validateDeviceKeyRegistration(
        validDeviceKey, IdentitySvc, IdentitySvc, Now
      ) shouldBe Right(())
    }

    it("rejects unauthorized senders") {
      IdentityValidations.validateDeviceKeyRegistration(
        validDeviceKey, UnauthSender, IdentitySvc, Now
      ).isLeft shouldBe true
    }

    it("rejects a non-did:key subject") {
      val u = validDeviceKey.copy(subjectDID = "did:web:example.com")
      IdentityValidations.validateDeviceKeyRegistration(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects wrong public key hex length") {
      val u = validDeviceKey.copy(publicKeyHex = "04abcd")
      IdentityValidations.validateDeviceKeyRegistration(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects uppercase hex in public key") {
      val u = validDeviceKey.copy(publicKeyHex = "04" + ("A" * 128))
      IdentityValidations.validateDeviceKeyRegistration(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }

    it("rejects empty device label") {
      val u = validDeviceKey.copy(deviceLabel = "")
      IdentityValidations.validateDeviceKeyRegistration(u, IdentitySvc, IdentitySvc, Now).isLeft shouldBe true
    }
  }

  // -- TrustTier sanity ---------------------------------------------------

  describe("TrustTier") {
    it("declares the 5 allowed tiers (Phase 1)") {
      TrustTier.All shouldBe Set(1, 2, 3, 4, 5)
      TrustTier.Min shouldBe 1
      TrustTier.Max shouldBe 5
    }
  }
}
