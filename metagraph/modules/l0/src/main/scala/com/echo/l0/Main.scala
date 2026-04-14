package com.echo.l0

import org.tessellation.currency.l0.CurrencyL0App
import org.tessellation.schema.balance.Amount

/**
 * Echo Metagraph L0 — consensus layer.
 *
 * Responsibilities:
 *   - Genesis: mint 1B ECHO, allocate across 5 pools (META-001)
 *   - Rewards: auto-scaling tier multipliers (META-002)
 *   - State combination: fold Currency L1 + Data L1 updates into snapshots
 *
 * TODO: Wire dataApplication with EchoOnChainState/EchoCalculatedState combiners
 * TODO: Override Rewards trait for custom emission schedule
 * TODO: Generate genesis.csv with 5-pool allocation
 */
object Main extends CurrencyL0App(
  name = "echo-metagraph-l0",
  header = "Echo Metagraph L0",
  clusterId = java.util.UUID.fromString("00000000-0000-0000-0000-000000000000"), // replace with real ID
  version = "0.1.0"
) {

  /**
   * Total ECHO supply: 1,000,000,000 tokens (1e8 datum each).
   * Genesis allocation:
   *   - Community rewards pool: 40% (400M)
   *   - Development fund:       20% (200M)
   *   - Ecosystem grants:       15% (150M)
   *   - Team & advisors:        15% (150M, 4-year vest)
   *   - Liquidity reserve:      10% (100M)
   */
  val TotalSupply: Long        = 1000000000L * 100000000L // 1B ECHO in datum
  val CommunityPool: Long      = (TotalSupply * 0.40).toLong
  val DevelopmentFund: Long    = (TotalSupply * 0.20).toLong
  val EcosystemGrants: Long    = (TotalSupply * 0.15).toLong
  val TeamAdvisors: Long       = (TotalSupply * 0.15).toLong
  val LiquidityReserve: Long   = (TotalSupply * 0.10).toLong
}
