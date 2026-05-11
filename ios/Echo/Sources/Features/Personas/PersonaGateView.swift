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
    // Design review: dark surface for private moments so "privacy feels different, not just looks it".
    // Minimal copy — no logo, no hint at what's hidden, no decorative chrome.

    private var gateScreen: some View {
        ZStack {
            Color.echoNight.ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer()

                // Double-ring circle around Face ID glyph
                ZStack {
                    Circle()
                        .stroke(Color.echoNightHair, lineWidth: 1)
                        .frame(width: 136, height: 136)

                    Circle()
                        .stroke(Color.echoNightHair.opacity(0.5), lineWidth: 1)
                        .frame(width: 120, height: 120)
                        .padding(8)

                    Circle()
                        .fill(Color.echoNightHi)
                        .frame(width: 120, height: 120)

                    Image(systemName: "faceid")
                        .font(.system(size: 48, weight: .ultraLight))
                        .foregroundStyle(Color.echoNightInk)
                        .scaleEffect(isUnlocking ? 1.06 : 1.0)
                        .animation(.easeInOut(duration: 0.9).repeatForever(autoreverses: true),
                                   value: isUnlocking)
                }
                .padding(.bottom, 28)

                Text("Verify to continue.")
                    .font(.system(size: 22, weight: .semibold))
                    .tracking(-0.5)
                    .foregroundStyle(Color.echoNightInk)

                Text("This area requires biometric confirmation.\nIt will lock again after two minutes in the background.")
                    .font(.system(size: 13.5))
                    .lineSpacing(4)
                    .foregroundStyle(Color.echoNightInk70)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 40)
                    .padding(.top, 10)

                // Mono status tag
                VStack(spacing: 0) {
                    if let error = errorMessage {
                        Text(error)
                            .font(.echomono(11))
                            .foregroundStyle(Color.echoAlert)
                    } else {
                        Text(isUnlocking ? "● scanning" : "● awaiting Face ID")
                            .font(.echomono(11))
                            .foregroundStyle(Color.echoNightInk40)
                    }
                }
                .padding(.top, 36)

                Spacer()

                // Cancel + Try again — quiet, no filled button
                HStack(spacing: 16) {
                    Button("Cancel") { isUnlocked = false }
                        .font(.system(size: 13))
                        .foregroundStyle(Color.echoNightInk70)
                        .padding(8)

                    Button("Try again") { Task { await unlock() } }
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(Color.echoNightInk)
                        .padding(8)
                        .disabled(isUnlocking)
                }
                .padding(.bottom, 48)
            }
        }
        .preferredColorScheme(.dark)
        .onAppear {
            Task {
                try? await Task.sleep(nanoseconds: 300_000_000)
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
