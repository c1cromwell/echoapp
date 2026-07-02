#if os(iOS)
import Foundation

/// Auto-backup scheduling for hidden-folder archives (WO-69).
enum HiddenFolderBackupScheduler {
    enum Frequency: String, CaseIterable {
        case manual, daily, weekly, monthly

        var label: String {
            switch self {
            case .manual: return "Manual"
            case .daily: return "Daily"
            case .weekly: return "Weekly"
            case .monthly: return "Monthly"
            }
        }

        var interval: TimeInterval? {
            switch self {
            case .manual: return nil
            case .daily: return 86_400
            case .weekly: return 604_800
            case .monthly: return 2_592_000
            }
        }
    }

    private static let frequencyKey = "echo.hidden.backup.frequency"
    private static let lastRunKey = "echo.hidden.backup.lastRun"
    private static let phraseConfiguredKey = "echo.hidden.backup.phraseConfigured"

    static var frequency: Frequency {
        get {
            let raw = UserDefaults.standard.string(forKey: frequencyKey) ?? Frequency.manual.rawValue
            return Frequency(rawValue: raw) ?? .manual
        }
        set { UserDefaults.standard.set(newValue.rawValue, forKey: frequencyKey) }
    }

    static var isPhraseConfigured: Bool {
        get { UserDefaults.standard.bool(forKey: phraseConfiguredKey) }
        set { UserDefaults.standard.set(newValue, forKey: phraseConfiguredKey) }
    }

    static func markPhraseConfigured() {
        isPhraseConfigured = true
    }

    static func runIfDue(phrase: RecoveryPhrase) async {
        guard isPhraseConfigured,
              frequency != .manual,
              let interval = frequency.interval else { return }
        let last = UserDefaults.standard.object(forKey: lastRunKey) as? Date ?? .distantPast
        guard Date().timeIntervalSince(last) >= interval else { return }
        guard await HiddenChatsSession.shared.isUnlocked, await !HiddenChatsSession.shared.isDuressMode else { return }
        do {
            _ = try await HiddenFolderBackupManager.createBackup(phrase: phrase)
            UserDefaults.standard.set(Date(), forKey: lastRunKey)
        } catch {
            // Best-effort hidden backup.
        }
    }
}
#endif
