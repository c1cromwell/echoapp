package com.echo.identity_l1

import com.echo.shared_data.cluster.ClusterIds
import com.echo.shared_data.types._
import com.echo.shared_data.validations.IdentityValidations
import org.tessellation.currency.l1.CurrencyL1App

/**
 * Echo Identity Metagraph L1 — receives identity-update submissions from
 * the Identity Service and validates them before forwarding to L0 for
 * snapshot inclusion.
 *
 * Validators wired in Phase 1:
 *   - VC issuance              (W3C Verifiable Credentials 2.0 metadata)
 *   - Trust tier commitments   (H(tier || nonce))
 *   - StatusList2021 batches   (revocation bit vectors, 5-min cadence)
 *   - EchoOrgRoleCredentials   (org membership)
 *
 * Authorization model: only the Identity Service DID (env var
 * `IDENTITY_SERVICE_DID`) may submit. Phase 4+ this expands to a
 * delegated-issuer registry (per-org issuer DIDs proven by an
 * EchoOrgRoleCredential anchored on this same metagraph).
 *
 * The validators are pure functions in `IdentityValidations`. Wiring
 * them into Tessellation 4.0.0-rc.0's L1 update pipeline happens by
 * overriding `CustomContextualTransactionValidator` (or the equivalent
 * `BaseDataApplicationL1Service` hook for non-currency updates) — see
 * the dispatch helper below.
 */
object Main extends CurrencyL1App(
  name      = "echo-identity-l1",
  header    = "Echo Identity Metagraph L1",
  clusterId = ClusterIds.identityL1,
  version   = "0.1.0"
) {

  /**
   * Authorized sender for Phase 1: read once at JVM start.
   * Production deployments MUST set `IDENTITY_SERVICE_DID`.
   */
  val authorizedSenderDid: String =
    sys.env.getOrElse("IDENTITY_SERVICE_DID", "did:key:z__UNSET__IDENTITY_SERVICE_DID__")

  /** Pluggable clock so tests can pin "now". */
  def now(): Long = System.currentTimeMillis()

  /** Last-known StatusList2021 sequence per issuerOrgDID — populated by the
   *  L0 snapshot reader on application start. In Phase 1 we keep this in
   *  memory; persistence belongs to the Tessellation snapshot stream.
   *  NOTE: tests inject sequences directly via `dispatch`. */
  @volatile private var revocationSequences: Map[String, Long] = Map.empty.withDefaultValue(0L)

  /** Update the cached sequence after a successful StatusList2021 batch
   *  is folded into a snapshot. Called from the L0 combiner. */
  def recordPublishedSequence(issuerOrgDID: String, sequence: Long): Unit =
    synchronized {
      revocationSequences = revocationSequences.updated(issuerOrgDID, sequence)
    }

  /**
   * Single dispatch point: runs the correct pure validator for the update
   * type. Returns Right(()) if the update should be accepted into the
   * mempool, Left(reason) to reject it.
   *
   * Tessellation's `CustomContextualTransactionValidator` callback
   * delegates here so the L1 application logic stays a thin wiring layer
   * around the pure rules.
   */
  def dispatch(
    update: IdentityUpdate,
    sender: String,
    nowMs:  Long = now()
  ): Either[String, Unit] = update match {
    case u: VCIssuanceUpdate =>
      IdentityValidations.validateVCIssuance(u, sender, authorizedSenderDid, nowMs)

    case u: TrustTierCommitmentUpdate =>
      IdentityValidations.validateTrustTierCommitment(u, sender, authorizedSenderDid, nowMs)

    case u: StatusList2021BatchUpdate =>
      val prev = revocationSequences.getOrElse(u.issuerOrgDID, 0L)
      IdentityValidations.validateStatusList2021(u, sender, authorizedSenderDid, prev, nowMs)

    case u: EchoOrgRoleCredentialUpdate =>
      IdentityValidations.validateEchoOrgRoleCredential(u, sender, authorizedSenderDid, nowMs)
  }
}
