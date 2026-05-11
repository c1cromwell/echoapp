#if os(iOS)
import SwiftUI

// Wave 12: Biometric re-auth gate for hidden personas and folders.
//
// Usage:
//   PersonaGateView(personaID: persona.id) {
//       HiddenConversationListView(persona: persona)
//   }
//
// Security properties:
//   - Triggers SecureEnclaveManager.sign() with a per-persona nonce — Face ID prompts
//   - The resulting signature is discarded immediately (proof-of-presence, not a secret)
//   - Auto-locks after 2 minutes of the app being in the background
//   - Re-auth required each time the user returns from background while gated

struct PersonaGateView<Content: View>: View {
    let personaID: String
    @ViewBuilder let protectedContent: () -> Content

    @State private var isUnlocked = false
    @State private var isUnlocking = false
    @State private var errorMessage: String?
    @State private var unlockTime: Date?
    @Environment(\.scenePhase) private var scenePhase

    private let lockAfterBackground: TimeInterval = 120 // 2 minutes

    var body: some View {
        Group {
            if isUnlocked {
                protectedContent()
                    .onChange(of: scenePhase) { _, newPhase in
                        if newPhase == .background {
                            // Start the lock countdown.
                        } else if newPhase == .active {
                            // If the app was in the background for too long, re-lock.
                            if let t = unlockTime, Date().timeIntervalSince(t) > lockAfterBackground {
                                isUnlocked = false
                                unlockTime = nil
                            }
                        }
                    }
            } else {
                gateScreen
            }
        }
        .onChange(of: scenePhase) { _, newPhase in
            if newPhase == .background {
                unlockTime = isUnlocked ? Date() : nil
            }
        }
    }

    // MARK: - Gate screen

    private var gateScreen: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            VStack(spacing: 32) {
                Spacer()

                ZStack {
                    Circle()
                        .fill(Color.echoPrimary.opacity(0.10))
                        .frame(width: 100, height: 100)
                    Image(systemName: "lock.shield.fill")
                        .font(.system(size: 44, weight: .semibold))
                        .foregroundColor(.echoPrimary)
                }

                VStack(spacing: 12) {
                    Text("Hidden area")
                        .font(.system(size: 22, weight: .bold))
                        .foregroundColor(.echoPrimaryText)
                    Text("Verify with Face ID to access this protected space.")
                        .font(.system(size: 14))
                        .foregroundColor(.echoSecondaryText)
                        .multilineTextAlignment(.center)
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

                VStack(spacing: 12) {
                    EchoButton(
                        isUnlocking ? "Verifying…" : "Unlock with Face ID",
                        style: .primary
                    ) {
                        guard !isUnlocking else { return }
                        Task { await unlock() }
                    }
                    .disabled(isUnlocking)
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 48)
            }
        }
        .onAppear {
            // Auto-trigger on appear for a Signal-like seamless experience.
            Task {
                try? await Task.sleep(nanoseconds: 300_000_000) // 300ms
                await unlock()
            }
        }
    }

    // MARK: - Unlock

    private func unlock() async {
        isUnlocking = true
        errorMessage = nil

        do {
            // Proof-of-presence: sign a per-persona nonce with the identity key.
            // Face ID is triggered here. The signature is discarded — we just need
            // to know the biometric passed.
            let nonce = Data("persona-access-\(personaID)-\(Date().timeIntervalSince1970)".utf8)
            _ = try await SecureEnclaveManager.shared.sign(
                data: nonce,
                keyId: "echo-identity-signing"
            )
            isUnlocked = true
            unlockTime = Date()
        } catch {
            errorMessage = "Verification failed. Please try again."
        }

        isUnlocking = false
    }
}

// MARK: - Persona model extension

extension Persona {
    /// Whether this persona requires a biometric gate to access.
    var requiresGate: Bool { visibility == .hidden }
}
#endif
