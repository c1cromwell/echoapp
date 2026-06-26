#if os(iOS)
import Foundation

/// Client preference for Double Ratchet 1:1 encryption (WO-SX1). Enabled by default.
enum DoubleRatchetPreferences {
    private static let key = "echo.double_ratchet.enabled"

    static var isEnabled: Bool {
        get {
            if UserDefaults.standard.object(forKey: key) == nil { return true }
            return UserDefaults.standard.bool(forKey: key)
        }
        set { UserDefaults.standard.set(newValue, forKey: key) }
    }
}
#endif
