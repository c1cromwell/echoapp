#if os(iOS)
import Foundation

/// OAuth token registry for cloud file pickers (WO-46 subset).
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

    private static let tokenPrefix = "echo.cloud.token."

    static func connectedProviders() -> [Provider] {
        Provider.allCases.filter { provider in
            (try? KeychainManager.shared.retrieve(key: tokenKey(provider)))?.isEmpty == false
        }
    }

    static func saveToken(_ token: String, provider: Provider) async throws {
        try await KeychainManager.shared.store(value: token, key: tokenKey(provider))
    }

    static func revoke(provider: Provider) async throws {
        try await KeychainManager.shared.delete(key: tokenKey(provider))
    }

    static func authorizationURL(provider: Provider, redirectURI: String = "echo://oauth/cloud") -> URL {
        var components = URLComponents(url: EchoAPIBaseURL.url(path: "/v3/integrations/cloud/\(provider.rawValue)"), resolvingAgainstBaseURL: false)!
        components.queryItems = [URLQueryItem(name: "redirect_uri", value: redirectURI)]
        return components.url ?? EchoAPIBaseURL.url(path: "/v3/integrations/cloud/\(provider.rawValue)")
    }

    private static func tokenKey(_ provider: Provider) -> String {
        tokenPrefix + provider.rawValue
    }
}
#endif
