package com.echo.l1

import org.tessellation.currency.l1.CurrencyL1App

/**
 * Echo Currency L1 — token transaction layer.
 *
 * Responsibilities:
 *   - TokenLock validation: tier minimums, lock durations (META-003)
 *   - Reward claim validation: tier multipliers, daily caps (META-002)
 *   - Stake delegation and withdrawal cooldown enforcement
 *
 * TODO: Override CustomContextualTransactionValidator for TokenLock
 * TODO: Implement reward claim validation with auto-scaling multipliers
 * TODO: Enforce 14-day unstaking cooldown (governance-adjustable)
 */
object Main extends CurrencyL1App(
  name = "echo-currency-l1",
  header = "Echo Currency L1",
  clusterId = java.util.UUID.fromString("00000000-0000-0000-0000-000000000000"), // replace with real ID
  version = "0.1.0"
)
