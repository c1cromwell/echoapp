#if os(iOS)
import Foundation

/// Post-quantum mode opt-in (WO-257) — mirrors PQ hybrid ratchet bootstrap.
enum PostQuantumModeStore {
    private static let key = "echo.pq.mode.enabled"

    static var isEnabled: Bool {
        get { PQHybridPreferences.usePQBootstrap }
        set {
            PQHybridPreferences.usePQBootstrap = newValue
            UserDefaults.standard.set(newValue, forKey: key)
            if newValue {
                Task { _ = try? await PQHybridKeyStore.shared.loadOrGenerateBundle() }
            }
        }
    }

    static var isAvailable: Bool { PQHybridBootstrap.isPlatformSupported }
}
#endif
