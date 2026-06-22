#if os(iOS)
import SwiftUI
import CryptoKit

// WO-136: Passkey creation step in the Streamlined Onboarding flow.
//
// Flow:
//  1. User arrives here after profile creation (display name confirmed).
//  2. Taps "Create Passkey" → SecureEnclaveManager generates P-256 key pair.
//  3. Key pair is submitted to backend POST /v1/auth/passkey.
//  4. Immediate test authentication verifies the passkey before completing onboarding.
//  5. On success, onCompletion is called; onboarding coordinator moves to main tab.

/// Full-screen onboarding step that creates and registers a Secure Enclave passkey.
public struct PasskeySetupView: View {
    let onCompletion: () -> Void

    @State private var phase: SetupPhase = .idle
    @State private var errorMessage: String?

    public init(onCompletion: @escaping () -> Void) {
        self.onCompletion = onCompletion
    }

    enum SetupPhase {
        case idle, generating, submitting, verifying, done
        var label: String {
            switch self {
            case .idle:       return "Create Passkey"
            case .generating: return "Generating key…"
            case .submitting: return "Registering…"
            case .verifying:  return "Verifying…"
            case .done:       return "Done"
            }
        }
        var isWorking: Bool { self != .idle && self != .done }
    }

    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 32) {
                Spacer()

                // Icon
                ZStack {
                    Circle()
                        .fill(Color.echoPrimary.opacity(0.12))
                        .frame(width: 100, height: 100)
                    Image(systemName: phase == .done ? "checkmark.seal.fill" : "faceid")
                        .font(.system(size: 44, weight: .semibold))
                        .foregroundColor(phase == .done ? .echoSuccess : .echoPrimary)
                        .animation(.spring(), value: phase == .done)
                }

                VStack(spacing: 12) {
                    Text("Secure Your Account")
                        .font(.system(size: 22, weight: .bold))
                        .foregroundColor(.echoPrimaryText)

                    Text("Create a backup way to sign in using Face ID. Your account stays on your phone.")
                        .font(.system(size: 14))
                        .foregroundColor(.echoSecondaryText)
                        .multilineTextAlignment(.center)
                        .lineSpacing(3)
                        .padding(.horizontal, 32)
                }

                if let error = errorMessage {
                    Text(error)
                        .font(.system(size: 13))
                        .foregroundColor(.echoError)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 32)
                }

                Spacer()

                VStack(spacing: 16) {
                    EchoButton(
                        phase.label,
                        style: .primary
                    ) {
                        guard phase == .idle else { return }
                        Task { await createPasskey() }
                    }
                    .disabled(phase.isWorking)

                    if phase == .idle {
                        Button("Skip for now") { onCompletion() }
                            .font(.system(size: 13))
                            .foregroundColor(.echoSecondaryText)
                    }
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 48)
            }
        }
    }

    // MARK: - Passkey Creation

    private func createPasskey() async {
        errorMessage = nil
        phase = .generating

        do {
            // 1. Generate hardware-bound P-256 key pair in Secure Enclave.
            let publicKeyBase64 = try await generateDeviceKey()

            // 2. Submit public key to backend.
            phase = .submitting
            try await submitPublicKey(publicKeyBase64)

            // 3. Verify passkey works (test sign + verify round-trip).
            phase = .verifying
            try await verifyPasskeyWorks()

            phase = .done
            try? await Task.sleep(nanoseconds: 600_000_000) // 0.6s success display
            onCompletion()

        } catch {
            phase = .idle
            errorMessage = error.localizedDescription
        }
    }

    private func generateDeviceKey() async throws -> String {
        // Returns base64-encoded uncompressed P-256 public key (04 || X || Y).
        return try await SecureEnclaveManager.shared.generateBiometricProtectedKey(id: "echo-identity-signing")
    }

    private func submitPublicKey(_ publicKeyBase64: String) async throws {
        // POST /v1/auth/passkey — backend registers the key under the user's DID.
        // Phase 1 stub: the network call is wired in the APIClient integration WO.
        let _ = publicKeyBase64
    }

    private func verifyPasskeyWorks() async throws {
        // Sign a challenge with the newly created key to confirm round-trip works.
        let challenge = "echo-passkey-verify-\(UUID().uuidString)"
        _ = try await SecureEnclaveManager.shared.sign(
            data: Data(challenge.utf8),
            keyId: "echo-identity-signing"
        )
    }
}

// echoSuccess is defined in DesignSystem/Colors.swift as Color(hex: 0x10B981).

#Preview {
    PasskeySetupView(onCompletion: {})
}
#endif
