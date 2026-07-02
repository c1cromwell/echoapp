package com.echo.l1

import com.echo.shared_data.cluster.ClusterIds
import com.echo.shared_data.types._
import com.echo.shared_data.validations.Validations
import io.constellationnetwork.currency.l1.CurrencyL1App
import io.constellationnetwork.schema.semver.{TessellationVersion, MetagraphVersion}

/**
 * Echo Currency L1 — token transactions (lock, delegate, withdraw, claim).
 *
 * Validators wired:
 *   - TokenLock   → Validations.validateTokenLock
 *   - RewardClaim → Validations.validateRewardClaim
 *
 * StakeDelegationUpdate / WithdrawLockUpdate carry no per-update invariants
 * yet (cooldowns and balance checks live at the L0 combiner stage); they
 * pass through the dispatcher to keep the surface explicit.
 */
object Main extends CurrencyL1App(
  name      = "echo-currency-l1",
  header    = "Echo Currency L1",
  clusterId = ClusterIds.currencyL1,
  tessellationVersion = TessellationVersion.unsafeFrom("4.0.0-rc.0"),
  metagraphVersion    = MetagraphVersion.unsafeFrom("0.1.0")
) {

  def dispatch(update: EchoUpdate): Either[String, Unit] = update match {
    case u: TokenLockUpdate          => Validations.validateTokenLock(u)
    case u: RewardClaimUpdate        => Validations.validateRewardClaim(u)
    case u: MintUpdate               => Validations.validateMint(u)
    case u: FounderRevocationUpdate  => Validations.validateFounderRevocation(u)
    case _: StakeDelegationUpdate    => Right(())
    case _: WithdrawLockUpdate       => Right(())
    case other =>
      Left(s"Currency L1 does not accept update type ${other.getClass.getSimpleName}")
  }
}
