package com.echo.shared_data.types

import io.circe.{Decoder, DecodingFailure, Encoder, HCursor}
import io.circe.generic.semiauto._
import io.constellationnetwork.currency.dataApplication.DataUpdate

/**
 * On-chain state persisted in metagraph snapshots.
 * Mirrors Go backend types in internal/metagraph/transactions.go.
 */
case class EchoOnChainState(
  tokenLocks:       Map[String, TokenLockState],
  delegations:      Map[String, DelegationState],
  rewardClaims:     Map[String, RewardClaimState],
  trustCommitments: Map[String, TrustCommitment],
  merkleRoots:      Map[String, MerkleRoot]
)

object EchoOnChainState {
  val empty: EchoOnChainState = EchoOnChainState(Map.empty, Map.empty, Map.empty, Map.empty, Map.empty)
  implicit val encoder: Encoder[EchoOnChainState] = deriveEncoder
  implicit val decoder: Decoder[EchoOnChainState] = deriveDecoder
}

/**
 * Calculated state derived from on-chain state for efficient querying.
 */
case class EchoCalculatedState(
  totalStaked:      Long,
  activeValidators: Set[String],
  dailyRewardCaps:  Map[String, Long]
)

object EchoCalculatedState {
  val empty: EchoCalculatedState = EchoCalculatedState(0L, Set.empty, Map.empty)
  implicit val encoder: Encoder[EchoCalculatedState] = deriveEncoder
  implicit val decoder: Decoder[EchoCalculatedState] = deriveDecoder
}

// --- Currency L1 types ---

case class TokenLockState(
  txId:         String,
  senderDid:    String,
  amount:       Long,        // in datum (1 ECHO = 1e8 datum)
  tierName:     String,
  lockDays:     Int,
  unlocksAt:    Long,        // epoch millis
  minimumStake: Long
)

object TokenLockState {
  implicit val encoder: Encoder[TokenLockState] = deriveEncoder
  implicit val decoder: Decoder[TokenLockState] = deriveDecoder
}

case class DelegationState(
  txId:          String,
  senderDid:     String,
  tokenLockTxId: String,
  validatorDid:  String,
  delegatedStake: Long
)

object DelegationState {
  implicit val encoder: Encoder[DelegationState] = deriveEncoder
  implicit val decoder: Decoder[DelegationState] = deriveDecoder
}

case class RewardClaimState(
  txId:       String,
  claimerDid: String,
  amount:     Long,
  tier:       String,
  claimedAt:  Long
)

object RewardClaimState {
  implicit val encoder: Encoder[RewardClaimState] = deriveEncoder
  implicit val decoder: Decoder[RewardClaimState] = deriveDecoder
}

// --- Data L1 types ---

case class TrustCommitment(
  txId:       String,
  senderDid:  String,
  commitment: String,   // H(score || nonce), hex-encoded
  epoch:      Long,
  createdAt:  Long
)

object TrustCommitment {
  implicit val encoder: Encoder[TrustCommitment] = deriveEncoder
  implicit val decoder: Decoder[TrustCommitment] = deriveDecoder
}

case class MerkleRoot(
  txId:      String,
  senderDid: String,
  root:      String,    // hex-encoded SHA-256 Merkle root
  leafCount: Int,
  createdAt: Long
)

object MerkleRoot {
  implicit val encoder: Encoder[MerkleRoot] = deriveEncoder
  implicit val decoder: Decoder[MerkleRoot] = deriveDecoder
}

// --- Update types submitted to L1 endpoints ---

/** Updates accepted by the Echo Data L1 data application (Tessellation POST /data). */
sealed trait DataLayerUpdate extends DataUpdate

sealed trait EchoUpdate

case class TokenLockUpdate(
  amount:   Long,
  tierName: String,
  lockDays: Int
) extends EchoUpdate

case class StakeDelegationUpdate(
  tokenLockTxId: String,
  validatorDid:  String,
  amount:        Long
) extends EchoUpdate

case class WithdrawLockUpdate(
  tokenLockTxId: String,
  amount:        Long
) extends EchoUpdate

case class RewardClaimUpdate(
  amount: Long,
  tier:   String
) extends EchoUpdate

case class TrustCommitmentUpdate(
  commitment: String,
  epoch:      Long
) extends EchoUpdate with DataLayerUpdate

case class MerkleRootUpdate(
  root:      String,
  leafCount: Int
) extends EchoUpdate with DataLayerUpdate

/** Post-genesis minting is forbidden (WO-214 fixed supply). */
case class MintUpdate(
  amount: Long,
  pool:   String
) extends EchoUpdate

/** Founder departure revocation credits Future Team pool (WO-225). */
case class FounderRevocationUpdate(
  targetFounderDid: String,
  amount:           Long,
  revokerDids:      Seq[String],
  destinationPool:  String
) extends EchoUpdate

object DataLayerUpdate {

  private val merkleDec = deriveDecoder[MerkleRootUpdate]
  private val trustDec  = deriveDecoder[TrustCommitmentUpdate]

  implicit val decoder: Decoder[DataLayerUpdate] = Decoder.instance { c: HCursor =>
    val keys = c.keys.map(_.toSet).getOrElse(Set.empty)
    if (keys.contains("root")) merkleDec(c)
    else if (keys.contains("commitment")) trustDec(c)
    else Left(DecodingFailure("Unrecognized DataLayerUpdate JSON shape", c.history))
  }

  implicit val encoder: Encoder[DataLayerUpdate] = Encoder.instance {
    case u: MerkleRootUpdate      => deriveEncoder[MerkleRootUpdate].apply(u)
    case u: TrustCommitmentUpdate => deriveEncoder[TrustCommitmentUpdate].apply(u)
  }
}

object EchoUpdate {
  implicit val encoder: Encoder[EchoUpdate] = Encoder.instance {
    case u: TokenLockUpdate        => deriveEncoder[TokenLockUpdate].apply(u)
    case u: StakeDelegationUpdate  => deriveEncoder[StakeDelegationUpdate].apply(u)
    case u: WithdrawLockUpdate     => deriveEncoder[WithdrawLockUpdate].apply(u)
    case u: RewardClaimUpdate      => deriveEncoder[RewardClaimUpdate].apply(u)
    case u: TrustCommitmentUpdate  => deriveEncoder[TrustCommitmentUpdate].apply(u)
    case u: MerkleRootUpdate       => deriveEncoder[MerkleRootUpdate].apply(u)
    case u: MintUpdate             => deriveEncoder[MintUpdate].apply(u)
    case u: FounderRevocationUpdate => deriveEncoder[FounderRevocationUpdate].apply(u)
  }
}
