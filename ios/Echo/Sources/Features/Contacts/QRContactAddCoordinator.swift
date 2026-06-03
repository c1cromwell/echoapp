#if os(iOS)
import Foundation
import Observation

/// Resolves a scanned profile QR, adds the contact, and prepares a DM thread.
@MainActor
@Observable
final class QRContactAddCoordinator {
    var isWorking = false
    var resultMessage: String?
    var resultIsError = false
    var createdConversation: StoredConversation?

    func handleScan(_ raw: String) async {
        guard let scanned = ScannedIdentityParser.parse(raw) else {
            presentError("Unrecognized QR — scan an Echo profile code.")
            return
        }

        guard let myDID = await CurrentUserSession.currentDID() else {
            presentError("Sign in before adding contacts.")
            return
        }
        if scanned.did == myDID {
            presentError("That’s your own profile code.")
            return
        }

        guard let client = DIContainer.shared.resolveAPIClient() else {
            presentError("Network unavailable.")
            return
        }

        isWorking = true
        defer { isWorking = false }

        let social = ContactSocialAPIClient(apiClient: client)
        let discovery = ContactDiscoveryAPIClient(apiClient: client)

        do {
            var displayHandle = scanned.username.map { "@\($0)" } ?? scanned.did
            if let profile = try? await discovery.resolveIdentity(did: scanned.did) {
                if let u = profile.username, !u.isEmpty {
                    displayHandle = "@\(u)"
                } else if let name = profile.display_name, !name.isEmpty {
                    displayHandle = name
                }
            }

            _ = try await social.addContact(did: scanned.did, addedVia: "qr_scan")

            let threadId = ConversationID.direct(localDID: myDID, peerDID: scanned.did)
            let conversation = StoredConversation(
                id: threadId,
                contactName: displayHandle,
                peerDID: scanned.did
            )
            ConversationStore.shared.upsert(conversation)
            createdConversation = conversation

            resultIsError = false
            resultMessage = "Added \(displayHandle). Open Messages to start chatting."
        } catch {
            presentError(error.localizedDescription)
        }
    }

    private func presentError(_ message: String) {
        resultIsError = true
        resultMessage = message
        createdConversation = nil
    }

    func reset() {
        resultMessage = nil
        resultIsError = false
        createdConversation = nil
    }
}
#endif
