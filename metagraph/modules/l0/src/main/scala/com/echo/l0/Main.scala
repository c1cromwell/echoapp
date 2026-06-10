package com.echo.l0

import cats.data.NonEmptyList
import cats.effect.{IO, Resource}
import cats.syntax.all._
import com.echo.shared_data.cluster.ClusterIds
import com.echo.shared_data.deserializers.DataDeserializers
import com.echo.shared_data.lifecycle.{DataCalculatedStateService, DataLifecycle}
import com.echo.shared_data.serializers.DataSerializers
import com.echo.shared_data.types._
import io.circe.{Decoder, Encoder}
import io.constellationnetwork.currency.dataApplication._
import io.constellationnetwork.currency.dataApplication.dataApplication.{DataApplicationBlock, DataApplicationValidationErrorOr}
import io.constellationnetwork.currency.l0.CurrencyL0App
import io.constellationnetwork.ext.cats.effect.ResourceIO
import io.constellationnetwork.schema.SnapshotOrdinal
import io.constellationnetwork.schema.semver.{MetagraphVersion, TessellationVersion}
import io.constellationnetwork.security.SecurityProvider
import io.constellationnetwork.security.hash.Hash
import io.constellationnetwork.security.signature.Signed
import org.http4s.HttpRoutes
import org.http4s.circe.CirceEntityCodec.circeEntityDecoder

/**
 * Echo Metagraph L0 — consensus layer.
 *
 * Phase 1: data application folds Data L1 Merkle root and trust commitments
 * into metagraph snapshots.
 */
object Main extends CurrencyL0App(
  name      = "echo-metagraph-l0",
  header    = "Echo Metagraph L0",
  clusterId = ClusterIds.metagraphL0,
  tessellationVersion = TessellationVersion.unsafeFrom("4.0.0-rc.0"),
  metagraphVersion    = MetagraphVersion.unsafeFrom("0.1.0")
) {

  /**
   * Total ECHO supply: 1,000,000,000 tokens (1e8 datum each).
   */
  val TotalSupply: Long      = 1000000000L * 100000000L
  val CommunityPool: Long    = (TotalSupply * 0.40).toLong
  val DevelopmentFund: Long  = (TotalSupply * 0.20).toLong
  val EcosystemGrants: Long  = (TotalSupply * 0.15).toLong
  val TeamAdvisors: Long     = (TotalSupply * 0.15).toLong
  val LiquidityReserve: Long = (TotalSupply * 0.10).toLong

  private def makeBaseDataApplicationL0Service(
    calculatedStateService: DataCalculatedStateService[IO]
  ): BaseDataApplicationL0Service[IO] =
    BaseDataApplicationL0Service(
      new DataApplicationL0Service[IO, DataLayerUpdate, DataLayerOnChainState, DataLayerCalculatedState] {
        override def genesis: DataState[DataLayerOnChainState, DataLayerCalculatedState] =
          DataState(DataLayerOnChainState.empty, DataLayerCalculatedState.empty)

        override def validateData(
          state:   DataState[DataLayerOnChainState, DataLayerCalculatedState],
          updates: NonEmptyList[Signed[DataLayerUpdate]]
        )(implicit context: L0NodeContext[IO]): IO[DataApplicationValidationErrorOr[Unit]] = {
          implicit val sp: SecurityProvider[IO] = context.securityProvider
          DataLifecycle.validateData[IO](state, updates)
        }

        override def combine(
          state:   DataState[DataLayerOnChainState, DataLayerCalculatedState],
          updates: List[Signed[DataLayerUpdate]]
        )(implicit context: L0NodeContext[IO]): IO[DataState[DataLayerOnChainState, DataLayerCalculatedState]] =
          DataLifecycle.combine[IO](state, updates)

        override def dataEncoder: Encoder[DataLayerUpdate] =
          implicitly[Encoder[DataLayerUpdate]]

        override def calculatedStateEncoder: Encoder[DataLayerCalculatedState] =
          implicitly[Encoder[DataLayerCalculatedState]]

        override def dataDecoder: Decoder[DataLayerUpdate] =
          implicitly[Decoder[DataLayerUpdate]]

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

        override def getCalculatedState(implicit context: L0NodeContext[IO]): IO[(SnapshotOrdinal, DataLayerCalculatedState)] =
          calculatedStateService.getCalculatedState

        override def setCalculatedState(
          ordinal: SnapshotOrdinal,
          state:   DataLayerCalculatedState
        )(implicit context: L0NodeContext[IO]): IO[Boolean] =
          calculatedStateService.setCalculatedState(ordinal, state)

        override def hashCalculatedState(state: DataLayerCalculatedState)(implicit context: L0NodeContext[IO]): IO[Hash] =
          calculatedStateService.hashCalculatedState(state)

        override def routes(implicit context: L0NodeContext[IO]): HttpRoutes[IO] =
          HttpRoutes.empty

        override def serializeCalculatedState(state: DataLayerCalculatedState): IO[Array[Byte]] =
          IO(DataSerializers.serializeCalculatedState(state))

        override def deserializeCalculatedState(bytes: Array[Byte]): IO[Either[Throwable, DataLayerCalculatedState]] =
          IO(DataDeserializers.deserializeCalculatedState(bytes))
      }
    )

  private def makeL0Service: IO[BaseDataApplicationL0Service[IO]] =
    for {
      calc <- DataCalculatedStateService.make[IO]
    } yield makeBaseDataApplicationL0Service(calc)

  override def dataApplication: Option[Resource[IO, BaseDataApplicationL0Service[IO]]] =
    makeL0Service.asResource.some
}
