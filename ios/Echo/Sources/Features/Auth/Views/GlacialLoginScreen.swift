// Features/Auth/Views/GlacialLoginScreen.swift
// Wave 12: Signal-style biometric login.
// Primary path: saved username + Face ID auto-trigger on appear.
// Secondary: passkey (if enrolled in Settings).
// Recovery: "Recover account" link → RecoveryCoordinator.

import SwiftUI
import LocalAuthentication

struct GlacialLoginScreen: View {
    @State private var showSMSSection = false
    @State private var phoneNumber = ""
    @State private var savedUsername: String = ""
    @State private var isUnlocking = false
    @State private var unlockError: String?
    let onPasskeyLogin: () -> Void
    let onSMSLogin: (String) -> Void
    let onGetStarted: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()

            ScrollView {
                VStack(spacing: 32) {
                    Spacer(minLength: 60)

                    // Frosted Glass Header
                    VStack(spacing: 8) {
                        EchoLogo(size: 64)
                        Text("ECHO")
                            .font(Font.Echo.displayMedium)
                            .foregroundStyle(Color.Echo.primaryContainer)

                        // Saved username greeting
                        if !savedUsername.isEmpty {
                            Text("Welcome back, \(savedUsername)")
                                .font(.system(size: 14))
                                .foregroundStyle(Color.Echo.onSurfaceVariant)
                                .padding(.top, 4)
                        }
                    }
                    .padding(.vertical, 24)
                    .frame(maxWidth: .infinity)
                    .background(.ultraThinMaterial.opacity(0.6))
                    .clipShape(RoundedRectangle(cornerRadius: 32))
                    .ghostBorder(opacity: 0.15)
                    .glacialShadow(radius: 32, opacity: 0.04)
                    .padding(.horizontal)

                    // Primary biometric login button
                    SignatureGradientButton(
                        title: isUnlocking ? "Verifying…" : "Unlock with Face ID",
                        subtitle: savedUsername.isEmpty ? "Your face is your key" : "Verify to continue",
                        icon: "faceid"
                    ) {
                        Task { await triggerBiometricLogin() }
                    }
                    .disabled(isUnlocking)
                    .padding(.horizontal)

                    if let error = unlockError {
                        Text(error)
                            .font(.system(size: 13))
                            .foregroundStyle(Color.echoError)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 32)
                    }

                    // Secondary options
                    VStack(spacing: 12) {
                        Button("Use passkey instead") { onPasskeyLogin() }
                            .font(.system(size: 13))
                            .foregroundStyle(Color.Echo.primaryContainer)

                        Button("Recover account") { onSMSLogin("") }
                            .font(.system(size: 13))
                            .foregroundStyle(Color.Echo.onSurfaceVariant)
                    }

                    // SMS section toggle (legacy / dev only)
                    Button {
                        withAnimation(.glacial) {
                            showSMSSection.toggle()
                        }
                    } label: {
                        HStack(spacing: 12) {
                            Rectangle()
                                .fill(Color.Echo.outlineVariant.opacity(0.2))
                                .frame(height: 1)

                            HStack(spacing: 4) {
                                Text("SECURE ALTERNATIVE")
                                    .font(Font.Echo.labelSm)
                                    .tracking(1.5)
                                    .foregroundStyle(Color.Echo.outline.opacity(0.6))
                                Image(systemName: showSMSSection ? "chevron.up" : "chevron.down")
                                    .font(.system(size: 8, weight: .bold))
                                    .foregroundStyle(Color.Echo.outline.opacity(0.6))
                            }

                            Rectangle()
                                .fill(Color.Echo.outlineVariant.opacity(0.2))
                                .frame(height: 1)
                        }
                        .padding(.horizontal)
                    }

                    // SMS Section (Expandable)
                    if showSMSSection {
                        VStack(spacing: 16) {
                            TextField("Phone number", text: $phoneNumber)
                                .font(Font.Echo.bodyLarge)
                                .foregroundStyle(Color.Echo.onSurface)
                                .padding(16)
                                .background(Color.Echo.surfaceContainerLowest)
                                .clipShape(RoundedRectangle(cornerRadius: 32))
                                .ghostBorder(opacity: 0.10)
                                #if os(iOS)
                                .keyboardType(.phonePad)
                                #endif

                            Button {
                                onSMSLogin(phoneNumber)
                            } label: {
                                Text("Send Code")
                                    .font(Font.Echo.bodyLarge)
                                    .foregroundStyle(Color.Echo.onSurface)
                                    .frame(maxWidth: .infinity)
                                    .padding(.vertical, 14)
                                    .background(Color.Echo.surfaceContainerHighest)
                                    .clipShape(RoundedRectangle(cornerRadius: 32))
                                    .ghostBorder(opacity: 0.20)
                            }
                            .disabled(phoneNumber.count < 10)
                            .opacity(phoneNumber.count < 10 ? 0.5 : 1.0)
                        }
                        .padding(20)
                        .background(Color.Echo.surfaceContainerLow)
                        .clipShape(RoundedRectangle(cornerRadius: 32))
                        .ghostBorder(opacity: 0.15)
                        .padding(.horizontal)
                        .transition(.opacity.combined(with: .move(edge: .top)))
                    }

                    Spacer(minLength: 40)

                    HStack(spacing: 4) {
                        Text("New to ECHO?")
                            .font(.system(size: 12))
                            .foregroundStyle(Color.Echo.onSurfaceVariant)
                        Button("Create Account") {
                            onGetStarted()
                        }
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundStyle(Color.Echo.primaryContainer)
                    }
                    .padding(.bottom, 32)
                }
            }
        }
        .icyBackground()
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
