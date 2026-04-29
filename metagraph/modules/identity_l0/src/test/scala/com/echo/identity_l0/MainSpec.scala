package com.echo.identity_l0

import org.scalatest.funspec.AnyFunSpec
import org.scalatest.matchers.should.Matchers

/**
 * Smoke spec for the Identity Metagraph L0 application object.
 *
 * Just touching `Main` forces the SDK constructor expression
 * (`extends CurrencyL0App(...)`) to evaluate, which in turn parses the
 * cluster UUID — so this also asserts that ClusterIds defaults are valid
 * UUIDs at JVM startup.
 */
final class MainSpec extends AnyFunSpec with Matchers {

  describe("Identity L0 Main") {
    it("declares the Phase-1 supported VC types") {
      Main.SupportedCredentialTypes should contain allOf (
        "TrustTierCredential",
        "EchoOrgRoleCredential",
        "KYCCredential",
        "ProfessionalCredential"
      )
    }

    it("constructs without throwing (cluster UUID parses, SDK accepts the args)") {
      noException should be thrownBy Main.toString
    }
  }
}
