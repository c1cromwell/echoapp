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

    fileprivate var interval: TimeInterval? {
        switch self {
        case .manual: return nil
        case .weekly: return 7 * 24 * 60 * 60
        case .monthly: return 30 * 24 * 60 * 60
        }
    }
}

enum ContactDiscoverySyncPreferences {
    private static let cadenceKey = "echo.contactDiscovery.syncCadence"
    private static let lastSyncedKey = "echo.contactDiscovery.lastSyncedAt"
    private static let firstScanKey = "echo.contactDiscovery.hasCompletedFirstScan"

    static var cadence: ContactDiscoverySyncCadence {
        get {
            guard let raw = UserDefaults.standard.string(forKey: cadenceKey),
                  let value = ContactDiscoverySyncCadence(rawValue: raw) else {
                return .manual
            }
            return value
        }
        set {
            UserDefaults.standard.set(newValue.rawValue, forKey: cadenceKey)
        }
    }

    static var lastSyncedAt: Date? {
        let ts = UserDefaults.standard.double(forKey: lastSyncedKey)
        guard ts > 0 else { return nil }
        return Date(timeIntervalSince1970: ts)
    }

    static var hasCompletedFirstScan: Bool {
        get { UserDefaults.standard.bool(forKey: firstScanKey) }
        set { UserDefaults.standard.set(newValue, forKey: firstScanKey) }
    }

    static func markSyncedNow() {
        UserDefaults.standard.set(Date().timeIntervalSince1970, forKey: lastSyncedKey)
        hasCompletedFirstScan = true
    }

    static func shouldRunAutomaticSync(now: Date = Date()) -> Bool {
        guard let interval = cadence.interval else { return false }
        guard let last = lastSyncedAt else { return true }
        return now.timeIntervalSince(last) >= interval
    }
}
#endif
