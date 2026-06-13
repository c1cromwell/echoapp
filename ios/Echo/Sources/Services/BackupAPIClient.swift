#if os(iOS)
import Foundation

struct BackupPushRequest: Encodable, Sendable {
    let ciphertextBase64: String
    let contentHash: String?

    enum CodingKeys: String, CodingKey {
        case ciphertextBase64 = "ciphertext_base64"
        case contentHash = "content_hash"
    }
}

struct BackupPushResponse: Decodable, Sendable {
    let storageURI: String?
    let contentHash: String?
    let byteSize: Int?
    let version: Int?
    let updatedAt: String?

    enum CodingKeys: String, CodingKey {
        case storageURI = "storage_uri"
        case contentHash = "content_hash"
        case byteSize = "byte_size"
        case version
        case updatedAt = "updated_at"
    }
}

struct BackupPullResponse: Decodable, Sendable {
    let storageURI: String?
    let contentHash: String?
    let byteSize: Int?
    let version: Int?
    let updatedAt: String?
    let ciphertextBase64: String

    enum CodingKeys: String, CodingKey {
        case storageURI = "storage_uri"
        case contentHash = "content_hash"
        case byteSize = "byte_size"
        case version
        case updatedAt = "updated_at"
        case ciphertextBase64 = "ciphertext_base64"
    }
}

protocol BackupAPIClient: Sendable {
    func push(ciphertext: Data) async throws -> BackupPushResponse
    func pull() async throws -> Data
}

enum BackupEndpoint: APIEndpoint {
    case push
    case pull

    var path: String {
        switch self {
        case .push: return "/v3/backup/push"
        case .pull: return "/v3/backup/pull"
        }
    }
}

actor LiveBackupAPIClient: BackupAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func push(ciphertext: Data) async throws -> BackupPushResponse {
        let body = BackupPushRequest(
            ciphertextBase64: Self.encodeBase64URL(ciphertext),
            contentHash: nil
        )
        return try await apiClient.post(endpoint: BackupEndpoint.push, body: body)
    }

    func pull() async throws -> Data {
        let resp: BackupPullResponse = try await apiClient.get(endpoint: BackupEndpoint.pull)
        guard let data = Self.decodeBase64URL(resp.ciphertextBase64) else {
            throw MessageBackupError.invalidCloudPayload
        }
        return data
    }

    private static func encodeBase64URL(_ data: Data) -> String {
        data.base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }

    private static func decodeBase64URL(_ raw: String) -> Data? {
        if raw.contains("+") || raw.contains("/") {
            return Data(base64Encoded: raw)
        }
        var padded = raw
        let rem = padded.count % 4
        if rem > 0 { padded += String(repeating: "=", count: 4 - rem) }
        let std = padded
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        return Data(base64Encoded: std)
    }
}
#endif
