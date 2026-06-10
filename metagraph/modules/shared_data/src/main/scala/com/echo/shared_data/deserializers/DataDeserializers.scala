package com.echo.shared_data.deserializers

import com.echo.shared_data.types.{DataLayerCalculatedState, DataLayerOnChainState, DataLayerUpdate}
import io.circe.Decoder
import io.circe.parser.decode
import io.constellationnetwork.currency.dataApplication.DataUpdate
import io.constellationnetwork.currency.dataApplication.dataApplication.DataApplicationBlock
import io.constellationnetwork.security.signature.Signed

object DataDeserializers {

  def deserializeUpdate(bytes: Array[Byte]): Either[Throwable, DataLayerUpdate] =
    decode[DataLayerUpdate](new String(bytes, java.nio.charset.StandardCharsets.UTF_8))

  def deserializeState(bytes: Array[Byte]): Either[Throwable, DataLayerOnChainState] =
    decode[DataLayerOnChainState](new String(bytes, java.nio.charset.StandardCharsets.UTF_8))

  def deserializeCalculatedState(bytes: Array[Byte]): Either[Throwable, DataLayerCalculatedState] =
    decode[DataLayerCalculatedState](new String(bytes, java.nio.charset.StandardCharsets.UTF_8))

  def deserializeBlock(bytes: Array[Byte])(implicit d: Decoder[DataUpdate]): Either[Throwable, Signed[DataApplicationBlock]] =
    decode[Signed[DataApplicationBlock]](new String(bytes, java.nio.charset.StandardCharsets.UTF_8))
}
