#if os(iOS)
import SwiftUI

/// Accept `echo://invite?u=` (`@username`) or legacy `echo://invite?code=` (WO-222).
struct AcceptInviteSheet: View {
    @Environment(\.dismiss) private var dismiss

    let inviteCode: String
    var onAccepted: (() -> Void)?

    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var success = false

    private var isUsernameInvite: Bool { InviteHandle.isUsernameToken(inviteCode) }
    private var displayedHandle: String { InviteHandle.display(inviteCode) }

    var body: some View {
        NavigationStack {
            VStack(spacing: Spacing.xl.rawValue) {
                if success {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 48))
                        .foregroundStyle(Color.echoTrustGreen)
                    Text("Added to contacts")
                        .typographyStyle(.h4, color: .echoInk)
                    Text(isUsernameInvite
                         ? "\(displayedHandle) is in your contacts."
                         : "They've been added to your contacts.")
                        .multilineTextAlignment(.center)
                        .typographyStyle(.body, color: .echoInk55)
                } else {
                    Text(isUsernameInvite ? "Add \(displayedHandle)?" : "Accept contact invite?")
                        .typographyStyle(.h4, color: .echoInk)
                    Text(isUsernameInvite ? displayedHandle : "Code: \(inviteCode)")
                        .font(Font.echomono(13, weight: .medium))
                        .foregroundStyle(Color.echoInk55)
                    if let errorMessage {
                        Text(errorMessage)
                            .typographyStyle(.caption, color: .echoError)
                    }
                    Button {
                        Task { await accept() }
                    } label: {
                        if isLoading {
                            ProgressView()
                        } else {
                            Text(isUsernameInvite ? "Add contact" : "Accept invite")
                                .frame(maxWidth: .infinity)
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(isLoading)
                }
            }
            .padding(Spacing.lg.rawValue)
            .navigationTitle("Contact invite")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(success ? "Done" : "Cancel") {
                        if success { onAccepted?() }
                        dismiss()
                    }
                }
            }
        }
    }

    private func accept() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        guard let client = DIContainer.shared.resolveAPIClient() else {
            errorMessage = "Sign in required"
            return
        }
        let social = ContactSocialAPIClient(apiClient: client)
        do {
            let useCase = InviteLinkUseCase(client: social)
            try await useCase.acceptInvite(code: inviteCode)
            success = true
        } catch InviteAcceptError.userNotFound {
            errorMessage = isUsernameInvite
                ? "No one found with \(displayedHandle)."
                : "Invite is invalid or expired."
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
#endif
