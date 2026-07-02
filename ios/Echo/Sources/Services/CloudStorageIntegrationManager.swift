#if os(iOS)
import Foundation

/// OAuth + file proxy client for cloud pickers (WO-46).
enum CloudStorageIntegrationManager {
    enum Provider: String, CaseIterable, Sendable {
        case googleDrive = "google_drive"
        case dropbox = "dropbox"
        case oneDrive = "onedrive"

        var title: String {
            switch self {
            case .googleDrive: return "Google Drive"
            case .dropbox: return "Dropbox"
            case .oneDrive: return "OneDrive"
            }
        }
    }

    struct RemoteFile: Decodable, Identifiable, Sendable {
        let id: String
        let name: String
        let mimeType: String
        let size: Int64

        enum CodingKeys: String, CodingKey {
            case id, name, size
            case mimeType = "mime_type"
        }
    }

    private struct AuthorizeResponse: Decodable {
        let authorizeURL: String
        enum CodingKeys: String, CodingKey { case authorizeURL = "authorize_url" }
    }

    private struct ProvidersResponse: Decodable {
        let providers: [String]
    }

    private struct FilesResponse: Decodable {
        let files: [RemoteFile]
    }

    static func connectedProviders() async -> [Provider] {
        guard let client = DIContainer.shared.resolveAPIClient() else { return [] }
        do {
            let resp: ProvidersResponse = try await client.get(endpoint: CloudEndpoint.listProviders)
            return resp.providers.compactMap { Provider(rawValue: $0) }
        } catch {
            return []
        }
    }

    static func fetchAuthorizeURL(provider: Provider, redirectURI: String = "echo://oauth/cloud") async throws -> URL {
        guard let client = DIContainer.shared.resolveAPIClient() else {
            throw CloudStorageError.clientUnavailable
        }
        let resp: AuthorizeResponse = try await client.get(
            endpoint: CloudEndpoint.authorize(provider: provider, redirectURI: redirectURI)
        )
        guard let url = URL(string: resp.authorizeURL) else {
            throw CloudStorageError.invalidAuthorizeURL
        }
        return url
    }

    static func exchangeCode(_ code: String, provider: Provider, redirectURI: String = "echo://oauth/cloud") async throws {
        guard let client = DIContainer.shared.resolveAPIClient() else {
            throw CloudStorageError.clientUnavailable
        }
        struct Body: Encodable {
            let code: String
            let redirectURI: String
            enum CodingKeys: String, CodingKey {
                case code
                case redirectURI = "redirect_uri"
            }
        }
        struct Resp: Decodable { let status: String }
        let _: Resp = try await client.post(
            endpoint: CloudEndpoint.oauthCallback(provider: provider),
            body: Body(code: code, redirectURI: redirectURI)
        )
    }

    static func listFiles(provider: Provider) async throws -> [RemoteFile] {
        guard let client = DIContainer.shared.resolveAPIClient() else {
            throw CloudStorageError.clientUnavailable
        }
        let resp: FilesResponse = try await client.get(endpoint: CloudEndpoint.files(provider: provider))
        return resp.files
    }

    static func downloadFile(provider: Provider, fileID: String) async throws -> (Data, String) {
        guard let client = DIContainer.shared.resolveAPIClient() else {
            throw CloudStorageError.clientUnavailable
        }
        let data = try await client.getRaw(endpoint: CloudEndpoint.stream(provider: provider, fileID: fileID))
        return (data, mimeType(for: data))
    }

    private static func mimeType(for data: Data) -> String {
        if data.starts(with: [0xFF, 0xD8]) { return "image/jpeg" }
        if data.starts(with: [0x89, 0x50, 0x4E, 0x47]) { return "image/png" }
        if data.starts(with: [0x25, 0x50, 0x44, 0x46]) { return "application/pdf" }
        return "application/octet-stream"
    }
}

enum CloudStorageError: LocalizedError {
    case clientUnavailable
    case invalidAuthorizeURL

    var errorDescription: String? {
        switch self {
        case .clientUnavailable: return "API client unavailable"
        case .invalidAuthorizeURL: return "Invalid authorize URL from server"
        }
    }
}

enum CloudEndpoint: APIEndpoint {
    case listProviders
    case authorize(provider: CloudStorageIntegrationManager.Provider, redirectURI: String)
    case oauthCallback(provider: CloudStorageIntegrationManager.Provider)
    case files(provider: CloudStorageIntegrationManager.Provider)
    case stream(provider: CloudStorageIntegrationManager.Provider, fileID: String)

    var path: String {
        switch self {
        case .listProviders:
            return "/v3/integrations/cloud"
        case .authorize(let provider, let redirectURI):
            let enc = redirectURI.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? redirectURI
            return "/v3/integrations/cloud/\(provider.rawValue)?redirect_uri=\(enc)"
        case .oauthCallback(let provider):
            return "/v3/integrations/cloud/\(provider.rawValue)/callback"
        case .files(let provider):
            return "/v3/integrations/cloud/\(provider.rawValue)/files"
        case .stream(let provider, let fileID):
            return "/v3/integrations/cloud/\(provider.rawValue)/files/\(fileID)/stream"
        }
    }
}
#endif
