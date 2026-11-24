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

  /**
   * Public keys (Tessellation peer-id hex, 128 chars) allowed to sign Data L1
   * updates, from the comma-separated `DATA_AUTHORIZED_SIGNERS`. Unset or empty
   * means no restriction, which is only acceptable on a local cluster: without
   * it, anyone who can reach a Data L1 port can anchor roots. The production
   * launcher (deploy/sovereign/node/echo-node.sh) refuses to start Metagraph L0
   * unless it is set.
   */
  def parseAuthorizedSigners(raw: Option[String]): Option[Set[String]] =
    raw.map(_.split(',').map(_.trim.toLowerCase).filter(_.nonEmpty).toSet).filter(_.nonEmpty)

  lazy val authorizedSigners: Option[Set[String]] =
    parseAuthorizedSigners(sys.env.get("DATA_AUTHORIZED_SIGNERS"))

  /** Every proof on the update must come from an authorised key. */
  def checkSigners(proofIds: List[String], allowed: Option[Set[String]]): Either[String, Unit] =
    allowed match {
      case None => Right(())
      case Some(keys) =>
        proofIds.find(id => !keys.contains(id.toLowerCase)) match {
          case None if proofIds.nonEmpty => Right(())
          case None                      => Left("Data L1 update carries no signature proof")
          case Some(id) =>
            Left(s"Data L1 update signed by a key that is not an authorised signer (${id.take(16)}…)")
        }
    }

  def validateUpdate[F[_]: Async](
    update: DataLayerUpdate
  ): F[DataApplicationValidationErrorOr[Unit]] =
    Async[F].delay {
      dispatch(update) match {
        case Right(()) =>
          update match {
            // L1 acceptance is NOT finality: the update is only in this node's
            // mempool. It becomes final when Metagraph L0 folds it in `combine`.
            case u: MerkleRootUpdate =>
              DataMerkleRootIndex.record(u.root, u.leafCount, finalized = false)
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
    validateSigned[F](updates, authorizedSigners)

  def validateSigned[F[_]: Async](
    updates: NonEmptyList[Signed[DataLayerUpdate]],
    allowed: Option[Set[String]]
  ): F[DataApplicationValidationErrorOr[Unit]] =
    updates.traverse { signed =>
      Async[F].delay {
        val proofIds = signed.proofs.toNonEmptyList.toList.map(_.id.hex.value)
        checkSigners(proofIds, allowed).flatMap(_ => dispatch(signed.value)) match {
          case Right(()) => DataErrors.valid
          case Left(msg) => DataErrors.invalid(msg)
        }
      }
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
