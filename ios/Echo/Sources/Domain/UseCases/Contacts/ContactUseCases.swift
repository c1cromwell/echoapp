#if os(iOS)
import Foundation

// MARK: - WO-39 contact use cases (Domain/UseCases/Contacts/)

/// Hashes device contacts and runs OPRF-PSI discovery (WO-221).
struct ContactDiscoveryUseCase: Sendable {
    private let service: ContactDiscoveryService

    init(service: ContactDiscoveryService) {
        self.service = service
    }

    func discoverContacts() async throws -> [DiscoveredContact] {
        try await service.discoverFromDeviceContacts()
    }
}

/// Generates and parses identity share payloads for zero-server QR exchange.
struct QRContactExchangeUseCase: Sendable {
    func shareURL(did: String, username: String) -> URL? {
        CurrentUserSession.identityShareURL(did: did, username: username)
    }

    func parseScannedPayload(_ raw: String) -> (did: String, username: String?)? {
        if let parsed = ScannedIdentityParser.parse(raw) {
            return (parsed.did, parsed.username)
        }
        return nil
    }
}

/// Referral invite links. Public share is `@username` (`echo://invite?u=`).
/// Opaque `code=` accept remains for links already in the wild (WO-222).
enum InviteAcceptError: Error {
    case userNotFound
}

struct InviteLinkUseCase: Sendable {
    private let client: ContactSocialAPIClient

    init(client: ContactSocialAPIClient) {
        self.client = client
    }

    func generateInviteLink() async throws -> URL {
        let username = await CurrentUserSession.currentUsername()
        guard let url = InviteHandle.shareURL(username: username) else {
            throw URLError(.badURL)
        }
        return url
    }

    func acceptInvite(code: String) async throws {
        if InviteHandle.isUsernameToken(code) {
            try await acceptUsername(InviteHandle.normalize(code))
            return
        }
        _ = try await client.acceptInvite(code: code)
    }

    func acceptUsername(_ handle: String) async throws {
        let normalized = InviteHandle.normalize(handle)
        guard !normalized.isEmpty else { throw URLError(.badURL) }
        let hits = try await client.searchUsername(normalized)
        let match = hits.first {
            InviteHandle.normalize($0.username).caseInsensitiveCompare(normalized) == .orderedSame
        } ?? hits.first
        guard let match else { throw InviteAcceptError.userNotFound }
        _ = try await client.addContact(did: match.did, addedVia: "username_invite")
    }
}

/// Public handle search (WO-222).
struct UsernameSearchUseCase: Sendable {
    private let client: ContactSocialAPIClient

    init(client: ContactSocialAPIClient) {
        self.client = client
    }

    func search(handle: String) async throws -> [UsernameSearchHit] {
        try await client.searchUsername(handle)
    }
}

/// Add a contact after PSI / QR / search (WO-39).
struct AddContactUseCase: Sendable {
    private let client: ContactSocialAPIClient

    init(client: ContactSocialAPIClient) {
        self.client = client
    }

    func add(did: String, addedVia: String) async throws {
        _ = try await client.addContact(did: did, addedVia: addedVia)
    }
}

/// Block a contact (WO-39 / WO-190).
struct BlockContactUseCase: Sendable {
    private let client: ContactSocialAPIClient

    init(client: ContactSocialAPIClient) {
        self.client = client
    }

    func block(did: String) async throws {
        try await client.blockContact(did: did)
    }
}

/// Unblock a contact (WO-190).
struct UnblockContactUseCase: Sendable {
    private let client: ContactSocialAPIClient

    init(client: ContactSocialAPIClient) {
        self.client = client
    }

    func unblock(did: String) async throws {
        try await client.unblockContact(did: did)
    }
}

/// List blocked contacts (WO-39 / WO-190).
struct ListBlockedContactsUseCase: Sendable {
    private let client: ContactSocialAPIClient

    init(client: ContactSocialAPIClient) {
        self.client = client
    }

    func list() async throws -> [RemoteContact] {
        try await client.listBlockedContacts()
    }
}

/// Sync profile privacy to the server (WO-187 / WO-190).
struct SyncProfilePrivacyUseCase: Sendable {
    private let client: ContactSocialAPIClient

    init(client: ContactSocialAPIClient) {
        self.client = client
    }

    func sync(_ settings: EnhancedPrivacySettings) async throws {
        let payload = ContactSocialAPIClient.ProfilePrivacySettings(
            showLastSeen: settings.showLastSeen ? "everyone" : "nobody",
            showOnlineStatus: settings.showOnlineStatus,
            showProfilePicture: settings.showProfilePicture ? "everyone" : "nobody",
            showStatusMessage: settings.showStatusMessage ? "contacts" : "nobody",
            allowGroupInvites: settings.whoCanMessage == "everyone" ? "everyone" : "contacts",
            allowCalls: settings.whoCanCall == "everyone" ? "everyone" : "contacts",
            showTrustScore: settings.showTrustScore
        )
        _ = try await client.updatePrivacy(payload)
    }
}

/// Per-contact privacy overrides (WO-39).
struct UpdateContactPrivacyUseCase: Sendable {
    private let client: ContactSocialAPIClient

    init(client: ContactSocialAPIClient) {
        self.client = client
    }

    func update(peerDID: String, notificationsEnabled: Bool, disappearingEnabled: Bool) async throws {
        try await client.updateContactPrivacy(
            peerDID: peerDID,
            notificationsEnabled: notificationsEnabled,
            disappearingEnabled: disappearingEnabled
        )
    }
}
#endif
