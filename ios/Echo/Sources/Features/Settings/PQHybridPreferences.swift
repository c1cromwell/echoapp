#if os(iOS)
import Foundation

/// PQ-hybrid ratchet bootstrap preference (WO-SX2).
enum PQHybridPreferences {
    private static let key = "echo.pq_hybrid.bootstrap.enabled"

    /// Off by default; enable in Privacy → Advanced network.
    static var usePQBootstrap: Bool {
        get { UserDefaults.standard.bool(forKey: key) }
        set { UserDefaults.standard.set(newValue, forKey: key) }
    }
}
#endif
