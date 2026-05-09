package com.echo.shared_data.state

/**
 * In-process StatusList2021 sequence baseline per issuer org DID (WO-272).
 *
 * Identity L1 mempool validation reads the previous sequence via
 * [[previousFor]]; after L0 folds a finalized snapshot containing a batch,
 * the L0 combiner must call [[recordPublished]] so the next batch's
 * monotonicity check sees the correct baseline.
 *
 * Phase 1 keeps this map in memory on each validator process; persistence
 * is the Tessellation snapshot stream.
 */
object IdentityRevocationSequences {

  @volatile private var sequences: Map[String, Long] = Map.empty.withDefaultValue(0L)

  def previousFor(issuerOrgDID: String): Long =
    sequences.getOrElse(issuerOrgDID, 0L)

  def recordPublished(issuerOrgDID: String, sequence: Long): Unit =
    synchronized {
      sequences = sequences.updated(issuerOrgDID, sequence)
    }

  /** Test-only reset (not for production use). */
  private[echo] def resetForTests(): Unit =
    synchronized {
      sequences = Map.empty.withDefaultValue(0L)
    }
}
