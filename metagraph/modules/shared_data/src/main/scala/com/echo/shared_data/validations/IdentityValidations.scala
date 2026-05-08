package com.echo.shared_data.validations

import com.echo.shared_data.types._

/**
 * Validation logic for the Constellation Identity Metagraph.
 *
 * All validators are pure: they take an update plus the registered
 * authorized-sender DID and return Right(()) on success or Left(reason)
 * on rejection. Authorized-sender enforcement is the L1 application's
 * job — we accept the sender as a parameter so tests stay deterministic.
 *
 * Phase 1 trust model: the Identity Service DID (registered at deploy
 * time via `IDENTITY_SERVICE_DID` env var) is the only authorized sender.
 * Phase 4+ this expands to per-org delegated issuers via OrgRoleCredentials.
 */
object IdentityValidations {

  // -- helpers ------------------------------------------------------------

  private val DidKeyPrefix     = "did:key:z"
  private val MaxFieldLen      = 4096
  private val MaxRoleLen       = 64
  private val MaxCredTypeLen   = 256
  private val MaxSchemaVersion = 32

  private def isHexLower(s: String): Boolean =
    s.nonEmpty && s.forall(c => c.isDigit || ('a' to 'f').contains(c))

  private def isHex(s: String): Boolean =
    s.nonEmpty && s.forall(c => c.isDigit || ('a' to 'f').contains(c.toLower))

  private def nonEmpty(field: String, value: String): Either[String, Unit] =
    Either.cond(value != null && value.nonEmpty, (), s"$field must not be empty")

  private def boundedLen(field: String, value: String, max: Int): Either[String, Unit] =
    Either.cond(value.length <= max, (), s"$field exceeds maximum length $max (got ${value.length})")

  private def isDidKey(s: String): Boolean =
    s != null && s.startsWith(DidKeyPrefix) && s.length > DidKeyPrefix.length

  private def validateDidKey(field: String, did: String): Either[String, Unit] =
    Either.cond(isDidKey(did), (), s"$field must be a did:key (got '$did')")

  private def requireAuthorizedSender(actualSender: String, authorized: String): Either[String, Unit] =
    Either.cond(
      actualSender != null && actualSender == authorized,
      (),
      s"Sender '$actualSender' is not the authorized Identity Service DID"
    )

  // -- validators ---------------------------------------------------------

  /**
   * VC Issuance Record validator.
   * Phase 1: the transaction sender and `issuerDID` must both match the
   * registered Identity Service DID. Subject must be did:key. Issuance
   * time must be in the past or near-future (small clock skew allowed).
   */
  def validateVCIssuance(
    update:           VCIssuanceUpdate,
    sender:           String,
    authorizedSender: String,
    nowMillis:        Long
  ): Either[String, Unit] =
    for {
      _ <- requireAuthorizedSender(sender, authorizedSender)
      _ <- nonEmpty("credentialId", update.credentialId)
      _ <- boundedLen("credentialId", update.credentialId, MaxFieldLen)
      _ <- validateDidKey("subjectDID", update.subjectDID)
      _ <- validateDidKey("issuerDID", update.issuerDID)
      _ <- Either.cond(
        update.issuerDID == authorizedSender,
        (),
        "issuerDID must equal the Identity Service DID in Phase 1"
      )
      _ <- nonEmpty("credentialType", update.credentialType)
      _ <- boundedLen("credentialType", update.credentialType, MaxCredTypeLen)
      _ <- nonEmpty("schemaVersion", update.schemaVersion)
      _ <- boundedLen("schemaVersion", update.schemaVersion, MaxSchemaVersion)
      _ <- Either.cond(update.issuedAt > 0, (), "issuedAt must be positive epoch millis")
      // allow up to 5 minutes of clock skew
      _ <- Either.cond(
        update.issuedAt <= nowMillis + 300000L,
        (),
        s"issuedAt ${update.issuedAt} is too far in the future (now=$nowMillis)"
      )
    } yield ()

  /**
   * Trust Tier Commitment validator.
   * Enforces the H(tier || nonce) invariant: the commitment field is a
   * 32-byte SHA-256 (64 lowercase hex chars). The raw tier/nonce preimage
   * is NOT submitted on-chain — only the digest. Validators reject any
   * submission whose commitment is not 32 bytes hex.
   */
  def validateTrustTierCommitment(
    update:           TrustTierCommitmentUpdate,
    sender:           String,
    authorizedSender: String,
    nowMillis:        Long
  ): Either[String, Unit] =
    for {
      _ <- requireAuthorizedSender(sender, authorizedSender)
      _ <- validateDidKey("subjectDID", update.subjectDID)
      _ <- Either.cond(
        update.commitment != null && update.commitment.length == 64,
        (),
        s"commitment must be 64 hex chars (SHA-256), got ${Option(update.commitment).map(_.length).getOrElse(0)}"
      )
      _ <- Either.cond(isHexLower(update.commitment), (), "commitment must be lowercase hex (a-f, 0-9)")
      _ <- Either.cond(update.anchoredAt > 0, (), "anchoredAt must be positive epoch millis")
      _ <- Either.cond(
        update.anchoredAt <= nowMillis + 300000L,
        (),
        s"anchoredAt ${update.anchoredAt} is too far in the future"
      )
    } yield ()

