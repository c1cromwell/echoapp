package com.echo.shared_data.lifecycle

import cats.data.NonEmptyList
import cats.effect.Async
import cats.syntax.all._
import com.echo.shared_data.combiners.IdentityCombiners
import com.echo.shared_data.errors.IdentityErrors
import com.echo.shared_data.types._
import com.echo.shared_data.validations.IdentityValidations
import io.constellationnetwork.currency.dataApplication.dataApplication.DataApplicationValidationErrorOr
import io.constellationnetwork.currency.dataApplication.{DataState, L0NodeContext}
import io.constellationnetwork.security.SecurityProvider
import io.constellationnetwork.security.signature.Signed

object IdentityLifecycle {

  def authorizedSenderDid: String =
    sys.env.getOrElse("IDENTITY_SERVICE_DID", "did:key:z__UNSET__IDENTITY_SERVICE_DID__")

  def authorizedPublicKeyHex: Option[String] =
    sys.env.get("IDENTITY_SERVICE_PUBLIC_KEY_HEX").filter(_.nonEmpty)

  def validateUpdate[F[_]: Async](
    update: IdentityUpdate,
    nowMs:  Long = System.currentTimeMillis()
  ): F[DataApplicationValidationErrorOr[Unit]] =
    Async[F].delay {
      dispatch(update, authorizedSenderDid, nowMs) match {
        case Right(()) => IdentityErrors.valid
        case Left(msg) => IdentityErrors.invalid(msg)
      }
    }

  def validateData[F[_]: Async: SecurityProvider](
    state:   DataState[IdentityOnChainState, IdentityCalculatedState],
    updates: NonEmptyList[Signed[IdentityUpdate]]
  ): F[DataApplicationValidationErrorOr[Unit]] =
    updates.traverse { signed =>
      verifySigner[F](signed).attempt.flatMap {
        case Left(err) => Async[F].pure(IdentityErrors.invalid(err.getMessage))
        case Right(_)  => Async[F].pure(validateSignedUpdate(signed, state))
      }
    }.map(_.reduce)

  def combine[F[_]: Async](
    state:   DataState[IdentityOnChainState, IdentityCalculatedState],
    updates: List[Signed[IdentityUpdate]]
  ): F[DataState[IdentityOnChainState, IdentityCalculatedState]] =
    Async[F].pure(updates.foldLeft(state)((acc, signed) => IdentityCombiners.combineUpdate(signed, acc)))

  def dispatch(
    update: IdentityUpdate,
    sender: String,
    nowMs:  Long,
    state:  Option[DataState[IdentityOnChainState, IdentityCalculatedState]] = None
  ): Either[String, Unit] =
    update match {
      case u: VCIssuanceUpdate =>
        IdentityValidations.validateVCIssuance(u, sender, authorizedSenderDid, nowMs)
      case u: TrustTierCommitmentUpdate =>
        IdentityValidations.validateTrustTierCommitment(u, sender, authorizedSenderDid, nowMs)
      case u: StatusList2021BatchUpdate =>
        val prev = com.echo.shared_data.state.IdentityRevocationSequences.previousFor(u.issuerOrgDID)
        IdentityValidations.validateStatusList2021(u, sender, authorizedSenderDid, prev, nowMs)
      case u: EchoOrgRoleCredentialUpdate =>
        IdentityValidations.validateEchoOrgRoleCredential(u, sender, authorizedSenderDid, nowMs)
      case u: DeviceKeyRegistrationUpdate =>
        IdentityValidations.validateDeviceKeyRegistration(u, sender, authorizedSenderDid, nowMs)
      case u: UsernameRegistrationUpdate =>
        val owner = state.flatMap(s => s.onChain.usernames.get(u.username.toLowerCase).map(_.subjectDID))
        IdentityValidations.validateUsernameRegistration(u, sender, authorizedSenderDid, owner, nowMs)
    }

  private def validateSignedUpdate(
    signed: Signed[IdentityUpdate],
    state:  DataState[IdentityOnChainState, IdentityCalculatedState]
  ): DataApplicationValidationErrorOr[Unit] =
    dispatch(signed.value, authorizedSenderDid, System.currentTimeMillis(), Some(state)) match {
      case Right(()) => IdentityErrors.valid
      case Left(msg) => IdentityErrors.invalid(msg)
    }

  private def verifySigner[F[_]: Async](signed: Signed[IdentityUpdate]): F[Unit] =
    authorizedPublicKeyHex match {
      case None => Async[F].unit
      case Some(expected) =>
        val proofIds = signed.proofs.toNonEmptyList.toList.map(_.id.toString)
        Async[F].raiseUnless(proofIds.exists(_.equalsIgnoreCase(expected)))(
          new IllegalArgumentException("proof id must match IDENTITY_SERVICE_PUBLIC_KEY_HEX")
        )
    }
}
