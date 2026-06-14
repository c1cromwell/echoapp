#if os(iOS)
import Foundation

/// WO-64 auto-backup scheduling (daily/weekly) using a device-stored backup key.
enum BackupScheduler {
    enum Frequency: String, CaseIterable {
        case manual
        case daily
        case weekly

        var displayName: String {
            switch self {
            case .manual: return "Manual"
            case .daily: return "Daily"
            case .weekly: return "Weekly"
            }
        }

        var interval: TimeInterval? {
            switch self {
            case .manual: return nil
            case .daily: return 86_400
            case .weekly: return 604_800
            }
        }
    }

    private static let frequencyKey = "echo.backup.frequency"
    private static let wifiOnlyKey = "echo.backup.wifiOnly"
    private static let autoEnabledKey = "echo.backup.autoEnabled"

    static var frequency: Frequency {
        get {
            let raw = UserDefaults.standard.string(forKey: frequencyKey) ?? Frequency.manual.rawValue
            return Frequency(rawValue: raw) ?? .manual
        }
        set { UserDefaults.standard.set(newValue.rawValue, forKey: frequencyKey) }
    }

    static var wifiOnly: Bool {
        get { UserDefaults.standard.object(forKey: wifiOnlyKey) as? Bool ?? true }
        set { UserDefaults.standard.set(newValue, forKey: wifiOnlyKey) }
    }

    static var autoBackupEnabled: Bool {
        get { UserDefaults.standard.bool(forKey: autoEnabledKey) }
        set { UserDefaults.standard.set(newValue, forKey: autoEnabledKey) }
    }

    /// Call on app foreground / Messages tab open.
    static func runIfDue() {
        Task { @MainActor in
            guard autoBackupEnabled,
                  frequency != .manual,
                  BackupSessionKeyStore.hasStoredKey(),
                  let service = DIContainer.shared.resolveMessageBackup(),
                  service.isBackupDue(interval: frequency.interval ?? 0) else { return }
            if wifiOnly, !NetworkReachability.isOnWiFi { return }
            do {
                try await service.uploadCloudBackupWithStoredKey()
                try await service.createLocalBackupWithStoredKey()
            } catch {
                // Best-effort; user can retry manually from Backup & Security.
            }
        }
    }
}

enum NetworkReachability {
    static func isOnWiFi() -> Bool {
        // Conservative default until NWPathMonitor wiring lands in app target.
        true
    }
}
#endif
