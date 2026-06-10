package com.echo.shared_data.lifecycle

import cats.effect.{Ref, Sync}
import cats.syntax.all._
import com.echo.shared_data.serializers.IdentitySerializers
import com.echo.shared_data.types.IdentityCalculatedState
import io.constellationnetwork.schema.SnapshotOrdinal
import io.constellationnetwork.security.hash.Hash

/** In-memory calculated-state store for Identity L0 (Phase 1 dev). */
final class IdentityCalculatedStateService[F[_]: Sync](
  ref: Ref[F, (SnapshotOrdinal, IdentityCalculatedState)]
) {

  def getCalculatedState: F[(SnapshotOrdinal, IdentityCalculatedState)] =
    ref.get

  def setCalculatedState(ordinal: SnapshotOrdinal, state: IdentityCalculatedState): F[Boolean] =
    ref.set((ordinal, state)).as(true)

  def hashCalculatedState(state: IdentityCalculatedState): F[Hash] =
    Sync[F].delay(Hash.fromBytes(IdentitySerializers.serializeCalculatedState(state)))
}

object IdentityCalculatedStateService {
  def make[F[_]: Sync]: F[IdentityCalculatedStateService[F]] =
    Ref.of[F, (SnapshotOrdinal, IdentityCalculatedState)](
      (SnapshotOrdinal.MinValue, IdentityCalculatedState.empty)
    ).map(new IdentityCalculatedStateService[F](_))
}
