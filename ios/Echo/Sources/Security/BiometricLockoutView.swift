#if os(iOS)
// Security/BiometricLockoutView.swift
// WO-211: Shown when the biometric lockout policy trips.
//   - .requiresPasscode (5 failures): prompt passcode fallback
//   - .hardLocked       (10 failures): 15-minute countdown before retrying

import SwiftUI
import LocalAuthentication

/// Full-screen overlay shown when `BiometricLockState` is locked.
/// Clears automatically once the hard lockout expires or the user
/// authenticates with a device passcode.
public struct BiometricLockoutView: View {
    let lockState: BiometricLockState
    let onUnlocked: () -> Void
    let onRecovery: () -> Void

    @State private var secondsRemaining: Int = 0
    @State private var timer: Timer?

    public init(
        lockState: BiometricLockState,
        onUnlocked: @escaping () -> Void = {},
        onRecovery: @escaping () -> Void = {}
    ) {
        self.lockState = lockState
        self.onUnlocked = onUnlocked
        self.onRecovery = onRecovery
    }

    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 32) {
                Spacer()

                // Lock icon
                ZStack {
                    Circle()
                        .fill(iconBackground)
                        .frame(width: 96, height: 96)
                    Image(systemName: iconName)
                        .font(.system(size: 40, weight: .semibold))
                        .foregroundColor(iconColor)
                }

                // Title and message
                VStack(spacing: 12) {
                    Text(title)
                        .font(.system(size: 22, weight: .bold))
                        .foregroundColor(.echoPrimaryText)
                        .multilineTextAlignment(.center)

                    Text(subtitle)
                        .font(.system(size: 15))
                        .foregroundColor(.echoSecondaryText)
                        .multilineTextAlignment(.center)
                        .lineSpacing(3)
                        .padding(.horizontal, 32)
                }

                // Countdown (hard lockout only)
                if case .hardLocked = lockState {
                    CountdownRing(secondsRemaining: $secondsRemaining, total: 15 * 60)
                }

                Spacer()

                // Action buttons
                VStack(spacing: 12) {
                    if showPasscodeButton {
                        EchoButton("Use Device Passcode", style: .primary) {
                            authenticateWithPasscode()
                        }
                    }

                    if showRetryButton {
                        EchoButton("Try \(biometricLabel)", style: .secondary) {
                            retryBiometric()
                        }
                    }

                    // Recover via SMS — shown in hard-locked state
                    if case .hardLocked = lockState {
                        Button("Recover account via SMS") { onRecovery() }
                            .font(.system(size: 13, weight: .medium))
                            .foregroundStyle(Color.echoInk55)
                            .padding(.top, 4)
                    }
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 40)
            }
        }
        .onAppear(perform: startCountdownIfNeeded)
        .onDisappear { timer?.invalidate() }
    }

    // MARK: - Computed

    private var title: String {
        switch lockState {
        case .requiresPasscode: return "Biometric Locked"
        case .hardLocked:       return "Temporarily Locked"
        case .allowed:          return ""
        }
    }

    private var subtitle: String {
        switch lockState {
        case .requiresPasscode(let count):
            return "Too many failed attempts (\(count)/5). Enter your device passcode to continue."
        case .hardLocked:
            let mins = max(1, (secondsRemaining + 59) / 60)
            let unit = mins == 1 ? "minute" : "minutes"
            return secondsRemaining > 0
                ? "Too many failed attempts. Wait \(mins) \(unit) before trying again."
                : "Lockout expired. You may try again."
        case .allowed:
            return ""
        }
    }

    private var iconName: String {
        switch lockState {
        case .requiresPasscode: return "lock.fill"
        case .hardLocked:       return "lock.shield.fill"
        case .allowed:          return "faceid"
        }
    }

    private var iconBackground: Color {
        switch lockState {
        case .requiresPasscode: return .echoWarning.opacity(0.15)
        case .hardLocked:       return .echoError.opacity(0.15)
        case .allowed:          return .echoSuccess.opacity(0.15)
        }
    }

    private var iconColor: Color {
        switch lockState {
        case .requiresPasscode: return .echoWarning
        case .hardLocked:       return .echoError
        case .allowed:          return .echoSuccess
        }
    }

    private var showPasscodeButton: Bool {
        switch lockState {
        case .requiresPasscode: return true
        case .hardLocked:       return secondsRemaining == 0
        default:                return false
        }
    }

    private var showRetryButton: Bool {
        guard case .hardLocked = lockState else { return false }
        return secondsRemaining == 0
    }

    private var biometricLabel: String {
        let ctx = LAContext()
        return ctx.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: nil)
            ? (ctx.biometryType == .faceID ? "Face ID" : "Touch ID")
            : "Biometrics"
    }

    // MARK: - Actions

    private func startCountdownIfNeeded() {
        guard case .hardLocked = lockState,
              let remaining = lockState.remainingLockout else { return }
        secondsRemaining = Int(ceil(remaining))
        timer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { _ in
            if secondsRemaining > 0 {
                secondsRemaining -= 1
            } else {
                timer?.invalidate()
            }
        }
    }

    private func authenticateWithPasscode() {
        let ctx = LAContext()
        ctx.evaluatePolicy(.deviceOwnerAuthentication,
                           localizedReason: "Unlock ECHO") { success, _ in
            if success {
                DispatchQueue.main.async {
                    SecureEnclaveManager().recordBiometricSuccess()
                    onUnlocked()
                }
            }
        }
    }

    private func retryBiometric() {
        let ctx = LAContext()
        ctx.evaluatePolicy(.deviceOwnerAuthenticationWithBiometrics,
                           localizedReason: "Unlock ECHO") { success, _ in
            if success {
                DispatchQueue.main.async {
                    SecureEnclaveManager().recordBiometricSuccess()
                    onUnlocked()
                }
            }
        }
    }
}

// MARK: - Countdown ring

private struct CountdownRing: View {
    @Binding var secondsRemaining: Int
    let total: Int

    private var progress: Double {
        total > 0 ? Double(secondsRemaining) / Double(total) : 0
    }

    var body: some View {
        ZStack {
            Circle()
                .stroke(Color.echoGray200, lineWidth: 8)
                .frame(width: 120, height: 120)

            Circle()
                .trim(from: 0, to: CGFloat(progress))
                .stroke(Color.echoError, style: StrokeStyle(lineWidth: 8, lineCap: .round))
                .frame(width: 120, height: 120)
                .rotationEffect(.degrees(-90))
                .animation(.linear(duration: 1), value: secondsRemaining)

            VStack(spacing: 2) {
                Text(timeString)
                    .font(.system(size: 26, weight: .bold, design: .monospaced))
                    .foregroundColor(.echoPrimaryText)
                Text("remaining")
                    .font(.system(size: 11))
                    .foregroundColor(.echoSecondaryText)
            }
        }
    }

    private var timeString: String {
        let m = secondsRemaining / 60
        let s = secondsRemaining % 60
        return String(format: "%d:%02d", m, s)
    }
}

// MARK: - Preview

#Preview("Requires Passcode") {
    BiometricLockoutView(lockState: .requiresPasscode(failureCount: 5)) {}
}

#Preview("Hard Locked") {
    BiometricLockoutView(lockState: .hardLocked(until: Date().addingTimeInterval(720))) {}
}
#endif
