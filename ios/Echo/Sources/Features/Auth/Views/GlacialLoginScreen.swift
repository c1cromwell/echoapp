#if os(iOS)
// Features/Auth/Views/GlacialLoginScreen.swift
// Login screen — username + Face ID auto-login. No password.
//
// States:
//   .normal           — saved username found, Face ID auto-triggers 400ms post-appear
//   .softLocked       — 5 Face ID failures → "Use device passcode" button shown
//   .hardLocked(Date) — 10 failures → 15-min countdown + "Recover account" link
//   .biometricsUnavailable — Face ID removed/unenrolled → passcode + recover shown
//   .noAccount        — no saved username → "Create account" CTA

import SwiftUI
import LocalAuthentication

public struct GlacialLoginScreen: View {
    @State private var savedUsername: String = ""
    @State private var loginState: LoginState = .normal
    @State private var isUnlocking = false
    @State private var unlockError: String?
    @State private var showLinkDevice = false

    let onSuccess: () -> Void
    let onRecovery: () -> Void
    let onGetStarted: () -> Void

    // Legacy callbacks kept for existing call sites
    let onPasskeyLogin: () -> Void
    let onSMSLogin: (String) -> Void

    public init(
        onPasskeyLogin: @escaping () -> Void = {},
        onSMSLogin: @escaping (String) -> Void = { _ in },
        onGetStarted: @escaping () -> Void = {},
        onSuccess: @escaping () -> Void = {},
        onRecovery: @escaping () -> Void = {}
    ) {
        self.onPasskeyLogin = onPasskeyLogin
        self.onSMSLogin = onSMSLogin
        self.onGetStarted = onGetStarted
        self.onSuccess = onSuccess
        self.onRecovery = onRecovery
    }

    enum LoginState {
        case normal
        case softLocked
        case hardLocked(until: Date)
        case biometricsUnavailable
        case noAccount
    }

