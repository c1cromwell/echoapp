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
        if raw.hasPrefix("did:") {
            return (raw, nil)
        }
        guard let url = URL(string: raw),
              url.scheme == "echo",
              url.host == "profile",
              let components = URLComponents(url: url, resolvingAgainstBaseURL: false) else {
            return nil
        }
        let did = components.queryItems?.first(where: { $0.name == "did" })?.value
        let username = components.queryItems?.first(where: { $0.name == "u" })?.value
        guard let did, !did.isEmpty else { return nil }
        return (did, username)
    }
}

/// Referral invite links via Contacts Service (WO-222).
struct InviteLinkUseCase: Sendable {
    private let client: ContactSocialAPIClient

    init(client: ContactSocialAPIClient) {
        self.client = client
    }

    func generateInviteLink() async throws -> URL {
        let response = try await client.createInviteLink()
        guard let url = response.shareURL else {
            throw URLError(.badURL)
        }
        return url
    }

    func acceptInvite(code: String) async throws {
        _ = try await client.acceptInvite(code: code)
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
#endif
