import Foundation
import Observation

/// Local conversation list for Messages tab (Wave 0.1). Persists peer DIDs for Phase 3 signals.
struct StoredConversation: Identifiable, Codable, Hashable, Sendable {
    let id: String
    var contactName: String
    let peerDID: String
    var personaId: String
    var lastMessage: String
    var timestamp: String
    var unreadCount: Int
    var isOnline: Bool

    enum CodingKeys: String, CodingKey {
        case id, contactName, peerDID, personaId, lastMessage, timestamp, unreadCount, isOnline
    }

    init(
        id: String = UUID().uuidString,
        contactName: String,
        peerDID: String,
        personaId: String = UserDefaults.standard.string(forKey: "echo.activePersonaId") ?? "default",
        lastMessage: String = "",
        timestamp: String = "Now",
        unreadCount: Int = 0,
        isOnline: Bool = false
    ) {
        self.id = id
        self.contactName = contactName
        self.peerDID = peerDID
        self.personaId = personaId
        self.lastMessage = lastMessage
        self.timestamp = timestamp
        self.unreadCount = unreadCount
        self.isOnline = isOnline
    }

    init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        id = try c.decode(String.self, forKey: .id)
        contactName = try c.decode(String.self, forKey: .contactName)
        peerDID = try c.decode(String.self, forKey: .peerDID)
        personaId = try c.decodeIfPresent(String.self, forKey: .personaId) ?? "default"
        lastMessage = try c.decodeIfPresent(String.self, forKey: .lastMessage) ?? ""
        timestamp = try c.decodeIfPresent(String.self, forKey: .timestamp) ?? "Now"
        unreadCount = try c.decodeIfPresent(Int.self, forKey: .unreadCount) ?? 0
        isOnline = try c.decodeIfPresent(Bool.self, forKey: .isOnline) ?? false
    }
}

@MainActor
@Observable
final class ConversationStore {
    static let shared = ConversationStore()

    private(set) var conversations: [StoredConversation] = []
    private let storageKey = "echo.conversations.v1"

    private init() {
        load()
        if conversations.isEmpty {
            seedDemoIfNeeded()
        }
    }

    func upsert(_ conversation: StoredConversation) {
        if let idx = conversations.firstIndex(where: { $0.id == conversation.id }) {
            conversations[idx] = conversation
        } else if let idx = conversations.firstIndex(where: { $0.peerDID == conversation.peerDID }) {
            conversations[idx] = conversation
        } else {
            conversations.insert(conversation, at: 0)
        }
        persist()
    }

    func appendMessagePreview(conversationId: String, preview: String) {
        guard let idx = conversations.firstIndex(where: { $0.id == conversationId }) else { return }
        conversations[idx].lastMessage = preview
        conversations[idx].timestamp = "Now"
        persist()
    }

    func conversation(id: String) -> StoredConversation? {
        conversations.first { $0.id == id }
    }

    private func persist() {
        guard let data = try? JSONEncoder().encode(conversations) else { return }
        UserDefaults.standard.set(data, forKey: storageKey)
    }

    private func load() {
        guard let data = UserDefaults.standard.data(forKey: storageKey),
              let decoded = try? JSONDecoder().decode([StoredConversation].self, from: data) else {
            return
        }
        conversations = decoded
    }

    /// Keeps list non-empty in dev until real contacts exist; remove when WO-39 list API is wired.
    private func seedDemoIfNeeded() {
        #if DEBUG
        conversations = [
            StoredConversation(
                id: "demo-1",
                contactName: "Echo Support",
                peerDID: "did:key:z6Mkdemo000000000000000000000000000000000000000000000000",
                lastMessage: "Tap to open secure chat",
                timestamp: "Now",
                unreadCount: 0,
                isOnline: true
            )
        ]
        persist()
        #endif
    }
}
