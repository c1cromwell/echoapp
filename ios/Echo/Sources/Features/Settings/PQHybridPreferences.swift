#if os(iOS)
import Foundation

/// PQ-hybrid ratchet bootstrap preference (WO-SX2).
///
/// Go server + `internal/crypto/pqhybrid.go` support ML-KEM-768 + P-256 hybrid
/// bootstrap today. iOS defers PQ until CryptoKit exposes ML-KEM; when enabled
/// here, clients should negotiate hybrid ciphertext in the ratchet pre-key signal.
enum PQHybridPreferences {
    private static let key = "echo.pq_hybrid.bootstrap.enabled"

    /// Off by default on iOS until ML-KEM is available in the platform SDK.
    static var usePQBootstrap: Bool {
        get { UserDefaults.standard.bool(forKey: key) }
        set { UserDefaults.standard.set(newValue, forKey: key) }
    }
}
#endif
