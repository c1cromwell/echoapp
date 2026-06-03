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
            contacts = remote.compactMap { row in
                guard let did = row.contactDid, !did.isEmpty else { return nil }
                let label = row.trustBadge?.capitalized ?? "Contact"
                return ContactModel(
                    id: did,
                    name: ContactThreadHelper.truncatedDID(did),
                    username: did,
                    trustLevel: label
                )
            }
            .sorted { $0.name < $1.name }
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
#endif
