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
    var remoteId: String?

    enum Status: String, Codable, Sendable {
        case pending, delivered, cancelled
    }
}

/// UserDefaults-backed queue for client-side scheduled sends.
@MainActor
enum ScheduledMessageStore {
    private static let key = "echo.scheduled.messages.v1"

    static func all() -> [ScheduledMessageRecord] {
        guard let data = UserDefaults.standard.data(forKey: key) else { return [] }
        if let rows = try? ScheduledMessageCrypto.decrypt(data) { return rows }
        if let legacy = try? JSONDecoder().decode([ScheduledMessageRecord].self, from: data) {
            persist(legacy)
            return legacy
        }
        return []
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
            status: .pending,
            remoteId: nil
        )
        rows.append(record)
        persist(rows)
        ScheduledMessageBGTask.scheduleNext(fireAt: fireAt)
        Task { await pushRemote(record) }
        return record
    }

    static func cancel(id: String) {
        var rows = all()
        guard let idx = rows.firstIndex(where: { $0.id == id }) else { return }
        let remoteId = rows[idx].remoteId
        rows[idx].status = .cancelled
        persist(rows)
        if let remoteId {
            Task { await deleteRemote(id: remoteId) }
        }
    }

    static func markDelivered(id: String) {
        var rows = all()
        guard let idx = rows.firstIndex(where: { $0.id == id }) else { return }
        let remoteId = rows[idx].remoteId
        rows[idx].status = .delivered
        persist(rows)
        if let remoteId {
            Task { await deleteRemote(id: remoteId) }
        }
    }

    static func attachRemoteId(localId: String, remoteId: String) {
        var rows = all()
        guard let idx = rows.firstIndex(where: { $0.id == localId }) else { return }
        rows[idx].remoteId = remoteId
        persist(rows)
    }

    /// Pull pending rows from the gateway so a second device can see them (content omitted remotely).
    static func hydrateFromServer() async {
        guard let api = DIContainer.shared.resolveAPIClient() else { return }
        let client = ScheduledMessageAPI(apiClient: api)
        guard let remote = try? await client.list() else { return }
        var rows = all()
        var changed = false
        for dto in remote where dto.status == "pending" {
            if rows.contains(where: { $0.remoteId == dto.id }) { continue }
            let fireAt = ISO8601DateFormatter().date(from: dto.scheduledAt) ?? Date()
            var plaintext = ""
            var peerDID = ""
            if let full = try? await client.get(id: dto.id),
               let envelope = try? await client.decryptEnvelope(from: full) {
                plaintext = envelope.body
                peerDID = envelope.peerDID
            }
            rows.append(ScheduledMessageRecord(
                id: UUID().uuidString,
                conversationId: dto.conversationId,
                peerDID: peerDID,
                plaintext: plaintext,
                fireAt: fireAt,
                createdAt: Date(),
                status: .pending,
                remoteId: dto.id
            ))
            changed = true
        }
        if changed {
            persist(rows)
            ScheduledMessageBGTask.scheduleNext()
        }
    }

    private static func pushRemote(_ record: ScheduledMessageRecord) async {
        guard let api = DIContainer.shared.resolveAPIClient() else { return }
        let client = ScheduledMessageAPI(apiClient: api)
        guard let dto = try? await client.create(
            conversationId: record.conversationId,
            peerDID: record.peerDID,
            plaintext: record.plaintext,
            fireAt: record.fireAt
        ) else { return }
        attachRemoteId(localId: record.id, remoteId: dto.id)
    }

    private static func deleteRemote(id: String) async {
        guard let api = DIContainer.shared.resolveAPIClient() else { return }
        let client = ScheduledMessageAPI(apiClient: api)
        try? await client.cancel(id: id)
    }

    private static func persist(_ rows: [ScheduledMessageRecord]) {
        guard let data = try? ScheduledMessageCrypto.encrypt(rows) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }
}
#endif
