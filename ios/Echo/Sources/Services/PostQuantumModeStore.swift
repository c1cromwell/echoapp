#if os(iOS)
import Foundation

/// Post-quantum mode opt-in stub (WO-257).
enum PostQuantumModeStore {
    private static let key = "echo.pq.mode.enabled"

    static var isEnabled: Bool {
        get { UserDefaults.standard.bool(forKey: key) }
        set { UserDefaults.standard.set(newValue, forKey: key) }
    }

    static var isAvailable: Bool { false }
}
#endif