  /**
   * StatusList2021 batch update validator.
   * The bit vector must be exactly 131,072 bits (32,768 hex chars per
   * W3C StatusList2021). The sequence number must be monotonically
   * increasing — the L1 application supplies the previously stored
   * sequence so this validator can enforce the bound.
   */
  def validateStatusList2021(
    update:           StatusList2021BatchUpdate,
    sender:           String,
    authorizedSender: String,
    previousSequence: Long,
    nowMillis:        Long
  ): Either[String, Unit] =
    for {
      _ <- requireAuthorizedSender(sender, authorizedSender)
      _ <- validateDidKey("issuerOrgDID", update.issuerOrgDID)
      _ <- Either.cond(
        update.bitVector != null && update.bitVector.length == StatusList2021Vector.ExpectedHexLength,
        (),
        s"bitVector must be exactly ${StatusList2021Vector.ExpectedHexLength} hex chars " +
          s"(${StatusList2021Vector.ExpectedBitLength} bits), got ${Option(update.bitVector).map(_.length).getOrElse(0)}"
      )
      _ <- Either.cond(isHex(update.bitVector), (), "bitVector must be hex-encoded")
      _ <- Either.cond(
        update.sequence > previousSequence,
        (),
        s"sequence ${update.sequence} must be greater than previous $previousSequence"
      )
      _ <- Either.cond(update.publishedAt > 0, (), "publishedAt must be positive epoch millis")
      _ <- Either.cond(
        update.publishedAt <= nowMillis + 300000L,
        (),
        s"publishedAt ${update.publishedAt} is too far in the future"
      )
    } yield ()

  /**
   * EchoOrgRoleCredential validator.
   * Enforces did:key DIDs, a known role, and a future expiry. The
   * issuerOrgDID must equal the authorized Phase-1 issuer (Identity
   * Service DID). Phase 4+: this expands to a delegated-issuer registry
   * lookup so each org can issue its own role credentials.
   */
  def validateEchoOrgRoleCredential(
    update:           EchoOrgRoleCredentialUpdate,
    sender:           String,
    authorizedSender: String,
    nowMillis:        Long
  ): Either[String, Unit] =
    for {
      _ <- requireAuthorizedSender(sender, authorizedSender)
      _ <- nonEmpty("credentialId", update.credentialId)
      _ <- boundedLen("credentialId", update.credentialId, MaxFieldLen)
      _ <- validateDidKey("issuerOrgDID", update.issuerOrgDID)
      _ <- validateDidKey("memberDID", update.memberDID)
      _ <- nonEmpty("role", update.role)
      _ <- boundedLen("role", update.role, MaxRoleLen)
      _ <- Either.cond(
        OrgRole.All.contains(update.role),
        (),
        s"role '${update.role}' is not a recognized org role (allowed: ${OrgRole.All.toSeq.sorted.mkString(", ")})"
      )
      _ <- Either.cond(update.issuedAt > 0, (), "issuedAt must be positive epoch millis")
      _ <- Either.cond(
        update.expiry > update.issuedAt,
        (),
        s"expiry ${update.expiry} must be after issuedAt ${update.issuedAt}"
      )
      _ <- Either.cond(
        update.expiry > nowMillis,
        (),
        s"expiry ${update.expiry} is already in the past (now=$nowMillis)"
      )
    } yield ()

  /**
   * Device key registration — multi-device mapping for one subject DID.
   * Phase 1: only the Identity Service DID may anchor device rows.
   */
  def validateDeviceKeyRegistration(
    update:           DeviceKeyRegistrationUpdate,
    sender:           String,
    authorizedSender: String,
    nowMillis:        Long
  ): Either[String, Unit] = {
    val MaxDeviceLabelLen = 128
    val PubKeyHexLen      = 130 // 04 + 64-byte point
    for {
      _ <- requireAuthorizedSender(sender, authorizedSender)
      _ <- validateDidKey("subjectDID", update.subjectDID)
      _ <- nonEmpty("publicKeyHex", update.publicKeyHex)
      _ <- Either.cond(
        update.publicKeyHex != null && update.publicKeyHex.length == PubKeyHexLen,
        (),
        s"publicKeyHex must be SEC1 uncompressed P-256 hex ($PubKeyHexLen lowercase chars)"
      )
      _ <- Either.cond(
        update.publicKeyHex.startsWith("04"),
        (),
        "publicKeyHex must start with 04 (uncompressed point)"
      )
      _ <- Either.cond(isHexLower(update.publicKeyHex), (), "publicKeyHex must be lowercase hex (a-f, 0-9)")
      _ <- nonEmpty("deviceLabel", update.deviceLabel)
      _ <- boundedLen("deviceLabel", update.deviceLabel, MaxDeviceLabelLen)
      _ <- Either.cond(update.addedAt > 0, (), "addedAt must be positive epoch millis")
      _ <- Either.cond(
        update.addedAt <= nowMillis + 300000L,
        (),
        s"addedAt ${update.addedAt} is too far in the future (now=$nowMillis)"
      )
    } yield ()
  }
}
