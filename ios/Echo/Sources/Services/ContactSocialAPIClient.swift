#if os(iOS)
import Foundation

// MARK: - Models (WO-222)

struct UsernameSearchHit: Codable, Sendable, Identifiable {
    let did: String
    let username: String
    let tier: Int?

    var id: String { did }
}

struct InviteLinkResponse: Codable, Sendable {
    let code: String
    let creatorDid: String?
    let accepted: Bool?
    let createdAt: String?
    let expiresAt: String?

    enum CodingKeys: String, CodingKey {
        case code
        case creatorDid
        case accepted
        case createdAt
        case expiresAt
    }

    var shareURL: URL? {
        URL(string: "echo://invite?code=\(code.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? code)")
    }
}

struct ContactListResponse: Codable, Sendable {
    let contacts: [RemoteContact]?
    let count: Int?
}

struct RemoteContact: Codable, Sendable, Identifiable {
    let contactDid: String?
    let ownerDid: String?
    let addedVia: String?
    let trustBadge: String?

    var id: String { contactDid ?? UUID().uuidString }
}

// MARK: - Client

struct ContactSocialAPIClient: Sendable {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func searchUsername(_ handle: String) async throws -> [UsernameSearchHit] {
        let trimmed = handle.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return [] }
        return try await apiClient.get(endpoint: ContactsEndpoint.search(handle: trimmed))
    }

    func createInviteLink() async throws -> InviteLinkResponse {
        struct Empty: Encodable {}
        return try await apiClient.post(endpoint: ContactsEndpoint.invite, body: Empty())
    }

    func acceptInvite(code: String) async throws -> InviteLinkResponse {
        struct Body: Encodable { let code: String }
        return try await apiClient.post(endpoint: ContactsEndpoint.verifyInvite, body: Body(code: code))
    }

    func listContacts() async throws -> [RemoteContact] {
        let decoded: ContactListResponse = try await apiClient.get(endpoint: ContactsEndpoint.list)
        return decoded.contacts ?? []
    }

    func addContact(did: String, addedVia: String = "psi_discovery") async throws -> RemoteContact {
        struct Body: Encodable {
            let contactDid: String
            let addedVia: String
        }
        return try await apiClient.post(
            endpoint: ContactsEndpoint.add,
            body: Body(contactDid: did, addedVia: addedVia)
        )
    }

    func blockContact(did: String) async throws {
        struct Body: Encodable { let contactDid: String }
        struct Status: Decodable { let status: String? }
        let _: Status = try await apiClient.post(
            endpoint: ContactsEndpoint.block,
            body: Body(contactDid: did)
        )
    }
}
#endif
