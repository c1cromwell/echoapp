#if os(iOS)
// Features/Auth/Views/GlacialLoginScreen.swift
// Design review: "No 'Welcome back' + logo + frosted card stack.
// One greeting, one trigger." — quiet return, warm paper, breathing Face ID target.

import SwiftUI
import LocalAuthentication

public struct GlacialLoginScreen: View {
    @State private var savedUsername: String = ""
    @State private var isUnlocking = false
    @State private var unlockError: String?
    let onPasskeyLogin: () -> Void
    let onSMSLogin: (String) -> Void
    let onGetStarted: () -> Void

    public init(
        onPasskeyLogin: @escaping () -> Void = {},
        onSMSLogin: @escaping (String) -> Void = { _ in },
        onGetStarted: @escaping () -> Void = {}
    ) {
        self.onPasskeyLogin = onPasskeyLogin
        self.onSMSLogin = onSMSLogin
        self.onGetStarted = onGetStarted
    }

    public var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(spacing: 0) {
                // Minimal mark — top-left, no hero logo lockup
                HStack {
                    EchoRippleMark(size: 22, color: .echoSignal)
                    Spacer()
                }
                .padding(.horizontal, 28)
                .padding(.top, 8)

                Spacer()

                // Username + heading
                VStack(alignment: .leading, spacing: 0) {
                    Text(savedUsername.isEmpty ? "" : "@\(savedUsername)")
                        .font(.echomono(11))
                        .foregroundStyle(Color.echoInk40)
                        .padding(.bottom, 8)

                    Text("Look at\nyour phone.")
                        .font(.system(size: 32, weight: .semibold))
                        .tracking(-0.8)
                        .lineSpacing(2)
                        .foregroundStyle(Color.echoInk)

                    Text("Face ID will unlock in a moment.")
                        .font(.system(size: 14))
                        .foregroundStyle(Color.echoInk55)
                        .padding(.top, 10)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal, 28)

                // Breathing Face ID target — the main interactive element
                VStack(spacing: 18) {
                    Button(action: { Task { await triggerBiometricLogin() } }) {
                        ZStack {
                            Circle()
                                .fill(Color.echoPaperDim)
                                .frame(width: 112, height: 112)
                                .overlay(Circle().stroke(Color.echoHair, lineWidth: 1))

                            Image(systemName: "faceid")
                                .font(.system(size: 48, weight: .ultraLight))
                                .foregroundStyle(isUnlocking ? Color.echoSignal : Color.echoInk)
                                .scaleEffect(isUnlocking ? 1.05 : 1.0)
                                .animation(.easeInOut(duration: 0.8).repeatForever(autoreverses: true),
                                           value: isUnlocking)
                        }
                    }
                    .disabled(isUnlocking)

                    if let error = unlockError {
                        Text(error)
                            .font(.system(size: 13))
                            .foregroundStyle(Color.echoAlert)
                            .multilineTextAlignment(.center)
                    } else {
                        Text(isUnlocking ? "● SCANNING" : "")
                            .font(.echomono(10))
                            .tracking(1)
                            .foregroundStyle(Color.echoTrustGreen)
                    }
                }
                .padding(.top, 44)

                Spacer()

                // Secondary links — below the target, not competing
                VStack(spacing: 4) {
                    Button("Use passkey", action: onPasskeyLogin)
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(Color.echoInk)
                        .padding(8)

                    Button("Recover account") { onSMSLogin("") }
                        .font(.system(size: 12))
                        .foregroundStyle(Color.echoInk55)
                        .padding(4)

                    Button("New to Echo? Create account", action: onGetStarted)
                        .font(.system(size: 12))
                        .foregroundStyle(Color.echoSignal)
                        .padding(.top, 8)
                }
                .padding(.bottom, 48)
            }
        }
        .preferredColorScheme(.light)
        .onAppear { loadSavedUsername(); Task { await autoTriggerBiometric() } }
    }

    // MARK: - Biometric login

    private func loadSavedUsername() {
        Task {
            savedUsername = (try? await KeychainManager.shared.retrieve(
                key: "echo.username.current", as: String.self
            )) ?? ""
        }
    }

    private func autoTriggerBiometric() async {
        try? await Task.sleep(nanoseconds: 400_000_000)
        guard !savedUsername.isEmpty else { return }
        await triggerBiometricLogin()
    }

    private func triggerBiometricLogin() async {
        guard !isUnlocking else { return }
        isUnlocking = true
        unlockError = nil

        do {
            // Proof-of-presence: sign a login challenge with the Secure Enclave key.
            // Face ID is triggered here. The signature is discarded — we only need
            // to know biometric passed.
            let challenge = Data("echo-login-\(UUID().uuidString)".utf8)
            _ = try await SecureEnclaveManager.shared.sign(
                data: challenge,
                keyId: "echo-identity-signing"
            )
            onPasskeyLogin()
        } catch {
            isUnlocking = false
            let nsErr = error as NSError
            if nsErr.domain == "com.apple.LocalAuthentication" {
                unlockError = "Face ID failed. Tap to try again."
            } else {
                unlockError = "Authentication failed. Please try again."
            }
        }
    }
}

#Preview {
    GlacialLoginScreen(
        onPasskeyLogin: {},
        onSMSLogin: { _ in },
        onGetStarted: {}
    )
}
#endif
