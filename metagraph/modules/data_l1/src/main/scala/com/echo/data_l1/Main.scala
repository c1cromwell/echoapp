package com.echo.data_l1

import com.echo.shared_data.cluster.ClusterIds
import com.echo.shared_data.types._
import com.echo.shared_data.validations.Validations
import io.constellationnetwork.currency.l1.CurrencyL1App
import io.constellationnetwork.schema.semver.{TessellationVersion, MetagraphVersion}

/**
 * Echo Data L1 — anchors Merkle roots and trust-score commitments.
 *
 * Validators wired:
 *   - Merkle root      → Validations.validateMerkleRoot
 *   - Trust commitment → Validations.validateTrustCommitment
 *
 * Wiring is done through the `dispatch` helper, which is invoked from the
 * Tessellation `BaseDataApplicationL1Service` override (or the equivalent
 * mempool admission hook in the SDK). Keeping dispatch as a pure function
 * lets us unit-test the wired behavior without booting the L1 cluster.
 */
object Main extends CurrencyL1App(
  name      = "echo-data-l1",
  header    = "Echo Data L1",
  clusterId = ClusterIds.dataL1,
  tessellationVersion = TessellationVersion.unsafeFrom("4.0.0-rc.0"),
  metagraphVersion    = MetagraphVersion.unsafeFrom("0.1.0")
) {

  /** Pure dispatch: runs the correct validator for each Data-L1 update. */
  def dispatch(update: EchoUpdate): Either[String, Unit] = update match {
    case u: MerkleRootUpdate      => Validations.validateMerkleRoot(u)
    case u: TrustCommitmentUpdate => Validations.validateTrustCommitment(u)
    case other =>
      Left(s"Data L1 does not accept update type ${other.getClass.getSimpleName}")
  }
}
