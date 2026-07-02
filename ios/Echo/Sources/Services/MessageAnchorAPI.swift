#if os(iOS)
import Foundation

// MARK: - Wire models (WO-15 / WO-227)

struct MessageAnchorProofResponse: Codable, Sendable, Equatable {
    let messageId: String
    let commitment: String
    let siblings: [String]
    let merkleLeafIndex: Int?
    let snapshotHash: String
    let snapshotHeight: Int64?
    let merkleRoot: String
}

struct AnchorConfirmationPayload: Codable, Sendable {
    let type: String?
    let messageId: String
    let snapshotHash: String
    let snapshotHeight: Int64?
    let merkleProof: [String]?
    let merkleRoot: String?
    let merkleLeafIndex: Int?
}

// MARK: - Endpoints

enum MessageAnchorEndpoint: APIEndpoint {
    case merkleProof(messageId: String)
    case existenceProof(messageId: String)

    var path: String {
        switch self {
        case .merkleProof(let id):
            return "/v1/messages/\(id)/merkle-proof"
        case .existenceProof(let id):
            return "/v1/messages/\(id)/proof"
        }
    }
}

// MARK: - Client

protocol MessageAnchorAPIClient: Sendable {
    func fetchMerkleProof(messageId: String) async throws -> MessageAnchorProofResponse
}

actor MessageAnchorAPI: MessageAnchorAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func fetchMerkleProof(messageId: String) async throws -> MessageAnchorProofResponse {
        try await apiClient.get(endpoint: MessageAnchorEndpoint.merkleProof(messageId: messageId))
    }
}

// MARK: - WS envelope decode

enum AnchorConfirmationCodec {
    private struct Outer: Decodable {
        let type: String
        let payload: AnchorConfirmationPayload?
    }

    static func decodeConfirmation(from text: String) -> AnchorConfirmationPayload? {
        guard let data = text.data(using: .utf8),
              let outer = try? JSONDecoder().decode(Outer.self, from: data),
              outer.type == "confirmation" else {
            return nil
        }
        if let payload = outer.payload, !payload.messageId.isEmpty {
            return payload
        }
        return try? JSONDecoder().decode(AnchorConfirmationPayload.self, from: data)
    }
}
#endif
