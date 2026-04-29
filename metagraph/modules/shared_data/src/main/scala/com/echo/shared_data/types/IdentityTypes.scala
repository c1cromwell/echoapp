package com.echo.shared_data.types

import io.circe.{Decoder, Encoder}
import io.circe.generic.semiauto._

/**
 * On-chain state persisted in Identity Metagraph snapshots.
 *
 * The Identity Metagraph is a dedicated L1 (separate from Currency L1 and
 * Data L1) that anchors W3C VC 2.0 issuance records, trust tier commitments,
 * StatusList2021 revocation bit vectors, and EchoOrgRoleCredential org
 * membership metadata.
 *
 * Phase 1: project-operated validators only. Phase 4+: community validators.
 */
case class IdentityOnChainState(
  vcIssuances:        Map[String, VCIssuanceRecord],     // keyed by credentialId
  trustTierAnchors:   Map[String, TrustTierAnchor],      // keyed by subjectDID
  revocationLists:    Map[String, StatusList2021Vector], // keyed by issuerOrgDID
  orgRoleCredentials: Map[String, EchoOrgRoleCredential] // keyed by credentialId
)

object IdentityOnChainState {
  val empty: IdentityOnChainState =
    IdentityOnChainState(Map.empty, Map.empty, Map.empty, Map.empty)
  implicit val encoder: Encoder[IdentityOnChainState] = deriveEncoder
  implicit val decoder: Decoder[IdentityOnChainState] = deriveDecoder
}

/**
 * Calculated state derived from on-chain identity state for efficient
 * querying (e.g. "how many tier-3 commitments has issuer X published this
 * epoch?").
 */
case class IdentityCalculatedState(
  totalCredentialsIssued: Long,
  revokedCredentialCount: Long,
  trustTierCounts:        Map[Int, Long] // tier -> count of subjects at that tier
)

object IdentityCalculatedState {
  val empty: IdentityCalculatedState =
    IdentityCalculatedState(0L, 0L, Map.empty)
  implicit val encoder: Encoder[IdentityCalculatedState] = deriveEncoder
  implicit val decoder: Decoder[IdentityCalculatedState] = deriveDecoder
}

// --- Persisted records (anchored on-chain) ---

/**
 * VC Issuance Record (W3C Verifiable Credentials 2.0).
 * Only the metadata of the credential is anchored. The full VC JWT lives
 * off-chain in the Identity Service / IPFS, signed by the issuer.
 */
case class VCIssuanceRecord(
  credentialId:   String, // unique, e.g. uuid:v4 or content-hash
  subjectDID:     String, // did:key of credential subject
  issuerDID:      String, // did:key of issuer (for Phase 1, this is the Identity Service)
  credentialType: String, // e.g. "TrustTierCredential", "EchoOrgRoleCredential", "KYCCredential"
  issuedAt:       Long,   // epoch millis
  schemaVersion:  String  // e.g. "v1.0.0"
)

object VCIssuanceRecord {
  implicit val encoder: Encoder[VCIssuanceRecord] = deriveEncoder
  implicit val decoder: Decoder[VCIssuanceRecord] = deriveDecoder
}

/**
 * Trust Tier Commitment anchored on the Identity Metagraph.
 *
 *   commitment = SHA-256(tier || nonce)
 *
 * The raw `tier` and `nonce` are revealed only when the holder needs to
 * prove their tier (selective disclosure). The on-chain commitment binds
 * the holder to a specific tier without revealing it to the public.
 */
case class TrustTierAnchor(
  subjectDID: String, // did:key of the holder
  commitment: String, // 64 hex chars (SHA-256 of tier||nonce)
  anchoredAt: Long    // epoch millis
)

object TrustTierAnchor {
  implicit val encoder: Encoder[TrustTierAnchor] = deriveEncoder
  implicit val decoder: Decoder[TrustTierAnchor] = deriveDecoder
}

