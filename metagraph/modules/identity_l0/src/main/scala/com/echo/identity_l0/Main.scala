package com.echo.identity_l0

import cats.data.NonEmptyList
import cats.effect.{IO, Resource}
import cats.syntax.all._
import com.echo.shared_data.cluster.ClusterIds
import com.echo.shared_data.deserializers.IdentityDeserializers
import com.echo.shared_data.lifecycle.{IdentityCalculatedStateService, IdentityLifecycle}
import com.echo.shared_data.serializers.IdentitySerializers
import com.echo.shared_data.types.IdentityUpdate
import com.echo.shared_data.types.{IdentityCalculatedState, IdentityOnChainState}
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
 * Echo Identity Metagraph L0 — consensus layer for the Identity Metagraph.
 */
object Main extends CurrencyL0App(
  name      = "echo-identity-l0",
  header    = "Echo Identity Metagraph L0",
  clusterId = ClusterIds.identityL0,
  tessellationVersion = TessellationVersion.unsafeFrom("4.0.0-rc.0"),
  metagraphVersion    = MetagraphVersion.unsafeFrom("0.1.0")
) {

  val SupportedCredentialTypes: Set[String] = Set(
    "TrustTierCredential",
    "EchoOrgRoleCredential",
    "KYCCredential",
    "ProfessionalCredential"
  )

  def onStatusListSnapshotFinalized(issuerOrgDID: String, sequence: Long): Unit =
    com.echo.shared_data.state.IdentityRevocationSequences.recordPublished(issuerOrgDID, sequence)

  private def makeBaseDataApplicationL0Service(
    calculatedStateService: IdentityCalculatedStateService[IO]
  ): BaseDataApplicationL0Service[IO] =
    BaseDataApplicationL0Service(
      new DataApplicationL0Service[IO, IdentityUpdate, IdentityOnChainState, IdentityCalculatedState] {
        override def genesis: DataState[IdentityOnChainState, IdentityCalculatedState] =
          DataState(IdentityOnChainState.empty, IdentityCalculatedState.empty)

        override def validateData(
          state:   DataState[IdentityOnChainState, IdentityCalculatedState],
          updates: NonEmptyList[Signed[IdentityUpdate]]
        )(implicit context: L0NodeContext[IO]): IO[DataApplicationValidationErrorOr[Unit]] = {
          implicit val sp: SecurityProvider[IO] = context.securityProvider
          IdentityLifecycle.validateData[IO](state, updates)
        }

        override def combine(
          state:   DataState[IdentityOnChainState, IdentityCalculatedState],
          updates: List[Signed[IdentityUpdate]]
        )(implicit context: L0NodeContext[IO]): IO[DataState[IdentityOnChainState, IdentityCalculatedState]] =
          IdentityLifecycle.combine[IO](state, updates)

        override def dataEncoder: Encoder[IdentityUpdate] =
          implicitly[Encoder[IdentityUpdate]]

        override def calculatedStateEncoder: Encoder[IdentityCalculatedState] =
          implicitly[Encoder[IdentityCalculatedState]]

        override def dataDecoder: Decoder[IdentityUpdate] =
          implicitly[Decoder[IdentityUpdate]]

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

        override def getCalculatedState(implicit context: L0NodeContext[IO]): IO[(SnapshotOrdinal, IdentityCalculatedState)] =
          calculatedStateService.getCalculatedState

        override def setCalculatedState(
          ordinal: SnapshotOrdinal,
          state:   IdentityCalculatedState
        )(implicit context: L0NodeContext[IO]): IO[Boolean] =
          calculatedStateService.setCalculatedState(ordinal, state)

        override def hashCalculatedState(state: IdentityCalculatedState)(implicit context: L0NodeContext[IO]): IO[Hash] =
          calculatedStateService.hashCalculatedState(state)

        override def routes(implicit context: L0NodeContext[IO]): HttpRoutes[IO] =
          HttpRoutes.empty

        override def serializeCalculatedState(state: IdentityCalculatedState): IO[Array[Byte]] =
          IO(IdentitySerializers.serializeCalculatedState(state))

        override def deserializeCalculatedState(bytes: Array[Byte]): IO[Either[Throwable, IdentityCalculatedState]] =
          IO(IdentityDeserializers.deserializeCalculatedState(bytes))
      }
    )

  private def makeL0Service: IO[BaseDataApplicationL0Service[IO]] =
    for {
      calc <- IdentityCalculatedStateService.make[IO]
    } yield makeBaseDataApplicationL0Service(calc)

  override def dataApplication: Option[Resource[IO, BaseDataApplicationL0Service[IO]]] =
    makeL0Service.asResource.some
}
