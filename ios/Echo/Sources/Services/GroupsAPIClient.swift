#if os(iOS)
import Foundation

// MARK: - REST models (match internal/api/v3_handlers.go group routes)

struct GroupCreateRequest: Codable, Sendable {
    let groupId: String
    let groupType: String
    let name: String
    let description: String

    enum CodingKeys: String, CodingKey {
        case groupId, groupType, name, description
    }
}

struct GroupMemberAddRequest: Codable, Sendable {
    let groupId: String
    let memberDid: String

    enum CodingKeys: String, CodingKey {
        case groupId
        case memberDid
    }
}

struct GroupMemberAddResponse: Codable, Sendable {
    let requiresRekey: Bool

    enum CodingKeys: String, CodingKey {
        case requiresRekey = "requires_rekey"
    }
}

struct GroupKeyDistributeRequest: Codable, Sendable {
    let groupId: String
    let version: Int
    let distributedBy: String
    let packages: [GroupKeyPackageWire]

    enum CodingKeys: String, CodingKey {
        case groupId = "group_id"
        case version
        case distributedBy = "distributed_by"
        case packages
    }
}

struct GroupKeyPackageWire: Codable, Sendable {
    let to: String
    let payload: Data
}

struct GroupKeyDistributeResponse: Codable, Sendable {
    let groupId: String
    let version: Int
    let delivered: Int
    let total: Int

    enum CodingKeys: String, CodingKey {
        case groupId = "group_id"
        case version, delivered, total
    }
}

struct GroupMemberWire: Codable, Sendable, Identifiable {
    let memberId: String
    let groupId: String
    let displayName: String?
    let role: String

    var id: String { memberId }

    var isAdmin: Bool {
        role == "owner" || role == "admin"
    }
}

struct GroupMembersListResponse: Codable, Sendable {
    let groupId: String
    let members: [GroupMemberWire]
    let count: Int
}

// MARK: - Client

protocol GroupsAPIClient: Sendable {
    func createGroup(groupId: String, name: String, type: String) async throws
    func listMembers(groupId: String) async throws -> [GroupMemberWire]
    func addMember(groupId: String, memberDid: String) async throws -> Bool
    func removeMember(groupId: String, memberDid: String) async throws -> Bool
    func distributeKeys(
        groupId: String,
        version: Int,
        distributedBy: String,
        packages: [GroupKeyManager.KeyPackage]
    ) async throws -> GroupKeyDistributeResponse
    func muteMember(groupId: String, memberDid: String, durationHours: Int) async throws
    func banMember(groupId: String, memberDid: String) async throws
}

#if os(iOS)

enum GroupsEndpoint: APIEndpoint {
    case create
    case addMember
    case removeMember
    case distributeKeys
    case listMembers(groupId: String)
    case muteMember
    case banMember

    var path: String {
        switch self {
        case .create: return "/v3/groups/create"
        case .addMember: return "/v3/groups/members/add"
        case .removeMember: return "/v3/groups/members/remove"
        case .distributeKeys: return "/v3/groups/key/distribute"
        case .listMembers(let groupId):
            return "/v3/groups/members?groupId=\(groupId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? groupId)"
        case .muteMember: return "/v3/groups/members/mute"
        case .banMember: return "/v3/groups/members/ban"
        }
    }
}

actor LiveGroupsAPIClient: GroupsAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func createGroup(groupId: String, name: String, type: String) async throws {
        let body = GroupCreateRequest(
            groupId: groupId,
            groupType: type,
            name: name,
            description: ""
        )
        let _: EmptyResponse = try await apiClient.post(endpoint: GroupsEndpoint.create, body: body)
    }

    func listMembers(groupId: String) async throws -> [GroupMemberWire] {
        let resp: GroupMembersListResponse = try await apiClient.get(endpoint: GroupsEndpoint.listMembers(groupId: groupId))
        return resp.members
    }

    func addMember(groupId: String, memberDid: String) async throws -> Bool {
        let body = GroupMemberAddRequest(groupId: groupId, memberDid: memberDid)
        let resp: GroupMemberAddResponse = try await apiClient.post(endpoint: GroupsEndpoint.addMember, body: body)
        return resp.requiresRekey
    }

    func removeMember(groupId: String, memberDid: String) async throws -> Bool {
        struct Body: Codable { let groupId: String; let memberId: String }
        struct Resp: Codable { let requiresRekey: Bool; enum CodingKeys: String, CodingKey { case requiresRekey = "requires_rekey" } }
        let resp: Resp = try await apiClient.post(
            endpoint: GroupsEndpoint.removeMember,
            body: Body(groupId: groupId, memberId: memberDid)
        )
        return resp.requiresRekey
    }

    func distributeKeys(
        groupId: String,
        version: Int,
        distributedBy: String,
        packages: [GroupKeyManager.KeyPackage]
    ) async throws -> GroupKeyDistributeResponse {
        let wire = packages.map { GroupKeyPackageWire(to: $0.recipientDID, payload: $0.encryptedKey) }
        let body = GroupKeyDistributeRequest(
            groupId: groupId,
            version: version,
            distributedBy: distributedBy,
            packages: wire
        )
        return try await apiClient.post(endpoint: GroupsEndpoint.distributeKeys, body: body)
    }

    func muteMember(groupId: String, memberDid: String, durationHours: Int) async throws {
        struct Body: Codable { let groupId: String; let memberId: String; let durationHours: Int
            enum CodingKeys: String, CodingKey { case groupId; case memberId; case durationHours = "duration_hours" }
        }
        let _: [String: String] = try await apiClient.post(
            endpoint: GroupsEndpoint.muteMember,
            body: Body(groupId: groupId, memberId: memberDid, durationHours: durationHours)
        )
    }

    func banMember(groupId: String, memberDid: String) async throws {
        struct Body: Codable { let groupId: String; let memberId: String }
        let _: [String: String] = try await apiClient.post(
            endpoint: GroupsEndpoint.banMember,
            body: Body(groupId: groupId, memberId: memberDid)
        )
    }
}

private struct EmptyResponse: Decodable {}
#endif
#endif
