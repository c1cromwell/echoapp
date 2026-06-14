#if os(iOS)
import Foundation

/// Encrypts media file bytes for relay upload (pairwise P-256 ECDH, same as sync).
actor MediaMessageCrypto {
    private let syncCrypto: DeviceSyncCrypto
    private let identityResolve: IdentityResolveClient
    private var peerKeyCache: [String: String] = [:]

    init(
        identityResolve: IdentityResolveClient,
        syncCrypto: DeviceSyncCrypto = DeviceSyncCrypto()
    ) {
        self.identityResolve = identityResolve
        self.syncCrypto = syncCrypto
    }

    func encryptFile(_ plaintext: Data, peerDID: String) async throws -> Data {
        let hex = try await cachedPeerKeyHex(peerDID: peerDID)
        let pub = try TextMessageCrypto.dataFromPublicKeyHex(hex)
        return try syncCrypto.wrap(plaintext: plaintext, recipientPublicKey: pub)
    }

    func decryptFile(_ ciphertext: Data) async throws -> Data {
        try await syncCrypto.unwrapWithLocalKey(ciphertext: ciphertext)
    }

    private func cachedPeerKeyHex(peerDID: String) async throws -> String {
        if let cached = peerKeyCache[peerDID] { return cached }
        let hex = try await identityResolve.primaryPublicKeyHex(peerDID: peerDID)
        peerKeyCache[peerDID] = hex
        return hex
    }
}
#endif
