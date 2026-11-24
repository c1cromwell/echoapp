import Foundation

// MARK: - Safety reports (POST /v1/safety/reports)
//
// Platform-neutral so the wire shape and the submit flow are unit-tested headlessly.
// Field names and allowed values mirror `internal/api/safety_handlers.go`.

enum SafetyReportReason: String, CaseIterable, Identifiable, Sendable {
    case spam
    case harassment
    case impersonation
    case other

    var id: String { rawValue }

    var label: String {
        switch self {
        case .spam: return "Spam"
        case .harassment: return "Harassment"
        case .impersonation: return "Impersonation"
        case .other: return "Something else"
        }
    }
}

enum SafetyReportContext: String, Sendable {
    case contact
    case message
    case channel
}

/// Body of a safety report. The optional fields are encoded only when present.
struct SafetyReportRequest: Encodable, Equatable, Sendable {
    static let maxDetailsLength = 1000

    let subjectDid: String
    let reason: String
    let context: String
    let messageId: String?
    let conversationId: String?
    let channelId: String?
    let details: String?

    enum CodingKeys: String, CodingKey {
        case subjectDid = "subject_did"
        case reason
        case context
        case messageId = "message_id"
        case conversationId = "conversation_id"
        case channelId = "channel_id"
        case details
    }

    init(
        subjectDID: String,
        reason: SafetyReportReason,
        context: SafetyReportContext,
        messageId: String? = nil,
        conversationId: String? = nil,
        channelId: String? = nil,
        details: String? = nil
    ) {
        self.subjectDid = subjectDID
        self.reason = reason.rawValue
        self.context = context.rawValue
        self.messageId = Self.present(messageId)
        self.conversationId = Self.present(conversationId)
        self.channelId = Self.present(channelId)
        self.details = Self.present(details).map { String($0.prefix(Self.maxDetailsLength)) }
    }

    private static func present(_ value: String?) -> String? {
        guard let trimmed = value?.trimmingCharacters(in: .whitespacesAndNewlines), !trimmed.isEmpty else {
            return nil
        }
        return trimmed
    }
}

/// What is being reported, and the words the report sheet uses for it.
struct SafetyReportTarget: Identifiable, Equatable, Sendable {
    let subjectDID: String
    let context: SafetyReportContext
    var messageId: String?
    var conversationId: String?
    var channelId: String?
    var title: String
    var prompt: String
    /// "Block & report": the subject is blocked before the report is sent.
    var alsoBlocks = false

    var id: String {
        [context.rawValue, subjectDID, messageId ?? "", conversationId ?? "", channelId ?? "", alsoBlocks ? "block" : ""]
            .joined(separator: "|")
    }

    func request(reason: SafetyReportReason, details: String?) -> SafetyReportRequest {
        SafetyReportRequest(
            subjectDID: subjectDID,
            reason: reason,
            context: context,
            messageId: messageId,
            conversationId: conversationId,
            channelId: channelId,
            details: details
        )
    }

    static func contact(did: String, conversationId: String? = nil, alsoBlocks: Bool = false) -> SafetyReportTarget {
        SafetyReportTarget(
            subjectDID: did,
            context: .contact,
            conversationId: conversationId,
            title: alsoBlocks ? "Block & report" : "Report contact",
            prompt: alsoBlocks
                ? "This person will be blocked and Echo will get a report. Tell Echo what is wrong."
                : "Tell Echo what is wrong with this contact. The report is stored for review.",
            alsoBlocks: alsoBlocks
        )
    }

    static func message(id: String, senderDID: String, conversationId: String) -> SafetyReportTarget {
        SafetyReportTarget(
            subjectDID: senderDID,
            context: .message,
            messageId: id,
            conversationId: conversationId,
            title: "Report message",
            prompt: "Tell Echo what is wrong with this message. Echo gets the message's id and who sent it, not its contents."
        )
    }

    static func channel(id: String, ownerDID: String) -> SafetyReportTarget {
        SafetyReportTarget(
            subjectDID: ownerDID,
            context: .channel,
            channelId: id,
            title: "Report channel",
            prompt: "Tell Echo what is wrong with this channel. The report names the channel and its owner."
        )
    }

    /// Groups use the `channel` context (the server has no separate one); the
    /// group is identified by its conversation id.
    static func group(id: String, ownerDID: String) -> SafetyReportTarget {
        SafetyReportTarget(
            subjectDID: ownerDID,
            context: .channel,
            conversationId: id.hasPrefix("group:") ? id : "group:\(id)",
            title: "Report group",
            prompt: "Tell Echo what is wrong with this group. The report names the group and its owner."
        )
    }
}

enum SafetyReportError: LocalizedError, Equatable {
    case notSignedIn
    case blockedButNotReported(String)

