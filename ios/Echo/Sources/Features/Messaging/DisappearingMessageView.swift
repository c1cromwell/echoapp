#if os(iOS)
import SwiftUI

/// Live countdown for disappearing messages (WO-105).
struct CountdownTimer: View {
    let expiresAt: Date
    let onExpired: () -> Void

    @State private var remaining: TimeInterval = 0
    @State private var fired = false

    private let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()

    var body: some View {
        Text(formatted(remaining))
            .font(.system(size: 11, weight: .medium, design: .monospaced))
            .foregroundColor(.echoInk55)
            .onAppear { tick() }
            .onReceive(timer) { _ in tick() }
    }

    private func tick() {
        remaining = max(0, expiresAt.timeIntervalSinceNow)
        guard remaining <= 0, !fired else { return }
        fired = true
        onExpired()
    }

    private func formatted(_ interval: TimeInterval) -> String {
        let total = Int(interval.rounded(.up))
        if total >= 86400 { return "\(total / 86400)d" }
        if total >= 3600 { return "\(total / 3600)h" }
        if total >= 60 { return "\(total / 60)m" }
        return "\(max(0, total))s"
    }
}

/// Bubble wrapper that shows countdown and triggers local deletion (WO-105).
struct DisappearingMessageView<Content: View>: View {
    let messageId: String
    let expiresAt: Date?
    let onExpired: (String) -> Void
    @ViewBuilder let content: () -> Content

    var body: some View {
        HStack(alignment: .bottom, spacing: 6) {
            content()
            if let expiresAt {
                CountdownTimer(expiresAt: expiresAt) {
                    onExpired(messageId)
                }
            }
        }
    }
}
#endif
