import SwiftUI

struct AccountLockedView: View {
    let reason: LockReason
    let retryAfter: Date?
    let onRecovery: () -> Void

    @State private var timeRemaining: String = ""
    @State private var timer: Timer?

    var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 32) {
                Spacer()

                Image(systemName: "lock.shield.fill")
                    .font(.system(size: 64))
                    .foregroundColor(.echoGray400)

                VStack(spacing: 12) {
                    Text(titleText)
                        .typographyStyle(.h3, color: .echoPrimaryText)
                        .multilineTextAlignment(.center)

                    Text(descriptionText)
                        .typographyStyle(.body, color: .echoSecondaryText)
                        .multilineTextAlignment(.center)

                    if !timeRemaining.isEmpty {
                        Text("Try again in \(timeRemaining)")
                            .typographyStyle(.bodyLarge, color: .echoPrimary)
                            .padding(.top, 8)
                    }
                }
                .padding(.horizontal, 32)

                Spacer()

                EchoButton(
                    "Recover Account",
                    style: .secondary,
                    size: .large,
                    icon: Image(systemName: "arrow.counterclockwise"),
                    action: onRecovery
                )
                .padding(.horizontal, 24)
                .padding(.bottom, 32)
            }
        }
        .onAppear { startCountdown() }
        .onDisappear { timer?.invalidate() }
    }

    private var titleText: String {
        switch reason {
        case .tooManyAttempts: return "Too Many Attempts"
        case .suspiciousActivity: return "Suspicious Activity Detected"
        case .accountSuspended: return "Account Suspended"
        }
    }

    private var descriptionText: String {
        switch reason {
        case .tooManyAttempts:
            return "Your account has been temporarily locked due to too many failed login attempts."
        case .suspiciousActivity:
            return "We detected unusual activity on your account and locked it for your protection."
        case .accountSuspended:
            return "Your account has been suspended. Please contact support or try account recovery."
        }
    }

    private func startCountdown() {
        guard let retryAfter else {
            timeRemaining = ""
            return
        }
        updateTimeRemaining(retryAfter)
        timer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { _ in
            updateTimeRemaining(retryAfter)
        }
    }

    private func updateTimeRemaining(_ target: Date) {
        let remaining = target.timeIntervalSinceNow
        if remaining <= 0 {
            timeRemaining = ""
            timer?.invalidate()
            return
        }
        let minutes = Int(remaining) / 60
        let seconds = Int(remaining) % 60
        timeRemaining = String(format: "%d:%02d", minutes, seconds)
    }
}
