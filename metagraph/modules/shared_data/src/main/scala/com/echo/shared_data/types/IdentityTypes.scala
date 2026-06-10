package com.echo.shared_data.types

import io.circe.{Decoder, DecodingFailure, Encoder, HCursor}
import io.circe.generic.semiauto._
import io.constellationnetwork.currency.dataApplication.{DataCalculatedState, DataOnChainState, DataUpdate}

/**
 * On-chain state persisted in Identity Metagraph snapshots.
 *
 * The Identity Metagraph is a dedicated L1 (separate from Currency L1 and
 * Data L1) that anchors W3C VC 2.0 issuance records, trust tier commitments,
 * StatusList2021 revocation bit vectors, and EchoOrgRoleCredential org
 * membership metadata, and multi-device public-key registrations.
 *
 * Phase 1: project-operated validators only. Phase 4+: community validators.
 */
case class IdentityOnChainState(
  vcIssuances:        Map[String, VCIssuanceRecord],     // keyed by credentialId
  trustTierAnchors:   Map[String, TrustTierAnchor],      // keyed by subjectDID
  revocationLists:    Map[String, StatusList2021Vector], // keyed by issuerOrgDID
  orgRoleCredentials: Map[String, EchoOrgRoleCredential], // keyed by credentialId
  deviceKeys:         Map[String, DeviceKeyRecord],      // keyed by subjectDID#publicKeyHex
  usernames:          Map[String, UsernameRecord]        // keyed by lowercased username (public Data L1 index)
) extends DataOnChainState

object IdentityOnChainState {
  val empty: IdentityOnChainState =
    IdentityOnChainState(Map.empty, Map.empty, Map.empty, Map.empty, Map.empty, Map.empty)
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
) extends DataCalculatedState

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

/**
 * Device key anchored for a subject DID (multi-device did:key mapping).
 * Rows are emitted when the Identity Service registers an additional device
 * public key on-chain (Identity Service is the Phase-1 authorized sender).
 */
case class DeviceKeyRecord(
  subjectDID:   String,
  publicKeyHex: String,
  deviceLabel:  String,
  addedAt:      Long
)

object DeviceKeyRecord {
  implicit val encoder: Encoder[DeviceKeyRecord] = deriveEncoder
  implicit val decoder: Decoder[DeviceKeyRecord]  = deriveDecoder
}

/**
 * Public `@username` -> DID binding anchored on the Identity Metagraph.
 *
 * Usernames are a PUBLIC Data L1 index (privacy class T7): anyone can resolve
 * `@username` to its owning `did:key` without ECHO's involvement, and the owner
 * can prove the binding is theirs. The backend Postgres `users` table is a
 * read-through cache of this index, not the source of truth.
 *
 * Keyed in [[IdentityOnChainState.usernames]] by the lowercased username so
 * uniqueness is case-insensitive.
 */
case class UsernameRecord(
  username:     String, // as registered (display case preserved)
  subjectDID:   String, // did:key of the owner
  registeredAt: Long    // epoch millis
)

object UsernameRecord {
  implicit val encoder: Encoder[UsernameRecord] = deriveEncoder
  implicit val decoder: Decoder[UsernameRecord] = deriveDecoder
}

// --- Update messages submitted to Identity L1 endpoints ---

sealed trait IdentityUpdate extends DataUpdate

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

/**
 * Additional device key registration for an existing subject `did:key`.
 *
 * JSON wire format (Identity L1 `POST /transactions` body, Phase 1): a flat
 * object with these four fields — discriminated from other [[IdentityUpdate]]
 * variants by the presence of `publicKeyHex`, `deviceLabel`, and `addedAt`
 * (mirrors `internal/metagraph.DeviceKeyRegistrationUpdate` in Go).
 *
 * Only the Identity Service DID may submit in Phase 1.
 */
case class DeviceKeyRegistrationUpdate(
  subjectDID:   String,
  publicKeyHex: String,
  deviceLabel:  String,
  addedAt:      Long
) extends IdentityUpdate

object DeviceKeyRegistrationUpdate {
  implicit val encoder: Encoder[DeviceKeyRegistrationUpdate] = deriveEncoder
  implicit val decoder: Decoder[DeviceKeyRegistrationUpdate] = deriveDecoder
}

/**
 * Public `@username` registration for a subject `did:key` (decentralization D1).
 *
 * JSON wire format (Identity L1 `POST /transactions` body): a flat object with
 * `subjectDID`, `username`, `registeredAt` — discriminated from other
 * [[IdentityUpdate]] variants by the presence of `username` (mirrors
 * `internal/metagraph.UsernameRegistrationUpdate` in Go).
 *
 * Only the Identity Service DID may submit in Phase 1.
 */
case class UsernameRegistrationUpdate(
  subjectDID:   String,
  username:     String,
  registeredAt: Long
) extends IdentityUpdate

object UsernameRegistrationUpdate {
  implicit val encoder: Encoder[UsernameRegistrationUpdate] = deriveEncoder
  implicit val decoder: Decoder[UsernameRegistrationUpdate] = deriveDecoder
}

object IdentityUpdate {

  private val vcDec           = deriveDecoder[VCIssuanceUpdate]
  private val trustDec        = deriveDecoder[TrustTierCommitmentUpdate]
  private val statusDec       = deriveDecoder[StatusList2021BatchUpdate]
  private val orgRoleDec      = deriveDecoder[EchoOrgRoleCredentialUpdate]
  private val deviceKeyDec    = DeviceKeyRegistrationUpdate.decoder
  private val usernameDec     = UsernameRegistrationUpdate.decoder

  /** Classify variant from JSON keys (no explicit `type` discriminator). */
  implicit val decoder: Decoder[IdentityUpdate] = Decoder.instance { c: HCursor =>
    val keys      = c.keys.map(_.toSet).getOrElse(Set.empty)
    val bitVec    = keys.contains("bitVector")
    val commitment = keys.contains("commitment")
    val device    =
      keys.contains("publicKeyHex") && keys.contains("deviceLabel") && keys.contains(
        "addedAt"
      )
    val username  = keys.contains("username")
    val vc        = keys.contains("schemaVersion")
    val orgRole   = keys.contains("role") && keys.contains("memberDID")

    if (bitVec) statusDec(c)
    else if (commitment) trustDec(c)
    else if (device) deviceKeyDec(c)
    else if (username) usernameDec(c)
    else if (vc) vcDec(c)
    else if (orgRole) orgRoleDec(c)
    else Left(DecodingFailure("Unrecognized IdentityUpdate JSON shape", c.history))
  }

  implicit val encoder: Encoder[IdentityUpdate] = Encoder.instance {
    case u: VCIssuanceUpdate              => deriveEncoder[VCIssuanceUpdate].apply(u)
    case u: TrustTierCommitmentUpdate     => deriveEncoder[TrustTierCommitmentUpdate].apply(u)
    case u: StatusList2021BatchUpdate     => deriveEncoder[StatusList2021BatchUpdate].apply(u)
    case u: EchoOrgRoleCredentialUpdate   => deriveEncoder[EchoOrgRoleCredentialUpdate].apply(u)
    case u: DeviceKeyRegistrationUpdate   => DeviceKeyRegistrationUpdate.encoder.apply(u)
    case u: UsernameRegistrationUpdate    => UsernameRegistrationUpdate.encoder.apply(u)
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
