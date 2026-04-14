package com.echo.data_l1

import org.tessellation.currency.l1.CurrencyL1App

/**
 * Echo Data L1 — custom data application layer.
 *
 * Responsibilities:
 *   - Merkle root validation: structure, authorized sender DID (META-004)
 *   - Trust commitment validation: H(score||nonce) format (META-005)
 *
 * TODO: Override dataApplication with BaseDataApplicationL1Service
 * TODO: Implement MerkleRoot update validator
 * TODO: Implement TrustCommitment update validator
 */
object Main extends CurrencyL1App(
  name = "echo-data-l1",
  header = "Echo Data L1",
  clusterId = java.util.UUID.fromString("00000000-0000-0000-0000-000000000000"), // replace with real ID
  version = "0.1.0"
)
