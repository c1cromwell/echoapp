package com.echo.identity_l1

import cats.effect.{IO, Resource}
import cats.syntax.all._
import com.echo.shared_data.cluster.ClusterIds
import com.echo.shared_data.deserializers.IdentityDeserializers
import com.echo.shared_data.lifecycle.IdentityLifecycle
import com.echo.shared_data.serializers.IdentitySerializers
import com.echo.shared_data.types.IdentityUpdate
import com.echo.shared_data.types.{IdentityCalculatedState, IdentityOnChainState}
import io.circe.{Decoder, Encoder}
import io.constellationnetwork.currency.dataApplication._
import io.constellationnetwork.currency.dataApplication.dataApplication.{DataApplicationBlock, DataApplicationValidationErrorOr}
import io.constellationnetwork.currency.l1.CurrencyL1App
import io.constellationnetwork.ext.cats.effect.ResourceIO
import io.constellationnetwork.schema.semver.{MetagraphVersion, TessellationVersion}
import io.constellationnetwork.security.signature.Signed
import org.http4s.HttpRoutes
import org.http4s.circe.CirceEntityCodec.circeEntityDecoder

/**
 * Echo Identity Metagraph L1 — receives identity-update submissions from
 * the Identity Service and validates them before forwarding to L0 for
 * snapshot inclusion.
 */
object Main extends CurrencyL1App(
  name      = "echo-identity-l1",
  header    = "Echo Identity Metagraph L1",
  clusterId = ClusterIds.identityL1,
  tessellationVersion = TessellationVersion.unsafeFrom("4.0.0-rc.0"),
  metagraphVersion    = MetagraphVersion.unsafeFrom("0.1.0")
) {

  val authorizedSenderDid: String = IdentityLifecycle.authorizedSenderDid

  def now(): Long = System.currentTimeMillis()

  def recordPublishedSequence(issuerOrgDID: String, sequence: Long): Unit =
    com.echo.shared_data.state.IdentityRevocationSequences.recordPublished(issuerOrgDID, sequence)

  /** Pure dispatch — delegates to shared lifecycle (unit-tested). */
  def dispatch(
    update: IdentityUpdate,
    sender: String,
    nowMs:  Long = now()
  ): Either[String, Unit] =
    IdentityLifecycle.dispatch(update, sender, nowMs)

  private def makeBaseDataApplicationL1Service: BaseDataApplicationL1Service[IO] =
    BaseDataApplicationL1Service(
      new DataApplicationL1Service[IO, IdentityUpdate, IdentityOnChainState, IdentityCalculatedState] {
        override def validateUpdate(
          update: IdentityUpdate
        )(implicit context: L1NodeContext[IO]): IO[DataApplicationValidationErrorOr[Unit]] =
          IdentityLifecycle.validateUpdate[IO](update, now())

        override def routes(implicit context: L1NodeContext[IO]): HttpRoutes[IO] =
          HttpRoutes.empty

        override def dataEncoder: Encoder[IdentityUpdate] =
          implicitly[Encoder[IdentityUpdate]]

        override def dataDecoder: Decoder[IdentityUpdate] =
          implicitly[Decoder[IdentityUpdate]]

        override def calculatedStateEncoder: Encoder[IdentityCalculatedState] =
          implicitly[Encoder[IdentityCalculatedState]]

        override def calculatedStateDecoder: Decoder[IdentityCalculatedState] =
          implicitly[Decoder[IdentityCalculatedState]]

        override def signedDataEntityDecoder: org.http4s.EntityDecoder[IO, Signed[IdentityUpdate]] =
          circeEntityDecoder

        override def serializeBlock(block: Signed[DataApplicationBlock]): IO[Array[Byte]] =
          IO(IdentitySerializers.serializeBlock(block)(dataEncoder.asInstanceOf[Encoder[DataUpdate]]))

        override def deserializeBlock(bytes: Array[Byte]): IO[Either[Throwable, Signed[DataApplicationBlock]]] =
          IO(IdentityDeserializers.deserializeBlock(bytes)(dataDecoder.asInstanceOf[Decoder[DataUpdate]]))

        override def serializeState(state: IdentityOnChainState): IO[Array[Byte]] =
          IO(IdentitySerializers.serializeState(state))

        override def deserializeState(bytes: Array[Byte]): IO[Either[Throwable, IdentityOnChainState]] =
          IO(IdentityDeserializers.deserializeState(bytes))

        override def serializeUpdate(update: IdentityUpdate): IO[Array[Byte]] =
          IO(IdentitySerializers.serializeUpdate(update))

        override def deserializeUpdate(bytes: Array[Byte]): IO[Either[Throwable, IdentityUpdate]] =
          IO(IdentityDeserializers.deserializeUpdate(bytes))

        override def serializeCalculatedState(state: IdentityCalculatedState): IO[Array[Byte]] =
          IO(IdentitySerializers.serializeCalculatedState(state))

        override def deserializeCalculatedState(bytes: Array[Byte]): IO[Either[Throwable, IdentityCalculatedState]] =
          IO(IdentityDeserializers.deserializeCalculatedState(bytes))
      }
    )

  override def dataApplication: Option[Resource[IO, BaseDataApplicationL1Service[IO]]] =
    makeBaseDataApplicationL1Service.pure[IO].asResource.some
}
