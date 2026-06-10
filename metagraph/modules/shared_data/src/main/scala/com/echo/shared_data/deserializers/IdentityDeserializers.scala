package com.echo.shared_data.deserializers

import com.echo.shared_data.types.{IdentityCalculatedState, IdentityOnChainState, IdentityUpdate}
import io.circe.Decoder
import io.circe.parser.decode
import io.constellationnetwork.currency.dataApplication.DataUpdate
import io.constellationnetwork.currency.dataApplication.dataApplication.DataApplicationBlock
import io.constellationnetwork.security.signature.Signed

object IdentityDeserializers {

  def deserializeUpdate(bytes: Array[Byte]): Either[Throwable, IdentityUpdate] =
    decode[IdentityUpdate](new String(bytes, java.nio.charset.StandardCharsets.UTF_8))

  def deserializeState(bytes: Array[Byte]): Either[Throwable, IdentityOnChainState] =
    decode[IdentityOnChainState](new String(bytes, java.nio.charset.StandardCharsets.UTF_8))

  def deserializeCalculatedState(bytes: Array[Byte]): Either[Throwable, IdentityCalculatedState] =
    decode[IdentityCalculatedState](new String(bytes, java.nio.charset.StandardCharsets.UTF_8))

  def deserializeBlock(bytes: Array[Byte])(implicit d: Decoder[DataUpdate]): Either[Throwable, Signed[DataApplicationBlock]] =
    decode[Signed[DataApplicationBlock]](new String(bytes, java.nio.charset.StandardCharsets.UTF_8))
}
