#if os(iOS)
import Foundation
import Observation

@MainActor
@Observable
final class ContactsListViewModel {
    var contacts: [ContactModel] = []
    var isLoading = false
    var errorMessage: String?

    private let socialAPI: ContactSocialAPIClient?

    init(socialAPI: ContactSocialAPIClient? = nil) {
        if let socialAPI {
            self.socialAPI = socialAPI
        } else if let resolved = DIContainer.shared.resolveContactSocialAPI() {
            self.socialAPI = resolved
        } else if let client = DIContainer.shared.resolveAPIClient() {
            self.socialAPI = ContactSocialAPIClient(apiClient: client)
        } else {
            self.socialAPI = nil
        }
    }

    func refresh() async {
        guard let socialAPI else {
            errorMessage = "Sign in required to load contacts."
            return
        }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            let remote = try await socialAPI.listContacts()
            ContactTrustIndex.shared.ingestRemoteContacts(remote)
            let discoveryClient: ContactDiscoveryAPIClient? = {
                guard let client = DIContainer.shared.resolveAPIClient() else { return nil }
                return ContactDiscoveryAPIClient(apiClient: client)
            }()

            var models: [ContactModel] = []
            for row in remote {
                guard let did = row.contactDid, !did.isEmpty else { continue }
                if row.blocked == true { continue }
                let badge = row.trustBadge?.capitalized ?? "Contact"
                var name = ContactThreadHelper.truncatedDID(did)
                var username = ContactThreadHelper.truncatedDID(did)
                if let discoveryClient,
                   let profile = try? await discoveryClient.resolveIdentity(did: did),
                   let handle = profile.username, !handle.isEmpty {
                    username = "@\(handle)"
                    name = username
                }
                models.append(ContactModel(id: did, name: name, username: username, trustLevel: badge))
            }
            contacts = models.sorted { $0.name < $1.name }
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
#endif
