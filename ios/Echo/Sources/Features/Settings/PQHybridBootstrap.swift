#if os(iOS)
import Foundation

/// PQ-hybrid ratchet bootstrap handshake hook (WO-SX2 finish).
///
/// Full ML-KEM encapsulation/decapsulation lives in Go (`internal/crypto/pqhybrid.go`).
/// iOS defers crypto until CryptoKit exposes ML-KEM; this type gates preference + payload wiring.
enum PQHybridBootstrap {
    /// `false` until the platform SDK ships ML-KEM-768.
    static var isPlatformSupported: Bool { false }

    /// User wants PQ bootstrap and the device can perform it.
    static var isActive: Bool {
        PQHybridPreferences.usePQBootstrap && isPlatformSupported
    }

    /// Hybrid bundle to attach to outbound ratchet pre-keys when active.
    static func outboundHybridBundle() async -> HybridPublicBundleWire? {
        guard isActive else { return nil }
        // Future: generate via CryptoKit ML-KEM + P-256 and persist in keychain.
        return nil
    }

    /// Cache a peer's hybrid bundle for a future PQ session establishment.
    static func cachePeerHybridBundle(peerDID: String, bundle: HybridPublicBundleWire?) {
        guard let bundle else {
            UserDefaults.standard.removeObject(forKey: peerKey(peerDID))
            return
        }
        if let data = try? JSONEncoder().encode(bundle) {
            UserDefaults.standard.set(data, forKey: peerKey(peerDID))
        }
    }

    static func cachedPeerHybridBundle(peerDID: String) -> HybridPublicBundleWire? {
        guard let data = UserDefaults.standard.data(forKey: peerKey(peerDID)) else { return nil }
        return try? JSONDecoder().decode(HybridPublicBundleWire.self, from: data)
    }

    private static func peerKey(_ peerDID: String) -> String {
        "echo.pq_hybrid.peer.bundle." + peerDID
    }
}
#endif
