#if os(iOS)
import Foundation

/// Append-only local audit trail for hidden-folder access (WO-78).
struct HiddenFolderAuditEvent: Codable, Sendable, Equatable {
    enum Kind: String, Codable, Sendable {
        case unlock, lock, failedUnlock, screenshotLock, duressEntry
    }

    let kind: Kind
    let timestamp: Date
    let detail: String?
}

enum HiddenFolderAuditLog {
    private static let key = "echo.hidden.audit.v1"
    private static let maxEvents = 200

    static func record(_ kind: HiddenFolderAuditEvent.Kind, detail: String? = nil) {
        var events = load()
        events.append(HiddenFolderAuditEvent(kind: kind, timestamp: Date(), detail: detail))
        if events.count > maxEvents {
            events.removeFirst(events.count - maxEvents)
        }
        if let data = try? JSONEncoder().encode(events) {
            UserDefaults.standard.set(data, forKey: key)
        }
    }

    static func recent(limit: Int = 50) -> [HiddenFolderAuditEvent] {
        Array(load().suffix(limit))
    }

    private static func load() -> [HiddenFolderAuditEvent] {
        guard let data = UserDefaults.standard.data(forKey: key),
              let events = try? JSONDecoder().decode([HiddenFolderAuditEvent].self, from: data) else {
            return []
        }
        return events
    }
}
#endif
