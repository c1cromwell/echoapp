#if os(iOS)
import Foundation

/// Persisted chat row for local thread history (Wave 0.1 / Phase 3 E2E).
struct StoredThreadMessage: Codable, Equatable, Sendable, Identifiable {
    let id: String
    let senderDID: String
    var content: String
    var timestamp: String
    var deliveryStatus: DeliveryStatus?
    var replyToMessageId: String?
    var replyPreview: String?
    var forwardedFromMessageId: String?
    var forwardedFromConversationId: String?
    var sentAtISO: String?
    var expiresAtISO: String?
    var commitmentHex: String?
    var snapshotHash: String?
    var snapshotHeight: Int?

    func asChatDetailMessage(currentUserDID: String) -> ChatDetailMessage {
        ChatDetailMessage(
            id: id,
            senderDID: senderDID,
            currentUserDID: currentUserDID,
            content: content,
            timestamp: timestamp,
            deliveryStatus: deliveryStatus,
            replyToMessageId: replyToMessageId,
            replyPreview: replyPreview,
            forwardedFromMessageId: forwardedFromMessageId,
            forwardedFromConversationId: forwardedFromConversationId,
            sentAt: StoredThreadMessage.parseSentAt(sentAtISO),
            expiresAt: StoredThreadMessage.parseSentAt(expiresAtISO),
            commitmentHex: commitmentHex,
            snapshotHash: snapshotHash,
            snapshotHeight: snapshotHeight
        )
    }

    init(
        id: String,
        senderDID: String,
        content: String,
        timestamp: String,
        deliveryStatus: DeliveryStatus? = nil,
        replyToMessageId: String? = nil,
        replyPreview: String? = nil,
        forwardedFromMessageId: String? = nil,
        forwardedFromConversationId: String? = nil,
        sentAtISO: String? = nil,
        expiresAtISO: String? = nil,
        commitmentHex: String? = nil,
        snapshotHash: String? = nil,
        snapshotHeight: Int? = nil
    ) {
        self.id = id
        self.senderDID = senderDID
        self.content = content
        self.timestamp = timestamp
        self.deliveryStatus = deliveryStatus
        self.replyToMessageId = replyToMessageId
        self.replyPreview = replyPreview
        self.forwardedFromMessageId = forwardedFromMessageId
        self.forwardedFromConversationId = forwardedFromConversationId
        self.sentAtISO = sentAtISO
        self.expiresAtISO = expiresAtISO
        self.commitmentHex = commitmentHex
        self.snapshotHash = snapshotHash
        self.snapshotHeight = snapshotHeight
    }

    init(from message: ChatDetailMessage) {
        id = message.id
        senderDID = message.senderDID
        content = message.content
        timestamp = message.timestamp
        deliveryStatus = message.deliveryStatus
        replyToMessageId = message.replyToMessageId
        replyPreview = message.replyPreview
        forwardedFromMessageId = message.forwardedFromMessageId
        forwardedFromConversationId = message.forwardedFromConversationId
        sentAtISO = StoredThreadMessage.formatSentAt(message.sentAt)
        expiresAtISO = StoredThreadMessage.formatSentAt(message.expiresAt)
        commitmentHex = message.commitmentHex
        snapshotHash = message.snapshotHash
        snapshotHeight = message.snapshotHeight
    }

    private static let sentAtFormatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return f
    }()

    static func formatSentAt(_ date: Date?) -> String? {
        guard let date else { return nil }
        return sentAtFormatter.string(from: date)
    }

    static func parseSentAt(_ iso: String?) -> Date? {
        guard let iso, !iso.isEmpty else { return nil }
        return sentAtFormatter.date(from: iso)
            ?? ISO8601DateFormatter().date(from: iso)
    }
}

/// UserDefaults-backed message history per conversation id.
@MainActor
enum ConversationThreadStore {
    private static let keyPrefix = "echo.thread.v1."

    static func load(conversationId: String, currentUserDID: String) -> [ChatDetailMessage] {
        guard !conversationId.isEmpty else { return [] }
        return loadStored(conversationId: conversationId)
            .map { $0.asChatDetailMessage(currentUserDID: currentUserDID) }
    }

    static func replace(conversationId: String, messages: [ChatDetailMessage]) {
        guard !conversationId.isEmpty else { return }
        let stored = messages.map(StoredThreadMessage.init)
        persist(conversationId: conversationId, messages: stored)
    }

    static func replaceStored(conversationId: String, messages: [StoredThreadMessage]) {
        persist(conversationId: conversationId, messages: messages)
    }

    static func loadStoredOnly(conversationId: String) -> [StoredThreadMessage] {
        loadStored(conversationId: conversationId)
    }

    static func allConversationIds() -> [String] {
        UserDefaults.standard.dictionaryRepresentation().keys
            .filter { $0.hasPrefix(keyPrefix) }
            .map { String($0.dropFirst(keyPrefix.count)) }
    }

    static func appendIfNew(conversationId: String, message: ChatDetailMessage) {
        appendStoredIfNew(conversationId: conversationId, message: StoredThreadMessage(from: message))
    }

