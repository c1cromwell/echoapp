package com.echo.shared_data.serializers

import com.echo.shared_data.types.{DataLayerCalculatedState, DataLayerOnChainState, DataLayerUpdate}
import io.circe.Encoder
import io.circe.syntax.EncoderOps
import io.constellationnetwork.currency.dataApplication.DataUpdate
import io.constellationnetwork.currency.dataApplication.dataApplication.DataApplicationBlock
import io.constellationnetwork.security.signature.Signed

import java.nio.charset.StandardCharsets

/** JSON UTF-8 serialization for Data L1 data-application payloads (Tessellation 4.x). */
object DataSerializers {

  private def serialize[A: Encoder](value: A): Array[Byte] =
    value.asJson.deepDropNullValues.noSpaces.getBytes(StandardCharsets.UTF_8)

  def serializeUpdate(update: DataLayerUpdate): Array[Byte] =
    serialize[DataLayerUpdate](update)

  def serializeState(state: DataLayerOnChainState): Array[Byte] =
    serialize[DataLayerOnChainState](state)

  def serializeCalculatedState(state: DataLayerCalculatedState): Array[Byte] =
    serialize[DataLayerCalculatedState](state)

  def serializeBlock(block: Signed[DataApplicationBlock])(implicit e: Encoder[DataUpdate]): Array[Byte] =
    serialize[Signed[DataApplicationBlock]](block)
}
