#if os(iOS)
import Foundation

/// PQ-hybrid ratchet bootstrap handshake (WO-SX2).
enum PQHybridBootstrap {
    static var isPlatformSupported: Bool { true }

    static var isActive: Bool {
        PQHybridPreferences.usePQBootstrap && isPlatformSupported
    }

    static func outboundHybridBundle() async -> HybridPublicBundleWire? {
        guard isActive else { return nil }
        return try? await PQHybridKeyStore.shared.loadOrGenerateBundle()
    }

    /// Initiator encapsulates against a cached peer bundle and stores the derived ratchet root secret.
    static func encapsulateForPeer(
        peerDID: String,
        remoteBundle: HybridPublicBundleWire
    ) throws -> (ciphertext: HybridCiphertextWire, secret: Data) {
        let (wire, secret) = try PQHybridCrypto.encapsulate(remote: remoteBundle)
        cacheBootstrapSecret(secret, peerDID: peerDID)
        return (wire, secret)
    }

    /// Responder decapsulates an initiator ciphertext and stores the ratchet root secret.
    static func decapsulateFromPeer(
        peerDID: String,
        ciphertext: HybridCiphertextWire
    ) async throws {
        guard let keys = try await PQHybridKeyStore.shared.loadPrivateKeys() else {
            throw PQHybridCrypto.Error.invalidKeyMaterial
        }
        let secret = try PQHybridCrypto.decapsulate(ec: keys.ec, pq: keys.pq, ciphertext: ciphertext)
        cacheBootstrapSecret(secret, peerDID: peerDID)
    }

    static func cachedBootstrapSecret(peerDID: String) -> Data? {
        UserDefaults.standard.data(forKey: secretKey(peerDID))
    }

    static func cachePeerHybridBundle(peerDID: String, bundle: HybridPublicBundleWire?) {
        guard let bundle else {
            UserDefaults.standard.removeObject(forKey: bundleKey(peerDID))
            return
        }
        if let data = try? JSONEncoder().encode(bundle) {
            UserDefaults.standard.set(data, forKey: bundleKey(peerDID))
        }
    }

    static func cachedPeerHybridBundle(peerDID: String) -> HybridPublicBundleWire? {
        guard let data = UserDefaults.standard.data(forKey: bundleKey(peerDID)) else { return nil }
        return try? JSONDecoder().decode(HybridPublicBundleWire.self, from: data)
    }

    private static func cacheBootstrapSecret(_ secret: Data, peerDID: String) {
        UserDefaults.standard.set(secret, forKey: secretKey(peerDID))
    }

    private static func bundleKey(_ peerDID: String) -> String {
        "echo.pq_hybrid.peer.bundle." + peerDID
    }

    private static func secretKey(_ peerDID: String) -> String {
        "echo.pq_hybrid.peer.secret." + peerDID
    }
}
#endif
