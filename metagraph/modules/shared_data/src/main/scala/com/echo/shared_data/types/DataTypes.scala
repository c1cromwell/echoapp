package com.echo.shared_data.types

import io.circe.{Decoder, Encoder}
import io.circe.generic.semiauto._
import io.constellationnetwork.currency.dataApplication.{DataCalculatedState, DataOnChainState}

/**
 * On-chain state for the Data L1 data application (Merkle roots + trust commitments).
 */
case class DataLayerOnChainState(
  merkleRoots:      Map[String, MerkleRoot],
  trustCommitments: Map[String, TrustCommitment]
) extends DataOnChainState

object DataLayerOnChainState {
  val empty: DataLayerOnChainState = DataLayerOnChainState(Map.empty, Map.empty)
  implicit val encoder: Encoder[DataLayerOnChainState] = deriveEncoder
  implicit val decoder: Decoder[DataLayerOnChainState] = deriveDecoder
}

case class DataLayerCalculatedState(
  totalMerkleRoots:      Long,
  totalTrustCommitments: Long
) extends DataCalculatedState

object DataLayerCalculatedState {
  val empty: DataLayerCalculatedState = DataLayerCalculatedState(0L, 0L)
  implicit val encoder: Encoder[DataLayerCalculatedState] = deriveEncoder
  implicit val decoder: Decoder[DataLayerCalculatedState] = deriveDecoder
}

/** JSON view returned by GET /data-application/merkle-roots/{root} (WO-230 Step 5). */
case class MerkleRootStatus(
  root:      String,
  leafCount: Int,
  finalized: Boolean
)

object MerkleRootStatus {
  implicit val encoder: Encoder[MerkleRootStatus] = deriveEncoder
  implicit val decoder: Decoder[MerkleRootStatus] = deriveDecoder
}
