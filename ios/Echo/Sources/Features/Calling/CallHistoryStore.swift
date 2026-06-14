#if os(iOS)
import Foundation

/// Local call history (WO-196) — UserDefaults-backed until SwiftData wiring lands.
enum CallHistoryStore {
    private static let storageKey = "echo.call.history"
    private static let maxRecords = 200

    static func load() -> [CallRecord] {
        guard let data = UserDefaults.standard.data(forKey: storageKey),
              let records = try? JSONDecoder().decode([CallRecord].self, from: data) else {
            return []
        }
        return records
    }

    static func append(_ record: CallRecord) {
        var records = load()
        records.insert(record, at: 0)
        if records.count > maxRecords {
            records = Array(records.prefix(maxRecords))
        }
        if let data = try? JSONEncoder().encode(records) {
            UserDefaults.standard.set(data, forKey: storageKey)
        }
    }

    static func missedCount(for peerDID: String) -> Int {
        load().filter { $0.peerDID == peerDID && $0.missedByCurrentUser }.count
    }

    #if DEBUG
    static func resetForTesting() {
        UserDefaults.standard.removeObject(forKey: storageKey)
    }
    #endif
}

struct CallRecord: Codable, Sendable, Identifiable, Equatable {
    let id: String
    let callType: CallType
    let peerDID: String
    let peerName: String
    let startedAt: Date
    let duration: TimeInterval
    let missedByCurrentUser: Bool
    let initiatorDID: String

    init(
        id: String = UUID().uuidString,
        callType: CallType,
        peerDID: String,
        peerName: String,
        startedAt: Date = Date(),
        duration: TimeInterval = 0,
        missedByCurrentUser: Bool = false,
        initiatorDID: String
    ) {
        self.id = id
        self.callType = callType
        self.peerDID = peerDID
        self.peerName = peerName
        self.startedAt = startedAt
        self.duration = duration
        self.missedByCurrentUser = missedByCurrentUser
        self.initiatorDID = initiatorDID
    }
}
#endif
