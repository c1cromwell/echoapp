package com.echo.shared_data.lifecycle

import cats.data.NonEmptyList
import cats.effect.Async
import cats.syntax.all._
import com.echo.shared_data.combiners.DataCombiners
import com.echo.shared_data.errors.DataErrors
import com.echo.shared_data.state.DataMerkleRootIndex
import com.echo.shared_data.types._
import com.echo.shared_data.validations.Validations
import io.constellationnetwork.currency.dataApplication.dataApplication.DataApplicationValidationErrorOr
import io.constellationnetwork.currency.dataApplication.DataState
import io.constellationnetwork.security.SecurityProvider
import io.constellationnetwork.security.signature.Signed

object DataLifecycle {

  def validateUpdate[F[_]: Async](
    update: DataLayerUpdate
  ): F[DataApplicationValidationErrorOr[Unit]] =
    Async[F].delay {
      dispatch(update) match {
        case Right(()) =>
          update match {
            case u: MerkleRootUpdate =>
              DataMerkleRootIndex.record(u.root, u.leafCount, finalized = true)
            case _ => ()
          }
          DataErrors.valid
        case Left(msg) => DataErrors.invalid(msg)
      }
    }

  def validateData[F[_]: Async: SecurityProvider](
    state:   DataState[DataLayerOnChainState, DataLayerCalculatedState],
    updates: NonEmptyList[Signed[DataLayerUpdate]]
  ): F[DataApplicationValidationErrorOr[Unit]] =
    updates.traverse { signed =>
      Async[F].delay(dispatch(signed.value) match {
        case Right(()) => DataErrors.valid
        case Left(msg) => DataErrors.invalid(msg)
      })
    }.map(_.reduce)

  def combine[F[_]: Async](
    state:   DataState[DataLayerOnChainState, DataLayerCalculatedState],
    updates: List[Signed[DataLayerUpdate]]
  ): F[DataState[DataLayerOnChainState, DataLayerCalculatedState]] =
    Async[F].delay {
      updates.foldLeft(state) { (acc, signed) =>
        val next = DataCombiners.combineUpdate(signed, acc)
        signed.value match {
          case u: MerkleRootUpdate =>
            DataMerkleRootIndex.record(u.root, u.leafCount, finalized = true)
          case _ => ()
        }
        next
      }
    }

  def dispatch(update: DataLayerUpdate): Either[String, Unit] =
    update match {
      case u: MerkleRootUpdate =>
        Validations.rejectUpdatePII(u).flatMap(_ => Validations.validateMerkleRoot(u))
      case u: TrustCommitmentUpdate =>
        Validations.rejectUpdatePII(u).flatMap(_ => Validations.validateTrustCommitment(u))
      case other =>
        Left(s"Data L1 does not accept update type ${other.getClass.getSimpleName}")
    }
}
