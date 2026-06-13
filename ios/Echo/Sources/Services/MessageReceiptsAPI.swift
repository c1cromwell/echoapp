import Foundation

// MARK: - REST models (match internal/api/v3_handlers.go handleMessageReceipt)

/// POST /v3/messages/{id}/receipt body. `receiptType` is "delivered" or "read".
struct MessageReceiptRequest: Codable, Sendable, Equatable {
    let receiptType: String
}

/// Response of the receipt POST.
struct MessageReceiptResponse: Codable, Sendable, Equatable {
    let messageId: String
    let receiptType: String
    let timestamp: String
}

/// GET /v3/messages/{id}/status — durable delivery state for reconnect sync (WO-48).
struct MessageStatusResponse: Codable, Sendable, Equatable {
    let messageId: String
    let conversationId: String?
    let status: String           // "queued" | "delivered" | "read" | ...
    let deliveredAt: String?
    let readAt: String?

    /// Maps the server status string onto the UI delivery status (never regresses upstream).
    var deliveryStatus: DeliveryStatus? {
        switch status {
        case "read": return .read
        case "delivered": return .delivered
        case "queued": return .sent
        default: return nil
        }
    }
}

// MARK: - Client protocol (mock in tests)

protocol MessageReceiptsAPIClient: Sendable {
    /// Durably records a read receipt; the server then pushes a live read_receipt to the sender.
    @discardableResult
    func markRead(messageId: String) async throws -> MessageReceiptResponse
    @discardableResult
    func markDelivered(messageId: String) async throws -> MessageReceiptResponse
    /// Pulls durable delivery state so a reconnecting client can reconcile missed receipts.
    func status(messageId: String) async throws -> MessageStatusResponse
}

#if os(iOS)

enum MessageReceiptEndpoint: APIEndpoint {
    case receipt(messageId: String)
    case status(messageId: String)

    private static func encode(_ id: String) -> String {
        id.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? id
    }

    var path: String {
        switch self {
        case .receipt(let messageId):
            return "/v3/messages/\(Self.encode(messageId))/receipt"
        case .status(let messageId):
            return "/v3/messages/\(Self.encode(messageId))/status"
        }
    }
}

/// Durable receipts via signed REST (PasskeySigningInterceptor on APIClient), matching
/// the WS read_receipt signal which the server now fans out authoritatively.
actor MessageReceiptsAPI: MessageReceiptsAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    @discardableResult
    func markRead(messageId: String) async throws -> MessageReceiptResponse {
        try await apiClient.post(
            endpoint: MessageReceiptEndpoint.receipt(messageId: messageId),
            body: MessageReceiptRequest(receiptType: "read")
        )
    }

    @discardableResult
    func markDelivered(messageId: String) async throws -> MessageReceiptResponse {
        try await apiClient.post(
            endpoint: MessageReceiptEndpoint.receipt(messageId: messageId),
            body: MessageReceiptRequest(receiptType: "delivered")
        )
    }

    func status(messageId: String) async throws -> MessageStatusResponse {
        try await apiClient.get(endpoint: MessageReceiptEndpoint.status(messageId: messageId))
    }
}

#endif
