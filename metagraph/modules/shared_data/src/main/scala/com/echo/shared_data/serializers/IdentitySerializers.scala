package com.echo.shared_data.serializers

import com.echo.shared_data.types.{IdentityCalculatedState, IdentityOnChainState, IdentityUpdate}
import io.circe.Encoder
import io.circe.syntax.EncoderOps
import io.constellationnetwork.currency.dataApplication.DataUpdate
import io.constellationnetwork.currency.dataApplication.dataApplication.DataApplicationBlock
import io.constellationnetwork.security.signature.Signed

import java.nio.charset.StandardCharsets

/** JSON UTF-8 serialization for Identity data-application payloads (Tessellation 4.x). */
object IdentitySerializers {

  private def serialize[A: Encoder](value: A): Array[Byte] =
    value.asJson.deepDropNullValues.noSpaces.getBytes(StandardCharsets.UTF_8)

  def serializeUpdate(update: IdentityUpdate): Array[Byte] =
    serialize[IdentityUpdate](update)

  def serializeState(state: IdentityOnChainState): Array[Byte] =
    serialize[IdentityOnChainState](state)

  def serializeCalculatedState(state: IdentityCalculatedState): Array[Byte] =
    serialize[IdentityCalculatedState](state)

  def serializeBlock(block: Signed[DataApplicationBlock])(implicit e: Encoder[DataUpdate]): Array[Byte] =
    serialize[Signed[DataApplicationBlock]](block)
}
