package com.echo.shared_data.lifecycle

import cats.effect.{Ref, Sync}
import cats.syntax.all._
import com.echo.shared_data.serializers.DataSerializers
import com.echo.shared_data.types.DataLayerCalculatedState
import io.constellationnetwork.schema.SnapshotOrdinal
import io.constellationnetwork.security.hash.Hash

/** In-memory calculated-state store for Data L0 (Phase 1 dev). */
final class DataCalculatedStateService[F[_]: Sync](
  ref: Ref[F, (SnapshotOrdinal, DataLayerCalculatedState)]
) {

  def getCalculatedState: F[(SnapshotOrdinal, DataLayerCalculatedState)] =
    ref.get

  def setCalculatedState(ordinal: SnapshotOrdinal, state: DataLayerCalculatedState): F[Boolean] =
    ref.set((ordinal, state)).as(true)

  def hashCalculatedState(state: DataLayerCalculatedState): F[Hash] =
    Sync[F].delay(Hash.fromBytes(DataSerializers.serializeCalculatedState(state)))
}

object DataCalculatedStateService {
  def make[F[_]: Sync]: F[DataCalculatedStateService[F]] =
    Ref
      .of[F, (SnapshotOrdinal, DataLayerCalculatedState)](
        (SnapshotOrdinal.MinValue, DataLayerCalculatedState.empty)
      )
      .map(new DataCalculatedStateService[F](_))
}