/**
 * StatusList2021 revocation bit vector for an organization.
 *
 * Each VC issued by this org references a position in the bit vector.
 * Setting bit N to 1 revokes the credential at position N. Vectors are
 * fixed at 131,072 bits (16 KiB) per the W3C StatusList2021 spec, which
 * yields one list per org until ~131k credentials are issued.
 *
 * Bit vectors are stored hex-encoded (16,384 hex pairs = 32,768 hex chars).
 */
case class StatusList2021Vector(
  issuerOrgDID: String, // did:key of issuing org
  bitVector:    String, // hex-encoded 131,072-bit vector (32,768 hex chars)
  publishedAt:  Long,   // epoch millis of last batch publication
  sequence:     Long    // monotonic publication sequence (for clients to detect updates)
)

object StatusList2021Vector {
  /** W3C StatusList2021 fixed size: 131,072 bits = 16,384 bytes = 32,768 hex chars. */
  val ExpectedBitLength:    Int = 131072
  val ExpectedHexLength:    Int = ExpectedBitLength / 4 // 32768 hex chars
  val ExpectedByteLength:   Int = ExpectedBitLength / 8 // 16384 bytes
  implicit val encoder: Encoder[StatusList2021Vector] = deriveEncoder
  implicit val decoder: Decoder[StatusList2021Vector] = deriveDecoder
}

/**
 * EchoOrgRoleCredential — org membership VC metadata anchored on-chain.
 * Used to prove "DID X holds role Y in org Z, valid until expiry E".
 */
case class EchoOrgRoleCredential(
  credentialId: String,
  issuerOrgDID: String, // did:key of the issuing org
  memberDID:    String, // did:key of the member
  role:         String, // e.g. "owner", "admin", "moderator", "member"
  expiry:       Long,   // epoch millis
  issuedAt:     Long
)

object EchoOrgRoleCredential {
  implicit val encoder: Encoder[EchoOrgRoleCredential] = deriveEncoder
  implicit val decoder: Decoder[EchoOrgRoleCredential] = deriveDecoder
}

// --- Update messages submitted to Identity L1 endpoints ---

sealed trait IdentityUpdate

case class VCIssuanceUpdate(
  credentialId:   String,
  subjectDID:     String,
  issuerDID:      String,
  credentialType: String,
  issuedAt:       Long,
  schemaVersion:  String
) extends IdentityUpdate

case class TrustTierCommitmentUpdate(
  subjectDID: String,
  commitment: String,
  anchoredAt: Long
) extends IdentityUpdate

case class StatusList2021BatchUpdate(
  issuerOrgDID: String,
  bitVector:    String,
  publishedAt:  Long,
  sequence:     Long
) extends IdentityUpdate

case class EchoOrgRoleCredentialUpdate(
  credentialId: String,
  issuerOrgDID: String,
  memberDID:    String,
  role:         String,
  expiry:       Long,
  issuedAt:     Long
) extends IdentityUpdate

object IdentityUpdate {
  implicit val encoder: Encoder[IdentityUpdate] = Encoder.instance {
    case u: VCIssuanceUpdate            => deriveEncoder[VCIssuanceUpdate].apply(u)
    case u: TrustTierCommitmentUpdate   => deriveEncoder[TrustTierCommitmentUpdate].apply(u)
    case u: StatusList2021BatchUpdate   => deriveEncoder[StatusList2021BatchUpdate].apply(u)
    case u: EchoOrgRoleCredentialUpdate => deriveEncoder[EchoOrgRoleCredentialUpdate].apply(u)
  }
}

/**
 * Allowed values for `tier` inside the (off-chain) preimage of a
 * TrustTierCommitment. Used by the validator to bound the disclosed tier
 * when a holder later reveals it.
 */
object TrustTier {
  val Min: Int = 1
  val Max: Int = 5
  val All: Set[Int] = (Min to Max).toSet
}

/**
 * Allowed values for `role` in EchoOrgRoleCredential. Reject any update
 * carrying an unknown role to prevent typo-driven privilege escalation.
 */
object OrgRole {
  val All: Set[String] = Set("owner", "admin", "moderator", "member")
}
