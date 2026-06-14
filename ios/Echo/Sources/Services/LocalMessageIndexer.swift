#if os(iOS)
import Foundation

/// Device-local inverted index for message search (WO-3 / M6).
actor LocalMessageIndexer {
    static let shared = LocalMessageIndexer()

    private let store = EncryptedIndexStore.shared
    private var snapshot = SearchIndexSnapshot()
    private var loaded = false

    func indexMessage(
        conversationId: String,
        messageId: String,
        senderDID: String,
        body: String,
        sentAt: Date?,
        contentType: String = "text"
    ) async {
        await ensureLoaded()
        indexMessageLocked(
            conversationId: conversationId,
            messageId: messageId,
            senderDID: senderDID,
            body: body,
            sentAt: sentAt,
            contentType: contentType
        )
        try? await store.save(snapshot)
    }

    func removeMessage(messageId: String) async {
        await ensureLoaded()
        removeMessageLocked(messageId: messageId)
        try? await store.save(snapshot)
    }

    func currentSnapshot() async -> SearchIndexSnapshot {
        await ensureLoaded()
        return snapshot
    }

    /// Bootstrap index from all persisted thread rows (idempotent).
    func rebuildFromLocalThreads() async {
        snapshot = SearchIndexSnapshot()
        loaded = true
        for conversationId in ConversationThreadStore.allStoredConversationIds() {
            for msg in ConversationThreadStore.exportMessages(conversationId: conversationId) {
                indexMessageLocked(
                    conversationId: conversationId,
                    messageId: msg.id,
                    senderDID: msg.senderDID,
                    body: msg.content,
                    sentAt: StoredThreadMessage.parseSentAt(msg.sentAtISO),
                    contentType: inferContentType(msg.content)
                )
            }
        }
        try? await store.save(snapshot)
    }

    private func ensureLoaded() async {
        guard !loaded else { return }
        if let saved = try? await store.load() {
            snapshot = saved
        }
        loaded = true
    }

    private func indexMessageLocked(
        conversationId: String,
        messageId: String,
        senderDID: String,
        body: String,
        sentAt: Date?,
        contentType: String
    ) {
        removeMessageLocked(messageId: messageId)
        let timestamp = sentAt?.timeIntervalSince1970 ?? Date().timeIntervalSince1970
        let preview = String(body.prefix(200))
        snapshot.documents[messageId] = SearchDocument(
            messageId: messageId,
            conversationId: conversationId,
            senderDID: senderDID,
            bodyPreview: preview,
            timestamp: timestamp,
            contentType: contentType
        )
        for token in Set(MessageSearchTokenizer.tokenize(body)) {
            var list = snapshot.postings[token] ?? []
            list.append(SearchPosting(
                messageId: messageId,
                conversationId: conversationId,
                timestamp: timestamp,
                fieldType: "body"
            ))
            snapshot.postings[token] = list
        }
    }

    private func removeMessageLocked(messageId: String) {
        guard let doc = snapshot.documents.removeValue(forKey: messageId) else { return }
        for token in MessageSearchTokenizer.tokenize(doc.bodyPreview) {
            snapshot.postings[token]?.removeAll { $0.messageId == messageId }
            if snapshot.postings[token]?.isEmpty == true {
                snapshot.postings.removeValue(forKey: token)
            }
        }
    }

    private func inferContentType(_ body: String) -> String {
        if body.contains("🎤") { return "audio" }
        if body.contains("📷") { return "image" }
        if body.contains("🎬") { return "video" }
        if body.contains("📎") { return "file" }
        if body.contains("http://") || body.contains("https://") { return "link" }
        return "text"
    }
}
#endif
