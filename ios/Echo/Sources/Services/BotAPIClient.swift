#if os(iOS)
import Foundation

enum BotEndpoint: APIEndpoint {
    case catalog
    case installed
    case install
    case uninstall(botDID: String)
    case setActive(botDID: String)

    private static func encode(_ did: String) -> String {
        did.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? did
    }

    var path: String {
        switch self {
        case .catalog:
            return "/v3/bots/catalog"
        case .installed:
            return "/v3/bots/installed"
        case .install:
            return "/v3/bots/install"
        case .uninstall(let botDID):
            return "/v3/bots/\(Self.encode(botDID))/install"
        case .setActive(let botDID):
            return "/v3/bots/\(Self.encode(botDID))/install"
        }
    }
}

struct BotManifestDTO: Codable, Sendable, Identifiable {
    let botDID: String
    let name: String
    let description: String
    let version: String
    let requiredPermissions: [String]
    let trustScore: Int

    var id: String { botDID }

    enum CodingKeys: String, CodingKey {
        case botDID = "bot_did"
        case name, description, version
        case requiredPermissions = "required_permissions"
        case trustScore = "trust_score"
    }
}

struct BotInstallationDTO: Codable, Sendable {
    let botDID: String
    let userDID: String
    let grantedPermissions: [String]
    let installedAt: Date
    let active: Bool

    enum CodingKeys: String, CodingKey {
        case botDID = "bot_did"
        case userDID = "user_did"
        case grantedPermissions = "granted_permissions"
        case installedAt = "installed_at"
        case active
    }
}

struct BotCatalogResponse: Codable, Sendable {
    let bots: [BotManifestDTO]
}

struct BotInstalledResponse: Codable, Sendable {
    let installed: [BotInstallationDTO]
}

struct BotInstallRequest: Codable, Sendable {
    let botDID: String
    let permissions: [String]

    enum CodingKeys: String, CodingKey {
        case botDID = "bot_did"
        case permissions
    }
}

struct BotInstallResponse: Codable, Sendable {
    let installation: BotInstallationDTO
}

struct BotActiveRequest: Codable, Sendable {
    let active: Bool
}

/// Bot marketplace + install API (Stage 4 / WO-11).
struct BotAPIClient: Sendable {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func fetchCatalog() async throws -> [BotManifestDTO] {
        let response: BotCatalogResponse = try await apiClient.get(endpoint: BotEndpoint.catalog)
        return response.bots
    }

    func fetchInstalled() async throws -> [BotInstallationDTO] {
        let response: BotInstalledResponse = try await apiClient.get(endpoint: BotEndpoint.installed)
        return response.installed
    }

    func install(botDID: String, permissions: [String]) async throws -> BotInstallationDTO {
        let response: BotInstallResponse = try await apiClient.post(
            endpoint: BotEndpoint.install,
            body: BotInstallRequest(botDID: botDID, permissions: permissions)
        )
        return response.installation
    }

    func uninstall(botDID: String) async throws {
        struct Removed: Codable { let removed: Bool }
        _ = try await apiClient.delete(endpoint: BotEndpoint.uninstall(botDID: botDID)) as Removed
    }

    func setActive(botDID: String, active: Bool) async throws -> BotInstallationDTO {
        let response: BotInstallResponse = try await apiClient.patch(
            endpoint: BotEndpoint.setActive(botDID: botDID),
            body: BotActiveRequest(active: active)
        )
        return response.installation
    }
}
#endif
