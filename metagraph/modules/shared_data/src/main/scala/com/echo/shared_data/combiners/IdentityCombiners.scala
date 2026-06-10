package com.echo.shared_data.combiners

import com.echo.shared_data.state.IdentityRevocationSequences
import com.echo.shared_data.types._
import io.constellationnetwork.currency.dataApplication.DataState
import io.constellationnetwork.security.signature.Signed

object IdentityCombiners {

  def combineUpdate(
    signed: Signed[IdentityUpdate],
    state:  DataState[IdentityOnChainState, IdentityCalculatedState]
  ): DataState[IdentityOnChainState, IdentityCalculatedState] = {
    val onChain = foldUpdate(signed.value, state.onChain)
    val calc    = recalculate(onChain)
    DataState(onChain, calc)
  }

  private def foldUpdate(update: IdentityUpdate, state: IdentityOnChainState): IdentityOnChainState =
    update match {
      case u: VCIssuanceUpdate =>
        val record = VCIssuanceRecord(
          credentialId   = u.credentialId,
          subjectDID     = u.subjectDID,
          issuerDID      = u.issuerDID,
          credentialType = u.credentialType,
          issuedAt       = u.issuedAt,
          schemaVersion  = u.schemaVersion
        )
        state.copy(vcIssuances = state.vcIssuances + (u.credentialId -> record))

      case u: TrustTierCommitmentUpdate =>
        val anchor = TrustTierAnchor(u.subjectDID, u.commitment, u.anchoredAt)
        state.copy(trustTierAnchors = state.trustTierAnchors + (u.subjectDID -> anchor))

      case u: StatusList2021BatchUpdate =>
        val vector = StatusList2021Vector(u.issuerOrgDID, u.bitVector, u.publishedAt, u.sequence)
        IdentityRevocationSequences.recordPublished(u.issuerOrgDID, u.sequence)
        state.copy(revocationLists = state.revocationLists + (u.issuerOrgDID -> vector))

      case u: EchoOrgRoleCredentialUpdate =>
        val cred = EchoOrgRoleCredential(
          credentialId = u.credentialId,
          issuerOrgDID = u.issuerOrgDID,
          memberDID    = u.memberDID,
          role         = u.role,
          expiry       = u.expiry,
          issuedAt     = u.issuedAt
        )
        state.copy(orgRoleCredentials = state.orgRoleCredentials + (u.credentialId -> cred))

      case u: DeviceKeyRegistrationUpdate =>
        val key = s"${u.subjectDID}#${u.publicKeyHex}"
        val rec = DeviceKeyRecord(u.subjectDID, u.publicKeyHex, u.deviceLabel, u.addedAt)
        state.copy(deviceKeys = state.deviceKeys + (key -> rec))

      case u: UsernameRegistrationUpdate =>
        val key  = u.username.toLowerCase
        val rec  = UsernameRecord(u.username, u.subjectDID, u.registeredAt)
        state.copy(usernames = state.usernames + (key -> rec))
    }

  private def recalculate(onChain: IdentityOnChainState): IdentityCalculatedState = {
    val revoked = onChain.revocationLists.values.map(_.bitVector.count(_ == '1')).sum.toLong
    IdentityCalculatedState(
      totalCredentialsIssued = onChain.vcIssuances.size.toLong,
      revokedCredentialCount   = revoked,
      trustTierCounts          = Map.empty
    )
  }
}
