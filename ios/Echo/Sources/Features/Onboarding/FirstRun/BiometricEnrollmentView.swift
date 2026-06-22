#if os(iOS)
import SwiftUI
import LocalAuthentication

// Wave 12 — Onboarding Step 3: Biometric enrollment.
//
// "Your face is your key." This screen generates the Secure Enclave P-256
// identity key (biometryCurrentSet) and registers the resulting did:key with
// the backend.  Face ID is triggered the first time the key is USED (via sign),
// not during creation.
//
// Security properties after this step:
//   - Private key is hardware-bound, never exported from the Secure Enclave
//   - Key is invalidated if new biometric enrollments are added (biometryCurrentSet)
//   - Storage key derived via HKDF from identity key signature (WO-224)

public struct BiometricEnrollmentView: View {
    let username: String
    let onComplete: (String) -> Void
    let onUnsupported: () -> Void

    @State private var phase: EnrollPhase = .idle
    @State private var errorMessage: String?

    public init(
        username: String,
        onComplete: @escaping (String) -> Void = { _ in },
        onUnsupported: @escaping () -> Void = {}
    ) {
        self.username = username
        self.onComplete = onComplete
        self.onUnsupported = onUnsupported
    }

    enum EnrollPhase {
        case idle, generating, registering, verifying, done
        var label: String {
            switch self {
            case .idle:        return "Set Up Face ID"
            case .generating:  return "Creating your key…"
            case .registering: return "Registering identity…"
            case .verifying:   return "Verifying…"
            case .done:        return "All set!"
            }
        }
        var isWorking: Bool { self != .idle && self != .done }
    }

    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            VStack(spacing: 32) {
                Spacer()

                // Hero icon
                ZStack {
                    Circle()
                        .fill(Color.echoPrimary.opacity(0.12))
                        .frame(width: 100, height: 100)
                    Image(systemName: phase == .done ? "checkmark.seal.fill" : "faceid")
                        .font(.system(size: 44, weight: .semibold))
                        .foregroundColor(phase == .done ? Color(hex: 0x10B981) : .echoPrimary)
                        .animation(.spring(), value: phase == .done)
                }

                VStack(spacing: 12) {
                    Text("Your face is your key")
                        .font(.system(size: 22, weight: .bold))
                        .foregroundColor(.echoPrimaryText)
                    Text("Echo uses Face ID to protect your identity. Your private key never leaves this device.")
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

                VStack(spacing: 8) {
                    EchoButton(phase.label, style: .primary) {
                        guard phase == .idle else { return }
                        Task { await enroll() }
                    }
                    .disabled(phase.isWorking)

                    Text("Your face unlocks ECHO and keeps your account safe.")
                        .font(.system(size: 11))
                        .foregroundColor(.echoSecondaryText)
                        .multilineTextAlignment(.center)
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 48)
            }
        }
        .onAppear { checkBiometricAvailability() }
    }

    // MARK: - Enrollment flow

    private func checkBiometricAvailability() {
        #if targetEnvironment(simulator)
        // Simulator may not have biometrics enrolled; allow proceeding for dev testing.
        return
        #else
        let ctx = LAContext()
        var error: NSError?
        guard ctx.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
            onUnsupported()
            return
        }
        #endif
    }

    private func enroll() async {
        errorMessage = nil
        phase = .generating

        do {
            // 1. Generate biometric-protected P-256 key in Secure Enclave.
            let publicKeyBase64 = try await SecureEnclaveManager.shared
                .generateBiometricProtectedKey(id: "echo-identity-signing")

            // 2. Trigger Face ID by signing a challenge — this proves the key works.
            phase = .verifying
            let challenge = Data("echo-identity-verify-\(UUID().uuidString)".utf8)
            _ = try await SecureEnclaveManager.shared.sign(
                data: challenge,
                keyId: "echo-identity-signing"
            )

            // 3. Register DID with backend.
            phase = .registering
            let did = try await registerDID(publicKeyBase64: publicKeyBase64)

            // 4. Persist DID + username to Keychain.
            try await KeychainManager.shared.store(key: "echo.did.current", value: did)
            try await KeychainManager.shared.store(key: "echo.username.current", value: username)

            // 5. Biometric integrity baseline is saved by SecureEnclaveManager
            //    on first successful sign() — no extra call needed here.

            phase = .done
            try? await Task.sleep(nanoseconds: 600_000_000)
            onComplete(did)

        } catch {
            phase = .idle
            errorMessage = error.localizedDescription
        }
    }

    /// Derives canonical `did:key` and registers with `POST /identity/register`.
    private func registerDID(publicKeyBase64: String) async throws -> String {
        guard let pubData = Data(base64Encoded: publicKeyBase64) else {
            throw EnrollError.invalidPublicKey
        }
        let pubHex = pubData.map { String(format: "%02x", $0) }.joined()
        let did = try DidKeyDeriver.deriveFromPublicKeyHex(pubHex)

        do {
            let response = try await IdentityService().registerDidKeyOnEchoAPI(
                baseURL: EchoAPIBaseURL.resolved,
                did: did,
                publicKeyHex: pubHex
            )
            return response.did
        } catch {
            // Offline / unreachable backend: keep canonical DID for local Keychain use.
            return did
        }
    }
}

enum EnrollError: LocalizedError {
    case invalidPublicKey
    var errorDescription: String? { "Invalid public key format" }
}
#endif
