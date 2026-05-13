#if os(iOS)
// Features/Auth/Views/StepUpSheetView.swift
// Sheet for sensitive account-level actions (revoke device, delete account, etc.)
// Supports Face ID / Passkey / Device Passcode — method selector at top.
// Local auth (Face ID / Passkey) then fetches an elevated backend token.
// Device Passcode uses StepUpAuthManager for local verification only.

import SwiftUI

struct StepUpSheetView: View {
    let action: StepUpAction
    let passkeyManager: PasskeyManagerProtocol
    let apiClient: AuthAPIClientProtocol
    let tokenManager: TokenManager
    let onVerified: (String) -> Void

    @Environment(\.dismiss) var dismiss
    @State private var selectedMethod: StepUpMethod = StepUpAuthManager.shared.preferredMethod
    @State private var isAuthenticating = false
    @State private var errorMessage: String?

    var body: some View {
        VStack(spacing: 24) {
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

            // Method selector
            HStack(spacing: 0) {
                ForEach(StepUpMethod.allCases.filter(\.isAvailable), id: \.rawValue) { method in
                    Button {
                        selectedMethod = method
                    } label: {
                        VStack(spacing: 4) {
                            Image(systemName: method.systemIcon)
                                .font(.system(size: 18))
                            Text(method.displayName)
                                .font(.system(size: 11, weight: .medium))
                        }
                        .foregroundStyle(selectedMethod == method ? Color.echoSignal : Color.echoInk55)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 10)
                        .background(
                            selectedMethod == method
                                ? Color.echoSignal.opacity(0.08)
                                : Color.clear,
                            in: RoundedRectangle(cornerRadius: 10)
                        )
                    }
                }
            }
            .padding(4)
            .background(Color.echoPaperDim, in: RoundedRectangle(cornerRadius: 14))

            // Verify button
            EchoButton(
                "Verify with \(selectedMethod.displayName)",
                style: .primary,
                size: .large,
                isLoading: isAuthenticating,
                isDisabled: isAuthenticating,
                icon: Image(systemName: selectedMethod.systemIcon),
                action: { Task { await performStepUp() } }
            )

            if let error = errorMessage {
                Text(error)
                    .typographyStyle(.caption, color: .red)
                    .multilineTextAlignment(.center)
            }

            Button("Cancel") { dismiss() }
                .typographyStyle(.body, color: .echoSecondaryText)

            Spacer()
        }
        .padding(.horizontal, 24)
        .presentationDetents([.fraction(0.52)])
        .presentationDragIndicator(.hidden)
    }

    private func performStepUp() async {
        isAuthenticating = true
        errorMessage = nil
        defer { isAuthenticating = false }

        do {
            switch selectedMethod {
            case .devicePasscode:
                // Local passcode verification — no elevated token from backend
                try await StepUpAuthManager.shared.authenticate(
                    reason: action.displayTitle,
                    override: .devicePasscode
                )
                // For device passcode, pass an empty token (caller should check for empty)
                onVerified("")
                dismiss()

            case .faceID, .passkey:
                // Full WebAuthn flow → elevated backend token
                let challenge = try await apiClient.getLoginChallenge()
                let assertion = try await passkeyManager.authenticateWithPasskey(
                    challenge: challenge.challengeData
                )
                let token = try await tokenManager.getValidAccessToken()
                let response = try await apiClient.requestStepUp(
                    action: action.rawValue,
                    assertion: assertion,
                    token: token
                )
                onVerified(response.elevatedToken)
                dismiss()
            }
        } catch is CancellationError {
            // User cancelled — not an error
        } catch {
            errorMessage = "Verification failed. Please try again."
        }
    }
}
#endif
