#if os(iOS)
import Foundation
import SwiftUI
@testable import Echo

enum ScreenCatalogFixtures {
    @MainActor
    static func firstRunCoordinator(displayName: String = "alex") -> FirstRunCoordinator {
        let coordinator = FirstRunCoordinator(onComplete: { _, _ in }, onRestoreComplete: { _ in })
        coordinator.displayName = displayName
        return coordinator
    }
}

struct CatalogUsernameClient: UsernameAvailabilityClient {
    let available: Bool

    func checkAvailability(username: String) async throws -> UsernameAvailabilityResult {
        UsernameAvailabilityResult(
            username: username,
            available: available,
            reason: available ? nil : "taken"
        )
    }
}

/// Composite chat thread preview for catalog (not a shipped screen).
struct CatalogChatThreadPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            catalogHeader
            ScrollView {
                VStack(spacing: 12) {
                    MessageBubble(
                        message: "Hey — did you get the invite link?",
                        isSent: false,
                        status: .delivered,
                        timestamp: "9:41 AM"
                    )
                    MessageBubble(
                        message: "Yes! End-to-end encrypted on this thread.",
                        isSent: true,
                        status: .read,
                        deliveryStatus: .read,
                        timestamp: "9:42 AM"
                    )
                    TypingIndicatorView(label: "Jordan is typing…")
                }
                .padding(.vertical, 16)
            }
            ReactionPickerView(onSelect: { _ in }, onDismiss: {})
                .padding(.bottom, 8)
            catalogComposer
        }
        .background(Color.Echo.surface)
    }

    private var catalogHeader: some View {
        HStack(spacing: 12) {
            Circle()
                .fill(Color.echoSignal.opacity(0.2))
                .frame(width: 36, height: 36)
                .overlay(Text("J").font(.headline).foregroundStyle(Color.echoSignal))
            VStack(alignment: .leading, spacing: 2) {
                Text("Jordan")
                    .font(.system(size: 17, weight: .semibold))
                Text("Verified · end-to-end encrypted")
                    .font(.caption)
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            Spacer()
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(Color.Echo.surfaceContainerLow)
    }

    private var catalogComposer: some View {
        HStack(spacing: 10) {
            Image(systemName: "plus.circle.fill")
                .font(.title2)
                .foregroundStyle(Color.echoSignal)
            Text("Message")
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            Spacer()
            Image(systemName: "mic.fill")
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(Color.Echo.surfaceContainerLow)
    }
}
#endif
