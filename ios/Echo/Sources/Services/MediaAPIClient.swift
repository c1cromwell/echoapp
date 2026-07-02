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
    case uploadInit
    case uploadChunk
    case uploadComplete
    case chunk(fileId: String, index: Int)
    case overflowManifest(fileId: String)

    var path: String {
        switch self {
        case .upload:
            return "/v3/media/upload"
        case .uploadInit:
            return "/v3/media/upload/init"
        case .uploadChunk:
            return "/v3/media/upload/chunk"
        case .uploadComplete:
            return "/v3/media/upload/complete"
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
    private static let chunkSize = 256 * 1024

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func uploadEncrypted(data: Data, mimeType: String, trustTier: Int) async throws -> MediaUploadResponse {
        if data.count > Self.chunkSize {
            return try await uploadChunked(data: data, mimeType: mimeType, trustTier: trustTier)
        }
        return try await apiClient.postRaw(
            endpoint: MediaEndpoint.upload,
            body: data,
            extraHeaders: [
                "Content-Type": mimeType,
                "X-Encrypted-Size": String(data.count),
                "X-Trust-Tier": String(trustTier),
            ]
        )
    }

    private func uploadChunked(data: Data, mimeType: String, trustTier: Int) async throws -> MediaUploadResponse {
        struct InitResponse: Decodable {
            let fileId: String
            let totalChunks: Int
            enum CodingKeys: String, CodingKey {
                case fileId = "file_id"
                case totalChunks = "total_chunks"
            }
        }
        struct ManifestResponse: Decodable {
            let fileId: String
            let totalChunks: Int
            let receivedChunks: Int?
            enum CodingKeys: String, CodingKey {
                case fileId = "file_id"
                case totalChunks = "total_chunks"
                case receivedChunks = "received_chunks"
            }
        }
        struct CompleteBody: Encodable {
            let fileId: String
            enum CodingKeys: String, CodingKey { case fileId = "file_id" }
        }

        var resumeState = MediaUploadResumeStore.pending().first {
            $0.encryptedSize == data.count && $0.mimeType == mimeType && $0.trustTier == trustTier
        }
        let fileId: String
        let chunkCount: Int
        if let existing = resumeState {
            fileId = existing.fileId
            chunkCount = existing.totalChunks
        } else {
            let initResp: InitResponse = try await apiClient.postRaw(
                endpoint: MediaEndpoint.uploadInit,
                body: Data(),
                extraHeaders: [
                    "Content-Type": mimeType,
                    "X-Encrypted-Size": String(data.count),
                    "X-Trust-Tier": String(trustTier),
                ]
            )
            fileId = initResp.fileId
            chunkCount = initResp.totalChunks
            let created = MediaUploadResumeState(
                fileId: fileId,
                totalChunks: chunkCount,
                mimeType: mimeType,
                trustTier: trustTier,
                encryptedSize: data.count,
                receivedChunks: [],
                createdAt: Date()
            )
            resumeState = created
            MediaUploadResumeStore.save(created)
        }

        var received = resumeState?.receivedChunks ?? []
        if received.isEmpty,
           let manifest: ManifestResponse = try? await apiClient.get(endpoint: MediaEndpoint.overflowManifest(fileId: fileId)) {
            let already = manifest.receivedChunks ?? 0
            if already > 0 {
                received = Set(0..<min(already, chunkCount))
            }
        }

        for index in 0..<chunkCount where !received.contains(index) {
            let start = index * Self.chunkSize
            let end = min(start + Self.chunkSize, data.count)
            let slice = data.subdata(in: start..<end)
            try await apiClient.putRaw(
                endpoint: MediaEndpoint.uploadChunk,
                body: slice,
                extraHeaders: [
                    "X-File-Id": fileId,
                    "X-Chunk-Index": String(index),
                ]
            )
            received.insert(index)
            var state = resumeState ?? MediaUploadResumeState(
                fileId: fileId,
                totalChunks: chunkCount,
                mimeType: mimeType,
                trustTier: trustTier,
                encryptedSize: data.count,
                receivedChunks: received,
                createdAt: Date()
            )
            state.receivedChunks = received
            MediaUploadResumeStore.save(state)
        }
        let result: MediaUploadResponse = try await apiClient.post(
            endpoint: MediaEndpoint.uploadComplete,
            body: CompleteBody(fileId: fileId)
        )
        MediaUploadResumeStore.clear(fileId: fileId)
        return result
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
