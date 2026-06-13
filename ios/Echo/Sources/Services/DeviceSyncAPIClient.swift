import Foundation

// MARK: - Wire models (match internal/api/sync_handlers.go + database.SyncEntry)

enum DeviceSyncEntryType {
    static let history = "history"
    static let message = "message"
    static let tombstone = "tombstone"
}

struct SyncPushRequest: Codable, Sendable {
    let targetDeviceId: String
    let entryType: String?
    let ciphertext: Data

    enum CodingKeys: String, CodingKey {
        case targetDeviceId = "target_device_id"
        case entryType = "entry_type"
        case ciphertext
    }
}

struct SyncPushResponse: Codable, Sendable {
    let targetDeviceId: String?
    let seq: Int64

    enum CodingKeys: String, CodingKey {
        case targetDeviceId = "target_device_id"
        case seq
    }
}

struct SyncEntryWire: Codable, Sendable, Identifiable {
    let controllerDid: String?
    let targetDeviceId: String?
    let seq: Int64
    let entryType: String?
    let ciphertext: Data
    let createdAt: String?

    var id: Int64 { seq }

    enum CodingKeys: String, CodingKey {
        case controllerDid
        case targetDeviceId
        case seq
        case entryType
        case ciphertext
        case createdAt
    }
}

struct SyncPullResponse: Codable, Sendable {
    let deviceId: String?
    let entries: [SyncEntryWire]
    let nextCursor: Int64

    enum CodingKeys: String, CodingKey {
        case deviceId = "device_id"
        case entries
        case nextCursor = "next_cursor"
    }
}

struct SyncHeadResponse: Codable, Sendable {
    let deviceId: String?
    let seq: Int64

    enum CodingKeys: String, CodingKey {
        case deviceId = "device_id"
        case seq
    }
}

struct SyncRevokeRequest: Codable, Sendable {
    let targetDeviceId: String

    enum CodingKeys: String, CodingKey {
        case targetDeviceId = "target_device_id"
    }
}

struct SyncRevokeResponse: Codable, Sendable {
    let targetDeviceId: String?
    let revoked: Bool

    enum CodingKeys: String, CodingKey {
        case targetDeviceId = "target_device_id"
        case revoked
    }
}

// MARK: - Client

protocol DeviceSyncAPIClient: Sendable {
    func push(targetDeviceId: String, ciphertext: Data, entryType: String) async throws -> Int64
    func pull(deviceId: String, after: Int64, limit: Int) async throws -> SyncPullResponse
    func head(deviceId: String) async throws -> Int64
    func revoke(targetDeviceId: String) async throws
}

#if os(iOS)

enum DeviceSyncEndpoint: APIEndpoint {
    case push
    case pull(deviceId: String, after: Int64, limit: Int)
    case head(deviceId: String)
    case revoke

    var path: String {
        switch self {
        case .push:
            return "/v3/sync/push"
        case .pull(let deviceId, let after, let limit):
            let encoded = deviceId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? deviceId
            return "/v3/sync/pull?device_id=\(encoded)&after=\(after)&limit=\(limit)"
        case .head(let deviceId):
            let encoded = deviceId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? deviceId
            return "/v3/sync/head?device_id=\(encoded)"
        case .revoke:
            return "/v3/sync/revoke"
        }
    }
}

actor LiveDeviceSyncAPIClient: DeviceSyncAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func push(targetDeviceId: String, ciphertext: Data, entryType: String) async throws -> Int64 {
        let body = SyncPushRequest(
            targetDeviceId: targetDeviceId,
            entryType: entryType,
            ciphertext: ciphertext
        )
        let resp: SyncPushResponse = try await apiClient.post(endpoint: DeviceSyncEndpoint.push, body: body)
        return resp.seq
    }

    func pull(deviceId: String, after: Int64, limit: Int) async throws -> SyncPullResponse {
        try await apiClient.get(endpoint: DeviceSyncEndpoint.pull(deviceId: deviceId, after: after, limit: limit))
    }

    func head(deviceId: String) async throws -> Int64 {
        let resp: SyncHeadResponse = try await apiClient.get(endpoint: DeviceSyncEndpoint.head(deviceId: deviceId))
        return resp.seq
    }

    func revoke(targetDeviceId: String) async throws {
        let _: SyncRevokeResponse = try await apiClient.post(
            endpoint: DeviceSyncEndpoint.revoke,
            body: SyncRevokeRequest(targetDeviceId: targetDeviceId)
        )
    }
}

enum DeviceSyncAPIError: LocalizedError {
    case deviceRevoked
    case emptyCiphertext

    var errorDescription: String? {
        switch self {
        case .deviceRevoked: return "This device's sync stream was revoked."
        case .emptyCiphertext: return "Sync payload is empty."
        }
    }
}
#endif
