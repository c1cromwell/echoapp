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
    let blocked: Bool?

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

    func unblockContact(did: String) async throws {
        struct Body: Encodable { let contactDid: String }
        struct Status: Decodable { let status: String? }
        let _: Status = try await apiClient.post(
            endpoint: ContactsEndpoint.unblock,
            body: Body(contactDid: did)
        )
    }

    func listBlockedContacts() async throws -> [RemoteContact] {
        struct Response: Decodable { let contacts: [RemoteContact]? }
        let decoded: Response = try await apiClient.get(endpoint: ContactsEndpoint.blocked)
        return decoded.contacts ?? []
    }

    struct ProfilePrivacySettings: Codable, Sendable {
        var showLastSeen: String
        var showOnlineStatus: String
        var showProfilePicture: String
        var showStatusMessage: String
        var allowGroupInvites: String
        var allowCalls: String
        var showTrustScore: String
    }

    struct UserProfile: Codable, Sendable {
        let did: String
        let displayName: String?
        let username: String?
        let avatarURL: String?
        let bio: String?
        let statusMessage: String?
        let trustTier: Int?
        let isVerified: Bool?
        let isContact: Bool?
        let isBlocked: Bool?
    }

    struct ContactPrivacyOverrideBody: Codable, Sendable {
        let notificationsEnabled: Bool?
        let disappearingEnabled: Bool?
    }

    func fetchOwnProfile() async throws -> (UserProfile, ProfilePrivacySettings?) {
        struct Response: Decodable {
            let profile: UserProfile
            let privacy: ProfilePrivacySettings?
        }
        let decoded: Response = try await apiClient.get(endpoint: ProfileEndpoint.own)
        return (decoded.profile, decoded.privacy)
    }

    func fetchProfile(did: String) async throws -> UserProfile {
        try await apiClient.get(endpoint: ProfileEndpoint.view(did: did))
    }

    func updateProfile(displayName: String?, bio: String?, statusMessage: String?) async throws -> UserProfile {
        struct Body: Encodable {
            let displayName: String?
            let bio: String?
            let statusMessage: String?
        }
        return try await apiClient.patch(
            endpoint: ProfileEndpoint.own,
            body: Body(displayName: displayName, bio: bio, statusMessage: statusMessage)
        )
    }

    func updatePrivacy(_ settings: ProfilePrivacySettings) async throws -> ProfilePrivacySettings {
        try await apiClient.patch(endpoint: ProfileEndpoint.privacy, body: settings)
    }

    func updateContactPrivacy(peerDID: String, notificationsEnabled: Bool?, disappearingEnabled: Bool?) async throws {
        struct Body: Encodable {
            let peerDid: String
            let override: ContactPrivacyOverrideBody
        }
        struct Status: Decodable { let status: String? }
        let _: Status = try await apiClient.patch(
            endpoint: ContactsEndpoint.contactPrivacy,
            body: Body(
                peerDid: peerDID,
                override: ContactPrivacyOverrideBody(
                    notificationsEnabled: notificationsEnabled,
                    disappearingEnabled: disappearingEnabled
                )
            )
        )
    }

    struct ContactRelationship: Decodable, Sendable {
        let peer_did: String?
        let mutual_groups: [MutualGroup]?
        let mutual_groups_count: Int?
        let mutual_contacts: [MutualContact]?
        let mutual_contacts_count: Int?
        let blocked_by_me: Bool?
        let blocked_me: Bool?
        let is_blocked: Bool?
    }

    struct MutualGroup: Decodable, Sendable, Identifiable {
        let groupId: String?
        let name: String?
        let type: String?
        let member_count: Int?

        var id: String { groupId ?? UUID().uuidString }
    }

    struct MutualContact: Decodable, Sendable, Identifiable {
        let did: String
        let username: String?

        var id: String { did }
    }

    func fetchRelationship(peerDID: String) async throws -> ContactRelationship {
        try await apiClient.get(endpoint: ContactsEndpoint.relationship(peerDID: peerDID))
    }
}
#endif
