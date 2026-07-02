#if os(iOS)
import Foundation

/// Locally encrypted scheduled outbound message (WO-65 / WO-76).
struct ScheduledMessageRecord: Codable, Identifiable, Sendable, Equatable {
    let id: String
    let conversationId: String
    let peerDID: String
    var plaintext: String
    let fireAt: Date
    var createdAt: Date
    var status: Status

    enum Status: String, Codable, Sendable {
        case pending, delivered, cancelled
    }
}

/// UserDefaults-backed queue for client-side scheduled sends.
@MainActor
enum ScheduledMessageStore {
    private static let key = "echo.scheduled.messages.v1"

    static func all() -> [ScheduledMessageRecord] {
        guard let data = UserDefaults.standard.data(forKey: key),
              let rows = try? JSONDecoder().decode([ScheduledMessageRecord].self, from: data) else {
            return []
        }
        return rows
    }

    static func pending(now: Date = Date()) -> [ScheduledMessageRecord] {
        all().filter { $0.status == .pending && $0.fireAt <= now }
    }

    static func schedule(
        conversationId: String,
        peerDID: String,
        plaintext: String,
        fireAt: Date
    ) -> ScheduledMessageRecord {
        var rows = all()
        let record = ScheduledMessageRecord(
            id: UUID().uuidString,
            conversationId: conversationId,
            peerDID: peerDID,
            plaintext: plaintext,
            fireAt: fireAt,
            createdAt: Date(),
            status: .pending
        )
        rows.append(record)
        persist(rows)
        ScheduledMessageBGTask.scheduleNext(fireAt: fireAt)
        return record
    }

    static func cancel(id: String) {
        var rows = all()
        guard let idx = rows.firstIndex(where: { $0.id == id }) else { return }
        rows[idx].status = .cancelled
        persist(rows)
    }

    static func markDelivered(id: String) {
        var rows = all()
        guard let idx = rows.firstIndex(where: { $0.id == id }) else { return }
        rows[idx].status = .delivered
        persist(rows)
    }

    private static func persist(_ rows: [ScheduledMessageRecord]) {
        guard let data = try? JSONEncoder().encode(rows) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }
}
#endif
