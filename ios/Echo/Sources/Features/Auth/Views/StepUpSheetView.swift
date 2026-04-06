import SwiftUI

struct StepUpSheetView: View {
    let action: StepUpAction
    let passkeyManager: PasskeyManagerProtocol
    let apiClient: AuthAPIClientProtocol
    let tokenManager: TokenManager
    let onVerified: (String) -> Void

    @Environment(\.dismiss) var dismiss
    @State private var isAuthenticating = false
    @State private var errorMessage: String?

    var body: some View {
        VStack(spacing: 24) {
            // Drag indicator
            Capsule()
                .fill(Color.echoGray300)
                .frame(width: 36, height: 5)
                .padding(.top, 8)

            VStack(spacing: 8) {
                Text("Verify it's you")
                    .typographyStyle(.h3, color: .echoPrimaryText)

                Text("This action requires additional verification.")
                    .typographyStyle(.body, color: .echoSecondaryText)
                    .multilineTextAlignment(.center)

                Text(action.displayTitle)
                    .typographyStyle(.bodyLarge, color: .echoPrimary)
                    .padding(.top, 4)
            }

            // Biometric verify button
            EchoButton(
                "Verify with Face ID",
                style: .primary,
                size: .large,
                icon: Image(systemName: "faceid"),
                isLoading: isAuthenticating,
                isDisabled: isAuthenticating,
                action: { Task { await performStepUp() } }
            )

            // Error
            if let error = errorMessage {
                Text(error)
                    .typographyStyle(.caption, color: .red)
                    .multilineTextAlignment(.center)
            }

            // Cancel
            Button("Cancel") { dismiss() }
                .typographyStyle(.body, color: .echoSecondaryText)

            Spacer()
        }
        .padding(.horizontal, 24)
        .presentationDetents([.fraction(0.45)])
        .presentationDragIndicator(.hidden)
    }

    private func performStepUp() async {
        isAuthenticating = true
        errorMessage = nil
        defer { isAuthenticating = false }

        do {
            // Get challenge
            let challenge = try await apiClient.getLoginChallenge()

            // Passkey assertion
            let assertion = try await passkeyManager.authenticateWithPasskey(
                challenge: challenge.challengeData
            )

            // Request elevated token
            let token = try await tokenManager.getValidAccessToken()
            let response = try await apiClient.requestStepUp(
                action: action.rawValue,
                assertion: assertion,
                token: token
            )

            onVerified(response.elevatedToken)
            dismiss()
        } catch is CancellationError {
            // User cancelled — not an error
        } catch {
            errorMessage = "Verification failed. Please try again."
        }
    }
}
