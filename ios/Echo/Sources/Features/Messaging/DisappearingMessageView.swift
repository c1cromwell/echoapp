#if os(iOS)
import SwiftUI

/// Wrapper that tracks expiry of a disappearing message and fires a callback (WO-105).
struct DisappearingMessageView<Content: View>: View {
    let messageId: String
    let expiresAt: Date?
    let onExpired: (String) -> Void
    @ViewBuilder let content: () -> Content

    @State private var isExpired = false

    var body: some View {
        if !isExpired {
            content()
                .overlay(alignment: .bottomTrailing) {
                    if let expiresAt {
                        ExpiryIndicator(expiresAt: expiresAt)
                    }
                }
                .task {
                    guard let expiresAt else { return }
                    let remaining = expiresAt.timeIntervalSinceNow
                    if remaining <= 0 {
                        isExpired = true
                        onExpired(messageId)
                        return
                    }
                    try? await Task.sleep(for: .seconds(remaining))
                    guard !Task.isCancelled else { return }
                    isExpired = true
                    onExpired(messageId)
                }
        }
    }
}

/// Small countdown indicator for disappearing messages.
private struct ExpiryIndicator: View {
    let expiresAt: Date

    var body: some View {
        Image(systemName: "timer")
            .font(.system(size: 10))
            .foregroundStyle(Color.echoInk40)
            .padding(4)
    }
}
#endif
