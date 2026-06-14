#if os(iOS)
import Foundation

struct OverflowManifestPayload: Decodable, Sendable {
    let storageURIs: [String]

    enum CodingKeys: String, CodingKey {
        case storageURIs = "storage_uris"
    }
}

struct OverflowBlobResponse: Decodable, Sendable {
    let storageURI: String?
    let ciphertextBase64: String

    enum CodingKeys: String, CodingKey {
        case storageURI = "storage_uri"
        case ciphertextBase64 = "ciphertext_base64"
    }
}

enum OverflowRelayEndpoint: APIEndpoint {
    case fetch(uri: String)

    var path: String {
        switch self {
        case .fetch(let uri):
            return "/v3/relay/overflow/\(uri)"
        }
    }
}

actor OverflowRelayAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func fetchBlob(uri: String) async throws -> Data {
        let response: OverflowBlobResponse = try await apiClient.get(
            endpoint: OverflowRelayEndpoint.fetch(uri: uri)
        )
        guard let data = Data(base64Encoded: response.ciphertextBase64) else {
            throw OverflowRelayError.invalidPayload
        }
        return data
    }
}

enum OverflowRelayError: LocalizedError {
    case invalidPayload

    var errorDescription: String? {
        switch self {
        case .invalidPayload: return "Overflow blob payload was invalid."
        }
    }
}

/// Handles `overflow_manifest` WS messages (WO-237): fetches encblob URIs and replays queued payloads.
enum OverflowManifestHandler {
    private struct OverflowManifestWSMessage: Decodable {
        let type: String
        let payload: OverflowManifestPayload
    }

    static func tryHandle(text: String, reprocess: @escaping @Sendable (String) -> Void) -> Bool {
        guard let data = text.data(using: .utf8),
              let envelope = try? JSONDecoder().decode(OverflowManifestWSMessage.self, from: data),
              envelope.type == "overflow_manifest" else {
            return false
        }
        let uris = envelope.payload.storageURIs
        guard !uris.isEmpty else { return true }
        Task {
            guard let client = await DIContainer.shared.resolveAPIClient() else { return }
            let api = OverflowRelayAPIClient(apiClient: client)
            for uri in uris {
                guard let blob = try? await api.fetchBlob(uri: uri),
                      let replay = String(data: blob, encoding: .utf8) else { continue }
                reprocess(replay)
            }
        }
        return true
    }
}
#endif
