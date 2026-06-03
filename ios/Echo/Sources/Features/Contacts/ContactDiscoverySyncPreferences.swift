#if os(iOS)
import Foundation

enum ContactDiscoverySyncCadence: String, CaseIterable, Sendable {
    case manual
    case weekly
    case monthly

    var label: String {
        switch self {
        case .manual: return "Manual only"
        case .weekly: return "Weekly"
        case .monthly: return "Monthly"
        }
    }
}

enum ContactDiscoverySyncPreferences {
    private static let key = "echo.contactDiscovery.syncCadence"

    static var cadence: ContactDiscoverySyncCadence {
        get {
            guard let raw = UserDefaults.standard.string(forKey: key),
                  let value = ContactDiscoverySyncCadence(rawValue: raw) else {
                return .manual
            }
            return value
        }
        set {
            UserDefaults.standard.set(newValue.rawValue, forKey: key)
        }
    }
}
#endif
