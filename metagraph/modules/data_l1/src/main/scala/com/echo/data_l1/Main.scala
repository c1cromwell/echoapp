package com.echo.data_l1

import cats.effect.{IO, Resource}
import cats.syntax.all._
import com.echo.shared_data.cluster.ClusterIds
import com.echo.shared_data.deserializers.DataDeserializers
import com.echo.shared_data.lifecycle.DataLifecycle
import com.echo.shared_data.routes.DataRoutes
import com.echo.shared_data.serializers.DataSerializers
import com.echo.shared_data.types._
import io.circe.{Decoder, Encoder}
import io.constellationnetwork.currency.dataApplication._
import io.constellationnetwork.currency.dataApplication.dataApplication.{DataApplicationBlock, DataApplicationValidationErrorOr}
import io.constellationnetwork.currency.l1.CurrencyL1App
import io.constellationnetwork.ext.cats.effect.ResourceIO
import io.constellationnetwork.schema.semver.{MetagraphVersion, TessellationVersion}
import io.constellationnetwork.security.signature.Signed
import org.http4s.circe.CirceEntityCodec.circeEntityDecoder

/**
 * Echo Data L1 — anchors Merkle roots and trust-score commitments.
 */
object Main extends CurrencyL1App(
  name      = "echo-data-l1",
  header    = "Echo Data L1",
  clusterId = ClusterIds.dataL1,
  tessellationVersion = TessellationVersion.unsafeFrom("4.0.0-rc.0"),
  metagraphVersion    = MetagraphVersion.unsafeFrom("0.1.0")
) {

  /** Pure dispatch — delegates to shared lifecycle (unit-tested). */
  def dispatch(update: EchoUpdate): Either[String, Unit] =
    update match {
      case u: DataLayerUpdate => DataLifecycle.dispatch(u)
      case other              => Left(s"Data L1 does not accept update type ${other.getClass.getSimpleName}")
    }

  private def makeBaseDataApplicationL1Service: BaseDataApplicationL1Service[IO] =
    BaseDataApplicationL1Service(
      new DataApplicationL1Service[IO, DataLayerUpdate, DataLayerOnChainState, DataLayerCalculatedState] {
        override def validateUpdate(
          update: DataLayerUpdate
        )(implicit context: L1NodeContext[IO]): IO[DataApplicationValidationErrorOr[Unit]] =
          DataLifecycle.validateUpdate[IO](update)

        override def routes(implicit context: L1NodeContext[IO]): org.http4s.HttpRoutes[IO] =
          DataRoutes.merkleRootRoutes

        override def dataEncoder: Encoder[DataLayerUpdate] =
          implicitly[Encoder[DataLayerUpdate]]

        override def dataDecoder: Decoder[DataLayerUpdate] =
          implicitly[Decoder[DataLayerUpdate]]

        override def calculatedStateEncoder: Encoder[DataLayerCalculatedState] =
          implicitly[Encoder[DataLayerCalculatedState]]

        override def calculatedStateDecoder: Decoder[DataLayerCalculatedState] =
          implicitly[Decoder[DataLayerCalculatedState]]

        override def signedDataEntityDecoder: org.http4s.EntityDecoder[IO, Signed[DataLayerUpdate]] =
          circeEntityDecoder

        override def serializeBlock(block: Signed[DataApplicationBlock]): IO[Array[Byte]] =
          IO(DataSerializers.serializeBlock(block)(dataEncoder.asInstanceOf[Encoder[DataUpdate]]))

        override def deserializeBlock(bytes: Array[Byte]): IO[Either[Throwable, Signed[DataApplicationBlock]]] =
          IO(DataDeserializers.deserializeBlock(bytes)(dataDecoder.asInstanceOf[Decoder[DataUpdate]]))

        override def serializeState(state: DataLayerOnChainState): IO[Array[Byte]] =
          IO(DataSerializers.serializeState(state))

        override def deserializeState(bytes: Array[Byte]): IO[Either[Throwable, DataLayerOnChainState]] =
          IO(DataDeserializers.deserializeState(bytes))

        override def serializeUpdate(update: DataLayerUpdate): IO[Array[Byte]] =
          IO(DataSerializers.serializeUpdate(update))

        override def deserializeUpdate(bytes: Array[Byte]): IO[Either[Throwable, DataLayerUpdate]] =
          IO(DataDeserializers.deserializeUpdate(bytes))

        override def serializeCalculatedState(state: DataLayerCalculatedState): IO[Array[Byte]] =
          IO(DataSerializers.serializeCalculatedState(state))

        override def deserializeCalculatedState(bytes: Array[Byte]): IO[Either[Throwable, DataLayerCalculatedState]] =
          IO(DataDeserializers.deserializeCalculatedState(bytes))
      }
    )

  override def dataApplication: Option[Resource[IO, BaseDataApplicationL1Service[IO]]] =
    makeBaseDataApplicationL1Service.pure[IO].asResource.some
}