    var errorDescription: String? {
        switch self {
        case .notSignedIn:
            return "Sign in before sending a report."
        case .blockedButNotReported(let reason):
            return "They're blocked, but the report wasn't sent. \(reason)"
        }
    }
}

/// State for the report sheet: pick a reason, confirm, submit, then success or an error to retry.
@MainActor
final class SafetyReportViewModel: ObservableObject {
    enum Phase: Equatable {
        case editing
        case submitting
        case succeeded
        case failed(String)
    }

    let target: SafetyReportTarget
    @Published var selectedReason: SafetyReportReason?
    @Published var details = ""
    @Published private(set) var phase: Phase = .editing

    private let submit: (SafetyReportRequest) async throws -> Void
    private let block: (String) async throws -> Void
    private var didBlock = false

    init(
        target: SafetyReportTarget,
        submit: @escaping (SafetyReportRequest) async throws -> Void,
        block: @escaping (String) async throws -> Void = { _ in }
    ) {
        self.target = target
        self.submit = submit
        self.block = block
    }

    var canSubmit: Bool { selectedReason != nil && phase != .submitting && phase != .succeeded }

    var errorMessage: String? {
        if case .failed(let message) = phase { return message }
        return nil
    }

    var successMessage: String {
        target.alsoBlocks
            ? "They're blocked and Echo saved your report."
            : "Echo saved this report. Blocking is separate and still available."
    }

    func submitReport() async {
        guard canSubmit, let reason = selectedReason else { return }
        phase = .submitting
        do {
            if target.alsoBlocks, !didBlock {
                try await block(target.subjectDID)
                didBlock = true
            }
        } catch {
            phase = .failed(error.localizedDescription)
            return
        }
        do {
            try await submit(target.request(reason: reason, details: details))
            phase = .succeeded
        } catch {
            let reasonText = error.localizedDescription
            phase = .failed(
                didBlock
                    ? (SafetyReportError.blockedButNotReported(reasonText).errorDescription ?? reasonText)
                    : reasonText
            )
        }
    }
}

#if os(iOS)

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

    /// Legacy opaque-code URL. New shares use `InviteHandle.shareURL` (`echo://invite?u=`).
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
    let barAllowed: Bool?

    var id: String { contactDid ?? UUID().uuidString }
}

// MARK: - Client

struct ContactSocialAPIClient: Sendable {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func searchUsername(_ handle: String) async throws -> [UsernameSearchHit] {
        let trimmed = InviteHandle.normalize(handle)
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

    func allowBar(did: String) async throws -> RemoteContact {
        struct Body: Encodable { let contactDid: String }
        return try await apiClient.post(endpoint: ContactsEndpoint.bar, body: Body(contactDid: did))
    }

    func revokeBar(did: String) async throws {
        struct Body: Encodable { let contactDid: String }
        struct Status: Decodable { let revoked: Bool? }
        let _: Status = try await apiClient.post(endpoint: ContactsEndpoint.barRevoke, body: Body(contactDid: did))
    }

    func listBar() async throws -> [RemoteContact] {
        let decoded: ContactListResponse = try await apiClient.get(endpoint: ContactsEndpoint.bar)
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

    func reportContact(did: String, reason: String) async throws {
        guard let reason = SafetyReportReason(rawValue: reason) else {
            throw APIError.badRequest
        }
        try await submitReport(SafetyReportRequest(subjectDID: did, reason: reason, context: .contact))
    }

    /// Files a safety report about a contact, message, channel or group.
    func submitReport(_ request: SafetyReportRequest) async throws {
        struct Ack: Decodable { let id: String }
        let _: Ack = try await apiClient.post(endpoint: SafetyEndpoint.report, body: request)
    }

    func blockContact(did: String) async throws {
        struct Body: Encodable { let contactDid: String }
        struct Status: Decodable { let status: String? }
        let _: Status = try await apiClient.post(
            endpoint: ContactsEndpoint.block,
            body: Body(contactDid: did)
        )
        #if os(iOS)
        // The server cannot enforce a block on a sealed message, because it cannot see who sent
        // it. So this device drops what they send, and the delivery key they hold is replaced.
        BlockedSenders.set(did, blocked: true)
        DeliveryKeys.shared.forget(peerDID: did)
        await DeliveryKeyRegistration.rotateAfterBlocking(apiClient: apiClient)
        #endif
    }

    func unblockContact(did: String) async throws {
        struct Body: Encodable { let contactDid: String }
        struct Status: Decodable { let status: String? }
        let _: Status = try await apiClient.post(
            endpoint: ContactsEndpoint.unblock,
            body: Body(contactDid: did)
        )
        #if os(iOS)
        BlockedSenders.set(did, blocked: false)
        #endif
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
