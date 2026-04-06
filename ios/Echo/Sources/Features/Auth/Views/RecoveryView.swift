import SwiftUI

struct RecoveryView: View {
    let method: RecoveryMethod
    let onComplete: () -> Void
    let onCancel: () -> Void

    @State private var step: RecoveryStep = .intro
    @State private var isLoading = false
    @State private var errorMessage: String?

    var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Account Recovery",
                    showBackButton: true,
                    onBackPressed: onCancel
                )

                VStack(spacing: Spacing.xl.rawValue) {
                    // Progress indicator
                    HStack(spacing: 8) {
                        ForEach(0..<stepsForMethod.count, id: \.self) { index in
                            Circle()
                                .fill(index <= currentStepIndex
                                      ? Color.echoPrimary
                                      : Color.echoGray300)
                                .frame(width: 8, height: 8)
                        }
                    }
                    .padding(.top, Spacing.md.rawValue)

                    switch step {
                    case .intro:
                        recoveryIntroView

                    case .verification:
                        verificationView

                    case .newPasskey:
                        newPasskeyView

                    case .complete:
                        completionView
                    }

                    Spacer()
                }
                .padding(.horizontal, Spacing.lg.rawValue)
            }
        }
    }

    // MARK: - Step Views

    private var recoveryIntroView: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            Image(systemName: iconForMethod)
                .font(.system(size: 48))
                .foregroundColor(.echoPrimary)

            Text(titleForMethod)
                .typographyStyle(.h3, color: .echoPrimaryText)
                .multilineTextAlignment(.center)

            Text(descriptionForMethod)
                .typographyStyle(.body, color: .echoSecondaryText)
                .multilineTextAlignment(.center)

            EchoButton(
                "Begin Recovery",
                style: .primary,
                size: .large,
                action: { step = .verification }
            )
        }
    }

    private var verificationView: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            Text("Step \(currentStepIndex + 1): Verify Identity")
                .typographyStyle(.h4, color: .echoPrimaryText)

            switch method {
            case .recoveryPhrase:
                Text("Enter your 12-word recovery phrase to verify your identity.")
                    .typographyStyle(.body, color: .echoSecondaryText)
                    .multilineTextAlignment(.center)

            case .trustedContacts:
                Text("2 of your 3 trusted contacts must confirm your identity. We've sent them a notification.")
                    .typographyStyle(.body, color: .echoSecondaryText)
                    .multilineTextAlignment(.center)

            case .phoneReverification:
                Text("We'll send a verification code to your registered phone number.")
                    .typographyStyle(.body, color: .echoSecondaryText)
                    .multilineTextAlignment(.center)
            }

            EchoButton(
                "Continue",
                style: .primary,
                size: .large,
                isLoading: isLoading,
                action: { step = .newPasskey }
            )

            if let error = errorMessage {
                Text(error)
                    .typographyStyle(.caption, color: .red)
            }
        }
    }

    private var newPasskeyView: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            Image(systemName: "key.fill")
                .font(.system(size: 48))
                .foregroundColor(.echoPrimary)

            Text("Create New Passkey")
                .typographyStyle(.h3, color: .echoPrimaryText)

            Text("Set up a new passkey for this device to complete your account recovery.")
                .typographyStyle(.body, color: .echoSecondaryText)
                .multilineTextAlignment(.center)

            EchoButton(
                "Create Passkey",
                style: .primary,
                size: .large,
                icon: Image(systemName: "faceid"),
                action: { step = .complete }
            )
        }
    }

    private var completionView: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            Image(systemName: "checkmark.circle.fill")
                .font(.system(size: 64))
                .foregroundColor(.echoSuccess)

            Text("Recovery Complete")
                .typographyStyle(.h3, color: .echoPrimaryText)

            Text("Your account has been recovered and secured with a new passkey.")
                .typographyStyle(.body, color: .echoSecondaryText)
                .multilineTextAlignment(.center)

            EchoButton(
                "Continue to Echo",
                style: .primary,
                size: .large,
                action: onComplete
            )
        }
    }

    // MARK: - Helpers

    private enum RecoveryStep {
        case intro, verification, newPasskey, complete
    }

    private var stepsForMethod: [String] {
        switch method {
        case .recoveryPhrase: return ["Intro", "Phrase", "Passkey", "Done"]
        case .trustedContacts: return ["Intro", "Contacts", "Passkey", "Done"]
        case .phoneReverification: return ["Intro", "OTP", "Passkey", "Done"]
        }
    }

    private var currentStepIndex: Int {
        switch step {
        case .intro: return 0
        case .verification: return 1
        case .newPasskey: return 2
        case .complete: return 3
        }
    }

    private var iconForMethod: String {
        switch method {
        case .recoveryPhrase: return "doc.text.fill"
        case .trustedContacts: return "person.3.fill"
        case .phoneReverification: return "phone.fill"
        }
    }

    private var titleForMethod: String {
        switch method {
        case .recoveryPhrase: return "Recovery Phrase"
        case .trustedContacts: return "Trusted Contacts"
        case .phoneReverification: return "Phone Verification"
        }
    }

    private var descriptionForMethod: String {
        switch method {
        case .recoveryPhrase:
            return "Use your 12-word recovery phrase to regain access to your account."
        case .trustedContacts:
            return "Your trusted contacts can help verify your identity and restore access."
        case .phoneReverification:
            return "Verify your phone number to recover your account."
        }
    }
}
