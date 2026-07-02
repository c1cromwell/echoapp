#if os(iOS)
import Foundation

/// Post-quantum mode opt-in (WO-257).
enum PostQuantumModeStore {
    private static let key = "echo.pq.mode.enabled"

    static var isEnabled: Bool {
        get { UserDefaults.standard.bool(forKey: key) }
        set {
            UserDefaults.standard.set(newValue, forKey: key)
            if newValue {
                Task { _ = try? await PQHybridKeyStore.shared.loadOrGenerateBundle() }
            }
        }
    }

    static var isAvailable: Bool { true }
}
#endif
