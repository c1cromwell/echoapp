#if os(iOS)
import Foundation

struct ConversationArchiveRequest: Encodable, Sendable {
    let archived: Bool
}

struct ConversationArchiveResponse: Decodable, Sendable {
    let conversationId: String?
    let archived: Bool

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case archived
    }
}

enum ConversationArchiveEndpoint: APIEndpoint {
    case get(conversationId: String)
    case set(conversationId: String)

    var path: String {
        switch self {
        case .get(let id), .set(let id):
            return "/v3/conversations/\(id)/archive"
        }
    }
}

protocol ConversationArchiveAPIClient: Sendable {
    func setArchived(_ archived: Bool, conversationId: String) async throws
}

actor LiveConversationArchiveAPIClient: ConversationArchiveAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func setArchived(_ archived: Bool, conversationId: String) async throws {
        let _: ConversationArchiveResponse = try await apiClient.post(
            endpoint: ConversationArchiveEndpoint.set(conversationId: conversationId),
            body: ConversationArchiveRequest(archived: archived)
        )
        ConversationArchiveStore.setArchived(archived, conversationId: conversationId)
    }
}
#endif
