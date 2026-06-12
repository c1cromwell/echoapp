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
        let qrUseCase = DIContainer.shared.resolveQRContactExchangeUseCase() ?? QRContactExchangeUseCase()
        guard let scanned = qrUseCase.parseScannedPayload(raw) else {
            presentError("Unrecognized QR — scan an Echo profile code.")
            return
        }
        let (peerDID, username) = scanned

        guard let myDID = await CurrentUserSession.currentDID() else {
            presentError("Sign in before adding contacts.")
            return
        }
        if peerDID == myDID {
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
            var displayHandle = username.map { "@\($0)" } ?? peerDID
            if let profile = try? await discovery.resolveIdentity(did: peerDID) {
                if let u = profile.username, !u.isEmpty {
                    displayHandle = "@\(u)"
                } else if let name = profile.display_name, !name.isEmpty {
                    displayHandle = name
                }
            }

            _ = try await social.addContact(did: peerDID, addedVia: "qr_scan")

            createdConversation = await ContactThreadHelper.upsertDirectThread(
                peerDID: peerDID,
                displayName: displayHandle
            )

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
