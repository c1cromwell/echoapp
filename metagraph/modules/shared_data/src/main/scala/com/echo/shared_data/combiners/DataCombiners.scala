package com.echo.shared_data.combiners

import com.echo.shared_data.types._
import io.constellationnetwork.currency.dataApplication.DataState
import io.constellationnetwork.security.signature.Signed

object DataCombiners {

  def combineUpdate(
    signed: Signed[DataLayerUpdate],
    state:  DataState[DataLayerOnChainState, DataLayerCalculatedState],
    nowMs:  Long = System.currentTimeMillis()
  ): DataState[DataLayerOnChainState, DataLayerCalculatedState] = {
    val onChain = foldUpdate(signed.value, state.onChain, nowMs)
    val calc    = recalculate(onChain)
    DataState(onChain, calc)
  }

  private def foldUpdate(
    update: DataLayerUpdate,
    state:  DataLayerOnChainState,
    nowMs:  Long
  ): DataLayerOnChainState =
    update match {
      case u: MerkleRootUpdate =>
        val rec = MerkleRoot(
          txId      = "",
          senderDid = "",
          root      = u.root,
          leafCount = u.leafCount,
          createdAt = nowMs
        )
        state.copy(merkleRoots = state.merkleRoots + (u.root -> rec))

      case u: TrustCommitmentUpdate =>
        val key = u.commitment
        val rec = TrustCommitment(
          txId       = "",
          senderDid  = "",
          commitment = u.commitment,
          epoch      = u.epoch,
          createdAt  = nowMs
        )
        state.copy(trustCommitments = state.trustCommitments + (key -> rec))
    }

  private def recalculate(onChain: DataLayerOnChainState): DataLayerCalculatedState =
    DataLayerCalculatedState(
      totalMerkleRoots      = onChain.merkleRoots.size.toLong,
      totalTrustCommitments = onChain.trustCommitments.size.toLong
    )
}
