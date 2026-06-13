#if os(iOS)
import Foundation
import CryptoKit

/// Orchestrates group symmetric key generation, per-member sealing, and relay fan-out (M2 / WO-207).
actor GroupKeyDistributionService {
    private let keyManager: GroupKeyManager
    private let groupsAPI: any GroupsAPIClient
    private let encryption: KinnamiEncryption
    private let secureEnclave: SecureEnclaveManager

    init(
        keyManager: GroupKeyManager,
        groupsAPI: any GroupsAPIClient,
        encryption: KinnamiEncryption,
        secureEnclave: SecureEnclaveManager
    ) {
        self.keyManager = keyManager
        self.groupsAPI = groupsAPI
        self.encryption = encryption
        self.secureEnclave = secureEnclave
    }

    /// Admin: rotate key and push sealed packages to all members.
    func rotateAndDistribute(
        groupId: String,
        members: [GroupKeyManager.MemberTarget],
        distributedBy: String
    ) async throws -> GroupKeyDistributeResponse {
        let (info, packages) = try await keyManager.rotateAndDistribute(
            groupId: groupId,
            members: members,
            encryption: encryption
        )
        return try await groupsAPI.distributeKeys(
            groupId: groupId,
            version: info.version,
            distributedBy: distributedBy,
            packages: packages
        )
    }

    /// Member: decrypt inbound group key signal and store locally.
    func acceptInboundKey(
        groupId: String,
        version: Int,
        encryptedPackage: Data
    ) async throws {
        let privateKeyData = try await secureEnclave.getPublicKey(id: "messaging-key-agreement")
        let privateKey = try P256.KeyAgreement.PrivateKey(rawRepresentation: privateKeyData)
        let keyData = try await keyManager.decryptKeyPackage(
            encryptedPackage,
            ourPrivateKey: privateKey,
            encryption: encryption
        )
        await keyManager.storeReceivedKey(groupId: groupId, version: version, keyData: keyData)
    }

    /// Admin: rotate and distribute after membership change when API returns requires_rekey.
    func rekeyForMembers(
        groupId: String,
        memberDIDs: [String],
        distributedBy: String,
        identityResolve: IdentityResolveClient
    ) async throws -> GroupKeyDistributeResponse {
        let targets = try await GroupMemberResolver.memberTargets(
            peerDIDs: memberDIDs,
            currentDID: distributedBy,
            identityResolve: identityResolve
        )
        guard !targets.isEmpty else {
            throw GroupKeyError.keyRotationFailed
        }
        return try await rotateAndDistribute(
            groupId: groupId,
            members: targets,
            distributedBy: distributedBy
        )
    }
}
#endif