    static func appendStoredIfNew(conversationId: String, message: StoredThreadMessage) {
        guard !conversationId.isEmpty else { return }
        var stored = loadStored(conversationId: conversationId)
        guard !stored.contains(where: { $0.id == message.id }) else { return }
        stored.append(message)
        if !ConversationPreferencesStore.shared.isHidden(conversationId) {
            notifySearchIndex(conversationId: conversationId, message: message)
        }
        persist(conversationId: conversationId, messages: stored)
    }

    static func updateDeliveryStatus(
        conversationId: String,
        messageId: String,
        status: DeliveryStatus
    ) {
        guard !conversationId.isEmpty else { return }
        var stored = loadStored(conversationId: conversationId)
        guard let idx = stored.firstIndex(where: { $0.id == messageId }) else { return }
        let current = stored[idx].deliveryStatus
        stored[idx].deliveryStatus = DeliveryStatusAdvancement.advanced(current: current, incoming: status)
        persist(conversationId: conversationId, messages: stored)
    }

    private static func loadStored(conversationId: String) -> [StoredThreadMessage] {
        let key = keyPrefix + conversationId
        guard let data = UserDefaults.standard.data(forKey: key) else { return [] }
        if let decoded = try? JSONDecoder().decode([StoredThreadMessage].self, from: data) {
            return decoded
        }
        if let decrypted = try? HiddenThreadCrypto.decrypt(data: data, conversationId: conversationId) {
            return decrypted
        }
        return []
    }

    private static func persist(conversationId: String, messages: [StoredThreadMessage]) {
        let key = keyPrefix + conversationId
        if ConversationPreferencesStore.shared.isHidden(conversationId) {
            guard let data = try? HiddenThreadCrypto.encrypt(messages: messages, conversationId: conversationId) else { return }
            UserDefaults.standard.set(data, forKey: key)
            return
        }
        guard let data = try? JSONEncoder().encode(messages) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }

    private static func notifySearchIndex(conversationId: String, message: StoredThreadMessage) {
        guard !ConversationPreferencesStore.shared.isHidden(conversationId) else { return }
        Task {
            await LocalMessageIndexer.shared.indexMessage(
                conversationId: conversationId,
                messageId: message.id,
                senderDID: message.senderDID,
                body: message.content,
                sentAt: StoredThreadMessage.parseSentAt(message.sentAtISO)
            )
        }
    }

    // MARK: - M3 history sync export (WO-CA3)

    /// Conversation ids that have persisted thread rows.
    static func allStoredConversationIds() -> [String] {
        let prefix = keyPrefix
        return UserDefaults.standard.dictionaryRepresentation().keys
            .filter { $0.hasPrefix(prefix) }
            .map { String($0.dropFirst(prefix.count)) }
            .sorted()
    }

    static func exportMessages(conversationId: String) -> [StoredThreadMessage] {
        loadStored(conversationId: conversationId)
    }

    static func replaceStored(conversationId: String, messages: [StoredThreadMessage]) {
        persist(conversationId: conversationId, messages: messages)
    }

    /// Idempotent merge: skip message ids already present locally.
    static func mergeMessages(conversationId: String, incoming: [StoredThreadMessage]) {
        guard !conversationId.isEmpty, !incoming.isEmpty else { return }
        var stored = loadStored(conversationId: conversationId)
        let existing = Set(stored.map(\.id))
        for msg in incoming where !existing.contains(msg.id) {
            stored.append(msg)
            if !ConversationPreferencesStore.shared.isHidden(conversationId) {
                notifySearchIndex(conversationId: conversationId, message: msg)
            }
        }
        persist(conversationId: conversationId, messages: stored)
    }

    /// Upsert a single message for incremental device sync (updates in place when id matches).
    static func upsertMessage(conversationId: String, message: StoredThreadMessage) {
        guard !conversationId.isEmpty else { return }
        var stored = loadStored(conversationId: conversationId)
        if let idx = stored.firstIndex(where: { $0.id == message.id }) {
            stored[idx] = message
        } else {
            stored.append(message)
            if !ConversationPreferencesStore.shared.isHidden(conversationId) {
                notifySearchIndex(conversationId: conversationId, message: message)
            }
        }
        persist(conversationId: conversationId, messages: stored)
    }

    static func removeMessage(conversationId: String, messageId: String) {
        guard !conversationId.isEmpty else { return }
        var stored = loadStored(conversationId: conversationId)
        let before = stored.count
        stored.removeAll { $0.id == messageId }
        guard stored.count != before else { return }
        persist(conversationId: conversationId, messages: stored)
    }

    /// Re-write thread storage when hide preference changes (plaintext ↔ encrypted).
    static func migrateStorageEncryption(conversationId: String) {
        guard !conversationId.isEmpty else { return }
        let stored = loadStored(conversationId: conversationId)
        guard !stored.isEmpty else { return }
        persist(conversationId: conversationId, messages: stored)
    }
}
#endif
