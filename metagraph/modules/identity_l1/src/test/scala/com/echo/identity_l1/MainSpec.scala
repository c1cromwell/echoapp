package com.echo.identity_l1

import com.echo.shared_data.types._
import org.scalatest.funspec.AnyFunSpec
import org.scalatest.matchers.should.Matchers

/**
 * Wired-validator integration spec for the Identity L1 application.
 *
 * Demonstrates that the `dispatch` function the Tessellation L1 mempool
 * delegates to refuses malformed submissions exactly the same way the
 * pure validators do (WO-277 acceptance criterion #3).
 */
final class MainSpec extends AnyFunSpec with Matchers {

  // We can't easily mutate Main.authorizedSenderDid (it's read once from
  // env at JVM start). Instead we use whatever it resolved to and feed
  // the same value as `sender` so authorization passes — keeping these
  // tests focused on the wiring + rule application rather than env juggling.
  private val Sender = Main.authorizedSenderDid
  private val Now    = 1_700_000_000_000L

  private val Subject   = "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
  private val Issuer    = "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH"
  private val OrgIssuer = "did:key:z6Mkm5K4XEQq8YjHK5xmM7VxN3aUuVeEEgkwZxBgcs2yK4Ra"
  private val Member    = "did:key:z6MkrJVnaZkeFzdQOGNvL7TbMaMt4G2zHFp2hPBn9Az3vS5K"

  describe("Identity L1 Main.dispatch") {

    it("constructs without throwing (cluster UUID parses, env defaults applied)") {
      noException should be thrownBy Main.authorizedSenderDid
    }

    it("accepts a well-formed VC issuance update") {
      val u = VCIssuanceUpdate(
        credentialId   = "urn:uuid:11111111-1111-1111-1111-111111111111",
        subjectDID     = Subject,
        issuerDID      = Issuer,
        credentialType = "TrustTierCredential",
        issuedAt       = Now - 1000L,
        schemaVersion  = "v1.0.0"
      )
      Main.dispatch(u, Sender, Now) shouldBe Right(())
    }

    it("rejects a VC issuance whose subject is not did:key (wired rejection)") {
      val u = VCIssuanceUpdate(
        credentialId   = "urn:uuid:bad",
        subjectDID     = "did:web:example.com",
        issuerDID      = Issuer,
        credentialType = "TrustTierCredential",
        issuedAt       = Now - 1000L,
        schemaVersion  = "v1.0.0"
      )
      Main.dispatch(u, Sender, Now).isLeft shouldBe true
    }

    it("accepts a 32-byte hex trust tier commitment") {
      val u = TrustTierCommitmentUpdate(Subject, "a" * 64, Now - 100L)
      Main.dispatch(u, Sender, Now) shouldBe Right(())
    }

    it("rejects a trust tier commitment that isn't 32 bytes hex (wired rejection)") {
      val u = TrustTierCommitmentUpdate(Subject, "a" * 10, Now - 100L)
      Main.dispatch(u, Sender, Now).isLeft shouldBe true
    }

    it("accepts a StatusList2021 batch with a fresh sequence and rejects a stale one") {
      val orgDid = OrgIssuer
      val good   = StatusList2021BatchUpdate(
        issuerOrgDID = orgDid,
        bitVector    = "0" * StatusList2021Vector.ExpectedHexLength,
        publishedAt  = Now - 100L,
        sequence     = 1L
      )
      Main.dispatch(good, Sender, Now) shouldBe Right(())
      Main.recordPublishedSequence(orgDid, 1L)

      val stale = good.copy(sequence = 1L)
      val r     = Main.dispatch(stale, Sender, Now)
      r.isLeft shouldBe true
      r.left.toOption.get should include("must be greater than previous")

      val ahead = good.copy(sequence = 2L)
      Main.dispatch(ahead, Sender, Now) shouldBe Right(())
    }

    it("accepts a well-formed EchoOrgRoleCredential and rejects unknown roles") {
      val good = EchoOrgRoleCredentialUpdate(
        credentialId = "urn:uuid:role-1",
        issuerOrgDID = OrgIssuer,
        memberDID    = Member,
        role         = "moderator",
        expiry       = Now + 86_400_000L,
        issuedAt     = Now - 1000L
      )
      Main.dispatch(good, Sender, Now) shouldBe Right(())

      val bad = good.copy(role = "godmode")
      Main.dispatch(bad, Sender, Now).isLeft shouldBe true
    }

    it("rejects updates from an unauthorized sender") {
      val u = TrustTierCommitmentUpdate(Subject, "a" * 64, Now - 100L)
      Main.dispatch(u, "did:key:zNOPE", Now).isLeft shouldBe true
    }
  }
}
