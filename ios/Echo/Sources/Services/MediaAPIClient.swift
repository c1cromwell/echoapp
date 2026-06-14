#if os(iOS)
import Foundation

/// JSON envelope encrypted inside Kinnami text payloads for media messages (M5).
struct MediaMessageWire: Codable, Sendable, Equatable {
    let messageId: String
    let media: MediaAttachmentRef
    let caption: String?

    enum CodingKeys: String, CodingKey {
        case messageId = "message_id"
        case media
        case caption
    }
}

struct MediaUploadResponse: Decodable, Sendable {
    let fileId: String
    let chunkCount: Int
    let contentType: String
    let size: Int

    enum CodingKeys: String, CodingKey {
        case fileId
        case chunkCount
        case contentType
        case size
    }
}

struct OverflowManifest: Decodable, Sendable {
    let fileId: String
    let totalChunks: Int
    let chunkSizeBytes: Int
    let contentType: String
    let createdAt: String?

    enum CodingKeys: String, CodingKey {
        case fileId = "file_id"
        case totalChunks = "total_chunks"
        case chunkSizeBytes = "chunk_size_bytes"
        case contentType = "content_type"
        case createdAt = "created_at"
    }
}

enum MediaEndpoint: APIEndpoint {
    case upload
    case chunk(fileId: String, index: Int)
    case overflowManifest(fileId: String)

    var path: String {
        switch self {
        case .upload:
            return "/v3/media/upload"
        case .chunk(let fileId, let index):
            return "/v3/media/\(fileId)/chunks/\(index)"
        case .overflowManifest(let fileId):
            return "/v3/media/\(fileId)/manifest"
        }
    }
}

protocol MediaAPIClient: Sendable {
    func uploadEncrypted(data: Data, mimeType: String, trustTier: Int) async throws -> MediaUploadResponse
    func downloadChunks(fileId: String, chunkCount: Int) async throws -> Data
    func fetchOverflowManifest(fileId: String) async throws -> OverflowManifest
    func downloadWithManifest(fileId: String) async throws -> Data
}

actor LiveMediaAPIClient: MediaAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func uploadEncrypted(data: Data, mimeType: String, trustTier: Int) async throws -> MediaUploadResponse {
        try await apiClient.postRaw(
            endpoint: MediaEndpoint.upload,
            body: data,
            extraHeaders: [
                "Content-Type": mimeType,
                "X-Encrypted-Size": String(data.count),
                "X-Trust-Tier": String(trustTier),
            ]
        )
    }

    func downloadChunks(fileId: String, chunkCount: Int) async throws -> Data {
        guard chunkCount > 0 else { return Data() }
        var combined = Data()
        for index in 0..<chunkCount {
            let chunk = try await apiClient.getRaw(endpoint: MediaEndpoint.chunk(fileId: fileId, index: index))
            combined.append(chunk)
        }
        return combined
    }

    func fetchOverflowManifest(fileId: String) async throws -> OverflowManifest {
        try await apiClient.get(endpoint: MediaEndpoint.overflowManifest(fileId: fileId))
    }

    func downloadWithManifest(fileId: String) async throws -> Data {
        let manifest = try await fetchOverflowManifest(fileId: fileId)
        return try await downloadChunks(fileId: fileId, chunkCount: manifest.totalChunks)
    }
}
#endif