    public var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(spacing: 0) {
                HStack {
                    EchoRippleMark(size: 22, color: .echoSignal)
                    Spacer()
                }
                .padding(.horizontal, 28)
                .padding(.top, 8)

                Spacer()

                switch loginState {
                case .noAccount:
                    noAccountView

                case .biometricsUnavailable:
                    biometricsUnavailableView

                case .softLocked:
                    softLockedView

                case .hardLocked(let until):
                    hardLockedView(until: until)

                case .normal:
                    normalLoginView
                }

                Spacer()

                footerLinks
                    .padding(.bottom, 48)
            }
        }
        .preferredColorScheme(.light)
        .onAppear {
            Task { await setupLoginState() }
        }
        .sheet(isPresented: $showLinkDevice) {
            LinkDeviceScanView()
        }
    }

    // MARK: - Login views

    private var normalLoginView: some View {
        VStack(spacing: 0) {
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

                Group {
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
            }
            .padding(.top, 44)
        }
    }

    private var noAccountView: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Welcome to Echo.")
                .font(.system(size: 32, weight: .semibold))
                .tracking(-0.8)
                .foregroundStyle(Color.echoInk)
            Text("No password. No email. Your face is your key.")
                .font(.system(size: 14))
                .lineSpacing(3)
                .foregroundStyle(Color.echoInk55)

            Button(action: onGetStarted) {
                HStack {
                    Text("Create account")
                        .font(.system(size: 15, weight: .semibold))
                    Spacer()
                    Text("→").font(.system(size: 18))
                }
                .foregroundStyle(Color.echoPaper)
                .padding(.horizontal, 18)
                .padding(.vertical, 15)
                .background(Color.echoInk, in: RoundedRectangle(cornerRadius: 14))
            }
            .buttonStyle(SpringPressStyle())
            .padding(.top, 8)
        }
        .padding(.horizontal, 28)
    }

    private var biometricsUnavailableView: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text(savedUsername.isEmpty ? "" : "@\(savedUsername)")
                .font(.echomono(11))
                .foregroundStyle(Color.echoInk40)
                .padding(.bottom, 8)

            Text("Face ID\nunavailable.")
                .font(.system(size: 32, weight: .semibold))
                .tracking(-0.8)
                .lineSpacing(2)
                .foregroundStyle(Color.echoInk)

            Text("Face ID has been removed or is not enrolled on this device. Use your device passcode or recover your account.")
                .font(.system(size: 13))
                .lineSpacing(3)
                .foregroundStyle(Color.echoInk55)
                .padding(.top, 10)

            VStack(spacing: 10) {
                Button(action: { Task { await useDevicePasscode() } }) {
                    HStack {
                        Image(systemName: "lock.fill").font(.system(size: 14))
                        Text("Use device passcode")
                            .font(.system(size: 15, weight: .semibold))
                        Spacer()
                        Text("→").font(.system(size: 16))
                    }
                    .foregroundStyle(Color.echoPaper)
                    .padding(.horizontal, 18)
                    .padding(.vertical, 14)
                    .background(Color.echoInk, in: RoundedRectangle(cornerRadius: 14))
                }
                .buttonStyle(SpringPressStyle())

                Button("Recover account via SMS") { triggerRecovery() }
                    .font(.system(size: 13, weight: .medium))
                    .foregroundStyle(Color.echoInk55)
                    .padding(.vertical, 6)
            }
            .padding(.top, 24)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, 28)
    }

    private var softLockedView: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text(savedUsername.isEmpty ? "" : "@\(savedUsername)")
                .font(.echomono(11))
                .foregroundStyle(Color.echoInk40)
                .padding(.bottom, 8)

            Text("Too many\nattempts.")
                .font(.system(size: 32, weight: .semibold))
                .tracking(-0.8)
                .lineSpacing(2)
                .foregroundStyle(Color.echoInk)

            Text("Face ID is temporarily unavailable after multiple failed attempts. Use your device passcode to continue.")
                .font(.system(size: 13))
                .lineSpacing(3)
                .foregroundStyle(Color.echoInk55)
                .padding(.top, 10)

            Button(action: { Task { await useDevicePasscode() } }) {
                HStack {
                    Image(systemName: "lock.fill").font(.system(size: 14))
                    Text("Use device passcode")
                        .font(.system(size: 15, weight: .semibold))
                    Spacer()
                    Text("→").font(.system(size: 16))
                }
                .foregroundStyle(Color.echoPaper)
                .padding(.horizontal, 18)
                .padding(.vertical, 14)
                .background(Color.echoInk, in: RoundedRectangle(cornerRadius: 14))
            }
            .buttonStyle(SpringPressStyle())
            .padding(.top, 24)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, 28)
    }

    private func hardLockedView(until: Date) -> some View {
        VStack(alignment: .leading, spacing: 0) {
            Text(savedUsername.isEmpty ? "" : "@\(savedUsername)")
                .font(.echomono(11))
                .foregroundStyle(Color.echoInk40)
                .padding(.bottom, 8)

            Text("Account\nlocked.")
                .font(.system(size: 32, weight: .semibold))
                .tracking(-0.8)
                .lineSpacing(2)
                .foregroundStyle(Color.echoInk)

            Text("Too many failed attempts. Face ID is locked for 15 minutes.")
                .font(.system(size: 13))
                .lineSpacing(3)
                .foregroundStyle(Color.echoInk55)
                .padding(.top, 10)

            CountdownLabel(until: until)
                .padding(.top, 16)

            VStack(spacing: 10) {
                Button(action: { Task { await useDevicePasscode() } }) {
                    HStack {
                        Image(systemName: "lock.fill").font(.system(size: 14))
                        Text("Use device passcode")
                            .font(.system(size: 15, weight: .semibold))
                        Spacer()
                        Text("→").font(.system(size: 16))
                    }
                    .foregroundStyle(Color.echoPaper)
                    .padding(.horizontal, 18)
                    .padding(.vertical, 14)
                    .background(Color.echoInk, in: RoundedRectangle(cornerRadius: 14))
                }
                .buttonStyle(SpringPressStyle())

                Button("Recover account via SMS") { triggerRecovery() }
                    .font(.system(size: 13, weight: .medium))
                    .foregroundStyle(Color.echoInk55)
                    .padding(.vertical, 6)
            }
            .padding(.top, 20)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, 28)
    }

    // MARK: - Footer

    private var footerLinks: some View {
        VStack(spacing: 4) {
            if case .normal = loginState {
                Button("Use passkey", action: onPasskeyLogin)
                    .font(.system(size: 13, weight: .medium))
                    .foregroundStyle(Color.echoInk)
                    .padding(8)
            }

            if case .noAccount = loginState {} else {
                Button("Recover account") { triggerRecovery() }
                    .font(.system(size: 12))
                    .foregroundStyle(Color.echoInk55)
                    .padding(4)
            }

            Button("Sign in on new device") { showLinkDevice = true }
                .font(.system(size: 12))
                .foregroundStyle(Color.echoInk55)
                .padding(4)

            Button("New to Echo? Create account", action: onGetStarted)
                .font(.system(size: 12))
                .foregroundStyle(Color.echoSignal)
                .padding(.top, 4)
        }
    }

    // MARK: - Setup

    private func setupLoginState() async {
        // Load saved username. Prefer the Keychain copy, but fall back to the
        // display name persisted at first-run completion — older builds (and the
        // silent-provision path) didn't always mirror the username into the
        // Keychain, which previously stranded returning users on the "Create
        // account" screen even though they had an account.
        let keychainUsername = (try? await KeychainManager.shared.retrieve(
            key: "echo.username.current", as: String.self
        )) ?? ""
        savedUsername = keychainUsername.isEmpty
            ? (UserDefaults.standard.string(forKey: "echo.displayName") ?? "")
            : keychainUsername

        // The authoritative "has an account" signal is first-run completion —
        // the same flag AppState uses to route here — not the presence of a
        // Keychain username string.
        let hasAccount = UserDefaults.standard.bool(forKey: "echo.hasCompletedFirstRun")
        guard hasAccount else {
            loginState = .noAccount
            return
        }

        // Self-heal: if we recovered the username from UserDefaults, mirror it
        // back into the Keychain so future launches read it directly.
        if keychainUsername.isEmpty, !savedUsername.isEmpty {
            try? await KeychainManager.shared.store(
                key: "echo.username.current", value: savedUsername
            )
        }

        // Check biometric availability
        let ctx = LAContext()
        var laError: NSError?
        let biometricsAvailable = ctx.canEvaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics, error: &laError
        )

        if !biometricsAvailable {
            #if targetEnvironment(simulator)
            // Simulator typically has no enrolled Face ID. Match the onboarding
            // simulator handling and proceed to the normal flow — the Secure
            // Enclave simulator key path signs without a biometric prompt — so
            // the returning-user flow stays testable in the Simulator.
            #else
            loginState = .biometricsUnavailable
            return
            #endif
        }

        // Check lockout state
        switch SecureEnclaveManager.shared.currentLockState() {
        case .requiresPasscode:
            loginState = .softLocked
        case .hardLocked(let until):
            loginState = .hardLocked(until: until)
        case .allowed:
            loginState = .normal
            if SessionSignOut.shouldSkipLoginAutoUnlock {
                SessionSignOut.consumeSkipLoginAutoUnlock()
                return
            }
            // Auto-trigger Face ID after brief settle
            try? await Task.sleep(nanoseconds: 400_000_000)
            await triggerBiometricLogin()
        }
    }

    // MARK: - Auth actions

    private func triggerBiometricLogin() async {
        guard !isUnlocking, case .normal = loginState else { return }
        isUnlocking = true
        unlockError = nil

        do {
            let challenge = Data("echo-login-\(UUID().uuidString)".utf8)
            _ = try await SecureEnclaveManager.shared.sign(
                data: challenge,
                keyId: "echo-identity-signing"
            )
            SecureEnclaveManager.shared.recordBiometricSuccess()
            // Call both legacy and new callbacks
            onPasskeyLogin()
            onSuccess()
        } catch {
            isUnlocking = false
            let nsErr = error as NSError
            // Record failure and re-check lock state
            SecureEnclaveManager.shared.recordBiometricFailure()
            let newState = SecureEnclaveManager.shared.currentLockState()
            switch newState {
            case .requiresPasscode:
                loginState = .softLocked
            case .hardLocked(let until):
                loginState = .hardLocked(until: until)
            case .allowed:
                unlockError = nsErr.domain == "com.apple.LocalAuthentication"
                    ? "Face ID failed. Tap to try again."
                    : "Authentication failed. Please try again."
            }
        }
    }

    private func useDevicePasscode() async {
        do {
            try await StepUpAuthManager.shared.authenticateWithPasscodeOnly(
                reason: "Unlock Echo"
            )
            SecureEnclaveManager.shared.recordBiometricSuccess()
            onPasskeyLogin()
            onSuccess()
        } catch {
            unlockError = "Passcode verification failed."
        }
    }

    private func triggerRecovery() {
        onSMSLogin("")
        onRecovery()
    }
}

// MARK: - Countdown label

private struct CountdownLabel: View {
    let until: Date
    @State private var secondsRemaining: Int = 0
    @State private var timer: Timer?

    var body: some View {
        HStack(spacing: 6) {
            Image(systemName: "clock")
                .font(.system(size: 13))
                .foregroundStyle(Color.echoAlert)
            Text(timeString)
                .font(.echomono(13))
                .foregroundStyle(Color.echoAlert)
                .monospacedDigit()
            Text("remaining")
                .font(.system(size: 12))
                .foregroundStyle(Color.echoInk40)
        }
        .onAppear {
            secondsRemaining = max(0, Int(until.timeIntervalSinceNow.rounded()))
            timer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { _ in
                let rem = Int(until.timeIntervalSinceNow.rounded())
                secondsRemaining = max(0, rem)
                if secondsRemaining == 0 { timer?.invalidate() }
            }
        }
        .onDisappear { timer?.invalidate() }
    }

    private var timeString: String {
        let m = secondsRemaining / 60
        let s = secondsRemaining % 60
        return String(format: "%d:%02d", m, s)
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
