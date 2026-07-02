#if os(iOS)
import Foundation

/// Preserved commitment metadata after plaintext deletion (WO-105).
struct DisappearingMessageAuditRecord: Codable, Equatable, Sendable {
    let messageId: String
    let conversationId: String
    let commitmentHex: String?
    let expiresAtISO: String?
    let deletedAtISO: String
}

/// Retains commitment hashes after local purge for future proof generation.
enum DisappearingMessageAuditStore {
    private static let prefix = "echo.disappearing.audit."

    static func preserve(
        messageId: String,
        conversationId: String,
        commitmentHex: String?,
        expiresAt: Date?
    ) {
        let record = DisappearingMessageAuditRecord(
            messageId: messageId,
            conversationId: conversationId,
            commitmentHex: commitmentHex,
            expiresAtISO: expiresAt.map { StoredThreadMessage.formatSentAt($0) },
            deletedAtISO: StoredThreadMessage.formatSentAt(Date())
        )
        guard let data = try? JSONEncoder().encode(record) else { return }
        UserDefaults.standard.set(data, forKey: prefix + messageId)
    }

    static func record(for messageId: String) -> DisappearingMessageAuditRecord? {
        guard let data = UserDefaults.standard.data(forKey: prefix + messageId),
              let record = try? JSONDecoder().decode(DisappearingMessageAuditRecord.self, from: data) else {
            return nil
        }
        return record
    }
}

/// Client-side deletion sequence when a disappearing timer fires (WO-105).
@MainActor
enum DisappearingMessageLocalDelete {
    /// Deletes plaintext, keys, and media while preserving commitment audit metadata.
    static func deleteMessageLocally(
        messageId: String,
        conversationId: String,
        commitmentHex: String?,
        expiresAt: Date?
    ) async {
        DisappearingMessageAuditStore.preserve(
            messageId: messageId,
            conversationId: conversationId,
            commitmentHex: commitmentHex,
            expiresAt: expiresAt
        )

        try? await LocalDatabase.shared.deleteMessage(id: messageId)
        try? await KeychainManager.shared.delete(key: "echo.msgkey.\(messageId)")
        MediaAttachmentCache.shared.remove(messageId: messageId)

        var stored = ConversationThreadStore.exportMessages(conversationId: conversationId)
        stored.removeAll { $0.id == messageId }
        ConversationThreadStore.replaceStored(conversationId: conversationId, messages: stored)
    }
}

/// Lightweight on-disk media cache for chat attachments.
final class MediaAttachmentCache {
    static let shared = MediaAttachmentCache()
    private init() {}

    private var root: URL {
        FileManager.default.urls(for: .cachesDirectory, in: .userDomainMask)[0]
            .appendingPathComponent("echo-media", isDirectory: true)
    }

    func remove(messageId: String) {
        let url = root.appendingPathComponent(messageId)
        try? FileManager.default.removeItem(at: url)
    }
}
#endif
