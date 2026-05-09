package com.echo.shared_data.cluster

import io.constellationnetwork.schema.cluster.ClusterId

/**
 * Per-environment cluster IDs for every metagraph layer.
 *
 * Each `Main` extends `CurrencyL{0,1}App(...)` whose constructor takes a
 * `clusterId: ClusterId`. Production deployments MUST set the matching env
 * var below — the defaults are deterministic non-zero UUIDs pinned to the
 * local Phase-1 testnet so devs can boot without ceremony.
 *
 * Generate fresh UUIDs for shared environments with `uuidgen` or
 * `python -c "import uuid; print(uuid.uuid4())"`.
 */
object ClusterIds {
  private val DefaultIdentityL0   = "11111111-1111-4111-8111-111111111111"
  private val DefaultIdentityL1   = "22222222-2222-4222-8222-222222222222"
  private val DefaultMetagraphL0  = "33333333-3333-4333-8333-333333333333"
  private val DefaultCurrencyL1   = "44444444-4444-4444-8444-444444444444"
  private val DefaultDataL1       = "55555555-5555-4555-8555-555555555555"

  private def fromEnv(envVar: String, default: String): ClusterId =
    ClusterId(java.util.UUID.fromString(sys.env.getOrElse(envVar, default)))

  val identityL0:  ClusterId = fromEnv("IDENTITY_L0_CLUSTER_ID",  DefaultIdentityL0)
  val identityL1:  ClusterId = fromEnv("IDENTITY_L1_CLUSTER_ID",  DefaultIdentityL1)
  val metagraphL0: ClusterId = fromEnv("METAGRAPH_L0_CLUSTER_ID", DefaultMetagraphL0)
  val currencyL1:  ClusterId = fromEnv("CURRENCY_L1_CLUSTER_ID",  DefaultCurrencyL1)
  val dataL1:      ClusterId = fromEnv("DATA_L1_CLUSTER_ID",      DefaultDataL1)
}
