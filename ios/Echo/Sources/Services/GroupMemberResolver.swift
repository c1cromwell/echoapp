#if os(iOS)
import Foundation
import CryptoKit

/// Resolves peer DIDs to `GroupKeyManager.MemberTarget` for key distribution (M2).
enum GroupMemberResolver {
    static func memberTargets(
        peerDIDs: [String],
        currentDID: String,
        identityResolve: IdentityResolveClient
    ) async throws -> [GroupKeyManager.MemberTarget] {
        var seen = Set<String>()
        var targets: [GroupKeyManager.MemberTarget] = []

        let allDIDs = peerDIDs + [currentDID]
        for did in allDIDs {
            let trimmed = did.trimmingCharacters(in: .whitespacesAndNewlines)
            guard !trimmed.isEmpty, !seen.contains(trimmed) else { continue }
            seen.insert(trimmed)

            let hex: String
            if trimmed == currentDID {
                let privateKey = try await TextMessageCrypto.loadAgreementPrivateKey()
                hex = privateKey.publicKey.rawRepresentation.hexEncodedString()
            } else {
                hex = try await identityResolve.primaryPublicKeyHex(peerDID: trimmed)
            }
            let pubData = try TextMessageCrypto.dataFromPublicKeyHex(hex)
            targets.append(.init(did: trimmed, keyAgreementPublicKey: pubData))
        }
        return targets
    }
}

private extension Data {
    func hexEncodedString() -> String {
        map { String(format: "%02x", $0) }.joined()
    }
}
#endif
